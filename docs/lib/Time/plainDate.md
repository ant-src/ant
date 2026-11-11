<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.plainDate

```ant signature
func Time.plainDate(year: int, month: int, day: int): Time.PlainDate
```

Time.plainDate returns the date with the given year, month and day. It panics if there is no such
date, such as month 13 or February 30; check input from the world against
[`Time.daysInMonth`](daysInMonth.md) first, or hand its text to
[`Time.parsePlainDate`](parsePlainDate.md).

<!-- end generated -->

## Details

- The month counts from 1 for January. A month outside 1 to 12, or a day below 1 or past the end of
  the month, panics.
- Two dates are `==` when they are the same day. Order them with [Time.compare](compare.md), not
  `<`.

## Basic examples

```ant
Time.plainDate(2026, 10, 22) // 2026-10-22
Time.plainDate(2024, 2, 29)  // 2024-02-29
```

The fields read back through `.`:

```ant
let release = Time.plainDate(2026, 10, 22)
release.year()  // 2026
release.month() // 10
release.day()   // 22
```

## Scope

Name the arguments when the order is not obvious to the reader:

```ant
Time.plainDate(year: 2026, month: 3, day: 1) // 2026-03-01
```

Years outside 0 to 9999 render with a sign and six digits:

```ant
Time.plainDate(0, 1, 1)     // 0000-01-01
Time.plainDate(-1, 12, 31)  // -000001-12-31
Time.plainDate(12345, 6, 7) // +012345-06-07
```

A date renders the same way in string interpolation and in JSON:

```ant
let d = Time.plainDate(2026, 10, 22)
"due ${d}"        // "due 2026-10-22"
Json.stringify(d) // "\"2026-10-22\""
```

Build several dates with a lambda, such as the first day of each quarter:

```ant
[1, 4, 7, 10].map({ Time.plainDate(2026, x, 1) })
// [2026-01-01, 2026-04-01, 2026-07-01, 2026-10-01]
```

A date is a value, so it works as a map key and compares with `==`:

```ant
let holidays = Map{
  Time.plainDate(2026, 12, 25): "Christmas",
  Time.plainDate(2026, 1, 1): "New Year",
}
holidays[Time.plainDate(2026, 12, 25)]                   // "Christmas"
Time.plainDate(2026, 1, 1) == Time.plainDate(2026, 1, 1) // true
```

## Applications

Check a year, a month and a day from the world before building the date, so a bad one gives `null`
instead of a panic:

```ant
func dateOf(year: int, month: int, day: int): Time.PlainDate | null {
  if month < 1 || month > 12 { return null }
  if day < 1 || day > Time.plainDate(year, month, 1).daysInMonth() { return null }
  Time.plainDate(year, month, day)
}

dateOf(2026, 2, 29) // null
dateOf(2028, 2, 29) // 2028-02-29
dateOf(2026, 13, 1) // null
```

The day of the week a date falls on, 1 for Monday to 7 for Sunday:

```ant
Time.plainDate(2026, 10, 22).dayOfWeek() // 4
```

A birthday in a given year. [Time.with](with.md) moves February 29 to February 28 in a year that has
no leap day, where `Time.plainDate` would panic:

```ant
let born = Time.plainDate(2000, 2, 29)
born.with(year: 2026) // 2026-02-28
born.with(year: 2028) // 2028-02-29
```

Days from today to a deadline:

```ant
let today = Time.plainDate(2026, 11, 11)
let deadline = Time.plainDate(2026, 12, 24)
today.until(deadline) // P43D
```

Sort dates with [Time.compare](compare.md):

```ant
let dates = [Time.plainDate(2026, 3, 1), Time.plainDate(2025, 12, 31), Time.plainDate(2026, 1, 15)]
dates.sortWith(Time.compare) // [2025-12-31, 2026-01-15, 2026-03-01]
```

## Properties and relations

A date reads back the year, the month and the day it was made with:

```ant
let d = Time.plainDate(1969, 7, 20)
assert(d.year() == 1969 && d.month() == 7 && d.day() == 20)
```

The text of a date parses back to the same date with [Time.parsePlainDate](parsePlainDate.md):

```ant
for d in [Time.plainDate(2026, 10, 22), Time.plainDate(-44, 3, 15), Time.plainDate(12345, 1, 1)] {
  assert(Time.parsePlainDate("${d}").unwrap() == d)
}
```

`Time.with` on any date builds the same value as `Time.plainDate` with those fields, when the date
exists:

```ant
let d = Time.plainDate(2026, 1, 31)
assert(d.with(month: 3, day: 15) == Time.plainDate(2026, 3, 15))
```

## Possible issues

A date that does not exist panics, unlike [Time.with](with.md), which moves the day to the end of
the month. A century year is a leap year only when divisible by 400, so 1900 has no February 29:

```ant
Time.plainDate(2026, 2, 30) // panic: Time.plainDate: no such date: 2026-02-30
Time.plainDate(2026, 13, 1) // panic: Time.plainDate: no such date: 2026-13-01
Time.plainDate(2025, 2, 29) // panic: Time.plainDate: no such date: 2025-02-29
Time.plainDate(1900, 2, 29) // panic: Time.plainDate: no such date: 1900-02-29
```

The arguments are year, month, day. A date written day first, as much of the world writes it, panics
when the day is above 12 and silently gives the wrong date otherwise: 3 April written day first
becomes March 4.

```ant
Time.plainDate(2026, 22, 10) // panic: Time.plainDate: no such date: 2026-22-10
Time.plainDate(2026, 3, 4)   // 2026-03-04
```

For a date that arrives as text, use [Time.parsePlainDate](parsePlainDate.md), which returns an
`Err` instead of panicking:

```ant
Time.parsePlainDate("2026-02-30").isOk() // false
```

Dates do not order with `<`. Use [Time.compare](compare.md):

```ant
let a = Time.plainDate(2026, 1, 1)
let b = Time.plainDate(2025, 12, 31)
println(a < b)
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.PlainDate; narrow first
```

```ant
Time.compare(Time.plainDate(2026, 1, 1), Time.plainDate(2025, 12, 31)) // 1
```

## See also

[Time.parsePlainDate](parsePlainDate.md) · [Time.daysInMonth](daysInMonth.md) ·
[Time.plainTime](plainTime.md) · [Time.plainDateTime](plainDateTime.md) · [Time.with](with.md) ·
[Time.year](year.md) · [Time.month](month.md) · [Time.day](day.md) ·
[Time.PlainDate](types/PlainDate.md) · [Time](index.md)
