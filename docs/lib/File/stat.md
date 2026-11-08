<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.stat

```ant signature
func File.stat(path: string, follow: bool = true): Result[File.Info, Error]
```

File.stat returns the name, size, kind, modification time and mode of the file at path. It follows a
symbolic link and describes its target, unless follow is false, when it describes the link itself.

<!-- end generated -->
