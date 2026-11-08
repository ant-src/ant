<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Iter.filter

```ant signature
func Iter.filter[T](s: Seq[T], f: func(T): bool): Seq[T]
```

Iter.filter returns a sequence of the elements of s that f accepts.

<!-- end generated -->

## Basic examples

```ant
Iter.range(0, 30) | Iter.filter({ x % 7 == 0 }) | Iter.toArray() // [0, 7, 14, 21, 28]
```
