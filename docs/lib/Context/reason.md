<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Context.reason

```ant signature
func Context.reason(ctx: Context): string | null
```

Context.reason returns why ctx is done, or null while it is not. The reason is "cancelled",
"timeout", "interrupt", or the parent's reason.

<!-- end generated -->
