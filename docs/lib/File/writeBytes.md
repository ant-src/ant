<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.writeBytes

```ant signature
func File.writeBytes(path: string, bytes: Bytes, replace: bool = true): Result[void, Error]
```

File.writeBytes writes bytes as they are to the file at path, as [`File.write`](write.md) writes
text.

<!-- end generated -->
