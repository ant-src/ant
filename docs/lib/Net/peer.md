<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Net.peer

```ant signature
func Net.peer(f: File.Handle): string | null
```

Net.peer returns the address of the other end of the connection f, as "ip:port". It is null when f
is a file or a standard stream.

<!-- end generated -->
