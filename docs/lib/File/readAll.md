<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.readAll

```ant signature
func File.readAll(f: File.Handle, ctx: Context | null = null): Result[string, Error]
```

File.readAll reads the rest of f as text. Bytes that are not valid UTF-8 become U+FFFD.

<!-- end generated -->
