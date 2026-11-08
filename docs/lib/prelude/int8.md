<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# int8

```ant signature
func int8[T: Number](x: T): int8
```

int8 converts x to int8 exactly. It panics if x is out of range, has a fraction, or is NaN or
infinite; int(Math.trunc(x)) truncates. x may also be a named primitive over a number.

<!-- end generated -->
