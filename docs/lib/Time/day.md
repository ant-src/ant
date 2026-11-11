<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.day

```ant signature
func Time.day(t: Time.Dated): int
```

Time.day returns the day of the month of t, from 1.

<!-- end generated -->

## Details

- The day counts from 1, so the first of the month is 1 and the last is
  [Time.daysInMonth](daysInMonth.md) of the same value.
- On a `Time.ZonedDateTime`, the day is the one the wall clock in its zone shows. The same instant
  can have a different day in another zone.

## Basic examples

```ant
Time.day(Time.plainDate(2026, 10, 22)) // 22
Time.day(Time.plainDate(2024, 2, 29))  // 29
```

Call it through `.` on any value with a date:

```ant
let date = Time.plainDate(2026, 3, 5)
date.day() // 5
```

## Scope

A `Time.PlainDateTime` has the day of its date, whatever the time:

```ant
let late = Time.plainDateTime(Time.plainDate(2026, 2, 28), Time.plainTime(23, minute: 59))
late.day() // 28
```

A `Time.ZonedDateTime` has the day its zone's wall clock shows:

```ant
let stockholm = Time.parseZoned("2026-10-22T01:30:00+02:00[Europe/Stockholm]").unwrap()
stockholm.day() // 22
```

A value typed `Time.Dated` may hold any of the three types with a date, and `Time.day` takes it
without narrowing. Pass `Time.day` by name to read each one:

```ant
let meeting = Time.plainDateTime(Time.plainDate(2026, 1, 6), Time.plainTime(9))
let launch = Time.parseZoned("2026-01-07T12:00:00Z[UTC]").unwrap()
let dates: Array[Time.Dated] = [Time.plainDate(2026, 1, 5), meeting, launch]
dates.map(Time.day) // [5, 6, 7]
```

In a pipe, the value is the first argument:

```ant
"2026-12-24" | Time.parsePlainDate() | Result.map(Time.day) // Ok{ value: 24 }
```

## Applications

Write a date with an English ordinal day, such as "October 22nd":

```ant
let months = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December",
]

func ordinal(n: int): string {
  let teen = n % 100 >= 11 && n % 100 <= 13
  let suffix = if teen {
    "th"
  } else if n % 10 == 1 {
    "st"
  } else if n % 10 == 2 {
    "nd"
  } else if n % 10 == 3 {
    "rd"
  } else {
    "th"
  }
  "${n}${suffix}"
}

func label(d: Time.PlainDate): string {
  "${months[d.month() - 1]} ${ordinal(d.day())}"
}

println(label(Time.plainDate(2026, 10, 22)))
println(label(Time.plainDate(2026, 3, 1)))
println(label(Time.plainDate(2026, 6, 13)))
```

```text
October 22nd
March 1st
June 13th
```

Find the dates that fall on the last day of their month:

```ant
let dates = [
  Time.plainDate(2026, 1, 31),
  Time.plainDate(2026, 2, 28),
  Time.plainDate(2024, 2, 28),
  Time.plainDate(2026, 4, 30),
  Time.plainDate(2026, 5, 30),
]
let monthEnds = dates.filter({ x.day() == x.daysInMonth() })
println(monthEnds)
```

```text
[2026-01-31, 2026-02-28, 2026-04-30]
```

A bill due on the 31st falls on the last day of shorter months. [Time.with](with.md) constrains the
day, and `Time.day` reads the day it picked:

```ant
let firsts = [1, 2, 3, 4].map({ Time.plainDate(2026, x, 1) })
let dueDays = firsts.map({ x.with(day: 31).day() })
println(dueDays)
```

```text
[31, 28, 31, 30]
```

## Properties and relations

The day is between 1 and the number of days in the month:

```ant
let dates = [Time.plainDate(2026, 1, 1), Time.plainDate(2024, 2, 29), Time.plainDate(2026, 4, 30)]
for d in dates {
  assert(d.day() >= 1)
  assert(d.day() <= d.daysInMonth())
}
```

`Time.day` returns the day a date was made with, and setting the day with `Time.with` sets what
`Time.day` reads:

```ant
let d = Time.plainDate(2026, 7, 14)
assert(d.day() == 14)
assert(d.with(day: 3).day() == 3)
assert(d.with(day: 1).day() == 1)
```

A `Time.PlainDateTime` and a `Time.ZonedDateTime` have the day of their date:

```ant
let zoned = Time.parseZoned("2026-10-22T23:30:00+02:00[Europe/Stockholm]").unwrap()
let local = zoned.toPlainDateTime()
assert(zoned.day() == zoned.toPlainDate().day())
assert(local.day() == local.toPlainDate().day())
```

In January, the day of the month is the day of the year:

```ant
let d = Time.plainDate(2026, 1, 17)
assert(d.day() == d.dayOfYear())
```

## Possible issues

A `Time.Instant` has no day, so passing one is a compile error:

```ant
let epoch = Time.fromEpochMillis(0)
println(Time.day(epoch)) // error: cannot use Time.Instant where Time.Dated is expected
```

Put the instant in a zone first. The day depends on the zone: the same instant is October 21 in UTC
and October 22 in Stockholm:

```ant
let instant = Time.parseInstant("2026-10-21T23:30:00Z").unwrap()
let stockholm = Time.zone("Europe/Stockholm").unwrap()
instant.inZone(Time.utc).day()  // 21
instant.inZone(stockholm).day() // 22
```

`Time.day` is the day of the month, not the day of the week. Use [Time.dayOfWeek](dayOfWeek.md) for
Monday to Sunday:

```ant
let d = Time.plainDate(2026, 10, 22)
d.day()       // 22
d.dayOfWeek() // 4
```

Adding a month constrains the day to the end of a shorter month, and the day stays there. Two months
after January 31 is March 31, but one month and then another is March 28:

```ant
let jan31 = Time.plainDate(2026, 1, 31)
let month = Time.duration(months: 1)
jan31.add(Time.duration(months: 2)).day() // 31
jan31.add(month).add(month).day()         // 28
```

## See also

[Time.month](month.md) · [Time.year](year.md) · [Time.dayOfWeek](dayOfWeek.md) ·
[Time.dayOfYear](dayOfYear.md) · [Time.daysInMonth](daysInMonth.md) · [Time.with](with.md) ·
[Time.plainDate](plainDate.md) · [Time](index.md)
