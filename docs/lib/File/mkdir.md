<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.mkdir

```ant signature
func File.mkdir(path: string): Result[void, Error]
```

File.mkdir creates the directory at path, with any parents it lacks. It is Ok when the directory
already exists.

<!-- end generated -->
