<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Parallel.race

```ant signature
func Parallel.race[T](fs: Array[func(Context): T], ctx: Context | null = null): T | null
```

Parallel.race calls every function of fs at once with the scope context and returns the first value
one returns, an Err included, cancelling the scope for the rest. It returns null when the scope ends
before any function returns, so a T that includes null is wrapped in a record. It panics if fs is
empty.

<!-- end generated -->
