<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.Dated

```ant signature
type Time.Dated = Time.ZonedDateTime | Time.PlainDate | Time.PlainDateTime
```

Time.Dated is the time types with a date, the ones the date fields such as [`Time.year`](../year.md)
read.

<!-- end generated -->

## Details

- A `Time.Instant` has no date until `Time.inZone` places it in a zone, and a `Time.PlainTime` has
  none.
- A `Time.Dated` value passes as it is to the date fields, such as `Time.year` and `Time.dayOfWeek`,
  and to the `relativeTo` parameter of `Time.total`, `Time.compareDuration`, and
  `Time.roundDuration`.
- A generic function bounded by a union, such as `Time.add`, `Time.compare`, or `Time.with`, needs a
  single member. Narrow a `Time.Dated` value with `is` first.
- As `relativeTo`, a `Time.PlainDateTime` counts as its date and its time is ignored. Only a
  `Time.ZonedDateTime` gives days other than 24 hours.

## Basic examples

A parameter typed `Time.Dated` takes all three types with a date:

```ant
func label(t: Time.Dated): string {
  "${t.day()}/${t.month()}/${t.year()}"
}

let stockholm = Result.unwrap(Time.zone("Europe/Stockholm"))
let date = Time.plainDate(2026, 11, 11)
let morning = Time.plainDateTime(date, Time.plainTime(hour: 9))
label(date)                       // "11/11/2026"
label(morning)                    // "11/11/2026"
label(morning.toZoned(stockholm)) // "11/11/2026"
```

The date fields read a `Time.Dated` value without narrowing:

```ant
let when: Time.Dated = Time.plainDate(2026, 2, 14)
Time.year(when)    // 2026
when.dayOfWeek()   // 6
when.daysInMonth() // 28
```

## Scope

An array of `Time.Dated` holds the three types side by side. The date fields read each of them, and
each renders in its own form:

```ant
let stockholm = Result.unwrap(Time.zone("Europe/Stockholm"))
let date = Time.plainDate(2026, 11, 11)
let morning = Time.plainDateTime(date, Time.plainTime(hour: 9))
let all: Array[Time.Dated] = [date, morning, morning.toZoned(stockholm)]
all.map({ x.weekOfYear() }) // [46, 46, 46]

all // [2026-11-11, 2026-11-11T09:00:00, 2026-11-11T09:00:00+01:00[Europe/Stockholm]]
```

`is` tells the members apart and narrows the value to one of them:

```ant
func describe(t: Time.Dated): string {
  if t is Time.PlainDate {
    "a date"
  } else if t is Time.PlainDateTime {
    "a date at ${t.toPlainTime()}"
  } else {
    "a date in ${t.zoneOf()}"
  }
}

let date = Time.plainDate(2026, 1, 1)
describe(date)                                              // "a date"
describe(Time.plainDateTime(date, Time.plainTime(hour: 8))) // "a date at 08:00:00"
describe(Time.fromEpochMillis(0) | Time.inZone(Time.utc))   // "a date in UTC"
```

A `Time.ZonedDateTime` has the date its zone shows, so one instant falls on two dates in two zones:

```ant
let instant = Result.unwrap(Time.parseInstant("2026-11-11T23:30:00Z"))
let tokyo = Result.unwrap(Time.zone("Asia/Tokyo"))
let dates: Array[Time.Dated] = [instant.inZone(Time.utc), instant.inZone(tokyo)]
dates.map({ x.day() }) // [11, 12]
```

A type parameter bounded by `Time.Dated` stands for one member, so it reaches the generic functions
and the result keeps the argument's type:

```ant
func firstOfMonth[T: Time.Dated](t: T): T {
  Time.with(t, day: 1)
}

let stockholm = Result.unwrap(Time.zone("Europe/Stockholm"))
let date = Time.plainDate(2026, 11, 11)
let zoned = Time.plainDateTime(date, Time.plainTime(hour: 9)).toZoned(stockholm)
firstOfMonth(date)  // 2026-11-01
firstOfMonth(zoned) // 2026-11-01T09:00:00+01:00[Europe/Stockholm]
```

`relativeTo` takes any `Time.Dated` value. A month is as long as the month it starts at, and the
noon of the `Time.PlainDateTime` is ignored:

```ant
let stockholm = Result.unwrap(Time.zone("Europe/Stockholm"))
let month = Time.duration(months: 1)
let starts: Array[Time.Dated] = [
  Time.plainDate(2026, 2, 1),
  Time.plainDateTime(Time.plainDate(2028, 2, 1), Time.plainTime(hour: 12)),
  Time.plainDateTime(Time.plainDate(2026, 3, 1), Time.plainTime()).toZoned(stockholm),
]
starts.map({ Time.total(month, Time.Day{}, relativeTo: x) }) // [28.0, 29.0, 31.0]
```

## Applications

Label any value with a date by its quarter:

```ant
func quarter(t: Time.Dated): string {
  "Q${(t.month() - 1) / 3 + 1} ${t.year()}"
}

let stockholm = Result.unwrap(Time.zone("Europe/Stockholm"))
quarter(Time.plainDate(2026, 12, 24))                                             // "Q4 2026"
quarter(Time.plainDateTime(Time.plainDate(2026, 3, 1), Time.plainTime(hour: 18))) // "Q1 2026"
quarter(Time.fromEpochMillis(0) | Time.inZone(stockholm))                         // "Q1 1970"
```

Sort events of mixed types by their dates. Narrow each to a `Time.PlainDate`, a single type that
`Time.compare` orders; `dateOf` is the narrowing from Possible issues:

```ant
func dateOf(t: Time.Dated): Time.PlainDate {
  if t is Time.PlainDate { t } else { t.toPlainDate() }
}

let stockholm = Result.unwrap(Time.zone("Europe/Stockholm"))
let events: Array[Time.Dated] = [
  Time.plainDate(2026, 12, 24),
  Time.plainDateTime(Time.plainDate(2026, 3, 1), Time.plainTime(hour: 18)),
  Time.plainDateTime(Time.plainDate(2026, 6, 19), Time.plainTime(hour: 12)).toZoned(stockholm),
]
events | Array.map(dateOf) | Array.sortWith(Time.compare) // [2026-03-01, 2026-06-19, 2026-12-24]
```

Count the hours of a day. A `Time.ZonedDateTime` as `relativeTo` gives the zone's real day, which
loses an hour at the spring change:

```ant
let stockholm = Result.unwrap(Time.zone("Europe/Stockholm"))
let change = Time.plainDateTime(Time.plainDate(2026, 3, 29), Time.plainTime()).toZoned(stockholm)
let day = Time.duration(days: 1)
Time.total(day, Time.Hour{}, relativeTo: change)               // 23.0
Time.total(day, Time.Hour{}, relativeTo: change.toPlainDate()) // 24.0
```

## Properties and relations

Every member has the date fields of its `Time.PlainDate`:

```ant
func dateOf(t: Time.Dated): Time.PlainDate {
  if t is Time.PlainDate { t } else { t.toPlainDate() }
}

let stockholm = Result.unwrap(Time.zone("Europe/Stockholm"))
let date = Time.plainDate(2028, 2, 29)
let all: Array[Time.Dated] = [
  date,
  Time.plainDateTime(date, Time.plainTime(hour: 23)),
  Time.plainDateTime(date, Time.plainTime(hour: 23)).toZoned(stockholm),
]
for t in all {
  let d = dateOf(t)
  assert(t.year() == d.year() && t.month() == d.month() && t.day() == d.day())
  assert(t.dayOfYear() == d.dayOfYear() && t.inLeapYear() == d.inLeapYear())
}
```

As `relativeTo`, a `Time.PlainDateTime` counts the same as its date:

```ant
let month = Time.duration(months: 1)
let dt = Time.plainDateTime(Time.plainDate(2028, 2, 1), Time.plainTime(hour: 12))
assert(
  Time.total(month, Time.Day{}, relativeTo: dt)
    == Time.total(month, Time.Day{}, relativeTo: dt.toPlainDate()),
)
```

## Possible issues

A `Time.Instant` is not a `Time.Dated`. Place it in a zone first:

```ant
let t: Time.Dated = Time.now() // error: cannot use Time.Instant where Time.Dated is expected
```

```ant
let utc: Time.Dated = Time.fromEpochMillis(0) | Time.inZone(Time.utc)
utc.year() // 1970
```

A `Time.PlainTime` has no date:

```ant
let t: Time.Dated = Time.plainTime(hour: 9)
// error: cannot use Time.PlainTime where Time.Dated is expected
```

A generic function such as `Time.add` needs a single member, so a value typed `Time.Dated` does not
reach it:

```ant
let t: Time.Dated = Time.plainDate(2026, 1, 1)
let u = Time.add(t, Time.duration(days: 1))
// error: T would be Time.Dated, which is not a single member of its bound Time.Point; narrow first
```

Narrow it with `is`, one branch per member, or take a type parameter bounded by `Time.Dated`, as
`firstOfMonth` in Scope does:

```ant
func nextDay(t: Time.Dated): Time.Dated {
  let day = Time.duration(days: 1)
  if t is Time.ZonedDateTime {
    Time.add(t, day)
  } else if t is Time.PlainDate {
    Time.add(t, day)
  } else {
    Time.add(t, day)
  }
}

nextDay(Time.plainDate(2026, 12, 31)) // 2027-01-01
```

`is` cannot test `Time.Dated` itself, since an alias has no run-time identity:

```ant
func hasDate(t: Time.Point): bool {
  t is Time.Dated // error: Time.Dated is an alias and has no run-time identity; test its members instead
}
```

A `Time.Dated` value may have no time, so the time fields do not take it:

```ant
let t: Time.Dated = Time.plainDate(2026, 1, 1)
let h = Time.hour(t) // error: cannot use Time.Dated where Time.Timed is expected
```

`Time.toPlainDate` takes a `Time.ZonedDateTime` or a `Time.PlainDateTime`, not a `Time.Dated`, so it
does not take a value typed `Time.Dated`:

```ant
let when: Time.Dated = Time.plainDate(2026, 1, 1)
let d = when.toPlainDate()
// error: Time.toPlainDate takes Time.ZonedDateTime | Time.PlainDateTime first, so a Time.Dated does not reach it
```

Narrow it first; a `Time.PlainDate` is its own date:

```ant
func dateOf(t: Time.Dated): Time.PlainDate {
  if t is Time.PlainDate { t } else { t.toPlainDate() }
}

let date = Time.plainDate(2026, 1, 1)
dateOf(date)                                              // 2026-01-01
dateOf(Time.plainDateTime(date, Time.plainTime(hour: 9))) // 2026-01-01
```

Values of different members are never `==`, even on the same date:

```ant
let date = Time.plainDate(2026, 1, 1)
let a: Time.Dated = date
let b: Time.Dated = Time.plainDateTime(date, Time.plainTime())
a == b                          // false
a == Time.plainDate(2026, 1, 1) // true
```

## See also

[Time.Timed](Timed.md) · [Time.Wall](Wall.md) · [Time.Point](Point.md) ·
[Time.PlainDate](PlainDate.md) · [Time.PlainDateTime](PlainDateTime.md) ·
[Time.ZonedDateTime](ZonedDateTime.md) · [Time.year](../year.md) · [Time.dayOfWeek](../dayOfWeek.md)
· [Time.total](../total.md) · [Time.inZone](../inZone.md)
