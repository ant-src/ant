<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.pipe

```ant signature
func File.pipe(src: File.Handle, dst: File.Handle, ctx: Context | null = null): Result[void, Error]
```

File.pipe copies the rest of src to dst, writing each chunk as it is read, so a source that waits
streams. What it read has been written even when ctx ends the copy, which is then Err with ctx's
reason.

<!-- end generated -->
