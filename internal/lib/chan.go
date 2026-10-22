package lib

import (
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// Chan of section 9 (Waiting) and section 11: a channel as a sync type with functions. A select
// over channels is the interpreter's, in internal/interp/select.go.

func chanArg(g *interp.Goroutine, who string, v value.Value) *value.Chan {
	c, ok := v.(*value.Chan)
	if !ok {
		g.Panic("%s needs a Chan, got %s", who, value.Describe(v))
	}
	return c
}

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
}
