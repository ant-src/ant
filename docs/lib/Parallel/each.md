<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Parallel.each

```ant signature
func Parallel.each[T, R](xs: Array[T] | Channel[T], f: func(T, Context): R, limit: int, ctx: Context | null = null): void
```

Parallel.each calls f on every element of xs, or on every value received from xs until it is closed,
with the scope context, running at most limit callbacks at once, and discards f's value. Over a
channel it receives only when a slot is free and under the scope context, so a value it receives
always runs; it neither closes nor drains the channel, and a null value is passed to f like any
other. It reports nothing: a done ctx after the call does not say which elements ran. It panics if
limit < 1, even when xs is empty.

<!-- end generated -->
