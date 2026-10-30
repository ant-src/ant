package ant

import (
	"context"
	"fmt"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/internal/lsp"
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
	// Complete analyzes ahead of an input so the names it offers carry their types.
	decls []string
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

// Eval evaluates one input and returns the value of its last expression, or void. A syntax
// or scope error comes back as a *LoadError, a panic as a *PanicError after being reported.
func (s *Session) Eval(src string) (value.Value, error) {
	prog, diags := loader.Load("<input>", loader.Config{
		FS: s.cfg.fsys, Sources: withSource(s.cfg.sources, "<input>", src),
		Mode: parser.Session, Extra: s.extra, Session: s.s.Fields, Imports: s.importDecls,
	})
	if len(diags) > 0 {
		return nil, &LoadError{Diagnostics: convert(diags)}
	}
	if declares(prog.Root.File, len(s.importDecls)) {
		s.decls = append(s.decls, src)
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

// declares reports whether an input declares anything: a function, a type, a variable, or an
// import of its own beyond the first known inputs in force.
func declares(f *ast.File, known int) bool {
	if len(f.Decls) > 0 || len(f.Imports) > known {
		return true
	}
	for _, st := range f.Stmts {
		if _, ok := st.(*ast.VarDecl); ok {
			return true
		}
	}
	return false
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
// parameters of a call, and the keywords. The items are unsorted.
func (s *Session) Complete(src string, off int) []Completion {
	if off < 0 || off > len(src) {
		return nil
	}
	var sb strings.Builder
	for _, d := range s.decls {
		sb.WriteString(d)
		if !strings.HasSuffix(d, "\n") {
			sb.WriteByte('\n')
		}
	}
	base := sb.Len()
	sb.WriteString(src)
	cfg := loader.Config{FS: s.cfg.fsys, Sources: s.cfg.sources, Mode: parser.Session, Extra: s.extra,
		Session: &ast.ModuleScope{Index: map[string]int{}}}
	items := lsp.Complete(cfg, "<input>", sb.String(), base+off)
	out := make([]Completion, 0, len(items))
	for _, it := range items {
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
	return out
}

// Value returns a name the session has declared.
func (s *Session) Value(name string) (value.Value, bool) { return s.s.Lookup(name) }

// Close stops the session's goroutines and flushes its output.
func (s *Session) Close() { s.s.Instance().Close() }
