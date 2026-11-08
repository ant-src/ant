<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.has

```ant signature
func Map.has[K, V](m: Map[K, V], k: K): bool
```

Map.has reports whether m has an entry for k. When V includes null, it is the only way to tell a
missing key from a stored null.

<!-- end generated -->
