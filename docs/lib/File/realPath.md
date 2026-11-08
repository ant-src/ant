<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.realPath

```ant signature
func File.realPath(path: string): Result[string, Error]
```

File.realPath returns path made absolute and clean, with every symbolic link along it resolved. It
is Err when the path leads nowhere.

<!-- end generated -->
