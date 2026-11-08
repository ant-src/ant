<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.unique

```ant signature
func Array.unique[T](xs: Array[T]): Array[T]
```

Array.unique returns a new array holding the first of each == value in xs, in order.

<!-- end generated -->

## Basic examples

```ant
Array.unique([3, 1, 3, 2, 1]) // [3, 1, 2]
```
