<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.max

```ant signature
func Array.max[T: Ordered](xs: Array[T]): T | null
```

Array.max returns the greatest element of xs, or null when xs is empty. It returns NaN when any
element is NaN.

<!-- end generated -->
