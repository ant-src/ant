<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.map

```ant signature
func Result.map[T, U, E](r: Result[T, E], f: func(T): U): Result[U, E]
```

Result.map returns r with f applied to its value if r is an Ok, and r's error unchanged if it is an
Err.

<!-- end generated -->
