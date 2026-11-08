<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Channel.close

```ant signature
func Channel.close[T](c: Channel[T]): void
```

Channel.close closes c: receivers drain what is buffered and then get null. It panics if c is
already closed.

<!-- end generated -->
