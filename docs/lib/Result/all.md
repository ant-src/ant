<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.all

```ant signature
func Result.all[T, E](rs: Array[Result[T, E]]): Result[Array[T], E]
```

Result.all returns Ok of every value in rs, in order, or the first Err in rs.

<!-- end generated -->

## Basic examples

```ant
let rs: Array[Result[int, string]] = [Result.ok(1), Result.ok(2)]
Result.all(rs)                                   // Ok{ value: [1, 2] }
Result.all(Array.concat(rs, [Result.err("no")])) // Err{ error: "no" }
```
