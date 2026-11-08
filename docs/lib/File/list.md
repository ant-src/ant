<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.list

```ant signature
func File.list(dir: string): Result[Array[string], Error]
```

File.list returns the names of the entries directly in dir, in lexical order. Unlike
[`File.walk`](walk.md) it does not descend, and it returns names rather than paths.

<!-- end generated -->
