<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# RwMutex.writeUnlock

```ant signature
func RwMutex.writeUnlock(m: RwMutex): void
```

RwMutex.writeUnlock releases the write lock on m. It panics if m is not write-locked.

<!-- end generated -->
