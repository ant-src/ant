<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.groupBy

```ant signature
func Array.groupBy[T, K](xs: Array[T], key: func(T): K): Map[K, Array[T]]
```

Array.groupBy returns a map from each key to the elements of xs with that key. Keys are in
first-seen order and each group keeps the order of xs.

<!-- end generated -->

## Basic examples

```ant
Array.groupBy([1, 2, 3, 4, 5], { x % 2 }) // Map{ 1: [1, 3, 5], 0: [2, 4] }
```
