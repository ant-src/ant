<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.sortBy

```ant signature
func Array.sortBy[T, K: Ordered](xs: Array[T], key: func(T): K): Array[T]
```

Array.sortBy returns a new array holding the elements of xs in ascending order of key. The sort is
stable.

<!-- end generated -->
