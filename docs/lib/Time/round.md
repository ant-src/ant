<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.round

```ant signature
func Time.round[T: Time.Instant | Time.Timed](t: T, smallestUnit: Time.Unit, increment: int = 1, mode: Time.RoundingMode = Time.HalfExpand{}): T
```

Time.round rounds t to a multiple of increment of smallestUnit by mode, by default to the nearest
with a tie away from zero. It rounds the time since the start of the day, or since the epoch for an
Instant; a ZonedDateTime rounded to Day picks the start of its day or the next by how much of the
real day has passed. A PlainTime that rounds up to 24:00 wraps to 00:00:00, and the other wall-clock
types move to the next day. It panics on Year, Month or Week, on Day for an Instant or a PlainTime,
and unless increment is positive, divides the next larger unit (24, 60 or 1000) and is smaller than
it, or is 1 for Day.

<!-- end generated -->

## Basic examples

```ant
Time.round(Time.plainTime(14, 37), Time.Minute{}, increment: 15) // 14:30:00
Time.round(Time.plainTime(14, 37, 40), Time.Minute{})            // 14:38:00
```

A time exactly halfway rounds up by default:

```ant
Time.round(Time.plainTime(14, 30), Time.Hour{}) // 15:00:00
```

## Scope

### Every type with a time of day

A `PlainTime` wraps around the clock:

```ant
Time.plainTime(23, 50).round(Time.Hour{}) // 00:00:00
```

A `PlainDateTime` carries into the date, and rounds to a whole day:

```ant
let eve = Time.plainDateTime(Time.plainDate(2026, 12, 31), Time.plainTime(23, 50))
eve.round(Time.Hour{}) // 2027-01-01T00:00:00
let morning = Time.plainDateTime(Time.plainDate(2026, 5, 4), Time.plainTime(11, 59))
morning.round(Time.Day{}) // 2026-05-04T00:00:00
```

An `Instant` rounds in UTC:

```ant
let i = Result.unwrap(Time.parseInstant("2026-10-22T14:03:05.678Z"))
i.round(Time.Second{})                      // 2026-10-22T14:03:06Z
i.round(Time.Millisecond{}, increment: 100) // 2026-10-22T14:03:05.7Z
```

A `ZonedDateTime` keeps its zone:

```ant
let noon = Time.plainDateTime(Time.plainDate(2026, 5, 4), Time.plainTime(12)).toZoned(Time.utc)
noon.round(Time.Day{}) // 2026-05-05T00:00:00+00:00[UTC]
```

### Through a pipe

```ant
let t = Time.plainTime(9, 52, 10)
t | Time.round(Time.Minute{}, increment: 5) // 09:50:00
```

### Days across a daylight-saving change

New York sets its clocks forward on March 8, 2026, so that day has 23 hours. At 12:00 only 11 of
them have passed, and at 13:00 12 have, so noon rounds down and 13:00 rounds up:

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let day = Time.plainDate(2026, 3, 8)
Time.plainDateTime(day, Time.plainTime(12)).toZoned(ny).round(Time.Day{})
// 2026-03-08T00:00:00-05:00[America/New_York]
Time.plainDateTime(day, Time.plainTime(13)).toZoned(ny).round(Time.Day{})
// 2026-03-09T00:00:00-04:00[America/New_York]
```

A smaller unit works on the wall clock. 01:40 rounds to 02:00, which the change skips, so the result
moves forward out of the gap:

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let early = Time.plainDateTime(Time.plainDate(2026, 3, 8), Time.plainTime(1, 40)).toZoned(ny)
early.round(Time.Hour{}) // 2026-03-08T03:00:00-04:00[America/New_York]
```

In the hour that repeats when the clocks go back, the result keeps the value's offset:

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let late = Time.plainDateTime(Time.plainDate(2026, 11, 1), Time.plainTime(1, 40)).toZoned(ny, disambiguation: Time.Later{})
late.round(Time.Minute{}, increment: 30, mode: Time.Floor{}) // 2026-11-01T01:30:00-05:00[America/New_York]
```

## Named arguments

### increment

`increment` rounds to a multiple of several units. For the plain and zoned types it divides the next
larger unit, so quarter hours and six-hour blocks work:

```ant
let t = Time.plainTime(14, 37, 30)
t.round(Time.Minute{}, increment: 15) // 14:45:00
t.round(Time.Hour{}, increment: 6)    // 12:00:00
```

For an `Instant` it divides a day, and may be the whole day, which rounds to a UTC midnight:

```ant
let i = Result.unwrap(Time.parseInstant("2026-10-22T14:03:05Z"))
i.round(Time.Hour{}, increment: 24)     // 2026-10-23T00:00:00Z
i.round(Time.Minute{}, increment: 1440) // 2026-10-23T00:00:00Z
```

### mode

`mode` picks the direction. `Floor` and `Ceil` always go down or up, `Trunc` rounds toward zero,
which for a time of day is down, and `HalfEven` sends a tie to the even multiple:

```ant
let t = Time.plainTime(14, 37, 30)
t.round(Time.Minute{}, increment: 15, mode: Time.Floor{})          // 14:30:00
t.round(Time.Minute{}, increment: 15, mode: Time.Ceil{})           // 14:45:00
Time.plainTime(10, 7, 45).round(Time.Minute{}, mode: Time.Trunc{}) // 10:07:00
Time.plainTime(12, 30).round(Time.Hour{}, mode: Time.HalfEven{})   // 12:00:00
Time.plainTime(13, 30).round(Time.Hour{}, mode: Time.HalfEven{})   // 14:00:00
```

## Applications

Bill time in started quarter hours:

```ant
func billable(end: Time.PlainTime): Time.PlainTime {
  end.round(Time.Minute{}, increment: 15, mode: Time.Ceil{})
}

billable(Time.plainTime(16, 1)) // 16:15:00
billable(Time.plainTime(16, 0)) // 16:00:00
```

Group log events into five-minute buckets:

```ant
let events = [
  "2026-10-22T14:03:05Z",
  "2026-10-22T14:04:59Z",
  "2026-10-22T14:05:00Z",
  "2026-10-22T14:12:30Z",
]
for e in events {
  let at = Result.unwrap(Time.parseInstant(e))
  let bucket = at.round(Time.Minute{}, increment: 5, mode: Time.Floor{})
  println("${bucket} ${e}")
}
```

```text
2026-10-22T14:00:00Z 2026-10-22T14:03:05Z
2026-10-22T14:00:00Z 2026-10-22T14:04:59Z
2026-10-22T14:05:00Z 2026-10-22T14:05:00Z
2026-10-22T14:10:00Z 2026-10-22T14:12:30Z
```

## Properties and relations

`Floor` and `Ceil` bracket the value, and a value already on a multiple does not move:

```ant
let t = Time.plainDateTime(Time.plainDate(2026, 6, 1), Time.plainTime(8, 17, 3))
let down = t.round(Time.Minute{}, increment: 10, mode: Time.Floor{})
let up = t.round(Time.Minute{}, increment: 10, mode: Time.Ceil{})
assert(Time.compare(down, t) <= 0 && Time.compare(t, up) <= 0)
assert(down.round(Time.Minute{}, increment: 10) == down)
```

Rounding to a `Nanosecond` changes nothing:

```ant
let i = Result.unwrap(Time.parseInstant("2026-10-22T14:03:05.123456789Z"))
assert(i.round(Time.Nanosecond{}) == i)
```

## Possible issues

An `Instant` before 1970 counts as negative, so `Trunc`, which goes toward zero, rounds it forward
in time. Use `Floor` to always round down:

```ant
let i = Result.unwrap(Time.parseInstant("1969-12-31T23:59:59.5Z"))
i.round(Time.Second{}, mode: Time.Trunc{}) // 1970-01-01T00:00:00Z
i.round(Time.Second{}, mode: Time.Floor{}) // 1969-12-31T23:59:59Z
i.round(Time.Second{})                     // 1969-12-31T23:59:59Z
```

An increment must divide the next larger unit and be smaller than it; to round to the hour, use
`Time.Hour`, not 60 minutes:

```ant
Time.plainTime(9, 7).round(Time.Minute{}, increment: 7)
// panic: Time.round: the increment for minutes must divide 60 and be smaller, not 7
```

```ant
Time.plainTime(9, 7).round(Time.Minute{}, increment: 60)
// panic: Time.round: the increment for minutes must divide 60 and be smaller, not 60
```

A `PlainTime` has no date and an `Instant` no calendar, so neither rounds to a day, and no type
rounds to a week, month or year:

```ant
let start = Time.fromEpochMillis(0)
start.round(Time.Day{})               // panic: Time.round: an instant cannot be rounded to a day
Time.plainTime(9).round(Time.Day{})   // panic: Time.round: a time cannot be rounded to a day
Time.plainTime(9).round(Time.Month{}) // panic: Time.round: cannot round to months; round a duration with Time.roundDuration
```

A `PlainDate` has no time of day to round:

```ant
let d = Time.round(Time.plainDate(2026, 5, 4), Time.Day{})
// error: T would be Time.PlainDate, which is not a member of its bound Time.Instant | Time.Timed
```

## See also

[Time.roundDuration](roundDuration.md), [Time.startOfDay](startOfDay.md), [Time.with](with.md),
[Time.add](add.md), [Time.RoundingMode](types/RoundingMode.md), [Time.Unit](types/Unit.md),
[Math.round](../Math/round.md)
