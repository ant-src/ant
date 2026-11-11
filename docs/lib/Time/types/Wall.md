<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.Wall

```ant signature
type Time.Wall = Time.ZonedDateTime | Time.PlainDate | Time.PlainTime | Time.PlainDateTime
```

Time.Wall is the time types with fields, the ones [`Time.with`](../with.md) serves.

<!-- end generated -->

## Details

- A value typed `Time.Wall` does not pass to [`Time.with`](../with.md) until an `is` test narrows it
  to one member. A function with a type parameter bounded by `Time.Wall` passes its argument as it
  is.
- `Time.Wall` admits a `PlainDate` and a `PlainTime`, but each takes only its own fields:
  `Time.with` panics on a time field of a date or a date field of a time.

## Basic examples

A variable typed `Time.Wall` holds any of the four members:

```ant
let release: Time.Wall = Time.plainDate(2026, 4, 10)
release                   // 2026-04-10
release is Time.PlainDate // true
release is Time.PlainTime // false
```

A function whose type parameter is bounded by `Time.Wall` accepts every member and returns the type
it was given. Each member must still have the fields it replaces, so this one takes the three
members with a clock time:

```ant
func startOfMinute[T: Time.Wall](t: T): T {
  Time.with(t, second: 0, millisecond: 0, microsecond: 0, nanosecond: 0)
}

let meeting = Time.parseZoned("2026-05-04T10:15:42.5[Europe/Paris]").unwrap()
println(startOfMinute(Time.plainTime(9, 30, 45)))
println(startOfMinute(meeting.toPlainDateTime()))
println(startOfMinute(meeting))
```

```text
09:30:00
2026-05-04T10:15:00
2026-05-04T10:15:00+02:00[Europe/Paris]
```

On a `PlainDate` it panics, as [Possible issues](#possible-issues) shows.

## Scope

An array of `Time.Wall` mixes the members, and each renders as its own type does:

```ant
let meeting = Time.parseZoned("2026-05-04T10:00[Europe/Paris]").unwrap()
let marks: Array[Time.Wall] = [
  Time.plainDate(2026, 5, 4),
  Time.plainTime(10, 0),
  meeting.toPlainDateTime(),
  meeting,
]
marks                                 // [2026-05-04, 10:00:00, 2026-05-04T10:00:00, 2026-05-04T10:00:00+02:00[Europe/Paris]]
marks.filter({ x is Time.PlainDate }) // [2026-05-04]
```

Test the member with `is` to branch on it; after the last test, the remaining member is narrowed
too:

```ant
func describe(t: Time.Wall): string {
  if t is Time.PlainDate {
    "a date"
  } else if t is Time.PlainTime {
    "a clock time"
  } else if t is Time.PlainDateTime {
    "a date and time"
  } else {
    "a time in ${t.zoneOf().zoneId()}"
  }
}

let meeting = Time.parseZoned("2026-05-04T10:00[Europe/Paris]").unwrap()
let marks: Array[Time.Wall] = [Time.plainDate(2026, 5, 4), Time.plainTime(10, 0), meeting]
for m in marks {
  println("${m}: ${describe(m)}")
}
```

```text
2026-05-04: a date
10:00:00: a clock time
2026-05-04T10:00:00+02:00[Europe/Paris]: a time in Europe/Paris
```

Guards narrow a `Time.Wall` before the main path, so the last line calls `Time.with` on a
`ZonedDateTime`:

```ant
func firstOfMonth(t: Time.Wall): Time.Wall {
  if t is Time.PlainTime { return t }
  if t is Time.PlainDate { return t.with(day: 1) }
  if t is Time.PlainDateTime { return t.with(day: 1) }
  t.with(day: 1)
}

println(firstOfMonth(Time.plainDate(2026, 4, 10)))
println(firstOfMonth(Time.plainTime(9, 0)))
let flight = Time.parseZoned("2026-04-10T08:00[Asia/Tokyo]").unwrap()
println(firstOfMonth(flight))
```

```text
2026-04-01
09:00:00
2026-04-01T08:00:00+09:00[Asia/Tokyo]
```

## Applications

An agenda keeps all-day dates, daily clock times, and zoned meetings in one list, and renders each
entry by its member:

```ant
type Item = { title: string, when: Time.Wall }

func label(w: Time.Wall): string {
  if w is Time.PlainDate { return "all day ${w}" }
  if w is Time.PlainTime { return "daily at ${w}" }
  if w is Time.PlainDateTime { return "at ${w}" }
  "at ${w.toPlainDateTime()} in ${w.zoneOf().zoneId()}"
}

let day = Time.plainDate(2026, 5, 4)
let standup = Time.parseZoned("2026-05-04T09:30[America/New_York]").unwrap()
let agenda: Array[Item] = [
  Item{ title: "Release day", when: day },
  Item{ title: "Water the plants", when: Time.plainTime(8, 0) },
  Item{ title: "Dentist", when: Time.plainDateTime(day, Time.plainTime(15, 0)) },
  Item{ title: "Standup", when: standup },
]
for e in agenda {
  println("${e.title}: ${label(e.when)}")
}
```

```text
Release day: all day 2026-05-04
Water the plants: daily at 08:00:00
Dentist: at 2026-05-04T15:00:00
Standup: at 2026-05-04T09:30:00 in America/New_York
```

## Properties and relations

`Time.Wall` is [`Time.Point`](Point.md) without `Time.Instant`, and holds both
[`Time.Dated`](Dated.md) and [`Time.Timed`](Timed.md), so each assigns where a wider union is
expected:

```ant
let day: Time.Dated = Time.plainDate(2026, 1, 1)
let alarm: Time.Timed = Time.plainTime(7, 0)
let walls: Array[Time.Wall] = [day, alarm]
let points: Array[Time.Point] = [...walls]
assert(points.length() == 2)
```

## Possible issues

A value typed with the union does not pass to `Time.with`; narrow it first:

```ant
let release: Time.Wall = Time.plainDate(2026, 4, 10)
let first = Time.with(release, day: 1)
// error: T would be Time.Wall, which is not a single member of its bound Time.Wall; narrow first
```

`Time.Instant` is not a member: it has no fields to replace. Convert it to a `ZonedDateTime` with
[`Time.inZone`](../inZone.md) first:

```ant
let epoch = Time.fromEpochMillis(0)
let later = Time.with(epoch, year: 2000)
// error: T would be Time.Instant, which is not a member of its bound Time.Wall
```

A `Time.Wall` is not a `Time.Dated`, since a `PlainTime` has no date:

```ant
let release: Time.Wall = Time.plainDate(2026, 4, 10)
let year = Time.year(release) // error: cannot use Time.Wall where Time.Dated is expected
```

The bound admits every member, but a `PlainDate` takes no time fields, so a function bounded by
`Time.Wall` can still panic on one:

```ant
func startOfMinute[T: Time.Wall](t: T): T {
  Time.with(t, second: 0)
}

let today = startOfMinute(Time.plainDate(2026, 4, 10)) // panic: Time.with: a date has no time fields
```

A call to `unwrap` passed straight to a `Time.Wall` parameter takes `Time.Wall` as its expected
type, so it asks for a `Result[Time.Wall, Error]`, which a `Result[Time.ZonedDateTime, Error]` is
not. Bind the value with `let` first, as the examples above do:

```ant
func firstOfMonth(t: Time.Wall): Time.Wall {
  if t is Time.PlainTime { return t }
  if t is Time.PlainDate { return t.with(day: 1) }
  if t is Time.PlainDateTime { return t.with(day: 1) }
  t.with(day: 1)
}

println(firstOfMonth(Time.parseZoned("2026-04-10T08:00[Asia/Tokyo]").unwrap()))
// error: cannot use Result[Time.ZonedDateTime, Error] where Result[Time.Wall, Error] is expected
```

## See also

[Time.with](../with.md) · [Time.Point](Point.md) · [Time.Dated](Dated.md) · [Time.Timed](Timed.md) ·
[Time.PlainDate](PlainDate.md) · [Time.PlainTime](PlainTime.md) ·
[Time.PlainDateTime](PlainDateTime.md) · [Time.ZonedDateTime](ZonedDateTime.md)
