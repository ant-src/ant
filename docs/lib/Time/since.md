<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.since

```ant signature
func Time.since[T: Time.Point](t: T, other: T, largestUnit: Time.Unit | null = null): Time.Duration
```

Time.since returns the duration from other to t, counted from t. It is Time.negate(Time.until(t,
other)), and for calendar units can differ from Time.until(other, t), whose anchor is other. It
panics as [`Time.until`](until.md) does.

<!-- end generated -->

## Details

- The result is positive when `t` is later than `other`, the opposite sign of
  `Time.until(t, other)`.
- The default largest unit is `Second` for an `Instant`, `Hour` for a `ZonedDateTime` or a
  `PlainTime`, and `Day` for the others.
- It panics on a largest unit the type does not have, and on `Day` or larger between
  `ZonedDateTime`s in different zones.

## Basic examples

```ant
let feb29 = Time.plainDate(2024, 2, 29)
let mar31 = Time.plainDate(2024, 3, 31)
Time.since(mar31, feb29) // P31D
Time.since(feb29, mar31) // -P31D
```

`Time.since` counts months back from its first argument, so it can differ from `Time.until`:

```ant
let feb29 = Time.plainDate(2024, 2, 29)
let mar31 = Time.plainDate(2024, 3, 31)
Time.since(mar31, feb29, largestUnit: Time.Month{}) // P1M
Time.until(feb29, mar31, largestUnit: Time.Month{}) // P1M2D
```

## Scope

### Every point type

An `Instant` counts in seconds unless `largestUnit:` names a larger unit:

```ant
let started = Result.unwrap(Time.parseInstant("2026-10-22T14:03:05Z"))
let finished = Result.unwrap(Time.parseInstant("2026-10-22T16:20:35.25Z"))
finished.since(started) // PT8250.25S
```

A `PlainTime` counts in hours:

```ant
Time.plainTime(17, 30).since(Time.plainTime(9, 0)) // PT8H30M
Time.plainTime(9, 0).since(Time.plainTime(17, 30)) // -PT8H30M
```

A `PlainDateTime` counts in days and smaller:

```ant
let opened = Time.plainDateTime(Time.plainDate(2026, 1, 1), Time.plainTime(8))
let closed = Time.plainDateTime(Time.plainDate(2026, 3, 15), Time.plainTime(20, 15))
closed.since(opened) // P73DT12H15M
```

A `ZonedDateTime` counts in hours. New York sets its clocks forward on March 8, 2026, so noon to
noon across it is 23 hours, or one of the zone's real days:

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let saturday = Time.plainDateTime(Time.plainDate(2026, 3, 7), Time.plainTime(12)).toZoned(ny)
let sunday = Time.plainDateTime(Time.plainDate(2026, 3, 8), Time.plainTime(12)).toZoned(ny)
sunday.since(saturday)                          // PT23H
sunday.since(saturday, largestUnit: Time.Day{}) // P1D
```

### The same point, and pipes

A point since itself is a zero duration:

```ant
let day = Time.plainDate(2026, 5, 4)
day.since(day) // PT0S
```

In a pipe, the piped value is `t`, the later point of a positive result:

```ant
let opened = Time.plainDateTime(Time.plainDate(2026, 1, 1), Time.plainTime(8))
let closed = Time.plainDateTime(Time.plainDate(2026, 3, 15), Time.plainTime(20, 15))
closed | Time.since(opened) // P73DT12H15M
```

## Named arguments

`largestUnit:` sets the largest field of the result; everything above it is carried down:

```ant
let opened = Time.plainDateTime(Time.plainDate(2026, 1, 1), Time.plainTime(8))
let closed = Time.plainDateTime(Time.plainDate(2026, 3, 15), Time.plainTime(20, 15))
closed.since(opened, largestUnit: Time.Year{}) // P2M14DT12H15M
closed.since(opened, largestUnit: Time.Week{}) // P10W3DT12H15M
closed.since(opened, largestUnit: Time.Hour{}) // PT1764H15M
```

On an `Instant`, `Time.Hour{}` is the largest unit there is:

```ant
let started = Result.unwrap(Time.parseInstant("2026-10-22T14:03:05Z"))
let finished = Result.unwrap(Time.parseInstant("2026-10-22T16:20:35.25Z"))
finished.since(started, largestUnit: Time.Hour{}) // PT2H17M30.25S
```

## Applications

A person's age in whole years, read from the year field of a duration counted in years:

```ant
func age(birthday: Time.PlainDate, today: Time.PlainDate): int {
  today.since(birthday, largestUnit: Time.Year{}).part(Time.Year{})
}

age(Time.plainDate(1990, 7, 15), Time.plainDate(2026, 7, 14)) // 35
age(Time.plainDate(1990, 7, 15), Time.plainDate(2026, 7, 15)) // 36
```

How long a job took, in minutes:

```ant
let started = Result.unwrap(Time.parseInstant("2026-10-22T14:03:05Z"))
let finished = Result.unwrap(Time.parseInstant("2026-10-22T14:10:35Z"))
finished.since(started).total(Time.Minute{}) // 7.5
```

A short label for how long ago something happened:

```ant
func ago(then: Time.Instant, now: Time.Instant): string {
  let d = now.since(then, largestUnit: Time.Hour{})
  let hours = d.part(Time.Hour{})
  let minutes = d.part(Time.Minute{})
  if hours > 0 { "${hours} h ago" } else if minutes > 0 { "${minutes} min ago" } else { "just now" }
}

let now = Result.unwrap(Time.parseInstant("2026-10-22T14:00:00Z"))
for stamp in ["2026-10-22T13:59:30Z", "2026-10-22T13:45:00Z", "2026-10-22T09:10:00Z"] {
  println(ago(Result.unwrap(Time.parseInstant(stamp)), now))
}
```

```text
just now
15 min ago
4 h ago
```

## Properties and relations

`Time.since` negates `Time.until` with the same arguments, and for exact units it equals
`Time.until` with the arguments swapped:

```ant
let a = Result.unwrap(Time.parseInstant("2026-10-22T14:03:05Z"))
let b = Result.unwrap(Time.parseInstant("2026-10-23T09:00:00Z"))
assert(b.since(a) == Time.negate(Time.until(b, a)))
assert(b.since(a) == Time.until(a, b))
assert(a.add(b.since(a)) == b)
```

## Possible issues

The arguments read as "`t` since `other`", the reverse order of `Time.until`. Swapping them flips
the sign:

```ant
let earlier = Time.plainDate(2026, 1, 1)
let later = Time.plainDate(2026, 1, 8)
later.since(earlier) // P7D
earlier.since(later) // -P7D
```

A `PlainTime` does not wrap around midnight. From 23:00 to 01:00 is 22 hours back, not 2 hours
forward:

```ant
Time.plainTime(1).since(Time.plainTime(23)) // -PT22H
```

A largest unit the type does not have panics:

```ant
let a = Time.fromEpochMillis(0)
let b = Time.fromEpochMillis(90_000_000)
b.since(a, largestUnit: Time.Day{}) // panic: Time.since: an instant has no days
Time.plainDate(2026, 1, 2).since(Time.plainDate(2026, 1, 1), largestUnit: Time.Hour{})
// panic: Time.since: a date has no hours
```

Days between two zones do not line up, so `Day` or larger panics:

```ant
let tokyo = Result.unwrap(Time.zone("Asia/Tokyo"))
let t = Time.fromEpochMillis(0)
t.inZone(Time.utc).since(t.inZone(tokyo), largestUnit: Time.Day{})
// panic: Time.since: UTC and Asia/Tokyo are different zones; a difference in days or larger needs one zone
```

Both points must have the same type:

```ant
let start = Time.fromEpochMillis(0)
let end = Time.fromEpochMillis(1000)
let d = end.since(start.inZone(Time.utc)) // error: cannot use Time.ZonedDateTime where Time.Instant is expected
```

## See also

[Time.until](until.md), [Time.negate](negate.md), [Time.add](add.md), [Time.subtract](subtract.md),
[Time.total](total.md), [Time.part](part.md), [Time.roundDuration](roundDuration.md),
[Time.Unit](types/Unit.md)
