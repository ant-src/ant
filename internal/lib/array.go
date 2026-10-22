package lib

import (
	"math"
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
			fn("Array.first", "xs", func(g *interp.Goroutine, a []value.Value) value.Value {
				v, ok := arrayArg(g, "Array.first", a[0]).Get(0)
				if !ok {
					return value.Null
				}
				return v
			}),
			fn("Array.last", "xs", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.last", a[0])
				v, ok := xs.Get(xs.Len() - 1)
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
			fn("Array.insert", "xs, i, v", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.insert", a[0])
				i := intArg(g, "Array.insert", a[1])
				if i < 0 || i > int64(xs.Len()) || !xs.Insert(int(i), a[2]) {
					g.Panic("Array.insert: the index %d is not within 0..%d", i, xs.Len())
				}
				return value.Void
			}),
			fn("Array.removeAt", "xs, i", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.removeAt", a[0])
				i := intArg(g, "Array.removeAt", a[1])
				var v value.Value
				var ok bool
				if i >= 0 && i <= int64(xs.Len()) {
					v, ok = xs.Remove(int(i))
				}
				if !ok {
					g.Panic("Array.removeAt: the index %d is out of range for a length of %d", i, xs.Len())
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
			fn("Array.flatMap", "xs, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.flatMap", a[0]).Values()
				f := funcArg(g, "Array.flatMap", a[1])
				var out []value.Value
				for _, x := range xs {
					out = append(out, arrayArg(g, "the function given to Array.flatMap", g.Call(f, x)).Values()...)
				}
				return value.NewArray(out)
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
			fn("Array.findIndex", "xs, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.findIndex", a[0]).Values()
				f := funcArg(g, "Array.findIndex", a[1])
				for i, x := range xs {
					if boolResult(g, "Array.findIndex", g.Call(f, x)) {
						return int64(i)
					}
				}
				return value.Null
			}),
			fn("Array.findLast", "xs, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.findLast", a[0]).Values()
				f := funcArg(g, "Array.findLast", a[1])
				for i := len(xs) - 1; i >= 0; i-- {
					if boolResult(g, "Array.findLast", g.Call(f, xs[i])) {
						return xs[i]
					}
				}
				return value.Null
			}),
			fn("Array.findLastIndex", "xs, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.findLastIndex", a[0]).Values()
				f := funcArg(g, "Array.findLastIndex", a[1])
				for i := len(xs) - 1; i >= 0; i-- {
					if boolResult(g, "Array.findLastIndex", g.Call(f, xs[i])) {
						return int64(i)
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
			fn("Array.countWhere", "xs, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.countWhere", a[0]).Values()
				f := funcArg(g, "Array.countWhere", a[1])
				n := int64(0)
				for _, x := range xs {
					if boolResult(g, "Array.countWhere", g.Call(f, x)) {
						n++
					}
				}
				return n
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
			fn("Array.lastIndexOf", "xs, v", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.lastIndexOf", a[0]).Values()
				for i := len(xs) - 1; i >= 0; i-- {
					if value.Equal(xs[i], a[1]) {
						return int64(i)
					}
				}
				return value.Null
			}),
			extreme("Array.min", true),
			extreme("Array.max", false),
			extremeBy("Array.minBy", true),
			extremeBy("Array.maxBy", false),
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
			fn("Array.unique", "xs", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.unique", a[0]).Values()
				return value.NewArray(uniqueBy(xs, xs))
			}),
			fn("Array.uniqueBy", "xs, key", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.uniqueBy", a[0]).Values()
				f := funcArg(g, "Array.uniqueBy", a[1])
				keys := make([]value.Value, len(xs))
				for i, x := range xs {
					keys[i] = g.Call(f, x)
				}
				return value.NewArray(uniqueBy(xs, keys))
			}),
			fn("Array.chunk", "xs, n", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.chunk", a[0]).Values()
				n := intArg(g, "Array.chunk", a[1])
				if n < 1 {
					g.Panic("Array.chunk needs a size of at least 1, got %d", n)
				}
				var out []value.Value
				for len(xs) > 0 {
					k := min(int(n), len(xs))
					out = append(out, value.NewArray(append([]value.Value(nil), xs[:k]...)))
					xs = xs[k:]
				}
				return value.NewArray(out)
			}),
			fn("Array.groupBy", "xs, key", func(g *interp.Goroutine, a []value.Value) value.Value {
				xs := arrayArg(g, "Array.groupBy", a[0]).Values()
				f := funcArg(g, "Array.groupBy", a[1])
				m := value.NewMap()
				for _, x := range xs {
					k := g.Call(f, x)
					if group, ok := m.Get(k); ok {
						group.(*value.Array).Push(x)
					} else {
						m.Set(k, value.NewArray([]value.Value{x}))
					}
				}
				return m
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

// isNaN reports whether v is a float NaN.
func isNaN(v value.Value) bool {
	switch x := v.(type) {
	case float64:
		return math.IsNaN(x)
	case float32:
		return math.IsNaN(float64(x))
	}
	return false
}

// extreme is Array.min and Array.max: null on an empty array, NaN when any element is NaN, and
// otherwise the least or greatest element by <.
func extreme(name string, least bool) *interp.Builtin {
	return fn(name, "xs", func(g *interp.Goroutine, a []value.Value) value.Value {
		xs := arrayArg(g, name, a[0]).Values()
		if len(xs) == 0 {
			return value.Null
		}
		best := xs[0]
		if isNaN(best) {
			return best
		}
		for _, x := range xs[1:] {
			if isNaN(x) {
				return x
			}
			if least && less(g, name, x, best) || !least && less(g, name, best, x) {
				best = x
			}
		}
		return best
	})
}

// extremeBy is Array.minBy and Array.maxBy: the first element whose key is least or greatest.
// The held element is replaced only when a key is strictly less, or greater, than its key, so
// of equal keys the first wins and a NaN key never replaces one.
func extremeBy(name string, least bool) *interp.Builtin {
	return fn(name, "xs, key", func(g *interp.Goroutine, a []value.Value) value.Value {
		xs := arrayArg(g, name, a[0]).Values()
		f := funcArg(g, name, a[1])
		if len(xs) == 0 {
			return value.Null
		}
		best := 0
		bestKey := g.Call(f, xs[0])
		for i := 1; i < len(xs); i++ {
			k := g.Call(f, xs[i])
			if least && less(g, name, k, bestKey) || !least && less(g, name, bestKey, k) {
				best, bestKey = i, k
			}
		}
		return xs[best]
	})
}

// uniqueBy keeps the first of the elements whose keys are ==, in order. A map keyed by == is
// the set of keys seen, so the work is linear.
func uniqueBy(xs, keys []value.Value) []value.Value {
	seen := value.NewMap()
	out := make([]value.Value, 0, len(xs))
	for i, x := range xs {
		if seen.Has(keys[i]) {
			continue
		}
		seen.Set(keys[i], value.Void)
		out = append(out, x)
	}
	return out
}
