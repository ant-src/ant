<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# RwMutex

```ant signature
type RwMutex
```

RwMutex is a reader-writer lock: many goroutines may hold it for reading, or one for writing.
RwMutexes compare by identity.

<!-- end generated -->
