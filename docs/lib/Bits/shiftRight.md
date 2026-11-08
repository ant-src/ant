<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.shiftRight

```ant signature
func Bits.shiftRight[T: Integer](a: T, n: int): T
```

Bits.shiftRight returns a shifted right by n bits. The shift is arithmetic for a signed type and
logical for an unsigned one. It panics if n \< 0 or n is at least the bit width of T.

<!-- end generated -->

## Basic examples

```ant
Bits.shiftRight(int8(-8), 1)   // -4
Bits.shiftRight(uint8(248), 1) // 124
```
