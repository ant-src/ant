// Package parser builds the syntax tree of an Ant module. It implements sections 2, 5, 6, and 10
// of the specification with one token of lookahead and no feedback from the checker.
package parser

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/lexer"
	"github.com/ant-src/ant/internal/token"
	"github.com/ant-src/ant/value"
)

// Mode selects what a source may contain at the top level.
type Mode int

const (
	// Program is a module: imports followed by declarations and statements (section 10).
	Program Mode = iota
	// Spec also allows top-level statements and "..." placeholders, for the specification's
	// own examples.
	Spec
	// Session also allows top-level statements, for a REPL.
	Session
	// Declarations is a declaration file (section 10, Declaration files): declarations only,
	// each qualified by its owner, a func without a body, a type without one, and a var as a
	// typed constant; the prelude and a host namespace's signatures.
	Declarations
)

// ModeOf is the mode a file's path chooses: Declarations for a declaration file, whose name
// ends in .d.ant, and Program otherwise.
func ModeOf(path string) Mode {
	if strings.HasSuffix(path, ".d.ant") {
		return Declarations
	}
	return Program
}

// declarations reports whether the source holds signatures rather than a program.
func (p *parser) declarations() bool { return p.mode == Declarations }

// Error is a syntax error. Incomplete is set when the input ended too early, which a REPL
// uses to keep reading.
type Error struct {
	Pos        ast.Pos
	Msg        string
	Incomplete bool
}

func (e *Error) Error() string { return fmt.Sprintf("%d:%d: %s", e.Pos.Line, e.Pos.Col, e.Msg) }

// bail unwinds a recovering parser to the nearest point that can go on, once fail has recorded
// the error: the statement, list item, or interpolation being parsed.
type bail struct{}

// ParseRecover parses one module in the mode of its path (ModeOf) as ParseLayout does, but goes
// on past its syntax errors, for an editor. It returns the tree of what parsed, with a BadStmt,
// BadExpr, or BadType standing where a statement, an expression, or a type could not be read,
// the layout, and every error in source order: the lexer's, then the parser's as it met them.
// The error Parse reports is among them, and the rest are those Parse would report once the
// earlier ones were fixed, as far as the parser can tell. A Bad token, which the lexer has
// reported, ends a statement for the parser, as a line break would, so that it is reported once.
// The tree is for analysis only and must not be run.
func ParseRecover(path, src string) (*ast.File, *Layout, []*Error) {
	mode := ModeOf(path)
	toks, comments, lerrs := lexer.LexRecover(src, mode == Program)
	l := &Layout{Tokens: toks, Comments: comments, Roles: map[int]Role{}, Prefix: map[int]bool{}}
	parsed := toks
	if slices.ContainsFunc(toks, func(t token.Token) bool { return t.Kind == token.Bad }) {
		parsed = slices.Clone(toks)
		for i := range parsed {
			if parsed[i].Kind == token.Bad {
				parsed[i] = token.Token{Kind: token.Newline, Pos: parsed[i].Pos, EndPos: parsed[i].Pos, Off: parsed[i].Off, End: parsed[i].Off}
			}
		}
	}
	p := &parser{mode: mode, src: lexer.NormalizeBreaks(src), toks: parsed, comments: comments,
		roles: l.Roles, prefix: l.Prefix, recovering: true, recorded: map[Error]bool{}}
	for _, e := range lerrs {
		p.record(&Error{Pos: cpos(e.Pos), Msg: e.Msg})
	}
	var f *ast.File
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(bail); !ok {
					panic(r)
				}
			}
		}()
		f = p.file(path)
	}()
	if f == nil {
		f = &ast.File{Path: path, End: epos(toks[len(toks)-1])}
	}
	slices.SortStableFunc(p.errs, func(a, b *Error) int {
		if a.Pos.Line != b.Pos.Line {
			return a.Pos.Line - b.Pos.Line
		}
		return a.Pos.Col - b.Pos.Col
	})
	return f, l, p.errs
}

// Parse parses one module. path only labels the result.
func Parse(path, src string, mode Mode) (f *ast.File, err error) {
	toks, comments, lerr := lexer.LexLayout(src, mode == Program)
	if lerr != nil {
		return nil, lexError(src, lerr)
	}
	p := &parser{mode: mode, src: lexer.NormalizeBreaks(src), toks: toks, comments: comments}
	defer p.recover(&err)
	return p.file(path), nil
}

func lexError(src string, err error) *Error {
	le := err.(*lexer.Error)
	return &Error{Pos: ast.Pos{Line: le.Pos.Line, Col: le.Pos.Col}, Msg: le.Msg, Incomplete: incompleteLex(src, le)}
}

// Layout is what the formatter needs beyond the tree: the tokens with their spans, the
// comments, and the role of every opening bracket, keyed by its byte offset.
type Layout struct {
	Tokens   []token.Token
	Comments []token.Comment
	Roles    map[int]Role
	// Prefix holds the offsets of the - and ! tokens the parser read as prefix operators.
	Prefix map[int]bool
}

// Role is what an opening bracket is, as the parser decided while parsing it.
type Role struct {
	Block bool // a block or a lambda body, in which newlines separate statements
	Typed bool // the { of a record or map literal, directly after its type
	Comma bool // a list that allows a comma after its last item
}

// ParseLayout parses one module, as Parse does in the mode of its path (ModeOf), and also
// returns its Layout.
func ParseLayout(path, src string) (f *ast.File, l *Layout, err error) {
	mode := ModeOf(path)
	toks, comments, lerr := lexer.LexLayout(src, mode == Program)
	if lerr != nil {
		return nil, nil, lexError(src, lerr)
	}
	l = &Layout{Tokens: toks, Comments: comments, Roles: map[int]Role{}, Prefix: map[int]bool{}}
	p := &parser{mode: mode, src: lexer.NormalizeBreaks(src), toks: toks, comments: comments, roles: l.Roles, prefix: l.Prefix}
	defer p.recover(&err)
	return p.file(path), l, nil
}

// incompleteLex reports whether a lexer error means the source ended too early: an open string,
// raw string, or block comment, which may span lines, or a regular expression open on the last
// line. A regular expression stays on one line, so a line break ends it in an error instead.
func incompleteLex(src string, le *lexer.Error) bool {
	if !strings.HasPrefix(le.Msg, "unterminated") {
		return false
	}
	if le.Msg == "unterminated regular expression" {
		return le.Pos.Line == strings.Count(lexer.NormalizeBreaks(src), "\n")+1
	}
	return true
}

// ParseExpr parses a single expression, in spec mode.
func ParseExpr(src string) (e ast.Expr, err error) {
	toks, lerr := lexer.Lex(src)
	if lerr != nil {
		le := lerr.(*lexer.Error)
		return nil, &Error{Pos: ast.Pos{Line: le.Pos.Line, Col: le.Pos.Col}, Msg: le.Msg}
	}
	p := &parser{mode: Spec, src: lexer.NormalizeBreaks(src), toks: toks}
	defer p.recover(&err)
	e = p.expression()
	p.skipTerms()
	if !p.atKind(token.EOF) {
		p.fail(p.peek().Pos, "unexpected %s", p.peek())
	}
	return e, nil
}

type parser struct {
	mode     Mode
	src      string // with every line break holding a \n (lexer.NormalizeBreaks)
	toks     []token.Token
	i        int
	comments []token.Comment // for doc comments
	ci       int             // comments before ci end before the declaration being parsed
	nl       int             // > 0: newlines are ignored (inside a list)
	header   bool            // inside an if, while, or for header: Path { is not a literal
	boundary string          // "", "script", "func", "lambda", "go", "defer", or "default" (a field's)
	loops    int             // loops entered inside the current boundary
	// inDefault > 0: inside a parameter default, which runs in the callee, so a try there returns
	// from it
	inDefault int
	// bodyOnly is the first try, return, or defer at the top level of a program module, which
	// only a script body allows; the module is a script only once a statement has been seen.
	bodyOnly *Error
	curBlock *ast.Block   // the innermost block whose statements are being parsed
	depth    int          // nesting of expressions, types, and blocks, bounded by maxNesting
	roles    map[int]Role // filled for ParseLayout, nil otherwise
	prefix   map[int]bool // likewise
	// recovering makes the parser go on past an error (ParseRecover): fail records the error
	// in errs and unwinds to the nearest recovery point, and soft records it and goes on.
	recovering bool
	errs       []*Error
	recorded   map[Error]bool // the errors in errs, by position and message
}

// mark records the role of the opening bracket t, for ParseLayout.
func (p *parser) mark(t token.Token, r Role) {
	if p.roles != nil {
		p.roles[t.Off] = r
	}
}

// maxNesting bounds the nesting of expressions, types, blocks and destructurings, so that no input
// can overflow the Go stack in the parser or in the passes that walk the tree after it.
const maxNesting = 1000

func (p *parser) nest() {
	p.depth++
	if p.depth > maxNesting {
		p.fail(p.peek().Pos, "nesting deeper than %d levels", maxNesting)
	}
}

func (p *parser) unnest() { p.depth-- }

// link counts one more node of a left-nested chain such as a + b + c or f()()(), which the passes
// after the parser recurse through as deeply as through parentheses. unlink releases the links a
// function counted when it returns.
func (p *parser) link(n *int) {
	p.nest()
	*n++
}

func (p *parser) unlink(n *int) { p.depth -= *n }

func (p *parser) recover(err *error) {
	if r := recover(); r != nil {
		if e, ok := r.(*Error); ok {
			*err = e
			return
		}
		panic(r)
	}
}

func (p *parser) fail(pos token.Pos, format string, args ...any) {
	incomplete := p.i < len(p.toks) && p.toks[p.i].Kind == token.EOF
	e := &Error{Pos: ast.Pos{Line: pos.Line, Col: pos.Col}, Msg: fmt.Sprintf(format, args...), Incomplete: incomplete}
	if p.recovering {
		p.record(e)
		panic(bail{})
	}
	panic(e)
}

// soft reports an error the parser can go on from when recovering, and returns true so that
// the caller takes the way out it names; otherwise it fails as fail does.
func (p *parser) soft(pos token.Pos, format string, args ...any) bool {
	if !p.recovering {
		p.fail(pos, format, args...)
	}
	p.record(&Error{Pos: cpos(pos), Msg: fmt.Sprintf(format, args...)})
	return true
}

// record keeps an error, once per position and message.
func (p *parser) record(e *Error) {
	key := Error{Pos: e.Pos, Msg: e.Msg}
	if p.recorded[key] {
		return
	}
	p.recorded[key] = true
	p.errs = append(p.errs, e)
}

// state is what a recovery point restores: the nesting a failed construct did not leave.
type state struct {
	nl        int
	header    bool
	block     *ast.Block
	boundary  string
	loops     int
	inDefault int
}

func (p *parser) save() state {
	return state{p.nl, p.header, p.curBlock, p.boundary, p.loops, p.inDefault}
}

func (p *parser) restore(s state) {
	p.nl, p.header, p.curBlock, p.boundary, p.loops, p.inDefault = s.nl, s.header, s.block, s.boundary, s.loops, s.inDefault
}

// attempt runs f. When recovering, a syntax error inside f ends it: attempt restores the state
// around it and returns false, with the error recorded and the cursor where the error was.
func (p *parser) attempt(f func()) (ok bool) {
	if !p.recovering {
		f()
		return true
	}
	saved := p.save()
	defer func() {
		if r := recover(); r != nil {
			if _, isBail := r.(bail); !isBail {
				panic(r)
			}
			p.restore(saved)
			ok = false
		}
	}()
	f()
	return true
}

// statementStart lists the keywords that begin a statement a list cannot hold, so that a list
// left open above one of them ends there when recovering.
var statementStart = map[string]bool{"var": true, "type": true, "func": true, "export": true, "import": true,
	"return": true, "for": true, "while": true, "if": true, "select": true, "go": true, "defer": true, "break": true, "continue": true}

// sync skips to where parsing can go on after an error, when recovering: to a line break or
// semicolon outside brackets, or to a comma inside a list, or up to the closer close of the
// enclosing block or list, or to a closing brace that must end an enclosing block, or to the
// end of the file. A statement on a line of its own, which the lexer may have joined to the line
// before, ends the skipping too, once the item that failed, begun at from, has consumed
// something, so that the parser makes progress. A stray closer is skipped.
func (p *parser) sync(close string, list bool, from int) {
	depth := 0
	for {
		t := p.raw()
		switch {
		case t.Kind == token.EOF:
			return
		case t.Kind == token.Newline || t.Is(";"):
			if depth <= 0 && !list {
				return
			}
		case depth <= 0 && p.i > from && p.statementStarts(p.i):
			return
		case t.Is(",") && list && depth <= 0:
			return
		case t.Is("(") || t.Is("[") || t.Is("{"):
			depth++
		case t.Is(")") || t.Is("]") || t.Is("}"):
			if depth <= 0 && (t.Is(close) || t.Is("}") && close != "") {
				return
			}
			depth--
		}
		p.i++
	}
}

// statementStarts reports whether the token at j begins a statement on a line of its own: a
// keyword that begins one, after a line break, which the lexer may have joined to the line
// before when that line ended in an operator or an open bracket (section 2).
func (p *parser) statementStarts(j int) bool {
	t := p.toks[j]
	if t.Kind != token.Keyword || !statementStart[t.Text] {
		return false
	}
	for k := j - 1; k >= 0; k-- {
		switch prev := p.toks[k]; {
		case prev.Kind == token.Newline:
			continue
		default:
			return t.Pos.Line > prev.EndPos.Line
		}
	}
	return true
}

// listEnds reports whether, when recovering, the list closed by close has ended without its
// closer: at the end of the file, at a brace that ends an enclosing block, or at a statement
// that begins on a line of its own.
func (p *parser) listEnds(close string) bool {
	if !p.recovering {
		return false
	}
	t := p.peek()
	if t.Kind == token.EOF || t.Is("}") && close != "}" {
		return true
	}
	return p.statementStarts(p.i)
}

// badExpr is a BadExpr for the tokens from start to the cursor, or an empty one just after the
// last token read when nothing was, so that it lies inside the node that lacks it.
func (p *parser) badExpr(start int) *ast.BadExpr {
	if p.i <= start {
		return &ast.BadExpr{Pos: p.end(), End: p.end()}
	}
	return &ast.BadExpr{Pos: pos(p.toks[start]), End: p.end()}
}

func (p *parser) badStmt(start int) *ast.BadStmt {
	e := p.badExpr(start)
	return &ast.BadStmt{Pos: e.Pos, End: e.End}
}

// failAt reports an error at t, unless t is a reserved word: then the error says so, since the
// word is the likelier mistake than whatever was expected in its place.
func (p *parser) failAt(t token.Token, format string, args ...any) {
	if t.Kind == token.Reserved {
		p.fail(t.Pos, "'%s' is a reserved word", t.Text)
	}
	p.fail(t.Pos, format, args...)
}

func pos(t token.Token) ast.Pos { return ast.Pos{Line: t.Pos.Line, Col: t.Pos.Col} }

// epos is the position just after t.
func epos(t token.Token) ast.Pos { return ast.Pos{Line: t.EndPos.Line, Col: t.EndPos.Col} }

// cpos converts a lexer position.
func cpos(p token.Pos) ast.Pos { return ast.Pos{Line: p.Line, Col: p.Col} }

// end is the position just after the last token read, skipping the line breaks peek may have
// stepped over inside a list.
func (p *parser) end() ast.Pos {
	for j := p.i - 1; j >= 0; j-- {
		if t := p.toks[j]; t.Kind != token.Newline {
			return epos(t)
		}
	}
	return ast.Pos{Line: 1, Col: 1}
}

// name makes the target a name token binds.
func name(t token.Token) *ast.NameTarget {
	return &ast.NameTarget{Pos: pos(t), End: epos(t), Name: t.Text}
}

// ---- tokens

func (p *parser) raw() token.Token { return p.toks[p.i] }

func (p *parser) peek() token.Token {
	for p.nl > 0 && p.toks[p.i].Kind == token.Newline {
		p.i++
	}
	return p.toks[p.i]
}

// peekAfter returns the token after the next one, skipping newlines inside a list.
func (p *parser) peekAfter() token.Token {
	p.peek()
	j := p.i + 1
	for p.nl > 0 && j < len(p.toks) && p.toks[j].Kind == token.Newline {
		j++
	}
	if j >= len(p.toks) {
		return p.toks[len(p.toks)-1]
	}
	return p.toks[j]
}

func (p *parser) next() token.Token {
	t := p.peek()
	if t.Kind != token.EOF {
		p.i++
	}
	return t
}

func (p *parser) at(text string) bool      { return p.peek().Is(text) }
func (p *parser) atKind(k token.Kind) bool { return p.peek().Kind == k }
func (p *parser) unexpected(t token.Token) { p.failAt(t, "unexpected %s", t) }

func (p *parser) expect(text string) token.Token {
	t := p.peek()
	if !t.Is(text) {
		p.failAt(t, "expected '%s', got %s", text, t)
	}
	p.i++
	return t
}

// expectIdent accepts an identifier; keywords and reserved words are errors.
func (p *parser) expectIdent(what string) token.Token {
	t := p.peek()
	if t.Kind != token.Ident {
		if t.Kind == token.Reserved {
			p.fail(t.Pos, "'%s' is a reserved word and cannot be %s", t.Text, what)
		}
		if t.Kind == token.Keyword {
			p.fail(t.Pos, "'%s' is a keyword and cannot be %s", t.Text, what)
		}
		p.fail(t.Pos, "expected %s, got %s", what, t)
	}
	p.i++
	return t
}

// wordKind names the list a keyword or reserved word comes from, for an error message.
func wordKind(t token.Token) string {
	if t.Kind == token.Reserved {
		return "a reserved word"
	}
	return "a keyword"
}

// expectName accepts a field name: an identifier, a keyword, or a reserved word (section 2).
func (p *parser) expectName(what string) token.Token {
	t := p.peek()
	if t.Kind != token.Ident && t.Kind != token.Keyword && t.Kind != token.Reserved {
		p.fail(t.Pos, "expected %s, got %s", what, t)
	}
	p.i++
	return t
}

func (p *parser) skipTerms() int {
	n := 0
	for {
		t := p.raw()
		if t.Kind == token.Newline || t.Is(";") {
			p.i++
			n++
		} else {
			return n
		}
	}
}

func (p *parser) expectTerm() {
	t := p.raw()
	if t.Kind == token.EOF || t.Is("}") {
		return
	}
	if p.recovering && p.i > 0 && (p.toks[p.i-1].Kind == token.Newline || t.Pos.Line > p.toks[p.i-1].EndPos.Line) {
		// A list left open stepped over the line break, or the lexer joined the lines because
		// the statement was unfinished: the break is there in the source, and the error that
		// left the statement unfinished has been reported.
		return
	}
	if p.skipTerms() == 0 {
		if t.Is(":") && p.i >= 2 && p.toks[p.i-2].Is("{") && (p.toks[p.i-1].Kind == token.Ident || p.toks[p.i-1].Kind == token.String) {
			// { name: value } reads as a block or lambda; the author meant a literal without a head.
			p.failAt(t, "a record or map literal names its type, as in Point{ x: 1 } or Map{ \"a\": 1 }")
		}
		if t.Kind == token.Reserved {
			p.failAt(t, "")
		}
		p.soft(t.Pos, "expected a newline or ';' before %s", t) // when recovering, the next statement starts here
	}
}

// ---- nesting state

type nesting struct {
	nl     int
	header bool
	block  *ast.Block
}

// enterList starts a parenthesis, bracket, or non-block brace, inside which newlines are ignored.
func (p *parser) enterList() nesting {
	s := nesting{p.nl, p.header, p.curBlock}
	p.nl++
	p.header = false
	return s
}

// enterBlock starts a block or a lambda body, inside which newlines separate statements.
func (p *parser) enterBlock(b *ast.Block) nesting {
	s := nesting{p.nl, p.header, p.curBlock}
	p.nl = 0
	p.header = false
	p.curBlock = b
	return s
}

func (p *parser) leave(s nesting) { p.nl, p.header, p.curBlock = s.nl, s.header, s.block }

// list parses comma-separated items up to and including close, after the opening token. A
// trailing comma is allowed, and newlines inside are ignored.
func (p *parser) list(close string, item func()) {
	p.mark(p.toks[p.i-1], Role{Comma: true})
	saved := p.enterList()
	for !p.at(close) {
		if p.listEnds(close) {
			break
		}
		start := p.i
		if !p.attempt(item) || p.i == start {
			// The item could not be read, or read nothing: skip to the next item or the end.
			p.sync(close, true, start)
			if p.at(",") {
				p.next()
			}
			continue
		}
		if !p.at(close) {
			if p.recovering && !p.at(",") {
				if p.listEnds(close) {
					break
				}
				p.soft(p.peek().Pos, "expected ',', got %s", p.peek())
				continue
			}
			p.expect(",")
		}
	}
	if p.recovering && !p.at(close) {
		p.soft(p.peek().Pos, "expected '%s', got %s", close, p.peek())
	} else {
		p.expect(close)
	}
	p.leave(saved)
}

// expectClose expects the brace that ends a block. When recovering, a block left open at the end
// of the file ends there.
func (p *parser) expectClose() {
	if p.recovering && p.atKind(token.EOF) {
		p.soft(p.peek().Pos, "expected '}', got end of file")
		return
	}
	p.expect("}")
}

type boundary struct {
	kind  string
	loops int
}

func (p *parser) enterBoundary(kind string) boundary {
	s := boundary{p.boundary, p.loops}
	p.boundary = kind
	p.loops = 0
	return s
}

func (p *parser) leaveBoundary(s boundary) { p.boundary, p.loops = s.kind, s.loops }

// ---- module

func (p *parser) file(path string) *ast.File {
	f := &ast.File{Path: path, Declarations: p.declarations()}
	if p.mode == Program || p.mode == Session {
		// a script's top level, and each input of a session, is a body that try, return, and
		// defer can leave or register on (section 10, Scripts)
		p.boundary = "script"
	}
	p.skipTerms()
	f.Doc = p.moduleDoc(p.peek())
	for p.at("import") {
		start := p.i
		if !p.attempt(func() {
			if p.declarations() {
				p.fail(p.peek().Pos, "a declaration file takes no imports: it sees the prelude and its own declarations")
			}
			f.Imports = append(f.Imports, p.importDecl())
		}) {
			p.sync("", false, start)
			f.Stmts = append(f.Stmts, p.badStmt(start))
		}
		p.terminate("")
	}
	for !p.atKind(token.EOF) {
		start := p.i
		added := false
		if !p.attempt(func() { added = p.item(f) }) || !added {
			if p.i == start {
				p.sync("", false, start)
			}
			f.Stmts = append(f.Stmts, p.badStmt(start))
		}
		p.terminate("")
	}
	if p.mode == Program {
		p.scriptRules(f)
	}
	f.End = epos(p.toks[len(p.toks)-1])
	return f
}

// terminate expects the end of a statement and skips the terminators that follow. When
// recovering, what cannot end one, such as a reserved word, is skipped to the next statement.
func (p *parser) terminate(close string) {
	if !p.attempt(p.expectTerm) {
		p.sync(close, false, -1)
	}
	p.skipTerms()
}

// item parses one top-level item into f: a declaration, or a statement, which makes a program
// module a script. It reports whether it added one; a statement that did not parse adds nothing.
func (p *parser) item(f *ast.File) bool {
	t := p.peek()
	switch {
	case t.Is("import"):
		p.fail(t.Pos, "imports must come before declarations")
	case t.Is("export"), t.Is("func"), t.Is("type"):
		doc := p.doc(t)
		d := p.declaration()
		switch d := d.(type) {
		case *ast.VarDecl:
			d.Doc = doc
		case *ast.FuncDecl:
			d.Doc = doc
		case *ast.TypeDecl:
			d.Doc = doc
		}
		f.Decls = append(f.Decls, d)
	case t.Is("var"):
		doc := p.doc(t)
		d := p.varDecl()
		d.Doc = doc
		f.Decls = append(f.Decls, d)
		if !p.declarations() {
			f.Stmts = append(f.Stmts, d)
		}
	default:
		if p.declarations() {
			p.failAt(t, "expected a declaration, got %s", t)
		}
		s := p.statement()
		if _, bad := s.(*ast.BadExpr); bad {
			return false // what did not parse does not make the module a script
		}
		if p.mode == Program && !f.Script {
			f.Script = true
			f.FirstStmt = pos(t)
		}
		if p.mode == Program {
			p.discardedHtml(s)
		}
		f.Stmts = append(f.Stmts, s)
	}
	return true
}

// scriptRules enforces what a module's statements decide (section 10, Scripts): a script
// exports nothing, and only a script's top level is a body that try, return, and defer can
// leave or register on.
func (p *parser) scriptRules(f *ast.File) {
	if !f.Script {
		if p.bodyOnly != nil {
			if p.recovering {
				p.record(p.bodyOnly)
				return
			}
			panic(p.bodyOnly)
		}
		return
	}
	for _, d := range f.Decls {
		if d.Exported() {
			e := &Error{Pos: f.FirstStmt, Msg: "a module with exports cannot have top-level statements; this statement makes it a script, which cannot be imported"}
			if p.recovering {
				p.record(e)
				return
			}
			panic(e)
		}
	}
}

// topLevelOnly records a try, return, or defer at the top level of a program module, which is
// an error unless the module turns out to be a script.
func (p *parser) topLevelOnly(at token.Pos, msg string) {
	if p.bodyOnly == nil {
		p.bodyOnly = &Error{Pos: ast.Pos{Line: at.Line, Col: at.Col}, Msg: msg}
	}
}

func (p *parser) importDecl() *ast.Import {
	t := p.expect("import")
	pt := p.peek()
	if pt.Kind != token.String {
		p.fail(pt.Pos, "expected a module path string, got %s", pt)
	}
	p.next()
	imp := &ast.Import{Pos: pos(t), Path: pt.Text, PathPos: pos(pt), PathEnd: epos(pt)}
	if p.atKind(token.Ident) {
		ns := p.next()
		imp.Namespace, imp.NamespacePos = ns.Text, pos(ns)
		imp.End = p.end()
		return imp
	}
	p.expect("{")
	p.list("}", func() {
		name := p.expectIdent("an export name")
		item := ast.ImportName{Pos: pos(name), Name: name.Text, Alias: name.Text}
		if p.at(":") {
			p.next()
			alias := p.expectIdent("a local name")
			item.Alias, item.AliasPos = alias.Text, pos(alias)
		}
		item.End = p.end()
		imp.Names = append(imp.Names, item)
	})
	imp.End = p.end()
	return imp
}

func (p *parser) declaration() ast.Decl {
	export := false
	if p.at("export") {
		if p.declarations() {
			p.fail(p.peek().Pos, "a declaration file takes no export: every declaration in it is visible")
		}
		p.next()
		export = true
	}
	t := p.peek()
	switch {
	case t.Is("var"):
		d := p.varDecl()
		d.Export = export
		return d
	case t.Is("func"):
		d := p.funcDecl()
		d.Export = export
		return d
	case t.Is("type"):
		d := p.typeDecl()
		d.Export = export
		return d
	}
	p.fail(t.Pos, "expected var, func, or type after export, got %s", t)
	return nil
}

func (p *parser) varDecl() *ast.VarDecl {
	t := p.expect("var")
	d := &ast.VarDecl{Pos: pos(t)}
	switch {
	case p.at("{") || p.at("["):
		if p.declarations() {
			// when recovering, the var is read as written, so that one error is reported
			p.soft(p.peek().Pos, "a declaration file declares a constant as var Owner.NAME: T")
		}
		d.Target = p.destructure()
	case p.atKind(token.Ident):
		n := p.next()
		d.Target = name(n)
		if p.declarations() && p.at(".") {
			p.next()
			d.Owner, d.OwnerPos = n.Text, pos(n)
			m := p.expectName("a constant name")
			d.Target = name(m)
			p.expect(":")
			d.Type = p.parseType()
			d.Sig = p.since(t)
			d.End = p.end()
			return d
		}
		if p.declarations() {
			// when recovering, the var is read as written, so that one error is reported
			p.soft(p.peek().Pos, "a declaration file declares a constant as var Owner.NAME: T")
			if p.at(":") {
				p.next()
				d.Type = p.parseType()
			}
			if p.at("=") {
				p.next()
				d.Init = p.expression()
			}
			d.Sig = p.since(t)
			d.End = p.end()
			return d
		}
		if p.at(":") {
			p.next()
			d.Type = p.parseType()
		}
	default:
		n := p.peek()
		if n.Kind == token.Reserved || n.Kind == token.Keyword {
			p.fail(n.Pos, "'%s' cannot be a variable name", n.Text)
		}
		p.fail(n.Pos, "expected a name or a destructuring after var, got %s", n)
	}
	d.Sig = p.since(t)
	if p.recovering && !p.at("=") {
		p.soft(p.peek().Pos, "expected '=', got %s", p.peek())
		d.Init = p.badExpr(p.i)
		d.End = p.end()
		return d
	}
	p.expect("=")
	d.Init = p.expression()
	d.End = p.end()
	return d
}

func (p *parser) funcDecl() *ast.FuncDecl {
	t := p.expect("func")
	name := p.expectIdent("a function name")
	d := &ast.FuncDecl{Pos: pos(t), Name: name.Text, NamePos: pos(name)}
	if p.at(".") {
		p.next()
		d.Owner, d.OwnerPos = d.Name, d.NamePos
		m := p.expectName("a function name")
		d.Name, d.NamePos = m.Text, pos(m)
		if p.at(".") {
			p.fail(p.peek().Pos, "a function belongs to one type: func Type.name")
		}
	}
	if p.at("[") {
		d.TypeParams = p.typeParams()
	}
	d.Params, d.Return = p.signature()
	d.Sig = p.since(t)
	if p.declarations() {
		if p.at("{") {
			// when recovering, the body is read and dropped, so that one error is reported
			p.soft(p.peek().Pos, "a func in a declaration file has a signature and no body")
			saved := p.enterBoundary("func")
			p.block()
			p.leaveBoundary(saved)
		}
		d.End = p.end()
		return d
	}
	if p.recovering && !p.at("{") {
		p.soft(p.peek().Pos, "expected '{', got %s", p.peek())
		d.End = p.end()
		return d // a declaration whose body is not written yet
	}
	saved := p.enterBoundary("func")
	d.Body = p.block()
	p.leaveBoundary(saved)
	d.End = p.end()
	return d
}

func (p *parser) typeDecl() *ast.TypeDecl {
	t := p.expect("type")
	name := p.expectIdent("a type name")
	d := &ast.TypeDecl{Pos: pos(t), Name: name.Text, NamePos: pos(name)}
	if p.declarations() && p.at(".") {
		p.next()
		d.Owner, d.OwnerPos = d.Name, d.NamePos
		m := p.expectIdent("a type name")
		d.Name, d.NamePos = m.Text, pos(m)
	}
	if p.at("[") {
		d.TypeParams = p.typeParams()
	}
	if p.declarations() && !p.at("=") {
		d.Sig = p.since(t)
		d.End = p.end()
		return d // an opaque type
	}
	if p.recovering && !p.at("=") {
		p.soft(p.peek().Pos, "expected '=', got %s", p.peek())
		d.Sig = p.since(t)
		d.Body = &ast.BadType{Pos: p.end(), End: p.end()}
		d.End = p.end()
		return d
	}
	p.expect("=")
	if p.at("{") {
		rt := p.recordType(true)
		if p.at("|") && p.mode != Spec && slices.ContainsFunc(rt.Fields, func(f ast.FieldType) bool { return f.Default != nil }) {
			p.fail(p.peek().Pos, "a field default belongs in the record body of a type declaration, not in a union member")
		}
		d.Body = p.unionFrom(rt)
	} else {
		d.Body = p.parseType()
	}
	d.Sig = p.since(t)
	d.End = p.end()
	return d
}

// since is the source from t to the end of the last token read.
func (p *parser) since(t token.Token) string { return p.src[t.Off:p.toks[p.i-1].End] }

// doc is the doc comment of the declaration that starts at t (section 2): the run of line
// comments, each alone on its line, on the lines directly above t.
func (p *parser) doc(t token.Token) string {
	for p.ci < len(p.comments) && p.comments[p.ci].End <= t.Off {
		p.ci++
	}
	var lines []string
	end := t.Off
	for k := p.ci - 1; k >= 0; k-- {
		c := p.comments[k]
		if !strings.HasPrefix(c.Text, "//") || strings.HasPrefix(c.Text, "//!") || !oneLineBreak(p.src[c.End:end]) || !lineStart(p.src, c.Off) {
			break
		}
		line := strings.TrimRight(c.Text[2:], " \t\r")
		lines = append(lines, strings.TrimPrefix(line, " "))
		end = c.Off
	}
	slices.Reverse(lines)
	return strings.Join(lines, "\n")
}

// moduleDoc is the module doc (section 2): the //! lines, which must all come before first, the
// module's first token, alone on their lines and on consecutive lines.
func (p *parser) moduleDoc(first token.Token) string {
	var lines []string
	last := -1 // the index in p.comments of the previous //! line
	for k, c := range p.comments {
		if !strings.HasPrefix(c.Text, "//!") {
			continue
		}
		switch {
		case c.Off >= first.Off:
			p.soft(c.Pos, "a //! module doc line must come before the first import or declaration")
		case !lineStart(p.src, c.Off):
			p.soft(c.Pos, "a //! module doc line must be alone on its line")
		case last >= 0 && (k != last+1 || !oneLineBreak(p.src[p.comments[last].End:c.Off])):
			p.soft(c.Pos, "the //! module doc lines must be consecutive, with no blank line or other comment between")
		}
		line := strings.TrimRight(c.Text[3:], " \t\r")
		lines = append(lines, strings.TrimPrefix(line, " "))
		last = k
	}
	return strings.Join(lines, "\n")
}

// oneLineBreak reports whether s is blank apart from exactly one line break.
func oneLineBreak(s string) bool {
	return strings.Trim(s, " \t\r\n") == "" && strings.Count(s, "\n") == 1
}

// lineStart reports whether only spaces and tabs come before off on its line.
func lineStart(src string, off int) bool {
	i := strings.LastIndexByte(src[:off], '\n')
	return strings.Trim(src[i+1:off], " \t") == ""
}

func (p *parser) typeParams() []ast.TypeParam {
	p.expect("[")
	var params []ast.TypeParam
	p.list("]", func() {
		n := p.expectIdent("a type parameter")
		tp := ast.TypeParam{Pos: pos(n), Name: n.Text}
		if p.at(":") {
			p.next()
			tp.Bound = p.parseType()
		}
		tp.End = p.end()
		params = append(params, tp)
	})
	return params
}

func (p *parser) signature() ([]ast.Param, ast.Type) {
	p.expect("(")
	var params []ast.Param
	p.list(")", func() {
		if p.mode == Spec && p.at("...") {
			p.next()
			return
		}
		n := p.expectIdent("a parameter name")
		p.expect(":")
		prm := ast.Param{Pos: pos(n), Name: n.Text, Type: p.parseType()}
		if p.at("=") {
			p.next()
			p.inDefault++
			prm.Default = p.expression()
			p.inDefault--
		}
		prm.End = p.end()
		params = append(params, prm)
	})
	var ret ast.Type
	if p.at(":") {
		p.next()
		ret = p.parseType()
	}
	return params, ret
}

// ---- types

func (p *parser) parseType() ast.Type {
	p.nest()
	defer p.unnest()
	return p.unionFrom(p.typeTerm())
}

// unionFrom parses the rest of a union type whose first member has been read, if there is one.
func (p *parser) unionFrom(first ast.Type) ast.Type {
	if !p.at("|") {
		return first
	}
	u := &ast.UnionType{Pos: first.Position(), Members: []ast.Type{first}}
	for p.at("|") {
		p.next()
		u.Members = append(u.Members, p.typeTerm())
	}
	u.End = p.end()
	return u
}

func (p *parser) typeTerm() ast.Type {
	t := p.peek()
	switch {
	case t.Is("null") || t.Is("void"):
		p.next()
		return &ast.TypeName{Pos: pos(t), End: epos(t), Path: []string{t.Text}, PathPos: []ast.Pos{pos(t)}}
	case t.Is("{"):
		return p.recordType(false)
	case t.Is("func"):
		return p.funcType()
	case t.Is("("):
		p.next()
		p.mark(t, Role{})
		saved := p.enterList()
		inner := p.parseType()
		p.expect(")")
		p.leave(saved)
		return inner
	case t.Kind == token.Ident:
		p.next()
		tn := &ast.TypeName{Pos: pos(t), Path: []string{t.Text}, PathPos: []ast.Pos{pos(t)}}
		for p.at(".") {
			p.next()
			part := p.expectIdent("a type name")
			tn.Path = append(tn.Path, part.Text)
			tn.PathPos = append(tn.PathPos, pos(part))
		}
		if p.at("[") {
			p.next()
			p.list("]", func() { tn.Args = append(tn.Args, p.parseType()) })
		}
		tn.End = p.end()
		return tn
	}
	if p.recovering && t.Kind != token.Reserved {
		p.soft(t.Pos, "expected a type, got %s", t)
		return &ast.BadType{Pos: p.end(), End: p.end()}
	}
	p.failAt(t, "expected a type, got %s", t)
	return nil
}

// recordType parses a record type. Its fields take defaults only when it is the body of a type
// declaration (section 4, Records).
func (p *parser) recordType(defaults bool) *ast.RecordType {
	t := p.expect("{")
	rt := &ast.RecordType{Pos: pos(t)}
	p.list("}", func() {
		n := p.expectName("a field name")
		p.expect(":")
		f := ast.FieldType{Pos: pos(n), Name: n.Text, Type: p.parseType()}
		if p.at("=") {
			if !defaults && p.mode != Spec {
				p.fail(p.peek().Pos, "a field default belongs in the record body of a type declaration")
			}
			p.next()
			f.Default = p.fieldDefault()
		}
		f.End = p.end()
		rt.Fields = append(rt.Fields, f)
	})
	rt.End = p.end()
	return rt
}

// fieldDefault parses a field's default, which lies outside every boundary and every block, as
// a library's top-level initializer does (section 5), so return, try, break, continue, and
// defer have nothing to leave or attach to.
func (p *parser) fieldDefault() ast.Expr {
	saved := p.enterBoundary("default")
	inDefault, block := p.inDefault, p.curBlock
	p.inDefault, p.curBlock = 0, nil
	e := p.expression()
	p.inDefault, p.curBlock = inDefault, block
	p.leaveBoundary(saved)
	return e
}

func (p *parser) funcType() *ast.FuncType {
	t := p.expect("func")
	p.expect("(")
	ft := &ast.FuncType{Pos: pos(t)}
	p.list(")", func() {
		typ := p.parseType()
		prm := ast.FuncTypeParam{Type: typ}
		if p.at(":") {
			tn, ok := typ.(*ast.TypeName)
			if !ok || len(tn.Path) != 1 || len(tn.Args) != 0 {
				p.fail(p.peek().Pos, "expected ',' or ')' in a function type")
			}
			p.next()
			prm = ast.FuncTypeParam{Name: tn.Path[0], Type: p.parseType()}
		}
		ft.Params = append(ft.Params, prm)
	})
	p.expect(":")
	ft.Return = p.parseType()
	ft.End = p.end()
	return ft
}

// ---- blocks and statements

func (p *parser) block() *ast.Block {
	p.nest()
	defer p.unnest()
	t := p.expect("{")
	p.mark(t, Role{Block: true})
	b := &ast.Block{Pos: pos(t)}
	saved := p.enterBlock(b)
	p.blockStatements(b)
	p.expectClose()
	p.leave(saved)
	b.End = p.end()
	return b
}

// blockStatements parses a block's statements and lists the types it declares.
func (p *parser) blockStatements(b *ast.Block) {
	b.Stmts = p.statements()
	for _, s := range b.Stmts {
		if td, ok := s.(*ast.TypeDecl); ok {
			b.Types = append(b.Types, td)
		}
	}
}

func (p *parser) statements() []ast.Stmt {
	var stmts []ast.Stmt
	p.skipTerms()
	for !p.at("}") {
		if p.atKind(token.EOF) {
			if p.recovering {
				break // expectClose reports it
			}
			p.fail(p.peek().Pos, "expected '}', got end of file")
		}
		if len(stmts) > 0 {
			p.discardedHtml(stmts[len(stmts)-1])
		}
		start := p.i
		var s ast.Stmt
		if !p.attempt(func() { s = p.statement() }) || p.i == start {
			p.sync("}", false, start)
			s = p.badStmt(start)
		}
		stmts = append(stmts, s)
		if p.at("}") {
			break
		}
		p.terminate("}")
	}
	return stmts
}

func (p *parser) statement() ast.Stmt {
	t := p.peek()
	if t.Kind == token.Keyword {
		switch t.Text {
		case "var":
			return p.varDecl()
		case "type":
			return p.typeDecl()
		case "export", "import":
			p.fail(t.Pos, "'%s' is only allowed at the top level", t.Text)
		}
	}
	return p.expression()
}

// ---- expressions

func isAssignOp(s string) bool {
	switch s {
	case "=", "+=", "-=", "*=", "/=", "%=":
		return true
	}
	return false
}

func (p *parser) expression() ast.Expr {
	p.nest()
	defer p.unnest()
	left := p.pipe()
	t := p.peek()
	if t.Kind == token.Op && isAssignOp(t.Text) {
		p.next()
		target := p.assignTarget(left, t)
		right := p.pipe()
		end := p.end()
		if n := p.peek(); n.Kind == token.Op && isAssignOp(n.Text) {
			p.fail(n.Pos, "assignment cannot be chained")
		}
		return &ast.Assign{Pos: pos(t), End: end, Op: t.Text, Target: target, Value: right}
	}
	return left
}

func (p *parser) assignTarget(e ast.Expr, op token.Token) ast.Node {
	switch x := e.(type) {
	case *ast.Ident:
		return x
	case *ast.Field:
		if !x.Optional {
			return x
		}
	case *ast.Index:
		if !x.Optional {
			return x
		}
	case *ast.ArrayLit:
		if op.Text != "=" {
			p.fail(op.Pos, "a destructuring assignment allows only '='")
		}
		return p.arrayAssignTarget(x, op)
	}
	p.fail(op.Pos, "the left side of '%s' must be a variable, a field, an element, or an array destructuring", op.Text)
	return nil
}

func (p *parser) arrayAssignTarget(lit *ast.ArrayLit, op token.Token) *ast.AssignArray {
	t := &ast.AssignArray{Pos: lit.Pos, End: lit.End}
	if len(lit.Elems) == 0 {
		p.fail(op.Pos, "an array destructuring binds at least one position")
	}
	for i, el := range lit.Elems {
		if el.Spread {
			if i != len(lit.Elems)-1 {
				p.fail(op.Pos, "the rest must be the last position of a destructuring")
			}
			if !isPlace(el.Value) {
				p.fail(op.Pos, "the rest position of a destructuring must be an assignable place")
			}
			if lit.SpreadComma {
				p.fail(op.Pos, "a destructuring takes no comma after its rest position")
			}
			p.mark(token.Token{Off: lit.Off}, Role{}) // no comma after the rest
			t.Rest = el.Value
			continue
		}
		switch x := el.Value.(type) {
		case *ast.ArrayLit:
			t.Elems = append(t.Elems, p.arrayAssignTarget(x, op))
		default:
			if !isPlace(x) {
				p.fail(op.Pos, "the positions of a destructuring assignment must be assignable places")
			}
			t.Elems = append(t.Elems, x)
		}
	}
	return t
}

func isPlace(e ast.Expr) bool {
	switch x := e.(type) {
	case *ast.Ident:
		return true
	case *ast.Field:
		return !x.Optional
	case *ast.Index:
		return !x.Optional
	}
	return false
}

func (p *parser) pipe() ast.Expr {
	links := 0
	defer p.unlink(&links)
	left := p.pipeOperand()
	var stages []*ast.Call
	for p.at("|") {
		p.link(&links)
		bar := p.next()
		right := p.postfix()
		call, ok := right.(*ast.Call)
		if !ok || call.Optional || call.Pipe {
			p.fail(bar.Pos, "the right side of a pipe must be a call")
		}
		call.Args = append([]ast.Arg{{Pos: ast.Start(left), End: left.EndPosition(), Value: left}}, call.Args...)
		call.Pipe = true
		call.Assert = nil // assert(...) on a pipe's right is a use the resolver rejects
		call.Bar = pos(bar)
		stages = append(stages, call)
		left = call
	}
	if len(stages) > 1 {
		for _, c := range stages {
			c.Stage = true
		}
	}
	return left
}

func (p *parser) pipeOperand() ast.Expr {
	links := 0
	defer p.unlink(&links)
	left := p.equality()
	if p.at("??") {
		for p.at("??") {
			p.link(&links)
			t := p.next()
			b := &ast.Binary{Pos: pos(t), Op: "??", X: left, Y: p.equality()}
			b.End = p.end()
			left = b
		}
		if p.at("&&") || p.at("||") {
			p.fail(p.peek().Pos, "'??' cannot mix with '&&' or '||' without parentheses")
		}
		return left
	}
	if p.at("&&") || p.at("||") {
		left = p.andTail(left)
		for p.at("||") {
			p.link(&links)
			t := p.next()
			right := p.andTail(p.equality())
			left = &ast.Binary{Pos: pos(t), End: p.end(), Op: "||", X: left, Y: right}
		}
		if p.at("??") {
			p.fail(p.peek().Pos, "'??' cannot mix with '&&' or '||' without parentheses")
		}
	}
	return left
}

func (p *parser) andTail(left ast.Expr) ast.Expr {
	links := 0
	defer p.unlink(&links)
	for p.at("&&") {
		p.link(&links)
		t := p.next()
		b := &ast.Binary{Pos: pos(t), Op: "&&", X: left, Y: p.equality()}
		b.End = p.end()
		left = b
	}
	return left
}

func (p *parser) equality() ast.Expr { return p.binaryLeft(p.comparison, "==", "!=") }

// binaryLeft parses a left-associative level: operands from next, joined by any of ops. Each
// link of the chain counts toward the nesting limit, as a nested expression does.
func (p *parser) binaryLeft(next func() ast.Expr, ops ...string) ast.Expr {
	links := 0
	defer p.unlink(&links)
	left := next()
	for slices.ContainsFunc(ops, p.at) {
		p.link(&links)
		t := p.next()
		b := &ast.Binary{Pos: pos(t), Op: t.Text, X: left, Y: next()}
		b.End = p.end()
		left = b
	}
	return left
}

var kindNames = map[string]bool{
	"bool": true, "string": true, "Bytes": true, "Html": true, "Array": true, "Map": true, "Set": true,
	"int8": true, "int16": true, "int32": true, "int64": true,
	"uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"float32": true, "float64": true, "int": true, "byte": true,
}

func (p *parser) comparison() ast.Expr {
	links := 0
	defer p.unlink(&links)
	left := p.additive()
	for {
		t := p.peek()
		switch {
		case t.Is("<") || t.Is("<=") || t.Is(">") || t.Is(">="):
			p.next()
			b := &ast.Binary{Pos: pos(t), Op: t.Text, X: left, Y: p.additive()}
			b.End = p.end()
			left = b
		case t.Is("is"):
			p.next()
			is := &ast.Is{Pos: pos(t), X: left}
			is.Kind, is.KindPos, is.Path = p.isTarget()
			is.End = p.end()
			left = is
		default:
			return left
		}
		p.link(&links)
	}
}

func (p *parser) isTarget() (kind string, kindPos ast.Pos, path ast.Expr) {
	t := p.peek()
	if t.Is("null") || t.Is("void") || t.Is("func") {
		p.next()
		return t.Text, pos(t), nil
	}
	if t.Kind != token.Ident {
		p.fail(t.Pos, "expected a type name after 'is', got %s", t)
	}
	p.next()
	var e ast.Expr = &ast.Ident{Pos: pos(t), End: epos(t), Name: t.Text}
	single := true
	links := 0
	defer p.unlink(&links)
	for p.at(".") {
		p.link(&links)
		d := p.next()
		name := p.expectIdent("a type name")
		e = &ast.Field{Pos: pos(d), End: epos(name), X: e, Name: name.Text, NamePos: pos(name)}
		single = false
	}
	if single && kindNames[t.Text] {
		return t.Text, pos(t), nil
	}
	return "", ast.Pos{}, e
}

func (p *parser) additive() ast.Expr { return p.binaryLeft(p.multiplicative, "+", "-") }

func (p *parser) multiplicative() ast.Expr { return p.binaryLeft(p.unary, "*", "/", "%") }

func (p *parser) unary() ast.Expr {
	t := p.peek()
	if t.Is("-") || t.Is("!") || t.Is("try") {
		if t.Is("try") && p.boundary == "" && p.inDefault == 0 && p.mode != Spec {
			p.fail(t.Pos, "'try' outside a function")
		}
		if t.Is("try") && p.boundary == "default" && p.mode != Spec {
			p.fail(t.Pos, "'try' in a field default")
		}
		if t.Is("try") && p.boundary == "script" && p.inDefault == 0 {
			p.topLevelOnly(t.Pos, "'try' outside a function")
		}
		p.nest()
		defer p.unnest()
		p.next()
		if p.prefix != nil && t.Kind == token.Op {
			p.prefix[t.Off] = true
		}
		u := &ast.Unary{Pos: pos(t), Op: t.Text, X: p.unary()}
		u.End = p.end()
		return u
	}
	return p.postfix()
}

func (p *parser) postfix() ast.Expr {
	links := 0
	defer p.unlink(&links)
	node := p.primary()
	_, isPath := node.(*ast.Ident)
	for {
		t := p.peek()
		if t.Kind != token.Op {
			return node
		}
		switch t.Text {
		case ".":
			p.next()
			if p.recovering && !p.atKind(token.Ident) && p.raw().Kind != token.Reserved && p.raw().Kind != token.Keyword {
				// A dot with no name after it, as while the name is typed: the field is empty.
				p.soft(p.raw().Pos, "expected a field name, got %s", p.raw())
				return &ast.Field{Pos: pos(t), End: epos(t), X: node, NamePos: epos(t)}
			}
			name := p.expectIdent("a field name")
			node = &ast.Field{Pos: pos(t), End: epos(name), X: node, Name: name.Text, NamePos: pos(name)}
			if isPath && p.at("{") && !p.header {
				return p.braceLit(node)
			}
		case "?.":
			p.next()
			isPath = false
			switch {
			case p.at("["):
				lb := p.peek()
				idx := &ast.Index{Pos: pos(lb), X: node, Index: p.index(), Optional: true}
				idx.End = p.end()
				node = idx
			case p.at("("):
				lp := p.peek()
				call := &ast.Call{Pos: pos(lp), Fn: node, Args: p.callArgs(), Optional: true}
				call.End = p.end()
				node = call
			default:
				if p.recovering && !p.atKind(token.Ident) && p.raw().Kind != token.Reserved && p.raw().Kind != token.Keyword {
					p.soft(p.raw().Pos, "expected a field name, got %s", p.raw())
					return &ast.Field{Pos: pos(t), End: epos(t), X: node, NamePos: epos(t), Optional: true}
				}
				name := p.expectIdent("a field name")
				node = &ast.Field{Pos: pos(t), End: epos(name), X: node, Name: name.Text, NamePos: pos(name), Optional: true}
			}
		case "[":
			isPath = false
			idx := &ast.Index{Pos: pos(t), X: node, Index: p.index()}
			idx.End = p.end()
			node = idx
		case "(":
			isPath = false
			call := &ast.Call{Pos: pos(t), Fn: node, Args: p.callArgs()}
			call.End = p.end()
			if id, ok := call.Fn.(*ast.Ident); ok && id.Name == "assert" {
				call.Assert = p.assertInfo(call)
			}
			node = call
		default:
			return node
		}
		p.link(&links)
	}
}

// index parses [ expression ].
func (p *parser) index() ast.Expr {
	p.mark(p.expect("["), Role{})
	saved := p.enterList()
	idx := p.expression()
	p.expect("]")
	p.leave(saved)
	return idx
}

func (p *parser) callArgs() []ast.Arg {
	p.expect("(")
	var args []ast.Arg
	named := false
	p.list(")", func() {
		if p.mode == Spec && p.at("...") && (p.peekAfter().Is(")") || p.peekAfter().Is(",")) {
			p.next()
			return
		}
		start := p.peek()
		e := p.expression()
		if id, ok := e.(*ast.Ident); ok && p.at(":") {
			p.next()
			a := ast.Arg{Pos: pos(start), Name: id.Name, Value: p.expression()}
			a.End = p.end()
			args = append(args, a)
			named = true
		} else {
			if named {
				p.fail(start.Pos, "a positional argument cannot follow a named argument")
			}
			args = append(args, ast.Arg{Pos: pos(start), End: p.end(), Value: e})
		}
	})
	return args
}

func (p *parser) primary() ast.Expr {
	t := p.peek()
	switch t.Kind {
	case token.Int:
		p.next()
		return &ast.IntLit{Pos: pos(t), End: epos(t), Value: t.Value}
	case token.Float:
		p.next()
		return &ast.FloatLit{Pos: pos(t), End: epos(t), Value: t.Float}
	case token.String:
		p.next()
		return &ast.StringLit{Pos: pos(t), End: epos(t), Value: t.Text}
	case token.Regex:
		p.next()
		re, err := value.NewRegex(t.Text, t.Flags)
		if err != nil {
			p.fail(t.Pos, "%s", err)
		}
		return &ast.RegexLit{Pos: pos(t), End: epos(t), Source: re.Source, Flags: re.Flags, Pattern: re}
	case token.Template:
		p.next()
		return p.template(t)
	case token.Html:
		p.next()
		return &ast.Html{Pos: pos(t), End: epos(t), Content: p.markup(t.Markup), Close: cpos(t.Markup.Close)}
	case token.Ident:
		p.next()
		id := &ast.Ident{Pos: pos(t), End: epos(t), Name: t.Text}
		if p.at("{") && !p.header {
			return p.braceLit(id)
		}
		return id
	case token.Keyword:
		switch t.Text {
		case "true", "false":
			p.next()
			return &ast.BoolLit{Pos: pos(t), End: epos(t), Value: t.Text == "true"}
		case "null":
			p.next()
			return &ast.NullLit{Pos: pos(t), End: epos(t)}
		case "void":
			p.next()
			return &ast.VoidLit{Pos: pos(t), End: epos(t)}
		case "func":
			p.next()
			if p.atKind(token.Ident) {
				p.fail(p.peek().Pos, "a named function is only allowed at the top level; bind it with var")
			}
			params, ret := p.signature()
			saved := p.enterBoundary("func")
			body := p.block()
			p.leaveBoundary(saved)
			return &ast.FuncLit{Pos: pos(t), End: p.end(), Params: params, Return: ret, Body: body}
		case "if":
			return p.ifExpr()
		case "select":
			return p.selectExpr()
		case "while":
			p.next()
			cond := p.headerExpr("while")
			w := &ast.While{Pos: pos(t), Cond: cond, Body: p.loopBody()}
			w.End = p.end()
			return w
		case "for":
			return p.forExpr()
		case "go", "defer":
			p.next()
			enclosing := p.curBlock
			if enclosing == nil && t.Text == "defer" && p.boundary == "script" {
				p.topLevelOnly(t.Pos, "defer needs an enclosing block")
			} else if enclosing == nil && t.Text == "defer" && p.mode != Spec {
				p.fail(t.Pos, "defer needs an enclosing block")
			}
			saved := p.enterBoundary(t.Text)
			body := p.block()
			p.leaveBoundary(saved)
			if t.Text == "go" {
				return &ast.Go{Pos: pos(t), End: p.end(), Body: body}
			}
			if enclosing != nil {
				enclosing.HasDefer = true
			}
			return &ast.Defer{Pos: pos(t), End: p.end(), Body: body}
		case "return":
			if p.boundary == "" && p.mode != Spec {
				p.fail(t.Pos, "'return' outside a function")
			}
			if p.boundary == "default" && p.mode != Spec {
				p.fail(t.Pos, "'return' in a field default")
			}
			if p.boundary == "script" {
				p.topLevelOnly(t.Pos, "'return' outside a function")
			}
			p.next()
			r := p.raw()
			if r.Kind == token.Newline || r.Kind == token.EOF ||
				r.Is(";") || r.Is("}") || r.Is(")") || r.Is("]") || r.Is(",") {
				return &ast.Return{Pos: pos(t), End: epos(t)}
			}
			if p.boundary == "go" || p.boundary == "defer" {
				p.fail(t.Pos, "'return' takes no value in a %s block", p.boundary)
			}
			ret := &ast.Return{Pos: pos(t), Value: p.expression()}
			ret.End = p.end()
			return ret
		case "break", "continue":
			if p.loops == 0 && p.mode != Spec {
				p.fail(t.Pos, "'%s' outside a loop", t.Text)
			}
			p.next()
			if t.Text == "break" {
				return &ast.Break{Pos: pos(t), End: epos(t)}
			}
			return &ast.Continue{Pos: pos(t), End: epos(t)}
		}
	case token.Op:
		switch t.Text {
		case "{":
			return p.lambda()
		case "[":
			return p.arrayLit()
		case "(":
			p.next()
			p.mark(t, Role{})
			saved := p.enterList()
			e := p.expression()
			p.expect(")")
			p.leave(saved)
			// the parentheses end a postfix chain, so a ?. inside skips nothing after them
			switch x := e.(type) {
			case *ast.Field:
				x.Grouped = true
			case *ast.Index:
				x.Grouped = true
			case *ast.Call:
				x.Grouped = true
			}
			return e
		case "...":
			if p.mode == Spec {
				p.next()
				return &ast.Placeholder{Pos: pos(t), End: epos(t)}
			}
		}
	}
	if p.recovering && t.Kind != token.Reserved {
		// What is here cannot start an expression. A token that ends something, a line break, a
		// closer, a comma, or the end of the file, is left for the enclosing construct and the
		// expression is empty; anything else is skipped as the expression.
		p.soft(t.Pos, "unexpected %s", t)
		if t.Kind == token.Newline || t.Kind == token.EOF || t.Kind == token.Keyword ||
			t.Is(")") || t.Is("]") || t.Is("}") || t.Is(",") || t.Is(";") || t.Is(":") {
			return &ast.BadExpr{Pos: p.end(), End: p.end()}
		}
		p.next()
		return &ast.BadExpr{Pos: pos(t), End: epos(t)}
	}
	p.unexpected(t)
	return nil
}

func (p *parser) template(t token.Token) *ast.Template {
	n := &ast.Template{Pos: pos(t), End: epos(t)}
	for _, part := range t.Parts {
		if part.Tokens == nil {
			n.Parts = append(n.Parts, ast.TemplatePart{Pos: cpos(part.Pos), End: cpos(part.EndPos), Text: part.Text})
			continue
		}
		e := p.subExpression(part.Tokens, "an interpolation")
		if bad, ok := e.(*ast.BadExpr); ok && bad.Pos == (ast.Pos{}) {
			bad.Pos, bad.End = cpos(part.Pos), cpos(part.EndPos)
		}
		n.Parts = append(n.Parts, ast.TemplatePart{Pos: cpos(part.Pos), End: cpos(part.EndPos), Expr: e})
	}
	return n
}

// markup converts the content the lexer read, parsing each interpolation on its own tokens as
// template does for a string.
func (p *parser) markup(m *token.Markup) []ast.HtmlNode {
	p.nest()
	defer p.unnest()
	var out []ast.HtmlNode
	for _, n := range m.Nodes {
		switch n.Kind {
		case token.MarkupText:
			out = append(out, ast.HtmlNode{Pos: cpos(n.Pos), End: cpos(n.EndPos), Text: n.Text})
		case token.MarkupExpr:
			e := p.subExpression(n.Tokens, "markup")
			if bad, ok := e.(*ast.BadExpr); ok && bad.Pos == (ast.Pos{}) {
				bad.Pos, bad.End = cpos(n.Pos), cpos(n.EndPos)
			}
			out = append(out, ast.HtmlNode{Pos: cpos(n.Pos), End: cpos(n.EndPos), Expr: e})
		case token.MarkupElement:
			el := &ast.HtmlElement{Pos: cpos(n.Pos), End: cpos(n.EndPos), Name: n.Name, TextOnly: n.TextOnly}
			el.NamePos = ast.Pos{Line: el.Pos.Line, Col: el.Pos.Col + 1}
			if n.Children != nil {
				el.Close = cpos(n.Children.Close)
			}
			for _, a := range n.Attrs {
				attr := ast.HtmlAttr{Pos: cpos(a.Pos), End: cpos(a.EndPos), Name: a.Name, Text: a.Text}
				switch a.Kind {
				case token.AttrExpr:
					attr.Expr = p.subExpression(a.Tokens, "an attribute")
				case token.AttrSpread:
					attr.Spread = true
					attr.Expr = p.subExpression(a.Tokens, "a spread")
				}
				if bad, ok := attr.Expr.(*ast.BadExpr); ok && bad.Pos == (ast.Pos{}) {
					bad.Pos, bad.End = cpos(a.Pos), cpos(a.EndPos)
				}
				el.Attrs = append(el.Attrs, attr)
			}
			el.Content = p.markup(n.Children)
			out = append(out, ast.HtmlNode{Pos: el.Pos, End: el.End, Element: el})
		}
	}
	return out
}

// subExpression parses one expression from toks, an interpolation's tokens ending with EOF. It is
// parsed by this parser on its own tokens, so it keeps the source, the nesting depth, and
// everything else of the enclosing expression.
func (p *parser) subExpression(toks []token.Token, where string) (e ast.Expr) {
	saved, i, nl, header := p.toks, p.i, p.nl, p.header
	p.toks, p.i, p.nl, p.header = toks, 0, 1, false
	defer func() { p.toks, p.i, p.nl, p.header = saved, i, nl, header }()
	if len(toks) == 0 {
		return &ast.BadExpr{} // an interpolation the recovering lexer could not read
	}
	if !p.attempt(func() { e = p.expression() }) {
		return &ast.BadExpr{Pos: pos(toks[0]), End: epos(toks[len(toks)-1])}
	}
	if !p.atKind(token.EOF) {
		if p.soft(p.peek().Pos, "unexpected %s in %s", p.peek(), where) {
			return e
		}
	}
	return e
}

// discardedHtml fails at an HTML expression that is a statement whose value is discarded
// (section 5, HTML expressions).
func (p *parser) discardedHtml(s ast.Stmt) {
	if h, ok := s.(*ast.Html); ok {
		p.soft(token.Pos{Line: h.Pos.Line, Col: h.Pos.Col}, "this HTML expression's value is discarded; put siblings in a fragment <>...</>")
	}
}

// headerExpr parses the condition of an if or while, or the iterable of a for: an expression in
// which a brace after an identifier path ends the header rather than starting a literal.
func (p *parser) headerExpr(what string) ast.Expr {
	saved := p.header
	p.header = true
	e := p.expression()
	p.header = saved
	if !p.at("{") {
		p.fail(p.peek().Pos, "expected '{' after the %s header, got %s", what, p.peek())
	}
	if p.toks[p.i-1].Kind == token.Newline {
		// Inside parentheses the parser skips the line break, but the { still belongs on the
		// header's line (section 5, Blocks).
		p.fail(p.peek().Pos, "expected '{' after the %s header on its line, got a line break", what)
	}
	if p.entryFollows() {
		p.fail(p.peek().Pos, "a record or map literal in the %s header must be parenthesized", what)
	}
	return e
}

// entryFollows reports whether the brace at the cursor is followed by an entry, `name:` or
// `"key":`, which no block can begin with: the brace opens a record or map literal that the
// header stopped before.
func (p *parser) entryFollows() bool {
	j := p.i + 1
	for j < len(p.toks) && p.toks[j].Kind == token.Newline {
		j++
	}
	if j >= len(p.toks) {
		return false
	}
	switch p.toks[j].Kind {
	case token.Ident, token.String, token.Int, token.Float, token.Reserved:
	default:
		return false
	}
	for j++; j < len(p.toks) && p.toks[j].Kind == token.Newline; j++ {
	}
	return j < len(p.toks) && p.toks[j].Is(":")
}

func (p *parser) loopBody() *ast.Block {
	p.loops++
	b := p.block()
	p.loops--
	return b
}

func (p *parser) ifExpr() *ast.If {
	p.nest() // an else-if chain nests each if in the one before
	defer p.unnest()
	t := p.expect("if")
	cond := p.headerExpr("if")
	n := &ast.If{Pos: pos(t), Cond: cond, Then: p.block()}
	if p.at("else") {
		p.next()
		if p.at("if") {
			n.Else = p.ifExpr()
		} else {
			n.Else = p.block()
		}
	}
	n.End = p.end()
	return n
}

// selectHeads are the five blocking calls a select clause may wait on, by namespace and
// function; no declaration may shadow a prelude name, so the path is the function.
var selectHeads = map[string]map[string]bool{
	"Channel":   {"receive": true, "send": true},
	"Context":   {"wait": true},
	"WaitGroup": {"wait": true},
	"Time":      {"sleep": true},
}

const selectHeadNames = "Channel.receive, Channel.send, Context.wait, WaitGroup.wait or Time.sleep"

// selectExpr parses select { clauses }: each clause a head, one of the five blocking calls
// written and not performed, with an optional var x = binding and a block, or else and a block,
// written last (section 5). Newlines separate clauses, as they do statements.
func (p *parser) selectExpr() *ast.Select {
	t := p.expect("select")
	s := &ast.Select{Pos: pos(t)}
	open := p.expect("{")
	p.mark(open, Role{Block: true})
	saved := p.enterBlock(p.curBlock)
	heads, hasElse := 0, false
	p.skipClauseBreaks()
	for !p.at("}") {
		if p.atKind(token.EOF) {
			if p.recovering {
				break
			}
			p.fail(p.peek().Pos, "expected '}', got end of file")
		}
		begin := p.i
		var c *ast.SelectClause
		if !p.attempt(func() { c = p.selectClause(&heads, &hasElse) }) || p.i == begin {
			p.sync("}", false, begin)
			p.skipClauseBreaks()
			continue
		}
		s.Clauses = append(s.Clauses, c)
		if p.at("}") {
			break
		}
		// A line that starts with else continues the line before it (section 2), so no
		// terminator separates a clause's block from an else clause.
		if !p.at("else") && p.skipClauseBreaks() == 0 {
			if p.raw().Kind == token.Reserved {
				p.failAt(p.raw(), "")
			}
			p.soft(p.raw().Pos, "expected a line break before %s: a select has one clause per line", p.raw())
		}
		p.skipClauseBreaks()
	}
	p.expectClose()
	p.leave(saved)
	s.End = p.end()
	if heads == 0 {
		p.fail(open.Pos, "a select needs at least one clause with a head: %s", selectHeadNames)
	}
	return s
}

// selectClause parses one clause of a select, counting the heads and the else.
func (p *parser) selectClause(heads *int, hasElse *bool) *ast.SelectClause {
	start := p.peek()
	c := &ast.SelectClause{Pos: pos(start)}
	if p.at("var") {
		p.next()
		n := p.expectIdent("a name for the clause's result")
		c.Target = name(n)
		p.expect("=")
	}
	if p.at("else") {
		if c.Target != nil {
			p.fail(start.Pos, "an else clause binds nothing; it runs when no head can proceed")
		}
		if *hasElse {
			p.fail(start.Pos, "a select has at most one else clause")
		}
		*hasElse = true
		p.next()
		c.Body = p.block()
	} else {
		if *hasElse {
			p.fail(start.Pos, "the else clause must be the last clause of a select")
		}
		c.Head = p.selectHead(p.headerExpr("select clause"))
		*heads++
		c.Body = p.block()
	}
	c.End = p.end()
	return c
}

// skipClauseBreaks skips the line breaks between select clauses and returns how many it
// skipped. A ; there is an error, since a select has one clause per line (section 5).
func (p *parser) skipClauseBreaks() int {
	n := 0
	for ; p.raw().Kind == token.Newline; n++ {
		p.i++
	}
	if p.raw().Is(";") {
		p.soft(p.raw().Pos, "a select has one clause per line, so no ';' separates its clauses")
		p.i++
		return n + 1
	}
	return n
}

// selectHead checks that a clause's head is one of the five calls, written plainly.
func (p *parser) selectHead(e ast.Expr) *ast.Call {
	if call, ok := e.(*ast.Call); ok && !call.Optional && !call.Pipe {
		if f, ok := call.Fn.(*ast.Field); ok && !f.Optional {
			if ns, ok := f.X.(*ast.Ident); ok && selectHeads[ns.Name][f.Name] {
				return call
			}
		}
	}
	at := e.Position()
	p.fail(token.Pos{Line: at.Line, Col: at.Col}, "a select clause's head must be a call of %s", selectHeadNames)
	return nil
}

func (p *parser) forExpr() *ast.For {
	t := p.expect("for")
	var target ast.Target
	if p.at("{") || p.at("[") {
		target = p.destructure()
	} else {
		n := p.expectIdent("a loop variable")
		target = name(n)
	}
	p.expect("in")
	iter := p.headerExpr("for")
	fr := &ast.For{Pos: pos(t), Target: target, Iter: iter, Body: p.loopBody()}
	fr.End = p.end()
	return fr
}

// ---- destructuring

func (p *parser) destructure() ast.Target {
	p.nest()
	defer p.unnest()
	if p.at("[") {
		return p.arrayDestructure()
	}
	t := p.expect("{")
	rt := &ast.RecordTarget{Pos: pos(t)}
	p.list("}", func() {
		w := p.peek()
		field := p.expectName("a field name")
		f := ast.RecordTargetField{Pos: pos(w), Name: field.Text}
		if p.at(":") {
			p.next()
			f.Target = p.target()
		} else {
			if w.Kind != token.Ident {
				p.fail(w.Pos, "'%s' is %s and cannot be bound: write { %s: name }", w.Text, wordKind(w), w.Text)
			}
			f.Target = name(w)
		}
		f.End = p.end()
		rt.Fields = append(rt.Fields, f)
	})
	rt.End = p.end()
	if len(rt.Fields) == 0 {
		p.fail(t.Pos, "a record destructuring binds at least one field")
	}
	return rt
}

func (p *parser) arrayDestructure() *ast.ArrayTarget {
	t := p.expect("[")
	at := &ast.ArrayTarget{Pos: pos(t)}
	p.list("]", func() {
		if p.at("...") {
			p.next()
			n := p.expectIdent("a rest name")
			at.Rest = name(n)
			p.mark(t, Role{}) // no comma after the rest
			if !p.at("]") {
				p.fail(p.peek().Pos, "the rest must be the last position of a destructuring")
			}
			return
		}
		at.Elems = append(at.Elems, p.target())
	})
	at.End = p.end()
	if len(at.Elems) == 0 && at.Rest == nil {
		p.fail(t.Pos, "an array destructuring binds at least one position")
	}
	return at
}

func (p *parser) target() ast.Target {
	if p.at("{") || p.at("[") {
		return p.destructure()
	}
	n := p.expectIdent("a name or a destructuring")
	return name(n)
}

// ---- literals

// braceLit parses the entries of a record or map literal after its head.
func (p *parser) braceLit(head ast.Expr) ast.Expr {
	t := p.expect("{")
	id, _ := head.(*ast.Ident)
	isMap := id != nil && id.Name == "Map"
	var rec []ast.RecordEntry
	var entries []ast.MapEntry
	p.list("}", func() {
		w := p.peek()
		after := p.peekAfter()
		switch {
		case w.Is("..."):
			p.next()
			if p.mode == Spec && (p.at("}") || p.at(",")) {
				break
			}
			if isMap {
				p.fail(w.Pos, "spread has no meaning in a map literal")
			}
			e := ast.RecordEntry{Pos: pos(w), Spread: p.expression()}
			e.End = p.end()
			rec = append(rec, e)
		case w.Kind == token.Keyword || w.Kind == token.Reserved:
			p.next()
			if !p.at(":") {
				p.fail(w.Pos, "'%s' is %s: as a key it needs ':'", w.Text, wordKind(w))
			}
			p.next()
			v := p.expression()
			if isMap {
				entries = append(entries, ast.MapEntry{Pos: pos(w), End: p.end(), Key: &ast.StringLit{Pos: pos(w), End: epos(w), Value: w.Text}, Value: v})
			} else {
				rec = append(rec, ast.RecordEntry{Pos: pos(w), End: p.end(), Name: w.Text, Value: v})
			}
		case w.Kind == token.Ident && (after.Is(":") || after.Is(",") || after.Is("}")):
			p.next()
			var v ast.Expr
			if p.at(":") {
				p.next()
				v = p.expression()
			} else {
				v = &ast.Ident{Pos: pos(w), End: epos(w), Name: w.Text}
			}
			if isMap {
				entries = append(entries, ast.MapEntry{Pos: pos(w), End: p.end(), Key: &ast.StringLit{Pos: pos(w), End: epos(w), Value: w.Text}, Value: v})
			} else {
				rec = append(rec, ast.RecordEntry{Pos: pos(w), End: p.end(), Name: w.Text, Value: v})
			}
		default:
			if !isMap {
				if p.recovering && w.Kind == token.Ident {
					// A name with no colon or comma after it yet, as while an entry is typed:
					// the shorthand entry, and the list reports what follows.
					p.next()
					rec = append(rec, ast.RecordEntry{Pos: pos(w), End: epos(w), Name: w.Text, Value: &ast.Ident{Pos: pos(w), End: epos(w), Name: w.Text}})
					return
				}
				p.fail(w.Pos, "a record field must be a name")
			}
			k := p.expression()
			p.expect(":")
			e := ast.MapEntry{Pos: pos(w), Key: k, Value: p.expression()}
			e.End = p.end()
			entries = append(entries, e)
		}
	})
	p.mark(t, Role{Typed: true, Comma: true})
	if isMap {
		return &ast.MapLit{Pos: pos(t), End: p.end(), Entries: entries}
	}
	return &ast.RecordLit{Pos: head.Position(), End: p.end(), Head: head, Entries: rec}
}

func (p *parser) arrayLit() *ast.ArrayLit {
	t := p.expect("[")
	lit := &ast.ArrayLit{Pos: pos(t), Off: t.Off}
	p.list("]", func() {
		lit.SpreadComma = false
		if !p.at("...") {
			lit.Elems = append(lit.Elems, ast.ArrayElem{Value: p.expression()})
			return
		}
		p.next()
		if p.mode == Spec && (p.at("]") || p.at(",")) {
			return
		}
		lit.Elems = append(lit.Elems, ast.ArrayElem{Value: p.expression(), Spread: true})
		lit.SpreadComma = p.at(",")
	})
	lit.End = p.end()
	return lit
}

func (p *parser) lambda() *ast.Lambda {
	t := p.expect("{")
	p.mark(t, Role{Block: true})
	b := &ast.Block{Pos: pos(t)}
	savedN := p.enterBlock(b)
	savedB := p.enterBoundary("lambda")
	p.blockStatements(b)
	p.leaveBoundary(savedB)
	p.expectClose()
	p.leave(savedN)
	b.End = p.end()
	return &ast.Lambda{Pos: pos(t), End: b.End, Body: b}
}

// ---- assert

// assertInfo records what a call of assert reports when it fails (section 8, assert): the
// condition, and the parts of it whose values the report shows.
func (p *parser) assertInfo(call *ast.Call) *ast.Assert {
	a := &ast.Assert{X: -1, Y: -1, Index: map[ast.Expr]int{}}
	for i, arg := range call.Args {
		if i == 0 && arg.Name == "" || arg.Name == "cond" {
			a.Cond = arg.Value
			break
		}
	}
	if a.Cond == nil {
		return a
	}
	a.Text = p.text(a.Cond)
	add := func(e ast.Expr, shown bool) int {
		a.Index[e] = len(a.Parts)
		a.Parts = append(a.Parts, ast.AssertPart{Expr: e, Text: p.text(e), Shown: shown})
		return len(a.Parts) - 1
	}
	var part, inner func(e ast.Expr)
	// part adds e when it is a name, a field access, an index, or a call, and the parts inside it.
	part = func(e ast.Expr) {
		switch e.(type) {
		case *ast.Ident, *ast.Field, *ast.Index, *ast.Call:
			if id, ok := e.(*ast.Ident); !ok || id.Name != "_" {
				add(e, true)
			}
		}
		inner(e)
	}
	// inner adds the parts inside e, leaving out a callee and the object of an access.
	inner = func(e ast.Expr) {
		switch x := e.(type) {
		case *ast.Field:
			inner(x.X)
		case *ast.Index:
			inner(x.X)
			part(x.Index)
		case *ast.Call:
			if x.Pipe {
				part(x.Args[0].Value)
				inner(x.Fn)
				for _, arg := range x.Args[1:] {
					part(arg.Value)
				}
				return
			}
			inner(x.Fn)
			for _, arg := range x.Args {
				part(arg.Value)
			}
		case *ast.Unary:
			part(x.X)
		case *ast.Binary:
			part(x.X)
			part(x.Y)
		case *ast.Is:
			part(x.X)
		case *ast.Template:
			for _, tp := range x.Parts {
				if tp.Expr != nil {
					part(tp.Expr)
				}
			}
		case *ast.ArrayLit:
			for _, el := range x.Elems {
				part(el.Value)
			}
		case *ast.MapLit:
			for _, en := range x.Entries {
				part(en.Key)
				part(en.Value)
			}
		case *ast.RecordLit:
			for _, en := range x.Entries {
				if en.Spread != nil {
					part(en.Spread)
				} else {
					part(en.Value)
				}
			}
		}
	}
	switch x := a.Cond.(type) {
	case *ast.Binary:
		switch x.Op {
		case "==", "!=", "<", "<=", ">", ">=":
			a.Op = x.Op
			a.X = add(x.X, false)
			inner(x.X)
			a.Y = add(x.Y, false)
			inner(x.Y)
			return a
		}
	case *ast.Is:
		a.Op = "is"
		a.X = add(x.X, false)
		inner(x.X)
		return a
	}
	inner(a.Cond)
	return a
}

// text is the source of a node, from its first token to its last.
func (p *parser) text(n ast.Node) string {
	from, to := ast.Start(n), n.EndPosition()
	first := sort.Search(len(p.toks), func(i int) bool { return !before(p.toks[i].Pos, from) })
	last := sort.Search(len(p.toks), func(i int) bool { return !before(p.toks[i].Pos, to) }) - 1
	if first >= len(p.toks) || last < first {
		return ""
	}
	return p.src[p.toks[first].Off:p.toks[last].End]
}

func before(a token.Pos, b ast.Pos) bool {
	return a.Line < b.Line || a.Line == b.Line && a.Col < b.Col
}
