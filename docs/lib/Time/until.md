<!-- status: reviewed -->
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

## Details

- The default largest unit is `Second` for an `Instant`, `Hour` for a `ZonedDateTime` or a
  `PlainTime`, and `Day` for a `PlainDate` or a `PlainDateTime`. Ask for months or years with
  `largestUnit:`.
- Months and years count from `from`: a month has passed when the day of the month of `to` reaches
  that of `from`. The rest is days from that date, constrained to the month's length, so February
  29, 2024 to March 1, 2025 is `P1Y1D`. Swapping the arguments can change more than the sign.
- It panics on a largest unit the type does not have, on `Day` or larger between `ZonedDateTime`s in
  different zones, and when a field of the result does not fit an `int`.

## Basic examples

```ant
Time.until(Time.plainDate(2026, 1, 1), Time.plainDate(2026, 3, 15)) // P73D
Time.until(Time.plainDate(2026, 3, 15), Time.plainDate(2026, 1, 1)) // -P73D
```

Months count from the first date. January 31 to February 28 is no whole month, although January 31
plus a month is February 28:

```ant
let jan31 = Time.plainDate(2026, 1, 31)
Time.until(jan31, Time.plainDate(2026, 2, 28), largestUnit: Time.Month{}) // P28D
Time.until(jan31, Time.plainDate(2026, 3, 31), largestUnit: Time.Month{}) // P2M
```

## Scope

### Every point type

An `Instant` counts in seconds unless you ask for larger units:

```ant
let started = Result.unwrap(Time.parseInstant("2026-10-22T14:03:05Z"))
let finished = Result.unwrap(Time.parseInstant("2026-10-22T16:20:35.25Z"))
started.until(finished) // PT8250.25S
```

A `PlainTime` counts in hours:

```ant
Time.plainTime(9).until(Time.plainTime(17, 30)) // PT8H30M
```

A `PlainDateTime` counts in days and smaller:

```ant
let opened = Time.plainDateTime(Time.plainDate(2026, 1, 1), Time.plainTime(8))
let closed = Time.plainDateTime(Time.plainDate(2026, 3, 15), Time.plainTime(20, 15))
opened.until(closed) // P73DT12H15M
```

A `ZonedDateTime` counts in hours. New York sets its clocks forward on March 8, 2026, so noon to
noon across it is 23 hours, or one of the zone's real days:

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let saturday = Time.plainDateTime(Time.plainDate(2026, 3, 7), Time.plainTime(12)).toZoned(ny)
let sunday = Time.plainDateTime(Time.plainDate(2026, 3, 8), Time.plainTime(12)).toZoned(ny)
saturday.until(sunday)                          // PT23H
saturday.until(sunday, largestUnit: Time.Day{}) // P1D
```

Between two zones, hours and smaller still work:

```ant
let tokyo = Result.unwrap(Time.zone("Asia/Tokyo"))
let t = Time.fromEpochMillis(0)
t.inZone(Time.utc).until(t.add(Time.duration(hours: 30)).inZone(tokyo)) // PT30H
```

### The same point, and pipes

A point until itself is a zero duration:

```ant
let day = Time.plainDate(2026, 5, 4)
day.until(day) // PT0S
```

In a pipe, the piped value is `from`:

```ant
Time.plainDate(2026, 1, 1) | Time.until(Time.plainDate(2026, 12, 25)) // P358D
```

## Named arguments

`largestUnit:` sets the largest field of the result; everything above it is carried down:

```ant
let opened = Time.plainDateTime(Time.plainDate(2026, 1, 1), Time.plainTime(8))
let closed = Time.plainDateTime(Time.plainDate(2026, 3, 15), Time.plainTime(20, 15))
opened.until(closed, largestUnit: Time.Year{})   // P2M14DT12H15M
opened.until(closed, largestUnit: Time.Week{})   // P10W3DT12H15M
opened.until(closed, largestUnit: Time.Minute{}) // PT105855M
```

On an `Instant`, `Time.Hour{}` is the largest unit there is:

```ant
let started = Result.unwrap(Time.parseInstant("2026-10-22T14:03:05Z"))
let finished = Result.unwrap(Time.parseInstant("2026-10-22T16:20:35.25Z"))
started.until(finished, largestUnit: Time.Hour{}) // PT2H17M30.25S
```

February 28, 2025 does not reach day 29, so no year has passed. On March 1 a year has passed, and
the one day left counts from February 28, the constrained date:

```ant
let leap = Time.plainDate(2024, 2, 29)
leap.until(Time.plainDate(2025, 2, 28), largestUnit: Time.Year{}) // P11M30D
leap.until(Time.plainDate(2025, 3, 1), largestUnit: Time.Year{})  // P1Y1D
```

All three arguments can be named:

```ant
let start = Time.plainDate(2026, 1, 1)
let end = Time.plainDate(2026, 12, 25)
Time.until(from: start, to: end, largestUnit: Time.Month{}) // P11M24D
```

## Applications

Days left until a deadline:

```ant
func daysLeft(today: Time.PlainDate, deadline: Time.PlainDate): int {
  today.until(deadline).part(Time.Day{})
}

daysLeft(Time.plainDate(2026, 11, 11), Time.plainDate(2026, 12, 31)) // 50
daysLeft(Time.plainDate(2027, 1, 2), Time.plainDate(2026, 12, 31))   // -2
```

How many whole months a subscription has run, for billing:

```ant
func monthsBilled(start: Time.PlainDate, today: Time.PlainDate): int {
  start.until(today, largestUnit: Time.Month{}).part(Time.Month{})
}

let start = Time.plainDate(2026, 1, 15)
monthsBilled(start, Time.plainDate(2026, 4, 14)) // 2
monthsBilled(start, Time.plainDate(2026, 4, 15)) // 3
```

Hours worked per shift, from clock-in and clock-out times:

```ant
type Shift = { start: Time.PlainTime, end: Time.PlainTime }

let shifts = [
  Shift{ start: Time.plainTime(9), end: Time.plainTime(17, 30) },
  Shift{ start: Time.plainTime(8, 15), end: Time.plainTime(12) },
]
for shift in shifts {
  println(shift.start.until(shift.end).total(Time.Hour{}))
}
```

```text
8.5
3.75
```

## Properties and relations

For exact units, adding the result to `from` gives `to`, and swapping the arguments negates it:

```ant
let a = Result.unwrap(Time.parseInstant("2026-10-22T14:03:05Z"))
let b = Result.unwrap(Time.parseInstant("2026-10-23T09:00:00Z"))
assert(a.add(a.until(b)) == b)
assert(b.until(a) == a.until(b).negate())
assert(a.until(b) == b.since(a))
```

With calendar units the anchor matters, so swapping the arguments can give a different length:

```ant
let feb29 = Time.plainDate(2024, 2, 29)
let mar31 = Time.plainDate(2024, 3, 31)
feb29.until(mar31, largestUnit: Time.Month{}) // P1M2D
mar31.until(feb29, largestUnit: Time.Month{}) // -P1M
```

## Possible issues

A `PlainTime` does not wrap around midnight. From 23:00 to 01:00 is 22 hours back, not 2 hours
forward:

```ant
Time.plainTime(23).until(Time.plainTime(1)) // -PT22H
```

A largest unit the type does not have panics:

```ant
let a = Time.fromEpochMillis(0)
let b = Time.fromEpochMillis(90_000_000)
a.until(b, largestUnit: Time.Day{}) // panic: Time.until: an instant has no days
Time.plainDate(2026, 1, 1).until(Time.plainDate(2026, 1, 2), largestUnit: Time.Hour{})
// panic: Time.until: a date has no hours
Time.plainTime(9).until(Time.plainTime(10), largestUnit: Time.Day{})
// panic: Time.until: a time has no days
```

Days between two zones do not line up, so `Day` or larger panics:

```ant
let tokyo = Result.unwrap(Time.zone("Asia/Tokyo"))
let t = Time.fromEpochMillis(0)
t.inZone(Time.utc).until(t.inZone(tokyo), largestUnit: Time.Day{})
// panic: Time.until: UTC and Asia/Tokyo are different zones; a difference in days or larger needs one zone
```

A long span in a small unit can overflow an `int`. Three centuries in nanoseconds panics:

```ant
let a = Time.fromEpochMillis(0)
let b = Result.unwrap(Time.parseInstant("2300-01-01T00:00:00Z"))
a.until(b, largestUnit: Time.Nanosecond{})
// panic: Time.until: a balanced field does not fit an int
```

Both points must have the same type:

```ant
let start = Time.fromEpochMillis(0)
let end = Time.fromEpochMillis(1000)
let d = start.until(end.inZone(Time.utc)) // error: cannot use Time.ZonedDateTime where Time.Instant is expected
```

## See also

[Time.since](since.md), [Time.add](add.md), [Time.negate](negate.md), [Time.total](total.md),
[Time.part](part.md), [Time.roundDuration](roundDuration.md), [Time.compare](compare.md),
[Time.Unit](types/Unit.md)
