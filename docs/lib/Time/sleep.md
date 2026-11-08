<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.sleep

```ant signature
func Time.sleep(d: Time.Duration, ctx: Context | null = null): void
```

Time.sleep blocks the calling goroutine for d, or until ctx is done when ctx is not null. Hours and
smaller are exact and a day is 24 hours; a zero or negative d returns at once, and a span beyond
about 292 years is taken as that. It panics if d has years, months or weeks.

<!-- end generated -->
