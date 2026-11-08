<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Crypto.randomBytes

```ant signature
func Crypto.randomBytes(n: int): Bytes
```

Crypto.randomBytes returns n bytes from the operating system's secure source of randomness. It
panics if n \< 0 or n is above 2^31 - 1.

<!-- end generated -->
