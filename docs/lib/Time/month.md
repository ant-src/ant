<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.month

```ant signature
func Time.month(t: Time.Dated): int
```

Time.month returns the month of t, 1 for January to 12 for December.

<!-- end generated -->

## Details

- Months count from 1, not 0. Subtract 1 before you index an array of month names.
- On a `Time.ZonedDateTime`, the month is the one the wall clock in its zone shows. The same instant
  can fall in different months in different zones.
- A `Time.Instant` has no calendar and so no month. Put it in a zone with [Time.inZone](inZone.md)
  first.

## Basic examples

```ant
Time.month(Time.plainDate(2026, 10, 22)) // 10
Time.month(Time.plainDate(2026, 1, 1))   // 1
```

Call it through `.` on any value with a date:

```ant
let date = Time.plainDate(2026, 12, 25)
date.month() // 12
```

## Scope

A `Time.PlainDateTime` has the month of its date, whatever the time:

```ant
let late = Time.plainDateTime(Time.plainDate(2026, 4, 30), Time.plainTime(23, minute: 59))
late.month() // 4
```

A `Time.ZonedDateTime` has the month its zone's wall clock shows: 00:30 on March 1 in Stockholm is
still February 28 in UTC:

```ant
let stockholm = Time.parseZoned("2026-03-01T00:30:00+01:00[Europe/Stockholm]").unwrap()
stockholm.month()                    // 3
stockholm.withZone(Time.utc).month() // 2
```

A value typed `Time.Dated` may hold any of the three types with a date, and `Time.month` takes it
without narrowing. Pass `Time.month` by name to read each one:

```ant
let meeting = Time.plainDateTime(Time.plainDate(2026, 6, 3), Time.plainTime(9))
let launch = Time.parseZoned("2026-09-15T12:00:00Z[UTC]").unwrap()
let dates: Array[Time.Dated] = [Time.plainDate(2026, 2, 14), meeting, launch]
dates.map(Time.month) // [2, 6, 9]
```

In a pipe, the value is the first argument:

```ant
"2026-07-04" | Time.parsePlainDate() | Result.map(Time.month) // Ok{ value: 7 }
```

## Applications

Name the month. The month counts from 1, and the array from 0:

```ant
let names = [
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

func monthName(d: Time.Dated): string {
  names[d.month() - 1]
}

println(monthName(Time.plainDate(2026, 1, 5)))
println(monthName(Time.plainDate(2026, 10, 22)))
```

```text
January
October
```

The quarter of the year, 1 to 4:

```ant
func quarter(d: Time.Dated): int {
  (d.month() - 1) / 3 + 1
}

quarter(Time.plainDate(2026, 3, 31))  // 1
quarter(Time.plainDate(2026, 4, 1))   // 2
quarter(Time.plainDate(2026, 12, 24)) // 4
```

Group orders by the month they were placed in:

```ant
let orders = [
  Time.plainDate(2026, 1, 14),
  Time.plainDate(2026, 3, 2),
  Time.plainDate(2026, 3, 28),
  Time.plainDate(2026, 7, 9),
]
println(orders.groupBy(Time.month))
```

```text
Map{ 1: [2026-01-14], 3: [2026-03-02, 2026-03-28], 7: [2026-07-09] }
```

Keep the dates that fall in summer, June to August:

```ant
let dates = [
  Time.plainDate(2026, 5, 31),
  Time.plainDate(2026, 6, 1),
  Time.plainDate(2026, 8, 31),
  Time.plainDate(2026, 9, 1),
]
dates.filter({ x.month() >= 6 && x.month() <= 8 }) // [2026-06-01, 2026-08-31]
```

## Properties and relations

The month is between 1 and 12:

```ant
let dates = [Time.plainDate(2026, 1, 1), Time.plainDate(2024, 2, 29), Time.plainDate(2026, 12, 31)]
for d in dates {
  assert(d.month() >= 1)
  assert(d.month() <= 12)
}
```

`Time.month` returns the month a date was made with, and setting the month with `Time.with` sets
what `Time.month` reads:

```ant
let d = Time.plainDate(2026, 7, 14)
assert(d.month() == 7)
assert(d.with(month: 2).month() == 2)
```

Adding months wraps into the next year, and `Time.year` follows:

```ant
let d = Time.plainDate(2026, 11, 30).add(Time.duration(months: 3))
assert(d.month() == 2)
assert(d.year() == 2027)
```

## Possible issues

`Time.now` returns a `Time.Instant`, which has no month, so asking it for the month is a compile
error:

```ant
println(Time.now().month()) // error: Time.month takes Time.Dated first, so a Time.Instant does not reach it; call it as Time.month(...)
```

For the current month, take the time in a zone with `Time.nowIn`:

```ant
let month = Time.nowIn(Time.localZone()).month()
assert(month >= 1 && month <= 12)
```

Put the instant in a zone first. The month depends on the zone: the same instant is December in UTC
and January in Stockholm:

```ant
let instant = Time.parseInstant("2026-12-31T23:30:00Z").unwrap()
let stockholm = Time.zone("Europe/Stockholm").unwrap()
instant.inZone(Time.utc).month()  // 12
instant.inZone(stockholm).month() // 1
```

Months count from 1, so a month used as an index without subtracting 1 is off by one, and December
is past the end:

```ant
let short = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]
let d = Time.plainDate(2026, 12, 1)
short[d.month() - 1] // "Dec"
short[d.month()]     // panic: index 12 out of range for an array of length 12
```

`Time.with` constrains a month above 12, but a month below 1 panics:

```ant
let d = Time.plainDate(2026, 3, 31)
d.with(month: 13).month() // 12
d.with(month: 0)          // panic: Time.with: month below 1: 0
```

## See also

[Time.year](year.md) · [Time.day](day.md) · [Time.daysInMonth](daysInMonth.md) ·
[Time.with](with.md) · [Time.add](add.md) · [Time.plainDate](plainDate.md) · [Time](index.md)
