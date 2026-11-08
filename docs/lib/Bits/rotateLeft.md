<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.rotateLeft

```ant signature
func Bits.rotateLeft[T: Integer](a: T, n: int): T
```

Bits.rotateLeft returns a rotated left by n bits. The count is taken modulo the bit width of T, so n
may exceed it. It panics if n \< 0.

<!-- end generated -->

## Basic examples

```ant
Bits.rotateLeft(uint8(129), 1) // 3
Bits.rotateLeft(uint8(1), 9)   // 2
```
