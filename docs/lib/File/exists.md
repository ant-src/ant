<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.exists

```ant signature
func File.exists(path: string): bool
```

File.exists reports whether anything exists at path, a directory included. It follows a symbolic
link, so a link that leads nowhere is false.

<!-- end generated -->
