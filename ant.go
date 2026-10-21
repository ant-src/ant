// Package ant loads, runs, and checks programs written in Ant.
//
// Load parses a program and every module it imports into an immutable Program, which can be run
// any number of times, concurrently. Program.Run executes it with the options given and returns
// nil when main returns, a *PanicError when the main goroutine panics, or an *ExitError when the
// program calls Os.exit. Program.Check runs the type checker and returns its diagnostics.
//
// The library reads sources through an fs.FS, writes only to the writers it is given, takes its
// arguments and environment from options, and never exits the process.
package ant

import (
	"context"
	"io"
	"io/fs"
	"os"

	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/parser"
	"github.com/ant-src/ant/value"
)

type config struct {
	fsys    fs.FS
	sources map[string]string
	stdout  io.Writer
	stderr  io.Writer
	args    []string
	env     func(name string) (string, bool)
	// host namespaces added and prelude namespaces removed
	namespaces []Namespace
	removed    []string
}

func defaultConfig() config {
	return config{stdout: os.Stdout, stderr: os.Stderr, env: os.LookupEnv}
}

func (c config) with(opts []Option) config {
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

// An Option configures Load, Run, or both. Options given to Run override those given to Load.
type Option func(*config)

// WithFS reads sources from fsys instead of the operating system. Paths are then fs.FS paths:
// slash-separated and relative to the root of fsys.
func WithFS(fsys fs.FS) Option { return func(c *config) { c.fsys = fsys } }

// WithSource supplies the text of the module at path, overriding whatever the file system
// holds, so a program can be loaded from a string.
func WithSource(path, src string) Option {
	return func(c *config) {
		if c.sources == nil {
			c.sources = map[string]string{}
		}
		c.sources[path] = src
	}
}

// WithStdout sets where print writes. The default is os.Stdout.
func WithStdout(w io.Writer) Option { return func(c *config) { c.stdout = w } }

// WithStderr sets where panic reports are written. The default is os.Stderr.
func WithStderr(w io.Writer) Option { return func(c *config) { c.stderr = w } }

// WithArgs sets what Os.args returns.
func WithArgs(args []string) Option { return func(c *config) { c.args = args } }

// WithEnv sets what Os.env consults. The default is os.LookupEnv.
func WithEnv(lookup func(name string) (string, bool)) Option {
	return func(c *config) { c.env = lookup }
}

// A Program is a loaded, parsed, and resolved module graph. It is immutable.
type Program struct {
	prog    *loader.Program
	cfg     config
	prelude map[string]value.Value
	nsDecls []check.HostNamespace
}

// Load parses the module at path and every module it imports. It returns a *LoadError when a
// module has a syntax error, cannot be found, takes part in an import cycle, or when the root
// module does not declare main.
func Load(path string, opts ...Option) (*Program, error) {
	cfg := defaultConfig().with(opts)
	table, extra, nsDecls, err := cfg.prelude()
	if err != nil {
		return nil, &LoadError{Diagnostics: []Diagnostic{{Message: err.Error()}}}
	}
	prog, diags := loader.Load(path, loader.Config{
		FS: cfg.fsys, Sources: cfg.sources, Mode: parser.Program, RequireMain: true, Extra: extra,
	})
	if len(diags) > 0 {
		return nil, &LoadError{Diagnostics: convert(diags)}
	}
	return &Program{prog: prog, cfg: cfg, prelude: table, nsDecls: nsDecls}, nil
}

func convert(diags []loader.Diagnostic) []Diagnostic {
	out := make([]Diagnostic, len(diags))
	for i, d := range diags {
		out[i] = Diagnostic{Pos: Position{Path: d.Path, Line: d.Pos.Line, Col: d.Pos.Col}, Message: d.Msg}
	}
	return out
}

// Run executes the program: it initializes every module in dependency order, then calls main.
// It returns nil when main returns, a *PanicError when the main goroutine panics, an *ExitError
// when the program calls Os.exit, and ctx.Err() when the context is cancelled first.
func (p *Program) Run(ctx context.Context, opts ...Option) error {
	cfg := p.cfg.with(opts)
	return convertRunError(interp.Run(ctx, p.prog, interp.Config{
		Prelude: p.prelude, Stdout: cfg.stdout, Stderr: cfg.stderr, Args: cfg.args, Env: cfg.env,
	}))
}

// Check type-checks every module of the program, including code nothing calls, and returns the
// errors sorted by position. A program with no diagnostics is well typed.
func (p *Program) Check() []Diagnostic {
	diags := check.Program(p.prog, check.Config{Namespaces: p.nsDecls, Removed: p.cfg.removed})
	out := make([]Diagnostic, len(diags))
	for i, d := range diags {
		out[i] = Diagnostic{Pos: Position{Path: d.Path, Line: d.Pos.Line, Col: d.Pos.Col}, Message: d.Msg}
	}
	return out
}
