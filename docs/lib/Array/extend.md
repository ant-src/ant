<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.extend

```ant signature
func Array.extend[T](xs: Array[T], ys: Array[T]): void
```

Array.extend appends the elements of ys to the end of xs, in order, as ys was at the call, so
Array.extend(xs, xs) doubles xs. It mutates xs.

<!-- end generated -->
