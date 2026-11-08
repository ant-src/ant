<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.close

```ant signature
func File.close(f: File.Handle): Result[void, Error]
```

File.close closes f, and is Err when the close fails; f is closed either way. A buffered handle
writes its buffer out first and then closes the handle under it, reporting the first error. When a
write of it is in progress on another goroutine, it closes the handle under first, which ends that
write, and is Err when what the buffer held is lost. It does nothing and is Ok on a standard stream
and on a handle already closed, so a deferred close can follow an explicit one. Called from another
goroutine, it ends a read or write in progress on f with an Err.

<!-- end generated -->
