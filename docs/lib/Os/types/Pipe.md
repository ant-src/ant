<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.Pipe

```ant signature
type Os.Pipe = { read: File.Handle, write: File.Handle }
```

Os.Pipe is a pair of connected handles, made by [`Os.openPipe`](../openPipe.md): what is written to
write is read from read.

<!-- end generated -->
