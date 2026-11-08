<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# RwMutex.writeLock

```ant signature
func RwMutex.writeLock(m: RwMutex): void
```

RwMutex.writeLock locks m for writing, blocking while any reader or writer holds it.

<!-- end generated -->
