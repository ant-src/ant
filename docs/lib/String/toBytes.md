<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.toBytes

```ant signature
func String.toBytes(s: string): Bytes
```

String.toBytes returns the UTF-8 encoding of s.

<!-- end generated -->

## Basic examples

```ant
String.toBytes("é") // Bytes(c3 a9)
```
