<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.readLine

```ant signature
func File.readLine(f: File.Handle, ctx: Context | null = null): Result[string | null, Error]
```

File.readLine reads the next line from f as text, without its \\n or \\r\\n, and returns null at end
of file. When ctx is done first it is Err with ctx's reason, and what it had read stays for the next
read of f.

<!-- end generated -->
