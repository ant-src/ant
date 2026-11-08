<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Atomic.compareAndSwap

```ant signature
func Atomic.compareAndSwap[T](a: Atomic[T], expected: T, v: T): bool
```

Atomic.compareAndSwap sets the value in a to v if it equals expected, and reports whether it did. It
compares with ==.

<!-- end generated -->
