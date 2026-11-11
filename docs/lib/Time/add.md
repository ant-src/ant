<!-- status: reviewed -->
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

## Details

- Months and years constrain the day to the end of a shorter month instead of panicking.
- One duration applies its years and months before its days, which can differ from adding the days
  first.

## Basic examples

```ant
Time.add(Time.plainDate(2024, 1, 31), Time.duration(months: 1)) // 2024-02-29
Time.add(Time.plainTime(23, 30), Time.duration(hours: 2))       // 01:30:00
```

A negative duration moves back:

```ant
Time.add(Time.plainDate(2026, 3, 1), Time.duration(days: -1)) // 2026-02-28
```

## Scope

### Every point type

A `PlainDate` takes calendar units and whole days of the exact part:

```ant
let day = Time.plainDate(2026, 1, 1)
day.add(Time.duration(years: 1, weeks: 2)) // 2027-01-15
day.add(Time.duration(hours: 47))          // 2026-01-02
day.add(Time.duration(hours: -1))          // 2026-01-01
```

A `PlainTime` wraps around the clock and ignores days:

```ant
let noon = Time.plainTime(12)
noon.add(Time.duration(hours: 13))          // 01:00:00
noon.add(Time.duration(days: 3, hours: 1))  // 13:00:00
noon.add(Time.duration(milliseconds: -500)) // 11:59:59.5
```

A `PlainDateTime` carries the clock into the date:

```ant
let eve = Time.plainDateTime(Time.plainDate(2026, 12, 31), Time.plainTime(23))
eve.add(Time.duration(hours: 2))          // 2027-01-01T01:00:00
eve.add(Time.duration(days: 1, hours: 2)) // 2027-01-02T01:00:00
```

An `Instant` takes hours and smaller units:

```ant
let start = Time.fromEpochMillis(0)
start.add(Time.duration(hours: 36, minutes: 30)) // 1970-01-02T12:30:00Z
start.add(Time.duration(seconds: -1))            // 1969-12-31T23:59:59Z
```

A `ZonedDateTime` keeps its zone:

```ant
let utc = Time.fromEpochMillis(0).inZone(Time.utc)
utc.add(Time.duration(months: 1, hours: 6)) // 1970-02-01T06:00:00+00:00[UTC]
```

### Durations with several fields

Months go first, then days, then the exact part. From January 30, a month gives February 29 and the
day after it is March 1:

```ant
let day = Time.plainDate(2024, 1, 30)
day.add(Time.duration(months: 1, days: 1))                    // 2024-03-01
day.add(Time.duration(days: 1)).add(Time.duration(months: 1)) // 2024-02-29
```

### Zero and pipes

Adding a zero duration gives the same value:

```ant
let day = Time.plainDate(2026, 5, 4)
day.add(Time.duration()) == day // true
```

Call it through `.` or in a pipe:

```ant
Time.plainDate(2026, 5, 4).add(Time.duration(days: 7))        // 2026-05-11
Time.plainDate(2026, 5, 4) | Time.add(Time.duration(days: 7)) // 2026-05-11
```

### Days and hours across a daylight-saving change

On a `ZonedDateTime`, a day moves the wall clock and an hour moves the instant. New York sets its
clocks forward on March 8, 2026, so that day has 23 hours:

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let saturday = Time.plainDateTime(Time.plainDate(2026, 3, 7), Time.plainTime(12)).toZoned(ny)
saturday.add(Time.duration(days: 1))   // 2026-03-08T12:00:00-04:00[America/New_York]
saturday.add(Time.duration(hours: 24)) // 2026-03-08T13:00:00-04:00[America/New_York]
```

It sets them back on November 1, 2026, so that day has 25 hours and 24 hours end an hour short of
noon:

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let saturday = Time.plainDateTime(Time.plainDate(2026, 10, 31), Time.plainTime(12)).toZoned(ny)
saturday.add(Time.duration(days: 1))   // 2026-11-01T12:00:00-05:00[America/New_York]
saturday.add(Time.duration(hours: 24)) // 2026-11-01T11:00:00-05:00[America/New_York]
```

A day that lands in the skipped hour moves forward out of it:

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let early = Time.plainDateTime(Time.plainDate(2026, 3, 7), Time.plainTime(2, 30)).toZoned(ny)
early.add(Time.duration(days: 1)) // 2026-03-08T03:30:00-04:00[America/New_York]
```

## Applications

A due date 30 days after an invoice:

```ant
func dueDate(issued: Time.PlainDate): Time.PlainDate {
  issued.add(Time.duration(days: 30))
}

dueDate(Time.plainDate(2026, 11, 15)) // 2026-12-15
```

A deadline five minutes from a fixed instant, checked with `Time.compare`:

```ant
let started = Result.unwrap(Time.parseInstant("2026-10-22T14:03:05Z"))
let deadline = started.add(Time.duration(minutes: 5))
let now = Result.unwrap(Time.parseInstant("2026-10-22T14:09:00Z"))
Time.compare(now, deadline) > 0 // true
```

Monthly billing dates anchored on the 31st. Adding `i` months to the first date keeps the day where
the month allows it:

```ant
let first = Time.plainDate(2026, 1, 31)
for i in Iter.range(0, 4) {
  println(first.add(Time.duration(months: i)))
}
```

```text
2026-01-31
2026-02-28
2026-03-31
2026-04-30
```

A weekly meeting stays at 09:00 in its zone across the change of clocks:

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let first = Time.plainDateTime(Time.plainDate(2026, 2, 26), Time.plainTime(9)).toZoned(ny)
for week in Iter.range(0, 3) {
  println(first.add(Time.duration(weeks: week)))
}
```

```text
2026-02-26T09:00:00-05:00[America/New_York]
2026-03-05T09:00:00-05:00[America/New_York]
2026-03-12T09:00:00-04:00[America/New_York]
```

## Properties and relations

`Time.subtract` adds the negated duration:

```ant
let t = Time.plainDateTime(Time.plainDate(2026, 6, 1), Time.plainTime(8, 15))
let d = Time.duration(days: 2, hours: 3)
assert(t.subtract(d) == t.add(d.negate()))
```

For exact units on an `Instant`, subtracting undoes adding, and `Time.until` measures the step:

```ant
let t = Time.fromEpochMillis(1_000_000)
let d = Time.duration(hours: 5, seconds: 7)
assert(t.add(d).subtract(d) == t)
assert(t.until(t.add(d)) == Time.duration(seconds: 18007))
```

## Possible issues

An `Instant` has no calendar, so a day panics. Use 24 hours, or convert to a `ZonedDateTime`:

```ant
let start = Time.fromEpochMillis(0)
start.add(Time.duration(days: 1))   // panic: Time.add: an instant has no calendar, so it takes no days; convert it with Time.inZone first
start.add(Time.duration(hours: 24)) // 1970-01-02T00:00:00Z
```

Months do not undo each other. The day is constrained at each step, so two single months from
January 31 end on March 28, while two months at once end on March 31:

```ant
let day = Time.plainDate(2025, 1, 31)
day.add(Time.duration(months: 1)).add(Time.duration(months: 1))      // 2025-03-28
day.add(Time.duration(months: 2))                                    // 2025-03-31
day.add(Time.duration(months: 1)).subtract(Time.duration(months: 1)) // 2025-01-28
```

A `PlainDate` drops what is less than a day, so small steps add up to nothing:

```ant
let day = Time.plainDate(2026, 1, 1)
day.add(Time.duration(hours: 12)).add(Time.duration(hours: 12)) // 2026-01-01
day.add(Time.duration(hours: 24))                               // 2026-01-02
```

A value typed with a union such as `Time.Dated` does not pass until an `is` test narrows it:

```ant
func tomorrow(t: Time.Dated): Time.Dated {
  Time.add(t, Time.duration(days: 1)) // error: T would be Time.Dated, which is not a single member of its bound Time.Point; narrow first
}
```

Each branch of an `is` test narrows `t` to a single member, and there the call passes:

```ant
func tomorrow(t: Time.Dated): Time.Dated {
  let day = Time.duration(days: 1)
  if t is Time.ZonedDateTime {
    t.add(day)
  } else if t is Time.PlainDate {
    t.add(day)
  } else {
    t.add(day)
  }
}

println(tomorrow(Time.plainDate(2026, 5, 4)))
```

```text
2026-05-05
```

A result beyond the range of its type panics:

```ant
Time.plainDate(275760, 9, 13).add(Time.duration(days: 2)) // panic: Time.add: outside the range of time values
```

## See also

[Time.subtract](subtract.md), [Time.duration](duration.md), [Time.until](until.md),
[Time.since](since.md), [Time.with](with.md), [Time.addDuration](addDuration.md),
[Time.compare](compare.md)
