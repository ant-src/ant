<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.min

```ant signature
func Array.min[T: Ordered](xs: Array[T]): T | null
```

Array.min returns the least element of xs, or null when xs is empty. It returns NaN when any element
is NaN.

<!-- end generated -->
