<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.epochNanos

```ant signature
func Time.epochNanos(t: Time.Instant | Time.ZonedDateTime): int
```

Time.epochNanos returns the nanoseconds from the Unix epoch to t. It panics if t is outside
1677-09-21 to 2262-04-11, beyond which the count does not fit in an int.

<!-- end generated -->
