<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.wrapErr

```ant signature
func Result.wrapErr[T](r: Result[T, Error], context: string): Result[T, Error]
```

Result.wrapErr returns r itself if it is an Ok, and an Err holding Error.wrap(r.error, context) if
it is an Err.

<!-- end generated -->
