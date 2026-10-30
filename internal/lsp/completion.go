package lsp

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/internal/lsp/htmldata"
	"github.com/ant-src/ant/internal/prelude"
	"github.com/ant-src/ant/internal/token"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/parser"
	"github.com/ant-src/ant/value"
)

// placeholder stands in for the name being typed when the server parses the document again to
// learn what the cursor is in (see fixup).
const placeholder = "__ant_cmp__"

// completion answers textDocument/completion (spec/lsp.md section 4).
func (s *Server) completion(p *CompletionParams) *CompletionList {
	d := s.doc(p.TextDocument.URI)
	if d == nil {
		return nil
	}
	return &CompletionList{Items: s.completeAt(d, d.offset(p.Position, s.utf16))}
}

// completeAt collects what can be written at byte offset off of d.
func (s *Server) completeAt(d *document, off int) []CompletionItem {
	a := s.analyze(d)
	c := &completer{s: s, a: a, d: d, off: off, toks: a.module().Layout.Tokens, items: []CompletionItem{}, seen: map[string]bool{}}
	word, start := d.wordAt(off)
	c.wordStart, c.wordEnd = start, start+len(word)
	c.run()
	return c.items
}

// Complete lists what can be written at byte offset off of text, the input of a session at
// path, for a REPL: cfg is the session's loader configuration, whose Session scope and Extra
// names the input is resolved against; the text itself overrides whatever cfg.Sources holds at
// path. The text is parsed as a script, since the checker types the statements of a script
// and not those of a session input; the scope still lets a name be declared again. The items
// are those the language server would offer (spec/lsp.md section 4), unsorted.
func Complete(cfg loader.Config, path, text string, off int) []CompletionItem {
	s := &Server{docs: map[string]*document{}, cache: map[string]*analysis{}}
	s.load = func(p string, sources map[string]string) (*loader.Program, []loader.Diagnostic) {
		c := cfg
		c.Mode, c.Recover = parser.Program, true
		c.Sources = make(map[string]string, len(cfg.Sources)+len(sources))
		for k, v := range cfg.Sources {
			c.Sources[k] = v
		}
		for k, v := range sources {
			c.Sources[k] = v
		}
		return loader.Load(p, c)
	}
	d := newDocument(fileURI(path), 0, text)
	d.path = path
	s.docs[d.uri] = d
	if off < 0 || off > len(text) {
		return nil
	}
	return s.completeAt(d, off)
}

// completer collects the items for one request.
type completer struct {
	s         *Server
	a         *analysis
	d         *document
	off       int
	wordStart int // where the identifier under the cursor starts
	wordEnd   int // where it ends, past the cursor when it goes on
	// toks is the token stream the cursor is in: the document's, or an interpolation's inside a
	// template string.
	toks  []token.Token
	items []CompletionItem
	seen  map[string]bool
}

func (c *completer) add(item CompletionItem) {
	key := item.Label + "\x00" + item.Detail
	if c.seen[key] {
		return
	}
	c.seen[key] = true
	c.items = append(c.items, item)
}

// run decides what the cursor is in and collects what can be written there.
func (c *completer) run() {
	m := c.a.module()
	toks := m.Layout.Tokens
	// Inside a comment, a string, or a regular expression nothing is completed.
	for _, cm := range m.Layout.Comments {
		if cm.Off < c.off && c.off <= cm.End {
			return
		}
	}
	for i := range toks {
		t := &toks[i]
		if t.Off < c.wordStart && c.off <= t.End || t.Off < c.off && c.off < t.End {
			switch t.Kind {
			case token.String:
				if i > 0 && toks[i-1].Is("import") {
					c.importPaths(t)
				}
				return
			case token.Regex:
				return
			case token.Html:
				c.markup(t)
				return
			case token.Template:
				part := c.interpolation(t)
				if part == nil {
					return // in the text of the template
				}
				c.toks = part.Tokens
			}
		}
	}
	prev := c.tokenBefore(c.toks, c.wordStart)
	switch {
	case prev != nil && (prev.Is(".") || prev.Is("?.")) && prev.End == c.wordStart:
		c.members(prev)
		return
	case prev != nil && prev.Is("is"):
		c.isTargets()
		return
	}
	if c.importNames() {
		return
	}
	if c.typePosition(prev) {
		c.typeNames("")
		return
	}
	if c.pipe(prev) {
		return
	}
	c.recordFields(prev)
	c.namedArgs(prev)
	c.scopeNames()
	c.keywords(prev)
}

// tokenBefore is the last token that ends at or before off, line breaks left out.
func (c *completer) tokenBefore(toks []token.Token, off int) *token.Token {
	var prev *token.Token
	for i := range toks {
		t := &toks[i]
		if t.Kind == token.Newline || t.Kind == token.EOF {
			continue
		}
		if t.End <= off {
			prev = t
		} else {
			break
		}
	}
	return prev
}

// interpolation is the ${...} of a template string the cursor is in, or nil in its text.
func (c *completer) interpolation(t *token.Token) *token.Part {
	for i := range t.Parts {
		part := &t.Parts[i]
		if part.Tokens != nil && part.Off < c.off && c.off < part.End {
			return part
		}
	}
	return nil
}

// fixup analyzes the document again with the word under the cursor replaced by placeholder, and
// with extra text after it, so that an unfinished statement parses and the tree says what the
// cursor is in. It returns the new analysis and the placeholder's identifier, when it is one.
func (c *completer) fixup(extra string) (*analysis, *ast.Ident) {
	text := c.d.text[:c.wordStart] + placeholder + extra + c.d.text[c.wordEnd:]
	a := c.s.analyzeText(c.d, text)
	var found *ast.Ident
	ast.Inspect(a.module().File, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == placeholder {
			found = id
		}
		return found == nil
	})
	return a, found
}

// analyzeText analyzes the program rooted at d with text in place of its own, for completion.
func (s *Server) analyzeText(d *document, text string) *analysis {
	sources := map[string]string{}
	for _, od := range s.docs {
		if od.path != "" {
			sources[od.path] = od.text
		}
	}
	sources[d.path] = text
	a := &analysis{root: d, modules: map[string]*loader.Module{}, docs: map[string]*document{}, server: s}
	prog, diags := s.loadProgram(d.path, sources)
	if prog == nil {
		prog, _ = s.loadProgram(d.path, map[string]string{d.path: text})
	}
	a.prog, a.diags = prog, diags
	for _, m := range prog.Modules {
		a.modules[m.File.Path] = m
		a.docs[m.File.Path] = newDocument(fileURI(m.File.Path), 0, m.Src)
	}
	a.info = check.NewInfo()
	check.Program(prog, check.Config{Info: a.info})
	return a
}

// ---- members

// members completes after a dot: the members of a namespace, a module, or a type, or the
// fields of a record.
func (c *completer) members(dot *token.Token) {
	extra := ""
	if c.inPipeRight(dot) {
		extra = "()" // the right side of a pipe must be a call for the statement to parse
	}
	a, _ := c.fixup(extra)
	// the field whose name the placeholder is
	var fld *ast.Field
	ast.Inspect(a.module().File, func(n ast.Node) bool {
		if f, ok := n.(*ast.Field); ok && f.Name == placeholder {
			fld = f
		}
		return fld == nil
	})
	if fld == nil {
		c.typeMembers(a)
		return
	}
	piped := c.pipedLeft(a, fld)
	if obj := a.entityOf(fld.X); obj != nil {
		for _, m := range obj.Members() {
			if obj.Kind == check.Module && !m.Exported {
				continue
			}
			if piped != nil && m.Kind == check.Func && !acceptsFirst(m, piped) {
				continue
			}
			c.add(c.objectItem(a, m, ""))
		}
		return
	}
	recv := a.info.Types[fld.X]
	if recv == nil {
		return
	}
	if fld.Optional {
		recv = types.Without(recv, types.Null{})
	}
	for _, f := range commonFields(recv) {
		c.add(CompletionItem{Label: f.Name, Kind: ItemField, Detail: types.String(f.Type)})
	}
}

// inPipeRight reports whether the dot ends a path that is the right side of a pipe: xs | Array.
func (c *completer) inPipeRight(dot *token.Token) bool {
	toks := c.toks
	i := -1
	for j := range toks {
		if toks[j].Off == dot.Off {
			i = j
		}
	}
	for i--; i >= 0; i-- {
		t := toks[i]
		switch {
		case t.Kind == token.Newline:
		case t.Kind == token.Ident:
		case t.Is(".") || t.Is("?."):
		case t.Is("|"):
			return true
		default:
			return false
		}
	}
	return false
}

// typeMembers completes after a dot in a type: the types a namespace or a module holds, found
// through the type name whose last part the placeholder is.
func (c *completer) typeMembers(a *analysis) {
	var tn *ast.TypeName
	ast.Inspect(a.module().File, func(n ast.Node) bool {
		if x, ok := n.(*ast.TypeName); ok && x.Path[len(x.Path)-1] == placeholder {
			tn = x
		}
		return tn == nil
	})
	if tn == nil {
		return
	}
	objs := a.info.TypeNames[tn]
	if len(objs) < len(tn.Path)-1 || len(objs) == 0 {
		return
	}
	owner := objs[len(tn.Path)-2]
	for _, m := range owner.Members() {
		if m.Kind == check.TypeKind || m.Kind == check.Namespace && hasTypes(m) {
			if owner.Kind == check.Module && !m.Exported {
				continue
			}
			c.add(c.objectItem(a, m, ""))
		}
	}
}

// entityOf is the object an expression names when it is a path to a namespace, a module, or a
// type, or nil for a value.
func (a *analysis) entityOf(e ast.Expr) *check.Object {
	var o *check.Object
	switch x := e.(type) {
	case *ast.Ident:
		o = a.info.Objects[x]
	case *ast.Field:
		o = a.info.Members[x]
	}
	if o != nil && (o.Kind == check.Namespace || o.Kind == check.Module || o.Kind == check.TypeKind) {
		return o
	}
	return nil
}

// commonFields lists the fields every member of a type declares (section 4, Unions), or the
// fields of a record.
func commonFields(t types.Type) []types.Field {
	members := types.Members(t)
	var out []types.Field
	for i, m := range members {
		fields := types.FieldsOf(m)
		if i == 0 {
			out = fields
			continue
		}
		var keep []types.Field
		for _, f := range out {
			for _, g := range fields {
				if g.Name == f.Name {
					keep = append(keep, f)
					break
				}
			}
		}
		out = keep
	}
	return out
}

// pipedLeft is the type of the left side when the field being completed is the function of a
// pipe's right side, so that the members can be filtered to those that take it.
func (c *completer) pipedLeft(a *analysis, fld *ast.Field) types.Type {
	var left types.Type
	ast.Inspect(a.module().File, func(n ast.Node) bool {
		if call, ok := n.(*ast.Call); ok && call.Pipe && call.Fn == fld && len(call.Args) > 0 {
			left = a.info.Types[call.Args[0].Value]
		}
		return left == nil
	})
	return left
}

// ---- pipes

// pipe completes after |: the functions whose first parameter takes the left side's type, from
// the prelude, the imported modules, and the types in scope.
func (c *completer) pipe(prev *token.Token) bool {
	if prev == nil || !prev.Is("|") {
		return false
	}
	a, id := c.fixup("()")
	if id == nil {
		return true
	}
	var left types.Type
	ast.Inspect(a.module().File, func(n ast.Node) bool {
		if call, ok := n.(*ast.Call); ok && call.Pipe && call.Fn == id && len(call.Args) > 0 {
			left = a.info.Types[call.Args[0].Value]
		}
		return left == nil
	})
	if left == nil || types.IsInvalid(left) {
		c.scopeNames()
		return true
	}
	offer := func(owner *check.Object) {
		for _, m := range owner.Members() {
			if m.Kind == check.Func && (owner.Kind != check.Module || m.Exported) && acceptsFirst(m, left) {
				item := c.objectItem(a, m, owner.Name+".")
				c.add(item)
			}
		}
	}
	for _, o := range a.info.Prelude {
		if o.Kind == check.Namespace || o.Kind == check.TypeKind {
			offer(o)
		}
	}
	if mi, ok := a.info.Modules[a.module().File.Path]; ok {
		for _, o := range mi.Names {
			switch o.Kind {
			case check.Module, check.TypeKind:
				offer(o)
			case check.Func:
				if acceptsFirst(o, left) {
					c.add(c.objectItem(a, o, ""))
				}
			}
		}
	}
	return true
}

// acceptsFirst reports whether a function's first parameter can take a value of type t.
func acceptsFirst(fn *check.Object, t types.Type) bool {
	ft, ok := fn.Type().(types.Func)
	if !ok || len(ft.Params) == 0 {
		return false
	}
	return accepts(ft.Params[0], t)
}

// accepts reports whether a parameter type can take an argument type, binding a type parameter
// to anything, as a call's inference would (section 4, Inference).
func accepts(param, arg types.Type) bool {
	if !types.HasParam(param) {
		return types.Assignable(arg, param)
	}
	switch p := types.Expand(param).(type) {
	case types.Param:
		if p.Decl.Bound == nil {
			return true
		}
		return accepts(p.Decl.Bound, arg)
	case types.Array:
		a, ok := types.Expand(arg).(types.Array)
		return ok && accepts(p.Elem, a.Elem)
	case types.Map:
		m, ok := types.Expand(arg).(types.Map)
		return ok && accepts(p.Key, m.Key) && accepts(p.Val, m.Val)
	case types.Set:
		s, ok := types.Expand(arg).(types.Set)
		return ok && accepts(p.Elem, s.Elem)
	case types.Named:
		n, ok := types.Expand(arg).(types.Named)
		if !ok || n.Decl != p.Decl || len(n.Args) != len(p.Args) {
			return false
		}
		for i := range p.Args {
			if !accepts(p.Args[i], n.Args[i]) {
				return false
			}
		}
		return true
	case types.Func:
		f, ok := types.Expand(arg).(types.Func)
		return ok && len(f.Params) == len(p.Params)
	case types.Union:
		for _, m := range p.Members {
			if accepts(m, arg) {
				return true
			}
		}
		return false
	case types.Record:
		fields := types.FieldsOf(arg)
		for _, f := range p.Fields {
			found := false
			for _, g := range fields {
				if g.Name == f.Name && accepts(f.Type, g.Type) {
					found = true
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
	return false
}

// ---- names in scope

// scopeNames completes the names in scope: locals declared above, parameters, x in a lambda, the
// module's declarations, the imported names, and the prelude.
func (c *completer) scopeNames() {
	a := c.a
	m := a.module()
	pos := c.d.posOf(c.wordStart)
	if m.Resolve != nil {
		nodes := ast.Path(m.File, pos)
		for i := len(nodes) - 1; i >= 0; i-- {
			sc, ok := m.Resolve.Scopes[nodes[i]]
			if !ok {
				continue
			}
			_, module := nodes[i].(*ast.File)
			for _, b := range sc.Names {
				if !module && !b.Type && !b.Pos.Before(pos) {
					continue // declared below
				}
				if b.Name == "" || b.Name == "_" {
					continue
				}
				o := a.info.Decls[b.Node]
				if o != nil {
					item := c.objectItem(a, o, "")
					item.Label = b.Name // an import's alias
					c.add(item)
				} else {
					c.add(CompletionItem{Label: b.Name, Kind: ItemVariable})
				}
			}
		}
	}
	for _, o := range a.info.Prelude {
		c.add(c.objectItem(a, o, ""))
	}
}

// objectItem is the completion item for an object, with its kind, type, and summary.
func (c *completer) objectItem(a *analysis, o *check.Object, prefix string) CompletionItem {
	item := CompletionItem{Label: prefix + o.Name}
	switch o.Kind {
	case check.Func:
		item.Kind = ItemFunction
		if o.Owner != "" {
			item.Kind = ItemMethod
		}
		item.Detail = strings.TrimPrefix(a.signatureLabel(o), o.Owner+".")
		item.Detail = strings.TrimPrefix(item.Detail, o.Name)
	case check.Var:
		item.Kind = ItemVariable
		item.Detail = typeString(o.Type())
	case check.Param:
		item.Kind = ItemVariable
		item.Detail = typeString(o.Type())
	case check.TypeKind:
		item.Kind = ItemClass
		if d := o.TypeDecl(); d != nil && d.Kind != types.RecordDecl {
			item.Kind = ItemInterface
		}
		item.Detail = "type"
	case check.TypeParam:
		item.Kind = ItemTypeParameter
		item.Detail = "type parameter"
	case check.Namespace:
		item.Kind = ItemModule
		item.Detail = "namespace"
	case check.Module:
		item.Kind = ItemModule
		item.Detail = "module"
	case check.Const:
		item.Kind = ItemConstant
		item.Detail = typeString(o.Type())
	case check.Field:
		item.Kind = ItemField
		item.Detail = typeString(o.Type())
	}
	doc := o.Doc()
	if o.Kind == check.Namespace {
		_, doc = prelude.NamespaceDoc(o.Name)
	}
	if doc != "" {
		item.Documentation = markdown(summary(doc, a.knows))
	}
	return item
}

// ---- keywords

var exprKeywords = []string{"true", "false", "null", "void", "try", "func", "if", "select", "while", "for", "go", "defer", "return", "break", "continue"}
var stmtKeywords = []string{"var", "type"}
var topKeywords = []string{"import", "export"}

// keywords completes the keywords that can start an expression or a statement here.
func (c *completer) keywords(prev *token.Token) {
	atStatementStart := prev == nil || prev.Is("{") || prev.Is(";") || prev.Is("}") || c.lineStart(prev)
	for _, k := range exprKeywords {
		c.add(CompletionItem{Label: k, Kind: ItemKeyword})
	}
	if atStatementStart {
		for _, k := range stmtKeywords {
			c.add(CompletionItem{Label: k, Kind: ItemKeyword})
		}
		if c.atTopLevel() {
			for _, k := range topKeywords {
				c.add(CompletionItem{Label: k, Kind: ItemKeyword})
			}
		}
	}
}

// lineStart reports whether the cursor's word starts a line after prev.
func (c *completer) lineStart(prev *token.Token) bool {
	return c.d.lineOf(prev.End) < c.d.lineOf(c.wordStart)
}

// atTopLevel reports whether the cursor is outside every declaration.
func (c *completer) atTopLevel() bool {
	nodes := ast.Path(c.a.module().File, c.d.posOf(c.wordStart))
	for _, n := range nodes {
		switch n.(type) {
		case *ast.FuncDecl, *ast.TypeDecl, *ast.VarDecl, *ast.Block:
			return false
		}
	}
	return true
}

// ---- types

// typePosition reports whether a type is written at the cursor: inside a type expression, or
// after the colon of a parameter, a field, a variable's annotation, or a type parameter's bound,
// or after the = of a type declaration.
func (c *completer) typePosition(prev *token.Token) bool {
	nodes := ast.Path(c.a.module().File, c.d.posOf(c.wordStart))
	pos := c.d.posOf(c.wordStart)
	for i := len(nodes) - 1; i >= 0; i-- {
		switch n := nodes[i].(type) {
		case *ast.TypeName, *ast.RecordType, *ast.FuncType, *ast.UnionType, *ast.BadType:
			return true
		case *ast.FieldType:
			return !pos.Before(after(n.Pos, n.Name))
		case *ast.Param:
			return !pos.Before(after(n.Pos, n.Name)) && (n.Default == nil || pos.Before(ast.Start(n.Default)))
		case *ast.TypeParam:
			return !pos.Before(after(n.Pos, n.Name))
		case *ast.VarDecl:
			if n.Init == nil || pos.Before(ast.Start(n.Init)) {
				return prev != nil && prev.Is(":")
			}
			return false
		case *ast.TypeDecl:
			return prev != nil && (prev.Is("=") || prev.Is("|") || prev.Is("["))
		case *ast.Call, *ast.RecordLit:
			return false
		case *ast.Block:
			return c.tokensSayType(prev)
		}
	}
	return c.tokensSayType(prev)
}

// tokensSayType reads a type position from the tokens before the cursor, where the tree holds
// no type yet: after the colon of a parameter or a var, or after the = or | of a type
// declaration.
func (c *completer) tokensSayType(prev *token.Token) bool {
	if prev == nil {
		return false
	}
	switch {
	case prev.Is(":"):
		return c.afterParamName(prev)
	case prev.Is("="):
		return c.lineStartsWith(prev, "type")
	case prev.Is("|"):
		// a union being written: the tokens before the bar are a type that follows the colon
		// of a parameter or a var, or the = of a type declaration
		return c.lineStartsWith(prev, "type") || c.unionAfterColon(prev)
	}
	return false
}

// unionAfterColon reports whether the bar at t follows a type written after the colon of a
// parameter or a var, reading back over the tokens a type is made of.
func (c *completer) unionAfterColon(bar *token.Token) bool {
	toks := c.toks
	i := -1
	for j := range toks {
		if &toks[j] == bar {
			i = j
		}
	}
	for i--; i >= 0; i-- {
		t := toks[i]
		switch {
		case t.Kind == token.Newline:
		case t.Kind == token.Ident, t.Is("."), t.Is("["), t.Is("]"), t.Is(","), t.Is("|"), t.Is("("), t.Is(")"), t.Is("null"), t.Is("void"), t.Is("func"):
		case t.Is(":"):
			return c.afterParamName(&toks[i])
		default:
			return false
		}
	}
	return false
}

// afterParamName reports whether the colon before the cursor follows a name in a parameter list
// of a function literal, or in a var declaration, which the tree may not hold yet.
func (c *completer) afterParamName(colon *token.Token) bool {
	toks := c.toks
	for i := range toks {
		if &toks[i] != colon || i < 2 {
			continue
		}
		return toks[i-1].Kind == token.Ident && (toks[i-2].Is("(") || toks[i-2].Is(",") || toks[i-2].Is("var"))
	}
	return false
}

// lineStartsWith reports whether the line holding t begins with the keyword word.
func (c *completer) lineStartsWith(t *token.Token, word string) bool {
	toks := c.toks
	for i := range toks {
		if &toks[i] != t {
			continue
		}
		for j := i; j >= 0; j-- {
			if toks[j].Kind == token.Newline || j > 0 && toks[j-1].EndPos.Line < toks[j].Pos.Line {
				return toks[j].Is(word)
			}
		}
		return toks[0].Is(word)
	}
	return false
}

var builtinTypes = []string{"int", "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "byte",
	"float32", "float64", "string", "bool", "null", "void", "never", "Array", "Map", "Set"}

// typeNames completes where a type is written: the types in scope, the type parameters of the
// enclosing declarations, the prelude's types and namespaces, and the built-in names.
func (c *completer) typeNames(prefix string) {
	a := c.a
	for _, name := range builtinTypes {
		c.add(CompletionItem{Label: name, Kind: ItemClass, Detail: "type"})
	}
	for _, o := range a.info.Prelude {
		if o.Kind == check.TypeKind || o.Kind == check.Namespace && hasTypes(o) {
			c.add(c.objectItem(a, o, ""))
		}
	}
	c.moduleTypes()
	nodes := ast.Path(a.module().File, c.d.posOf(c.wordStart))
	for _, n := range nodes {
		var tps []ast.TypeParam
		switch x := n.(type) {
		case *ast.FuncDecl:
			tps = x.TypeParams
		case *ast.TypeDecl:
			tps = x.TypeParams
		}
		for i := range tps {
			c.add(CompletionItem{Label: tps[i].Name, Kind: ItemTypeParameter, Detail: "type parameter"})
		}
	}
	if m := a.module(); m.Resolve != nil {
		for i := len(nodes) - 1; i >= 0; i-- {
			if sc, ok := m.Resolve.Scopes[nodes[i]]; ok {
				for _, b := range sc.Names {
					if b.Type {
						c.add(CompletionItem{Label: b.Name, Kind: ItemClass, Detail: "type"})
					}
				}
			}
		}
	}
}

// moduleTypes completes the types and modules the module binds at its top level, under the
// names it binds them by: its own types, and imports under their aliases.
func (c *completer) moduleTypes() {
	a := c.a
	m := a.module()
	if m.Resolve == nil {
		return
	}
	sc := m.Resolve.Scopes[m.File]
	if sc == nil {
		return
	}
	for _, b := range sc.Names {
		o := a.info.Decls[b.Node]
		if o == nil || o.Kind != check.TypeKind && o.Kind != check.Module {
			continue
		}
		item := c.objectItem(a, o, "")
		item.Label = b.Name
		c.add(item)
	}
}

// hasTypes reports whether a namespace holds a type, so that it leads somewhere in a type path.
func hasTypes(o *check.Object) bool {
	for _, m := range o.Members() {
		if m.Kind == check.TypeKind {
			return true
		}
	}
	return false
}

var isKinds = []string{"null", "void", "func", "bool", "string", "Bytes", "Html", "Array", "Map", "Set",
	"int", "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "byte", "float32", "float64"}

// isTargets completes after is: the kinds is tests and the types in scope.
func (c *completer) isTargets() {
	for _, k := range isKinds {
		c.add(CompletionItem{Label: k, Kind: ItemKeyword})
	}
	a := c.a
	for _, o := range a.info.Prelude {
		if o.Kind == check.TypeKind || o.Kind == check.Namespace && hasTypes(o) {
			c.add(c.objectItem(a, o, ""))
		}
	}
	c.moduleTypes()
}

// ---- arguments and entries

// namedArgs completes the parameters not yet given, as name:, when the cursor starts an argument.
func (c *completer) namedArgs(prev *token.Token) {
	if prev == nil || !(prev.Is("(") || prev.Is(",")) {
		return
	}
	a, id := c.fixup("")
	if id == nil {
		return
	}
	var call *ast.Call
	ast.Inspect(a.module().File, func(n ast.Node) bool {
		if x, ok := n.(*ast.Call); ok {
			for _, arg := range x.Args {
				if arg.Value == ast.Expr(id) {
					call = x
				}
			}
		}
		return call == nil
	})
	if call == nil {
		return
	}
	ci := a.info.Calls[call]
	if ci == nil || len(ci.Names) == 0 {
		return
	}
	given := map[int]bool{}
	for i, arg := range call.Args {
		if arg.Value == ast.Expr(id) {
			continue
		}
		if i < len(ci.Slots) && ci.Slots[i] >= 0 {
			given[ci.Slots[i]] = true
		}
	}
	for i, name := range ci.Names {
		if given[i] {
			continue
		}
		// sorted first, in the order of the parameters
		item := CompletionItem{Label: name + ":", Kind: ItemProperty, InsertText: name + ": ", SortText: fmt.Sprintf("0%03d", i)}
		if i < len(ci.Fn.Params) {
			item.Detail = types.String(ci.Fn.Params[i])
		}
		c.add(item)
	}
}

// recordFields completes the fields a record literal does not give yet, when the cursor starts
// an entry.
func (c *completer) recordFields(prev *token.Token) {
	if prev == nil || !(prev.Is("{") || prev.Is(",")) {
		return
	}
	a, id := c.fixup("")
	if id == nil {
		return
	}
	var lit *ast.RecordLit
	var entry *ast.RecordEntry
	ast.Inspect(a.module().File, func(n ast.Node) bool {
		if x, ok := n.(*ast.RecordLit); ok {
			for i := range x.Entries {
				if x.Entries[i].Value == ast.Expr(id) || x.Entries[i].Name == placeholder {
					lit, entry = x, &x.Entries[i]
				}
			}
		}
		return lit == nil
	})
	if lit == nil {
		return
	}
	named, ok := types.Expand(a.info.Types[lit]).(types.Named)
	if !ok || named.Decl.Kind != types.RecordDecl {
		return
	}
	given := map[string]bool{}
	for i := range lit.Entries {
		if &lit.Entries[i] != entry && lit.Entries[i].Name != "" {
			given[lit.Entries[i].Name] = true
		}
	}
	for i, f := range named.FieldsOf() {
		if given[f.Name] {
			continue
		}
		// sorted first, in the order of the fields
		item := CompletionItem{Label: f.Name, Kind: ItemField, Detail: types.String(f.Type), InsertText: f.Name + ": ", SortText: fmt.Sprintf("0%03d", i)}
		if f.Default {
			item.LabelDetails = &CompletionLabelDetails{Detail: " = …"}
		}
		c.add(item)
	}
}

// ---- imports

// importNames completes the exports of a module between an import's braces.
func (c *completer) importNames() bool {
	pos := c.d.posOf(c.wordStart)
	for _, imp := range c.a.module().File.Imports {
		if !imp.PathEnd.Before(pos) || imp.End.Before(pos) && !(pos.Line == imp.End.Line) {
			continue
		}
		if imp.Namespace != "" || imp.Module < 0 || imp.Module >= len(c.a.prog.Modules) {
			return true
		}
		m := c.a.prog.Modules[imp.Module]
		mi, ok := c.a.info.Modules[m.File.Path]
		if !ok {
			return true
		}
		given := map[string]bool{}
		for _, n := range imp.Names {
			given[n.Name] = true
		}
		for _, o := range mi.Exports {
			if !given[o.Name] {
				c.add(c.objectItem(c.a, o, ""))
			}
		}
		return true
	}
	return false
}

// importPaths completes the path of an import: the .ant files and directories at the path
// typed so far, relative to the document.
func (c *completer) importPaths(t *token.Token) {
	typed := c.d.text[t.Off+1 : c.off]
	dir := path.Dir(typed)
	if typed == "" || !strings.HasPrefix(typed, ".") && !strings.HasPrefix(typed, "/") {
		dir = "."
	}
	base := filepath.Dir(c.d.path)
	full := filepath.Join(base, filepath.FromSlash(dir))
	if path.IsAbs(dir) {
		full = filepath.FromSlash(dir)
	}
	entries, err := os.ReadDir(full)
	if err != nil {
		return
	}
	if typed == "" {
		c.add(CompletionItem{Label: "./", Kind: ItemFolder})
		c.add(CompletionItem{Label: "../", Kind: ItemFolder})
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		switch {
		case e.IsDir():
			c.add(CompletionItem{Label: name + "/", Kind: ItemFolder, InsertText: name + "/", CommitCharacters: []string{"/"}})
		case strings.HasSuffix(name, ".ant") && filepath.Join(full, name) != c.d.path:
			c.add(CompletionItem{Label: name, Kind: ItemFile})
		}
	}
}

// ---- markup

// markup completes inside an HTML expression: element names after <, the attributes of the
// element in a start tag, the end tag after </, and expressions in an interpolation.
func (c *completer) markup(t *token.Token) {
	ctx := scanMarkup(c.d.text[t.Off:c.off])
	switch ctx.kind {
	case inTagName:
		for _, el := range htmldata.Elements {
			if value.HtmlObsolete[el.Name] {
				continue
			}
			item := CompletionItem{Label: el.Name, Kind: ItemProperty, Documentation: markdown(el.Doc)}
			if value.HtmlVoid[el.Name] {
				item.Detail = "void element"
			}
			c.add(item)
		}
	case inAttrs, inAttrName:
		for _, name := range htmldata.Attributes(ctx.element) {
			if value.HtmlCodeAttr(name) {
				continue
			}
			item := CompletionItem{Label: name, Kind: ItemProperty}
			switch {
			case value.HtmlBoolean[name]:
				item.Detail = "boolean attribute"
			case value.HtmlURL[name]:
				item.Detail = "URL attribute"
			}
			c.add(item)
		}
	case inEndTag:
		if n := len(ctx.open); n > 0 {
			name := ctx.open[n-1]
			c.add(CompletionItem{Label: name + ">", Kind: ItemProperty, FilterText: name, InsertText: name + ">"})
		}
	case inInterp:
		c.expressionInMarkup()
	}
}

// expressionInMarkup completes an expression in an interpolation, where the tokens of the
// document do not reach: the names in scope, and the members after a dot.
func (c *completer) expressionInMarkup() {
	if c.wordStart >= 2 && c.d.text[c.wordStart-1] == '.' {
		dot := token.Token{Kind: token.Op, Text: ".", Off: c.wordStart - 1, End: c.wordStart}
		if c.wordStart >= 2 && c.d.text[c.wordStart-2] == '?' {
			dot = token.Token{Kind: token.Op, Text: "?.", Off: c.wordStart - 2, End: c.wordStart}
		}
		c.members(&dot)
		return
	}
	c.scopeNames()
	for _, k := range exprKeywords {
		c.add(CompletionItem{Label: k, Kind: ItemKeyword})
	}
}
