<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.with

```ant signature
func Time.with[T: Time.Wall](t: T, year: int | null = null, month: int | null = null, day: int | null = null, hour: int | null = null, minute: int | null = null, second: int | null = null, millisecond: int | null = null, microsecond: int | null = null, nanosecond: int | null = null, disambiguation: Time.Disambiguation = Time.Compatible{}): T
```

Time.with returns t with the named fields replaced and the others kept. A field out of range is
constrained, so month 13 is December and day 31 in April is April 30. A PlainDate takes only the
date fields and a PlainTime only the time fields, and panics on another; on a ZonedDateTime the new
wall-clock time keeps the offset of t when the zone has it then, and is placed by disambiguation
otherwise.

<!-- end generated -->

## Details

- A field above its range becomes its largest value: month 13 is December, day 31 in April is April
  30, hour 25 is 23.
- A month or a day below 1 panics, while a time field below 0 becomes 0.

## Basic examples

```ant
Time.with(Time.plainDate(2026, 4, 10), day: 31)              // 2026-04-30
Time.with(Time.plainTime(14, 30), minute: 0)                 // 14:00:00
Time.with(Time.plainDate(2026, 4, 10), year: 2024, month: 2) // 2024-02-10
```

## Scope

### Every wall-clock type

A `PlainDate` takes `year`, `month` and `day`:

```ant
let day = Time.plainDate(2026, 4, 10)
day.with(month: 13)                     // 2026-12-10
day.with(year: 2024, month: 2, day: 30) // 2024-02-29
```

A `PlainTime` takes the six time fields. An hour or a minute past its range becomes the largest
value, and one below 0 becomes 0:

```ant
let t = Time.plainTime(14, 30)
t.with(millisecond: 250, nanosecond: 7) // 14:30:00.250000007
t.with(hour: 25, minute: 99)            // 23:59:00
t.with(second: -5)                      // 14:30:00
```

A `PlainDateTime` takes all nine:

```ant
let dt = Time.plainDateTime(Time.plainDate(2026, 4, 10), Time.plainTime(14, 30))
dt.with(hour: 0, minute: 0) // 2026-04-10T00:00:00
dt.with(day: 1, hour: 9)    // 2026-04-01T09:30:00
```

A `ZonedDateTime` keeps its zone:

```ant
let utc = Time.fromEpochMillis(0).inZone(Time.utc)
utc.with(year: 2000, hour: 12) // 2000-01-01T12:00:00+00:00[UTC]
```

### No fields and pipes

With no field named, the result equals `t`. Call it through `.` or in a pipe:

```ant
let day = Time.plainDate(2026, 4, 10)
day.with() == day       // true
day | Time.with(day: 1) // 2026-04-01
```

### Keeping the offset in a repeated hour

New York sets its clocks back on November 1, 2026, so 01:30 that day happens twice, at `-04:00` and
at `-05:00`. A `ZonedDateTime` moved into that hour keeps its own offset when it is one of the two:

```ant
let ny = Time.zone("America/New_York").unwrap()
let oct31 = Time.plainDateTime(Time.plainDate(2026, 10, 31), Time.plainTime(1, 30)).toZoned(ny)
let nov2 = Time.plainDateTime(Time.plainDate(2026, 11, 2), Time.plainTime(1, 30)).toZoned(ny)
oct31.with(month: 11, day: 1) // 2026-11-01T01:30:00-04:00[America/New_York]
nov2.with(day: 1)             // 2026-11-01T01:30:00-05:00[America/New_York]
```

## Named arguments

### disambiguation

When the new wall-clock time does not exist in the zone, or exists twice and the offset of `t` does
not settle it, `disambiguation` picks the instant. New York skips 02:00 to 03:00 on March 8, 2026:

```ant
let ny = Time.zone("America/New_York").unwrap()
let sat = Time.plainDateTime(Time.plainDate(2026, 3, 7), Time.plainTime(2, 30)).toZoned(ny)
sat.with(day: 8)                                 // 2026-03-08T03:30:00-04:00[America/New_York]
sat.with(day: 8, disambiguation: Time.Earlier{}) // 2026-03-08T01:30:00-05:00[America/New_York]
sat.with(day: 8, disambiguation: Time.Later{})   // 2026-03-08T03:30:00-04:00[America/New_York]
```

The plain types ignore it.

## Applications

The first and the last day of a month. A day past the end of every month becomes the last day:

```ant
let t = Time.plainDate(2026, 2, 14)
t.with(day: 1)  // 2026-02-01
t.with(day: 31) // 2026-02-28
```

An alarm at 07:15 on a given date:

```ant
let trip = Time.plainDateTime(Time.plainDate(2026, 6, 12), Time.plainTime())
trip.with(hour: 7, minute: 15) // 2026-06-12T07:15:00
```

Truncate a time to the start of its hour. Every field below the hour must be named, or a leftover
microsecond or nanosecond stays; [Time.round](round.md) with `Time.Trunc{}` does the same in one
step:

```ant
let t = Time.plainTime(9, 47, 12, 500, 3, 7)
t.with(minute: 0, second: 0, millisecond: 0)                                // 09:00:00.000003007
t.with(minute: 0, second: 0, millisecond: 0, microsecond: 0, nanosecond: 0) // 09:00:00
t.round(Time.Hour{}, mode: Time.Trunc{})                                    // 09:00:00
```

## Properties and relations

A day beyond the end of the month gives the same date as `Time.daysInMonth`:

```ant
let t = Time.plainDate(2024, 2, 10)
assert(t.with(day: 99) == t.with(day: t.daysInMonth()))
```

Replacing a field with its own value changes nothing, and the field reads back as set when it is in
range:

```ant
let t = Time.plainDateTime(Time.plainDate(2026, 4, 10), Time.plainTime(14, 30))
assert(t.with(month: t.month()) == t)
assert(t.with(hour: 8).hour() == 8)
```

## Possible issues

`Time.with` constrains where a constructor panics, so a month from outside the program is not
checked:

```ant
Time.plainDate(2026, 1, 1).with(month: 13) // 2026-12-01
Time.plainDate(2026, 13, 1)                // panic: Time.plainDate: no such date: 2026-13-01
```

A month or a day below 1 panics:

```ant
Time.plainDate(2026, 4, 10).with(day: 0)   // panic: Time.with: day below 1: 0
Time.plainDate(2026, 4, 10).with(month: 0) // panic: Time.with: month below 1: 0
```

A field the type lacks panics at run time:

```ant
Time.plainDate(2026, 4, 10).with(hour: 9) // panic: Time.with: a date has no time fields
Time.plainTime(9).with(day: 1)            // panic: Time.with: a time has no date fields
```

The day is constrained at each call, so changing the month twice can lose it. Name both fields in
one call instead:

```ant
let t = Time.plainDate(2026, 1, 31)
t.with(month: 2).with(month: 3) // 2026-03-28
t.with(month: 3)                // 2026-03-31
```

An `Instant` has no fields. Convert it to a `ZonedDateTime` first:

```ant
let start = Time.fromEpochMillis(0)
let later = Time.with(start, year: 2000) // error: T would be Time.Instant, which is not a member of its bound Time.Wall
```

## See also

[Time.plainDate](plainDate.md), [Time.plainTime](plainTime.md), [Time.add](add.md),
[Time.round](round.md), [Time.startOfDay](startOfDay.md), [Time.daysInMonth](daysInMonth.md),
[Time.Disambiguation](types/Disambiguation.md)
