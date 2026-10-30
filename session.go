package ant

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/internal/lsp"
	"github.com/ant-src/ant/internal/token"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/parser"
	"github.com/ant-src/ant/value"
)

// A Session evaluates one input at a time on a persistent module, which is what the REPL is
// built on. Each input may hold declarations and statements; declarations accumulate, a name
// may be declared again, imports stay in force, and goroutines keep running between inputs.
type Session struct {
	cfg         config
	prelude     map[string]value.Value
	extra       []string
	s           *interp.Session
	imports     []string      // the imports in force, as text, to keep each once
	importDecls []*ast.Import // the same, parsed, which every later input takes before its own
	// decls is the text of every input that declared or imported something, in order, which
	// Complete analyzes ahead of an input so the names it offers carry their types. An input
	// whose every name a later input declares again is dropped, so the text stays bounded.
	decls []declInput
}

type declInput struct {
	src   string
	names []string
}

// NewSession creates a session. The options are those of Load and Run.
func NewSession(ctx context.Context, opts ...Option) (*Session, error) {
	cfg := defaultConfig().with(opts)
	table, extra, _, err := cfg.prelude()
	if err != nil {
		return nil, &LoadError{Diagnostics: []Diagnostic{{Message: err.Error()}}}
	}
	s := interp.NewSession(ctx, cfg.run(table))
	return &Session{cfg: cfg, prelude: table, extra: extra, s: s}, nil
}

// Incomplete reports whether src ends too early to be one input, so a REPL should keep
// reading: an open brace, bracket, string, or template.
func Incomplete(src string) bool {
	_, err := parser.Parse("<input>", src, parser.Session)
	if err == nil {
		return false
	}
	return err.(*parser.Error).Incomplete
}

// Eval evaluates one input and returns the value of its last expression, or void. The input
// is a script body of its own: the value it returns, by return or ?? return, is its result, and
// an Err it returns, as try does, comes back as a *ResultError after being reported. A syntax
// or scope error comes back as a *LoadError, a panic as a *PanicError after being reported.
func (s *Session) Eval(src string) (value.Value, error) {
	prog, diags := loader.Load("<input>", loader.Config{
		FS: s.cfg.fsys, Sources: withSource(s.cfg.sources, "<input>", src),
		Mode: parser.Session, Extra: s.extra, Session: s.s.Fields, Imports: s.importDecls,
	})
	if len(diags) > 0 {
		return nil, &LoadError{Diagnostics: convert(diags)}
	}
	if names := declaredNames(prog.Root.File, len(s.importDecls)); len(names) > 0 {
		s.decls = slices.DeleteFunc(s.decls, func(d declInput) bool {
			return !slices.ContainsFunc(d.names, func(n string) bool { return !slices.Contains(names, n) })
		})
		s.decls = append(s.decls, declInput{src, names})
	}
	v, err := s.s.Eval(prog)
	// the imports stay in force even when the input fails, as its other declarations do
	for _, imp := range prog.Root.File.Imports[len(s.importDecls):] {
		s.remember(imp)
	}
	if err != nil {
		return nil, convertRunError(err)
	}
	return v, nil
}

// remember keeps an import in force for later inputs.
func (s *Session) remember(imp *ast.Import) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "import %q ", imp.Path)
	if imp.Namespace != "" {
		sb.WriteString(imp.Namespace)
	} else {
		sb.WriteString("{ ")
		for i, n := range imp.Names {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(n.Name)
			if n.Alias != n.Name {
				sb.WriteString(": " + n.Alias)
			}
		}
		sb.WriteString(" }")
	}
	text := sb.String()
	for _, have := range s.imports {
		if have == text {
			return
		}
	}
	s.imports = append(s.imports, text)
	s.importDecls = append(s.importDecls, imp)
}

func withSource(sources map[string]string, path, src string) map[string]string {
	out := make(map[string]string, len(sources)+1)
	for k, v := range sources {
		out[k] = v
	}
	out[path] = src
	return out
}

// declaredNames lists what an input declares: its functions, with their owner, its types, its
// variables, and the names its own imports bind, beyond the known imports in force.
func declaredNames(f *ast.File, known int) []string {
	var names []string
	for _, imp := range f.Imports[known:] {
		if imp.Namespace != "" {
			names = append(names, imp.Namespace)
		}
		for _, n := range imp.Names {
			names = append(names, n.Alias)
		}
	}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			if d.Owner != "" {
				names = append(names, d.Owner+"."+d.Name)
			} else {
				names = append(names, d.Name)
			}
		case *ast.TypeDecl:
			names = append(names, d.Name)
		case *ast.VarDecl:
			names = targetNames(names, d.Target)
		}
	}
	for _, st := range f.Stmts {
		if v, ok := st.(*ast.VarDecl); ok {
			names = targetNames(names, v.Target)
		}
	}
	return names
}

func targetNames(names []string, t ast.Target) []string {
	switch t := t.(type) {
	case *ast.NameTarget:
		if t.Name != "_" {
			names = append(names, t.Name)
		}
	case *ast.RecordTarget:
		for _, f := range t.Fields {
			names = targetNames(names, f.Target)
		}
	case *ast.ArrayTarget:
		for _, e := range t.Elems {
			names = targetNames(names, e)
		}
		if t.Rest != nil {
			names = targetNames(names, t.Rest)
		}
	}
	return names
}

// A Completion is one thing that can be written at a position of an input, as Complete lists
// them.
type Completion struct {
	Label string // the name, as a list shows it
	// Insert is the text to write: the label, or more where the position asks for it, as a
	// named argument gets its colon and a directory its slash.
	Insert string
	// Kind says what it is: "variable", "function", "field", "type", "namespace", "module",
	// "constant", "keyword", "file", or "folder".
	Kind string
	// Detail is what hover would show beside it: a function's parameters and result, a
	// variable's type, or what a name is.
	Detail string
	// Preferred marks an item the position asks for first: a parameter the call has not been
	// given, or a field the record literal lacks.
	Preferred bool
}

// Complete lists what can be written at byte offset off of src, the input being typed, as the
// language server's completion does (spec/lsp.md section 4): the names the session has
// declared, with their types, the imports in force, the prelude, and what the position allows,
// as the members after a dot, the functions that take the left side after a pipe, the
// parameters of a call, and the keywords. After a dot it also offers the fields of the record
// a variable of the session holds at runtime, read through its fields, which a union's static
// type does not promise. The items are unsorted.
func (s *Session) Complete(src string, off int) []Completion {
	if off < 0 || off > len(src) {
		return nil
	}
	text, base := s.transcript(src)
	items := lsp.Complete(s.analysisConfig(), "<input>", text, base+off)
	out := make([]Completion, 0, len(items))
	seen := map[string]bool{}
	for _, it := range items {
		seen[it.Label] = true
		c := Completion{Label: it.Label, Insert: it.InsertText, Detail: it.Detail, Preferred: it.SortText != ""}
		if c.Insert == "" {
			c.Insert = it.Label
		}
		switch it.Kind {
		case lsp.ItemVariable:
			c.Kind = "variable"
		case lsp.ItemFunction, lsp.ItemMethod:
			c.Kind = "function"
		case lsp.ItemField, lsp.ItemProperty:
			c.Kind = "field"
		case lsp.ItemClass, lsp.ItemInterface, lsp.ItemTypeParameter:
			c.Kind = "type"
		case lsp.ItemModule:
			c.Kind = "module"
			if it.Detail == "namespace" {
				c.Kind = "namespace"
			}
		case lsp.ItemConstant:
			c.Kind = "constant"
		case lsp.ItemKeyword:
			c.Kind = "keyword"
		case lsp.ItemFile:
			c.Kind = "file"
		case lsp.ItemFolder:
			c.Kind = "folder"
		}
		out = append(out, c)
	}
	return s.runtimeFields(src, off, seen, out)
}

// runtimeFields adds the fields of the record a path of names reaches at runtime: a variable
// of the session and the fields read from it, before the dot the cursor's word follows. A
// session evaluates without types, so a field the static type cannot promise, as the value of
// a Result that is an Ok, is still there to read, and offered with the type it holds.
func (s *Session) runtimeFields(src string, off int, seen map[string]bool, out []Completion) []Completion {
	start := off
	for start > 0 && token.IsIdentByte(src[start-1]) {
		start--
	}
	var names []string
	i := start
	for {
		switch {
		case i >= 2 && src[i-2:i] == "?.":
			i -= 2
		case i >= 1 && src[i-1] == '.':
			i--
		default:
			if len(names) == 0 {
				return out
			}
			if i > 0 && (token.IsIdentByte(src[i-1]) || strings.IndexByte(")]}\"`'", src[i-1]) >= 0) {
				return out // the path begins in an expression, not at a variable
			}
			slices.Reverse(names)
			rec := s.recordAt(names)
			if rec == nil {
				return out
			}
			for _, f := range rec.Tag.Fields {
				if seen[f] {
					continue
				}
				fv, _ := rec.Field(f)
				out = append(out, Completion{Label: f, Insert: f, Kind: "field", Detail: typeName(fv)})
			}
			return out
		}
		j := i
		for j > 0 && token.IsIdentByte(src[j-1]) {
			j--
		}
		if j == i {
			return out
		}
		names = append(names, src[j:i])
		i = j
	}
}

// recordAt is the record the session's variable names[0] holds, read through the fields
// names[1:], or nil.
func (s *Session) recordAt(names []string) *value.Record {
	v, ok := s.Value(names[0])
	for _, name := range names[1:] {
		rec, isRec := v.(*value.Record)
		if !ok || !isRec || rec.Tag.Kind != value.RecordTag {
			return nil
		}
		v, ok = rec.Field(name)
	}
	rec, isRec := v.(*value.Record)
	if !ok || !isRec || rec.Tag.Kind != value.RecordTag {
		return nil
	}
	return rec
}

// typeName names the type of a value as the checker would: a record by its type, the rest
// by its kind.
func typeName(v value.Value) string {
	if rec, ok := v.(*value.Record); ok && rec.Tag.Kind == value.RecordTag {
		return rec.Tag.Name
	}
	return value.KindOf(v).String()
}

// transcript is the text of the inputs that declared something, followed by src, and where
// src begins in it: what Complete and TypeOf analyze, so that the names src uses have types.
func (s *Session) transcript(src string) (string, int) {
	var sb strings.Builder
	for _, d := range s.decls {
		sb.WriteString(d.src)
		if !strings.HasSuffix(d.src, "\n") {
			sb.WriteByte('\n')
		}
	}
	base := sb.Len()
	sb.WriteString(src)
	return sb.String(), base
}

// analysisConfig loads a transcript for analysis: as a script, whose statements the checker
// types, in a scope where a name may be declared again.
func (s *Session) analysisConfig() loader.Config {
	return loader.Config{FS: s.cfg.fsys, Sources: s.cfg.sources, Extra: s.extra,
		Session: &ast.ModuleScope{Index: map[string]int{}}}
}

// TypeOf returns the type the checker gives expr, an expression over what the session has
// declared, as the checker writes it. A syntax error, or a type error in expr, comes back as a
// *LoadError, positioned in expr.
func (s *Session) TypeOf(expr string) (string, error) {
	text, base := s.transcript(expr)
	baseLine := strings.Count(text[:base], "\n")
	cfg := s.analysisConfig()
	cfg.Mode = parser.Program
	cfg.Sources = withSource(cfg.Sources, "<input>", text)
	prog, diags := loader.Load("<input>", cfg)
	inExpr := func(ds []loader.Diagnostic) []Diagnostic {
		var out []Diagnostic
		for _, d := range ds {
			if d.Path == "<input>" && d.Pos.Line > baseLine {
				d.Pos.Line -= baseLine
				out = append(out, convert([]loader.Diagnostic{d})...)
			}
		}
		return out
	}
	if prog == nil {
		if ds := inExpr(diags); len(ds) > 0 {
			return "", &LoadError{Diagnostics: ds}
		}
		return "", &LoadError{Diagnostics: convert(diags)}
	}
	info := check.NewInfo()
	if ds := inExpr(check.Program(prog, check.Config{Info: info})); len(ds) > 0 {
		return "", &LoadError{Diagnostics: ds}
	}
	stmts := prog.Root.File.Stmts
	if len(stmts) == 0 {
		return "", &LoadError{Diagnostics: []Diagnostic{{Message: "expected an expression"}}}
	}
	e, ok := stmts[len(stmts)-1].(ast.Expr)
	if !ok || ast.Start(e).Line <= baseLine {
		return "", &LoadError{Diagnostics: []Diagnostic{{Message: "expected an expression"}}}
	}
	t := info.Types[e]
	if t == nil {
		return "", &LoadError{Diagnostics: []Diagnostic{{Message: "the expression has no type"}}}
	}
	return types.String(t), nil
}

// Value returns a name the session has declared.
func (s *Session) Value(name string) (value.Value, bool) { return s.s.Lookup(name) }

// Close stops the session's goroutines and flushes its output.
func (s *Session) Close() { s.s.Instance().Close() }
