<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.wrapAdd

```ant signature
func Bits.wrapAdd[T: Integer](a: T, b: T): T
```

Bits.wrapAdd returns a + b, wrapping around in two's complement instead of panicking on overflow.

<!-- end generated -->

## Basic examples

```ant
Bits.wrapAdd(uint8(255), uint8(1)) // 0
```
