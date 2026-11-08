<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Crypto.hmacSha256

```ant signature
func Crypto.hmacSha256(key: Bytes, message: Bytes): Bytes
```

Crypto.hmacSha256 returns the 32-byte HMAC-SHA-256 of message under key. Compare it with a received
signature using [Crypto.equal](equal.md).

<!-- end generated -->
