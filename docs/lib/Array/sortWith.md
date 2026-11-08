<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.sortWith

```ant signature
func Array.sortWith[T](xs: Array[T], cmp: func(T, T): int): Array[T]
```

Array.sortWith returns a new array holding the elements of xs ordered by cmp. cmp(a, b) < 0 puts a
before b, and the sort is stable.

<!-- end generated -->

## Basic examples

```ant
Array.sortWith([3, 1, 2], func(a: int, b: int) { b - a }) // [3, 2, 1]
```
