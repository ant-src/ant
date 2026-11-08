<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set.union

```ant signature
func Set.union[T](a: Set[T], b: Set[T]): Set[T]
```

Set.union returns a new set holding the members of a, then those of b that a lacks.

<!-- end generated -->

## Basic examples

```ant
Set.union(Set.of([1, 2]), Set.of([3, 2])) // Set.of([1, 2, 3])
```
