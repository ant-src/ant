<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.merge

```ant signature
func Map.merge[K, V](m: Map[K, V], n: Map[K, V]): Map[K, V]
```

Map.merge returns a new map holding the entries of m and n, where the entries of n win. The keys of
m come first, followed by the keys of n that m lacks.

<!-- end generated -->

## Basic examples

```ant
Map.merge(Map{ a: 1, b: 2 }, Map{ b: 20, c: 30 }) // Map{ "a": 1, "b": 20, "c": 30 }
```
