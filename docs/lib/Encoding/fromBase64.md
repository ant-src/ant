<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Encoding.fromBase64

```ant signature
func Encoding.fromBase64(s: string, url: bool = false): Result[Bytes, Error]
```

Encoding.fromBase64 decodes s as base64 in the alphabet url names, with the padding present or
absent. It ignores \\r and \\n, and is Err on any other character outside the alphabet or on a
length no encoding produces.

<!-- end generated -->
