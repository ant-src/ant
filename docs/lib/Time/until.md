<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.until

```ant signature
func Time.until[T: Time.Point](from: T, to: T, largestUnit: Time.Unit | null = null): Time.Duration
```

Time.until returns the duration from from to to, negative when to is earlier, balanced up to
largestUnit. That defaults to Time.Second for an Instant, Time.Hour for a ZonedDateTime or a
PlainTime, and Time.Day otherwise; months count from from, one passing when the day of the month
reaches that of from, and days between ZonedDateTimes in one zone are the zone's real days. It
panics on a unit above Hour for an Instant or a PlainTime or below Day for a PlainDate, and on Day
or larger between ZonedDateTimes in different zones.

<!-- end generated -->

## Basic examples

```ant
var jan31 = Time.plainDate(2026, 1, 31)
Time.until(jan31, Time.plainDate(2026, 2, 28), largestUnit: Time.Month{}) // P28D
```
