<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Atomic

```ant signature
type Atomic[T]
```

Atomic is a cell holding one value of T, read and written atomically. All operations on all Atomic
cells form one total order. Atomics compare by identity.

<!-- end generated -->
