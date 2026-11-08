<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set.add

```ant signature
func Set.add[T](s: Set[T], x: T): bool
```

Set.add adds x to s and reports whether it was not already a member. It mutates s. Adding an
existing member changes nothing, so it keeps its position.

<!-- end generated -->
