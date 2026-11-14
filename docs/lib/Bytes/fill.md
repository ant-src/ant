<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.fill

```ant signature
func Bytes.fill(b: Bytes, x: byte, from: int = 0, to: int | null = null): void
```

Bytes.fill sets the bytes of b from index from up to but excluding index to, the length when null, to x. It panics as Bytes.slice does.

<!-- end generated -->
