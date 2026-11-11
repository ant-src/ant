<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.parsePlainTime

```ant signature
func Time.parsePlainTime(s: string): Result[Time.PlainTime, Error]
```

Time.parsePlainTime parses a time of day from s, such as 14:03:05.123. It takes HH:MM, HH:MM:SS or
up to nine digits of fraction, alone or after a date, and keeps the time of a longer form; the whole
string must still be well formed, with a date that exists and a valid offset, though the name in a
bracketed zone is not looked up. Z with no bracketed zone, and text it cannot parse, are an Err.

<!-- end generated -->

## Details

- The hour and the minute take two digits each, and the seconds and a fraction after a point are
  optional. Missing seconds are 0.
- After a date and `T` or a space, it takes the time and ignores the offset and the bracketed zone.
  The date must still exist.
- A string with `Z` and no bracketed zone is an `Err`: `Z` names a moment, not a wall-clock time.
- It never panics. A malformed string, an hour of 24 or a leap second `:60` is an `Err`.

## Basic examples

```ant
Time.parsePlainTime("14:03:05") // Ok{ value: 14:03:05 }
Time.parsePlainTime("09:30")    // Ok{ value: 09:30:00 }
```

Text that is not a time is an `Err`:

```ant
Time.parsePlainTime("noon") // Err{ error: Error{ message: "not a date or time: \"noon\"" } }
```

## Scope

A fraction of a second takes up to nine digits, down to the nanosecond. The time renders its
fraction to the last nonzero digit:

```ant
Time.parsePlainTime("14:03:05.123")       // Ok{ value: 14:03:05.123 }
Time.parsePlainTime("14:03:05.120")       // Ok{ value: 14:03:05.12 }
Time.parsePlainTime("14:03:05.000")       // Ok{ value: 14:03:05 }
Time.parsePlainTime("23:59:59.999999999") // Ok{ value: 23:59:59.999999999 }
```

A date and time gives its time, with `T` or a space between them:

```ant
Time.parsePlainTime("2026-10-22T14:03:05") // Ok{ value: 14:03:05 }
Time.parsePlainTime("2026-10-22 08:15")    // Ok{ value: 08:15:00 }
```

A full zoned string gives the wall-clock time it shows, and an offset alone is ignored too:

```ant
Time.parsePlainTime("2026-10-22T16:03:05+02:00[Europe/Stockholm]") // Ok{ value: 16:03:05 }
Time.parsePlainTime("2026-10-22T16:03:05+02:00")                   // Ok{ value: 16:03:05 }
Time.parsePlainTime("2026-10-22T14:03:05Z[UTC]")                   // Ok{ value: 14:03:05 }
```

In a pipe, the string is the first argument:

```ant
"07:05" | Time.parsePlainTime() | Result.map(Time.minute) // Ok{ value: 5 }
```

## Applications

Read opening hours from a config value, with `try` to return the first bad time:

```ant
type Hours = { open: Time.PlainTime, close: Time.PlainTime }

func parseHours(s: string): Result[Hours, Error] {
  let parts = s.split("-")
  if parts.length() != 2 {
    return Result.fail("want open-close, got ${s}")
  }
  let open = try Time.parsePlainTime(parts[0]).wrapErr("opening time")
  let close = try Time.parsePlainTime(parts[1]).wrapErr("closing time")
  Result.ok(Hours{ open, close })
}

println(parseHours("09:00-17:30"))
println(parseHours("09:00-25:00"))
println(parseHours("9 to 5"))
```

```text
Ok{ value: Hours{ open: 09:00:00, close: 17:30:00 } }
Err{ error: Error{ message: "closing time: hour out of range: 25" } }
Err{ error: Error{ message: "want open-close, got 9 to 5" } }
```

Sort the times of a schedule written as text:

```ant
let alarms = ["07:30", "06:45", "22:00", "12:15"]
let sorted = alarms
  | Array.map({ Time.parsePlainTime(x).unwrap() })
  | Array.sortWith(Time.compare)
println(sorted)
```

```text
[06:45:00, 07:30:00, 12:15:00, 22:00:00]
```

Keep only the entries whose time parses, and fall back to midnight for the rest:

```ant
let entries = ["08:00", "lunch", "13:30", "24:00"]
let valid = entries.filter({ Time.parsePlainTime(x).isOk() })
println(valid)
let times = entries.map({ Time.parsePlainTime(x).unwrapOr(Time.plainTime()) })
println(times)
```

```text
["08:00", "13:30"]
[08:00:00, 00:00:00, 13:30:00, 00:00:00]
```

Measure the length of a shift given as two times:

```ant
let start = Time.parsePlainTime("08:30").unwrap()
let end = Time.parsePlainTime("17:15").unwrap()
println(Time.until(start, end))
```

```text
PT8H45M
```

## Properties and relations

A time renders as text that parses back to the same time:

```ant
let times = [Time.plainTime(), Time.plainTime(14, 3, 5, 123), Time.plainTime(23, 59, nanosecond: 1)]
for t in times {
  assert(Time.parsePlainTime(String.from(t)).unwrap() == t)
}
```

Parsing a date and time and taking its time with [Time.toPlainTime](toPlainTime.md) gives the same
value as parsing the time directly:

```ant
let s = "2026-10-22T14:03:05.5"
let viaDateTime = Time.parsePlainDateTime(s).unwrap().toPlainTime()
assert(Time.parsePlainTime(s).unwrap() == viaDateTime)
```

The parsed fields are the ones written:

```ant
let t = Time.parsePlainTime("14:03:05.123456789").unwrap()
assert(t.hour() == 14 && t.minute() == 3 && t.second() == 5)
assert(t.millisecond() == 123 && t.microsecond() == 456 && t.nanosecond() == 789)
```

## Possible issues

Each field takes exactly two digits, so a single-digit hour is an `Err`:

```ant
Time.parsePlainTime("9:15")  // Err{ error: Error{ message: "not a date or time: \"9:15\"" } }
Time.parsePlainTime("09:15") // Ok{ value: 09:15:00 }
```

The parser does not trim spaces, so trim the input first:

```ant
Time.parsePlainTime(" 14:03")        // Err{ error: Error{ message: "not a date or time: \" 14:03\"" } }
Time.parsePlainTime(" 14:03".trim()) // Ok{ value: 14:03:00 }
```

`Time.parsePlainTime` takes a string but lives in the `Time` namespace, so a dot call on a string
does not find it. Call it by its namespace or in a pipe:

```ant
let t = "18:45".parsePlainTime() // error: string has no field or function parsePlainTime
```

Midnight is `00:00`. An hour of 24 and a leap second are out of range:

```ant
Time.parsePlainTime("24:00")    // Err{ error: Error{ message: "hour out of range: 24" } }
Time.parsePlainTime("23:59:60") // Err{ error: Error{ message: "second out of range: 60" } }
```

A string with `Z` and no bracketed zone names a moment, not a wall-clock time, so it is an `Err`:

```ant
Time.parsePlainTime("2026-10-22T14:03:05Z")
// Err{ error: Error{ message: "Z names an exact time; a plain value needs a wall-clock time" } }
```

Parse it with [Time.parseInstant](parseInstant.md) and put it in a zone to read its time there:

```ant
let instant = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
instant.inZone(Time.utc).toPlainTime() // 14:03:05
```

The date before the time must exist, though the result has no date:

```ant
Time.parsePlainTime("2026-02-30T10:00") // Err{ error: Error{ message: "no such date: 2026-02-30" } }
```

A date alone, a time with `T` in front, more than nine digits of fraction, a comma before the
fraction, and trailing text are all an `Err`:

```ant
Time.parsePlainTime("2026-10-22").isOk()          // false
Time.parsePlainTime("T14:03").isOk()              // false
Time.parsePlainTime("14:03:05.1234567890").isOk() // false
Time.parsePlainTime("14:03:05,5").isOk()          // false
Time.parsePlainTime("14:03 pm").isOk()            // false
```

## See also

[Time.plainTime](plainTime.md) · [Time.parsePlainDateTime](parsePlainDateTime.md) ·
[Time.parsePlainDate](parsePlainDate.md) · [Time.parseInstant](parseInstant.md) ·
[Time.parseZoned](parseZoned.md) · [Time.toPlainTime](toPlainTime.md) ·
[Time.PlainTime](types/PlainTime.md) · [Time](index.md)
