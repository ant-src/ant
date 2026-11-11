<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.yearOfWeek

```ant signature
func Time.yearOfWeek(t: Time.Dated): int
```

Time.yearOfWeek returns the year that the ISO 8601 week of t belongs to, which differs from
[`Time.year`](year.md) around New Year.

<!-- end generated -->

## Details

- The result equals `Time.year` except for up to three days at the start of January, which can
  belong to the previous year, and up to three days at the end of December, which can belong to the
  next.
- The week's year is the year of its Thursday.
- A `Time.ZonedDateTime` counts the date its zone shows, not the date in UTC.

## Basic examples

January 1, 2027 is a Friday, so it belongs to the last week of 2026:

```ant
Time.yearOfWeek(Time.plainDate(2027, 1, 1)) // 2026
```

December 30, 2024 is a Monday, and its week holds January 2, 2025, the first Thursday of 2025:

```ant
Time.yearOfWeek(Time.plainDate(2024, 12, 30)) // 2025
Time.weekOfYear(Time.plainDate(2024, 12, 30)) // 1
```

Away from New Year, it is the calendar year:

```ant
Time.yearOfWeek(Time.plainDate(2026, 6, 15)) // 2026
```

## Scope

A `PlainDateTime` counts its date and ignores the time:

```ant
let party = Time.plainDateTime(Time.plainDate(2027, 1, 1), Time.plainTime(hour: 20))
party.yearOfWeek() // 2026
```

A `ZonedDateTime` counts the date of its own zone. Sunday evening in New York is already Monday in
UTC, the first day of the first week of 2025:

```ant
let ny = Time.zone("America/New_York").unwrap()
let eve = Time.plainDateTime(Time.plainDate(2024, 12, 29), Time.plainTime(hour: 22)).toZoned(ny)
eve.yearOfWeek()                    // 2024
eve.withZone(Time.utc).yearOfWeek() // 2025
```

An instant reaches it once it has a zone:

```ant
Time.fromEpochMillis(0) | Time.inZone(Time.utc) | Time.yearOfWeek() // 1970
```

Pass it as a function value to map a parsed date:

```ant
Time.parsePlainDate("2021-01-03") | Result.map(Time.yearOfWeek) // Ok{ value: 2020 }
```

Years before the common era follow the same rule. January 1 of year 0 is a Saturday:

```ant
Time.plainDate(0, 1, 1).yearOfWeek() // -1
```

## Applications

Format a date as an ISO 8601 week date, such as `2026-W53-5`:

```ant
func weekDate(d: Time.PlainDate): string {
  let week = "${d.weekOfYear()}".padStart(2, "0")
  "${d.yearOfWeek()}-W${week}-${d.dayOfWeek()}"
}

weekDate(Time.plainDate(2027, 1, 1))   // "2026-W53-5"
weekDate(Time.plainDate(2024, 12, 30)) // "2025-W01-1"
weekDate(Time.plainDate(2026, 11, 11)) // "2026-W46-3"
```

Keep the days that belong to the weeks of one year, the way a weekly report counts them. December
31, 2025 is a Wednesday, so its week, which holds January 1, 2026, is week 1 of 2026:

```ant
let days = [
  Time.plainDate(2025, 12, 31),
  Time.plainDate(2026, 1, 1),
  Time.plainDate(2026, 12, 31),
  Time.plainDate(2027, 1, 3),
  Time.plainDate(2027, 1, 4),
]
let report = days | Array.filter({ x.yearOfWeek() == 2026 })
for d in report {
  println(d)
}
```

```text
2025-12-31
2026-01-01
2026-12-31
2027-01-03
```

Find the years whose first days belong to the previous year's last week:

```ant
Array.range(2020, 2030)
  | Array.filter({ Time.plainDate(x, 1, 1).yearOfWeek() < x }) // [2021, 2022, 2023, 2027, 2028]
```

## Properties and relations

The week's year is the year of the week's Thursday, and it never differs from the calendar year by
more than one:

```ant
let start = Time.plainDate(2024, 12, 20)
for i in Array.range(0, 800) {
  let d = start.add(Time.duration(days: i))
  let thursday = d.add(Time.duration(days: 4 - d.dayOfWeek()))
  assert(d.yearOfWeek() == thursday.year())
  assert(Math.abs(d.yearOfWeek() - d.year()) <= 1)
}
```

January 4 and December 28 always belong to their own year's weeks:

```ant
for year in Array.range(2020, 2031) {
  assert(Time.plainDate(year, 1, 4).yearOfWeek() == year)
  assert(Time.plainDate(year, 12, 28).yearOfWeek() == year)
}
```

## Possible issues

Pairing the ISO week with `Time.year` mislabels the days around New Year:

```ant
let d = Time.plainDate(2024, 12, 30)
"${d.year()}-W${d.weekOfYear()}"       // "2024-W1"
"${d.yearOfWeek()}-W${d.weekOfYear()}" // "2025-W1"
```

An `Instant` has no date, so it does not compile. Put it in a zone first:

```ant
println(Time.yearOfWeek(Time.fromEpochMillis(0)))
// error: cannot use Time.Instant where Time.Dated is expected
```

## See also

[Time.weekOfYear](weekOfYear.md) · [Time.year](year.md) · [Time.dayOfWeek](dayOfWeek.md) ·
[Time.add](add.md) · [Time.inZone](inZone.md) · [Time.PlainDate](types/PlainDate.md)
