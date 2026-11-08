<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.create

```ant signature
func File.create(path: string, replace: bool = true): Result[File.Handle, Error]
```

File.create opens the file at path for writing, creating it or truncating what it held. With replace
false it creates the file only when nothing is at path, and is otherwise Err.

<!-- end generated -->
