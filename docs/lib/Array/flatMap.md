<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.flatMap

```ant signature
func Array.flatMap[T, U](xs: Array[T], f: func(T): Array[U]): Array[U]
```

Array.flatMap applies f to each element of xs and returns the arrays it returns, concatenated in
order.

<!-- end generated -->

## Basic examples

```ant
Array.flatMap([1, 2, 3], { [x, x * 10] }) // [1, 10, 2, 20, 3, 30]
```
