<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Channel.receive

```ant signature
func Channel.receive[T](c: Channel[T], ctx: Context | null = null): T | null
```

Channel.receive takes the next value from c, blocking until one is available. It returns null when c
is closed and drained, or when ctx is done first; [`Context.isDone`](../Context/isDone.md) tells the
two apart. When T includes null, a null value looks like the end, so wrap such values.

<!-- end generated -->
