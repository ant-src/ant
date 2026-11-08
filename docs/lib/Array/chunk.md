<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.chunk

```ant signature
func Array.chunk[T](xs: Array[T], n: int): Array[Array[T]]
```

Array.chunk splits xs into arrays of n elements each, the last holding fewer when the length is not
a multiple of n. It panics if n \< 1. An empty xs gives \[\].

<!-- end generated -->

## Basic examples

```ant
Array.chunk([1, 2, 3, 4, 5], 2) // [[1, 2], [3, 4], [5]]
```
