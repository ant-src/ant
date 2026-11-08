<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.unwrap

```ant signature
func Result.unwrap[T, E](r: Result[T, E]): T
```

Result.unwrap returns the value of r. It panics if r is an Err.

<!-- end generated -->
