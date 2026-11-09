<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.nextTransition

```ant signature
func Time.nextTransition(t: Time.ZonedDateTime): Time.ZonedDateTime | null
```

Time.nextTransition returns the earliest moment after t at which the zone of t changes its offset,
in that zone. It is null when there is none, as for an offset zone or a zone that has stopped
changing.

<!-- end generated -->

## Basic examples

```ant
let t = Result.unwrap(Time.parseZoned("2024-06-01T00:00[America/New_York]"))
Time.nextTransition(t) // 2024-11-03T01:00:00-05:00[America/New_York]
```
