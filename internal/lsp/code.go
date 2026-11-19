package lsp

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/internal/prelude"
	"github.com/ant-src/ant/internal/token"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/parser"
)

// Code answers the requests of ant code (private/tools/code.md) with the server's own analysis,
// in process: a server whose workspace folder is the root, which counts columns in bytes, and
// whose open documents are the files as they are on disk, each read once.
type Code struct {
	s      *Server
	walked bool
}

// NewCode makes the analysis of ant code for the workspace under root.
func NewCode(root string) (*Code, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}
	s := &Server{log: log.New(io.Discard, "", 0), docs: map[string]*document{}, cache: map[string]*analysis{}, roots: []string{abs}}
	return &Code{s: s}, nil
}

// CodeError is a request ant code could not answer, with its code (code.md section 3).
type CodeError struct {
	Code       string
	Message    string
	Candidates []CodeLoc  // for ambiguous
	Errors     []CodeDiag // for check
	Skipped    []string   // for incomplete
	err        error      // the error of a file that could not be read
}

func (e *CodeError) Error() string { return e.Message }

func codeErr(code, format string, args ...any) *CodeError {
	return &CodeError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// CodeLoc is the location of a name or an expression: the file, as the analysis spells its path
// (absolute, or prelude.Path for the prelude), and the span, in lines and byte columns from 1.
type CodeLoc struct {
	Path       string
	Prelude    bool
	Start, End ast.Pos
	// Line is the source line the location starts on, without its line break.
	Line string
}

// CodeDiag is an error of a program a rename changes.
type CodeDiag struct {
	Path     string
	Pos, End ast.Pos
	Message  string
	Code     string
}

// loc makes the location of a span in the module at path.
func (a *analysis) codeLoc(path string, from, to ast.Pos) CodeLoc {
	l := CodeLoc{Path: path, Prelude: path == prelude.Path, Start: from, End: to}
	if d := a.doc(path); d != nil && from.Line >= 1 && from.Line <= len(d.lines) {
		l.Line = d.text[d.lines[from.Line-1]:d.lineEnd(from.Line-1)]
	}
	return l
}

// workspace opens every .ant file under the root as a document, once, so that every analysis
// sees the same text.
func (c *Code) workspace() {
	if c.walked {
		return
	}
	c.walked = true
	for _, f := range c.s.walk() {
		if _, ok := c.s.docs[f.doc.uri]; !ok {
			f.doc.version = 1
			c.s.docs[f.doc.uri] = f.doc
		}
	}
	// The walk stands for the rest of the run, and analyses made before it are made again.
	c.s.watching = true
	c.s.changed()
}

// open returns the document of a file, reading it the first time.
func (c *Code) open(path string) (*document, *CodeError) {
	uri := fileURI(path)
	if d, ok := c.s.docs[uri]; ok {
		return d, nil
	}
	src, err := os.ReadFile(path)
	if err != nil {
		e := codeErr("load", "%v", err)
		e.err = err
		return nil, e
	}
	d := newDocument(uri, 1, string(src))
	c.s.docs[uri] = d
	c.s.changed()
	return d, nil
}

// broken lists the files a search across the workspace and the programs of progs may have missed
// uses in: the files of the workspace that cannot be read or have syntax errors, and the modules
// of the programs that have syntax errors (code.md section 10).
func (c *Code) broken(progs []program) []string {
	c.workspace()
	set := map[string]bool{}
	for _, p := range c.s.unreadable {
		set[p] = true
	}
	inWorkspace := map[string]bool{}
	for _, f := range c.s.workspaceFiles() {
		inWorkspace[f.doc.path] = true
		if f.broken {
			set[f.doc.path] = true
		}
	}
	for _, pr := range progs {
		for path, m := range pr.a.modules {
			if !inWorkspace[path] && path != prelude.Path {
				if _, _, errs := parser.ParseRecover(path, m.Src); len(errs) > 0 {
					set[path] = true
				}
			}
		}
	}
	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// spot is what a location selects: the program rooted at its file, the position, and the name
// there, if any.
type spot struct {
	a   *analysis
	d   *document
	pos ast.Pos
	t   *target
}

var lineNumber = regexp.MustCompile(`^[0-9]+$`)

// resolve finds what a location selects (code.md section 2). A location that selects no name
// gives a spot without a target, which only type answers.
func (c *Code) resolve(loc string) (*spot, *CodeError) {
	i := strings.LastIndex(loc, ":")
	if i < 0 {
		if !validQualified(loc) || strings.HasSuffix(loc, ".ant") {
			return nil, usageLoc(loc)
		}
		return c.byName("", loc)
	}
	rest, last := loc[:i], loc[i+1:]
	if j := strings.LastIndex(rest, ":"); j >= 0 && lineNumber.MatchString(rest[j+1:]) {
		file := rest[:j]
		line, _ := strconv.Atoi(rest[j+1:])
		if lineNumber.MatchString(last) {
			col, _ := strconv.Atoi(last)
			return c.byPosition(file, line, col)
		}
		if !validIdent(last) {
			return nil, usageLoc(loc)
		}
		return c.byLine(file, line, last)
	}
	if rest == "" || !validQualified(last) {
		return nil, usageLoc(loc)
	}
	return c.byName(rest, last)
}

func usageLoc(loc string) *CodeError {
	return codeErr("usage", "%q is not a location: file:line:col, file:line:name, file:Name, or Name", loc)
}

func validIdent(s string) bool {
	if s == "" || s[0] >= '0' && s[0] <= '9' {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !token.IsIdentByte(s[i]) {
			return false
		}
	}
	return true
}

// validQualified reports whether s is Name or Owner.name.
func validQualified(s string) bool {
	owner, name, ok := strings.Cut(s, ".")
	if !ok {
		return validIdent(s)
	}
	return validIdent(owner) && validIdent(name)
}

// program opens a location's file and analyzes the program rooted at it, which must parse.
func (c *Code) program(file string) (*analysis, *document, *CodeError) {
	abs, err := filepath.Abs(file)
	if err != nil {
		return nil, nil, codeErr("load", "%v", err)
	}
	d, cerr := c.open(abs)
	if cerr != nil {
		var pe *fs.PathError
		if errors.As(cerr.err, &pe) {
			return nil, nil, codeErr("load", "cannot read %s: %v", file, pe.Err)
		}
		return nil, nil, cerr
	}
	if _, _, errs := parser.ParseRecover(d.path, d.text); len(errs) > 0 {
		return nil, nil, codeErr("load", "%s:%d:%d: %s", file, errs[0].Pos.Line, errs[0].Pos.Col, errs[0].Msg)
	}
	return c.s.analyze(d), d, nil
}

// lineText is the text of line n of d, counted from 1, or false for no such line.
func lineText(d *document, n int) (string, bool) {
	if n < 1 || n > len(d.lines) || n == len(d.lines) && d.lines[n-1] == len(d.text) && n > 1 {
		return "", false
	}
	return d.text[d.lines[n-1]:d.lineEnd(n-1)], true
}

// nameAt is the target whose name contains pos, from its first byte to its last.
func (a *analysis) nameAt(path string, pos ast.Pos) *target {
	t := a.at(path, pos)
	if t == nil || t.from.Line != pos.Line || pos.Col < t.from.Col || pos.Col >= t.to.Col || t.to.Line != t.from.Line {
		return nil
	}
	return t
}

func (c *Code) byPosition(file string, line, col int) (*spot, *CodeError) {
	a, d, err := c.program(file)
	if err != nil {
		return nil, err
	}
	text, ok := lineText(d, line)
	if !ok {
		return nil, codeErr("usage", "%s has no line %d", file, line)
	}
	if col < 1 || col > len(text)+1 {
		return nil, codeErr("usage", "line %d of %s has no column %d", line, file, col)
	}
	if col <= len(text) && !utf8.RuneStart(text[col-1]) {
		return nil, codeErr("usage", "column %d of line %d of %s is inside a character", col, line, file)
	}
	pos := ast.Pos{Line: line, Col: col}
	return &spot{a: a, d: d, pos: pos, t: a.nameAt(d.path, pos)}, nil
}

func (c *Code) byLine(file string, line int, name string) (*spot, *CodeError) {
	a, d, err := c.program(file)
	if err != nil {
		return nil, err
	}
	text, ok := lineText(d, line)
	if !ok {
		return nil, codeErr("usage", "%s has no line %d", file, line)
	}
	var found []*target
	var unresolved *target
	for i := 0; i+len(name) <= len(text); i++ {
		if text[i:i+len(name)] != name || i > 0 && token.IsIdentByte(text[i-1]) || i+len(name) < len(text) && token.IsIdentByte(text[i+len(name)]) {
			continue
		}
		pos := ast.Pos{Line: line, Col: i + 1}
		t := a.nameAt(d.path, pos)
		if t == nil || t.from != pos || t.to.Col != pos.Col+len(name) {
			continue
		}
		if t.obj == nil {
			if unresolved == nil {
				unresolved = t
			}
			continue
		}
		same := false
		for _, f := range found {
			same = same || sameBinding(a, f, t)
		}
		if !same {
			found = append(found, t)
		}
	}
	switch {
	case len(found) > 1:
		e := codeErr("ambiguous", "%s names %d bindings on line %d", name, len(found), line)
		for _, t := range found {
			e.Candidates = append(e.Candidates, a.codeLoc(d.path, t.from, t.to))
		}
		return nil, e
	case len(found) == 1:
		return &spot{a: a, d: d, pos: found[0].from, t: found[0]}, nil
	case unresolved != nil:
		return &spot{a: a, d: d, pos: unresolved.from, t: unresolved}, nil
	}
	return nil, codeErr("not-found", "no name %s on line %d of %s", name, line, file)
}

// sameBinding reports whether two targets name the same binding: the same object, reached
// through the same alias or through none.
func sameBinding(a *analysis, x, y *target) bool {
	ax, ay := a.aliasAt(x), a.aliasAt(y)
	return x.obj == y.obj && ax == ay
}

// declared is a declaration found by name.
type declared struct {
	path string
	pos  ast.Pos
	end  ast.Pos
}

// declsNamed finds the top-level declarations of f called name, Name or Owner.name: a function,
// a type, an interface, a variable, a function on Owner or in the interface Owner, or a field of
// the record type Owner.
func declsNamed(f *ast.File, name string) []declared {
	var out []declared
	add := func(owner, n string, pos ast.Pos) {
		full := n
		if owner != "" {
			full = owner + "." + n
		}
		if full == name {
			out = append(out, declared{path: f.Path, pos: pos, end: after(pos, n)})
		}
	}
	for _, decl := range f.Decls {
		switch x := decl.(type) {
		case *ast.FuncDecl:
			add(x.Owner, x.Name, x.NamePos)
		case *ast.TypeDecl:
			add(x.Owner, x.Name, x.NamePos)
			if rt, ok := x.Body.(*ast.RecordType); ok && x.Owner == "" {
				for i := range rt.Fields {
					add(x.Name, rt.Fields[i].Name, rt.Fields[i].Pos)
				}
			}
		case *ast.InterfaceDecl:
			add("", x.Name, x.NamePos)
			for i := range x.Funcs {
				add(x.Name, x.Funcs[i].Name, x.Funcs[i].NamePos)
			}
		case *ast.VarDecl:
			for _, nt := range targetNames(x.Target) {
				add(x.Owner, nt.Name, nt.Pos)
			}
		}
	}
	return out
}

// byName finds a declaration by name: in the module file, or, with no file, among the top-level
// declarations of the workspace and the prelude's.
func (c *Code) byName(file, name string) (*spot, *CodeError) {
	var found []declared
	if file != "" {
		a, d, err := c.program(file)
		if err != nil {
			return nil, err
		}
		for _, x := range declsNamed(a.module().File, name) {
			x.path = d.path
			found = append(found, x)
		}
		if len(found) == 0 {
			return nil, codeErr("not-found", "%s declares no %s", file, name)
		}
	} else {
		c.workspace()
		for _, f := range c.s.workspaceFiles() {
			for _, x := range declsNamed(f.file, name) {
				x.path = f.doc.path
				found = append(found, x)
			}
		}
		if f, err := prelude.Decls(); err == nil {
			for _, x := range declsNamed(f, name) {
				x.path = prelude.Path
				found = append(found, x)
			}
		}
		if len(found) == 0 {
			return nil, codeErr("not-found", "no declaration %s in the workspace or the prelude", name)
		}
	}
	if len(found) > 1 {
		e := codeErr("ambiguous", "%s names %d declarations", name, len(found))
		for _, x := range found {
			var a *analysis
			if x.path == prelude.Path {
				a = c.s.analyze(c.s.preludeDoc())
			} else {
				d, err := c.open(x.path)
				if err != nil {
					return nil, err
				}
				a = c.s.analyze(d)
			}
			e.Candidates = append(e.Candidates, a.codeLoc(x.path, x.pos, x.end))
		}
		return nil, e
	}
	x := found[0]
	var a *analysis
	var d *document
	if x.path == prelude.Path {
		d = c.s.preludeDoc()
		a = c.s.analyze(d)
	} else {
		var err *CodeError
		if a, d, err = c.program(x.path); err != nil {
			return nil, err
		}
	}
	return &spot{a: a, d: d, pos: x.pos, t: a.nameAt(d.path, x.pos)}, nil
}

// named resolves a location that must select a name with a declaration.
func (c *Code) named(loc string) (*spot, *CodeError) {
	sp, err := c.resolve(loc)
	if err != nil {
		return nil, err
	}
	switch {
	case sp.t == nil:
		return nil, codeErr("not-found", "%s selects no name", loc)
	case sp.t.imp == nil && sp.t.obj == nil:
		if sp.t.element != nil || sp.t.attr != nil {
			return nil, codeErr("not-found", "%s is markup, which has no declaration", loc)
		}
		return nil, codeErr("not-found", "%s is not declared", sp.a.nameText(sp.t))
	}
	return sp, nil
}

// nameText is the name a target spans, as written.
func (a *analysis) nameText(t *target) string {
	d := a.doc(t.path)
	if d == nil {
		return ""
	}
	return d.text[d.offsetOf(t.from):d.offsetOf(t.to)]
}

// Def answers ant code def: where the name at loc is declared, and its signature on one line.
func (c *Code) Def(loc string) (CodeLoc, string, *CodeError) {
	sp, err := c.named(loc)
	if err != nil {
		return CodeLoc{}, "", err
	}
	a, t := sp.a, sp.t
	if t.imp != nil {
		if t.imp.Module < 0 || t.imp.Module >= len(a.prog.Modules) {
			return CodeLoc{}, "", codeErr("not-found", "cannot read the module %q", t.imp.Path)
		}
		m := a.prog.Modules[t.imp.Module]
		start := ast.Pos{Line: 1, Col: 1}
		return a.codeLoc(m.File.Path, start, start), "import " + quote(t.imp.Path), nil
	}
	path, from, to, ok := a.declRange(t.obj)
	if !ok {
		return CodeLoc{}, "", codeErr("not-found", "%s has no declaration", t.obj.Name)
	}
	if path == "" {
		path = a.pathOf(t.obj)
	}
	head, _ := a.describe(t.obj)
	return a.codeLoc(path, from, to), oneLine(head), nil
}

// oneLine joins the lines of a signature: a line break after an opening bracket or before a
// closing one goes, with a comma before it, and any other becomes a space.
func oneLine(sig string) string {
	if !strings.Contains(sig, "\n") {
		return sig
	}
	lines := strings.Split(sig, "\n")
	var b strings.Builder
	for i, l := range lines {
		l = strings.TrimSpace(l)
		if i > 0 {
			prev := b.String()
			opened := strings.HasSuffix(prev, "(") || strings.HasSuffix(prev, "[")
			closing := strings.HasPrefix(l, ")") || strings.HasPrefix(l, "]") || strings.HasPrefix(l, "}")
			if closing {
				trimmed := strings.TrimSuffix(prev, ",")
				b.Reset()
				b.WriteString(trimmed)
			}
			switch {
			case opened, closing && !strings.HasPrefix(l, "}"):
			default:
				b.WriteString(" ")
			}
		}
		b.WriteString(l)
	}
	return b.String()
}

// CodeRef is a use refs reports, with the alias it is written through, if any.
type CodeRef struct {
	Loc   CodeLoc
	Alias string
}

// RefsResult is the answer of ant code refs.
type RefsResult struct {
	// Decl is the declaration's location; nil for a namespace of the prelude, which has none.
	Decl    *CodeLoc
	Refs    []CodeRef
	Skipped []string
}

// isPrelude reports whether o is declared by the prelude.
func isPrelude(o *check.Object) bool {
	return o.Path == prelude.Path || o.Decl == nil && (o.Kind == check.Namespace || o.Kind == check.Func || o.Kind == check.TypeKind || o.Kind == check.Const)
}

// Refs answers ant code refs: the uses of the declaration loc selects, and the declaration itself
// with decl.
func (c *Code) Refs(loc string, decl bool) (RefsResult, *CodeError) {
	sp, err := c.named(loc)
	if err != nil {
		return RefsResult{}, err
	}
	a, t := sp.a, sp.t
	if t.imp != nil {
		return RefsResult{}, codeErr("not-found", "%s is an import's path, which declares nothing", loc)
	}
	var res RefsResult
	var ms []mention
	var progs []program
	switch {
	case isPrelude(t.obj):
		progs = c.preludePrograms(t.obj)
		ms = mentionsIn(progs)
		if path, from, to, ok := a.declRange(t.obj); ok {
			l := a.codeLoc(path, from, to)
			res.Decl = &l
		}
		res.Skipped = c.broken(progs)
	case a.aliasAt(t) != nil:
		ms = c.s.mentionsAt(a, t)
		alias := a.aliasAt(t)
		l := a.codeLoc(t.path, alias.AliasPos, after(alias.AliasPos, alias.Alias))
		res.Decl = &l
	default:
		ms = c.s.mentionsAt(a, t)
		path, from, to, ok := a.declRange(t.obj)
		if path == "" {
			path = a.pathOf(t.obj)
		}
		if ok {
			l := a.codeLoc(path, from, to)
			res.Decl = &l
		}
		if a.reachesImporters(t.obj) {
			res.Skipped = c.broken(c.s.programs(a, t.obj))
		}
	}
	for _, m := range ms {
		if m.decl && !decl {
			continue
		}
		r := CodeRef{Loc: m.a.codeLoc(m.path, m.from, m.to)}
		if m.viaAlias {
			r.Alias = m.a.nameText(&target{path: m.path, from: m.from, to: m.to})
		}
		res.Refs = append(res.Refs, r)
	}
	sortLocs(res.Refs, func(r CodeRef) CodeLoc { return r.Loc })
	return res, nil
}

// preludePrograms lists the program rooted at each file of the workspace, with the object of the
// prelude's declaration o in it.
func (c *Code) preludePrograms(o *check.Object) []program {
	c.workspace()
	var out []program
	for _, f := range c.s.workspaceFiles() {
		a := c.s.analyze(f.doc)
		if same := a.preludeObject(o); same != nil {
			out = append(out, program{a, same})
		}
	}
	return out
}

// preludeObject finds the object of this analysis for a prelude declaration's object of another.
func (a *analysis) preludeObject(o *check.Object) *check.Object {
	same := func(p *check.Object) bool {
		if o.Decl != nil {
			return p.Decl == o.Decl
		}
		return p.Name == o.Name && p.Kind == o.Kind && p.Owner == o.Owner
	}
	for _, p := range a.info.Prelude {
		if same(p) {
			return p
		}
		for _, m := range p.Members() {
			if same(m) {
				return m
			}
		}
	}
	return nil
}

// mentionsIn finds the mentions of each program's object, each once.
func mentionsIn(progs []program) []mention {
	var out []mention
	seen := map[reference]bool{}
	for _, pr := range progs {
		for _, r := range pr.a.referencesOf(pr.o) {
			key := reference{path: r.path, from: r.from}
			if !seen[key] {
				seen[key] = true
				out = append(out, mention{pr.a, r})
			}
		}
	}
	return out
}

func sortLocs[T any](xs []T, loc func(T) CodeLoc) {
	sort.SliceStable(xs, func(i, j int) bool {
		x, y := loc(xs[i]), loc(xs[j])
		if x.Path != y.Path {
			return x.Path < y.Path
		}
		return x.Start.Before(y.Start)
	})
}

// TypeResult is the answer of ant code type.
type TypeResult struct {
	Kind string
	Text string
	Doc  string
	Loc  CodeLoc
}

// operandKeywords are the keywords that are values.
var operandKeywords = map[string]bool{"true": true, "false": true, "null": true, "void": true}

// Type answers ant code type: what hover shows for the name loc selects, or the type of the
// innermost expression around a literal or an operator.
func (c *Code) Type(loc string) (TypeResult, *CodeError) {
	sp, err := c.resolve(loc)
	if err != nil {
		return TypeResult{}, err
	}
	a, t := sp.a, sp.t
	switch {
	case t != nil && t.imp != nil:
		res := TypeResult{Kind: "import", Text: "import " + quote(t.imp.Path), Loc: a.codeLoc(t.path, t.from, t.to)}
		if t.imp.Module >= 0 && t.imp.Module < len(a.prog.Modules) {
			res.Doc = a.prog.Modules[t.imp.Module].File.Doc
		}
		return res, nil
	case t != nil && t.obj != nil:
		head, doc := a.describeAt(t)
		return TypeResult{Kind: t.obj.Kind.String(), Text: head, Doc: doc, Loc: a.codeLoc(t.path, t.from, t.to)}, nil
	case t != nil:
		return TypeResult{}, codeErr("not-found", "%s has no type the checker could tell", a.nameText(t))
	}
	tok, ok := tokenAt(sp.d, sp.pos)
	switch {
	case !ok:
		return TypeResult{}, codeErr("not-found", "%s is not on an expression", loc)
	case tok.Kind == token.Keyword && !operandKeywords[tok.Text], tok.Kind == token.Reserved, tok.Kind == token.Ident, tok.Kind == token.Bad:
		return TypeResult{}, codeErr("not-found", "%s is on %q, which has no type", loc, tok.Text)
	}
	nodes := ast.Path(a.module().File, sp.pos)
	for i := len(nodes) - 1; i >= 0; i-- {
		e, ok := nodes[i].(ast.Expr)
		if !ok {
			continue
		}
		ty, ok := a.info.Types[e]
		if !ok || ty == nil {
			continue
		}
		if types.IsInvalid(ty) {
			break
		}
		return TypeResult{Kind: "expression", Text: typeString(ty), Loc: a.codeLoc(sp.d.path, ast.Start(e), e.EndPosition())}, nil
	}
	return TypeResult{}, codeErr("not-found", "%s is in no expression the checker could type", loc)
}

// tokenAt finds the token that holds pos, looking into the interpolations of a string, or false
// on white space or a comment.
func tokenAt(d *document, pos ast.Pos) (token.Token, bool) {
	_, layout, _ := parser.ParseRecover(d.path, d.text)
	if layout == nil {
		return token.Token{}, false
	}
	off := d.offsetOf(pos)
	toks := layout.Tokens
	for {
		i := sort.Search(len(toks), func(i int) bool { return toks[i].End > off })
		if i == len(toks) || toks[i].Off > off || toks[i].Kind == token.Newline || toks[i].Kind == token.EOF {
			return token.Token{}, false
		}
		tok := toks[i]
		if tok.Kind != token.Template {
			return tok, true
		}
		inner := false
		for _, p := range tok.Parts {
			if p.Tokens != nil && p.Off <= off && off < p.End {
				toks, inner = p.Tokens, true
				break
			}
		}
		if !inner {
			return tok, true
		}
	}
}

// CodeSymbol is a declaration symbols lists.
type CodeSymbol struct {
	Loc       CodeLoc
	Kind      string
	Name      string
	Signature string
	Exported  bool
	Container string
}

// Symbols answers ant code symbols: the declarations of the modules under each path, the root
// when there is none, whose names contain match, without regard to case; with all, the fields
// of record types and the declarations in functions too. It also returns the files with syntax
// errors, whose declarations are those it recovered.
func (c *Code) Symbols(paths []string, match string, all bool) ([]CodeSymbol, []string, *CodeError) {
	if len(paths) == 0 {
		paths = c.s.roots
	}
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, nil, codeErr("load", "%v", err)
		}
		abs, _ := filepath.Abs(p)
		if !info.IsDir() {
			files = append(files, abs)
			continue
		}
		sub := &Server{roots: []string{abs}}
		for _, f := range sub.walk() {
			files = append(files, f.doc.path)
		}
	}
	sort.Strings(files)
	match = strings.ToLower(match)
	var out []CodeSymbol
	var broken []string
	seen := map[string]bool{}
	for _, path := range files {
		if seen[path] {
			continue
		}
		seen[path] = true
		d, err := c.open(path)
		if err != nil {
			return nil, nil, err
		}
		f, _, errs := parser.ParseRecover(d.path, d.text)
		if len(errs) > 0 {
			broken = append(broken, d.path)
		}
		var a *analysis
		if all {
			a = c.s.analyze(d)
			f = a.module().File
		}
		for _, sym := range c.fileSymbols(d, f, a) {
			if strings.Contains(strings.ToLower(sym.Name), match) {
				out = append(out, sym)
			}
		}
	}
	return out, broken, nil
}

// fileSymbols lists a module's declarations in source order, the functions on a type after it,
// the tests last; with an analysis, also each record type's fields and each function's
// declarations.
func (c *Code) fileSymbols(d *document, f *ast.File, a *analysis) []CodeSymbol {
	loc := func(from ast.Pos, name string) CodeLoc {
		l := CodeLoc{Path: d.path, Start: from, End: after(from, name)}
		l.Line, _ = lineText(d, from.Line)
		return l
	}
	sig := func(export bool, s string) string {
		s = oneLine(s)
		if export {
			return "export " + s
		}
		return s
	}
	isType := map[string]bool{}
	for _, decl := range f.Decls {
		if x, ok := decl.(*ast.TypeDecl); ok && x.Owner == "" {
			isType[x.Name] = true
		}
	}
	var out []CodeSymbol
	var funcSymbols func(x *ast.FuncDecl) []CodeSymbol
	funcSymbols = func(x *ast.FuncDecl) []CodeSymbol {
		out := []CodeSymbol{{Loc: loc(x.NamePos, x.Name), Kind: "function", Name: x.Name, Signature: sig(x.Export, x.Sig), Exported: x.Export, Container: x.Owner}}
		if a != nil && x.Body != nil {
			out = append(out, a.localSymbols(x, loc)...)
		}
		return out
	}
	var typeSymbols func(x *ast.TypeDecl) []CodeSymbol
	typeSymbols = func(x *ast.TypeDecl) []CodeSymbol {
		out := []CodeSymbol{{Loc: loc(x.NamePos, x.Name), Kind: "type", Name: x.Name, Signature: sig(x.Export, x.Sig), Exported: x.Export, Container: x.Owner}}
		if rt, ok := x.Body.(*ast.RecordType); ok && a != nil {
			for i := range rt.Fields {
				fl := &rt.Fields[i]
				out = append(out, CodeSymbol{Loc: loc(fl.Pos, fl.Name), Kind: "field", Name: fl.Name, Signature: fl.Name + ": " + oneLine(typeText(fl.Type, d)), Exported: x.Export, Container: x.Name})
			}
		}
		if x.Owner == "" {
			for _, decl := range f.Decls {
				if fd, ok := decl.(*ast.FuncDecl); ok && fd.Owner == x.Name {
					out = append(out, funcSymbols(fd)...)
				}
			}
		}
		return out
	}
	for _, decl := range f.Decls {
		switch x := decl.(type) {
		case *ast.FuncDecl:
			if x.Owner == "" || !isType[x.Owner] {
				out = append(out, funcSymbols(x)...)
			}
		case *ast.TypeDecl:
			out = append(out, typeSymbols(x)...)
		case *ast.InterfaceDecl:
			out = append(out, CodeSymbol{Loc: loc(x.NamePos, x.Name), Kind: "interface", Name: x.Name, Signature: sig(x.Export, x.Sig), Exported: x.Export})
			for i := range x.Funcs {
				fn := &x.Funcs[i]
				out = append(out, CodeSymbol{Loc: loc(fn.NamePos, fn.Name), Kind: "function", Name: fn.Name, Signature: oneLine(fn.Sig), Exported: x.Export, Container: x.Name})
			}
		case *ast.VarDecl:
			for _, nt := range targetNames(x.Target) {
				s := x.Sig
				if len(targetNames(x.Target)) > 1 || s == "" {
					s = varKeyword(x) + " " + nt.Name
				}
				out = append(out, CodeSymbol{Loc: loc(nt.Pos, nt.Name), Kind: "variable", Name: nt.Name, Signature: sig(x.Export, s), Exported: x.Export, Container: x.Owner})
			}
		}
	}
	var tests func(items []*testItem, parent string)
	tests = func(items []*testItem, parent string) {
		for _, it := range items {
			name := it.name
			if parent != "" {
				name = parent + "/" + it.name
			}
			lit := it.node.Name.(*ast.StringLit)
			l := CodeLoc{Path: d.path, Start: lit.Pos, End: lit.End}
			l.Line, _ = lineText(d, lit.Pos.Line)
			out = append(out, CodeSymbol{Loc: l, Kind: "test", Name: name, Signature: "test " + strconv.Quote(name), Container: parent})
			tests(it.children, name)
		}
	}
	tests(fileTests(f), "")
	return out
}

func varKeyword(x *ast.VarDecl) string {
	if x.Let {
		return "let"
	}
	return "var"
}

// localSymbols lists the declarations inside a function, with the types the checker gave them.
func (a *analysis) localSymbols(fd *ast.FuncDecl, loc func(ast.Pos, string) CodeLoc) []CodeSymbol {
	container := fd.Name
	if fd.Owner != "" {
		container = fd.Owner + "." + fd.Name
	}
	var out []CodeSymbol
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.VarDecl:
			for _, nt := range targetNames(x.Target) {
				s := varKeyword(x) + " " + nt.Name
				if o := a.info.Decls[nt]; o != nil {
					s, _ = a.describe(o)
				}
				out = append(out, CodeSymbol{Loc: loc(nt.Pos, nt.Name), Kind: "variable", Name: nt.Name, Signature: oneLine(s), Container: container})
			}
		case *ast.TypeDecl:
			out = append(out, CodeSymbol{Loc: loc(x.NamePos, x.Name), Kind: "type", Name: x.Name, Signature: oneLine(x.Sig), Container: container})
		case *ast.FuncDecl:
			if x != fd {
				out = append(out, CodeSymbol{Loc: loc(x.NamePos, x.Name), Kind: "function", Name: x.Name, Signature: oneLine(x.Sig), Container: container})
			}
		}
		return true
	})
	return out
}

// CodeEdit is one name a rename changes.
type CodeEdit struct {
	Loc     CodeLoc
	NewText string
}

// CodeFile is a file a rename changes: its edits in source order, the text they were found in,
// and the text they make.
type CodeFile struct {
	Path     string
	Edits    []CodeEdit
	Old, New string
}

// RenameResult is the answer of ant code rename.
type RenameResult struct {
	Files   []CodeFile
	Skipped []string
	// Errors are the errors of the programs the rename changes, which force let through.
	Errors []CodeDiag
}

// Rename answers ant code rename: the edits that rename the declaration loc selects to name, or
// why it cannot be renamed. With force, a program with errors or an incomplete search does not
// stop it, and the result carries the errors and the files skipped.
func (c *Code) Rename(loc, name string, force bool) (RenameResult, *CodeError) {
	if err := validName(name); err != nil {
		return RenameResult{}, codeErr("refused", "%s", err.Message)
	}
	sp, cerr := c.named(loc)
	if cerr != nil {
		return RenameResult{}, cerr
	}
	a, t := sp.a, sp.t
	if t.imp != nil {
		return RenameResult{}, codeErr("refused", "an import's path is not renamed")
	}
	if a.nameText(t) == name {
		return RenameResult{}, nil
	}
	edits, err := c.s.renameAt(a, t, name)
	if err != nil {
		if err.Message == "nothing to rename here" {
			return RenameResult{}, codeErr("not-found", "%s", err.Message)
		}
		return RenameResult{}, codeErr("refused", "%s", err.Message)
	}
	var res RenameResult
	// The files changed, and the programs: each analysis an edit was found in.
	byPath := map[string]*CodeFile{}
	changed := map[*analysis]bool{}
	for _, e := range edits {
		if rel, err := filepath.Rel(c.s.roots[0], e.path); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return RenameResult{}, codeErr("refused", "%s is outside the root %s", e.path, c.s.roots[0])
		}
		changed[e.a] = true
		f := byPath[e.path]
		if f == nil {
			f = &CodeFile{Path: e.path, Old: e.a.doc(e.path).text}
			byPath[e.path] = f
		}
		f.Edits = append(f.Edits, CodeEdit{Loc: e.a.codeLoc(e.path, e.from, e.to), NewText: e.text})
	}
	for _, f := range byPath {
		sortLocs(f.Edits, func(e CodeEdit) CodeLoc { return e.Loc })
		d := newDocument("", 0, f.Old)
		var b strings.Builder
		last := 0
		for _, e := range f.Edits {
			from, to := d.offsetOf(e.Loc.Start), d.offsetOf(e.Loc.End)
			b.WriteString(f.Old[last:from])
			b.WriteString(e.NewText)
			last = to
		}
		b.WriteString(f.Old[last:])
		f.New = b.String()
		res.Files = append(res.Files, *f)
	}
	sort.Slice(res.Files, func(i, j int) bool { return res.Files[i].Path < res.Files[j].Path })
	// Checked before it is written: every program it changes (code.md section 8).
	seen := map[string]bool{}
	for pa := range changed {
		for _, dg := range pa.diags {
			if dg.Warning {
				continue // a warning does not make a use the rename may miss
			}
			key := fmt.Sprintf("%s:%d:%d:%s", dg.Path, dg.Pos.Line, dg.Pos.Col, dg.Msg)
			if !seen[key] {
				seen[key] = true
				res.Errors = append(res.Errors, CodeDiag{Path: dg.Path, Pos: dg.Pos, End: dg.End, Message: dg.Msg, Code: string(dg.Code)})
			}
		}
	}
	sort.SliceStable(res.Errors, func(i, j int) bool {
		x, y := res.Errors[i], res.Errors[j]
		if x.Path != y.Path {
			return x.Path < y.Path
		}
		return x.Pos.Before(y.Pos)
	})
	if a.aliasAt(t) == nil && a.reachesImporters(t.obj) {
		res.Skipped = c.broken(c.s.programs(a, t.obj))
	}
	if !force {
		if len(res.Errors) > 0 {
			return RenameResult{}, &CodeError{Code: "check", Message: "a program the rename changes has errors; -force renames anyway", Errors: res.Errors}
		}
		if len(res.Skipped) > 0 {
			return RenameResult{}, &CodeError{Code: "incomplete", Message: "some files of the workspace cannot be read or parsed, and may use the name; -force renames anyway", Skipped: res.Skipped}
		}
	}
	return res, nil
}
