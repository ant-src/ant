<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.write

```ant signature
func File.write(path: string, text: string, replace: bool = true): Result[void, Error]
```

File.write writes text as UTF-8 to the file at path, creating the file or replacing what it held; a
file that exists keeps its mode. With replace false it creates the file only when nothing is at
path, and is otherwise Err, having changed nothing.

<!-- end generated -->
