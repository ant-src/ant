<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.lock

```ant signature
func File.lock(path: string, ctx: Context | null = null): Result[File.Handle, Error]
```

File.lock opens the file at path for reading and writing, creating it when it is missing and never
truncating it, and waits for an exclusive lock on it. It is Err with ctx's reason when ctx is done
first. The lock is released when the handle is closed or the process ends, however it ends; a child
process starting at that moment can delay the release until its program has started.

<!-- end generated -->
