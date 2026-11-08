<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.not

```ant signature
func Bits.not[T: Integer](a: T): T
```

Bits.not returns a with every bit inverted.

<!-- end generated -->

## Basic examples

```ant
Bits.not(uint8(0)) // 255
Bits.not(0)        // -1
```
