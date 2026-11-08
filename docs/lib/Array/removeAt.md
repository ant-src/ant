<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.removeAt

```ant signature
func Array.removeAt[T](xs: Array[T], i: int): T
```

Array.removeAt removes element i from xs and returns it, shifting later elements down. It mutates
xs. It panics if i is out of range.

<!-- end generated -->
