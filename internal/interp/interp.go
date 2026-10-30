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
	Program string // what Os.program returns: the root module's path as the host was given it
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
	// Foreground is called as Cmd.run starts a child in the foreground, and the function it
	// returns as the child exits, with the signal that ended the child or nil, so the host can
	// hold back its keyboard signals meanwhile; that function reports whether one it held is
	// now the program's interrupt, which Cmd.run delivers before it returns. Nil holds nothing.
	Foreground func() func(os.Signal) bool
	// Debugger, when set, stops the run at its breakpoints and steps (dap.md).
	Debugger *Debugger
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
	localTypes  sync.Map // *ast.Block → the tags and constructors of the types it declares
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
	// onStop holds what a library runs as the run stops, such as killing its children.
	onStopMu sync.Mutex
	onStop   map[*func(aborted bool)]struct{}
	// released is set once the run has stopped and flushed: the host's writers are the host's
	// again, and what a child left running writes through the run is dropped.
	released bool
	stdin    io.Reader // cfg.Stdin, behind a lock when it is not a file
	dbg      *Debugger
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
	if g.d != nil {
		g.d.register(g)
	}
	in.active.Add(1)
	go func() {
		defer in.active.Done()
		defer in.recoverRoot(g)
		if g.d != nil {
			g.enterAnt()
		}
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
	switch r := cfg.Stdin.(type) {
	case nil:
	case *os.File:
		in.stdin = r
	default:
		in.stdin = &lockedReader{r: r}
	}
	in.ctx, in.cancel = context.WithCancel(ctx)
	// The context's end sets the flag, so a goroutine that never waits stops at its next check
	// too, whether Close, an interrupt, or the caller's context ended the run.
	context.AfterFunc(in.ctx, func() { in.done.Store(true) })
	in.setupModules()
	if cfg.Debugger != nil {
		in.dbg = cfg.Debugger
		in.dbg.attach(in)
	}
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
			in.deliverInterrupt()
		case <-in.ctx.Done():
			return
		}
	}
}

// deliverInterrupt delivers one interrupt: to the program's interrupt context while the
// program holds one that is not done, and otherwise by cancelling the run.
func (in *Instance) deliverInterrupt() {
	in.interruptMu.Lock()
	c := in.interrupt
	in.interruptMu.Unlock()
	if c != nil && !c.Done() {
		c.Cancel(value.ErrInterrupt)
		return
	}
	// As Close does: the flag first, so a goroutine whose wait the cancellation ends finds it
	// at its very next check and runs nothing more.
	in.interrupted.Store(true)
	in.done.Store(true)
	in.cancel()
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
	in.stop(errors.Is(err, ErrInterrupted) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
	return err
}

// stop ends the run: the flag first, so a goroutine whose wait the cancellation ends finds it
// at its very next check, then the context, then what the libraries run as the run stops,
// told whether a cancellation ended it, then the grace period for the goroutines still
// running, and a last flush.
func (in *Instance) stop(aborted bool) {
	in.done.Store(true)
	in.cancel()
	in.onStopMu.Lock()
	fs := in.onStop
	in.onStop = nil
	in.onStopMu.Unlock()
	for f := range fs {
		(*f)(aborted)
	}
	finished := make(chan struct{})
	go func() {
		in.active.Wait()
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(in.cfg.Grace):
	}
	in.outMu.Lock()
	in.out.Flush()
	in.errMu.Lock()
	in.released = true
	in.errMu.Unlock()
	in.outMu.Unlock()
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
	if g.d != nil {
		defer g.d.unregister(g)
	}
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
		io.WriteString(in.ErrWriter(), msg+"\n")
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
	io.WriteString(in.ErrWriter(), sb.String())
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
	if !in.released {
		in.out.Flush()
	}
	in.outMu.Unlock()
}

// Args returns the program's arguments.
func (in *Instance) Args() []string { return in.cfg.Args }

// Program returns the program's path, "" in a session.
func (in *Instance) Program() string { return in.cfg.Program }

// Stdin returns the run's standard input, an empty stream when none was given.
func (in *Instance) Stdin() io.Reader {
	if in.stdin == nil {
		return eofReader{}
	}
	return in.stdin
}

// lockedReader is a host's standard input that is not a file, read by one reader at a time:
// Os.stdin's pump and the copies into children that are given it.
type lockedReader struct {
	mu sync.Mutex
	r  io.Reader
}

func (r *lockedReader) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.r.Read(p)
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
	if w.in.released {
		return len(p), nil
	}
	return w.in.out.Write(p)
}

// ErrWriter returns a writer to the run's standard error that flushes stdout before each write,
// so that the two streams interleave in program order where they reach one terminal; eprint,
// eprintln, Os.stderr and the run's reports write through it.
func (in *Instance) ErrWriter() io.Writer { return errWriter{in} }

type errWriter struct{ in *Instance }

func (w errWriter) Write(p []byte) (int, error) {
	w.in.Flush()
	return stderrWriter{w.in}.Write(p)
}

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

// OnStop registers f to run as the run stops, before the grace period, told whether a
// cancellation ended the run: the host's, or an interrupt's. The returned function unregisters
// it.
func (in *Instance) OnStop(f func(aborted bool)) (remove func()) {
	key := &f
	in.onStopMu.Lock()
	if in.onStop == nil {
		in.onStop = map[*func(aborted bool)]struct{}{}
	}
	in.onStop[key] = struct{}{}
	in.onStopMu.Unlock()
	return func() {
		in.onStopMu.Lock()
		delete(in.onStop, key)
		in.onStopMu.Unlock()
	}
}

// Foreground tells the host that a child is starting in the foreground, and returns what tells
// it the child exited and which signal ended it, nil for none, and delivers the interrupt the
// host held back meanwhile when the host says it is now the program's.
func (in *Instance) Foreground() func(os.Signal) {
	if in.cfg.Foreground == nil {
		return func(os.Signal) {}
	}
	release := in.cfg.Foreground()
	return func(ended os.Signal) {
		if release(ended) {
			in.deliverInterrupt()
		}
	}
}

// ChildStdin returns the run's standard input as a child process is given it: the file itself
// when it is one, so a terminal is handed over as it is; nil, the null device, when the run has
// none; and otherwise the reader every reader of standard input shares.
func (in *Instance) ChildStdin() io.Reader {
	if in.stdin == nil {
		return nil
	}
	return in.stdin
}

// ChildStdout returns the run's standard output as a child process is given it, once the buffer
// is flushed: the file itself when it is one, and otherwise the buffer, so what the child writes
// and what the program prints never write the host's writer at once.
func (in *Instance) ChildStdout() io.Writer {
	if f, ok := in.cfg.Stdout.(*os.File); ok {
		return f
	}
	return in.Stdout()
}

// ChildStderr returns the run's standard error as a child process is given it: the file itself
// when it is one, and otherwise the writer behind the lock the run's reports take.
func (in *Instance) ChildStderr() io.Writer {
	if f, ok := in.cfg.Stderr.(*os.File); ok {
		return f
	}
	return stderrWriter{in}
}

// stderrWriter writes to the run's standard error under the lock its reports take, until the
// run is released.
type stderrWriter struct{ in *Instance }

func (w stderrWriter) Write(p []byte) (int, error) {
	w.in.errMu.Lock()
	defer w.in.errMu.Unlock()
	if w.in.released {
		return len(p), nil
	}
	return w.in.cfg.Stderr.Write(p)
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
	d     *Debugger // the run's debugger, nil without one and for an evaluation (debug.go)
	main  bool
	stack []frame
	pos   ast.Pos // the call site of the builtin in progress, for its panics
	// frames of blocks that registered deferred blocks, innermost last
	defers []*deferFrame
	// catching counts the Panic.catch calls in progress, for a debugger's uncaught panics.
	catching int
	// Under a debugger: the goroutine's number, where it started, how deep it is outside Ant
	// code, and whether the debugger counts it as active (debug.go).
	id        int
	startPath string
	startPos  ast.Pos
	outside   int
	dActive   bool
}

// frame is one call in progress. Under a debugger it has an ID and the environment and
// position of its last stop point, and the value flowing into a pipeline stage stopped at.
type frame struct {
	Frame
	id     uint64
	env    *Env
	at     ast.Pos
	piped  value.Value
	atStop bool
}

const maxDepth = 10000

func (in *Instance) newGoroutine(main bool) *Goroutine {
	return &Goroutine{in: in, main: main, d: in.dbg}
}

// Instance returns the run this goroutine belongs to.
func (g *Goroutine) Instance() *Instance { return g.in }

// Spawn runs body on a new Ant goroutine, as a go block does: a panic in it is reported and
// ends it alone, and the run's grace period waits for it.
func (g *Goroutine) Spawn(body func(*Goroutine)) {
	path := ""
	if len(g.stack) > 0 {
		path = g.stack[len(g.stack)-1].Path
	}
	g.spawn(path, g.pos, body)
}

// spawn starts a goroutine that a go block, or a builtin called at pos, starts.
func (g *Goroutine) spawn(path string, pos ast.Pos, body func(*Goroutine)) {
	ng := g.in.newGoroutine(false)
	if ng.d != nil {
		ng.startPath, ng.startPos = path, pos
		ng.d.register(ng)
	}
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
	p := &AntPanic{Message: fmt.Sprintf(format, args...), Trace: g.trace(pos)}
	if g.d != nil {
		g.panicked(p)
	}
	panic(p)
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
	if d := g.d; d != nil && d.halt.Load() && g.outside == 0 {
		d.park(g)
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
func (g *Goroutine) Repanic(p *AntPanic) {
	if g.d != nil {
		g.panicked(p)
	}
	panic(p)
}

// Catch calls fn and returns its value, or the Ant panic that left it once the deferred blocks
// inside it have run (section 8, Catching a panic). Anything else that unwinds, Os.exit or the
// run's end, goes on.
func (g *Goroutine) Catch(fn value.Value) (v value.Value, p *AntPanic) {
	// The frames and the call site unwind on their own, by the deferred pops and restores.
	g.catching++
	defer func() {
		g.catching--
		r := recover()
		if r == nil {
			return
		}
		ap, ok := r.(*AntPanic)
		if !ok {
			panic(r)
		}
		p = ap
	}()
	return g.Call(fn), nil
}

func (p *AntPanic) Error() string { return "panic: " + p.Message }
