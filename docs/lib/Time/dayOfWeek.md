<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.dayOfWeek

```ant signature
func Time.dayOfWeek(t: Time.Dated): int
```

Time.dayOfWeek returns the day of the week of t, 1 for Monday to 7 for Sunday, as in ISO 8601.

<!-- end generated -->

## Details

- A `Time.Instant` has no day of the week until you place it in a zone with `Time.inZone`.

## Basic examples

```ant
Time.dayOfWeek(Time.plainDate(2026, 11, 11)) // 3
Time.dayOfWeek(Time.plainDate(2026, 11, 15)) // 7
```

Call it through `.` on any date:

```ant
let release = Time.plainDate(2026, 11, 8)
release.dayOfWeek() // 7
```

## Scope

A `Time.PlainDateTime` gives the day of its date, whatever the time:

```ant
let late = Time.plainDateTime(Time.plainDate(2026, 11, 11), Time.plainTime(hour: 23, minute: 30))
late.dayOfWeek() // 3
```

A `Time.ZonedDateTime` gives the day as its zone sees it. The same instant can fall on two different
days in two zones:

```ant
let stockholm = Time.zone("Europe/Stockholm").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
let evening = Time.plainDateTime(Time.plainDate(2026, 11, 11), Time.plainTime(hour: 23))
let meeting = evening.toZoned(stockholm)
meeting.dayOfWeek()                 // 3
meeting.withZone(tokyo).dayOfWeek() // 4
```

An instant reaches it once it has a zone:

```ant
Time.fromEpochMillis(0) | Time.inZone(Time.utc) | Time.dayOfWeek() // 4
```

## Named arguments

The parameter is named `t`:

```ant
Time.dayOfWeek(t: Time.plainDate(2026, 11, 11)) // 3
```

## Applications

Name the day. Subtract 1 to index an array that starts with Monday:

```ant
let names = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]
names[Time.plainDate(2026, 11, 11).dayOfWeek() - 1] // "Wed"
```

Test for a weekend:

```ant
func isWeekend(d: Time.PlainDate): bool {
  d.dayOfWeek() >= 6
}

isWeekend(Time.plainDate(2026, 11, 14)) // true
isWeekend(Time.plainDate(2026, 11, 16)) // false
```

Find the next Monday after a date:

```ant
func nextMonday(d: Time.PlainDate): Time.PlainDate {
  d.add(Time.duration(days: 8 - d.dayOfWeek()))
}

nextMonday(Time.plainDate(2026, 11, 11)) // 2026-11-16
nextMonday(Time.plainDate(2026, 11, 16)) // 2026-11-23
```

Find the Monday that starts the week of a date:

```ant
func startOfWeek(d: Time.PlainDate): Time.PlainDate {
  d.subtract(Time.duration(days: d.dayOfWeek() - 1))
}

startOfWeek(Time.plainDate(2026, 11, 11)) // 2026-11-09
startOfWeek(Time.plainDate(2026, 11, 9))  // 2026-11-09
```

Count the working days of a month:

```ant
func weekdaysIn(year: int, month: int): int {
  let first = Time.plainDate(year, month, 1)
  Array.range(0, first.daysInMonth())
    | Array.countWhere({ first.add(Time.duration(days: x)).dayOfWeek() <= 5 })
}

weekdaysIn(2026, 11) // 21
weekdaysIn(2026, 2)  // 20
```

Convert to the Sunday-first numbering of JavaScript's `getDay`, 0 for Sunday to 6 for Saturday:

```ant
Time.plainDate(2026, 11, 15).dayOfWeek() % 7 // 0
Time.plainDate(2026, 11, 14).dayOfWeek() % 7 // 6
```

## Properties and relations

The result is always 1 to 7, a week later is the same day, and the next day is one later, wrapping
from Sunday to Monday:

```ant
let start = Time.plainDate(2026, 12, 28)
for i in Array.range(0, 14) {
  let d = start.add(Time.duration(days: i))
  let n = d.dayOfWeek()
  assert(n >= 1 && n <= 7)
  assert(d.add(Time.duration(weeks: 1)).dayOfWeek() == n)
  assert(d.add(Time.duration(days: 1)).dayOfWeek() == n % 7 + 1)
}
```

A date-time and a zoned date-time have the day of their date:

```ant
let stockholm = Time.zone("Europe/Stockholm").unwrap()
let dt = Time.plainDateTime(Time.plainDate(2026, 3, 29), Time.plainTime(hour: 2, minute: 30))
let zoned = dt.toZoned(stockholm)
assert(dt.dayOfWeek() == dt.toPlainDate().dayOfWeek())
assert(zoned.dayOfWeek() == zoned.toPlainDate().dayOfWeek())
```

## Possible issues

Monday is 1, not 0, so the result is not an index. Indexing a seven-element array with it panics on
a Sunday:

```ant
let names = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]
let sunday = Time.plainDate(2026, 11, 15)
names[sunday.dayOfWeek()] // panic: index 7 out of range for an array of length 7
```

An instant has no day of the week until you choose a zone:

```ant
println(Time.dayOfWeek(Time.now())) // error: cannot use Time.Instant where Time.Dated is expected
```

A time of day has no date:

```ant
Time.dayOfWeek(Time.plainTime(hour: 9)) // error: cannot use Time.PlainTime where Time.Dated is expected
```

## See also

[Time.day](day.md) · [Time.dayOfYear](dayOfYear.md) · [Time.weekOfYear](weekOfYear.md) ·
[Time.yearOfWeek](yearOfWeek.md) · [Time.add](add.md) · [Time.inZone](inZone.md) ·
[Time.PlainDate](types/PlainDate.md)
