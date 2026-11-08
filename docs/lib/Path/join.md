<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.join

```ant signature
func Path.join(parts: Array[string]): string
```

Path.join joins parts with the path separator and cleans the result, so no parts give ".".

<!-- end generated -->

## Basic examples

```ant
Path.join(["a", "b/../c", "d.ant"]) // "a/c/d.ant"
```
