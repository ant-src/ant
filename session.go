package ant

import (
	"context"
	"fmt"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/parser"
	"github.com/ant-src/ant/value"
)

// A Session evaluates one input at a time on a persistent module, which is what the REPL is
// built on. Each input may hold declarations and statements; declarations accumulate, a name
// may be declared again, imports stay in force, and goroutines keep running between inputs.
type Session struct {
	cfg     config
	prelude map[string]value.Value
	extra   []string
	s       *interp.Session
	imports []string
}

// NewSession creates a session. The options are those of Load and Run.
func NewSession(ctx context.Context, opts ...Option) (*Session, error) {
	cfg := defaultConfig().with(opts)
	table, extra, _, err := cfg.prelude()
	if err != nil {
		return nil, &LoadError{Diagnostics: []Diagnostic{{Message: err.Error()}}}
	}
	s := interp.NewSession(ctx, interp.Config{
		Prelude: table, Stdin: cfg.stdin, Stdout: cfg.stdout, Stderr: cfg.stderr, Args: cfg.args,
		Env: cfg.env,
	})
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
	prefix := strings.Join(s.imports, "; ")
	full := src
	if prefix != "" {
		full = prefix + "\n" + src
	}
	prog, diags := loader.Load("<input>", loader.Config{
		FS: s.cfg.fsys, Sources: withSource(s.cfg.sources, "<input>", full),
		Mode: parser.Session, Extra: s.extra, Session: s.s.Fields,
	})
	if len(diags) > 0 {
		if prefix != "" {
			for i := range diags {
				if diags[i].Path == "<input>" && diags[i].Pos.Line > 0 {
					diags[i].Pos.Line--
				}
			}
		}
		return nil, &LoadError{Diagnostics: convert(diags)}
	}
	v, err := s.s.Eval(prog)
	if err != nil {
		return nil, convertRunError(err)
	}
	for _, imp := range prog.Root.File.Imports {
		s.remember(imp)
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
}

func withSource(sources map[string]string, path, src string) map[string]string {
	out := make(map[string]string, len(sources)+1)
	for k, v := range sources {
		out[k] = v
	}
	out[path] = src
	return out
}

// Value returns a name the session has declared.
func (s *Session) Value(name string) (value.Value, bool) { return s.s.Lookup(name) }

// Close stops the session's goroutines and flushes its output.
func (s *Session) Close() { s.s.Instance().Close() }
