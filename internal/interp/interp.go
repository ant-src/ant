// Package interp evaluates a loaded program (section 12). An Instance is one run: it owns the
// module records, the writers, and the cancellation flag. Each goroutine carries its own call
// stack. Values come from the value package; functions and environments are defined here.
package interp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/value"
)

// Config configures a run.
type Config struct {
	Prelude map[string]value.Value
	Stdin   io.Reader // what Os.stdin reads; nil is an empty stream
	Stdout  io.Writer
	Stderr  io.Writer
	Args    []string
	Env     func(name string) (string, bool)
	Environ func() []string // every variable as name=value; nil is the process's environment
	// Grace is how long Run waits, after main returns, for the other goroutines to finish.
	// They are told to stop at once; the wait only lets reports in progress complete. Zero
	// means the default of 100 milliseconds.
	Grace time.Duration
}

// Frame is one entry of a panic trace, innermost first.
type Frame struct {
	Function string
	Path     string
	Pos      ast.Pos
}

// PanicError is a panic that ended the main goroutine.
type PanicError struct {
	Message string
	Trace   []Frame
}

func (e *PanicError) Error() string { return "panic: " + e.Message }

// ExitError is Os.exit.
type ExitError struct {
	Status int
}

func (e *ExitError) Error() string { return fmt.Sprintf("exit status %d", e.Status) }

// Instance is one run of a program.
type Instance struct {
	prog    *loader.Program
	cfg     Config
	modules []*module
	out     *bufio.Writer
	outMu   sync.Mutex
	errMu   sync.Mutex
	keep    map[string]any
	keepMu  sync.Mutex
	done    atomic.Bool
	result  chan error
	once    sync.Once
	ctx     context.Context
	cancel  context.CancelFunc
	active  sync.WaitGroup
}

// AntPanic is the Go panic value that carries an Ant panic while it unwinds.
type AntPanic struct {
	Message string
	Trace   []Frame
}

type exitSignal struct {
	status int
}

type cancelSignal struct{}

// Run executes prog: initializes the modules in order and calls main.
func Run(ctx context.Context, prog *loader.Program, cfg Config) error {
	in := New(ctx, prog, cfg)
	g := in.newGoroutine(true)
	in.active.Add(1)
	go func() {
		defer in.active.Done()
		defer in.recoverRoot(g)
		in.initModules(g)
		main, ok := in.modules[prog.Root.Index].rec.Field("main")
		if !ok {
			g.panicf(ast.Pos{}, "no func main")
		}
		g.callValue(main, nil, nil, ast.Pos{})
	}()
	return in.wait()
}

// New creates an instance without running anything. Sessions use it.
func New(ctx context.Context, prog *loader.Program, cfg Config) *Instance {
	if cfg.Stdout == nil {
		cfg.Stdout = io.Discard
	}
	if cfg.Stderr == nil {
		cfg.Stderr = io.Discard
	}
	if cfg.Grace == 0 {
		cfg.Grace = 100 * time.Millisecond
	}
	in := &Instance{prog: prog, cfg: cfg, out: bufio.NewWriter(cfg.Stdout), result: make(chan error, 1)}
	in.ctx, in.cancel = context.WithCancel(ctx)
	in.setupModules()
	return in
}

// wait blocks until main returns, a goroutine calls Os.exit, or the context is cancelled, then
// stops the run: every goroutine finds the flag at its next check, blocking calls see the
// cancelled context, and a report in progress gets the grace period to finish.
func (in *Instance) wait() error {
	var err error
	select {
	case err = <-in.result:
	case <-in.ctx.Done():
		err = in.ctx.Err()
	}
	in.done.Store(true)
	in.cancel()
	finished := make(chan struct{})
	go func() {
		in.active.Wait()
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(in.cfg.Grace):
	}
	in.Flush()
	return err
}

func (in *Instance) finish(err error) {
	in.once.Do(func() { in.result <- err })
}

// recoverRoot handles whatever ends a goroutine: a panic, Os.exit, cancellation, or a normal
// return of the main goroutine.
func (in *Instance) recoverRoot(g *Goroutine) {
	r := recover()
	switch x := r.(type) {
	case nil:
		if g.main {
			in.finish(nil)
		}
	case *AntPanic:
		in.report(x)
		if g.main {
			in.finish(&PanicError{Message: x.Message, Trace: x.Trace})
		}
	case *exitSignal:
		in.finish(&ExitError{Status: x.status})
	case *cancelSignal:
		if g.main {
			in.finish(in.ctx.Err())
		}
	case *loopUnwind:
		// cannot escape a function; treat as an internal error
		in.report(&AntPanic{Message: "internal error: loop unwind escaped its loop"})
		if g.main {
			in.finish(&PanicError{Message: "internal error"})
		}
	default:
		panic(r)
	}
}

func (in *Instance) report(p *AntPanic) {
	in.Flush()
	var sb strings.Builder
	sb.WriteString("panic: ")
	sb.WriteString(p.Message)
	sb.WriteByte('\n')
	for _, f := range p.Trace {
		fmt.Fprintf(&sb, "  at %s (%s:%d:%d)\n", f.Function, f.Path, f.Pos.Line, f.Pos.Col)
	}
	in.errMu.Lock()
	io.WriteString(in.cfg.Stderr, sb.String())
	in.errMu.Unlock()
}

// Print writes a line to the run's stdout.
func (in *Instance) Print(s string) {
	in.outMu.Lock()
	in.out.WriteString(s)
	in.out.WriteByte('\n')
	in.outMu.Unlock()
}

// Flush flushes stdout; blocking builtins call it so output is not held back.
func (in *Instance) Flush() {
	in.outMu.Lock()
	in.out.Flush()
	in.outMu.Unlock()
}

// Args returns the program's arguments.
func (in *Instance) Args() []string { return in.cfg.Args }

// Stdin returns the run's standard input, an empty stream when none was given.
func (in *Instance) Stdin() io.Reader {
	if in.cfg.Stdin == nil {
		return eofReader{}
	}
	return in.cfg.Stdin
}

type eofReader struct{}

func (eofReader) Read([]byte) (int, error) { return 0, io.EOF }

// Stdout returns a writer into the buffer print writes to, locked per write, so that what a
// program writes through it and what it prints come out in program order.
func (in *Instance) Stdout() io.Writer { return stdoutWriter{in} }

type stdoutWriter struct{ in *Instance }

func (w stdoutWriter) Write(p []byte) (int, error) {
	w.in.outMu.Lock()
	defer w.in.outMu.Unlock()
	return w.in.out.Write(p)
}

// Stderr returns the run's standard error.
func (in *Instance) Stderr() io.Writer { return in.cfg.Stderr }

// Keep returns the value stored under key for this run, creating it with create on the first
// call: a library's per-run state, such as the handles of the standard streams.
func (in *Instance) Keep(key string, create func() any) any {
	in.keepMu.Lock()
	defer in.keepMu.Unlock()
	if v, ok := in.keep[key]; ok {
		return v
	}
	if in.keep == nil {
		in.keep = map[string]any{}
	}
	v := create()
	in.keep[key] = v
	return v
}

// LookupEnv consults the run's environment.
func (in *Instance) LookupEnv(name string) (string, bool) {
	if in.cfg.Env == nil {
		return "", false
	}
	return in.cfg.Env(name)
}

// Environ returns the run's whole environment as name=value pairs.
func (in *Instance) Environ() []string {
	if in.cfg.Environ == nil {
		return os.Environ()
	}
	return in.cfg.Environ()
}

// Context returns the run's context, for blocking builtins.
func (in *Instance) Context() context.Context { return in.ctx }

// Exit ends the run with the given status, from any goroutine. Deferred blocks do not run on
// the way out; the run's other goroutines stop at their next check.
func (in *Instance) Exit(status int) {
	panic(&exitSignal{status: status})
}

// Cancelled reports whether the run is over or its context is done.
func (in *Instance) Cancelled() bool { return in.done.Load() }

// ---- goroutines

// Goroutine is the per-goroutine interpreter state: its call stack and depth.
type Goroutine struct {
	in    *Instance
	main  bool
	stack []Frame
	depth int
	pos   ast.Pos // the call site of the builtin in progress, for its panics
	// frames of blocks that registered deferred blocks, innermost last
	defers []*deferFrame
}

const maxDepth = 10000

func (in *Instance) newGoroutine(main bool) *Goroutine {
	return &Goroutine{in: in, main: main}
}

// Instance returns the run this goroutine belongs to.
func (g *Goroutine) Instance() *Instance { return g.in }

// Spawn runs body on a new goroutine.
func (g *Goroutine) spawn(body func(*Goroutine)) {
	ng := g.in.newGoroutine(false)
	g.in.active.Add(1)
	go func() {
		defer g.in.active.Done()
		defer g.in.recoverRoot(ng)
		body(ng)
	}()
}

// Panic raises an Ant panic at the current position.
func (g *Goroutine) Panic(format string, args ...any) {
	g.panicf(g.pos, format, args...)
}

func (g *Goroutine) panicf(pos ast.Pos, format string, args ...any) {
	panic(&AntPanic{Message: fmt.Sprintf(format, args...), Trace: g.trace(pos)})
}

// trace builds the stack trace, innermost first, with pos as the innermost position.
func (g *Goroutine) trace(pos ast.Pos) []Frame {
	var out []Frame
	for i := len(g.stack) - 1; i >= 0; i-- {
		f := g.stack[i]
		out = append(out, Frame{Function: f.Function, Path: f.Path, Pos: pos})
		pos = f.Pos // the call site, in the caller
	}
	return out
}

// check is the cancellation check at calls and loop back-edges.
func (g *Goroutine) check() {
	if g.in.done.Load() {
		panic(&cancelSignal{})
	}
}

// Catch runs f and returns the Ant panic that ended it, if one did. Other panics, such as
// Os.exit and cancellation, propagate.
func (g *Goroutine) Catch(f func()) (p *AntPanic) {
	defer func() {
		if r := recover(); r != nil {
			if ap, ok := r.(*AntPanic); ok {
				p = ap
				return
			}
			panic(r)
		}
	}()
	f()
	return nil
}

// Repanic raises a caught Ant panic again.
func (g *Goroutine) Repanic(p *AntPanic) { panic(p) }

func (p *AntPanic) Error() string { return "panic: " + p.Message }
