<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Context.new

```ant signature
func Context.new(timeout: Time.Duration | null = null, parent: Context | null = null): Context
```

Context.new returns a new context that is a child of parent, or of the run when parent is null. With
a timeout it is done once the timeout passes; a child of a done parent is done at once with the
parent's reason. It panics if timeout has years, months, or weeks, or is not positive.

<!-- end generated -->
