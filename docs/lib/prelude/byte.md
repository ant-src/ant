<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# byte

```ant signature
func byte[T: Number](x: T): byte
```

byte converts x to byte, which is uint8, exactly. It panics if x is out of range, has a fraction, or
is NaN or infinite; int(Math.trunc(x)) truncates. x may also be a named primitive over a number.

<!-- end generated -->
