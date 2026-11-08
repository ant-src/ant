<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Url.parse

```ant signature
func Url.parse(s: string): Result[Url.Parts, Error]
```

Url.parse divides s, an absolute or a relative reference, into its parts. It is Err on a malformed
reference, such as a bad %XX, escapes in the path or fragment that are not UTF-8, or a space in the
host.

<!-- end generated -->
