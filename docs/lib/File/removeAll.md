<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.removeAll

```ant signature
func File.removeAll(path: string): Result[void, Error]
```

File.removeAll removes path and everything below it. It is Ok when there is nothing to remove;
[File.remove](remove.md) is the form that refuses a directory with entries.

<!-- end generated -->
