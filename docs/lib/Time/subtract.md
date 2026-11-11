<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.subtract

```ant signature
func Time.subtract[T: Time.Point](t: T, d: Time.Duration): T
```

Time.subtract returns t moved back by d, which is Time.add(t, Time.negate(d)). It panics as
[`Time.add`](add.md) does.

<!-- end generated -->

## Details

- `Time.subtract` follows the same order as `Time.add`: years and months first with the day
  constrained, then weeks and days, then hours and smaller.
- A negative duration moves forward.
- It panics where `Time.add` does: days or larger on an `Instant`, or a result outside the range of
  its type.

## Basic examples

```ant
Time.subtract(Time.plainDate(2024, 3, 31), Time.duration(months: 1)) // 2024-02-29
Time.subtract(Time.plainTime(0, 30), Time.duration(hours: 1))        // 23:30:00
```

A negative duration moves forward:

```ant
Time.subtract(Time.plainDate(2026, 3, 1), Time.duration(days: -1)) // 2026-03-02
```

## Scope

### Every point type

A `PlainDate` drops what is less than a whole day:

```ant
let day = Time.plainDate(2026, 1, 1)
day.subtract(Time.duration(weeks: 1))  // 2025-12-25
day.subtract(Time.duration(hours: 25)) // 2025-12-31
day.subtract(Time.duration(hours: 1))  // 2026-01-01
```

A `PlainTime` wraps around the clock and ignores days:

```ant
let nine = Time.plainTime(9)
nine.subtract(Time.duration(hours: 10)) // 23:00:00
nine.subtract(Time.duration(days: 3))   // 09:00:00
```

A `PlainDateTime` borrows from the date, and an `Instant` takes hours and smaller units:

```ant
let newYear = Time.plainDateTime(Time.plainDate(2027, 1, 1), Time.plainTime(1))
newYear.subtract(Time.duration(hours: 2))                    // 2026-12-31T23:00:00
Time.fromEpochMillis(0).subtract(Time.duration(seconds: 90)) // 1969-12-31T23:58:30Z
```

A `ZonedDateTime` keeps its zone. A day moves the wall clock and an hour moves the instant, so
across a daylight-saving change they differ. New York sets its clocks forward on March 8, 2026:

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let sunday = Time.plainDateTime(Time.plainDate(2026, 3, 8), Time.plainTime(12)).toZoned(ny)
sunday.subtract(Time.duration(days: 1))   // 2026-03-07T12:00:00-05:00[America/New_York]
sunday.subtract(Time.duration(hours: 24)) // 2026-03-07T11:00:00-05:00[America/New_York]
```

### Durations with several fields

The month goes first and the day after it, so the order can change the result:

```ant
let end = Time.plainDate(2026, 3, 31)
end.subtract(Time.duration(months: 1, days: 1))                         // 2026-02-27
end.subtract(Time.duration(days: 1)).subtract(Time.duration(months: 1)) // 2026-02-28
```

### Pipes and callbacks

```ant
Time.plainDate(2026, 1, 1) | Time.subtract(Time.duration(weeks: 1)) // 2025-12-25
let ends = [Time.plainTime(10), Time.plainTime(0, 20)]
ends.map({ x.subtract(Time.duration(minutes: 45)) }) // [09:15:00, 23:35:00]
```

## Applications

When to leave so that a trip of a known length arrives on time:

```ant
func departure(arrival: Time.PlainDateTime, trip: Time.Duration): Time.PlainDateTime {
  arrival.subtract(trip)
}

let arrival = Time.plainDateTime(Time.plainDate(2026, 11, 20), Time.plainTime(8, 30))
departure(arrival, Time.duration(hours: 9, minutes: 45)) // 2026-11-19T22:45:00
```

The start of a 30-day window that ends today, for a report:

```ant
let today = Time.plainDate(2026, 11, 11)
let start = today.subtract(Time.duration(days: 29))
println("${start} to ${today}")
```

```text
2026-10-13 to 2026-11-11
```

Reminders before a meeting:

```ant
let meeting = Time.plainDateTime(Time.plainDate(2026, 11, 12), Time.plainTime(10))
let leads = [Time.duration(days: 1), Time.duration(minutes: 15)]
leads.map({ meeting.subtract(x) }) // [2026-11-11T10:00:00, 2026-11-12T09:45:00]
```

## Properties and relations

`Time.subtract` adds the negated duration, and for exact units on an `Instant` it undoes `Time.add`:

```ant
let t = Time.fromEpochMillis(1_000_000)
let ds = [Time.duration(hours: 5, seconds: 7), Time.duration(minutes: -3), Time.duration()]
for d in ds {
  assert(t.subtract(d) == t.add(d.negate()))
  assert(t.add(d).subtract(d) == t)
}
```

With its default unit, `Time.until` counts whole days between two dates, so subtracting that
duration goes back to the start:

```ant
let start = Time.plainDate(2026, 1, 31)
let end = Time.plainDate(2026, 3, 15)
end.subtract(Time.until(start, end)) == start // true
```

## Possible issues

An `Instant` has no calendar, so a day panics. Use 24 hours, or convert to a `ZonedDateTime`:

```ant
let start = Time.fromEpochMillis(0)
start.subtract(Time.duration(days: 1))
// panic: Time.subtract: an instant has no calendar, so it takes no days; convert it with Time.inZone first
start.subtract(Time.duration(hours: 24)) // 1969-12-31T00:00:00Z
```

Months do not undo each other, because the day is constrained on the way:

```ant
let end = Time.plainDate(2026, 3, 31)
end.subtract(Time.duration(months: 1))                               // 2026-02-28
end.subtract(Time.duration(months: 1)).add(Time.duration(months: 1)) // 2026-03-28
```

With a calendar `largestUnit`, the duration `Time.until` measures does not always lead back to the
start, because the day is constrained when the month is subtracted:

```ant
let start = Time.plainDate(2026, 1, 30)
let end = Time.plainDate(2026, 3, 1)
let d = Time.until(start, end, largestUnit: Time.Month{})
d               // P1M1D
end.subtract(d) // 2026-01-31
```

A result beyond the range of its type panics:

```ant
Time.plainDate(-271821, 4, 20).subtract(Time.duration(days: 2))
// panic: Time.subtract: outside the range of time values
```

A field of the most negative `int` cannot be negated:

```ant
Time.plainDate(2026, 1, 1).subtract(Time.duration(nanoseconds: Math.minInt))
// panic: Time.subtract: the negated duration has a field that does not fit an int
```

## See also

[Time.add](add.md), [Time.negate](negate.md), [Time.until](until.md), [Time.since](since.md),
[Time.duration](duration.md), [Time.addDuration](addDuration.md)
