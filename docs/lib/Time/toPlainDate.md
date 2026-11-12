<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.toPlainDate

```ant signature
func Time.toPlainDate(t: Time.ZonedDateTime | Time.PlainDateTime): Time.PlainDate
```

Time.toPlainDate returns the date of t, as its wall clock shows it.

<!-- end generated -->

## Details

- On a `Time.ZonedDateTime`, the date is the one the wall clock in its zone shows. The same instant
  can fall on a different date in another zone.
- The time of day, the offset and the zone are dropped. [Time.toPlainTime](toPlainTime.md) returns
  the time of day.

## Basic examples

```ant
let meeting = Time.parsePlainDateTime("2026-10-22T14:30").unwrap()
Time.toPlainDate(meeting) // 2026-10-22
```

```ant
let landing = Time.parseZoned("2026-10-22T23:45:00+09:00[Asia/Tokyo]").unwrap()
Time.toPlainDate(landing) // 2026-10-22
```

Call it through `.`:

```ant
let meeting = Time.plainDateTime(Time.plainDate(2026, 3, 1), Time.plainTime(9))
meeting.toPlainDate() // 2026-03-01
```

## Scope

The same instant falls on different dates in different zones:

```ant
let t = Time.parseInstant("2026-10-22T23:30:00Z").unwrap()
t.inZone(Time.utc).toPlainDate()                                  // 2026-10-22
t.inZone(Time.zone("Asia/Tokyo").unwrap()).toPlainDate()          // 2026-10-23
t.inZone(Time.zone("America/Los_Angeles").unwrap()).toPlainDate() // 2026-10-22
```

Midnight belongs to the day it starts, and the last nanosecond to the day it ends:

```ant
Time.parsePlainDateTime("2026-10-22T00:00").unwrap().toPlainDate()              // 2026-10-22
Time.parsePlainDateTime("2026-10-22T23:59:59.999999999").unwrap().toPlainDate() // 2026-10-22
```

A value typed as the union of the two types passes without narrowing, and the function passes by
name:

```ant
let alarm = Time.parsePlainDateTime("2026-12-24T07:00").unwrap()
let flight = Time.parseZoned("2026-12-25T10:15:00+01:00[Europe/Stockholm]").unwrap()
let times: Array[Time.ZonedDateTime | Time.PlainDateTime] = [alarm, flight]
times.map(Time.toPlainDate) // [2026-12-24, 2026-12-25]
```

In a pipe, the value is the first argument:

```ant
"2026-10-22T14:30" | Time.parsePlainDateTime() | Result.map(Time.toPlainDate) // Ok{ value: 2026-10-22 }
```

## Applications

Check whether two moments fall on the same day for a user in a given zone:

```ant
func sameDay(a: Time.Instant, b: Time.Instant, z: Time.Zone): bool {
  a.inZone(z).toPlainDate() == b.inZone(z).toPlainDate()
}

let a = Time.parseInstant("2026-10-22T20:00:00Z").unwrap()
let b = Time.parseInstant("2026-10-23T03:00:00Z").unwrap()
sameDay(a, b, Time.utc)                               // false
sameDay(a, b, Time.zone("America/New_York").unwrap()) // true
```

Group log entries by the date they were written on, in the server's zone:

```ant
let server = Time.zone("Europe/Berlin").unwrap()
let entries = [
  "2026-10-21T22:30:00Z",
  "2026-10-22T08:00:00Z",
  "2026-10-22T22:15:00Z",
]

let byDate = entries
  | Array.map({ Time.parseInstant(x).unwrap().inZone(server) })
  | Array.groupBy(Time.toPlainDate)

for { key, value } in byDate {
  println("${key}: ${value.length()}")
}
```

```text
2026-10-22: 2
2026-10-23: 1
```

Count calendar days to a departure, ignoring the time of day:

```ant
let now = Time.parsePlainDateTime("2026-10-22T23:00").unwrap()
let departure = Time.parsePlainDateTime("2026-10-24T06:00").unwrap()
now.toPlainDate().until(departure.toPlainDate()) // P2D
now.until(departure)                             // P1DT7H
```

## Properties and relations

The date and the time of day together give back the wall-clock time:

```ant
let t = Time.parseZoned("2026-10-22T14:30:00+02:00[Europe/Stockholm]").unwrap()
let wall = t.toPlainDateTime()
assert(Time.plainDateTime(t.toPlainDate(), t.toPlainTime()) == wall)
assert(t.toPlainDate() == wall.toPlainDate())
```

The date keeps the year, the month and the day of its input:

```ant
let t = Time.parsePlainDateTime("1969-07-20T20:17").unwrap()
let d = t.toPlainDate()
assert(d.year() == t.year() && d.month() == t.month() && d.day() == t.day())
```

[Time.startOfDay](startOfDay.md) keeps the date, even on a day that does not start at midnight. In
Santiago, 2024-09-08 starts at 01:00:

```ant
let t = Time.parseZoned("2024-09-08T12:00[America/Santiago]").unwrap()
assert(t.startOfDay().toPlainDate() == t.toPlainDate())
```

## Possible issues

An instant has no date until it is in a zone, so passing one is a compile error. Choose the zone
whose calendar you mean with [Time.inZone](inZone.md), as in Scope:

```ant
let t = Time.parseInstant("2026-10-22T23:30:00Z").unwrap()
let d = Time.toPlainDate(t)
// error: cannot use Time.Instant where Time.ZonedDateTime | Time.PlainDateTime is expected
```

"Today" depends on the zone. Take it in the user's zone with [Time.nowIn](nowIn.md), not in UTC or
the host's zone:

<!-- test: check-only reads the clock -->

```ant
let user = Time.zone("Pacific/Auckland").unwrap()
let today = Time.nowIn(user).toPlainDate()
println(today)
```

## See also

[Time.toPlainTime](toPlainTime.md) · [Time.toPlainDateTime](toPlainDateTime.md) ·
[Time.inZone](inZone.md) · [Time.startOfDay](startOfDay.md) · [Time.plainDate](plainDate.md) ·
[Time.PlainDate](types/PlainDate.md) · [Time](index.md)
