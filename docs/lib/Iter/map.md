<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Iter.map

```ant signature
func Iter.map[T, U](s: Seq[T], f: func(T): U): Seq[U]
```

Iter.map returns a sequence of f applied to each element of s. Like every Iter function but
[`Iter.toArray`](toArray.md), it runs nothing until the result is run.

<!-- end generated -->
