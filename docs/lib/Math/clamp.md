<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.clamp

```ant signature
func Math.clamp[T: Ordered](x: T, lo: T, hi: T): T
```

Math.clamp returns x held within lo ..= hi. It panics if lo > hi or a bound is NaN; a NaN x gives
NaN.

<!-- end generated -->

## Basic examples

```ant
Math.clamp(15, 0, 10) // 10
```
