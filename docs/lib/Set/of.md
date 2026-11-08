<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set.of

```ant signature
func Set.of[T](xs: Array[T]): Set[T]
```

Set.of returns a new set holding the first of each == value in xs, in order. Set.of([]) is the empty
set, its element type coming from the expected type.

<!-- end generated -->
