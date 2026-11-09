<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.total

```ant signature
func Time.total(d: Time.Duration, unit: Time.Unit, relativeTo: Time.Dated | null = null): float64
```

Time.total returns the length of d in unit, with a fraction: PT1H30M in hours is 1.5. Years, months
and weeks are measured from relativeTo, a PlainDateTime counting as its date, and a day is 24 hours
unless relativeTo is a ZonedDateTime. It panics on years, months or weeks, in d or as unit, when
relativeTo is null.

<!-- end generated -->

## Basic examples

```ant
let feb = Time.plainDate(2026, 2, 1)
Time.total(Time.duration(months: 1), Time.Day{}, relativeTo: feb) // 28.0
```
