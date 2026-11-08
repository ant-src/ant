<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Seq

```ant signature
type Seq[T] = func(func(T): bool): void
```

Seq is a sequence: a function that pushes each element to yield, in order, until yield returns false
or the elements run out. A for loop iterates one directly.

<!-- end generated -->
