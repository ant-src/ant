package interp

import (
	"context"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/value"
)

// Start creates an instance and initializes its modules without calling main. The caller
// runs functions with Call and ends the instance with Close.
func Start(ctx context.Context, prog *loader.Program, cfg Config) (*Instance, error) {
	in := New(ctx, prog, cfg)
	g := in.newGoroutine(true)
	if err := in.runOn(func() { in.initModules(g) }); err != nil {
		in.Close()
		return nil, err
	}
	return in, nil
}

// runOn runs f on the calling Go goroutine, converting what ends it into an error: an Ant
// panic is reported and returned as a *PanicError.
func (in *Instance) runOn(f func()) (err error) {
	in.active.Add(1)
	defer in.active.Done()
	defer func() {
		if r := recover(); r != nil {
			err = in.endErr(r)
		}
	}()
	f()
	return nil
}

// Call calls an Ant function value with positional arguments, on the calling Go goroutine.
func (in *Instance) Call(fn value.Value, args ...value.Value) (v value.Value, err error) {
	g := in.newGoroutine(false)
	// The call has no site, so the function's body stands in for it, for a panic of the call
	// itself such as a parameter left without an argument.
	var at ast.Pos
	if f, ok := fn.(*Func); ok && f.Body != nil {
		at = f.Body.Pos
	}
	err = in.runOn(func() { v = g.callValue(fn, args, nil, at) })
	return v, err
}

// Lookup returns a top-level name of the root module.
func (in *Instance) Lookup(name string) (value.Value, bool) {
	v, ok := in.modules[in.prog.Root.Index].rec.Field(name)
	if !ok || v == value.Uninitialized {
		return nil, false
	}
	return v, true
}

// Close stops the instance: its goroutines find the flag at their next check, blocking calls
// see the cancelled context, and reports in progress get the grace period to finish.
func (in *Instance) Close() {
	in.finish(context.Canceled)
	in.stop()
}

// Context returns the run's context, for blocking host functions.
func (g *Goroutine) Context() context.Context { return g.in.ctx }
