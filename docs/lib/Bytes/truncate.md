<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.truncate

```ant signature
func Bytes.truncate(b: Bytes, n: int): void
```

Bytes.truncate keeps the first n bytes of b. It panics if n is outside 0 ..= Bytes.length(b).

<!-- end generated -->
