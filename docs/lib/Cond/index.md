<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cond

```ant signature
type Cond
```

Cond is a condition variable: goroutines wait on it, under its mutex, until another signals. Conds
compare by identity.

<!-- end generated -->
