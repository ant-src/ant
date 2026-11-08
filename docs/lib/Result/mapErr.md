<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.mapErr

```ant signature
func Result.mapErr[T, E, F](r: Result[T, E], f: func(E): F): Result[T, F]
```

Result.mapErr returns r with f applied to its error if r is an Err, and r's value unchanged if it is
an Ok. With an identity function, as in Result.mapErr(r, \{ x \}), it widens the error type.

<!-- end generated -->
