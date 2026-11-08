<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Encoding.toBase64

```ant signature
func Encoding.toBase64(bytes: Bytes, url: bool = false): string
```

Encoding.toBase64 encodes bytes as padded base64 in the standard alphabet, with + and /, or when url
is true in the URL-safe one, with - and \_.

<!-- end generated -->

## Basic examples

```ant
Encoding.toBase64(String.toBytes("hi?"))            // "aGk/"
Encoding.toBase64(String.toBytes("hi?"), url: true) // "aGk_"
```
