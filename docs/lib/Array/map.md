<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.map

```ant signature
func Array.map[T, U](xs: Array[T], f: func(T): U): Array[U]
```

Array.map returns a new array holding f applied to each element of xs, in order.

<!-- end generated -->

## Basic examples

```ant
Array.map([1, 2, 3], { x * 2 }) // [2, 4, 6]
```
