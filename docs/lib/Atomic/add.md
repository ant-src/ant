<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Atomic.add

```ant signature
func Atomic.add[T: Number](a: Atomic[T], d: T): T
```

Atomic.add adds d to the value in a and returns the new value. It panics on overflow.

<!-- end generated -->
