<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.fromBytesLossy

```ant signature
func String.fromBytesLossy(bytes: Bytes): string
```

String.fromBytesLossy decodes bytes as UTF-8, replacing each maximal subpart of an ill-formed
sequence with U+FFFD.

<!-- end generated -->
