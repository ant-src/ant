<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# RwMutex.readUnlock

```ant signature
func RwMutex.readUnlock(m: RwMutex): void
```

RwMutex.readUnlock releases one read lock on m. It panics if m is not read-locked.

<!-- end generated -->
