<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.shiftLeft

```ant signature
func Bits.shiftLeft[T: Integer](a: T, n: int): T
```

Bits.shiftLeft returns a shifted left by n bits, filling with zeros. It panics if n \< 0 or n is at
least the bit width of T.

<!-- end generated -->
