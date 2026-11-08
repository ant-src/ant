<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Parallel.map

```ant signature
func Parallel.map[T, U](xs: Array[T], f: func(T, Context): U, limit: int, ctx: Context | null = null): Array[U] | null
```

Parallel.map returns f applied to every element of xs and the scope context, in input order, running
at most limit callbacks at once. It admits elements in index order until the scope ends, and returns
null when some element was never admitted; otherwise it returns every result, even when ctx ended
after the last admission. It panics if limit \< 1, even when xs is empty.

<!-- end generated -->

## Basic examples

```ant
Parallel.map([1, 2, 3], func(n: int, work: Context) { n * n }, limit: 2) // [1, 4, 9]
```
