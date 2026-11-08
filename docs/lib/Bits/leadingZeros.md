<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.leadingZeros

```ant signature
func Bits.leadingZeros[T: Integer](a: T): int
```

Bits.leadingZeros returns the number of zero bits above the highest one bit of a, counted at the bit
width of T over its two's complement bits. It is the full width for 0.

<!-- end generated -->

## Basic examples

```ant
Bits.leadingZeros(uint8(1)) // 7
Bits.leadingZeros(int8(-1)) // 0
```
