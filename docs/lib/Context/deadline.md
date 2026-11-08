<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Context.deadline

```ant signature
func Context.deadline(ctx: Context): Time.Instant | null
```

Context.deadline returns the earliest deadline of ctx and its parents, or null if none has one.

<!-- end generated -->
