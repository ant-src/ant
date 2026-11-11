<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.plainTime

```ant signature
func Time.plainTime(hour: int = 0, minute: int = 0, second: int = 0, millisecond: int = 0, microsecond: int = 0, nanosecond: int = 0): Time.PlainTime
```

Time.plainTime returns the time of day with the given fields, each 0 when omitted. It panics if a
field is out of range, such as hour 24 or minute 60.

<!-- end generated -->

## Details

- The hour is 0 to 23, the minute and the second 0 to 59, and each subsecond field 0 to 999. A field
  outside its range panics; nothing carries into the next field, so `minute: 90` panics rather than
  giving 01:30.
- A `Time.PlainTime` has no date and no zone. Combine it with a date in
  [Time.plainDateTime](plainDateTime.md), and that with a zone in [Time.toZoned](toZoned.md).
- To build a time from text, use [Time.parsePlainTime](parsePlainTime.md), which returns a `Result`
  instead of panicking.

## Basic examples

```ant
Time.plainTime(14, 3, 5) // 14:03:05
Time.plainTime(9, 30)    // 09:30:00
Time.plainTime()         // 00:00:00
```

A fractional second renders to its last nonzero digit:

```ant
Time.plainTime(14, 3, 5, 500)             // 14:03:05.5
Time.plainTime(23, 59, 59, 999, 999, 999) // 23:59:59.999999999
```

## Scope

Each subsecond field adds three digits:

```ant
Time.plainTime(second: 1, millisecond: 250) // 00:00:01.25
Time.plainTime(microsecond: 5)              // 00:00:00.000005
Time.plainTime(nanosecond: 1)               // 00:00:00.000000001
```

The result renders as text in interpolation and `String.from`:

```ant
let alarm = Time.plainTime(6, 45)
"Alarm at ${alarm}" // "Alarm at 06:45:00"
String.from(alarm)  // "06:45:00"
```

## Named arguments

Name the fields to set some and leave the ones before them at 0:

```ant
Time.plainTime(minute: 15)           // 00:15:00
Time.plainTime(hour: 7, second: 30)  // 07:00:30
Time.plainTime(12, millisecond: 120) // 12:00:00.12
```

## Applications

Build the half-hour slots of a morning:

```ant
let start = Time.plainTime(9)
let slots = Array.range(0, 4).map({ start.add(Time.duration(minutes: 30 * x)) })
println(slots)
```

```text
[09:00:00, 09:30:00, 10:00:00, 10:30:00]
```

Turn a count of minutes since midnight into a time, wrapping past the end of the day:

```ant
func fromMinutes(total: int): Time.PlainTime {
  Time.plainTime(total / 60 % 24, total % 60)
}

fromMinutes(510)  // 08:30:00
fromMinutes(1500) // 01:00:00
```

Check numbers from the outside world before constructing, so a bad field is an error and not a
panic:

```ant
func clockTime(h: int, m: int): Result[Time.PlainTime, Error] {
  if h < 0 || h > 23 {
    return Result.fail("hour ${h} is not 0 to 23")
  }
  if m < 0 || m > 59 {
    return Result.fail("minute ${m} is not 0 to 59")
  }
  Result.ok(Time.plainTime(h, m))
}

clockTime(7, 45) // Ok{ value: 07:45:00 }
clockTime(25, 0) // Err{ error: Error{ message: "hour 25 is not 0 to 23" } }
```

Put a time on a date and in a zone:

```ant
let meeting = Time.plainDateTime(Time.plainDate(2026, 11, 11), Time.plainTime(10, 30))
let stockholm = Time.zone("Europe/Stockholm").unwrap()
println(meeting)
println(meeting.toZoned(stockholm))
```

```text
2026-11-11T10:30:00
2026-11-11T10:30:00+01:00[Europe/Stockholm]
```

Check the current time of day against an opening time. [Time.toPlainTime](toPlainTime.md) takes the
time of day from a clock reading, and the output depends on when it runs:

<!-- test: check-only output changes between runs -->

```ant
let opens = Time.plainTime(9, 30)
let now = Time.nowIn(Time.localZone()).toPlainTime()
println(if Time.compare(now, opens) < 0 { "closed until ${opens}" } else { "open since ${opens}" })
```

Measure the length of a working day:

```ant
Time.plainTime(9).until(Time.plainTime(17, 30)) // PT8H30M
```

## Properties and relations

The fields read back what was given, and the rendered text parses back to the same time:

```ant
let t = Time.plainTime(14, 3, 5, 120)
assert(t.hour() == 14 && t.minute() == 3 && t.second() == 5)
assert(t.millisecond() == 120)
assert(Time.parsePlainTime(String.from(t)).unwrap() == t)
```

Omitted fields are 0, and two times with the same fields are `==`:

```ant
assert(Time.plainTime() == Time.plainTime(0, 0, 0, 0, 0, 0))
assert(Time.plainTime(9, 30) == Time.plainTime(minute: 30, hour: 9))
```

## Possible issues

A field out of range panics; an hour of 24 is not midnight:

```ant
Time.plainTime(24)                // panic: Time.plainTime: hour out of range: 24
Time.plainTime(0, 90)             // panic: Time.plainTime: minute out of range: 90
Time.plainTime(12, minute: -5)    // panic: Time.plainTime: minute out of range: -5
Time.plainTime(millisecond: 1000) // panic: Time.plainTime: millisecond out of range: 1000
```

[Time.with](with.md) constrains a field instead of panicking:

```ant
Time.plainTime(9).with(minute: 90) // 09:59:00
```

`<` does not apply to time values; order times with [Time.compare](compare.md):

```ant
let early = Time.plainTime(8)
let late = Time.plainTime(9)
println(early < late)
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.PlainTime; narrow first
```

```ant
let times = [Time.plainTime(17), Time.plainTime(8, 30), Time.plainTime(12)]
times.sortWith(Time.compare) // [08:30:00, 12:00:00, 17:00:00]
```

A `Time.PlainTime` has no date, so the time from 23:00 until 01:00 is 22 hours back, not two hours
forward:

```ant
Time.plainTime(23).until(Time.plainTime(1)) // -PT22H
```

## See also

[Time.parsePlainTime](parsePlainTime.md) · [Time.toPlainTime](toPlainTime.md) ·
[Time.plainDate](plainDate.md) · [Time.plainDateTime](plainDateTime.md) · [Time.with](with.md) ·
[Time.hour](hour.md) · [Time.compare](compare.md) · [Time.types.PlainTime](types/PlainTime.md) ·
[Time](index.md)
