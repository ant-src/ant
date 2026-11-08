<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.source

```ant signature
func Regex.source(p: Regex.Pattern): string
```

Regex.source returns the normalized source of p, in which \/ is written /.

<!-- end generated -->

## Basic examples

```ant
Regex.source(/a\/b/) // "a/b"
```
