<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.PlainDate

```ant signature
type Time.PlainDate
```

Time.PlainDate is a calendar date with no time and no zone, such as a birthday. It renders as
2026-10-22.

<!-- end generated -->

## Details

- A `Time.PlainDate` is immutable and always a real date. Functions that change it, such as
  [Time.add](../add.md) and [Time.with](../with.md), return a new value.
- Two dates are `==` when their year, month, and day are. The comparison operators such as `<` do
  not apply; [Time.compare](../compare.md) orders dates.
- [Time.plainDate](../plainDate.md) panics on a date that does not exist, and
  [Time.parsePlainDate](../parsePlainDate.md) returns an `Err` for it. Use the parser for text from
  outside the program.
- A duration's hours and smaller units count as whole 24-hour days on a date, and the rest is
  dropped.

## Basic examples

Build a date from its fields, or parse one from text:

```ant
Time.plainDate(2026, 10, 22)               // 2026-10-22
Time.parsePlainDate("2026-10-22")          // Ok{ value: 2026-10-22 }
Time.parsePlainDate("2026-10-22").unwrap() // 2026-10-22
```

Take the date a `Time.ZonedDateTime` shows in its zone. The same moment can be a different date in
another zone:

```ant
let late = Time.parseZoned("2026-10-22T23:30:00+02:00[Europe/Stockholm]").unwrap()
late.toPlainDate()                    // 2026-10-22
late.withZone(Time.utc).toPlainDate() // 2026-10-22
let tokyo = Time.zone("Asia/Tokyo").unwrap()
late.withZone(tokyo).toPlainDate() // 2026-10-23
```

Today's date depends on the zone you ask in, so take the current time in a zone and then its date:

```ant
let today = Time.nowIn(Time.localZone()).toPlainDate()
println(today.year() >= 2026)
```

```text
true
```

## Scope

### Fields

The date fields read a `Time.PlainDate` through `.`:

```ant
let d = Time.plainDate(2026, 10, 22)
d.year()        // 2026
d.month()       // 10
d.day()         // 22
d.dayOfWeek()   // 4
d.dayOfYear()   // 295
d.weekOfYear()  // 43
d.daysInMonth() // 31
d.daysInYear()  // 365
d.inLeapYear()  // false
```

### Changing a date

`Time.with` replaces fields and constrains the day to the month, and `Time.add` and `Time.subtract`
move by a duration:

```ant
let d = Time.plainDate(2026, 10, 22)
d.with(day: 1)                           // 2026-10-01
d.with(month: 2, day: 31)                // 2026-02-28
d.add(Time.duration(weeks: 1))           // 2026-10-29
d.subtract(Time.duration(years: 1))      // 2025-10-22
d.add(Time.duration(days: 1, hours: 30)) // 2026-10-24
```

### Measuring between dates

`Time.until` counts days by default, and larger units on request:

```ant
let start = Time.plainDate(2026, 1, 1)
let end = Time.plainDate(2026, 10, 22)
start.until(end)                            // P294D
start.until(end, largestUnit: Time.Month{}) // P9M21D
end.until(start)                            // -P294D
Time.compare(start, end)                    // -1
```

### Text and JSON

Interpolation, `String.from`, and `Json.stringify` write the ISO 8601 form. `Time.parsePlainDate`
also takes the date out of a longer string with a time and zone:

```ant
let d = Time.plainDate(2026, 10, 22)
"due ${d}"                                                         // "due 2026-10-22"
String.from(d)                                                     // "2026-10-22"
Json.stringify([d])                                                // "[\"2026-10-22\"]"
Time.parsePlainDate("2026-10-22T14:03:05+02:00[Europe/Stockholm]") // Ok{ value: 2026-10-22 }
Time.plainDate(-44, 3, 15)                                         // -000044-03-15
Time.plainDate(12345, 1, 1)                                        // +012345-01-01
```

### Records, maps, and unions

A date is a field like any other, and a map key:

```ant
type Person = { name: string, born: Time.PlainDate }
let ann = Person{ name: "Ann", born: Time.plainDate(1990, 7, 14) }
ann                 // Person{ name: "Ann", born: 1990-07-14 }
Json.stringify(ann) // "{\"name\":\"Ann\",\"born\":\"1990-07-14\"}"
let holidays = Map{ Time.plainDate(2026, 12, 25): "Christmas" }
holidays[Time.plainDate(2026, 12, 25)] // "Christmas"
```

`is` narrows a union of time types to `Time.PlainDate`:

```ant
func describe(t: Time.Dated): string {
  if t is Time.PlainDate { "the whole day ${t}" } else { "a moment on ${t.toPlainDate()}" }
}

println(describe(Time.plainDate(2026, 10, 22)))
println(describe(Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(9))))
```

```text
the whole day 2026-10-22
a moment on 2026-10-22
```

### Adding a time

`Time.plainDateTime` joins a date and a time, and `Time.toZoned` then places it in a zone:

```ant
let d = Time.plainDate(2026, 10, 22)
let meeting = Time.plainDateTime(d, Time.plainTime(9, 30))
meeting                   // 2026-10-22T09:30:00
meeting.toZoned(Time.utc) // 2026-10-22T09:30:00+00:00[UTC]
```

## Applications

A person's age in whole years on a given day:

```ant
func age(born: Time.PlainDate, on: Time.PlainDate): int {
  born.until(on, largestUnit: Time.Year{}).part(Time.Year{})
}

let born = Time.plainDate(1990, 7, 14)
age(born, Time.plainDate(2026, 7, 13)) // 35
age(born, Time.plainDate(2026, 7, 14)) // 36
```

The last day of a month:

```ant
let d = Time.plainDate(2024, 2, 10)
d.with(day: d.daysInMonth()) // 2024-02-29
```

The working days of two weeks, Monday to Friday:

```ant
let start = Time.plainDate(2026, 12, 21)
let workdays = Iter.range(0, 14)
  | Iter.map({ start.add(Time.duration(days: x)) })
  | Iter.filter({ x.dayOfWeek() <= 5 })
  | Iter.toArray()
println(workdays.length())
println(workdays[workdays.length() - 1])
```

```text
10
2027-01-01
```

Sort dates with `Time.compare`, and group them by month, keyed by the first day of each month:

```ant
let dates = [Time.plainDate(2026, 3, 1), Time.plainDate(2025, 12, 24), Time.plainDate(2026, 3, 9)]
dates.sortWith(Time.compare) // [2025-12-24, 2026-03-01, 2026-03-09]
let byMonth = dates.groupBy({ x.with(day: 1) })
byMonth[Time.plainDate(2026, 3, 1)] // [2026-03-01, 2026-03-09]
byMonth.keys()                      // [2026-03-01, 2025-12-01]
```

## Properties and relations

A date parses back from the text it renders as:

```ant
let d = Time.plainDate(2026, 10, 22)
assert(Time.parsePlainDate(String.from(d)).unwrap() == d)
let bc = Time.plainDate(-44, 3, 15)
assert(Time.parsePlainDate("${bc}").unwrap() == bc)
```

Adding the days `Time.until` counts gets back to the later date:

```ant
let a = Time.plainDate(2024, 2, 10)
let b = Time.plainDate(2026, 10, 22)
assert(a.add(a.until(b)) == b)
```

`Time.compare` is antisymmetric:

```ant
let a = Time.plainDate(2024, 2, 10)
let b = Time.plainDate(2026, 10, 22)
assert(Time.compare(a, b) == -Time.compare(b, a))
```

A date joined with a time and taken apart again is the same date:

```ant
let d = Time.plainDate(2026, 10, 22)
assert(Time.plainDateTime(d, Time.plainTime(23, 59)).toPlainDate() == d)
```

## Possible issues

A date that does not exist panics in `Time.plainDate`. Text from outside goes through
`Time.parsePlainDate`, which returns an `Err` instead:

```ant
Time.plainDate(2026, 2, 29)       // panic: Time.plainDate: no such date: 2026-02-29
Time.parsePlainDate("2026-02-29") // Err{ error: Error{ message: "no such date: 2026-02-29" } }
```

A string that ends in `Z` with no zone in brackets names a moment, not a wall-clock date, so the
parser refuses it:

```ant
Time.parsePlainDate("2026-10-22T14:03:05Z") // Err{ error: Error{ message: "Z names an exact time; a plain value needs a wall-clock time" } }
```

The ISO week number belongs to the week's year, which differs from the date's year around New Year.
Pair `weekOfYear` with `yearOfWeek`, not `year`:

```ant
let ny = Time.plainDate(2027, 1, 1)
ny.weekOfYear() // 53
ny.year()       // 2027
ny.yearOfWeek() // 2026
```

`<` does not compare dates, and `Array.sort` does not sort them. Use `Time.compare`:

```ant
let a = Time.plainDate(2026, 3, 1)
let b = Time.plainDate(2026, 4, 1)
let early = a < b
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.PlainDate
let sorted = [b, a].sort()
// error: T would be Time.PlainDate, which is not a member of its bound Ordered
```

A date has no time of day. The time fields and `Time.round` do not take it, and `Time.toZoned` needs
a `Time.PlainDateTime`:

```ant
let d = Time.plainDate(2026, 3, 1)
let h = Time.hour(d) // error: cannot use Time.PlainDate where Time.Timed is expected
let r = Time.round(d, Time.Day{})
// error: T would be Time.PlainDate, which is not a member of its bound Time.Instant | Time.Timed
let z = Time.toZoned(d, Time.utc)
// error: cannot use Time.PlainDate where Time.PlainDateTime is expected
```

`Time.with` and `Time.until` panic on a time field or unit:

```ant
let d = Time.plainDate(2026, 3, 1)
let april = Time.plainDate(2026, 4, 1)
d.with(hour: 3)                          // panic: Time.with: a date has no time fields
d.until(april, largestUnit: Time.Hour{}) // panic: Time.until: a date has no hours
```

Hours shorter than a day are dropped, so small steps add up to nothing:

```ant
let d = Time.plainDate(2026, 1, 1)
d.add(12h).add(12h) // 2026-01-01
d.add(24h)          // 2026-01-02
```

## See also

[Time.plainDate](../plainDate.md) · [Time.parsePlainDate](../parsePlainDate.md) ·
[Time.toPlainDate](../toPlainDate.md) · [Time.plainDateTime](../plainDateTime.md) ·
[Time.PlainDateTime](PlainDateTime.md) · [Time.PlainTime](PlainTime.md) ·
[Time.ZonedDateTime](ZonedDateTime.md) · [Time.Dated](Dated.md) · [Time.with](../with.md) ·
[Time.add](../add.md) · [Time.until](../until.md) · [Time.compare](../compare.md)
