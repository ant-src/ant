<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# WaitGroup.done

```ant signature
func WaitGroup.done(wg: WaitGroup): void
```

WaitGroup.done decrements the counter of wg by one. It panics if the counter is already zero.

<!-- end generated -->
