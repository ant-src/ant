package lib

import (
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

func init() {
	register("Set", func() value.Value {
		return namespace("Set",
			fn("Set.of", "xs", func(g *interp.Goroutine, a []value.Value) value.Value {
				s := value.NewSet()
				for _, x := range arrayArg(g, "Set.of", a[0]).Values() {
					s.Add(x)
				}
				return s
			}),
			fn("Set.size", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				return int64(setArg(g, "Set.size", a[0]).Len())
			}),
			fn("Set.isEmpty", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				return setArg(g, "Set.isEmpty", a[0]).Len() == 0
			}),
			fn("Set.has", "s, x", func(g *interp.Goroutine, a []value.Value) value.Value {
				return setArg(g, "Set.has", a[0]).Has(a[1])
			}),
			fn("Set.add", "s, x", func(g *interp.Goroutine, a []value.Value) value.Value {
				return setArg(g, "Set.add", a[0]).Add(a[1])
			}),
			fn("Set.remove", "s, x", func(g *interp.Goroutine, a []value.Value) value.Value {
				return setArg(g, "Set.remove", a[0]).Remove(a[1])
			}),
			fn("Set.values", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				return value.NewArray(setArg(g, "Set.values", a[0]).Values())
			}),
			fn("Set.union", "a, b", func(g *interp.Goroutine, a []value.Value) value.Value {
				x, y := setArg(g, "Set.union", a[0]), setArg(g, "Set.union", a[1])
				out := value.NewSet()
				for _, v := range x.Values() {
					out.Add(v)
				}
				for _, v := range y.Values() {
					out.Add(v)
				}
				return out
			}),
			fn("Set.intersect", "a, b", func(g *interp.Goroutine, a []value.Value) value.Value {
				x, y := setArg(g, "Set.intersect", a[0]), setArg(g, "Set.intersect", a[1])
				return setWhere(x, func(v value.Value) bool { return y.Has(v) })
			}),
			fn("Set.difference", "a, b", func(g *interp.Goroutine, a []value.Value) value.Value {
				x, y := setArg(g, "Set.difference", a[0]), setArg(g, "Set.difference", a[1])
				return setWhere(x, func(v value.Value) bool { return !y.Has(v) })
			}),
			fn("Set.isSubset", "a, b", func(g *interp.Goroutine, a []value.Value) value.Value {
				x, y := setArg(g, "Set.isSubset", a[0]), setArg(g, "Set.isSubset", a[1])
				for _, v := range x.Values() {
					if !y.Has(v) {
						return false
					}
				}
				return true
			}),
			fn("Set.same", "a, b", func(g *interp.Goroutine, a []value.Value) value.Value {
				return setArg(g, "Set.same", a[0]) == setArg(g, "Set.same", a[1])
			}),
		)
	})
}

// setWhere is the set of the members of s that keep accepts, in the order of s.
func setWhere(s *value.Set, keep func(value.Value) bool) *value.Set {
	out := value.NewSet()
	for _, v := range s.Values() {
		if keep(v) {
			out.Add(v)
		}
	}
	return out
}
