<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.filled

```ant signature
func Array.filled[T](n: int, v: T): Array[T]
```

Array.filled returns a new array of n elements, each the same v. Every slot holds v itself, not a
copy. It panics if n \< 0.

<!-- end generated -->
