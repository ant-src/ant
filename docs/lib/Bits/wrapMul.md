<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.wrapMul

```ant signature
func Bits.wrapMul[T: Integer](a: T, b: T): T
```

Bits.wrapMul returns a \* b, wrapping around in two's complement instead of panicking on overflow.

<!-- end generated -->

## Basic examples

```ant
Bits.wrapMul(int8(64), int8(2)) // -128
```
