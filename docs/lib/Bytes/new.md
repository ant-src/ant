<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.new

```ant signature
func Bytes.new(size: int = 0): Bytes
```

Bytes.new returns size zero bytes. It panics if size \< 0 or is above 2^31 − 1.

<!-- end generated -->
