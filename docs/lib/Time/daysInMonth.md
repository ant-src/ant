<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.daysInMonth

```ant signature
func Time.daysInMonth(t: Time.Dated): int
```

Time.daysInMonth returns the number of days in the month of t, 28 to 31.

<!-- end generated -->

## Details

- Only the year and the month of `t` matter: the day and the time of day do not change the result.
- February has 29 days in a leap year, the year for which `Time.inLeapYear` holds, and 28 otherwise.
- A `Time.ZonedDateTime` gives the month of its wall-clock date in its zone.

## Basic examples

The number of days in February 2026:

```ant
Time.daysInMonth(Time.plainDate(2026, 2, 1)) // 28
```

February 2024, a leap year:

```ant
let d = Time.plainDate(2024, 2, 10)
d.daysInMonth() // 29
```

## Scope

Every month of a year:

```ant
Array.range(1, 13) | Array.map({ Time.plainDate(2026, x, 1).daysInMonth() }) // [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31]
```

A `Time.PlainDateTime` gives the days of its date's month:

```ant
let end = Time.plainDateTime(Time.plainDate(2026, 4, 30), Time.plainTime(hour: 23, minute: 59))
end.daysInMonth() // 30
```

A `Time.ZonedDateTime` counts the month of its wall-clock date, so one instant can fall in two
months:

```ant
let at = Result.unwrap(Time.parseInstant("2026-02-28T23:30:00Z"))
at.inZone(Time.utc).daysInMonth()                                     // 28
at.inZone(Result.unwrap(Time.zone("Europe/Stockholm"))).daysInMonth() // 31
```

A value typed with the union `Time.Dated` passes as it is, so the function maps over a mixed array:

```ant
let dates: Array[Time.Dated] = [
  Time.plainDate(2026, 2, 1),
  Time.plainDateTime(Time.plainDate(2026, 6, 1), Time.plainTime()),
]
dates.map(Time.daysInMonth) // [28, 30]
```

Years down to 0 and below work too. Year 0 is a leap year:

```ant
Time.plainDate(0, 2, 1).daysInMonth()  // 29
Time.plainDate(-1, 2, 1).daysInMonth() // 28
```

On parsed text, map over the Result:

```ant
Time.parsePlainDate("2026-11-11") | Result.map(Time.daysInMonth) // Ok{ value: 30 }
```

## Applications

Check a day from the world before building the date, since `Time.plainDate` panics on a day the
month lacks:

```ant
func isDate(year: int, month: int, day: int): bool {
  if month < 1 || month > 12 { return false }
  day >= 1 && day <= Time.plainDate(year, month, 1).daysInMonth()
}

isDate(2026, 2, 29) // false
isDate(2028, 2, 29) // true
isDate(2026, 13, 1) // false
```

The last day of the month:

```ant
let d = Time.plainDate(2026, 4, 15)
d.with(day: d.daysInMonth()) // 2026-04-30
```

The days left in the month, today included:

```ant
let today = Time.plainDate(2026, 11, 11)
today.daysInMonth() - today.day() + 1 // 20
```

Prorate a monthly fee in cents for a customer who starts mid-month:

```ant
func prorate(cents: int, start: Time.PlainDate): int {
  let days = start.daysInMonth()
  cents * (days - start.day() + 1) / days
}

println(prorate(3000, Time.plainDate(2026, 11, 16)))
println(prorate(3000, Time.plainDate(2026, 2, 15)))
```

```text
1500
1500
```

## Properties and relations

The days of the twelve months add up to `Time.daysInYear`:

```ant
for year in [1900, 2000, 2024, 2026] {
  let total = Array.range(1, 13)
    | Array.map({ Time.plainDate(year, x, 1).daysInMonth() })
    | Array.sum()
  assert(total == Time.plainDate(year, 1, 1).daysInYear())
}
```

February has 29 days exactly when `Time.inLeapYear` holds:

```ant
for year in [1900, 2000, 2023, 2024] {
  let feb = Time.plainDate(year, 2, 1)
  assert((feb.daysInMonth() == 29) == feb.inLeapYear())
}
```

`Time.with` constrains a day past the end of the month to the last day, the one `Time.daysInMonth`
names:

```ant
let d = Time.plainDate(2026, 4, 15)
assert(d.with(day: 31).day() == d.daysInMonth())
```

## Possible issues

A `Time.Instant` has no date until it is in a zone, and a `Time.PlainTime` has no date at all:

```ant
let at = Result.unwrap(Time.parseInstant("2026-02-28T23:30:00Z"))
println(Time.daysInMonth(at)) // error: cannot use Time.Instant where Time.Dated is expected
```

```ant
println(Time.daysInMonth(Time.plainTime(hour: 9))) // error: cannot use Time.PlainTime where Time.Dated is expected
```

Place the instant in a zone first:

```ant
let at = Result.unwrap(Time.parseInstant("2026-02-28T23:30:00Z"))
at.inZone(Time.utc).daysInMonth() // 28
```

The function takes a date, not a year and a month. Build the first day of the month to ask about it:

```ant
println(Time.daysInMonth(2026, 2)) // error: Time.daysInMonth takes 1 argument, got 2
```

## See also

[Time.daysInYear](daysInYear.md) · [Time.inLeapYear](inLeapYear.md) · [Time.day](day.md) ·
[Time.month](month.md) · [Time.with](with.md) · [Time.plainDate](plainDate.md) ·
[Time.parsePlainDate](parsePlainDate.md)
