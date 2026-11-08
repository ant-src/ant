<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.Handle

```ant signature
type File.Handle
```

File.Handle is an opaque handle to an open file, a standard stream, a network connection, or a
buffer over one ([`File.buffered`](../buffered.md)). The reads of a handle share one buffer, so a
program can read lines and then bytes from it. A write is handed to the OS before it returns, all of
it or Err, except a write to a buffered handle. One goroutine may read a handle while another writes
it.

<!-- end generated -->
