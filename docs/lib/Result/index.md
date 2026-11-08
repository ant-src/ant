<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result

```ant signature
type Result[T, E] = Ok[T] | Err[E]
```

Result is the outcome of an operation that can fail: an Ok with a value or an Err with an error.
Test it with is, or use the Result functions.

```ant signature
type Ok[T] = { value: T }
```

Ok is the success case of a Result, holding its value.

```ant signature
type Err[E] = { error: E }
```

Err is the failure case of a Result, holding its error.

<!-- end generated -->
