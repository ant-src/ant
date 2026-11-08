<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.get

```ant signature
func Array.get[T](xs: Array[T], i: int): T | null
```

Array.get returns element i of xs, or null when i is out of range. Unlike xs\[i\], it never panics.

<!-- end generated -->
