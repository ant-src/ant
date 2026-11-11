<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.hoursInDay

```ant signature
func Time.hoursInDay(t: Time.ZonedDateTime): float64
```

Time.hoursInDay returns the length in hours of the date of t in its zone: 24.0 on most days, 23.0 or
25.0 on a daylight-saving change, 23.5 or 24.5 on Lord Howe Island.

<!-- end generated -->

## Details

- Only the date of `t` in its zone matters: every time of that day gives the same result.
- The result is a `float64`: a change of offset by half an hour gives a day such as 23.5 hours long.
- A zone with a fixed offset, such as UTC or `+05:30`, has days of 24.0 hours.

## Basic examples

The day New York moves its clocks forward is 23 hours long:

```ant
let t = Result.unwrap(Time.parseZoned("2026-03-08T12:00[America/New_York]"))
Time.hoursInDay(t) // 23.0
```

The day it moves them back is 25 hours long, and a summer day is 24:

```ant
let fallBack = Result.unwrap(Time.parseZoned("2026-11-01T12:00[America/New_York]"))
let summer = Result.unwrap(Time.parseZoned("2026-06-01T12:00[America/New_York]"))
Time.hoursInDay(fallBack) // 25.0
Time.hoursInDay(summer)   // 24.0
```

## Scope

Lord Howe Island moves its clocks forward by half an hour, so that day is 23.5 hours long:

```ant
Result.unwrap(Time.parseZoned("2026-10-04T12:00[Australia/Lord_Howe]")).hoursInDay() // 23.5
```

The first and the last minute of a day give the same length:

```ant
Result.unwrap(Time.parseZoned("2026-03-08T00:00[America/New_York]")).hoursInDay() // 23.0
Result.unwrap(Time.parseZoned("2026-03-08T23:59[America/New_York]")).hoursInDay() // 23.0
```

A zone with a fixed offset never changes its clocks:

```ant
Result.unwrap(Time.parseZoned("2026-03-08T12:00[UTC]")).hoursInDay()          // 24.0
Result.unwrap(Time.parseZoned("2026-03-08T12:00+05:30[+05:30]")).hoursInDay() // 24.0
```

One instant falls on days of different lengths in different zones:

```ant
let at = Result.unwrap(Time.parseInstant("2026-03-08T12:00:00Z"))
at.inZone(Time.utc).hoursInDay()                                     // 24.0
at.inZone(Result.unwrap(Time.zone("America/New_York"))).hoursInDay() // 23.0
```

In a pipeline, the zoned time is the first argument:

```ant
let text = "2026-10-25T12:00[Europe/London]"
text | Time.parseZoned() | Result.map(Time.hoursInDay) // Ok{ value: 25.0 }
```

## Applications

List the days of a year on which the clocks change:

```ant
let zone = Result.unwrap(Time.zone("Europe/Berlin"))
var day = Time.plainDate(2026, 1, 1)
while day.year() == 2026 {
  let t = Time.toZoned(Time.plainDateTime(day, Time.plainTime(hour: 12)), zone)
  let hours = t.hoursInDay()
  if hours != 24.0 {
    println("${day}: ${hours} hours")
  }
  day = day.add(Time.duration(days: 1))
}
```

```text
2026-03-29: 23.0 hours
2026-10-25: 25.0 hours
```

How much of the day has passed, measured on the real day rather than on a 24-hour clock:

```ant
func dayProgress(t: Time.ZonedDateTime): float64 {
  Time.since(t, t.startOfDay()).total(Time.Hour{}) / t.hoursInDay()
}

let t = Result.unwrap(Time.parseZoned("2026-11-01T12:00[America/New_York]"))
println(dayProgress(t))
```

```text
0.52
```

At noon on the 25-hour day, 13 hours have passed, not 12.

Size an hourly chart of a day, one bar per hour:

```ant
let t = Result.unwrap(Time.parseZoned("2026-03-08T09:30[America/New_York]"))
let bars = int(Math.ceil(t.hoursInDay()))
println("${bars} bars")
```

```text
23 bars
```

## Properties and relations

The length of the day is the time from its start to the start of the next day:

```ant
let zone = Result.unwrap(Time.zone("America/New_York"))
for date in [Time.plainDate(2026, 3, 8), Time.plainDate(2026, 6, 1), Time.plainDate(2026, 11, 1)] {
  let start = Time.toZoned(Time.plainDateTime(date, Time.plainTime()), zone)
  let next = Time.add(start, Time.duration(days: 1)).startOfDay()
  assert(Time.until(start, next).total(Time.Hour{}) == start.hoursInDay())
}
```

A day of a zone with a fixed offset is always 24 hours:

```ant
let zone = Result.unwrap(Time.zone("+05:30"))
let at = Result.unwrap(Time.parseInstant("2026-03-08T12:00:00Z"))
assert(at.inZone(zone).hoursInDay() == 24.0)
```

## Possible issues

The result is a `float64`, and Ant does not compare it with an `int`:

```ant
let t = Result.unwrap(Time.parseZoned("2026-03-08T12:00[America/New_York]"))
println(t.hoursInDay() == 24)
// error: float64 and int cannot be compared with ==: neither is assignable to the other
```

A `Time.PlainDate` has no zone, so it has no daylight-saving changes to count:

```ant
println(Time.hoursInDay(Time.plainDate(2026, 3, 8)))
// error: cannot use Time.PlainDate where Time.ZonedDateTime is expected
```

Place the date in a zone first:

```ant
let zone = Result.unwrap(Time.zone("America/New_York"))
let t = Time.toZoned(Time.plainDateTime(Time.plainDate(2026, 3, 8), Time.plainTime()), zone)
t.hoursInDay() // 23.0
```

## See also

[Time.startOfDay](startOfDay.md) · [Time.nextTransition](nextTransition.md) ·
[Time.previousTransition](previousTransition.md) · [Time.offset](offset.md) ·
[Time.offsetSeconds](offsetSeconds.md) · [Time.toZoned](toZoned.md) · [Time.until](until.md) ·
[Time.ZonedDateTime](types/ZonedDateTime.md)
