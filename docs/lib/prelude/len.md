<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# len

```ant signature
func len[T](x: T): int
```

len returns the size of x: an array's elements, a map's entries, a set's members, a string's code
points, the bytes of Bytes, or the values a channel buffers now. x may also be a named primitive
over string, which it unwraps.

<!-- end generated -->
