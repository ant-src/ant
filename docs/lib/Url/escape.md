<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Url.escape

```ant signature
func Url.escape(s: string): string
```

Url.escape escapes s for use as one query key or value: every character but the unreserved ones
becomes %XX, and a space becomes +.

<!-- end generated -->

## Basic examples

```ant
Url.escape("a b&c") // "a+b%26c"
```
