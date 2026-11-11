<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.PlainDateTime

```ant signature
type Time.PlainDateTime
```

Time.PlainDateTime is a date and a wall-clock time with no zone, such as a meeting before its zone
is known. It renders as 2026-10-22T14:03:05; [`Time.toZoned`](../toZoned.md) places it in a zone.

<!-- end generated -->

## Details

- It is opaque and immutable: read its parts with functions such as [`Time.year`](../year.md) and
  [`Time.hour`](../hour.md), and change them with [`Time.with`](../with.md) or
  [`Time.add`](../add.md), which return a new value.
- Arithmetic treats every day as 24 hours, with no daylight-saving changes.
- `==` compares the fields, and [`Time.compare`](../compare.md) orders by date, then time. `<` does
  not apply.

## Basic examples

Build one from a date and a time with [`Time.plainDateTime`](../plainDateTime.md):

```ant
Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(14, 3, 5)) // 2026-10-22T14:03:05
```

Parse one from text with [`Time.parsePlainDateTime`](../parsePlainDateTime.md), which returns a
`Result`:

```ant
Time.parsePlainDateTime("2026-10-22T14:03") // Ok{ value: 2026-10-22T14:03:00 }
Time.parsePlainDateTime("2026-10-22 14:03") // Ok{ value: 2026-10-22T14:03:00 }
Time.parsePlainDateTime("2026-02-30T10:00") // Err{ error: Error{ message: "no such date: 2026-02-30" } }
```

A full zoned string parses too. It keeps the wall clock and drops the zone:

```ant
Time.parsePlainDateTime("2026-10-22T14:03:05+02:00[Europe/Stockholm]") // Ok{ value: 2026-10-22T14:03:05 }
```

Read its date and its time fields:

```ant
let t = Time.parsePlainDateTime("2026-10-22T14:03:05").unwrap()
t.year()      // 2026
t.month()     // 10
t.dayOfWeek() // 4
t.hour()      // 14
t.second()    // 5
```

## Scope

Take the wall-clock date and time of a zoned value with
[`Time.toPlainDateTime`](../toPlainDateTime.md), and its parts with
[`Time.toPlainDate`](../toPlainDate.md) and [`Time.toPlainTime`](../toPlainTime.md):

```ant
let zoned = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let t = zoned.toPlainDateTime()
t               // 2026-10-22T16:03:05
t.toPlainDate() // 2026-10-22
t.toPlainTime() // 16:03:05
```

It renders as its ISO 8601 string in `String.from`, interpolation, and `Json.stringify`. The
fraction of a second is written to its last nonzero digit:

```ant
let t = Time.parsePlainDateTime("2026-10-22T14:03:05.120").unwrap()
String.from(t)    // "2026-10-22T14:03:05.12"
"due ${t}"        // "due 2026-10-22T14:03:05.12"
Json.stringify(t) // "\"2026-10-22T14:03:05.12\""
```

`Time.with` replaces any of the date and time fields, and constrains a day that does not fit the new
month:

```ant
let t = Time.parsePlainDateTime("2026-01-31T14:03:05").unwrap()
t.with(hour: 9, minute: 0, second: 0) // 2026-01-31T09:00:00
t.with(month: 4)                      // 2026-04-30T14:03:05
```

`Time.add` takes both calendar units and exact units. Hours carry over midnight:

```ant
let t = Time.parsePlainDateTime("2026-10-22T14:03:05").unwrap()
t.add(Time.duration(hours: 12))             // 2026-10-23T02:03:05
t.add(Time.duration(months: 1, minutes: 5)) // 2026-11-22T14:08:05
t.subtract(Time.duration(days: 22))         // 2026-09-30T14:03:05
```

`Time.until` counts up to days by default, and to larger units with `largestUnit:`.
[`Time.round`](../round.md) rounds the time of day:

```ant
let t = Time.parsePlainDateTime("2026-10-22T14:03:05").unwrap()
let end = Time.parsePlainDateTime("2026-12-25T00:00").unwrap()
t.until(end)                            // P63DT9H56M55S
t.until(end, largestUnit: Time.Month{}) // P2M2DT9H56M55S
t.round(Time.Hour{})                    // 2026-10-22T14:00:00
```

It is a member of `Time.Point`, `Time.Wall`, `Time.Dated`, and `Time.Timed`, so a parameter of one
of those unions takes it, and `is` narrows it back:

```ant
func describe(x: Time.Point): string {
  if x is Time.PlainDateTime { "a date and time" } else { "something else" }
}

let meeting = Time.parsePlainDateTime("2026-10-22T14:00").unwrap()
println(describe(meeting))
println(describe(Time.plainDate(2026, 10, 22)))
```

```text
a date and time
something else
```

It serves as a map key:

```ant
let t = Time.parsePlainDateTime("2026-10-22T14:00").unwrap()
let rooms: Map[Time.PlainDateTime, string] = Map{}
rooms[t] = "Aula"
rooms[Time.parsePlainDateTime("2026-10-22T14:00:00").unwrap()] // "Aula"
```

## Applications

Sort appointments with `Time.compare`, then group them by date:

```ant
let slots = ["2026-10-22T09:00", "2026-10-21T17:30", "2026-10-22T08:15"]
  .map({ Time.parsePlainDateTime(x).unwrap() })
let byDay = slots.sortWith(Time.compare).groupBy({ x.toPlainDate() })
for { key, value } in byDay {
  println("${key}: ${value.map({ x.toPlainTime() })}")
}
```

```text
2026-10-21: [17:30:00]
2026-10-22: [08:15:00, 09:00:00]
```

Keep an event as the wall-clock time it was agreed in, and place it in a zone only to show it:

```ant
let event = Time.parsePlainDateTime("2026-10-22T15:00").unwrap()
let start = event.toZoned(Time.zone("Europe/Stockholm").unwrap())
for name in ["Asia/Tokyo", "America/New_York"] {
  println("${name}: ${start.withZone(Time.zone(name).unwrap()).toPlainDateTime()}")
}
```

```text
Asia/Tokyo: 2026-10-22T22:00:00
America/New_York: 2026-10-22T09:00:00
```

## Properties and relations

The rendered string parses back to an equal value:

```ant
let t = Time.parsePlainDateTime("2026-10-22T14:03:05.5").unwrap()
assert(Time.parsePlainDateTime(String.from(t)).unwrap() == t)
```

Its date and time rebuild it with `Time.plainDateTime`:

```ant
let t = Time.parsePlainDateTime("2026-10-22T14:03:05").unwrap()
assert(Time.plainDateTime(t.toPlainDate(), t.toPlainTime()) == t)
```

Outside a daylight-saving change, placing it in a zone and taking the wall clock back gives it
again:

```ant
let t = Time.parsePlainDateTime("2026-10-22T14:03:05").unwrap()
let ny = Time.zone("America/New_York").unwrap()
assert(t.toZoned(ny).toPlainDateTime() == t)
```

## Possible issues

A day is always 24 hours. New York skips an hour on March 8, 2026, so from noon to noon is 24 hours
on the wall clock but 23 hours in the zone:

```ant
let ny = Time.zone("America/New_York").unwrap()
let sat = Time.parsePlainDateTime("2026-03-07T12:00").unwrap()
let sun = Time.parsePlainDateTime("2026-03-08T12:00").unwrap()
sat.until(sun, largestUnit: Time.Hour{}) // PT24H
sat.toZoned(ny).until(sun.toZoned(ny))   // PT23H
```

It has no zone, so functions that need a moment do not take it:

```ant
let t = Time.parsePlainDateTime("2026-10-22T14:00").unwrap()
let ms = Time.epochMillis(t)
// error: cannot use Time.PlainDateTime where Time.Instant | Time.ZonedDateTime is expected
```

It has no fields to read and no literal. Call the field functions, and build it with
`Time.plainDateTime`:

```ant
let t = Time.parsePlainDateTime("2026-10-22T14:00").unwrap()
let y = t.year               // error: Time.PlainDateTime is not a record, so it has no field year
let u = Time.PlainDateTime{} // error: Time.PlainDateTime is not a record type, so it has no literal
```

`<` and `Array.sort` need an ordered type. Use `Time.compare` and `Array.sortWith`:

```ant
let a = Time.parsePlainDateTime("2026-10-22T09:00").unwrap()
let b = Time.parsePlainDateTime("2026-10-21T17:30").unwrap()
let sorted = [a, b].sort() // error: T would be Time.PlainDateTime, which is not a member of its bound Ordered
```

A `Time.PlainDateTime` never equals a `Time.PlainDate`, even at midnight. Compare the date part:

```ant
let t = Time.parsePlainDateTime("2026-10-22T00:00").unwrap()
let d = Time.plainDate(2026, 10, 22)
let same = t == d
// error: Time.PlainDateTime and Time.PlainDate cannot be compared with ==: neither is assignable to the other
```

```ant
let t = Time.parsePlainDateTime("2026-10-22T00:00").unwrap()
t.toPlainDate() == Time.plainDate(2026, 10, 22) // true
```

Text with `Z` names a moment, not a wall-clock time, so it does not parse:

```ant
Time.parsePlainDateTime("2026-10-22T14:03:05Z") // Err{ error: Error{ message: "Z names an exact time; a plain value needs a wall-clock time" } }
```

A bare date does not parse, since it has no time. Parse it with
[`Time.parsePlainDate`](../parsePlainDate.md) and add a time with `Time.plainDateTime`:

```ant
Time.parsePlainDateTime("2026-10-22") // Err{ error: Error{ message: "a date-time needs a date and a time: \"2026-10-22\"" } }
```

```ant
let d = Time.parsePlainDate("2026-10-22").unwrap()
Time.plainDateTime(d, Time.plainTime(0, 0, 0)) // 2026-10-22T00:00:00
```

A wall-clock time that a daylight-saving change skipped has no instant in that zone. `Time.toZoned`
takes the later one by default, so the wall clock read back is not the one given. A program that
wants to refuse such a time tests the length of [`Time.possibleInstants`](../possibleInstants.md):

```ant
let ny = Time.zone("America/New_York").unwrap()
let gap = Time.parsePlainDateTime("2026-03-08T02:30").unwrap()
gap.toZoned(ny)                   // 2026-03-08T03:30:00-04:00[America/New_York]
gap.possibleInstants(ny).length() // 0
```

## See also

[Time.plainDateTime](../plainDateTime.md) · [Time.parsePlainDateTime](../parsePlainDateTime.md) ·
[Time.toPlainDateTime](../toPlainDateTime.md) · [Time.toZoned](../toZoned.md) ·
[Time.possibleInstants](../possibleInstants.md) · [Time.with](../with.md) ·
[Time.compare](../compare.md) · [Time.PlainDate](PlainDate.md) · [Time.PlainTime](PlainTime.md) ·
[Time.ZonedDateTime](ZonedDateTime.md) · [Time](../index.md)
