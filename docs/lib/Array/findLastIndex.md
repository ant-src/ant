<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.findLastIndex

```ant signature
func Array.findLastIndex[T](xs: Array[T], f: func(T): bool): int | null
```

Array.findLastIndex returns the index of the last element of xs that f accepts, or null when there
is none.

<!-- end generated -->
