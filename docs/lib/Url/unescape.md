<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Url.unescape

```ant signature
func Url.unescape(s: string): Result[string, Error]
```

Url.unescape reverses [Url.escape](escape.md). It is Err on a bad %XX or on escapes that decode to
bytes that are not UTF-8.

<!-- end generated -->

## Basic examples

```ant
Url.unescape("a+b%26c") // Ok{ value: "a b&c" }
```
