package lib

import (
	"sort"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

func init() {
	register("Array", func() value.Value {
		return namespace("Array",
			fn("Array.length", "xs", func(g *interp.Goroutine, a []value.Value) value.Value {
				return int64(arrayArg(g, "Array.length", a[0]).Len())
			}),
			fn("Array.isEmpty", "xs", func(g *interp.Goroutine, a []value.Value) value.Value {
				return arrayArg(g, "Array.isEmpty", a[0]).Len() == 0
			}),
			fn("Array.get", "xs, i", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.get", a[0])
				i := intArg(g, "Array.get", a[1])
				v, ok := xs.Get(int(i))
				if !ok {
					return value.Null
				}
				return v
			}),
			fn("Array.push", "xs, v", func(g *interp.Goroutine, a []value.Value) value.Value {
				arrayArg(g, "Array.push", a[0]).Push(a[1])
				return value.Void
			}),
			fn("Array.pop", "xs", func(g *interp.Goroutine, a []value.Value) value.Value {
				v, ok := arrayArg(g, "Array.pop", a[0]).Pop()
				if !ok {
					return value.Null
				}
				return v
			}),
			fn("Array.slice", "xs, from, to", func(g *interp.Goroutine, a []value.Value) value.Value {
				vals := arrayArg(g, "Array.slice", a[0]).Values()
				from, to := intArg(g, "Array.slice", a[1]), intArg(g, "Array.slice", a[2])
				sliceBounds(g, "Array.slice", from, to, len(vals))
				return value.NewArray(append([]value.Value(nil), vals[from:to]...))
			}),
			fn("Array.concat", "xs, ys", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.concat", a[0]).Values()
				ys := arrayArg(g, "Array.concat", a[1]).Values()
				return value.NewArray(append(append(make([]value.Value, 0, len(xs)+len(ys)), xs...), ys...))
			}),
			fn("Array.map", "xs, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.map", a[0]).Values()
				f := funcArg(g, "Array.map", a[1])
				out := make([]value.Value, len(xs))
				for i, x := range xs {
					out[i] = g.Call(f, x)
				}
				return value.NewArray(out)
			}),
			fn("Array.filter", "xs, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.filter", a[0]).Values()
				f := funcArg(g, "Array.filter", a[1])
				out := make([]value.Value, 0, len(xs))
				for _, x := range xs {
					if boolResult(g, "Array.filter", g.Call(f, x)) {
						out = append(out, x)
					}
				}
				return value.NewArray(out)
			}),
			fn("Array.reduce", "xs, init, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.reduce", a[0]).Values()
				f := funcArg(g, "Array.reduce", a[2])
				acc := a[1]
				for _, x := range xs {
					acc = g.Call(f, acc, x)
				}
				return acc
			}),
			fn("Array.find", "xs, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.find", a[0]).Values()
				f := funcArg(g, "Array.find", a[1])
				for _, x := range xs {
					if boolResult(g, "Array.find", g.Call(f, x)) {
						return x
					}
				}
				return value.Null
			}),
			fn("Array.some", "xs, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.some", a[0]).Values()
				f := funcArg(g, "Array.some", a[1])
				for _, x := range xs {
					if boolResult(g, "Array.some", g.Call(f, x)) {
						return true
					}
				}
				return false
			}),
			fn("Array.every", "xs, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.every", a[0]).Values()
				f := funcArg(g, "Array.every", a[1])
				for _, x := range xs {
					if !boolResult(g, "Array.every", g.Call(f, x)) {
						return false
					}
				}
				return true
			}),
			fn("Array.contains", "xs, v", func(g *interp.Goroutine, a []value.Value) value.Value {
				for _, x := range arrayArg(g, "Array.contains", a[0]).Values() {
					if value.Equal(x, a[1]) {
						return true
					}
				}
				return false
			}),
			fn("Array.indexOf", "xs, v", func(g *interp.Goroutine, a []value.Value) value.Value {
				for i, x := range arrayArg(g, "Array.indexOf", a[0]).Values() {
					if value.Equal(x, a[1]) {
						return int64(i)
					}
				}
				return value.Null
			}),
			fn("Array.sort", "xs", func(g *interp.Goroutine, a []value.Value) value.Value {
				vals := arrayArg(g, "Array.sort", a[0]).Values()
				sort.SliceStable(vals, func(i, j int) bool { return less(g, "Array.sort", vals[i], vals[j]) })
				return value.NewArray(vals)
			}),
			fn("Array.sortBy", "xs, key", func(g *interp.Goroutine, a []value.Value) value.Value {
				vals := arrayArg(g, "Array.sortBy", a[0]).Values()
				f := funcArg(g, "Array.sortBy", a[1])
				keys := make([]value.Value, len(vals))
				for i, x := range vals {
					keys[i] = g.Call(f, x)
				}
				idx := make([]int, len(vals))
				for i := range idx {
					idx[i] = i
				}
				sort.SliceStable(idx, func(i, j int) bool { return less(g, "Array.sortBy", keys[idx[i]], keys[idx[j]]) })
				out := make([]value.Value, len(vals))
				for i, k := range idx {
					out[i] = vals[k]
				}
				return value.NewArray(out)
			}),
			fn("Array.sortWith", "xs, cmp", func(g *interp.Goroutine, a []value.Value) value.Value {
				vals := arrayArg(g, "Array.sortWith", a[0]).Values()
				f := funcArg(g, "Array.sortWith", a[1])
				sort.SliceStable(vals, func(i, j int) bool {
					return intResult(g, "Array.sortWith", g.Call(f, vals[i], vals[j])) < 0
				})
				return value.NewArray(vals)
			}),
			fn("Array.reverse", "xs", func(g *interp.Goroutine, a []value.Value) value.Value {
				vals := arrayArg(g, "Array.reverse", a[0]).Values()
				for i, j := 0, len(vals)-1; i < j; i, j = i+1, j-1 {
					vals[i], vals[j] = vals[j], vals[i]
				}
				return value.NewArray(vals)
			}),
			fn("Array.entries", "xs", func(g *interp.Goroutine, a []value.Value) value.Value {
				vals := arrayArg(g, "Array.entries", a[0]).Values()
				out := make([]value.Value, len(vals))
				for i, x := range vals {
					out[i] = value.Indexed(int64(i), x)
				}
				return value.NewArray(out)
			}),
			fn("Array.range", "from, to", func(g *interp.Goroutine, a []value.Value) value.Value {
				from, to := intArg(g, "Array.range", a[0]), intArg(g, "Array.range", a[1])
				var out []value.Value
				for i := from; i < to; i++ {
					out = append(out, i)
				}
				return value.NewArray(out)
			}),
			fn("Array.filled", "n, v", func(g *interp.Goroutine, a []value.Value) value.Value {
				n := intArg(g, "Array.filled", a[0])
				if n < 0 {
					g.Panic("Array.filled needs a count of at least 0, got %d", n)
				}
				out := make([]value.Value, n)
				for i := range out {
					out[i] = a[1]
				}
				return value.NewArray(out)
			}),
			fn("Array.flatten", "xs", func(g *interp.Goroutine, a []value.Value) value.Value {
				var out []value.Value
				for _, x := range arrayArg(g, "Array.flatten", a[0]).Values() {
					out = append(out, arrayArg(g, "Array.flatten", x).Values()...)
				}
				return value.NewArray(out)
			}),
			fn("Array.sum", "xs", func(g *interp.Goroutine, a []value.Value) value.Value {
				vals := arrayArg(g, "Array.sum", a[0]).Values()
				if len(vals) == 0 {
					return value.Null
				}
				acc := vals[0]
				if !value.KindOf(acc).IsNumber() {
					g.Panic("Array.sum needs numbers, got %s", value.Describe(acc))
				}
				for _, x := range vals[1:] {
					v, err := value.Binary("+", acc, x)
					if err != nil {
						g.Panic("Array.sum: %s", err)
					}
					acc = v
				}
				return acc
			}),
			fn("Array.same", "xs, ys", func(g *interp.Goroutine, a []value.Value) value.Value {
				return arrayArg(g, "Array.same", a[0]) == arrayArg(g, "Array.same", a[1])
			}),
		)
	})
}
