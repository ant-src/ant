<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.rotateRight

```ant signature
func Bits.rotateRight[T: Integer](a: T, n: int): T
```

Bits.rotateRight returns a rotated right by n bits. The count is taken modulo the bit width of T, so
n may exceed it. It panics if n < 0.

<!-- end generated -->
