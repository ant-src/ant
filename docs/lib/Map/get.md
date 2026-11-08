<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.get

```ant signature
func Map.get[K, V](m: Map[K, V], k: K): V | null
```

Map.get returns the value m holds for k, or null when k is missing. It is the same as m[k]; use
[Map.has](has.md) when V itself includes null.

<!-- end generated -->
