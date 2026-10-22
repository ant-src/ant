package interp

import (
	"math"
	"reflect"
	"time"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/value"
)

// A select (section 5) waits on several heads at once: a receive or a send on a channel, a
// context's Done, the channel a WaitGroup closes at zero, or a timer. Each is a Go channel, so
// one reflect.Select chooses among them, atomically, as Go's select does.

// selectHeadParams names each head's parameters, for a named argument in a head.
var selectHeadParams = map[string][]string{
	"Chan.receive":   {"c"},
	"Chan.send":      {"c", "v"},
	"Context.wait":   {"ctx"},
	"WaitGroup.wait": {"wg"},
	"Time.sleep":     {"ms"},
}

// selectHeadName is the head's function: the parser checked it is a two-part path of one of
// the five.
func selectHeadName(head *ast.Call) string {
	f := head.Fn.(*ast.Field)
	return f.X.(*ast.Ident).Name + "." + f.Name
}

// selectExpr evaluates a select (section 12): the heads' operands in order, then one
// reflect.Select over the heads, an else, and the run's context; the chosen head's result is
// bound and its block evaluated, whose value is the select's.
func (g *Goroutine) selectExpr(e *ast.Select, env *Env) (value.Value, ctl) {
	cases := make([]reflect.SelectCase, 0, len(e.Clauses)+1)
	var timers []*time.Timer
	for _, c := range e.Clauses {
		if c.Head == nil {
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectDefault})
			continue
		}
		name := selectHeadName(c.Head)
		args, ctl := g.selectArgs(c.Head, name, env)
		if ctl != ctlNone {
			return args[0], ctl
		}
		pos := c.Head.Position()
		switch name {
		case "Chan.receive":
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(g.selectChan(pos, args[0]).Raw())})
		case "Chan.send":
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectSend, Chan: reflect.ValueOf(g.selectChan(pos, args[0]).Raw()), Send: reflect.ValueOf(args[1])})
		case "Context.wait":
			ctx, ok := args[0].(*value.Context)
			if !ok {
				g.panicf(pos, "Context.wait needs a Context, got %s", value.Describe(args[0]))
			}
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Ctx.Done())})
		case "WaitGroup.wait":
			wg, ok := args[0].(*value.WaitGroup)
			if !ok {
				g.panicf(pos, "WaitGroup.wait needs a WaitGroup, got %s", value.Describe(args[0]))
			}
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(wg.Zero())})
		case "Time.sleep":
			ms, ok := args[0].(int64)
			if !ok {
				g.panicf(pos, "Time.sleep needs an int, got %s", value.Describe(args[0]))
			}
			if ms > math.MaxInt64/int64(time.Millisecond) {
				ms = math.MaxInt64 / int64(time.Millisecond)
			}
			t := time.NewTimer(time.Duration(ms) * time.Millisecond)
			timers = append(timers, t)
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(t.C)})
		}
	}
	cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(g.in.ctx.Done())})
	g.in.Flush()
	chosen, recv, ok, closed := doSelect(cases)
	for _, t := range timers {
		t.Stop()
	}
	if closed {
		g.panicf(e.Pos, "select: a send clause was chosen on a closed channel")
	}
	if chosen == len(e.Clauses) {
		g.check() // the run is cancelled: the goroutine ends here or at its next check
		return value.Void, ctlNone
	}
	c := e.Clauses[chosen]
	var result value.Value = value.Void
	if c.Head != nil {
		switch selectHeadName(c.Head) {
		case "Chan.receive":
			if ok {
				result = recv.Interface().(value.Value)
			} else {
				result = value.Null
			}
		case "Chan.send":
			result = true
		}
	}
	cenv := env
	if c.Scope != nil {
		cenv = newEnv(c.Scope, env)
	}
	if c.Target != nil {
		g.bind(c.Target, result, cenv)
	}
	return g.evalBlock(c.Body, cenv)
}

// selectArgs evaluates a head's arguments in order, by position or by name, into the head's
// parameter order; the checker has already ruled out an argument the head does not take.
func (g *Goroutine) selectArgs(head *ast.Call, name string, env *Env) ([]value.Value, ctl) {
	params := selectHeadParams[name]
	args := make([]value.Value, len(params))
	next := 0
	for _, a := range head.Args {
		i := -1
		if a.Name == "" {
			i = next
			next++
		} else {
			for j, p := range params {
				if p == a.Name {
					i = j
				}
			}
		}
		v, c := g.eval(a.Value, env)
		if c != ctlNone {
			return []value.Value{v}, c
		}
		if i < 0 || i >= len(params) {
			g.panicf(a.Value.Position(), "%s takes %d arguments in a select clause", name, len(params))
		}
		args[i] = v
	}
	for i, v := range args {
		if v == nil {
			g.panicf(head.Position(), "%s needs its %s in a select clause", name, params[i])
		}
	}
	return args, ctlNone
}

func (g *Goroutine) selectChan(pos ast.Pos, v value.Value) *value.Chan {
	c, ok := v.(*value.Chan)
	if !ok {
		g.panicf(pos, "a select clause needs a Chan, got %s", value.Describe(v))
	}
	return c
}

// doSelect runs reflect.Select and reports a send chosen on a closed channel instead of the
// runtime's panic.
func doSelect(cases []reflect.SelectCase) (chosen int, recv reflect.Value, ok bool, closed bool) {
	defer func() {
		if r := recover(); r != nil {
			if e, isErr := r.(error); isErr && e.Error() == "send on closed channel" {
				closed = true
				return
			}
			panic(r)
		}
	}()
	chosen, recv, ok = reflect.Select(cases)
	return
}
