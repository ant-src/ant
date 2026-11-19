// Package ant loads, runs, and checks programs written in Ant.
//
// Load parses a program and every module it imports into an immutable Program, which can be run
// any number of times, concurrently. Program.Run executes it with the options given and returns
// nil when the script ends, a *ResultError when it returns an Err, a *PanicError when the main
// goroutine panics, or an *ExitError when the program calls Os.exit. Program.Check runs the type checker and returns its diagnostics.
//
// The library reads sources through an fs.FS, writes only to the writers it is given, takes its
// standard input, arguments and environment from options, and never exits the process. An
// interrupt reaches a program through WithInterrupts: the first goes to the context Os.interrupt
// returns when the program asked for one, and otherwise, or at the second, the run ends as it
// does when the context given to Run is cancelled.
//
// # Calling across the boundary
//
// Program.Start initializes a program and returns an Instance, so a host loads a library, whose
// initialization runs no statements, and calls into it: its Func
// method finds a top-level function, which Func.Call calls with positional arguments, and its
// Value method reads a top-level var or a type's tag. WithNamespace gives a program a set of Go
// functions that Ant code reaches as Name.function, declared in Ant notation so the checker
// can type the calls. A Session evaluates inputs one at a time on a persistent module.
//
// # Values
//
// Ant values cross the boundary as value.Value, which is a Go interface value holding one of
// these representations:
//
//	Ant                         Go
//	int, int64                  int64
//	int8 … int32, uint8 … uint64, byte   the Go integer type of that width (byte is uint8)
//	float32, float64            float32, float64
//	string                      string
//	Bytes                       *value.Bytes  (value.NewBytes adopts, value.CopyBytes copies; View, Snapshot)
//	bool                        bool
//	null, void                  value.Null, value.Void
//	Array[T]                    *value.Array  (value.NewArray, Values, Get, Set, Push)
//	Map[K, V]                   *value.Map    (value.NewMap, Get, Set, Keys, Entries)
//	Set[T]                      *value.Set    (value.NewSet, Has, Add, Remove, Values)
//	a record                    *value.Record (value.NewRecord with its *value.Tag, Field, SetField)
//	a named primitive type      *value.Tagged (Tag and Payload)
//	Result[T, E], Error         records built by value.Ok, value.Err, and value.ErrorValue
//	Entry[K, V], Indexed[T]     records built by value.Entry and value.Indexed
//	a function                  an opaque value.Callable: pass it back to Func.Call or Runtime.Call
//	a type                      *value.Tag, from Instance.Value("Name")
//	a time value                an opaque value.Time: its String is the RFC 9557 form; pass it back
//	a Regex.Pattern             *value.Regex, from value.NewRegex: its String is the literal form
//	a sync object, Channel, Context  an opaque value.Sync: pass it back
//
// A record's tag comes from the instance, since the program declares it: the Example_call test
// builds a Point that way. A host function receives its arguments in the order its signature
// lists them, already of these representations, and returns a value of them; an error it
// returns becomes an Ant panic with the error's message. value.Render renders any value the way
// String.from does.
package ant

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/parser"
	"github.com/ant-src/ant/value"
)

type config struct {
	fsys    fs.FS
	sources map[string]string
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
	args    []string
	program string
	env     func(name string) (string, bool)
	environ func() []string
	// host namespaces added and prelude namespaces removed
	namespaces []Namespace
	removed    []string
	interrupts <-chan struct{}
	foreground func() func(os.Signal) bool
}

func defaultConfig() config {
	return config{stdin: os.Stdin, stdout: os.Stdout, stderr: os.Stderr, env: os.LookupEnv}
}

func (c config) with(opts []Option) config {
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

// run is the interpreter's configuration for a run, an instance, or a session.
func (c config) run(prelude map[string]value.Value) interp.Config {
	return interp.Config{
		Prelude: prelude, Stdin: c.stdin, Stdout: c.stdout, Stderr: c.stderr,
		Args: c.args, Program: c.program, Env: c.env, Environ: c.environ, Interrupts: c.interrupts,
		Foreground: c.foreground,
	}
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

// WithStdin sets what Os.stdin reads. The default is os.Stdin.
func WithStdin(r io.Reader) Option { return func(c *config) { c.stdin = r } }

// WithStdout sets where print, println and Os.stdout write. The default is os.Stdout.
func WithStdout(w io.Writer) Option { return func(c *config) { c.stdout = w } }

// WithStderr sets where panic reports are written. The default is os.Stderr.
func WithStderr(w io.Writer) Option { return func(c *config) { c.stderr = w } }

// WithArgs sets what Os.args returns.
func WithArgs(args []string) Option { return func(c *config) { c.args = args } }

// WithEnv sets what Os.env consults. The default is os.LookupEnv. Os.environ and the
// environment a child of a Cmd inherits stay the process's; WithEnvironment replaces all three.
func WithEnv(lookup func(name string) (string, bool)) Option {
	return func(c *config) { c.env = lookup }
}

// WithInterrupts delivers interrupts to a run, an instance, or a session, one per value the host
// sends on c. The first goes to the context Os.interrupt returns, when the program has called it
// and the context is not done, so the program can drain and exit; otherwise, and for a program
// that never called it, an interrupt ends the run, and Run, the Func.Call or the Session.Eval in
// progress returns ErrInterrupted. The ant command sends one per Ctrl-C or SIGTERM.
func WithInterrupts(c <-chan struct{}) Option { return func(cfg *config) { cfg.interrupts = c } }

// WithForeground tells the host when Cmd.run has a child in the foreground: hold is called as the
// child starts, and the function it returns as the child exits, with the signal that ended it or
// nil. That function reports whether the run gets an interrupt for it, which Cmd.run delivers
// before it returns, as WithInterrupts would. The ant command holds back Ctrl-C meanwhile, which
// the child gets from the terminal, and passes it on only when the child died of it. Without it
// Cmd.run holds nothing back.
func WithForeground(hold func() (release func(ended os.Signal) (interrupt bool))) Option {
	return func(cfg *config) { cfg.foreground = hold }
}

// ErrInterrupted is what Run returns when an interrupt ended the run: one the program had not
// asked for, or the second. ExitStatus gives it 130, as a shell gives a process a signal ended.
var ErrInterrupted = interp.ErrInterrupted

// WithEnvironment gives the run its whole environment: what Os.env and Os.environ see, and
// what a child of a Cmd inherits when it is given no env of its own.
func WithEnvironment(vars map[string]string) Option {
	return func(c *config) {
		c.env = func(name string) (string, bool) { v, ok := vars[name]; return v, ok }
		c.environ = func() []string {
			out := make([]string, 0, len(vars))
			for k, v := range vars {
				out = append(out, k+"="+v)
			}
			sort.Strings(out)
			return out
		}
	}
}

// A Program is a loaded, parsed, and resolved module graph. It is immutable.
type Program struct {
	prog    *loader.Program
	cfg     config
	prelude map[string]value.Value
	nsDecls []check.HostNamespace
}

// Load parses the module at path and every module it imports. It returns a *LoadError when a
// module has a syntax error, cannot be found, takes part in an import cycle, or is a script that
// another module imports. A path ending in .d.ant is a declaration file (section 10, Declaration
// files), which Check checks and Run refuses.
func Load(path string, opts ...Option) (*Program, error) {
	cfg := defaultConfig().with(opts)
	cfg.program = path
	table, extra, nsDecls, err := cfg.prelude()
	if err != nil {
		return nil, &LoadError{Diagnostics: []Diagnostic{{Message: err.Error()}}}
	}
	prog, diags := loader.Load(path, loader.Config{
		FS: cfg.fsys, Sources: cfg.sources, Mode: parser.Program, Extra: extra,
	})
	if len(diags) > 0 {
		return nil, &LoadError{Diagnostics: convert(diags)}
	}
	return &Program{prog: prog, cfg: cfg, prelude: table, nsDecls: nsDecls}, nil
}

func convert(diags []loader.Diagnostic) []Diagnostic {
	out := make([]Diagnostic, len(diags))
	for i, d := range diags {
		out[i] = Diagnostic{
			Pos:      Position{Path: d.Path, Line: d.Pos.Line, Col: d.Pos.Col},
			End:      Position{Path: d.Path, Line: d.End.Line, Col: d.End.Col},
			Message:  d.Msg,
			Code:     string(d.Code),
			Severity: severity(d),
			Help:     d.Help,
		}
		for _, f := range d.Fixes {
			fix := Fix{Title: f.Title, Auto: f.Auto}
			for _, e := range f.Edits {
				fix.Edits = append(fix.Edits, Edit{Start: Position{Path: e.Path, Line: e.Start.Line, Col: e.Start.Col},
					End: Position{Path: e.Path, Line: e.End.Line, Col: e.End.Col}, Text: e.Text})
			}
			out[i].Fixes = append(out[i].Fixes, fix)
		}
		for _, r := range d.Related {
			out[i].Related = append(out[i].Related, Related{Pos: Position{Path: r.Path, Line: r.Pos.Line, Col: r.Pos.Col}, Message: r.Msg})
		}
	}
	return out
}

// Run executes the program: it initializes every module in dependency order, the root last, which
// for a script runs its body. It returns nil when the script ends, a *ResultError when it returns
// an Err, a *PanicError when the main goroutine panics, an *ExitError
// when the program calls Os.exit, and ctx.Err() when the context is cancelled first.
func (p *Program) Run(ctx context.Context, opts ...Option) error {
	if err := p.runnable(); err != nil {
		return err
	}
	return convertRunError(interp.Run(ctx, p.prog, p.cfg.with(opts).run(p.prelude)))
}

// runnable is the *LoadError Run and Start return for a root that is a declaration file (section
// 10, Declaration files), which declares names and holds nothing to run; Check checks one.
func (p *Program) runnable() error {
	if root := p.prog.Root; root != nil && parser.IsTestFile(root.File.Path) {
		return &LoadError{Diagnostics: []Diagnostic{{Pos: Position{Path: root.File.Path},
			Message: "a test file holds tests, which ant test runs, and cannot be run"}}}
	}
	if root := p.prog.Root; root != nil && root.File.Declarations {
		return &LoadError{Diagnostics: []Diagnostic{{Pos: Position{Path: root.File.Path},
			Message: "a declaration file declares names for the prelude and cannot be run"}}}
	}
	return nil
}

// Check type-checks every module of the program, including code nothing calls, and returns the
// errors sorted by position. A program with no diagnostics is well typed.
func (p *Program) Check() []Diagnostic {
	return convert(loader.Settle(check.Program(p.prog, check.Config{Namespaces: p.nsDecls, Removed: p.cfg.removed})))
}

// Check checks the programs rooted at paths together: the module at each path and every module
// it imports, each module once however many paths reach it. A path ending in .d.ant is a
// declaration file, checked on its own. A program that does not load, as Load reports, is not
// checked, and its load errors are returned in its place; the others are still checked. Check
// returns every error sorted by path and position, and whether every program loaded.
func Check(paths []string, opts ...Option) (diags []Diagnostic, loaded bool) {
	cfg := defaultConfig().with(opts)
	_, extra, nsDecls, err := cfg.prelude()
	if err != nil {
		return []Diagnostic{{Message: err.Error()}}, false
	}
	loaded = true
	var roots []string
	seen := map[string]bool{}
	for _, path := range paths {
		if parser.ModeOf(path) != parser.Declarations {
			roots = append(roots, path)
			continue
		}
		// the loader settles which roots are one module; a declaration file, checked on its own,
		// is named once
		if seen[filepath.Clean(path)] {
			continue
		}
		seen[filepath.Clean(path)] = true
		p, err := Load(path, opts...)
		if err != nil {
			diags = append(diags, loadDiagnostics(err)...)
			loaded = false
			continue
		}
		diags = append(diags, p.Check()...)
	}
	if len(roots) > 0 {
		// Recovery keeps checking past a syntax error, so that every error of a file is reported
		// in one run; a load error still makes the program one that cannot be loaded.
		prog, lds := loader.LoadRoots(roots, loader.Config{FS: cfg.fsys, Sources: cfg.sources, Mode: parser.Program, Extra: extra, Recover: true})
		loaded = loaded && len(lds) == 0
		if len(prog.Modules) > 0 {
			lds = append(lds, check.Program(prog, check.Config{Namespaces: nsDecls, Removed: cfg.removed, Prior: len(lds)})...)
		}
		diags = append(diags, convert(loader.Settle(lds))...)
	}
	sort.SliceStable(diags, func(i, j int) bool {
		a, b := diags[i].Pos, diags[j].Pos
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Col < b.Col
	})
	return diags, loaded
}

func loadDiagnostics(err error) []Diagnostic {
	var le *LoadError
	if errors.As(err, &le) && len(le.Diagnostics) > 0 {
		return le.Diagnostics
	}
	return []Diagnostic{{Message: err.Error()}}
}

// severity is "warning" for a warning and "error" for everything else.
func severity(d loader.Diagnostic) string {
	if d.Warning {
		return "warning"
	}
	return "error"
}

// Errors is the diagnostics of diags that are errors, without the warnings.
func Errors(diags []Diagnostic) []Diagnostic {
	var out []Diagnostic
	for _, d := range diags {
		if d.Severity != "warning" {
			out = append(out, d)
		}
	}
	return out
}
