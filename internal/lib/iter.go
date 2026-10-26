package lib

import (
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// A sequence is a function that pushes elements to a yield function until it returns false
// (section 5). The Iter functions build sequences as builtins over other sequences.

// seq builds a sequence function from a Go loop body. push sends one element and reports
// whether the consumer wants more.
func seq(name string, run func(g *interp.Goroutine, push func(v value.Value) bool)) *interp.Builtin {
	return fn(name, "yield", func(g *interp.Goroutine, _ string, a []value.Value) value.Value {
		y := funcArg(g, name, a[0])
		run(g, func(v value.Value) bool {
			return boolResult(g, "a sequence's yield", g.Call(y, v))
		})
		return value.Void
	})
}

// drive runs a sequence with a yield that calls push and stops the source once push says so,
// by for's rules (section 12, for).
func drive(g *interp.Goroutine, who string, s value.Value, push func(v value.Value) bool) {
	g.RunSeq(funcArg(g, who, s), push)
}

func init() {
	register("Iter", func() value.Value {
		return namespace("Iter",
			fn("Iter.range", "from, to", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				from, to := intArg(g, who, a[0]), intArg(g, who, a[1])
				return seq(who, func(g *interp.Goroutine, push func(value.Value) bool) {
					for i := from; i < to; i++ {
						if !push(i) {
							return
						}
					}
				})
			}),
			fn("Iter.fromArray", "xs", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				xs := arrayArg(g, who, a[0])
				return seq(who, func(g *interp.Goroutine, push func(value.Value) bool) {
					for i := 0; i < xs.Len(); i++ {
						v, ok := xs.Get(i)
						if !ok || !push(v) {
							return
						}
					}
				})
			}),
			fn("Iter.map", "s, f", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				src, f := funcArg(g, who, a[0]), funcArg(g, who, a[1])
				return seq(who, func(g *interp.Goroutine, push func(value.Value) bool) {
					drive(g, who, src, func(v value.Value) bool { return push(g.Call(f, v)) })
				})
			}),
			fn("Iter.filter", "s, f", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				src, f := funcArg(g, who, a[0]), funcArg(g, who, a[1])
				return seq(who, func(g *interp.Goroutine, push func(value.Value) bool) {
					drive(g, who, src, func(v value.Value) bool {
						if boolResult(g, who, g.Call(f, v)) {
							return push(v)
						}
						return true
					})
				})
			}),
			fn("Iter.take", "s, n", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				src := funcArg(g, who, a[0])
				n := intArg(g, who, a[1])
				if n < 0 {
					g.Panic("Iter.take needs a count of at least 0, got %d", n)
				}
				return seq(who, func(g *interp.Goroutine, push func(value.Value) bool) {
					if n == 0 {
						return
					}
					var taken int64
					drive(g, who, src, func(v value.Value) bool {
						taken++
						return push(v) && taken < n
					})
				})
			}),
			fn("Iter.toArray", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				var out []value.Value
				drive(g, who, a[0], func(v value.Value) bool {
					out = append(out, v)
					return true
				})
				return value.NewArray(out)
			}),
		)
	})
}
