<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Net.accept

```ant signature
func Net.accept(l: Net.Listener, ctx: Context | null = null): Result[File.Handle, Error]
```

Net.accept waits for the next connection on l and returns it as a handle. It is Err when l is
closed, or with ctx's reason when ctx is done. Several goroutines may accept on one listener, and
each connection goes to one of them.

<!-- end generated -->
