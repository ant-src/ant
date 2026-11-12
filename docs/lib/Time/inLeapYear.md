<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.inLeapYear

```ant signature
func Time.inLeapYear(t: Time.Dated): bool
```

Time.inLeapYear reports whether the year of t has 366 days: a year divisible by 4, except a century
year not divisible by 400, so 2000 and 0 are leap years and 1900 is not.

<!-- end generated -->

## Details

- Only the year of `t` matters: the month, the day, and the time of day do not change the result.
- A `Time.ZonedDateTime` answers for the year of its wall-clock date in its zone, so one instant can
  be in a leap year in one zone and not in another.

## Basic examples

2024 is a leap year and 2026 is not:

```ant
Time.inLeapYear(Time.plainDate(2024, 6, 1)) // true
Time.inLeapYear(Time.plainDate(2026, 6, 1)) // false
```

Call it through `.` on any date:

```ant
let launch = Time.plainDate(2028, 3, 15)
launch.inLeapYear() // true
```

## Scope

Century years are leap years only when they divide by 400:

```ant
[1900, 2000, 2100, 2400].map({ Time.plainDate(x, 1, 1).inLeapYear() }) // [false, true, false, true]
```

Year 0 and negative years follow the same calendar:

```ant
[0, -1, -4].map({ Time.plainDate(x, 1, 1).inLeapYear() }) // [true, false, true]
```

A `Time.PlainDateTime` answers for the year of its date:

```ant
let meeting = Time.plainDateTime(Time.plainDate(2032, 11, 2), Time.plainTime(hour: 10))
meeting.inLeapYear() // true
```

A `Time.ZonedDateTime` answers for the year its zone shows, so one instant gives two answers:

```ant
let at = Result.unwrap(Time.parseInstant("2024-12-31T23:30:00Z"))
at.inZone(Time.utc).inLeapYear()                               // true
at.inZone(Result.unwrap(Time.zone("Asia/Tokyo"))).inLeapYear() // false
```

A value typed with the union `Time.Dated` passes as it is, so the function maps over a mixed array:

```ant
let dates: Array[Time.Dated] = [
  Time.plainDate(2026, 2, 1),
  Time.plainDateTime(Time.plainDate(2028, 6, 1), Time.plainTime()),
]
dates.map(Time.inLeapYear) // [false, true]
```

In a pipeline, the date is the first argument:

```ant
Time.parsePlainDate("2028-03-01") | Result.map(Time.inLeapYear) // Ok{ value: true }
```

## Applications

The leap years of a range:

```ant
[2024, 2025, 2026, 2027, 2028, 2029, 2030, 2031, 2032].filter({
  Time.plainDate(x, 1, 1).inLeapYear()
}) // [2024, 2028, 2032]
```

The next leap year from a date:

```ant
func nextLeapYear(from: Time.PlainDate): int {
  var year = from.year() + 1
  while !Time.plainDate(year, 1, 1).inLeapYear() {
    year += 1
  }
  year
}

println(nextLeapYear(Time.plainDate(2026, 11, 11)))
println(nextLeapYear(Time.plainDate(2096, 1, 1)))
```

```text
2028
2104
```

Celebrate a February 29 birthday on March 1 in a common year:

```ant
func birthdayIn(year: int, born: Time.PlainDate): Time.PlainDate {
  let leapDay = born.month() == 2 && born.day() == 29
  if leapDay && !Time.plainDate(year, 1, 1).inLeapYear() {
    return Time.plainDate(year, 3, 1)
  }
  born.with(year: year)
}

let born = Time.plainDate(2000, 2, 29)
println(birthdayIn(2026, born))
println(birthdayIn(2028, born))
```

```text
2026-03-01
2028-02-29
```

## Properties and relations

A leap year has 366 days, the count `Time.daysInYear` gives:

```ant
for year in [1900, 2000, 2023, 2024] {
  let d = Time.plainDate(year, 1, 1)
  assert(d.inLeapYear() == (d.daysInYear() == 366))
}
```

February has 29 days exactly in a leap year:

```ant
for year in [1900, 2000, 2023, 2024] {
  let feb = Time.plainDate(year, 2, 1)
  assert(feb.inLeapYear() == (feb.daysInMonth() == 29))
}
```

It matches the Gregorian rule: divisible by 4, and not by 100 unless by 400:

```ant
for year in [-400, -100, -1, 0, 4, 1900, 2000, 2023, 2024, 2100] {
  let rule = year % 4 == 0 && (year % 100 != 0 || year % 400 == 0)
  assert(Time.plainDate(year, 1, 1).inLeapYear() == rule)
}
```

## Possible issues

The function takes a date, not a year:

```ant
println(Time.inLeapYear(2024)) // error: cannot use int where Time.Dated is expected
```

Build a date in the year instead:

```ant
Time.plainDate(2024, 1, 1).inLeapYear() // true
```

A `Time.Instant` has no year until it is in a zone:

```ant
let epoch = Time.fromEpochMillis(0)
println(Time.inLeapYear(epoch)) // error: cannot use Time.Instant where Time.Dated is expected
```

```ant
Time.fromEpochMillis(0).inZone(Time.utc).inLeapYear() // false
```

`Time.plainDate` panics on February 29 of a common year, so test the year before building that date:

```ant
Time.plainDate(2026, 2, 29) // panic: Time.plainDate: no such date: 2026-02-29
```

## See also

[Time.daysInYear](daysInYear.md) · [Time.daysInMonth](daysInMonth.md) ·
[Time.dayOfYear](dayOfYear.md) · [Time.year](year.md) · [Time.plainDate](plainDate.md) ·
[Time.with](with.md)
