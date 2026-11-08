<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.find

```ant signature
func Array.find[T](xs: Array[T], f: func(T): bool): T | null
```

Array.find returns the first element of xs that f accepts, or null when there is none.

<!-- end generated -->
