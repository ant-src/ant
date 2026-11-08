<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Random.int

```ant signature
func Random.int(r: Random, from: int, to: int): int
```

Random.int returns a uniformly random int from from, inclusive, to to, exclusive. It panics if to
\<= from.

<!-- end generated -->
