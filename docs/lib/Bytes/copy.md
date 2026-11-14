<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.copy

```ant signature
func Bytes.copy(dst: Bytes, at: int, src: Bytes, from: int = 0, to: int | null = null): void
```

Bytes.copy overwrites the bytes of dst from index at with those of src from index from up to but excluding index to, the length of src when null, as if src were read first. It never grows dst, and panics before writing unless 0 \<= at \<= Bytes.length(dst) and the range fits.

<!-- end generated -->
