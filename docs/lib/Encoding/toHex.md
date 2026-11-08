<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Encoding.toHex

```ant signature
func Encoding.toHex(bytes: Bytes): string
```

Encoding.toHex encodes bytes as two lower-case hex digits per byte.

<!-- end generated -->

## Basic examples

```ant
Encoding.toHex(String.toBytes("hi")) // "6869"
```
