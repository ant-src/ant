<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.daysInYear

```ant signature
func Time.daysInYear(t: Time.Dated): int
```

Time.daysInYear returns the number of days in the year of t, 365 or 366.

<!-- end generated -->

## Details

- `Time.daysInYear` takes any value with a date: a `PlainDate`, a `PlainDateTime`, or a
  `ZonedDateTime`. An `Instant` has no date until you give it a zone with `Time.inZone`.
- The year of a `ZonedDateTime` is the year in its zone, so one instant can fall in years of
  different lengths in two zones.

## Basic examples

```ant
Time.daysInYear(Time.plainDate(2024, 3, 1)) // 366
Time.daysInYear(Time.plainDate(2025, 3, 1)) // 365
```

Call it through `.` on a date:

```ant
let release = Time.plainDate(2028, 7, 4)
release.daysInYear() // 366
```

## Scope

The answer depends on the year alone, not on the month or day:

```ant
Time.plainDate(2024, 1, 1).daysInYear()   // 366
Time.plainDate(2024, 12, 31).daysInYear() // 366
```

Of the century years shown, only 2000 has 366 days:

```ant
Time.plainDate(1900, 6, 1).daysInYear() // 365
Time.plainDate(2000, 6, 1).daysInYear() // 366
Time.plainDate(2100, 6, 1).daysInYear() // 365
```

Years before 1 work too:

```ant
Time.plainDate(0, 6, 1).daysInYear()  // 366
Time.plainDate(-1, 6, 1).daysInYear() // 365
Time.plainDate(-4, 6, 1).daysInYear() // 366
```

A `PlainDateTime` and a `ZonedDateTime` answer for their date:

```ant
let eve = Time.plainDateTime(Time.plainDate(2024, 12, 31), Time.plainTime(hour: 23))
let stockholm = Result.unwrap(Time.zone("Europe/Stockholm"))
eve.daysInYear()                    // 366
eve.toZoned(stockholm).daysInYear() // 366
```

A value typed as the union `Time.Dated` passes as it is, with no `is` test:

```ant
func yearLength(t: Time.Dated): string {
  "${t.year()}: ${t.daysInYear()} days"
}

println(yearLength(Time.plainDate(2023, 5, 1)))
println(yearLength(Time.plainDateTime(Time.plainDate(2024, 5, 1), Time.plainTime())))
```

```text
2023: 365 days
2024: 366 days
```

## Applications

How far through the year a date is, as a fraction:

```ant
func yearProgress(d: Time.PlainDate): float64 {
  float64(d.dayOfYear()) / float64(d.daysInYear())
}

println(yearProgress(Time.plainDate(2024, 7, 1)))
println(yearProgress(Time.plainDate(2025, 7, 2)))
```

```text
0.5
0.5013698630136987
```

Days left until the end of the year:

```ant
let today = Time.plainDate(2026, 11, 11)
println("${today.daysInYear() - today.dayOfYear()} days left in ${today.year()}")
```

```text
50 days left in 2026
```

Daily interest under an actual/actual convention, where a day is a share of the year it falls in:

```ant
func dailyInterest(balance: float64, rate: float64, on: Time.PlainDate): float64 {
  balance * rate / float64(on.daysInYear())
}

println(dailyInterest(36600.0, 0.05, Time.plainDate(2024, 3, 1)))
println(dailyInterest(36500.0, 0.05, Time.plainDate(2025, 3, 1)))
```

```text
5.0
5.0
```

Count the days in a span of whole years:

```ant
let days = Array.range(2001, 2101)
  .map({ Time.plainDate(x, 1, 1).daysInYear() })
  .sum()
println(days)
```

```text
36524
```

## Properties and relations

A year has 366 days exactly when `Time.inLeapYear` is `true`:

```ant
for year in Array.range(1896, 1905) {
  let d = Time.plainDate(year, 1, 1)
  assert((d.daysInYear() == 366) == d.inLeapYear())
}
```

The last day of the year is day `daysInYear` of it, and the months add up to the year:

```ant
for year in [1900, 2000, 2023, 2024] {
  let d = Time.plainDate(year, 1, 1)
  assert(Time.plainDate(year, 12, 31).dayOfYear() == d.daysInYear())
  let months = Array.range(1, 13).map({ Time.plainDate(year, x, 1).daysInMonth() }).sum()
  assert(months == d.daysInYear())
}
```

Adding `daysInYear` days to January 1 gives January 1 of the next year:

```ant
let jan1 = Time.plainDate(2024, 1, 1)
let next = jan1.add(Time.duration(days: jan1.daysInYear()))
assert(next == Time.plainDate(2025, 1, 1))
```

## Possible issues

An `Instant` has no year, so passing one is a compile error:

```ant
let now = Result.unwrap(Time.parseInstant("2024-12-31T23:30:00Z"))
println(Time.daysInYear(now)) // error: cannot use Time.Instant where Time.Dated is expected
```

Give it a zone first. The zone decides the year, and with it the answer:

```ant
let now = Result.unwrap(Time.parseInstant("2024-12-31T23:30:00Z"))
let utc = Time.inZone(now, Time.utc)
let tokyo = Time.inZone(now, Result.unwrap(Time.zone("Asia/Tokyo")))
println("${utc.year()}: ${utc.daysInYear()}")
println("${tokyo.year()}: ${tokyo.daysInYear()}")
```

```text
2024: 366
2025: 365
```

A `PlainTime` has no date either:

```ant
Time.daysInYear(Time.plainTime(hour: 9)) // error: cannot use Time.PlainTime where Time.Dated is expected
```

## See also

[Time.inLeapYear](inLeapYear.md) · [Time.daysInMonth](daysInMonth.md) ·
[Time.dayOfYear](dayOfYear.md) · [Time.year](year.md) · [Time.inZone](inZone.md) · [Time](index.md)
