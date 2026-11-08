<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.maxBy

```ant signature
func Array.maxBy[T, K: Ordered](xs: Array[T], key: func(T): K): T | null
```

Array.maxBy returns the first element of xs with the greatest key, or null when xs is empty. An
element replaces the one held only when its key is strictly greater, so of equal keys the first wins
and a NaN key never wins.

<!-- end generated -->
