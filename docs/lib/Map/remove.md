<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.remove

```ant signature
func Map.remove[K, V](m: Map[K, V], k: K): V | null
```

Map.remove removes the entry for k from m and returns its value, or null when k is missing. It
mutates m.

<!-- end generated -->
