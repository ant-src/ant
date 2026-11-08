<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.map

```ant signature
func Map.map[K, V, W](m: Map[K, V], f: func(Entry[K, V]): W): Map[K, W]
```

Map.map returns a new map with the keys of m, each mapped to f applied to its entry. It keeps the
order of m.

<!-- end generated -->

## Basic examples

```ant
Map.map(Map{ a: 1, b: 2 }, { x.value * 10 }) // Map{ "a": 10, "b": 20 }
```
