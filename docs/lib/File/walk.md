<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.walk

```ant signature
func File.walk(dir: string): Result[Array[string], Error]
```

File.walk returns every file below dir, at any depth, as paths relative to dir. It lists files and
not directories, does not follow symbolic links, and returns the paths in lexical order.

<!-- end generated -->
