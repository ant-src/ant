<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.ticker

```ant signature
func Time.ticker(d: Time.Duration, ctx: Context | null = null): Channel[Time.Instant]
```

Time.ticker returns a channel that receives the current instant every d, counted from the call. It
holds one tick and drops a tick that finds it full, so a slow reader gets one tick, not a burst. It
stops when ctx is done, or when the run ends if ctx is null, and is never closed. It panics if d is
zero or negative or has years, months or weeks.

<!-- end generated -->
