package ant

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/lexer"
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
	// decls holds the latest declaration of each name the session declared, in order, without
	// imports, which come from importDecls: what Complete and TypeOf analyze ahead of an input
	// so that the names it uses carry their types. A function or a type is kept as written; a
	// variable as a stub of the type it had when declared, var s: T = panic(""), since a later
	// declaration of a name may read the one it replaces, as var s = [s] does.
	decls []declEntry
	// lastExpr is the source of the expression whose value _ reads, or "" when the last
	// result came from elsewhere, as a return; TypeOf puts it in place of _.
	lastExpr string
}

type declEntry struct {
	key, src string
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

// Eval evaluates one input and returns the value of its last expression, or void, which _
// reads in the inputs after it unless it is void (section 10, Scripts). The input
// is a script body of its own: the value it returns, by return or ?? return, is its result, and
// an Err it returns, as try does, comes back as a *ResultError after being reported. A syntax
// or scope error comes back as a *LoadError, a panic as a *PanicError after being reported.
func (s *Session) Eval(src string) (value.Value, error) { return s.evalAt("<input>", src) }

// EvalFile evaluates the file at path as an input of the session, its relative imports
// resolved against its own directory, and returns what Eval returns.
func (s *Session) EvalFile(path string) (value.Value, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, &LoadError{Diagnostics: []Diagnostic{{Pos: Position{Path: path}, Message: err.Error()}}}
	}
	return s.evalAt(path, string(src))
}

func (s *Session) evalAt(path, src string) (value.Value, error) {
	prog, diags := loader.Load(path, loader.Config{
		FS: s.cfg.fsys, Sources: withSource(s.cfg.sources, path, src),
		Mode: parser.Session, Extra: s.extra, Session: s.s.Fields, Imports: s.importDecls,
	})
	if len(diags) > 0 {
		return nil, &LoadError{Diagnostics: convert(diags)}
	}
	f := prog.Root.File
	own := f.Imports[len(s.importDecls):]
	if len(f.Decls) > 0 {
		s.record(path, src, f)
	}
	v, err := s.s.Eval(prog)
	// the imports stay in force even when the input fails, as its other declarations do
	for _, imp := range own {
		s.remember(imp, prog)
	}
	if err != nil {
		return nil, convertRunError(err)
	}
	if v != value.Void {
		s.s.SetLast(v)
		s.lastExpr = ""
		if n := len(f.Stmts); n > 0 {
			if e, ok := f.Stmts[n-1].(ast.Expr); ok {
				if _, ret := e.(*ast.Return); !ret {
					s.lastExpr = strings.TrimSpace(src[offsetOf(src, ast.Start(e)):])
				}
			}
		}
	}
	return v, nil
}

// offsetOf is the byte offset of a position in src.
func offsetOf(src string, p ast.Pos) int {
	off := 0
	for line := 1; line < p.Line; line++ {
		i := strings.IndexByte(src[off:], '\n')
		if i < 0 {
			return len(src)
		}
		off += i + 1
	}
	return min(off+p.Col-1, len(src))
}

// record keeps the declarations of an input, f parsed from src, for analysis: each as the
// latest of its name. The types of its variables come from analyzing it after what the
// session declared before it.
func (s *Session) record(path, src string, f *ast.File) {
	types := s.varTypes(path, src, f)
	put := func(key, text string) {
		s.decls = slices.DeleteFunc(s.decls, func(d declEntry) bool { return d.key == key })
		s.decls = append(s.decls, declEntry{key, text})
	}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			key := d.Name
			if d.Owner != "" {
				key = d.Owner + "." + d.Name
			}
			put(key, src[offsetOf(src, d.Pos):offsetOf(src, d.End)])
		case *ast.TypeDecl:
			put(d.Name, src[offsetOf(src, d.Pos):offsetOf(src, d.End)])
		case *ast.VarDecl:
			text := src[offsetOf(src, d.Pos):offsetOf(src, d.End)]
			ast.Inspect(d.Target, func(n ast.Node) bool {
				nt, ok := n.(*ast.NameTarget)
				if !ok || nt.Name == "_" {
					return true
				}
				if t, ok := types[nt.Name]; ok {
					put(nt.Name, fmt.Sprintf("var %s: %s = panic(\"\")", nt.Name, t))
				} else {
					put(nt.Name, text) // a type the checker cannot write; the declaration itself
				}
				return true
			})
		}
	}
}

// renamed is the suffix varTypes gives a variable an input declares again, so that what the
// input reads of the name before is the declaration it replaces, as in var s = [s].
const renamed = "__ant_new"

// varTypes analyzes src, the input at path parsed as f, after what the session declared, and
// returns the types the checker gives the variables it declares, where they can be written
// back.
func (s *Session) varTypes(path, src string, f *ast.File) map[string]string {
	known := map[string]bool{}
	for _, d := range s.decls {
		known[d.key] = true
	}
	var spans [][2]int
	for _, d := range f.Decls {
		if v, ok := d.(*ast.VarDecl); ok {
			ast.Inspect(v.Target, func(n ast.Node) bool {
				if nt, ok := n.(*ast.NameTarget); ok && known[nt.Name] {
					end := offsetOf(src, nt.Pos) + len(nt.Name)
					spans = append(spans, [2]int{offsetOf(src, nt.Pos), end})
				}
				return true
			})
		}
	}
	for i := len(spans) - 1; i >= 0; i-- {
		src = src[:spans[i][1]] + renamed + src[spans[i][1]:]
	}
	text, base := s.transcript(src)
	baseLine := strings.Count(text[:base], "\n")
	cfg := s.analysisConfig()
	cfg.Mode, cfg.Recover = parser.Program, true
	cfg.Sources = withSource(cfg.Sources, path, text)
	prog, _ := loader.Load(path, cfg)
	out := map[string]string{}
	if prog == nil {
		return out
	}
	info := check.NewInfo()
	check.Program(prog, check.Config{Info: info})
	for _, d := range prog.Root.File.Decls {
		v, ok := d.(*ast.VarDecl)
		if !ok || v.Pos.Line <= baseLine {
			continue
		}
		ast.Inspect(v.Target, func(n ast.Node) bool {
			if nt, ok := n.(*ast.NameTarget); ok {
				if o := info.Decls[nt]; o != nil && o.Type() != nil && !types.IsInvalid(o.Type()) {
					if t := types.String(o.Type()); !strings.Contains(t, "?") {
						out[strings.TrimSuffix(nt.Name, renamed)] = t
					}
				}
			}
			return true
		})
	}
	return out
}

// remember keeps an import in force for later inputs, by the path of the module it loaded, so
// that an input read from another directory, as :load reads one, imports the same module.
func (s *Session) remember(imp *ast.Import, prog *loader.Program) {
	c := *imp
	if imp.Module >= 0 && imp.Module < len(prog.Modules) {
		c.Path = prog.Modules[imp.Module].Path
	}
	imp = &c
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
// The imports in force come with analysisConfig.
func (s *Session) transcript(src string) (string, int) {
	var sb strings.Builder
	for _, d := range s.decls {
		sb.WriteString(d.src)
		sb.WriteByte('\n')
	}
	base := sb.Len()
	sb.WriteString(src)
	return sb.String(), base
}

// analysisConfig loads a transcript for analysis: as a script, whose statements the checker
// types, in a scope where a name may be declared again, with the imports in force.
func (s *Session) analysisConfig() loader.Config {
	return loader.Config{FS: s.cfg.fsys, Sources: s.cfg.sources, Extra: s.extra,
		Session: &ast.ModuleScope{Index: map[string]int{}}, Imports: s.importDecls}
}

// typeVar is the variable TypeOf gives the expression to, so that any expression, markup too,
// is a value used.
const (
	typeVar  = "__ant_type__"
	typeDecl = "var " + typeVar + " = "
)

// TypeOf returns the type the checker gives expr, an expression over what the session has
// declared, as the checker writes it; _ in it stands for the expression whose value it reads.
// A syntax error, or a type error in expr, comes back as a *LoadError, positioned in expr.
func (s *Session) TypeOf(expr string) (string, error) {
	expr, err := s.withLast(expr)
	if err != nil {
		return "", err
	}
	// a statement first, so that the transcript is a script whose statements the checker
	// types, and then the expression, which nothing after it can run on into
	text, base := s.transcript("void\n" + typeDecl + expr)
	baseLine := strings.Count(text[:base], "\n") + 1
	cfg := s.analysisConfig()
	cfg.Mode = parser.Program
	cfg.Sources = withSource(cfg.Sources, "<input>", text)
	prog, diags := loader.Load("<input>", cfg)
	inExpr := func(ds []loader.Diagnostic) []Diagnostic {
		var out []Diagnostic
		for _, d := range ds {
			if d.Path == "<input>" && d.Pos.Line > baseLine {
				d.Pos.Line -= baseLine
				if d.Pos.Line == 1 {
					d.Pos.Col = max(1, d.Pos.Col-len(typeDecl))
				}
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
	var init ast.Expr
	for _, st := range prog.Root.File.Stmts {
		if v, ok := st.(*ast.VarDecl); ok {
			if t, ok := v.Target.(*ast.NameTarget); ok && t.Name == typeVar {
				init = v.Init
			}
		}
	}
	if init == nil || info.Types[init] == nil {
		return "", &LoadError{Diagnostics: []Diagnostic{{Message: "the expression has no type"}}}
	}
	return types.String(info.Types[init]), nil
}

// withLast puts the expression whose value _ reads in place of each _ that expr reads.
func (s *Session) withLast(expr string) (string, error) {
	toks, _, _ := lexer.LexRecover(expr, false)
	var uses []token.Token
	for _, t := range toks {
		if t.Kind == token.Ident && t.Text == "_" {
			uses = append(uses, t)
		}
	}
	if len(uses) == 0 {
		return expr, nil
	}
	if s.lastExpr == "" {
		if v, ok := s.Value("_"); ok && strings.TrimSpace(expr) == "_" {
			return "", &LoadError{Diagnostics: []Diagnostic{{Message: "_ holds a " + typeName(v) + "; its type is not known, as it was returned"}}}
		}
		return "", &LoadError{Diagnostics: []Diagnostic{{Message: "_ holds no result whose type is known"}}}
	}
	for i := len(uses) - 1; i >= 0; i-- {
		expr = expr[:uses[i].Off] + "(" + s.lastExpr + ")" + expr[uses[i].End:]
	}
	return expr, nil
}

// Value returns a name the session has declared, or for _ the last result.
func (s *Session) Value(name string) (value.Value, bool) {
	if name == "_" {
		return s.s.Instance().LastResult()
	}
	return s.s.Lookup(name)
}

// Close stops the session's goroutines and flushes its output.
func (s *Session) Close() { s.s.Instance().Close() }
