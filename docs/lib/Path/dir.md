<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.dir

```ant signature
func Path.dir(p: string): string
```

Path.dir returns all but the last element of p, or "." when there is none.

<!-- end generated -->

## Basic examples

```ant
Path.dir("a/b.ant") // "a"
Path.dir("b.ant")   // "."
```
