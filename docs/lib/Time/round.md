<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.round

```ant signature
func Time.round[T: Time.Instant | Time.Timed](t: T, smallestUnit: Time.Unit, increment: int = 1, mode: Time.RoundingMode = Time.HalfExpand{}): T
```

Time.round rounds t to a multiple of increment of smallestUnit by mode, by default to the nearest
with a tie away from zero. It rounds the time since the start of the day, or since the epoch for an
Instant; a ZonedDateTime rounded to Day picks the start of its day or the next by how much of the
real day has passed. It panics on Year, Month or Week, on Day for an Instant or a PlainTime, and
unless increment divides the next larger unit (24, 60 or 1000) and is smaller than it, or is 1 for
Day.

<!-- end generated -->

## Basic examples

```ant
Time.round(Time.plainTime(14, 37), Time.Minute{}, increment: 15) // 14:30:00
```
