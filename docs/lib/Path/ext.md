<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.ext

```ant signature
func Path.ext(p: string): string
```

Path.ext returns the extension of p's last element, from its last dot, or "" when it has none.

<!-- end generated -->

## Basic examples

```ant
Path.ext("a/b.tar.gz") // ".gz"
Path.ext("a/b")        // ""
```
