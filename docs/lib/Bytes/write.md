<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.write

```ant signature
func Bytes.write[T: Number](b: Bytes, t: Reflect.Type[T], at: int, x: T, order: Bytes.Order): void
```

Bytes.write overwrites the bytes of b at index at with x in the given order. It never grows b, and panics unless the number's bytes are within 0 ..= Bytes.length(b).

<!-- end generated -->
