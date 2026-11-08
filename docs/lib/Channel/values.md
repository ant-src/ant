<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Channel.values

```ant signature
func Channel.values[T](c: Channel[T], ctx: Context | null = null): Seq[T]
```

Channel.values returns a sequence of every value received from c until it is closed and drained, or
until ctx is done.

<!-- end generated -->
