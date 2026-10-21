package lib

import (
	"time"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

var origin = time.Now()

func init() {
	register("Time", func() value.Value {
		return namespace("Time",
			fn("Time.now", "", func(g *interp.Goroutine, a []value.Value) value.Value {
				return time.Now().UnixMilli()
			}),
			fn("Time.sleep", "ms", func(g *interp.Goroutine, a []value.Value) value.Value {
				ms := intArg(g, "Time.sleep", a[0])
				in := g.Instance()
				in.Flush()
				select {
				case <-time.After(time.Duration(ms) * time.Millisecond):
				case <-in.Context().Done():
				}
				return value.Void
			}),
			fn("Time.monotonic", "", func(g *interp.Goroutine, a []value.Value) value.Value {
				return time.Since(origin).Milliseconds()
			}),
		)
	})
	register("Os", func() value.Value {
		return namespace("Os",
			fn("Os.args", "", func(g *interp.Goroutine, a []value.Value) value.Value {
				args := g.Instance().Args()
				out := make([]value.Value, len(args))
				for i, s := range args {
					out[i] = s
				}
				return value.NewArray(out)
			}),
			fn("Os.env", "name", func(g *interp.Goroutine, a []value.Value) value.Value {
				v, ok := g.Instance().LookupEnv(stringArg(g, "Os.env", a[0]))
				if !ok {
					return value.Null
				}
				return v
			}),
			fn("Os.exit", "status", func(g *interp.Goroutine, a []value.Value) value.Value {
				g.Instance().Exit(int(intArg(g, "Os.exit", a[0])))
				return nil
			}),
		)
	})
}
