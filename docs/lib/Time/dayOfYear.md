<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.dayOfYear

```ant signature
func Time.dayOfYear(t: Time.Dated): int
```

Time.dayOfYear returns the day of the year of t, 1 for January 1 to 365, or 366 in a leap year.

<!-- end generated -->

## Details

- The result runs from 1 to 365, or to 366 in a leap year, where every date from March 1 on is one
  higher.
- A `Time.ZonedDateTime` counts the date its zone shows, not the date in UTC.

## Basic examples

January 1 is day 1:

```ant
Time.dayOfYear(Time.plainDate(2026, 1, 1)) // 1
```

December 31 is day 365, or 366 in a leap year:

```ant
Time.dayOfYear(Time.plainDate(2026, 12, 31)) // 365
Time.dayOfYear(Time.plainDate(2024, 12, 31)) // 366
```

## Scope

A `PlainDateTime` counts its date and ignores the time:

```ant
let start = Time.plainDateTime(Time.plainDate(2026, 2, 1), Time.plainTime(hour: 9, minute: 15))
start.dayOfYear() // 32
```

A `ZonedDateTime` counts the date of its own zone:

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let eve = Time.plainDateTime(Time.plainDate(2026, 12, 31), Time.plainTime(hour: 23)).toZoned(ny)
eve.dayOfYear()                    // 365
eve.withZone(Time.utc).dayOfYear() // 1
```

Pass it as a function value to map a parsed date:

```ant
Time.parsePlainDate("2026-07-04") | Result.map(Time.dayOfYear) // Ok{ value: 185 }
```

A negative year counts the same way:

```ant
Time.plainDate(-1, 12, 31).dayOfYear() // 365
```

## Applications

Count the days left in the year:

```ant
let today = Time.plainDate(2026, 11, 11)
today.daysInYear() - today.dayOfYear() // 50
```

Show how far through the year a date is, as a percentage:

```ant
let today = Time.plainDate(2026, 11, 11)
today.dayOfYear() * 100 / today.daysInYear() // 86
```

Turn a year and a day number back into a date by adding days to January 1. Without the guard, a day
number past the end of the year would roll into the next year:

```ant
func fromDayOfYear(year: int, n: int): Time.PlainDate {
  let jan1 = Time.plainDate(year, 1, 1)
  if n < 1 || n > jan1.daysInYear() {
    panic("day ${n} is not in ${year}")
  }
  jan1.add(Time.duration(days: n - 1))
}

println(fromDayOfYear(2024, 60))
println(fromDayOfYear(2026, 60))
```

```text
2024-02-29
2026-03-01
```

## Properties and relations

The last day of the year is `daysInYear`:

```ant
let t = Time.plainDate(2024, 12, 31)
assert(t.dayOfYear() == t.daysInYear())
```

The day of the year is one more than the days since January 1:

```ant
let t = Time.plainDate(2026, 11, 11)
let jan1 = t.with(month: 1, day: 1)
assert(t.dayOfYear() == jan1.until(t).part(Time.Day{}) + 1)
```

The day before January 1 is the last day of the previous year:

```ant
let jan1 = Time.plainDate(2025, 1, 1)
let before = jan1.subtract(Time.duration(days: 1))
assert(before.dayOfYear() == before.daysInYear())
```

## Possible issues

An `Instant` has no date, so it does not compile. Put it in a zone first:

```ant
let epoch = Time.fromEpochMillis(0)
println(Time.dayOfYear(epoch))
// error: cannot use Time.Instant where Time.Dated is expected
```

```ant
Time.fromEpochMillis(0).inZone(Time.utc).dayOfYear() // 1
```

A `PlainTime` has no date either:

```ant
let alarm = Time.plainTime(hour: 7)
println(Time.dayOfYear(alarm))
// error: cannot use Time.PlainTime where Time.Dated is expected
```

The same calendar date has a different number in a leap year, so day numbers of different years do
not name the same date:

```ant
Time.plainDate(2024, 3, 1).dayOfYear() // 61
Time.plainDate(2026, 3, 1).dayOfYear() // 60
```

## See also

[Time.day](day.md) · [Time.dayOfWeek](dayOfWeek.md) · [Time.weekOfYear](weekOfYear.md) ·
[Time.daysInYear](daysInYear.md) · [Time.inLeapYear](inLeapYear.md) · [Time.until](until.md) ·
[Time.add](add.md) · [Time.PlainDate](types/PlainDate.md)
