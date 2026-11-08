<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.add

```ant signature
func Time.add[T: Time.Point](t: T, d: Time.Duration): T
```

Time.add returns t moved by d. The calendar part goes first, years then months with the day
constrained, then weeks and days, and hours and smaller after; on a ZonedDateTime the calendar
result is placed in the zone by Time.Compatible before the exact part moves the instant. A PlainDate
takes the exact part as whole 24-hour days and a PlainTime ignores the calendar part and wraps
around the clock. It panics if t is an Instant and d has years, months, weeks or days.

<!-- end generated -->

## Basic examples

```ant
Time.plainDate(2024, 1, 31) | Time.add(Time.duration(months: 1)) // 2024-02-29
Time.plainTime(23, 30) | Time.add(Time.duration(hours: 2))       // 01:30:00
```
