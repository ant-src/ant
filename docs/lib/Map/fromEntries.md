<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.fromEntries

```ant signature
func Map.fromEntries[K, V](entries: Array[Entry[K, V]]): Map[K, V]
```

Map.fromEntries returns a new map holding entries, in order. When a key repeats, the later entry's
value wins.

<!-- end generated -->
