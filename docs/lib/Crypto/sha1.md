<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Crypto.sha1

```ant signature
func Crypto.sha1(bytes: Bytes): Bytes
```

Crypto.sha1 returns the 20-byte SHA-1 digest of bytes. SHA-1 is broken against collisions and is
here for the formats that still use it, as git and TOTP do.

<!-- end generated -->
