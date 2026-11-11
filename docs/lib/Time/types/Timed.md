<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.Timed

```ant signature
type Time.Timed = Time.ZonedDateTime | Time.PlainTime | Time.PlainDateTime
```

Time.Timed is the time types with a time of day, the ones the time fields such as
[`Time.hour`](../hour.md) read.

<!-- end generated -->

## Details

- A function with a type parameter bounded by `Time.Timed` or a wider union, such as
  [`Time.round`](../round.md) or [`Time.add`](../add.md), takes one member at a time. Narrow a
  `Time.Timed` value with `is` before you pass it.

## Basic examples

Use `Time.Timed` as a parameter type to accept any value with a time of day:

```ant
func clock(t: Time.Timed): string {
  let minute = "${t.minute()}".padStart(2, "0")
  "${t.hour()}:${minute}"
}

let alarm = Time.plainTime(7, minute: 5)
let flight = Time.plainDateTime(Time.plainDate(2026, 6, 1), Time.plainTime(18, minute: 40))
let call = Time.parseZoned("2026-06-02T09:30:00+02:00[Europe/Stockholm]").unwrap()
println(clock(alarm))
println(clock(flight))
println(clock(call))
```

```text
7:05
18:40
9:30
```

An array of `Time.Timed` holds the three kinds side by side:

```ant
let breakfast = Time.plainTime(7, minute: 30)
let meeting = Time.plainDateTime(Time.plainDate(2026, 1, 6), Time.plainTime(10, minute: 15))
let launch = Time.parseZoned("2026-01-07T21:00:45Z[UTC]").unwrap()
let times: Array[Time.Timed] = [breakfast, meeting, launch]
times                  // [07:30:00, 2026-01-06T10:15:00, 2026-01-07T21:00:45+00:00[UTC]]
times.map(Time.minute) // [30, 15, 0]
times.map(Time.second) // [0, 0, 45]
```

## Scope

Test which member a value holds with `is`:

```ant
func kind(t: Time.Timed): string {
  if t is Time.PlainTime {
    "a time"
  } else if t is Time.PlainDateTime {
    "a date and time"
  } else {
    "a time in ${t.zoneOf().zoneId()}"
  }
}

let tokyo = Time.parseZoned("2026-03-01T12:00:00+09:00[Asia/Tokyo]").unwrap()
println(kind(Time.plainTime(12)))
println(kind(Time.plainDateTime(Time.plainDate(2026, 3, 1), Time.plainTime(12))))
println(kind(tokyo))
```

```text
a time
a date and time
a time in Asia/Tokyo
```

A `Time.Timed` field in a record, and `?.` on a value that may be `null`:

```ant
type Reminder = { text: string, at: Time.Timed | null }
let soon = Reminder{ text: "stretch", at: Time.plainTime(15) }
let someday = Reminder{ text: "learn Ant", at: null }
soon.at?.hour()    // 15
someday.at?.hour() // null
```

Bound a type parameter by `Time.Timed` to keep the caller's type. A call through a type parameter is
positional, so write the namespace form to pass a named argument:

```ant
func toQuarter[T: Time.Timed](t: T): T {
  Time.round(t, Time.Minute{}, increment: 15)
}
let start = Time.plainDateTime(Time.plainDate(2026, 3, 1), Time.plainTime(23, minute: 53))
toQuarter(Time.plainTime(9, minute: 8)) // 09:15:00
toQuarter(start)                        // 2026-03-02T00:00:00
```

The time fields go down to the nanosecond:

```ant
let t = Time.plainTime(1, millisecond: 2, microsecond: 3, nanosecond: 4)
[t.millisecond(), t.microsecond(), t.nanosecond()] // [2, 3, 4]
```

To compare or sort by time of day across members, turn each value into its `Time.PlainTime`, a
single type that [`Time.compare`](../compare.md) takes:

```ant
func clockOf(t: Time.Timed): Time.PlainTime {
  if t is Time.PlainTime { t } else { t.toPlainTime() }
}

let a: Time.Timed = Time.plainTime(9, minute: 30)
let b = Time.parseZoned("2026-05-04T09:30:00+02:00[Europe/Paris]").unwrap()
Time.compare(clockOf(a), clockOf(b)) // 0
```

## Applications

Order the day's events by the clock, whatever kind of value each holds:

```ant
type Event = { label: string, at: Time.Timed }

func clockOf(t: Time.Timed): Time.PlainTime {
  if t is Time.PlainTime { t } else { t.toPlainTime() }
}

let call = Time.parseZoned("2026-11-12T17:00:00-05:00[America/New_York]").unwrap()
let events = [
  Event{ label: "standup", at: Time.plainTime(9, minute: 30) },
  Event{
    label: "dentist",
    at: Time.plainDateTime(Time.plainDate(2026, 11, 12), Time.plainTime(8, minute: 15)),
  },
  Event{ label: "call", at: call },
]
let ordered = events.sortWith(func(a: Event, b: Event): int {
  Time.compare(clockOf(a.at), clockOf(b.at))
})
for e in ordered {
  println("${clockOf(e.at)} ${e.label}")
}
```

```text
08:15:00 dentist
09:30:00 standup
17:00:00 call
```

Round each value to the hour, narrowing first so that each call of `Time.round` gets one type:

```ant
func toHour(t: Time.Timed): Time.Timed {
  if t is Time.PlainTime {
    t.round(Time.Hour{})
  } else if t is Time.PlainDateTime {
    t.round(Time.Hour{})
  } else {
    t.round(Time.Hour{})
  }
}

let night = Time.plainDateTime(Time.plainDate(2026, 12, 31), Time.plainTime(23, minute: 40))
let times: Array[Time.Timed] = [Time.plainTime(10, minute: 29), night]
let rounded = times.map(toHour)
println(rounded)
```

```text
[10:00:00, 2027-01-01T00:00:00]
```

## Properties and relations

A `Time.ZonedDateTime` and a `Time.PlainDateTime` have the time fields of their
[`Time.toPlainTime`](../toPlainTime.md):

```ant
let zoned = Time.parseZoned("2026-10-22T23:30:15+02:00[Europe/Stockholm]").unwrap()
let local = zoned.toPlainDateTime()
let p = zoned.toPlainTime()
let q = local.toPlainTime()
assert(p.hour() == zoned.hour() && p.minute() == zoned.minute() && p.second() == zoned.second())
assert(q.hour() == local.hour() && q.minute() == local.minute() && q.second() == local.second())
```

Values of different members are never `==`, even with the same time of day:

```ant
let a: Time.Timed = Time.plainTime(9)
let b: Time.Timed = Time.plainDateTime(Time.plainDate(2026, 1, 1), Time.plainTime(9))
assert(a != b)
assert(a == Time.plainTime(9))
```

## Possible issues

A `Time.Instant` and a `Time.PlainDate` are not members:

```ant
let epoch: Time.Timed = Time.fromEpochMillis(0)
// error: cannot use Time.Instant where Time.Timed is expected
let date: Time.Timed = Time.plainDate(2026, 1, 1)
// error: cannot use Time.PlainDate where Time.Timed is expected
```

A `Time.Timed` value does not reach a generic function such as `Time.round`, `Time.with` or
`Time.compare` until `is` narrows it:

```ant
let t: Time.Timed = Time.plainTime(9, minute: 20)
let rounded = t.round(Time.Hour{})
// error: T would be Time.Timed, which is not a single member of its bound Time.Instant | Time.Timed; narrow first
let moved = t.with(minute: 0)
// error: T would be Time.Timed, which is not a single member of its bound Time.Wall; narrow first
```

`Time.Timed` is not `Time.Dated`: a `Time.PlainTime` has no date, so the date fields do not take a
`Time.Timed`:

```ant
let t: Time.Timed = Time.plainTime(9)
let year = Time.year(t) // error: cannot use Time.Timed where Time.Dated is expected
```

The expected type solves a generic call first, so an `unwrap` checked against `Time.Timed`, in an
annotation or as an argument, asks for a `Result[Time.Timed, Error]`, which a
`Result[Time.ZonedDateTime, Error]` is not:

```ant
let t: Time.Timed = Time.parseZoned("2026-06-02T09:30:00Z[UTC]").unwrap()
// error: cannot use Result[Time.ZonedDateTime, Error] where Result[Time.Timed, Error] is expected
```

Name the value first, then use it as a `Time.Timed`:

```ant
let zoned = Time.parseZoned("2026-06-02T09:30:00Z[UTC]").unwrap()
let t: Time.Timed = zoned
t.hour() // 9
```

Inside a function bounded by `Time.Timed`, a call through `.` takes no named arguments:

```ant
func toQuarter[T: Time.Timed](t: T): T {
  t.round(Time.Minute{}, increment: 15)
  // error: a call through T is positional, since each implementation names its parameters as it likes
}
```

## See also

[Time.hour](../hour.md) · [Time.minute](../minute.md) · [Time.round](../round.md) ·
[Time.Dated](Dated.md) · [Time.Wall](Wall.md) · [Time.Point](Point.md) ·
[Time.PlainTime](PlainTime.md) · [Time.PlainDateTime](PlainDateTime.md) ·
[Time.ZonedDateTime](ZonedDateTime.md)
