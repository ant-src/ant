<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Iter.take

```ant signature
func Iter.take[T](s: Seq[T], n: int): Seq[T]
```

Iter.take returns a sequence of at most the first n elements of s. It panics if n \< 0.

<!-- end generated -->
