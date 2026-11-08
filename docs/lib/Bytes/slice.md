<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.slice

```ant signature
func Bytes.slice(b: Bytes, from: int, to: int): Bytes
```

Bytes.slice returns the bytes of b from index from up to but excluding index to. It panics if from
\> to, or if either is outside 0 ..= Bytes.length(b).

<!-- end generated -->
