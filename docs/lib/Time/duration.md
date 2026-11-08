<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.duration

```ant signature
func Time.duration(years: int = 0, months: int = 0, weeks: int = 0, days: int = 0, hours: int = 0, minutes: int = 0, seconds: int = 0, milliseconds: int = 0, microseconds: int = 0, nanoseconds: int = 0): Time.Duration
```

Time.duration returns a duration with the given fields, each 0 when omitted and kept as given. It
panics if the fields have mixed signs, if years, months, weeks or days reach 2^32 in magnitude, or
if the time part, days counted as 24 hours, reaches 2^53 seconds.

<!-- end generated -->

## Basic examples

```ant
Time.duration(hours: 1, minutes: 30, milliseconds: 500) // PT1H30M0.5S
```
