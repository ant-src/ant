<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.base

```ant signature
func Path.base(p: string): string
```

Path.base returns the last element of p, or "." when p is empty.

<!-- end generated -->

## Basic examples

```ant
Path.base("a/b.ant") // "b.ant"
```
