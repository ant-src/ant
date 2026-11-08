<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Context.cancel

```ant signature
func Context.cancel(ctx: Context): void
```

Context.cancel makes ctx done now, with the reason "cancelled". It does nothing if ctx is already
done.

<!-- end generated -->
