<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# int64

```ant signature
func int64[T: Number](x: T): int64
```

int64 converts x to int64 exactly. It panics if x is out of range, has a fraction, or is NaN or
infinite; int(Math.trunc(x)) truncates. x may also be a named primitive over a number.

<!-- end generated -->
