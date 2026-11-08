<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# WaitGroup.wait

```ant signature
func WaitGroup.wait(wg: WaitGroup, ctx: Context | null = null): void
```

WaitGroup.wait blocks until the counter of wg is zero, or until ctx is done. When ctx ends the wait
it returns early; [Context.isDone](../Context/isDone.md) tells the two apart.

<!-- end generated -->
