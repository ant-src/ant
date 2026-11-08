<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Mutex.withLock

```ant signature
func Mutex.withLock[T](m: Mutex, f: func(): T): T
```

Mutex.withLock locks m, runs f, unlocks m, and returns f's result. It unlocks m even if f panics.

<!-- end generated -->
