<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.unwrapOr

```ant signature
func Result.unwrapOr[T, E](r: Result[T, E], fallback: T): T
```

Result.unwrapOr returns the value of r, or fallback if r is an Err.

<!-- end generated -->
