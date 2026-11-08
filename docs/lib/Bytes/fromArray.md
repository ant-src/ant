<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.fromArray

```ant signature
func Bytes.fromArray[T: Integer](xs: Array[T]): Bytes
```

Bytes.fromArray returns the bytes holding the integers in xs, one byte each. It panics if an element
is outside 0 ..= 255.

<!-- end generated -->

## Basic examples

```ant
Bytes.fromArray([0x89, 0x50, 0x4e, 0x47]) // Bytes(89 50 4e 47)
```
