<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.same

```ant signature
func Array.same[T](xs: Array[T], ys: Array[T]): bool
```

Array.same reports whether xs and ys are the same array, not merely equal ones.

<!-- end generated -->

## Basic examples

```ant
var xs = [1, 2]
Array.same(xs, xs)     // true
Array.same(xs, [1, 2]) // false
```
