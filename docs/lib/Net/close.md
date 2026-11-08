<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Net.close

```ant signature
func Net.close(l: Net.Listener): void
```

Net.close closes l, so pending and later [Net.accept](accept.md) calls are Err. Closing it again
does nothing.

<!-- end generated -->
