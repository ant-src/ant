package interp

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/value"
)

// Debugger stops a run at its stop points, for ant dap (dap.md): at breakpoints, at the end of a
// step, on a pause, and at a panic. When one goroutine stops, every goroutine stops.
//
// Each goroutine is in Ant code or outside it, in a builtin or the host's code. One in Ant code
// is active until it parks; one outside never touches its stack, so a stop waits only for the
// active ones to park, and a goroutine coming back into Ant code while the run is stopped waits
// there. The stacks are read, under mu, only while the run is stopped.
type Debugger struct {
	// OnStop is called once every goroutine has stopped, from the goroutine that stopped or, for
	// a pause, from one of its own.
	OnStop func(Stop)
	// OnLog is called with what a logpoint writes, or the error of a breakpoint's condition.
	OnLog func(text string)
	// OnGoroutine is called as a goroutine starts and as it ends.
	OnGoroutine func(id int, started bool)
	// StopOnEntry stops before the first statement of the root module.
	StopOnEntry bool

	in   atomic.Pointer[Instance] // set as the run starts
	root *module

	halt   atomic.Bool // a stop is under way: every active goroutine parks at its next chance
	frames atomic.Uint64

	mu       sync.Mutex
	cond     *sync.Cond
	gs       map[int]*Goroutine
	ids      int
	active   int    // goroutines in Ant code, not parked
	stopping bool   // a stop is under way or the run is stopped
	gen      uint64 // counts resumes, which parked goroutines wait for
	entry    bool   // StopOnEntry, until the entry stop
	step     *step
	stop     Stop
	bps      map[bpKey]*bp
	fbps     map[bpKey]*bp // function breakpoints
	uncaught bool
	all      bool
}

// Stop describes why the run stopped.
type Stop struct {
	Goroutine int    // the goroutine that stopped
	Reason    string // breakpoint, step, pause, exception, or entry
	Panic     *AntPanic
}

type bpKey struct {
	path string
	pos  ast.Pos
}

// step is a step in progress: the goroutine stepping, and the frames whose stop points end it,
// all of them for a step in.
type step struct {
	g      *Goroutine
	any    bool
	frames map[uint64]bool
}

// StepKind is a step request.
type StepKind int

const (
	StepIn StepKind = iota
	StepOver
	StepOut
)

// NewDebugger returns a debugger that stops at uncaught panics, as dap.md section 5 defaults.
func NewDebugger() *Debugger {
	d := &Debugger{gs: map[int]*Goroutine{}, uncaught: true}
	d.cond = sync.NewCond(&d.mu)
	return d
}

func (d *Debugger) attach(in *Instance) {
	d.in.Store(in)
	d.root = in.modules[in.prog.Root.Index]
	d.entry = d.StopOnEntry
}

// ---- configuration

// SetBreakpoints replaces the breakpoints of the module at path. It returns, for each, why it is
// not set, or nil.
func (d *Debugger) SetBreakpoints(path string, list []Breakpoint) []error {
	errs := make([]error, len(list))
	bps := map[bpKey]*bp{}
	for i, b := range list {
		p, err := newBP(b)
		if err != nil {
			errs[i] = err
			continue
		}
		bps[bpKey{path, b.Pos}] = p
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	for k, p := range d.bps {
		if k.path != path {
			bps[k] = p
		}
	}
	d.bps = bps
	return errs
}

// FunctionBreakpoint is a breakpoint at the first stop point of a function's body, in the
// module at Path.
type FunctionBreakpoint struct {
	Path string
	Breakpoint
}

// SetFunctionBreakpoints replaces every function breakpoint. It returns, for each, why it is not
// set, or nil.
func (d *Debugger) SetFunctionBreakpoints(list []FunctionBreakpoint) []error {
	errs := make([]error, len(list))
	bps := map[bpKey]*bp{}
	for i, b := range list {
		p, err := newBP(b.Breakpoint)
		if err != nil {
			errs[i] = err
			continue
		}
		bps[bpKey{b.Path, b.Pos}] = p
	}
	d.mu.Lock()
	d.fbps = bps
	d.mu.Unlock()
	return errs
}

// SetPanicFilters sets which panics stop the run (dap.md section 5).
func (d *Debugger) SetPanicFilters(uncaught, all bool) {
	d.mu.Lock()
	d.uncaught, d.all = uncaught, all
	d.mu.Unlock()
}

// Flush writes out what the program has printed, once the run has started.
func (d *Debugger) Flush() {
	if in := d.in.Load(); in != nil {
		in.Flush()
	}
}

// ---- control

// Continue resumes every goroutine.
func (d *Debugger) Continue() {
	d.mu.Lock()
	d.resume(nil)
	d.mu.Unlock()
}

// Step resumes every goroutine until the goroutine id reaches the stop point the step looks
// for (dap.md section 4, Stepping).
func (d *Debugger) Step(id int, kind StepKind) {
	d.mu.Lock()
	defer d.mu.Unlock()
	g := d.gs[id]
	if g == nil || !d.stopping {
		d.resume(nil)
		return
	}
	s := &step{g: g, any: kind == StepIn, frames: map[uint64]bool{}}
	cur := g.frameID()
	i := len(g.stack) - 1
	for i >= 0 && g.stack[i].id != cur {
		i--
	}
	if i < 0 {
		i = len(g.stack) - 1
	}
	if kind == StepOver {
		s.frames[cur] = true
	}
	for _, f := range g.stack[:i] {
		s.frames[f.id] = true
	}
	d.resume(s)
}

// Pause stops every goroutine. The stop names the goroutine id, or the main goroutine when that
// one has ended.
func (d *Debugger) Pause(id int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stopping {
		return
	}
	if _, ok := d.gs[id]; !ok {
		id = 1
	}
	d.stopping = true
	d.halt.Store(true)
	go func() {
		d.mu.Lock()
		for d.active > 0 {
			d.cond.Wait()
		}
		st := Stop{Goroutine: id, Reason: "pause"}
		d.stop = st
		d.mu.Unlock()
		d.Flush()
		d.OnStop(st)
	}()
}

// resume ends a stop, with the step to take or nil. mu is held.
func (d *Debugger) resume(s *step) {
	d.step = s
	d.stopping = false
	d.halt.Store(false)
	d.gen++
	d.cond.Broadcast()
}

// ---- goroutines

// register adds g, outside Ant code; it enters with enterAnt.
func (d *Debugger) register(g *Goroutine) {
	d.mu.Lock()
	d.ids++
	g.id = d.ids
	g.outside = 1
	d.gs[g.id] = g
	d.mu.Unlock()
	if d.OnGoroutine != nil {
		d.OnGoroutine(g.id, true)
	}
}

func (d *Debugger) unregister(g *Goroutine) {
	d.mu.Lock()
	if g.dActive {
		g.dActive = false
		d.active--
	}
	delete(d.gs, g.id)
	if d.step != nil && d.step.g == g {
		d.step = nil // the stepping goroutine ended: the program runs on
	}
	d.cond.Broadcast()
	d.mu.Unlock()
	if d.OnGoroutine != nil {
		d.OnGoroutine(g.id, false)
	}
}

// enterAnt marks g as running Ant code, after the stop under way if there is one.
func (g *Goroutine) enterAnt() {
	d := g.d
	d.mu.Lock()
	g.outside--
	if g.outside == 0 {
		for d.stopping {
			d.cond.Wait()
		}
		g.dActive = true
		d.active++
	}
	d.mu.Unlock()
}

// leaveAnt marks g as running outside Ant code, where a stop does not wait for it.
func (g *Goroutine) leaveAnt() {
	d := g.d
	d.mu.Lock()
	g.outside++
	if g.outside == 1 && g.dActive {
		g.dActive = false
		d.active--
		d.cond.Broadcast()
	}
	d.mu.Unlock()
}

// park stops g until the run resumes. mu is held.
func (d *Debugger) parkLocked(g *Goroutine) {
	if g.dActive {
		g.dActive = false
		d.active--
		d.cond.Broadcast()
	}
	gen := d.gen
	for d.gen == gen {
		d.cond.Wait()
	}
	if g.outside == 0 {
		g.dActive = true
		d.active++
	}
}

func (d *Debugger) park(g *Goroutine) {
	d.mu.Lock()
	if d.stopping {
		d.parkLocked(g)
	}
	d.mu.Unlock()
}

// stopAt stops the run for g. When another stop is under way, g parks, and stops the run for
// its own reason once that one ends: a breakpoint or a panic g reached is not lost, unless the
// run resumed by stepping g, which moves g on from it.
func (d *Debugger) stopAt(g *Goroutine, reason string, p *AntPanic) {
	d.mu.Lock()
	for d.stopping {
		d.parkLocked(g)
		if reason == "step" || reason == "entry" || d.step != nil && d.step.g == g {
			d.mu.Unlock()
			return
		}
	}
	d.stopping = true
	d.halt.Store(true)
	d.step = nil
	if g.dActive {
		g.dActive = false
		d.active--
	}
	for d.active > 0 {
		d.cond.Wait()
	}
	st := Stop{Goroutine: g.id, Reason: reason, Panic: p}
	d.stop = st
	gen := d.gen
	d.mu.Unlock()
	d.Flush() // what the program wrote before the stop shows at the stop
	d.OnStop(st)
	d.mu.Lock()
	for d.gen == gen {
		d.cond.Wait()
	}
	if g.outside == 0 {
		g.dActive = true
		d.active++
	}
	d.mu.Unlock()
}

// ---- hooks

// frameID is the frame g is running in: the frame of its last stop point's environment, which
// for a loop body over a sequence is the loop's frame, or else its innermost frame.
func (g *Goroutine) frameID() uint64 {
	if len(g.stack) == 0 {
		return 0
	}
	if f := g.stack[len(g.stack)-1]; f.env != nil && f.atStop {
		return f.env.frame
	}
	return g.stack[len(g.stack)-1].id
}

// stopPoint is the hook before a statement (dap.md section 4, Stop points).
func (g *Goroutine) stopPoint(env *Env, pos ast.Pos) {
	if len(g.stack) == 0 || g.outside > 0 {
		return
	}
	f := &g.stack[len(g.stack)-1]
	f.env, f.at, f.piped, f.atStop = env, pos, nil, true
	g.maybeStop(env, pos)
}

// stagePoint is the hook before a stage of a pipeline, with the value flowing into it.
func (g *Goroutine) stagePoint(env *Env, pos ast.Pos, piped value.Value) {
	if len(g.stack) == 0 || g.outside > 0 {
		return
	}
	f := &g.stack[len(g.stack)-1]
	f.env, f.at, f.piped, f.atStop = env, pos, piped, true
	g.maybeStop(env, pos)
}

func (g *Goroutine) maybeStop(env *Env, pos ast.Pos) {
	d := g.d
	reason, p := d.reason(g, env, pos)
	if p != nil && (p.Condition != "" || p.hitOp != "" || p.log != nil) {
		stop, log := d.hit(g, p, env)
		if log != "" && d.OnLog != nil {
			d.OnLog(log)
		}
		if !stop {
			reason = ""
		}
	}
	if reason != "" {
		d.stopAt(g, reason, nil)
		return
	}
	if d.halt.Load() {
		d.park(g)
	}
}

// reason says why g stops at the stop point, or "" when it does not, with the breakpoint there,
// whose expressions may still keep it from stopping.
func (d *Debugger) reason(g *Goroutine, env *Env, pos ast.Pos) (string, *bp) {
	d.mu.Lock()
	defer d.mu.Unlock()
	switch {
	case d.entry && g.id == 1 && env.mod == d.root:
		d.entry = false
		return "entry", nil
	case d.step != nil && d.step.g == g && (d.step.any || d.step.frames[env.frame]):
		return "step", nil
	}
	if p := d.bps[bpKey{env.mod.path, pos}]; p != nil {
		return "breakpoint", p
	}
	if p := d.fbps[bpKey{env.mod.path, pos}]; p != nil {
		return "function breakpoint", p
	}
	return "", nil
}

// panicked is the hook as a panic starts, at pos, before anything unwinds.
func (g *Goroutine) panicked(p *AntPanic) {
	d := g.d
	d.mu.Lock()
	stop := d.all || d.uncaught && g.catching == 0
	d.mu.Unlock()
	if !stop {
		return
	}
	if len(g.stack) > 0 && len(p.Trace) > 0 {
		f := &g.stack[len(g.stack)-1]
		f.at, f.piped = p.Trace[0].Pos, nil
	}
	d.stopAt(g, "exception", p)
}

// ---- inspection, while the run is stopped

// GoroutineInfo is a goroutine as a debugger lists it.
type GoroutineInfo struct {
	ID   int
	Name string
}

// Goroutines lists the live goroutines by ID.
func (d *Debugger) Goroutines() []GoroutineInfo {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]GoroutineInfo, 0, len(d.gs))
	for _, g := range d.gs {
		out = append(out, GoroutineInfo{ID: g.id, Name: g.name()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (g *Goroutine) name() string {
	if g.id == 1 {
		return "main"
	}
	return fmt.Sprintf("go at %s:%d #%d", g.startPath, g.startPos.Line, g.id)
}

// StackFrame is one frame of a stopped goroutine, innermost first.
type StackFrame struct {
	Function string
	Path     string
	Pos      ast.Pos
	g        *Goroutine
	i        int // index into the goroutine's stack
}

// Frames lists the frames of goroutine id, innermost first, while the run is stopped.
func (d *Debugger) Frames(id int) ([]StackFrame, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	g := d.gs[id]
	if g == nil || !d.stopping {
		return nil, false
	}
	var out []StackFrame
	var pos ast.Pos
	for i := len(g.stack) - 1; i >= 0; i-- {
		f := g.stack[i]
		if i == len(g.stack)-1 {
			pos = f.at
		}
		out = append(out, StackFrame{Function: f.Function, Path: f.Path, Pos: pos, g: g, i: i})
		pos = f.Pos // the call site, in the caller
	}
	return out, true
}

// Var is a named value a debugger shows.
type Var struct {
	Name  string
	Value value.Value
	set   func(value.Value) // writes the binding, nil when it cannot be set
}

// Settable reports whether SetVar can give v a new value.
func (v Var) Settable() bool { return v.set != nil }

// SetVar gives a variable of a stopped frame's scopes a new value (dap.md section 6, Setting
// variables).
func (d *Debugger) SetVar(v Var, x value.Value) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.stopping {
		return errors.New("the program is not stopped")
	}
	if v.set == nil {
		return fmt.Errorf("%s cannot be set", v.Name)
	}
	v.set(x)
	return nil
}

// Scopes returns the Locals, Closure and Module variables of a frame (dap.md section 3).
func (d *Debugger) Scopes(sf StackFrame) (locals, closure, module []Var) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if sf.g == nil || sf.i >= len(sf.g.stack) {
		return nil, nil, nil
	}
	f := sf.g.stack[sf.i]
	if f.piped != nil && sf.i == len(sf.g.stack)-1 {
		locals = append(locals, Var{Name: "(piped)", Value: f.piped})
	}
	seen := map[string]bool{}
	e := f.env
	for ; e != nil && e.frame == f.id; e = e.parent {
		locals = e.vars(locals, seen)
	}
	if f.env != nil && f.env.frame != f.id {
		// A loop body over a sequence: its environment belongs to the loop's frame.
		for e = f.env; e != nil && e.frame == f.env.frame; e = e.parent {
			locals = e.vars(locals, seen)
		}
	}
	for ; e != nil; e = e.parent {
		closure = e.vars(closure, seen)
	}
	if f.env != nil && f.env.mod != nil {
		module = f.env.mod.vars()
	}
	return locals, closure, module
}

// vars appends the initialized bindings of e that seen does not hold yet, in declaration order.
func (e *Env) vars(out []Var, seen map[string]bool) []Var {
	if e.scope == nil {
		return out
	}
	for _, n := range e.scope.Names {
		if n.Type || seen[n.Name] {
			continue
		}
		v := e.get(ast.Slot{Index: n.Index, Shared: n.Shared})
		if v == value.Uninitialized {
			continue
		}
		seen[n.Name] = true
		slot := ast.Slot{Index: n.Index, Shared: n.Shared}
		out = append(out, Var{Name: n.Name, Value: v, set: func(x value.Value) { e.set(slot, x) }})
	}
	return out
}

// Panic returns the panic the run stopped at, or nil.
func (d *Debugger) Panic() *AntPanic {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.stopping {
		return nil
	}
	return d.stop.Panic
}

// StopPoint is a place a run can stop (dap.md section 4).
type StopPoint struct {
	Pos ast.Pos
	// Body is the function, lambda, go block, defer block, or file the point is in.
	Body ast.Node
}

// StopPoints lists the stop points of a module, in source order.
func StopPoints(f *ast.File) []StopPoint {
	var out []StopPoint
	var walk func(n ast.Node, body ast.Node)
	walk = func(n ast.Node, body ast.Node) {
		ast.Inspect(n, func(x ast.Node) bool {
			if x == n {
				return true
			}
			switch x := x.(type) {
			case *ast.FuncDecl, *ast.FuncLit, *ast.Lambda, *ast.Go, *ast.Defer:
				walk(x, x)
				return false
			case *ast.Block:
				for _, s := range x.Stmts {
					out = append(out, StopPoint{Pos: ast.Start(s), Body: body})
				}
			case *ast.Call:
				if x.Stage {
					out = append(out, StopPoint{Pos: x.Bar, Body: body})
				}
			}
			return true
		})
	}
	if f.Script {
		for _, s := range f.Stmts {
			out = append(out, StopPoint{Pos: ast.Start(s), Body: f})
		}
	} else {
		for _, decl := range f.Decls {
			if vd, ok := decl.(*ast.VarDecl); ok {
				out = append(out, StopPoint{Pos: ast.Start(vd), Body: f})
			}
		}
	}
	walk(f, f)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Pos.Before(out[j].Pos) })
	return out
}

// vars returns the module's top-level variables that are initialized, in declaration order.
func (m *module) vars() []Var {
	var out []Var
	decls := m.m.File.Decls
	if m.m.File.Script {
		decls = nil
		for _, s := range m.m.File.Stmts {
			if vd, ok := s.(*ast.VarDecl); ok {
				decls = append(decls, vd)
			}
		}
	}
	for _, d := range decls {
		vd, ok := d.(*ast.VarDecl)
		if !ok {
			continue
		}
		ast.Inspect(vd.Target, func(n ast.Node) bool {
			t, ok := n.(*ast.NameTarget)
			if !ok || !t.Slot.Module {
				return true
			}
			if v := m.get(t.Slot.Index); v != value.Uninitialized {
				i := t.Slot.Index
				out = append(out, Var{Name: t.Name, Value: v, set: func(x value.Value) { m.rec.Set(i, x) }})
			}
			return true
		})
	}
	return out
}
