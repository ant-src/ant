<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.read

```ant signature
func File.read(path: string): Result[string, Error]
```

File.read reads the whole file at path as text. Bytes that are not valid UTF-8 become U+FFFD;
[File.readBytes](readBytes.md) keeps them as they are.

<!-- end generated -->
