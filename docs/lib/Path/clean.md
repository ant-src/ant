<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.clean

```ant signature
func Path.clean(p: string): string
```

Path.clean returns the shortest path equivalent to p, working on the text alone.

<!-- end generated -->

## Basic examples

```ant
Path.clean("a//b/./c/..") // "a/b"
```
