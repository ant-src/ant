<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Net.dial

```ant signature
func Net.dial(addr: string, ctx: Context | null = null): Result[File.Handle, Error]
```

Net.dial opens a TCP connection to addr, written "host:port". The connection is a handle read and
written with the File functions.

<!-- end generated -->
