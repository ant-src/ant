<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.flush

```ant signature
func File.flush(f: File.Handle): Result[void, Error]
```

File.flush writes out the buffer of a buffered handle, and of every buffered handle under it, so
that what was written to f has been handed to the OS. On a handle that is not buffered it does
nothing and is Ok.

<!-- end generated -->
