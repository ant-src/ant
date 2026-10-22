package lib

import (
	"math"
	"math/rand/v2"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

func init() {
	register("Math", func() value.Value {
		return namespace("Math",
			fn("Math.abs", "x", func(g *interp.Goroutine, a []value.Value) value.Value {
				v := a[0]
				if !value.KindOf(v).IsNumber() {
					g.Panic("Math.abs needs a number, got %s", value.Describe(v))
				}
				neg, err := value.Compare("<", v, zeroOf(v))
				if err != nil {
					g.Panic("Math.abs: %s", err)
				}
				if !neg {
					return v
				}
				r, err := value.Neg(v)
				if err != nil {
					g.Panic("Math.abs: %s", err)
				}
				return r
			}),
			fn("Math.min", "a, b", func(g *interp.Goroutine, a []value.Value) value.Value {
				if less(g, "Math.min", a[1], a[0]) {
					return a[1]
				}
				return a[0]
			}),
			fn("Math.max", "a, b", func(g *interp.Goroutine, a []value.Value) value.Value {
				if less(g, "Math.max", a[0], a[1]) {
					return a[1]
				}
				return a[0]
			}),
			fn("Math.pow", "x, y", func(g *interp.Goroutine, a []value.Value) value.Value {
				return math.Pow(floatArg(g, "Math.pow", a[0]), floatArg(g, "Math.pow", a[1]))
			}),
			fn("Math.sqrt", "x", func(g *interp.Goroutine, a []value.Value) value.Value {
				return math.Sqrt(floatArg(g, "Math.sqrt", a[0]))
			}),
			fn("Math.floor", "x", func(g *interp.Goroutine, a []value.Value) value.Value {
				return math.Floor(floatArg(g, "Math.floor", a[0]))
			}),
			fn("Math.ceil", "x", func(g *interp.Goroutine, a []value.Value) value.Value {
				return math.Ceil(floatArg(g, "Math.ceil", a[0]))
			}),
			fn("Math.round", "x", func(g *interp.Goroutine, a []value.Value) value.Value {
				return math.Round(floatArg(g, "Math.round", a[0]))
			}),
			fn("Math.trunc", "x", func(g *interp.Goroutine, a []value.Value) value.Value {
				return math.Trunc(floatArg(g, "Math.trunc", a[0]))
			}),
			fn("Math.toInt", "x", func(g *interp.Goroutine, a []value.Value) value.Value {
				v, err := value.Convert(value.KindInt64, floatArg(g, "Math.toInt", a[0]))
				if err != nil {
					return value.Null
				}
				return v
			}),
			unary("Math.log", math.Log),
			unary("Math.log2", math.Log2),
			unary("Math.log10", math.Log10),
			unary("Math.exp", math.Exp),
			unary("Math.sin", math.Sin),
			unary("Math.cos", math.Cos),
			unary("Math.tan", math.Tan),
			unary("Math.asin", math.Asin),
			unary("Math.acos", math.Acos),
			unary("Math.atan", math.Atan),
			fn("Math.atan2", "y, x", func(g *interp.Goroutine, a []value.Value) value.Value {
				return math.Atan2(floatArg(g, "Math.atan2", a[0]), floatArg(g, "Math.atan2", a[1]))
			}),
			fn("Math.hypot", "x, y", func(g *interp.Goroutine, a []value.Value) value.Value {
				return math.Hypot(floatArg(g, "Math.hypot", a[0]), floatArg(g, "Math.hypot", a[1]))
			}),
			unary("Math.cbrt", math.Cbrt),
			fn("Math.isNaN", "x", func(g *interp.Goroutine, a []value.Value) value.Value {
				return math.IsNaN(floatArg(g, "Math.isNaN", a[0]))
			}),
			fn("Math.isFinite", "x", func(g *interp.Goroutine, a []value.Value) value.Value {
				x := floatArg(g, "Math.isFinite", a[0])
				return !math.IsNaN(x) && !math.IsInf(x, 0)
			}),
			fn("Math.clamp", "x, lo, hi", func(g *interp.Goroutine, a []value.Value) value.Value {
				x, lo, hi := a[0], a[1], a[2]
				if less(g, "Math.clamp", hi, lo) {
					g.Panic("Math.clamp needs lo <= hi, got %s and %s", value.Render(lo), value.Render(hi))
				}
				if less(g, "Math.clamp", x, lo) {
					return lo
				}
				if less(g, "Math.clamp", hi, x) {
					return hi
				}
				return x
			}),
			fn("Math.random", "", func(g *interp.Goroutine, a []value.Value) value.Value {
				return rand.Float64()
			}),
			fn("Math.randomInt", "from, to", func(g *interp.Goroutine, a []value.Value) value.Value {
				from, to := intArg(g, "Math.randomInt", a[0]), intArg(g, "Math.randomInt", a[1])
				if to <= from {
					g.Panic("Math.randomInt needs from < to, got %d and %d", from, to)
				}
				// The span may exceed the int range, so it is taken as a uint64.
				return int64(uint64(from) + rand.Uint64N(uint64(to)-uint64(from)))
			}),
			constant("PI", math.Pi),
			constant("E", math.E),
			constant("NaN", math.NaN()),
			constant("Infinity", math.Inf(1)),
			constant("maxFloat64", math.MaxFloat64),
			constant("minInt", int64(math.MinInt64)),
			constant("maxInt", int64(math.MaxInt64)),
		)
	})
}

// unary is a Math function over one float64.
func unary(name string, f func(float64) float64) *interp.Builtin {
	return fn(name, "x", func(g *interp.Goroutine, a []value.Value) value.Value {
		return f(floatArg(g, name, a[0]))
	})
}

// zeroOf returns the zero of v's numeric kind.
func zeroOf(v value.Value) value.Value {
	if t, ok := v.(*value.Tagged); ok {
		return &value.Tagged{Tag: t.Tag, Payload: zeroOf(t.Payload)}
	}
	z, _ := value.Convert(value.KindOf(v), int64(0))
	return z
}
