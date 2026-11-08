<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.filter

```ant signature
func Map.filter[K, V](m: Map[K, V], f: func(Entry[K, V]): bool): Map[K, V]
```

Map.filter returns a new map holding the entries of m that f accepts, in the order of m.

<!-- end generated -->
