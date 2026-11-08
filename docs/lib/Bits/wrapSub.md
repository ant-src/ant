<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.wrapSub

```ant signature
func Bits.wrapSub[T: Integer](a: T, b: T): T
```

Bits.wrapSub returns a - b, wrapping around in two's complement instead of panicking on overflow.

<!-- end generated -->

## Basic examples

```ant
Bits.wrapSub(uint8(0), uint8(1)) // 255
```
