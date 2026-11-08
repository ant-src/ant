<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.uniqueBy

```ant signature
func Array.uniqueBy[T, K](xs: Array[T], key: func(T): K): Array[T]
```

Array.uniqueBy returns a new array holding the first element of xs for each == key, in order.

<!-- end generated -->
