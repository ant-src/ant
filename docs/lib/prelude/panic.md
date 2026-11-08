<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# panic

```ant signature
func panic(message: string): never
```

panic stops the current goroutine with message, running its deferred blocks on the way out, unless a
[Panic.catch](../Panic/catch.md) on the way stops it. A panic in the main goroutine ends the program
with exit status 1.

<!-- end generated -->
