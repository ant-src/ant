package lib

import (
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

func init() {
	register("Map", func() value.Value {
		return namespace("Map",
			fn("Map.size", "m", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return int64(mapArg(g, who, a[0]).Len())
			}),
			fn("Map.isEmpty", "m", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return mapArg(g, who, a[0]).Len() == 0
			}),
			fn("Map.has", "m, k", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return mapArg(g, who, a[0]).Has(a[1])
			}),
			fn("Map.get", "m, k", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				v, ok := mapArg(g, who, a[0]).Get(a[1])
				if !ok {
					return value.Null
				}
				return v
			}),
			fn("Map.set", "m, k, v", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				mapArg(g, who, a[0]).Set(a[1], a[2])
				return value.Void
			}),
			fn("Map.remove", "m, k", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				v, ok := mapArg(g, who, a[0]).Remove(a[1])
				if !ok {
					return value.Null
				}
				return v
			}),
			fn("Map.keys", "m", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return value.NewArray(mapArg(g, who, a[0]).Keys())
			}),
			fn("Map.values", "m", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return value.NewArray(mapArg(g, who, a[0]).Values())
			}),
			fn("Map.entries", "m", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return value.NewArray(entries(mapArg(g, who, a[0])))
			}),
			fn("Map.fromEntries", "entries", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				m := value.NewMap()
				for _, e := range arrayArg(g, who, a[0]).Values() {
					k, v := entryArg(g, who, e)
					m.Set(k, v)
				}
				return m
			}),
			fn("Map.map", "m, f", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				f := funcArg(g, who, a[1])
				out := value.NewMap()
				for _, e := range mapArg(g, who, a[0]).Entries() {
					out.Set(e.Key, g.Call(f, value.Entry(e.Key, e.Value)))
				}
				return out
			}),
			fn("Map.filter", "m, f", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				f := funcArg(g, who, a[1])
				out := value.NewMap()
				for _, e := range mapArg(g, who, a[0]).Entries() {
					if boolResult(g, who, g.Call(f, value.Entry(e.Key, e.Value))) {
						out.Set(e.Key, e.Value)
					}
				}
				return out
			}),
			fn("Map.merge", "m, n", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				out := value.NewMap()
				for _, e := range mapArg(g, who, a[0]).Entries() {
					out.Set(e.Key, e.Value)
				}
				for _, e := range mapArg(g, who, a[1]).Entries() {
					out.Set(e.Key, e.Value)
				}
				return out
			}),
			fn("Map.same", "m, n", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return mapArg(g, who, a[0]) == mapArg(g, who, a[1])
			}),
		)
	})
	register("Record", func() value.Value {
		return namespace("Record",
			fn("Record.same", "a, b", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				x, ok := a[0].(*value.Record)
				if !ok {
					g.Panic("Record.same needs a record, got %s", value.Describe(a[0]))
				}
				y, ok := a[1].(*value.Record)
				if !ok {
					g.Panic("Record.same needs a record, got %s", value.Describe(a[1]))
				}
				return x == y
			}),
		)
	})
}

func entries(m *value.Map) []value.Value {
	es := m.Entries()
	out := make([]value.Value, len(es))
	for i, e := range es {
		out[i] = value.Entry(e.Key, e.Value)
	}
	return out
}

// entryArg requires an Entry record and returns its key and value.
func entryArg(g *interp.Goroutine, who string, v value.Value) (value.Value, value.Value) {
	r, ok := v.(*value.Record)
	if !ok || r.Tag != value.EntryTag {
		g.Panic("%s needs Entry records, got %s", who, value.Describe(v))
	}
	return r.Get(0), r.Get(1)
}
