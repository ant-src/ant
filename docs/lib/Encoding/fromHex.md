<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Encoding.fromHex

```ant signature
func Encoding.fromHex(s: string): Result[Bytes, Error]
```

Encoding.fromHex decodes s, hex digits of either case, into bytes. It is Err on an odd length or a
character that is not a hex digit.

<!-- end generated -->

## Basic examples

```ant
Encoding.fromHex("6869") // Ok{ value: Bytes(68 69) }
```
