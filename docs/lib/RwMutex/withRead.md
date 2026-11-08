<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# RwMutex.withRead

```ant signature
func RwMutex.withRead[T](m: RwMutex, f: func(): T): T
```

RwMutex.withRead read-locks m, runs f, releases the lock, and returns f's result. It releases the
lock even if f panics.

<!-- end generated -->
