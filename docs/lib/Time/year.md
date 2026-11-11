<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.year

```ant signature
func Time.year(t: Time.Dated): int
```

Time.year returns the year of t.

<!-- end generated -->

## Details

- The calendar is ISO 8601's, which has a year 0 and negative years before it.
- On a `Time.ZonedDateTime`, the year is the one the wall clock in its zone shows. The same instant
  can fall in different years in different zones.
- A `Time.Instant` has no calendar and so no year. Put it in a zone with [Time.inZone](inZone.md)
  first.

## Basic examples

```ant
Time.year(Time.plainDate(2026, 10, 22)) // 2026
Time.year(Time.plainDate(1969, 7, 20))  // 1969
```

Call it through `.` on any value with a date:

```ant
let date = Time.plainDate(2000, 1, 1)
date.year() // 2000
```

## Scope

A `Time.PlainDateTime` has the year of its date, whatever the time:

```ant
let late = Time.plainDateTime(Time.plainDate(2026, 12, 31), Time.plainTime(23, minute: 59))
late.year() // 2026
```

A `Time.ZonedDateTime` has the year its zone's wall clock shows:

```ant
let tokyo = Time.parseZoned("2027-01-01T00:30:00+09:00[Asia/Tokyo]").unwrap()
tokyo.year() // 2027
```

A value typed `Time.Dated` may hold any of the three types with a date, and `Time.year` takes it
without narrowing. Pass `Time.year` by name to read each one:

```ant
let meeting = Time.plainDateTime(Time.plainDate(2025, 6, 3), Time.plainTime(9))
let launch = Time.parseZoned("2027-09-15T12:00:00Z[UTC]").unwrap()
let dates: Array[Time.Dated] = [Time.plainDate(2026, 2, 14), meeting, launch]
dates.map(Time.year) // [2026, 2025, 2027]
```

Year 0 and negative years are ordinary years. A date outside 0 to 9999 renders with a sign and six
digits:

```ant
let ides = Time.plainDate(-44, 3, 15)
ides                                                // -000044-03-15
ides.year()                                         // -44
Time.plainDate(0, 1, 1).year()                      // 0
Time.plainDate(12345, 1, 1).year()                  // 12345
Time.parsePlainDate("-000044-03-15").map(Time.year) // Ok{ value: -44 }
```

## Applications

Read the current year. [Time.now](now.md) returns a `Time.Instant`, which has no year, so ask for
the time in a zone with [Time.nowIn](nowIn.md):

```ant
let thisYear = Time.nowIn(Time.localZone()).year()
assert(thisYear >= 2026)
```

The local zone is the host's, so around New Year two machines can disagree. Pass `Time.utc` instead
when the result must not depend on the host.

Group dates by year:

```ant
let invoices = [
  Time.plainDate(2025, 3, 1),
  Time.plainDate(2026, 1, 9),
  Time.plainDate(2025, 12, 31),
]
println(invoices.groupBy(Time.year))
```

```text
Map{ 2025: [2025-03-01, 2025-12-31], 2026: [2026-01-09] }
```

Keep the dates of one year:

```ant
let dates = [
  Time.plainDate(2025, 12, 31),
  Time.plainDate(2026, 1, 1),
  Time.plainDate(2026, 7, 4),
  Time.plainDate(2027, 1, 1),
]
dates.filter({ x.year() == 2026 }) // [2026-01-01, 2026-07-04]
```

The decade a date falls in:

```ant
func decade(d: Time.Dated): int {
  d.year() / 10 * 10
}

decade(Time.plainDate(1969, 7, 20)) // 1960
decade(Time.plainDate(2026, 1, 1))  // 2020
```

## Properties and relations

`Time.year` returns the year a date was made with, and setting the year with `Time.with` sets what
`Time.year` reads:

```ant
let d = Time.plainDate(2026, 7, 14)
assert(Time.plainDate(1999, 12, 31).year() == 1999)
assert(d.with(year: 2030).year() == 2030)
```

Adding a duration of years moves the year by that many:

```ant
let d = Time.plainDate(2026, 5, 9)
assert(d.add(Time.duration(years: 3)).year() == d.year() + 3)
assert(d.subtract(Time.duration(years: 30)).year() == 1996)
```

A `Time.PlainDateTime` and a `Time.ZonedDateTime` have the year of their date:

```ant
let zoned = Time.parseZoned("2026-12-31T23:30:00+01:00[Europe/Stockholm]").unwrap()
let local = zoned.toPlainDateTime()
assert(zoned.year() == zoned.toPlainDate().year())
assert(local.year() == local.toPlainDate().year())
```

## Possible issues

A `Time.Instant` has no year, so passing one is a compile error:

```ant
let epoch = Time.fromEpochMillis(0)
println(Time.year(epoch)) // error: cannot use Time.Instant where Time.Dated is expected
```

Put the instant in a zone first. The year depends on the zone: the same instant is 2026 in UTC and
2027 in Tokyo:

```ant
let instant = Time.parseInstant("2026-12-31T23:30:00Z").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
instant.inZone(Time.utc).year() // 2026
instant.inZone(tokyo).year()    // 2027
```

The difference of two years is not an age. It counts a year that has not passed yet; ask
[Time.until](until.md) for whole years instead:

```ant
let birth = Time.plainDate(1990, 6, 15)
let today = Time.plainDate(2026, 6, 14)
today.year() - birth.year()                                    // 36
birth.until(today, largestUnit: Time.Year{}).part(Time.Year{}) // 35
```

`Time.year` is the calendar year, not the year of the ISO week. Around New Year they differ; use
[Time.yearOfWeek](yearOfWeek.md) with [Time.weekOfYear](weekOfYear.md):

```ant
let d = Time.plainDate(2024, 12, 30)
d.year()       // 2024
d.yearOfWeek() // 2025
```

Setting the year of February 29 to a year that is not a leap year constrains the day to February 28:

```ant
Time.plainDate(2024, 2, 29).with(year: 2025) // 2025-02-28
```

## See also

[Time.month](month.md) · [Time.day](day.md) · [Time.yearOfWeek](yearOfWeek.md) ·
[Time.daysInYear](daysInYear.md) · [Time.inLeapYear](inLeapYear.md) · [Time.with](with.md) ·
[Time.until](until.md) · [Time.plainDate](plainDate.md) · [Time](index.md)
