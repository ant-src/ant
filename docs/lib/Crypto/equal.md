<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Crypto.equal

```ant signature
func Crypto.equal(a: Bytes, b: Bytes): bool
```

Crypto.equal reports whether a and b hold the same bytes, in time that depends on their lengths and
not on their contents. It is for checking a received signature against a computed one; sequences of
different lengths are unequal.

<!-- end generated -->
