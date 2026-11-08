<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.addDuration

```ant signature
func Time.addDuration(a: Time.Duration, b: Time.Duration): Time.Duration
```

Time.addDuration returns the sum of a and b, with days as 24 hours, balanced up to the larger of
their largest units. It panics if either has years, months or weeks.

<!-- end generated -->

## Basic examples

```ant
Time.addDuration(Time.duration(minutes: 90), Time.duration(minutes: 30)) // PT120M
Time.addDuration(Time.duration(days: 1), Time.duration(hours: 25))       // P2DT1H
```
