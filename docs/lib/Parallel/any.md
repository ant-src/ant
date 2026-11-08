<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Parallel.any

```ant signature
func Parallel.any[T, E](fs: Array[func(Context): Result[T, E]], ctx: Context | null = null): Result[T, Array[E]] | null
```

Parallel.any calls every function of fs at once with the scope context and returns the first Ok one
returns, cancelling the scope for the rest. An Err cancels nothing; when every function returns Err,
it returns Err of the errors in input order, and Err([]) when fs is empty. It returns null when the
scope ends before an Ok or the last Err.

<!-- end generated -->
