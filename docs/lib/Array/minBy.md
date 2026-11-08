<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.minBy

```ant signature
func Array.minBy[T, K: Ordered](xs: Array[T], key: func(T): K): T | null
```

Array.minBy returns the first element of xs with the least key, or null when xs is empty. An element
replaces the one held only when its key is strictly less, so of equal keys the first wins and a NaN
key never wins.

<!-- end generated -->

## Basic examples

```ant
Array.minBy(["bb", "a", "c"], { String.length(x) }) // "a"
```
