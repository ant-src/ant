<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.abs

```ant signature
func Math.abs[T: Number](x: T): T
```

Math.abs returns the absolute value of x. It panics on the most negative value of an integer type,
whose absolute value does not fit.

<!-- end generated -->
