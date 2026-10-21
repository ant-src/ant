package lib

import (
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// A sequence is a function that pushes elements to a yield function until it returns false
// (section 5). The Iter functions build sequences as builtins over other sequences.

// seq builds a sequence function from a Go loop body. push sends one element and reports
// whether the consumer wants more.
func seq(name string, run func(g *interp.Goroutine, push func(value.Value) bool)) *interp.Builtin {
	return fn(name, "yield", func(g *interp.Goroutine, a []value.Value) value.Value {
		y := funcArg(g, name, a[0])
		run(g, func(v value.Value) bool {
			return boolResult(g, "a sequence's yield", g.Call(y, v))
		})
		return value.Void
	})
}

// drive runs a sequence with a yield that calls push and stops the source once push says so.
func drive(g *interp.Goroutine, who string, s value.Value, push func(value.Value) bool) {
	more := true
	y := fn("yield", "x", func(g *interp.Goroutine, a []value.Value) value.Value {
		if !more {
			return false
		}
		more = push(a[0])
		return more
	})
	g.Call(funcArg(g, who, s), y)
}

func init() {
	register("Iter", func() value.Value {
		return namespace("Iter",
			fn("Iter.range", "from, to", func(g *interp.Goroutine, a []value.Value) value.Value {
				from, to := intArg(g, "Iter.range", a[0]), intArg(g, "Iter.range", a[1])
				return seq("Iter.range", func(g *interp.Goroutine, push func(value.Value) bool) {
					for i := from; i < to; i++ {
						if !push(i) {
							return
						}
					}
				})
			}),
			fn("Iter.fromArray", "xs", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Iter.fromArray", a[0])
				return seq("Iter.fromArray", func(g *interp.Goroutine, push func(value.Value) bool) {
					for i := 0; i < xs.Len(); i++ {
						v, ok := xs.Get(i)
						if !ok || !push(v) {
							return
						}
					}
				})
			}),
			fn("Iter.map", "s, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				src, f := funcArg(g, "Iter.map", a[0]), funcArg(g, "Iter.map", a[1])
				return seq("Iter.map", func(g *interp.Goroutine, push func(value.Value) bool) {
					drive(g, "Iter.map", src, func(v value.Value) bool { return push(g.Call(f, v)) })
				})
			}),
			fn("Iter.filter", "s, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				src, f := funcArg(g, "Iter.filter", a[0]), funcArg(g, "Iter.filter", a[1])
				return seq("Iter.filter", func(g *interp.Goroutine, push func(value.Value) bool) {
					drive(g, "Iter.filter", src, func(v value.Value) bool {
						if boolResult(g, "Iter.filter", g.Call(f, v)) {
							return push(v)
						}
						return true
					})
				})
			}),
			fn("Iter.take", "s, n", func(g *interp.Goroutine, a []value.Value) value.Value {
				src := funcArg(g, "Iter.take", a[0])
				n := intArg(g, "Iter.take", a[1])
				if n < 0 {
					g.Panic("Iter.take needs a count of at least 0, got %d", n)
				}
				return seq("Iter.take", func(g *interp.Goroutine, push func(value.Value) bool) {
					if n == 0 {
						return
					}
					var taken int64
					drive(g, "Iter.take", src, func(v value.Value) bool {
						taken++
						return push(v) && taken < n
					})
				})
			}),
			fn("Iter.toArray", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				var out []value.Value
				drive(g, "Iter.toArray", a[0], func(v value.Value) bool {
					out = append(out, v)
					return true
				})
				return value.NewArray(out)
			}),
		)
	})
}
