<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.fail

```ant signature
func Result.fail[T](message: string): Result[T, Error]
```

Result.fail returns an Err carrying an Error with the message; Result.fail(m) is the same value as Result.err(Error\{ message: m \}). Its error type is Error itself, so where a wider one is expected, such as ParseError \| Error, write Result.err.

<!-- end generated -->
