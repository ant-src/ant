<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.sync

```ant signature
func File.sync(f: File.Handle): Result[void, Error]
```

File.sync flushes f as [`File.flush`](flush.md) does, then has the OS write the file to storage, so
that what was written survives a crash or a power loss. Sync a temporary file before renaming it
into place. It is Err where the OS refuses, as for a pipe, and for a connection or a closed handle.

<!-- end generated -->
