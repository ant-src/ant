<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Random.fork

```ant signature
func Random.fork(r: Random): Random
```

Random.fork returns a generator with r's state, and moves r 2^128 draws ahead, so neither sequence
reaches the other's in any run. The new generator goes on with the values r would have given.

<!-- end generated -->
