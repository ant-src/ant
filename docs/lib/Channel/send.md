<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Channel.send

```ant signature
func Channel.send[T](c: Channel[T], v: T, ctx: Context | null = null): bool
```

Channel.send sends v on c, blocking until a receiver takes it or the buffer has room. It returns
true once v is taken or buffered, and false if ctx is done first. It panics if c is closed.

<!-- end generated -->
