<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.PlainTime

```ant signature
type Time.PlainTime
```

Time.PlainTime is a wall-clock time of day with no date and no zone, such as an alarm. It renders as
14:03:05, with a fractional second to its last nonzero digit when there is one.

<!-- end generated -->

## Details

- Make one with [Time.plainTime](../plainTime.md), from text with
  [Time.parsePlainTime](../parsePlainTime.md), or from a date and time with
  [Time.toPlainTime](../toPlainTime.md). It has no literal.
- A value is immutable. [Time.with](../with.md), [Time.add](../add.md), and
  [Time.round](../round.md) return a new time.
- Arithmetic wraps around the clock: 23:59 plus two minutes is 00:01, and days, weeks, months, and
  years in a duration are ignored.
- Two times are `==` when every field is. Order them with [Time.compare](../compare.md); `<` does
  not apply.

## Basic examples

Build a time and read it back:

```ant
let t = Time.plainTime(14, 3, 5)
t              // 14:03:05
t.hour()       // 14
t.minute()     // 3
String.from(t) // "14:03:05"
```

Parse one from text. Given a date and time, it takes the time:

```ant
Time.parsePlainTime("07:30")            // Ok{ value: 07:30:00 }
Time.parsePlainTime("14:03:05.123")     // Ok{ value: 14:03:05.123 }
Time.parsePlainTime("2026-10-22T14:03") // Ok{ value: 14:03:00 }
```

Annotate a parameter with the type:

```ant
func greeting(t: Time.PlainTime): string {
  if t.hour() < 12 { "Good morning" } else { "Good afternoon" }
}

println(greeting(Time.plainTime(9, 15)))
println(greeting(Time.plainTime(15)))
```

```text
Good morning
Good afternoon
```

## Scope

A fractional second renders to its last nonzero digit, and down to the nanosecond:

```ant
Time.plainTime(8, 0, 0, 500)              // 08:00:00.5
Time.plainTime(23, 59, 59, 999, 999, 999) // 23:59:59.999999999
```

Take the time of day from a `Time.PlainDateTime` or a `Time.ZonedDateTime`:

```ant
let meeting = Time.parsePlainDateTime("2026-10-22T09:30:00").unwrap()
meeting.toPlainTime() // 09:30:00
```

Put it back on a date with `Time.plainDateTime`:

```ant
Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(18, 45)) // 2026-10-22T18:45:00
```

Replace fields with `Time.with`, which keeps the others:

```ant
let t = Time.plainTime(14, 3, 5)
t.with(hour: 8)              // 08:03:05
t.with(minute: 0, second: 0) // 14:00:00
```

Add, subtract, and round:

```ant
let t = Time.plainTime(14, 37)
t.add(45m)                            // 15:22:00
t.subtract(15h)                       // 23:37:00
t.round(Time.Minute{}, increment: 15) // 14:30:00
```

Measure between two times with `Time.until` and `Time.since`. The largest unit is `Hour` unless you
name a smaller one:

```ant
let open = Time.plainTime(9)
let close = Time.plainTime(17, 30)
Time.until(open, close)                             // PT8H30M
Time.until(open, close, largestUnit: Time.Minute{}) // PT510M
Time.since(open, close)                             // -PT8H30M
```

A `Time.PlainTime` is a member of `Time.Point`, `Time.Wall`, and `Time.Timed`, and `is` picks it
out:

```ant
func describe(p: Time.Point): string {
  if p is Time.PlainTime { "the time ${p}" } else { "a date or a moment: ${p}" }
}

println(describe(Time.plainTime(6, 45)))
println(describe(Time.plainDate(2026, 10, 22)))
```

```text
the time 06:45:00
a date or a moment: 2026-10-22
```

Times work as map keys, and `Json.stringify` writes one as a string:

```ant
let alarms = Map{ Time.plainTime(6, 45): "wake up", Time.plainTime(22): "sleep" }
alarms[Time.plainTime(6, 45)]         // "wake up"
Json.stringify(Time.plainTime(6, 45)) // "\"06:45:00\""
```

## Applications

Check whether a shop is open at a given time:

```ant
func isOpen(t: Time.PlainTime): bool {
  let open = Time.plainTime(9)
  let close = Time.plainTime(18)
  Time.compare(t, open) >= 0 && Time.compare(t, close) < 0
}

println(isOpen(Time.plainTime(8, 59)))
println(isOpen(Time.plainTime(12, 30)))
println(isOpen(Time.plainTime(18)))
```

```text
false
true
false
```

Sort a day's alarms:

```ant
let alarms = [Time.plainTime(18), Time.plainTime(7, 30), Time.plainTime(12, 15)]
alarms.sortWith(Time.compare) // [07:30:00, 12:15:00, 18:00:00]
```

Print a timetable every 20 minutes from a first departure:

```ant
let first = Time.plainTime(23, 10)
for i in Array.range(0, 4) {
  println(first.add(Time.duration(minutes: 20 * i)))
}
```

```text
23:10:00
23:30:00
23:50:00
00:10:00
```

## Properties and relations

A time renders as text that parses back to the same time:

```ant
let t = Time.plainTime(14, 3, 5, 120)
assert(Time.parsePlainTime(String.from(t)).unwrap() == t)
```

Adding 24 hours wraps around the clock back to the same time, and a duration of days is ignored:

```ant
let t = Time.plainTime(10, 30)
assert(t.add(24h) == t)
assert(t.add(Time.duration(days: 1)) == t)
```

`Time.compare` and `==` agree:

```ant
let a = Time.plainTime(9, 0)
let b = Time.parsePlainTime("09:00").unwrap()
assert(a == b)
assert(Time.compare(a, b) == 0)
```

## Possible issues

`<` and the other comparison operators do not apply to times. Use `Time.compare`:

```ant
let a = Time.plainTime(9)
let b = Time.plainTime(17)
let early = a < b
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.PlainTime
```

A time has no date, so the date fields are a compile error, and `Time.with` panics on one:

```ant
Time.year(Time.plainTime(9)) // error: cannot use Time.PlainTime where Time.Dated is expected
```

```ant
Time.plainTime(9).with(day: 3) // panic: Time.with: a time has no date fields
```

The constructor rejects a field out of range, while `Time.with` constrains it:

```ant
Time.plainTime(24)                           // panic: Time.plainTime: hour out of range: 24
Time.plainTime(9).with(hour: 25, minute: -3) // 23:00:00
```

`Time.until` gives no unit larger than an hour, and `Time.round` does not round to a day:

```ant
Time.until(Time.plainTime(9), Time.plainTime(10), largestUnit: Time.Day{}) // panic: Time.until: a time has no days
Time.plainTime(9).round(Time.Day{})                                        // panic: Time.round: a time cannot be rounded to a day
```

A span across midnight is negative, since 06:00 is earlier than 22:00 and a time has no date. Add a
date when the span crosses midnight:

```ant
Time.until(Time.plainTime(22), Time.plainTime(6)) // -PT16H
let start = Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(22))
let end = Time.plainDateTime(Time.plainDate(2026, 10, 23), Time.plainTime(6))
Time.until(start, end) // PT8H
```

`Time.PlainTime` is opaque and has no literal:

```ant
let t = Time.PlainTime{ hour: 9 } // error: Time.PlainTime is not a record type, so it has no literal
```

A `Z` or an offset names a moment, not a wall-clock time, so `Time.parsePlainTime` rejects it:

```ant
Time.parsePlainTime("2026-10-22T14:00Z") // Err{ error: Error{ message: "Z names an exact time; a plain value needs a wall-clock time" } }
```

## See also

[Time.plainTime](../plainTime.md) · [Time.parsePlainTime](../parsePlainTime.md) ·
[Time.toPlainTime](../toPlainTime.md) · [Time.plainDateTime](../plainDateTime.md) ·
[Time.hour](../hour.md) · [Time.with](../with.md) · [Time.add](../add.md) ·
[Time.until](../until.md) · [Time.round](../round.md) · [Time.compare](../compare.md) ·
[Time.PlainDate](PlainDate.md) · [Time.PlainDateTime](PlainDateTime.md) · [Time.Timed](Timed.md)
