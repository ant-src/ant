<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# WaitGroup.add

```ant signature
func WaitGroup.add(wg: WaitGroup, n: int): void
```

WaitGroup.add adds n, which may be negative, to the counter of wg. It panics if the counter would go
negative or past [Math.maxInt].

<!-- end generated -->
