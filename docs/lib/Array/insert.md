<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.insert

```ant signature
func Array.insert[T](xs: Array[T], i: int, v: T): void
```

Array.insert inserts v into xs so that it becomes element i, shifting later elements up. It mutates
xs. It panics if i is outside 0 ..= length, so i equal to the length appends.

<!-- end generated -->
