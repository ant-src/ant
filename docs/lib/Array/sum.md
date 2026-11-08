<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.sum

```ant signature
func Array.sum[T: Number](xs: Array[T]): T | null
```

Array.sum returns the sum of the elements of xs, or null when xs is empty. Write Array.sum(xs) ?? 0
for a zero default. It panics on overflow.

<!-- end generated -->
