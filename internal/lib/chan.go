package lib

import (
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// Channel of section 9 (Waiting) and section 11: a channel as a sync type with functions. A select
// over channels is the interpreter's, in internal/interp/select.go.

func chanArg(g *interp.Goroutine, who string, v value.Value) *value.Chan {
	return typedArg[*value.Chan](g, who, "a Channel", v)
}

func init() {
	register("Channel", func() value.Value {
		return syncTag("Channel",
			fn("Channel.new", "capacity?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				capacity := int64(0)
				if given(a[0]) {
					capacity = intArg(g, who, a[0])
				}
				if capacity < 0 {
					g.Panic("Channel.new needs a capacity of at least 0, got %d", capacity)
				}
				boundCount(g, who, "capacity", capacity)
				return value.NewChan(int(capacity))
			}),
			fn("Channel.send", "c, v, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				c := chanArg(g, who, a[0])
				ctx := waitContext(g, who, a[2])
				g.Instance().Flush()
				sent, err := c.Send(ctx, a[1])
				g.AfterWait()
				if err != nil {
					g.Panic("Channel.send: %s", err)
				}
				return sent
			}),
			fn("Channel.receive", "c, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				c := chanArg(g, who, a[0])
				ctx := waitContext(g, who, a[1])
				g.Instance().Flush()
				v, ok, _ := c.Receive(ctx)
				g.AfterWait()
				if !ok {
					return value.Null
				}
				return v
			}),
			fn("Channel.close", "c", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				if err := chanArg(g, who, a[0]).Close(); err != nil {
					g.Panic("Channel.close: %s already", err)
				}
				return value.Void
			}),
			fn("Channel.isClosed", "c", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return chanArg(g, who, a[0]).IsClosed()
			}),
			fn("Channel.length", "c", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return int64(chanArg(g, who, a[0]).Len())
			}),
			fn("Channel.values", "c, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				c := chanArg(g, who, a[0])
				ctx := waitContext(g, who, a[1])
				return seq(who, func(g *interp.Goroutine, push func(value.Value) bool) {
					for {
						g.Instance().Flush()
						v, ok, _ := c.Receive(ctx)
						g.AfterWait()
						if !ok || !push(v) {
							return
						}
					}
				})
			}),
		)
	})
}
