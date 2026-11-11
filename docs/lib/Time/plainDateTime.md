<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.plainDateTime

```ant signature
func Time.plainDateTime(date: Time.PlainDate, time: Time.PlainTime): Time.PlainDateTime
```

Time.plainDateTime returns the wall-clock date and time made of date and time.

<!-- end generated -->

## Details

- The result has no zone: it is what a calendar and a clock show, not a moment on the time line.
  Place it in a zone with [Time.toZoned](toZoned.md) to get one.
- Any date goes with any time, even a time that a daylight-saving change skips in some zone: the
  checks happen when you build the parts with [Time.plainDate](plainDate.md) and
  [Time.plainTime](plainTime.md).
- To build from text, use [Time.parsePlainDateTime](parsePlainDateTime.md). It returns `Err` on bad
  input, where `Time.plainDate` and `Time.plainTime` panic.

## Basic examples

```ant
let date = Time.plainDate(2026, 10, 22)
Time.plainDateTime(date, Time.plainTime(14, 3, 5)) // 2026-10-22T14:03:05
Time.plainDateTime(date, Time.plainTime(9))        // 2026-10-22T09:00:00
```

`Time.plainTime()` with no arguments is midnight, the start of the date:

```ant
Time.plainDateTime(Time.plainDate(2027, 1, 1), Time.plainTime()) // 2027-01-01T00:00:00
```

## Scope

The time keeps its fraction of a second, written to the last nonzero digit:

```ant
let date = Time.plainDate(2026, 10, 22)
Time.plainDateTime(date, Time.plainTime(9, 30, millisecond: 250)) // 2026-10-22T09:30:00.25
```

A year outside 0 to 9999 renders with a sign and six digits:

```ant
Time.plainDateTime(Time.plainDate(-50, 3, 1), Time.plainTime(12))   // -000050-03-01T12:00:00
Time.plainDateTime(Time.plainDate(12026, 3, 1), Time.plainTime(12)) // +012026-03-01T12:00:00
```

The parts may come from other values. Take the date a zoned value shows and set a new time on it:

```ant
let arrival = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
Time.plainDateTime(arrival.toPlainDate(), Time.plainTime(9)) // 2026-10-22T09:00:00
```

Inside a pipe, pass it in a lambda to combine a parsed date with a fixed time:

```ant
"2026-12-24"
  | Time.parsePlainDate()
  | Result.map({ Time.plainDateTime(x, Time.plainTime(18)) }) // Ok{ value: 2026-12-24T18:00:00 }
```

Naming the arguments lets them come in either order:

```ant
let date = Time.plainDate(2026, 10, 22)
Time.plainDateTime(time: Time.plainTime(8), date: date) // 2026-10-22T08:00:00
```

## Applications

List a week of stand-ups at 09:30, one each weekday:

```ant
let monday = Time.plainDate(2026, 10, 19)
let standup = Time.plainTime(9, 30)
let week = [0, 1, 2, 3, 4].map({ Time.plainDateTime(monday.add(Time.duration(days: x)), standup) })
for meeting in week {
  println(meeting)
}
```

```text
2026-10-19T09:30:00
2026-10-20T09:30:00
2026-10-21T09:30:00
2026-10-22T09:30:00
2026-10-23T09:30:00
```

Combine a date field and a time field from a form, each parsed on its own:

```ant
func appointment(day: string, at: string): Result[Time.PlainDateTime, Error] {
  let date = try Time.parsePlainDate(day).wrapErr("date")
  let time = try Time.parsePlainTime(at).wrapErr("time")
  Ok{ value: Time.plainDateTime(date, time) }
}

println(appointment("2026-11-03", "15:45"))
println(appointment("2026-11-03", "3pm"))
```

```text
Ok{ value: 2026-11-03T15:45:00 }
Err{ error: Error{ message: "time: not a date or time: \"3pm\"" } }
```

Agree on a meeting by date and time, then place it in the office's zone to get a moment everyone can
convert:

```ant
let meeting = Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(15))
let stockholm = Time.zone("Europe/Stockholm").unwrap()
let start = meeting.toZoned(stockholm)
println(start)
println(start.withZone(Time.zone("America/New_York").unwrap()))
```

```text
2026-10-22T15:00:00+02:00[Europe/Stockholm]
2026-10-22T09:00:00-04:00[America/New_York]
```

## Properties and relations

[Time.toPlainDate](toPlainDate.md) and [Time.toPlainTime](toPlainTime.md) take the parts back:

```ant
let date = Time.plainDate(2026, 10, 22)
let time = Time.plainTime(14, 3, 5)
let t = Time.plainDateTime(date, time)
assert(t.toPlainDate() == date)
assert(t.toPlainTime() == time)
```

The result is the value that the date and time written with a `T` between them parse to:

```ant
let date = Time.plainDate(2026, 10, 22)
let time = Time.plainTime(23, 30)
let parsed = Time.parsePlainDateTime("${date}T${time}").unwrap()
assert(Time.plainDateTime(date, time) == parsed)
```

[Time.compare](compare.md) orders by the date first and the time second:

```ant
let late = Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(23))
let early = Time.plainDateTime(Time.plainDate(2026, 10, 23), Time.plainTime(1))
assert(Time.compare(late, early) == -1)
assert(late.until(early) == Time.duration(hours: 2))
```

## Possible issues

`Time.plainDateTime` takes a date and a time, not their fields:

```ant
let t = Time.plainDateTime(2026, 10, 22) // error: Time.plainDateTime takes 2 arguments, got 3
```

There is no hour 24. For the end of a day, use midnight of the next date:

```ant
Time.plainTime(24) // panic: Time.plainTime: hour out of range: 24
```

```ant
let date = Time.plainDate(2026, 12, 31)
Time.plainDateTime(date.add(Time.duration(days: 1)), Time.plainTime()) // 2027-01-01T00:00:00
```

The result has no zone, so it is not a moment and does not convert to an instant:

```ant
let t = Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(9))
let i = Time.toInstant(t) // error: cannot use Time.PlainDateTime where Time.ZonedDateTime is expected
```

A wall-clock time that a zone skips is still a valid `Time.PlainDateTime`. Stockholm moves from
02:00 to 03:00 on March 29, 2026, so 02:30 has no instant there, and `Time.toZoned` moves it
forward:

```ant
let t = Time.plainDateTime(Time.plainDate(2026, 3, 29), Time.plainTime(2, 30))
let stockholm = Time.zone("Europe/Stockholm").unwrap()
t.possibleInstants(stockholm) // []
t.toZoned(stockholm)          // 2026-03-29T03:30:00+02:00[Europe/Stockholm]
```

`<` does not apply to time values. Use `Time.compare`:

```ant
let a = Time.plainDateTime(Time.plainDate(2026, 1, 1), Time.plainTime())
let b = Time.plainDateTime(Time.plainDate(2026, 1, 2), Time.plainTime())
let first = a < b
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.PlainDateTime
```

## See also

[Time.plainDate](plainDate.md) · [Time.plainTime](plainTime.md) ·
[Time.parsePlainDateTime](parsePlainDateTime.md) · [Time.toPlainDateTime](toPlainDateTime.md) ·
[Time.toPlainDate](toPlainDate.md) · [Time.toPlainTime](toPlainTime.md) · [Time.toZoned](toZoned.md)
· [Time.PlainDateTime](types/PlainDateTime.md) · [Time](index.md)
