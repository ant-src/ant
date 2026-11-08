<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.toInt

```ant signature
func Math.toInt(x: float64): int | null
```

Math.toInt returns x as an int, or null if x has a fraction, is out of range, or is NaN or an
infinity. It is int(x) without the panic.

<!-- end generated -->

## Basic examples

```ant
Math.toInt(2.0) // 2
Math.toInt(2.5) // null
```
