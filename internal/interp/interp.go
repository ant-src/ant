// Package interp evaluates a loaded program (section 12). An Instance is one run: it owns the
// module records, the writers, and the cancellation flag. Each goroutine carries its own call
// stack. Values come from the value package; functions and environments are defined here.
package interp

import (
	"bufio"
	"context"
	"errors"
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
	// Grace is how long Run waits, after the script ends, for the other goroutines to finish.
	// They are told to stop at once; the wait only lets reports in progress complete. Zero
	// means the default of 100 milliseconds.
	Grace time.Duration
	// Interrupts delivers the host's interrupts, one value each. The first goes to the context
	// Os.interrupt created, when the program created one and it is not done; otherwise the run
	// is cancelled, as the first is for a program that never asked. Nil delivers none.
	Interrupts <-chan struct{}
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

// ResultError is an Err the script returned (section 10, Scripts). Message is String.from of
// its error field, which the run has written to standard error.
type ResultError struct {
	Message string
}

func (e *ResultError) Error() string { return e.Message }

// ExitError is Os.exit.
type ExitError struct {
	Status int
}

func (e *ExitError) Error() string { return fmt.Sprintf("exit status %d", e.Status) }

// Instance is one run of a program.
type Instance struct {
	prog        *loader.Program
	cfg         Config
	modules     []*module
	out         *bufio.Writer
	outMu       sync.Mutex
	errMu       sync.Mutex
	keep        map[string]any
	keepMu      sync.Mutex
	localTypes  sync.Map // *ast.TypeDecl of a block → the tag or constructor it binds
	done        atomic.Bool
	interrupted atomic.Bool // the relay ended the run: Run returns ErrInterrupted
	result      chan error
	once        sync.Once
	ctx         context.Context
	cancel      context.CancelFunc
	active      sync.WaitGroup
	// interrupt is the context Os.interrupt returns, made at the first call.
	interruptMu sync.Mutex
	interrupt   *value.Context
}

// AntPanic is the Go panic value that carries an Ant panic while it unwinds.
type AntPanic struct {
	Message string
	Trace   []Frame
}

type exitSignal struct {
	status int
}

// resultSignal carries an Err the script returned out of the main goroutine, after its
// deferred blocks have run.
type resultSignal struct {
	err value.Value
}

type cancelSignal struct{}

// ErrInterrupted is the error Run returns when an interrupt from Config.Interrupts ended the
// run: one the program had not asked for, or the second.
var ErrInterrupted = errors.New("interrupted")

// Run executes prog: initializes the modules in order, the script last, which runs its body.
func Run(ctx context.Context, prog *loader.Program, cfg Config) error {
	in := New(ctx, prog, cfg)
	g := in.newGoroutine(true)
	in.active.Add(1)
	go func() {
		defer in.active.Done()
		defer in.recoverRoot(g)
		in.initModules(g)
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
	// The context's end sets the flag, so a goroutine that never waits stops at its next check
	// too, whether Close, an interrupt, or the caller's context ended the run.
	context.AfterFunc(in.ctx, func() { in.done.Store(true) })
	in.setupModules()
	if cfg.Interrupts != nil {
		go in.relayInterrupts(cfg.Interrupts)
	}
	return in
}

// relayInterrupts delivers each interrupt the host sends: to the program's interrupt context
// while the program holds one that is not done, and otherwise by cancelling the run.
func (in *Instance) relayInterrupts(interrupts <-chan struct{}) {
	for {
		select {
		case _, ok := <-interrupts:
			if !ok {
				return
			}
			in.interruptMu.Lock()
			c := in.interrupt
			in.interruptMu.Unlock()
			if c != nil && !c.Done() {
				c.Cancel(value.ErrInterrupt)
			} else {
				// As Close does: the flag first, so a goroutine whose wait the cancellation
				// ends finds it at its very next check and runs nothing more.
				in.interrupted.Store(true)
				in.done.Store(true)
				in.cancel()
			}
		case <-in.ctx.Done():
			return
		}
	}
}

// Interrupt returns the context Os.interrupt gives: the same one at every call, done when the
// host delivers an interrupt.
func (in *Instance) Interrupt() *value.Context {
	in.interruptMu.Lock()
	defer in.interruptMu.Unlock()
	if in.interrupt == nil {
		in.interrupt = value.NewContext(in.ctx, 0)
	}
	return in.interrupt
}

// wait blocks until the script ends, a goroutine calls Os.exit, or the context is cancelled, then
// stops the run: every goroutine finds the flag at its next check, blocking calls see the
// cancelled context, and a report in progress gets the grace period to finish.
func (in *Instance) wait() error {
	var err error
	select {
	case err = <-in.result:
	case <-in.ctx.Done():
		err = in.stopErr()
	}
	in.stop()
	return err
}

// stop ends the run: the flag first, so a goroutine whose wait the cancellation ends finds it
// at its very next check, then the context, then the grace period for the goroutines still
// running, and a last flush.
func (in *Instance) stop() {
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
}

// stopErr is the error of a run its context ended: ErrInterrupted when an interrupt did, which
// the relay records before it sets the flag, and otherwise the context's error. A goroutine may
// find the flag before the context is cancelled, so the error is never nil.
func (in *Instance) stopErr() error {
	if in.interrupted.Load() {
		return ErrInterrupted
	}
	if err := in.ctx.Err(); err != nil {
		return err
	}
	return context.Canceled
}

func (in *Instance) finish(err error) {
	in.once.Do(func() { in.result <- err })
}

// recoverRoot handles whatever ends a goroutine: a panic, Os.exit, cancellation, or a normal
// return of the main goroutine.
func (in *Instance) recoverRoot(g *Goroutine) {
	r := recover()
	if r == nil {
		if g.main {
			in.finish(nil)
		}
		return
	}
	err := in.endErr(r)
	if _, exit := err.(*ExitError); exit || g.main {
		in.finish(err)
	}
}

// endErr converts what a goroutine panicked with into the error that ends it: an Ant panic,
// which it reports, Os.exit, or the run's end. Any other panic is a Go one and goes on.
func (in *Instance) endErr(r any) error {
	switch x := r.(type) {
	case *AntPanic:
		in.report(x)
		return &PanicError{Message: x.Message, Trace: x.Trace}
	case *exitSignal:
		return &ExitError{Status: x.status}
	case *resultSignal:
		// An Error is written as its message alone, any other error as String.from renders it
		// (section 10, Scripts).
		msg := value.Render(x.err)
		if r, ok := x.err.(*value.Record); ok && r.Tag == value.ErrorTag {
			if m, ok := r.Get(0).(string); ok {
				msg = m
			}
		}
		in.Flush()
		in.errMu.Lock()
		io.WriteString(in.cfg.Stderr, msg+"\n")
		in.errMu.Unlock()
		return &ResultError{Message: msg}
	case *cancelSignal:
		return in.stopErr()
	case *loopUnwind:
		// forSeq catches its own on the goroutine that raised it, so this would be a bug.
		in.report(&AntPanic{Message: "internal error: loop unwind escaped its loop"})
		return &PanicError{Message: "internal error"}
	}
	panic(r)
}

// traceEnds is how many frames a reported trace keeps at each end.
const traceEnds = 10

func (in *Instance) report(p *AntPanic) {
	in.Flush()
	var sb strings.Builder
	sb.WriteString("panic: ")
	sb.WriteString(p.Message)
	sb.WriteByte('\n')
	for i, f := range p.Trace {
		// A long trace, such as a call-depth overflow's, keeps its innermost and outermost
		// frames and counts the ones between.
		if n := len(p.Trace); n > 2*traceEnds+1 && i >= traceEnds && i < n-traceEnds {
			if i == traceEnds {
				fmt.Fprintf(&sb, "  ... %d more frames\n", n-2*traceEnds)
			}
			continue
		}
		fmt.Fprintf(&sb, "  at %s (%s:%d:%d)\n", f.Function, f.Path, f.Pos.Line, f.Pos.Col)
	}
	in.errMu.Lock()
	io.WriteString(in.cfg.Stderr, sb.String())
	in.errMu.Unlock()
}

// Print writes s to the run's stdout.
func (in *Instance) Print(s string) {
	in.outMu.Lock()
	in.out.WriteString(s)
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

// Stdout returns a writer into the buffer print and println write to, locked per write, so that
// what a program writes through it and what it prints come out in program order.
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

// ---- goroutines

// Goroutine is the per-goroutine interpreter state: its call stack, whose length is the depth.
type Goroutine struct {
	in    *Instance
	main  bool
	stack []Frame
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

// Spawn runs body on a new Ant goroutine, as a go block does: a panic in it is reported and
// ends it alone, and the run's grace period waits for it.
func (g *Goroutine) Spawn(body func(*Goroutine)) {
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

// AfterWait ends the goroutine when the run's own context is done: a wait the run's end
// released runs nothing more, whichever path ended the run and whether or not the flag is set
// yet. Every blocking call calls it as it returns.
func (g *Goroutine) AfterWait() {
	if g.in.ctx.Err() != nil {
		panic(&cancelSignal{})
	}
}

// Repanic raises a caught Ant panic again.
func (g *Goroutine) Repanic(p *AntPanic) { panic(p) }

func (p *AntPanic) Error() string { return "panic: " + p.Message }
