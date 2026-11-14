<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.read

```ant signature
func Bytes.read[T: Number](b: Bytes, t: Reflect.Type[T], at: int, order: Bytes.Order): T
```

Bytes.read returns the number of type t stored in b at index at in the given order. It panics unless the number's bytes are within 0 ..= Bytes.length(b).

<!-- end generated -->
