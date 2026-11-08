<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.reduce

```ant signature
func Array.reduce[T, A](xs: Array[T], init: A, f: func(A, T): A): A
```

Array.reduce folds xs from the left, starting from init and calling f with the accumulator and each
element in turn. It returns init when xs is empty.

<!-- end generated -->
