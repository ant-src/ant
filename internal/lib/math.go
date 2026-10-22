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
			fn("Math.random", "", func(g *interp.Goroutine, a []value.Value) value.Value {
				return rand.Float64()
			}),
			constant("PI", math.Pi),
			constant("E", math.E),
			constant("minInt", int64(math.MinInt64)),
			constant("maxInt", int64(math.MaxInt64)),
		)
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
