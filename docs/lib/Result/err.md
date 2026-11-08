<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.err

```ant signature
func Result.err[T, E](error: E): Result[T, E]
```

Result.err returns an Err carrying error; Result.err(e) is the same value as Err\{ error: e \}.

<!-- end generated -->
