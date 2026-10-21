package lib

import (
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// result requires an Ok or an Err.
func result(g *interp.Goroutine, who string, v value.Value) *value.Record {
	r, ok := v.(*value.Record)
	if !ok || (r.Tag != value.OkTag && r.Tag != value.ErrTag) {
		g.Panic("%s needs a Result, got %s", who, value.Describe(v))
	}
	return r
}

func resultFuncs() []*interp.Builtin {
	return []*interp.Builtin{
		fn("Result.ok", "value", func(g *interp.Goroutine, a []value.Value) value.Value { return value.Ok(a[0]) }),
		fn("Result.err", "error", func(g *interp.Goroutine, a []value.Value) value.Value { return value.Err(a[0]) }),
		fn("Result.isOk", "r", func(g *interp.Goroutine, a []value.Value) value.Value {
			return result(g, "Result.isOk", a[0]).Tag == value.OkTag
		}),
		fn("Result.map", "r, f", func(g *interp.Goroutine, a []value.Value) value.Value {
			r := result(g, "Result.map", a[0])
			if r.Tag != value.OkTag {
				return r
			}
			return value.Ok(g.Call(a[1], r.Get(0)))
		}),
		fn("Result.mapErr", "r, f", func(g *interp.Goroutine, a []value.Value) value.Value {
			r := result(g, "Result.mapErr", a[0])
			if r.Tag != value.ErrTag {
				return r
			}
			return value.Err(g.Call(a[1], r.Get(0)))
		}),
		fn("Result.unwrap", "r", func(g *interp.Goroutine, a []value.Value) value.Value {
			r := result(g, "Result.unwrap", a[0])
			if r.Tag != value.OkTag {
				g.Panic("Result.unwrap on an Err: %s", value.RenderQuoted(r.Get(0)))
			}
			return r.Get(0)
		}),
		fn("Result.unwrapOr", "r, fallback", func(g *interp.Goroutine, a []value.Value) value.Value {
			r := result(g, "Result.unwrapOr", a[0])
			if r.Tag != value.OkTag {
				return a[1]
			}
			return r.Get(0)
		}),
		fn("Result.all", "rs", func(g *interp.Goroutine, a []value.Value) value.Value {
			arr, ok := a[0].(*value.Array)
			if !ok {
				g.Panic("Result.all needs an array, got %s", value.Describe(a[0]))
			}
			var values []value.Value
			for _, v := range arr.Values() {
				r := result(g, "Result.all", v)
				if r.Tag == value.ErrTag {
					return r
				}
				values = append(values, r.Get(0))
			}
			return value.Ok(value.NewArray(values))
		}),
	}
}
