<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.randomInt

```ant signature
func Math.randomInt(from: int, to: int): int
```

Math.randomInt returns a uniformly random int from from, inclusive, to to, exclusive. It draws from
the same unseeded sequence as [Math.random](random.md). It panics if to <= from.

<!-- end generated -->
