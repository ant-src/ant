<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.compareDuration

```ant signature
func Time.compareDuration(a: Time.Duration, b: Time.Duration, relativeTo: Time.Dated | null = null): int
```

Time.compareDuration returns -1, 0 or 1 as a is shorter than, as long as or longer than b. Years,
months and weeks are measured from relativeTo, and a day is 24 hours unless relativeTo is a
ZonedDateTime, whose days have their daylight-saving hours. It panics if either has years, months or
weeks and relativeTo is null.

<!-- end generated -->

## Basic examples

```ant
let month = Time.duration(months: 1)
let feb = Time.plainDate(2026, 2, 1)
Time.compareDuration(month, Time.duration(days: 30), relativeTo: feb) // -1
```
