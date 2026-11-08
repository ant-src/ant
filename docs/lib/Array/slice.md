<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.slice

```ant signature
func Array.slice[T](xs: Array[T], from: int, to: int): Array[T]
```

Array.slice returns a new array of the elements of xs from index from up to but excluding to. It
panics if from > to, or if either is outside 0 ..= length.

<!-- end generated -->

## Basic examples

```ant
Array.slice([10, 20, 30, 40], 1, 3) // [20, 30]
```
