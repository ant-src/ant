<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.set

```ant signature
func Map.set[K, V](m: Map[K, V], k: K, v: V): void
```

Map.set stores v under k in m, replacing any value already there. It mutates m and is the same as
m\[k\] = v. A replaced entry keeps its position; a new one goes to the end.

<!-- end generated -->
