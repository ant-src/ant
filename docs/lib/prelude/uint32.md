<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# uint32

```ant signature
func uint32[T: Number](x: T): uint32
```

uint32 converts x to uint32 exactly. It panics if x is out of range, has a fraction, or is NaN or
infinite; int(Math.trunc(x)) truncates. x may also be a named primitive over a number.

<!-- end generated -->
