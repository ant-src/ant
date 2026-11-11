<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.parsePlainDateTime

```ant signature
func Time.parsePlainDateTime(s: string): Result[Time.PlainDateTime, Error]
```

Time.parsePlainDateTime parses a date and time from s, such as 2026-10-22T14:03:05, with T or a
space between them. It keeps the date and time of a longer form, which must still be well formed,
with a valid offset, though the name in a bracketed zone is not looked up. Z with no bracketed zone,
and text it cannot parse, are an Err.

<!-- end generated -->

## Details

- The text needs both a date and a time. Seconds and a fraction of up to nine digits are optional.
- An offset or a bracketed zone after the time is ignored, not applied: the result is the wall-clock
  time as written. `Z` alone is an `Err`, since it names a moment, not a wall-clock time.
- A date or time that does not exist, such as February 30, hour 24, or a leap second `:60`, is an
  `Err`.
- Other spellings are an `Err`: a lowercase `t`, a comma before the fraction, a time with no
  minutes, or a leading space.

## Basic examples

```ant
Time.parsePlainDateTime("2026-10-22T14:03:05") // Ok{ value: 2026-10-22T14:03:05 }
Time.parsePlainDateTime("2026-10-22T14:03")    // Ok{ value: 2026-10-22T14:03:00 }
Time.parsePlainDateTime("2026-10-22 14:03")    // Ok{ value: 2026-10-22T14:03:00 }
```

Text that is not a date and a time is an `Err`:

```ant
Time.parsePlainDateTime("tomorrow at noon") // Err{ error: Error{ message: "not a date or time: \"tomorrow at noon\"" } }
```

## Scope

A fraction of a second keeps up to nine digits, down to the nanosecond:

```ant
Time.parsePlainDateTime("2026-10-22T14:03:05.5")         // Ok{ value: 2026-10-22T14:03:05.5 }
Time.parsePlainDateTime("2026-10-22T14:03:05.123456789") // Ok{ value: 2026-10-22T14:03:05.123456789 }
```

A longer string gives its date and time and drops the rest. An offset or a zone does not shift the
result:

```ant
Time.parsePlainDateTime("2026-10-22T16:03:05+02:00")                   // Ok{ value: 2026-10-22T16:03:05 }
Time.parsePlainDateTime("2026-10-22T16:03:05+02:00[Europe/Stockholm]") // Ok{ value: 2026-10-22T16:03:05 }
Time.parsePlainDateTime("2026-10-22T14:03:05Z[UTC]")                   // Ok{ value: 2026-10-22T14:03:05 }
```

The ISO calendar annotation is accepted, and any other calendar is an `Err`:

```ant
Time.parsePlainDateTime("2026-10-22T14:03[u-ca=iso8601]") // Ok{ value: 2026-10-22T14:03:00 }
Time.parsePlainDateTime("2026-10-22T14:03[u-ca=gregory]") // Err{ error: Error{ message: "calendar gregory is not supported; only iso8601 is" } }
```

A year outside 0 to 9999 is written with a sign and six digits, the form it renders in:

```ant
Time.parsePlainDateTime("+012026-10-22T00:00") // Ok{ value: +012026-10-22T00:00:00 }
Time.parsePlainDateTime("-000001-01-01T00:00") // Ok{ value: -000001-01-01T00:00:00 }
```

A date or a time alone is not enough:

```ant
Time.parsePlainDateTime("2026-10-22") // Err{ error: Error{ message: "a date-time needs a date and a time: \"2026-10-22\"" } }
Time.parsePlainDateTime("14:03:05")   // Err{ error: Error{ message: "a date-time needs a date and a time: \"14:03:05\"" } }
```

A field out of range names the field:

```ant
Time.parsePlainDateTime("2026-02-30T10:00")    // Err{ error: Error{ message: "no such date: 2026-02-30" } }
Time.parsePlainDateTime("2026-10-22T24:00")    // Err{ error: Error{ message: "hour out of range: 24" } }
Time.parsePlainDateTime("2026-12-31T23:59:60") // Err{ error: Error{ message: "second out of range: 60" } }
```

## Applications

Read a time from user input and fall back to a default when it does not parse:

```ant
let fallback = Time.plainDateTime(Time.plainDate(2026, 11, 12), Time.plainTime(hour: 9))
Time.parsePlainDateTime("next week").unwrapOr(fallback) // 2026-11-12T09:00:00
```

Parse a line of a schedule, and say which entry failed:

```ant
type Event = { name: string, at: Time.PlainDateTime }

func parseEvent(line: string): Result[Event, Error] {
  let parts = line.split(",")
  if parts.length() != 2 {
    return Result.fail("expected name,time: ${line}")
  }
  let at = try (Time.parsePlainDateTime(parts[1].trim()) | Result.wrapErr(parts[0]))
  Result.ok(Event{ name: parts[0], at })
}

println(parseEvent("standup, 2026-11-12T09:30"))
println(parseEvent("review, 2026-11-31T15:00"))
```

```text
Ok{ value: Event{ name: "standup", at: 2026-11-12T09:30:00 } }
Err{ error: Error{ message: "review: no such date: 2026-11-31" } }
```

Sort parsed times with [`Time.compare`](compare.md):

```ant
["2026-11-12T09:30", "2026-11-03T18:00", "2026-11-12T08:00"]
  | Array.map({ Time.parsePlainDateTime(x).unwrap() })
  | Array.sortWith(Time.compare) // [2026-11-03T18:00:00, 2026-11-12T08:00:00, 2026-11-12T09:30:00]
```

Measure the time between two wall-clock readings:

```ant
let start = Time.parsePlainDateTime("2026-11-12T09:30").unwrap()
let end = Time.parsePlainDateTime("2026-11-13 17:15").unwrap()
start.until(end)                           // P1DT7H45M
start.until(end, largestUnit: Time.Hour{}) // PT31H45M
```

Place a meeting time, given without a zone, in the zone it happens in:

```ant
let stockholm = Time.zone("Europe/Stockholm").unwrap()
let meeting = Time.parsePlainDateTime("2026-11-12T14:00").unwrap()
meeting.toZoned(stockholm) // 2026-11-12T14:00:00+01:00[Europe/Stockholm]
```

## Properties and relations

A `Time.PlainDateTime` renders as text that parses back to it:

```ant
let t = Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(hour: 14, minute: 3, millisecond: 250))
assert(Time.parsePlainDateTime("${t}").unwrap() == t)
```

Parsing a date and time agrees with parsing each part and joining them:

```ant
let s = "2026-10-22T14:03:05"
let whole = Time.parsePlainDateTime(s).unwrap()
let date = Time.parsePlainDate(s).unwrap()
let time = Time.parsePlainTime(s).unwrap()
assert(whole == Time.plainDateTime(date, time))
```

The plain date-time of a zoned string is its wall-clock time, as
[`Time.toPlainDateTime`](toPlainDateTime.md) gives it:

```ant
let s = "2026-10-22T16:03:05+02:00[Europe/Stockholm]"
let zoned = Time.parseZoned(s).unwrap()
assert(Time.parsePlainDateTime(s).unwrap() == zoned.toPlainDateTime())
```

## Possible issues

An offset is dropped, not applied. To get the moment the text names, use
[`Time.parseInstant`](parseInstant.md):

```ant
let s = "2026-10-22T16:03:05+02:00"
Time.parsePlainDateTime(s) // Ok{ value: 2026-10-22T16:03:05 }
Time.parseInstant(s)       // Ok{ value: 2026-10-22T14:03:05Z }
```

`Z` without a bracketed zone is an `Err`, since it names a moment. Parse it with
`Time.parseInstant`, then choose the zone to see it in:

```ant
let s = "2026-10-22T14:03:05Z"
Time.parsePlainDateTime(s)                                         // Err{ error: Error{ message: "Z names an exact time; a plain value needs a wall-clock time" } }
Time.parseInstant(s).map({ x.inZone(Time.utc).toPlainDateTime() }) // Ok{ value: 2026-10-22T14:03:05 }
```

Close spellings that ISO 8601 or other libraries allow are an `Err`:

```ant
Time.parsePlainDateTime("2026-10-22t14:03")      // Err{ error: Error{ message: "not a date or time: \"2026-10-22t14:03\"" } }
Time.parsePlainDateTime("2026-10-22T14:03:05,5") // Err{ error: Error{ message: "not a date or time: \"2026-10-22T14:03:05,5\"" } }
Time.parsePlainDateTime("2026-10-22T14")         // Err{ error: Error{ message: "not a date or time: \"2026-10-22T14\"" } }
Time.parsePlainDateTime(" 2026-10-22T14:03")     // Err{ error: Error{ message: "not a date or time: \" 2026-10-22T14:03\"" } }
```

Trim text from a file or a form before parsing it:

```ant
Time.parsePlainDateTime(" 2026-10-22T14:03\n".trim()) // Ok{ value: 2026-10-22T14:03:00 }
```

## See also

[Time.plainDateTime](plainDateTime.md) · [Time.parsePlainDate](parsePlainDate.md) ·
[Time.parsePlainTime](parsePlainTime.md) · [Time.parseZoned](parseZoned.md) ·
[Time.parseInstant](parseInstant.md) · [Time.toZoned](toZoned.md) ·
[Time.PlainDateTime](types/PlainDateTime.md)
