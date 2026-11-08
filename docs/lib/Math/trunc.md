<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.trunc

```ant signature
func Math.trunc(x: float64): float64
```

Math.trunc returns x with its fraction dropped, rounding toward zero. int(Math.trunc(x)) is the
explicit truncating conversion to int.

<!-- end generated -->

## Basic examples

```ant
Math.trunc(-2.7) // -2.0
```
