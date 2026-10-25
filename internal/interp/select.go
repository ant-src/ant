package interp

import (
	"reflect"
	"time"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/temporal"
	"github.com/ant-src/ant/value"
)

// A select (section 5) waits on several heads at once: a receive or a send on a channel, a
// context's Done, the channel a WaitGroup closes at zero, or a timer. Each is a Go channel, so
// one reflect.Select chooses among them, atomically, as Go's select does. A channel clause is
// two cases, the operation and the channel's closed signal, since a value.Chan never closes its
// channel of values.

// selectHeadParams names each head's parameters, for a named argument in a head.
var selectHeadParams = map[string][]string{
	"Channel.receive": {"c"},
	"Channel.send":    {"c", "v"},
	"Context.wait":    {"ctx"},
	"WaitGroup.wait":  {"wg"},
	"Time.sleep":      {"d"},
}

// selectHeadName is the head's function: the parser checked it is a two-part path of one of
// the five.
func selectHeadName(head *ast.Call) string {
	f := head.Fn.(*ast.Field)
	return f.X.(*ast.Ident).Name + "." + f.Name
}

// selectOwner says what a case of the reflect.Select stands for: a clause's operation, a
// clause's closed signal, or the run's context.
type selectOwner struct {
	clause int // -1 for the run's context
	closed bool
	ch     *value.Chan // a receive clause's channel
}

// selectExpr evaluates a select (section 12): the heads' operands in order, then one
// reflect.Select over the heads, an else, and the run's context; the chosen head's result is
// bound and its block evaluated, whose value is the select's.
func (g *Goroutine) selectExpr(e *ast.Select, env *Env) (value.Value, ctl) {
	var cases []reflect.SelectCase
	var owners []selectOwner
	var timers []*time.Timer
	add := func(sc reflect.SelectCase, owner selectOwner) {
		cases = append(cases, sc)
		owners = append(owners, owner)
	}
	recv := func(ch any, owner selectOwner) {
		add(reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}, owner)
	}
	for i, c := range e.Clauses {
		if c.Head == nil {
			add(reflect.SelectCase{Dir: reflect.SelectDefault}, selectOwner{clause: i})
			continue
		}
		name := selectHeadName(c.Head)
		args, ctl := g.selectArgs(c.Head, name, env)
		if ctl != ctlNone {
			return args[0], ctl
		}
		pos := c.Head.Position()
		switch name {
		case "Channel.receive":
			// A channel closed already has only its signal, which takes what is buffered
			// first: with the operation too, the clause would be two ready cases of the
			// select's uniform choice and so chosen twice as often as another.
			ch := g.selectChan(pos, args[0])
			if !ch.IsClosed() {
				recv(ch.Raw(), selectOwner{clause: i})
			}
			recv(ch.Closed(), selectOwner{clause: i, closed: true, ch: ch})
		case "Channel.send":
			// A channel closed already has only its signal, so the send is never buffered
			// after the close; one closed during the wait ends the blocked send the same way.
			ch := g.selectChan(pos, args[0])
			if !ch.IsClosed() {
				add(reflect.SelectCase{Dir: reflect.SelectSend, Chan: reflect.ValueOf(ch.Raw()), Send: reflect.ValueOf(args[1])}, selectOwner{clause: i})
			}
			recv(ch.Closed(), selectOwner{clause: i, closed: true})
		case "Context.wait":
			ctx, ok := args[0].(*value.Context)
			if !ok {
				g.panicf(pos, "Context.wait needs a Context, got %s", value.Describe(args[0]))
			}
			recv(ctx.Ctx.Done(), selectOwner{clause: i})
		case "WaitGroup.wait":
			wg, ok := args[0].(*value.WaitGroup)
			if !ok {
				g.panicf(pos, "WaitGroup.wait needs a WaitGroup, got %s", value.Describe(args[0]))
			}
			recv(wg.Zero(), selectOwner{clause: i})
		case "Time.sleep":
			span, ok := args[0].(temporal.Duration)
			if !ok {
				g.panicf(pos, "Time.sleep needs a Time.Duration, got %s", value.Describe(args[0]))
			}
			d, exact := span.Exact()
			if !exact {
				g.panicf(pos, "Time.sleep needs a duration without years, months or weeks, which have no fixed length")
			}
			t := time.NewTimer(max(d, 0))
			timers = append(timers, t)
			recv(t.C, selectOwner{clause: i})
		}
	}
	recv(g.in.ctx.Done(), selectOwner{clause: -1})
	g.in.Flush()
	chosen, got, ok := reflect.Select(cases)
	for _, t := range timers {
		t.Stop()
	}
	g.AfterWait() // the run's end, whichever case was chosen beside it
	owner := owners[chosen]
	if owner.clause < 0 {
		return value.Void, ctlNone
	}
	c := e.Clauses[owner.clause]
	var result value.Value = value.Void
	if c.Head != nil {
		switch selectHeadName(c.Head) {
		case "Channel.receive":
			switch {
			case owner.closed:
				// The closed signal: what is buffered comes first, then null, as Channel.receive.
				select {
				case v := <-owner.ch.Raw():
					result = v
				default:
					result = value.Null
				}
			case ok:
				result = got.Interface().(value.Value)
			default:
				result = value.Null
			}
		case "Channel.send":
			if owner.closed {
				g.panicf(e.Pos, "select: a send clause was chosen on a closed channel")
			}
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
		if i < 0 {
			g.panicf(a.Value.Position(), "%s takes %s in a select clause, not %s", name, describeParams(params), a.Name)
		}
		if i >= len(params) {
			g.panicf(a.Value.Position(), "%s takes %s in a select clause, got %d arguments", name, describeParams(params), len(head.Args))
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

// describeParams names a head's parameters: "c" or "c and v".
func describeParams(params []string) string {
	if len(params) == 1 {
		return params[0]
	}
	return params[0] + " and " + params[1]
}

func (g *Goroutine) selectChan(pos ast.Pos, v value.Value) *value.Chan {
	c, ok := v.(*value.Chan)
	if !ok {
		g.panicf(pos, "a select clause needs a Channel, got %s", value.Describe(v))
	}
	return c
}
