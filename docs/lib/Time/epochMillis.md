<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.epochMillis

```ant signature
func Time.epochMillis(t: Time.Instant | Time.ZonedDateTime): int
```

Time.epochMillis returns the milliseconds from the Unix epoch to t, rounded down. Every instant
fits.

<!-- end generated -->

## Basic examples

```ant
Time.epochMillis(Time.fromEpochNanos(-1)) // -1
```
