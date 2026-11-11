<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.weekOfYear

```ant signature
func Time.weekOfYear(t: Time.Dated): int
```

Time.weekOfYear returns the ISO 8601 week of t, 1 to 53. Weeks start on Monday and week 1 holds the
year's first Thursday, so days around New Year can fall in a week of the neighbouring year;
[`Time.yearOfWeek`](yearOfWeek.md) gives that year.

<!-- end generated -->

## Details

- Up to three days at the start of January can fall in week 52 or 53 of the previous year, and up to
  three days at the end of December in week 1 of the next. Pair the week with `Time.yearOfWeek`, not
  `Time.year`.
- A `Time.ZonedDateTime` counts the date its zone shows, not the date in UTC.

## Basic examples

November 11, 2026 is in week 46:

```ant
Time.weekOfYear(Time.plainDate(2026, 11, 11)) // 46
```

January 1, 2027 is a Friday, so it belongs to the last week of 2026:

```ant
Time.weekOfYear(Time.plainDate(2027, 1, 1)) // 53
Time.yearOfWeek(Time.plainDate(2027, 1, 1)) // 2026
```

## Scope

A `PlainDateTime` counts its date and ignores the time:

```ant
let standup = Time.plainDateTime(Time.plainDate(2026, 11, 11), Time.plainTime(hour: 9))
standup.weekOfYear() // 46
```

A `ZonedDateTime` counts the date of its own zone. Sunday evening in New York is already Monday in
UTC, the start of the next week:

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let sunday = Time.plainDateTime(Time.plainDate(2026, 12, 27), Time.plainTime(hour: 23)).toZoned(ny)
sunday.weekOfYear()                    // 52
sunday.withZone(Time.utc).weekOfYear() // 53
```

An instant reaches it once it has a zone:

```ant
Time.fromEpochMillis(0) | Time.inZone(Time.utc) | Time.weekOfYear() // 1
```

Pass it as a function value to map a parsed date:

```ant
Time.parsePlainDate("2026-07-04") | Result.map(Time.weekOfYear) // Ok{ value: 27 }
```

## Named arguments

The parameter is named `t`:

```ant
Time.weekOfYear(t: Time.plainDate(2026, 11, 11)) // 46
```

## Applications

Label a date with its ISO week, as in `2026-W46`:

```ant
func isoWeek(d: Time.PlainDate): string {
  let week = "${d.weekOfYear()}".padStart(2, "0")
  "${d.yearOfWeek()}-W${week}"
}

isoWeek(Time.plainDate(2026, 11, 11)) // "2026-W46"
isoWeek(Time.plainDate(2027, 1, 1))   // "2026-W53"
isoWeek(Time.plainDate(2025, 12, 29)) // "2026-W01"
```

Count the ISO weeks of a year. December 28 is always in the last week:

```ant
func weeksIn(year: int): int {
  Time.plainDate(year, 12, 28).weekOfYear()
}

weeksIn(2026) // 53
weeksIn(2027) // 52
```

Group a log of dates into weeks:

```ant
let days = [
  Time.plainDate(2026, 12, 30),
  Time.plainDate(2027, 1, 3),
  Time.plainDate(2027, 1, 4),
  Time.plainDate(2027, 1, 8),
]
let byWeek = days | Array.groupBy({ "${x.yearOfWeek()}-W${"${x.weekOfYear()}".padStart(2, "0")}" })
for { key, value } in byWeek {
  println("${key}: ${value.length()} days")
}
```

```text
2026-W53: 2 days
2027-W01: 2 days
```

Find the Monday that starts week `n` of a year, from January 4, which is always in week 1:

```ant
func mondayOfWeek(year: int, n: int): Time.PlainDate {
  let jan4 = Time.plainDate(year, 1, 4)
  let week1 = jan4.subtract(Time.duration(days: jan4.dayOfWeek() - 1))
  week1.add(Time.duration(weeks: n - 1))
}

mondayOfWeek(2026, 1)  // 2025-12-29
mondayOfWeek(2026, 46) // 2026-11-09
```

## Properties and relations

January 4 is always in week 1, and December 28 is always in the last week of its own year:

```ant
for year in Array.range(2020, 2031) {
  assert(Time.plainDate(year, 1, 4).weekOfYear() == 1)
  assert(Time.plainDate(year, 12, 28).yearOfWeek() == year)
}
```

Monday to Sunday share a week, and the next Monday starts the next one:

```ant
let monday = Time.plainDate(2026, 11, 9)
for i in Array.range(0, 7) {
  assert(monday.add(Time.duration(days: i)).weekOfYear() == 46)
}
assert(monday.add(Time.duration(weeks: 1)).weekOfYear() == 47)
```

## Possible issues

Pairing the week with `Time.year` mislabels the days around New Year. Use `Time.yearOfWeek`:

```ant
let d = Time.plainDate(2025, 12, 29)
"${d.year()}-W${d.weekOfYear()}"       // "2025-W1"
"${d.yearOfWeek()}-W${d.weekOfYear()}" // "2026-W1"
```

The week is not the day of the year divided by 7. Weeks start on Monday, so January 1 can be in week
52 or 53:

```ant
let newYear = Time.plainDate(2027, 1, 1)
newYear.weekOfYear()              // 53
(newYear.dayOfYear() - 1) / 7 + 1 // 1
```

An `Instant` has no date, so it does not compile. Put it in a zone first:

```ant
println(Time.weekOfYear(Time.fromEpochMillis(0)))
// error: cannot use Time.Instant where Time.Dated is expected
```

## See also

[Time.yearOfWeek](yearOfWeek.md) · [Time.dayOfWeek](dayOfWeek.md) · [Time.dayOfYear](dayOfYear.md) ·
[Time.year](year.md) · [Time.add](add.md) · [Time.inZone](inZone.md) ·
[Time.PlainDate](types/PlainDate.md)
