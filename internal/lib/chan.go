package lib

import (
	"reflect"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// Chan and Select of section 9 (Waiting) and section 11: a channel is a sync type with
// functions, and Select is a namespace whose clauses carry a function each, run on the calling
// goroutine when reflect.Select chooses them.

func chanArg(g *interp.Goroutine, who string, v value.Value) *value.Chan {
	c, ok := v.(*value.Chan)
	if !ok {
		g.Panic("%s needs a Chan, got %s", who, value.Describe(v))
	}
	return c
}

func caseArg(g *interp.Goroutine, who string, v value.Value) *value.SelectCase {
	c, ok := v.(*value.SelectCase)
	if !ok {
		g.Panic("%s needs an array of Select.Case, got an element of kind %s", who, value.Describe(v))
	}
	return c
}

// CaseTag is Select.Case, the type of a clause.
var CaseTag = value.NewTag("Select.Case", value.SyncTag, nil)

func init() {
	register("Chan", func() value.Value {
		return syncTag("Chan",
			fn("Chan.new", "capacity?", func(g *interp.Goroutine, a []value.Value) value.Value {
				capacity := int64(0)
				if given(a[0]) {
					capacity = intArg(g, "Chan.new", a[0])
				}
				if capacity < 0 {
					g.Panic("Chan.new needs a capacity of at least 0, got %d", capacity)
				}
				return value.NewChan(int(capacity))
			}),
			fn("Chan.send", "c, v, ctx?", func(g *interp.Goroutine, a []value.Value) value.Value {
				c := chanArg(g, "Chan.send", a[0])
				ctx := waitContext(g, optionalContext(g, "Chan.send", a[2]))
				g.Instance().Flush()
				sent, err := c.Send(ctx, a[1])
				if err != nil {
					g.Panic("Chan.send: %s", err)
				}
				return sent
			}),
			fn("Chan.receive", "c, ctx?", func(g *interp.Goroutine, a []value.Value) value.Value {
				c := chanArg(g, "Chan.receive", a[0])
				ctx := waitContext(g, optionalContext(g, "Chan.receive", a[1]))
				g.Instance().Flush()
				v, ok, _ := c.Receive(ctx)
				if !ok {
					return value.Null
				}
				return v
			}),
			fn("Chan.close", "c", func(g *interp.Goroutine, a []value.Value) value.Value {
				if err := chanArg(g, "Chan.close", a[0]).Close(); err != nil {
					g.Panic("Chan.close: %s already", err)
				}
				return value.Void
			}),
			fn("Chan.isClosed", "c", func(g *interp.Goroutine, a []value.Value) value.Value {
				return chanArg(g, "Chan.isClosed", a[0]).IsClosed()
			}),
			fn("Chan.length", "c", func(g *interp.Goroutine, a []value.Value) value.Value {
				return int64(chanArg(g, "Chan.length", a[0]).Len())
			}),
			fn("Chan.values", "c, ctx?", func(g *interp.Goroutine, a []value.Value) value.Value {
				c := chanArg(g, "Chan.values", a[0])
				ctx := waitContext(g, optionalContext(g, "Chan.values", a[1]))
				return seq("Chan.values", func(g *interp.Goroutine, push func(value.Value) bool) {
					for {
						g.Instance().Flush()
						v, ok, _ := c.Receive(ctx)
						if !ok || !push(v) {
							return
						}
					}
				})
			}),
		)
	})
	register("Select", func() value.Value {
		return namespace("Select",
			constant("Case", CaseTag),
			fn("Select.receive", "c, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				return &value.SelectCase{Chan: chanArg(g, "Select.receive", a[0]), Fn: funcArg(g, "Select.receive", a[1])}
			}),
			fn("Select.send", "c, v, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				return &value.SelectCase{Chan: chanArg(g, "Select.send", a[0]), Send: true, Value: a[1], Fn: funcArg(g, "Select.send", a[2])}
			}),
			fn("Select.otherwise", "f", func(g *interp.Goroutine, a []value.Value) value.Value {
				return &value.SelectCase{Fn: funcArg(g, "Select.otherwise", a[0])}
			}),
			fn("Select.wait", "cases, ctx?", selectWait),
		)
	})
}

// selectWait is Select.wait: one reflect.Select over the clauses and the context, then the
// chosen clause's function on the calling goroutine.
func selectWait(g *interp.Goroutine, a []value.Value) value.Value {
	const who = "Select.wait"
	clauses := arrayArg(g, who, a[0]).Values()
	ctx := waitContext(g, optionalContext(g, who, a[1]))
	if len(clauses) == 0 {
		g.Panic("%s needs at least one clause", who)
	}
	cases := make([]reflect.SelectCase, 0, len(clauses)+1)
	sel := make([]*value.SelectCase, 0, len(clauses))
	otherwise := false
	for _, v := range clauses {
		sc := caseArg(g, who, v)
		switch {
		case sc.Chan == nil:
			if otherwise {
				g.Panic("%s takes at most one otherwise clause", who)
			}
			otherwise = true
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectDefault})
		case sc.Send:
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectSend, Chan: reflect.ValueOf(sc.Chan.Raw()), Send: reflect.ValueOf(sc.Value)})
		default:
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(sc.Chan.Raw())})
		}
		sel = append(sel, sc)
	}
	cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())})
	g.Instance().Flush()
	chosen, recv, ok, closed := doSelect(cases)
	if closed {
		g.Panic("Select.send: the channel is closed")
	}
	if chosen == len(sel) {
		return value.Null // the context ended the wait
	}
	sc := sel[chosen]
	if sc.Chan == nil || sc.Send {
		return g.Call(sc.Fn)
	}
	if !ok {
		return g.Call(sc.Fn, value.Null)
	}
	return g.Call(sc.Fn, recv.Interface().(value.Value))
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
