// Package parser builds the syntax tree of an Ant module. It implements sections 2, 5, 6, and 10
// of the specification with one token of lookahead and no feedback from the checker.
package parser

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/lexer"
	"github.com/ant-src/ant/internal/token"
	"github.com/ant-src/ant/value"
)

// Mode selects what a source may contain at the top level.
type Mode int

const (
	// Program is a module: imports followed by declarations.
	Program Mode = iota
	// Spec also allows top-level statements and "..." placeholders, for the specification's
	// own examples.
	Spec
	// Session also allows top-level statements, for a REPL.
	Session
	// Declarations allows a func without a body, for the prelude's signatures.
	Declarations
)

// Error is a syntax error. Incomplete is set when the input ended too early, which a REPL
// uses to keep reading.
type Error struct {
	Pos        ast.Pos
	Msg        string
	Incomplete bool
}

func (e *Error) Error() string { return fmt.Sprintf("%d:%d: %s", e.Pos.Line, e.Pos.Col, e.Msg) }

// Parse parses one module. path only labels the result.
func Parse(path, src string, mode Mode) (f *ast.File, err error) {
	toks, lerr := lexer.Lex(src)
	if lerr != nil {
		return nil, lexError(src, lerr)
	}
	p := &parser{mode: mode, toks: toks}
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

// ParseLayout parses one module, as Parse does in Program mode, and also returns its Layout.
func ParseLayout(path, src string) (f *ast.File, l *Layout, err error) {
	toks, comments, lerr := lexer.LexLayout(src)
	if lerr != nil {
		return nil, nil, lexError(src, lerr)
	}
	l = &Layout{Tokens: toks, Comments: comments, Roles: map[int]Role{}, Prefix: map[int]bool{}}
	p := &parser{mode: Program, toks: toks, roles: l.Roles, prefix: l.Prefix}
	defer p.recover(&err)
	return p.file(path), l, nil
}

// incompleteLex reports whether a lexer error means the source ended too early: an open
// template string or block comment, which may span lines, or a plain string or regular
// expression open on the last line. Those stay on one line, so a line break ends them in an
// error instead.
func incompleteLex(src string, le *lexer.Error) bool {
	if !strings.HasPrefix(le.Msg, "unterminated") {
		return false
	}
	if le.Msg == "unterminated string" || le.Msg == "unterminated regular expression" {
		return le.Pos.Line == strings.Count(src, "\n")+1
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
	p := &parser{mode: Spec, toks: toks}
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
	toks     []token.Token
	i        int
	nl       int          // > 0: newlines are ignored (inside a list)
	header   bool         // inside an if, while, or for header: Path { is not a literal
	boundary string       // "", "func", "lambda", "go", or "defer"
	loops    int          // loops entered inside the current boundary
	curBlock *ast.Block   // the innermost block whose statements are being parsed
	depth    int          // nesting of expressions, types, and blocks, bounded by maxNesting
	roles    map[int]Role // filled for ParseLayout, nil otherwise
	prefix   map[int]bool // likewise
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
	panic(&Error{Pos: ast.Pos{Line: pos.Line, Col: pos.Col}, Msg: fmt.Sprintf(format, args...), Incomplete: incomplete})
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
	if p.skipTerms() == 0 {
		p.failAt(t, "expected a newline or ';' before %s", t)
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
		item()
		if !p.at(close) {
			p.expect(",")
		}
	}
	p.expect(close)
	p.leave(saved)
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
	f := &ast.File{Path: path}
	p.skipTerms()
	for p.at("import") {
		f.Imports = append(f.Imports, p.importDecl())
		p.expectTerm()
		p.skipTerms()
	}
	for !p.atKind(token.EOF) {
		t := p.peek()
		switch {
		case t.Is("import"):
			p.fail(t.Pos, "imports must come before declarations")
		case t.Is("export"), t.Is("func"), t.Is("type"):
			d := p.declaration()
			f.Decls = append(f.Decls, d)
		case t.Is("var"):
			d := p.varDecl()
			f.Decls = append(f.Decls, d)
			if p.mode == Spec || p.mode == Session {
				f.Stmts = append(f.Stmts, d)
			}
		default:
			if p.mode != Spec && p.mode != Session {
				p.failAt(t, "expected a declaration, got %s", t)
			}
			f.Stmts = append(f.Stmts, p.statement())
		}
		p.expectTerm()
		p.skipTerms()
	}
	return f
}

func (p *parser) importDecl() *ast.Import {
	t := p.expect("import")
	pt := p.peek()
	if pt.Kind != token.String {
		p.fail(pt.Pos, "expected a module path string, got %s", pt)
	}
	p.next()
	imp := &ast.Import{Pos: pos(t), Path: pt.Text}
	if p.atKind(token.Ident) {
		imp.Namespace = p.next().Text
		return imp
	}
	p.expect("{")
	p.list("}", func() {
		name := p.expectIdent("an export name")
		item := ast.ImportName{Pos: pos(name), Name: name.Text, Alias: name.Text}
		if p.at(":") {
			p.next()
			item.Alias = p.expectIdent("a local name").Text
		}
		imp.Names = append(imp.Names, item)
	})
	return imp
}

func (p *parser) declaration() ast.Decl {
	export := false
	if p.at("export") {
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
		d.Target = p.destructure()
	case p.atKind(token.Ident):
		n := p.next()
		d.Target = &ast.NameTarget{Pos: pos(n), Name: n.Text}
		if p.mode == Declarations && p.at(".") {
			p.next()
			d.Owner = n.Text
			m := p.expectName("a constant name")
			d.Target = &ast.NameTarget{Pos: pos(m), Name: m.Text}
			p.expect(":")
			d.Type = p.parseType()
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
	p.expect("=")
	d.Init = p.expression()
	return d
}

func (p *parser) funcDecl() *ast.FuncDecl {
	t := p.expect("func")
	name := p.expectIdent("a function name")
	d := &ast.FuncDecl{Pos: pos(t), Name: name.Text}
	if p.at(".") {
		p.next()
		d.Owner = d.Name
		d.Name = p.expectName("a function name").Text
		if p.at(".") {
			p.fail(p.peek().Pos, "a function belongs to one type: func Type.name")
		}
	}
	if p.at("[") {
		d.TypeParams = p.typeParams()
	}
	d.Params, d.Return = p.signature()
	if p.mode == Declarations && !p.at("{") {
		return d
	}
	saved := p.enterBoundary("func")
	d.Body = p.block()
	p.leaveBoundary(saved)
	return d
}

func (p *parser) typeDecl() *ast.TypeDecl {
	t := p.expect("type")
	name := p.expectIdent("a type name")
	d := &ast.TypeDecl{Pos: pos(t), Name: name.Text}
	if p.mode == Declarations && p.at(".") {
		p.next()
		d.Owner = d.Name
		d.Name = p.expectIdent("a type name").Text
	}
	if p.at("[") {
		d.TypeParams = p.typeParams()
	}
	if p.mode == Declarations && !p.at("=") {
		return d // an opaque type
	}
	p.expect("=")
	d.Body = p.parseType()
	return d
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
			prm.Default = p.expression()
		}
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
	first := p.typeTerm()
	if !p.at("|") {
		return first
	}
	u := &ast.UnionType{Pos: first.Position(), Members: []ast.Type{first}}
	for p.at("|") {
		p.next()
		u.Members = append(u.Members, p.typeTerm())
	}
	return u
}

func (p *parser) typeTerm() ast.Type {
	t := p.peek()
	switch {
	case t.Is("null") || t.Is("void"):
		p.next()
		return &ast.TypeName{Pos: pos(t), Path: []string{t.Text}}
	case t.Is("{"):
		return p.recordType()
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
		tn := &ast.TypeName{Pos: pos(t), Path: []string{t.Text}}
		for p.at(".") {
			p.next()
			tn.Path = append(tn.Path, p.expectIdent("a type name").Text)
		}
		if p.at("[") {
			p.next()
			p.list("]", func() { tn.Args = append(tn.Args, p.parseType()) })
		}
		return tn
	}
	p.failAt(t, "expected a type, got %s", t)
	return nil
}

func (p *parser) recordType() *ast.RecordType {
	t := p.expect("{")
	rt := &ast.RecordType{Pos: pos(t)}
	p.list("}", func() {
		n := p.expectName("a field name")
		p.expect(":")
		rt.Fields = append(rt.Fields, ast.FieldType{Pos: pos(n), Name: n.Text, Type: p.parseType()})
	})
	return rt
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
	b.Stmts = p.statements()
	p.expect("}")
	p.leave(saved)
	return b
}

func (p *parser) statements() []ast.Stmt {
	var stmts []ast.Stmt
	p.skipTerms()
	for !p.at("}") {
		if p.atKind(token.EOF) {
			p.fail(p.peek().Pos, "expected '}', got end of file")
		}
		stmts = append(stmts, p.statement())
		if p.at("}") {
			break
		}
		p.expectTerm()
		p.skipTerms()
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
			p.fail(t.Pos, "type declarations are only allowed at the top level")
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
		if n := p.peek(); n.Kind == token.Op && isAssignOp(n.Text) {
			p.fail(n.Pos, "assignment cannot be chained")
		}
		return &ast.Assign{Pos: pos(t), Op: t.Text, Target: target, Value: right}
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
	t := &ast.AssignArray{Pos: lit.Pos}
	for i, el := range lit.Elems {
		if el.Spread {
			if i != len(lit.Elems)-1 {
				p.fail(op.Pos, "the rest must be the last position of a destructuring")
			}
			if !isPlace(el.Value) {
				p.fail(op.Pos, "the rest position of a destructuring must be an assignable place")
			}
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
	for p.at("|") {
		p.link(&links)
		bar := p.next()
		right := p.postfix()
		call, ok := right.(*ast.Call)
		if !ok || call.Optional || call.Pipe {
			p.fail(bar.Pos, "the right side of a pipe must be a call")
		}
		call.Args = append([]ast.Arg{{Pos: left.Position(), Value: left}}, call.Args...)
		call.Pipe = true
		left = call
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
			left = &ast.Binary{Pos: pos(t), Op: "??", X: left, Y: p.equality()}
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
			left = &ast.Binary{Pos: pos(t), Op: "||", X: left, Y: right}
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
		left = &ast.Binary{Pos: pos(t), Op: "&&", X: left, Y: p.equality()}
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
		left = &ast.Binary{Pos: pos(t), Op: t.Text, X: left, Y: next()}
	}
	return left
}

var kindNames = map[string]bool{
	"bool": true, "string": true, "Bytes": true, "Array": true, "Map": true, "Set": true,
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
			left = &ast.Binary{Pos: pos(t), Op: t.Text, X: left, Y: p.additive()}
		case t.Is("is"):
			p.next()
			is := &ast.Is{Pos: pos(t), X: left}
			is.Kind, is.Path = p.isTarget()
			left = is
		default:
			return left
		}
		p.link(&links)
	}
}

func (p *parser) isTarget() (string, ast.Expr) {
	t := p.peek()
	if t.Is("null") || t.Is("void") || t.Is("func") {
		p.next()
		return t.Text, nil
	}
	if t.Kind != token.Ident {
		p.fail(t.Pos, "expected a type name after 'is', got %s", t)
	}
	p.next()
	var e ast.Expr = &ast.Ident{Pos: pos(t), Name: t.Text}
	single := true
	links := 0
	defer p.unlink(&links)
	for p.at(".") {
		p.link(&links)
		d := p.next()
		name := p.expectIdent("a type name")
		e = &ast.Field{Pos: pos(d), X: e, Name: name.Text}
		single = false
	}
	if single && kindNames[t.Text] {
		return t.Text, nil
	}
	return "", e
}

func (p *parser) additive() ast.Expr { return p.binaryLeft(p.multiplicative, "+", "-") }

func (p *parser) multiplicative() ast.Expr { return p.binaryLeft(p.unary, "*", "/", "%") }

func (p *parser) unary() ast.Expr {
	t := p.peek()
	if t.Is("-") || t.Is("!") || t.Is("try") {
		p.nest()
		defer p.unnest()
		p.next()
		if p.prefix != nil && t.Kind == token.Op {
			p.prefix[t.Off] = true
		}
		return &ast.Unary{Pos: pos(t), Op: t.Text, X: p.unary()}
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
			name := p.expectIdent("a field name")
			node = &ast.Field{Pos: pos(t), X: node, Name: name.Text}
			if isPath && p.at("{") && !p.header {
				return p.braceLit(node)
			}
		case "?.":
			p.next()
			isPath = false
			switch {
			case p.at("["):
				lb := p.peek()
				node = &ast.Index{Pos: pos(lb), X: node, Index: p.index(), Optional: true}
			case p.at("("):
				lp := p.peek()
				node = &ast.Call{Pos: pos(lp), Fn: node, Args: p.callArgs(), Optional: true}
			default:
				name := p.expectIdent("a field name")
				node = &ast.Field{Pos: pos(t), X: node, Name: name.Text, Optional: true}
			}
		case "[":
			isPath = false
			node = &ast.Index{Pos: pos(t), X: node, Index: p.index()}
		case "(":
			isPath = false
			node = &ast.Call{Pos: pos(t), Fn: node, Args: p.callArgs()}
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
			args = append(args, ast.Arg{Pos: pos(start), Name: id.Name, Value: p.expression()})
			named = true
		} else {
			if named {
				p.fail(start.Pos, "a positional argument cannot follow a named argument")
			}
			args = append(args, ast.Arg{Pos: pos(start), Value: e})
		}
	})
	return args
}

func (p *parser) primary() ast.Expr {
	t := p.peek()
	switch t.Kind {
	case token.Int:
		p.next()
		return &ast.IntLit{Pos: pos(t), Value: t.Value}
	case token.Float:
		p.next()
		return &ast.FloatLit{Pos: pos(t), Value: t.Float}
	case token.String:
		p.next()
		return &ast.StringLit{Pos: pos(t), Value: t.Text}
	case token.Regex:
		p.next()
		re, err := value.NewRegex(t.Text, t.Flags)
		if err != nil {
			p.fail(t.Pos, "%s", err)
		}
		return &ast.RegexLit{Pos: pos(t), Source: re.Source, Flags: re.Flags, Pattern: re}
	case token.Template:
		p.next()
		return p.template(t)
	case token.Ident:
		p.next()
		id := &ast.Ident{Pos: pos(t), Name: t.Text}
		if p.at("{") && !p.header {
			return p.braceLit(id)
		}
		return id
	case token.Keyword:
		switch t.Text {
		case "true", "false":
			p.next()
			return &ast.BoolLit{Pos: pos(t), Value: t.Text == "true"}
		case "null":
			p.next()
			return &ast.NullLit{Pos: pos(t)}
		case "void":
			p.next()
			return &ast.VoidLit{Pos: pos(t)}
		case "func":
			p.next()
			if p.atKind(token.Ident) {
				p.fail(p.peek().Pos, "a named function is only allowed at the top level; bind it with var")
			}
			params, ret := p.signature()
			saved := p.enterBoundary("func")
			body := p.block()
			p.leaveBoundary(saved)
			return &ast.FuncLit{Pos: pos(t), Params: params, Return: ret, Body: body}
		case "if":
			return p.ifExpr()
		case "select":
			return p.selectExpr()
		case "while":
			p.next()
			cond := p.headerExpr("while")
			return &ast.While{Pos: pos(t), Cond: cond, Body: p.loopBody()}
		case "for":
			return p.forExpr()
		case "go", "defer":
			p.next()
			enclosing := p.curBlock
			if enclosing == nil && t.Text == "defer" && p.mode != Spec {
				p.fail(t.Pos, "defer needs an enclosing block")
			}
			saved := p.enterBoundary(t.Text)
			body := p.block()
			p.leaveBoundary(saved)
			if t.Text == "go" {
				return &ast.Go{Pos: pos(t), Body: body}
			}
			if enclosing != nil {
				enclosing.HasDefer = true
			}
			return &ast.Defer{Pos: pos(t), Body: body}
		case "return":
			if p.boundary == "" && p.mode != Spec {
				p.fail(t.Pos, "'return' outside a function")
			}
			p.next()
			r := p.raw()
			if r.Kind == token.Newline || r.Kind == token.EOF ||
				r.Is(";") || r.Is("}") || r.Is(")") || r.Is("]") || r.Is(",") {
				return &ast.Return{Pos: pos(t)}
			}
			if p.boundary == "go" || p.boundary == "defer" {
				p.fail(t.Pos, "'return' takes no value in a %s block", p.boundary)
			}
			return &ast.Return{Pos: pos(t), Value: p.expression()}
		case "break", "continue":
			if p.loops == 0 && p.mode != Spec {
				p.fail(t.Pos, "'%s' outside a loop", t.Text)
			}
			p.next()
			if t.Text == "break" {
				return &ast.Break{Pos: pos(t)}
			}
			return &ast.Continue{Pos: pos(t)}
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
				return &ast.Placeholder{Pos: pos(t)}
			}
		}
	}
	p.unexpected(t)
	return nil
}

func (p *parser) template(t token.Token) *ast.Template {
	n := &ast.Template{Pos: pos(t)}
	for _, part := range t.Parts {
		if part.Tokens == nil {
			n.Parts = append(n.Parts, ast.TemplatePart{Text: part.Text})
			continue
		}
		sub := &parser{mode: p.mode, toks: part.Tokens, nl: 1, boundary: p.boundary, loops: p.loops, curBlock: p.curBlock, roles: p.roles, prefix: p.prefix}
		e := sub.expression()
		if !sub.atKind(token.EOF) {
			sub.fail(sub.peek().Pos, "unexpected %s in a template expression", sub.peek())
		}
		n.Parts = append(n.Parts, ast.TemplatePart{Expr: e})
	}
	return n
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
	p.skipTerms()
	for !p.at("}") {
		if p.atKind(token.EOF) {
			p.fail(p.peek().Pos, "expected '}', got end of file")
		}
		start := p.peek()
		c := &ast.SelectClause{Pos: pos(start)}
		if p.at("var") {
			p.next()
			n := p.expectIdent("a name for the clause's result")
			c.Target = &ast.NameTarget{Pos: pos(n), Name: n.Text}
			p.expect("=")
		}
		if p.at("else") {
			if c.Target != nil {
				p.fail(start.Pos, "an else clause binds nothing; it runs when no head can proceed")
			}
			if hasElse {
				p.fail(start.Pos, "a select has at most one else clause")
			}
			hasElse = true
			p.next()
			c.Body = p.block()
		} else {
			if hasElse {
				p.fail(start.Pos, "the else clause must be the last clause of a select")
			}
			c.Head = p.selectHead(p.headerExpr("select clause"))
			heads++
			c.Body = p.block()
		}
		s.Clauses = append(s.Clauses, c)
		if p.at("}") {
			break
		}
		// A line that starts with else continues the line before it (section 2), so no
		// terminator separates a clause's block from an else clause.
		if !p.at("else") {
			p.expectTerm()
		}
		p.skipTerms()
	}
	p.expect("}")
	p.leave(saved)
	if heads == 0 {
		p.fail(open.Pos, "a select needs at least one clause with a head: %s", selectHeadNames)
	}
	return s
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
		target = &ast.NameTarget{Pos: pos(n), Name: n.Text}
	}
	p.expect("in")
	iter := p.headerExpr("for")
	return &ast.For{Pos: pos(t), Target: target, Iter: iter, Body: p.loopBody()}
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
		name := p.expectName("a field name")
		f := ast.RecordTargetField{Pos: pos(w), Name: name.Text}
		if p.at(":") {
			p.next()
			f.Target = p.target()
		} else {
			if w.Kind != token.Ident {
				p.fail(w.Pos, "'%s' is %s and cannot be bound: write { %s: name }", w.Text, wordKind(w), w.Text)
			}
			f.Target = &ast.NameTarget{Pos: pos(w), Name: w.Text}
		}
		rt.Fields = append(rt.Fields, f)
	})
	return rt
}

func (p *parser) arrayDestructure() *ast.ArrayTarget {
	t := p.expect("[")
	at := &ast.ArrayTarget{Pos: pos(t)}
	p.list("]", func() {
		if p.at("...") {
			p.next()
			n := p.expectIdent("a rest name")
			at.Rest = &ast.NameTarget{Pos: pos(n), Name: n.Text}
			p.mark(t, Role{}) // no comma after the rest
			if !p.at("]") {
				p.fail(p.peek().Pos, "the rest must be the last position of a destructuring")
			}
			return
		}
		at.Elems = append(at.Elems, p.target())
	})
	return at
}

func (p *parser) target() ast.Target {
	if p.at("{") || p.at("[") {
		return p.destructure()
	}
	n := p.expectIdent("a name or a destructuring")
	return &ast.NameTarget{Pos: pos(n), Name: n.Text}
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
			rec = append(rec, ast.RecordEntry{Pos: pos(w), Spread: p.expression()})
		case w.Kind == token.Keyword || w.Kind == token.Reserved:
			p.next()
			if !p.at(":") {
				p.fail(w.Pos, "'%s' is %s: as a key it needs ':'", w.Text, wordKind(w))
			}
			p.next()
			v := p.expression()
			if isMap {
				entries = append(entries, ast.MapEntry{Pos: pos(w), Key: &ast.StringLit{Pos: pos(w), Value: w.Text}, Value: v})
			} else {
				rec = append(rec, ast.RecordEntry{Pos: pos(w), Name: w.Text, Value: v})
			}
		case w.Kind == token.Ident && (after.Is(":") || after.Is(",") || after.Is("}")):
			p.next()
			var v ast.Expr
			if p.at(":") {
				p.next()
				v = p.expression()
			} else {
				v = &ast.Ident{Pos: pos(w), Name: w.Text}
			}
			if isMap {
				entries = append(entries, ast.MapEntry{Pos: pos(w), Key: &ast.StringLit{Pos: pos(w), Value: w.Text}, Value: v})
			} else {
				rec = append(rec, ast.RecordEntry{Pos: pos(w), Name: w.Text, Value: v})
			}
		default:
			if !isMap {
				p.fail(w.Pos, "a record field must be a name")
			}
			k := p.expression()
			p.expect(":")
			entries = append(entries, ast.MapEntry{Pos: pos(w), Key: k, Value: p.expression()})
		}
	})
	p.mark(t, Role{Typed: true, Comma: true})
	if isMap {
		return &ast.MapLit{Pos: pos(t), Entries: entries}
	}
	return &ast.RecordLit{Pos: head.Position(), Head: head, Entries: rec}
}

func (p *parser) arrayLit() *ast.ArrayLit {
	t := p.expect("[")
	lit := &ast.ArrayLit{Pos: pos(t)}
	p.list("]", func() {
		if !p.at("...") {
			lit.Elems = append(lit.Elems, ast.ArrayElem{Value: p.expression()})
			return
		}
		p.next()
		if p.mode == Spec && (p.at("]") || p.at(",")) {
			return
		}
		lit.Elems = append(lit.Elems, ast.ArrayElem{Value: p.expression(), Spread: true})
	})
	return lit
}

func (p *parser) lambda() *ast.Lambda {
	t := p.expect("{")
	p.mark(t, Role{Block: true})
	b := &ast.Block{Pos: pos(t)}
	savedN := p.enterBlock(b)
	savedB := p.enterBoundary("lambda")
	b.Stmts = p.statements()
	p.leaveBoundary(savedB)
	p.expect("}")
	p.leave(savedN)
	return &ast.Lambda{Pos: pos(t), Body: b}
}
