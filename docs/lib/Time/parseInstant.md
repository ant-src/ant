<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.parseInstant

```ant signature
func Time.parseInstant(s: string): Result[Time.Instant, Error]
```

Time.parseInstant parses an instant from s, such as 2026-10-22T14:03:05Z. The text needs a date, a
time, and Z or an offset; a bracketed zone is ignored. Text it cannot parse is an Err.

<!-- end generated -->

## Details

- Without `Z` or an offset the text names no moment, so a date, or a date and time alone, is an
  `Err`; parse it with [Time.parsePlainDateTime](parsePlainDateTime.md).
- The offset only places the moment: the result is an instant, with no zone, and renders in UTC.
- To keep the bracketed zone, use [Time.parseZoned](parseZoned.md).

## Basic examples

```ant
Time.parseInstant("2026-10-22T14:03:05Z")      // Ok{ value: 2026-10-22T14:03:05Z }
Time.parseInstant("2026-10-22T16:03:05+02:00") // Ok{ value: 2026-10-22T14:03:05Z }
```

Text it cannot parse is an `Err` with a message:

```ant
Time.parseInstant("yesterday")
// Err{ error: Error{ message: "not a date or time: \"yesterday\"" } }
```

## Scope

The seconds are optional, and a fraction keeps up to nine digits, written back to its last nonzero
digit:

```ant
Time.parseInstant("2026-10-22T14:03Z")              // Ok{ value: 2026-10-22T14:03:00Z }
Time.parseInstant("2026-10-22T14:03:05.500Z")       // Ok{ value: 2026-10-22T14:03:05.5Z }
Time.parseInstant("2026-10-22T14:03:05.123456789Z") // Ok{ value: 2026-10-22T14:03:05.123456789Z }
```

A space may separate the date and the time:

```ant
Time.parseInstant("2026-10-22 14:03:05Z") // Ok{ value: 2026-10-22T14:03:05Z }
```

A negative offset, or one that is not a whole hour, moves the moment the other way:

```ant
Time.parseInstant("2026-10-22T09:03:05-05:00") // Ok{ value: 2026-10-22T14:03:05Z }
Time.parseInstant("2026-10-22T19:48:05+05:45") // Ok{ value: 2026-10-22T14:03:05Z }
```

An offset can carry the moment into another day:

```ant
Time.parseInstant("2026-10-23T01:30:00+11:00") // Ok{ value: 2026-10-22T14:30:00Z }
```

A bracketed zone and the calendar annotation are read and ignored; the offset alone decides the
moment:

```ant
Time.parseInstant("2026-10-22T16:03:05+02:00[Europe/Stockholm]") // Ok{ value: 2026-10-22T14:03:05Z }
Time.parseInstant("2026-10-22T14:03:05Z[u-ca=iso8601]")          // Ok{ value: 2026-10-22T14:03:05Z }
```

A year outside 0 to 9999 takes a sign and six digits, the form such an instant renders in:

```ant
Time.parseInstant("+012026-10-22T14:03:05Z") // Ok{ value: +012026-10-22T14:03:05Z }
Time.parseInstant("-000001-01-01T00:00:00Z") // Ok{ value: -000001-01-01T00:00:00Z }
```

The result can be piped on:

```ant
let stamp = "2026-10-22T14:03:05Z"
stamp | Time.parseInstant() | Result.map({ x.epochMillis() }) // Ok{ value: 1792677785000 }
```

## Named arguments

The parameter is named `s`:

```ant
Time.parseInstant(s: "1970-01-01T00:00:00Z") // Ok{ value: 1970-01-01T00:00:00Z }
```

## Applications

Read the timestamps of a log and report how long the job took, failing with the line that is wrong:

```ant
func elapsed(lines: Array[string]): Result[Time.Duration, Error] {
  if lines.length() < 2 {
    return Result.fail("need a start and an end")
  }
  let start = try (Time.parseInstant(lines[0]) | Result.wrapErr("first line"))
  let end = try (Time.parseInstant(lines[lines.length() - 1]) | Result.wrapErr("last line"))
  Result.ok(start.until(end))
}

println(elapsed(["2026-10-22T14:03:05Z", "2026-10-22T16:05:35.5+02:00"]))
println(elapsed(["2026-10-22T14:03:05Z", "2026-10-22T14:05:35"]))
```

```text
Ok{ value: PT150.5S }
Err{ error: Error{ message: "last line: an instant needs a date, a time and an offset: \"2026-10-22T14:05:35\"" } }
```

Order timestamps written with different offsets by the moment they name. Sorting the strings would
put them in the wrong order:

```ant
let stamps = ["2026-10-22T16:00:00+02:00", "2026-10-22T14:30:00Z", "2026-10-22T09:15:00-05:00"]
let ordered = stamps.sortBy({ Time.parseInstant(x).unwrap().epochNanos() })
for s in ordered {
  println(s)
}
```

```text
2026-10-22T16:00:00+02:00
2026-10-22T09:15:00-05:00
2026-10-22T14:30:00Z
```

Keep the valid timestamps of user input and drop the rest:

```ant
let input = ["2026-10-22T14:03:05Z", "not a time", "2026-10-22T14:03:05", "2026-10-23T08:00:00+01:00"]
let valid = input.filter({ Time.parseInstant(x).isOk() })
for s in valid {
  println(s)
}
```

```text
2026-10-22T14:03:05Z
2026-10-23T08:00:00+01:00
```

## Properties and relations

Every instant renders as text that parses back to it:

```ant
for ms in [0, -1, 1_792_677_785_500, 8_640_000_000_000_000] {
  let t = Time.fromEpochMillis(ms)
  assert(Time.parseInstant(String.from(t)).unwrap() == t)
}
```

The same moment written with any offset parses to the same instant:

```ant
let utc = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
assert(Time.parseInstant("2026-10-22T16:03:05+02:00").unwrap() == utc)
assert(Time.parseInstant("2026-10-22T05:03:05-09:00").unwrap() == utc)
```

Parsing a zoned string and taking its instant gives what `Time.parseInstant` gives:

```ant
let text = "2026-10-22T16:03:05+02:00[Europe/Stockholm]"
assert(Time.parseZoned(text).unwrap().toInstant() == Time.parseInstant(text).unwrap())
```

## Possible issues

Text without an offset is a wall-clock time, not a moment, and is an `Err`. Parse it with
[Time.parsePlainDateTime](parsePlainDateTime.md) and place it in a zone with
[Time.toZoned](toZoned.md):

```ant
Time.parseInstant("2026-10-22T14:03:05")
// Err{ error: Error{ message: "an instant needs a date, a time and an offset: \"2026-10-22T14:03:05\"" } }
Time.parseInstant("2026-10-22")
// Err{ error: Error{ message: "an instant needs a date, a time and an offset: \"2026-10-22\"" } }
```

Spellings that other ISO 8601 parsers accept are an `Err`: a lowercase `z`, a comma before the
fraction, a short offset, and white space around the text:

```ant
Time.parseInstant("2026-10-22T14:03:05z").isOk()     // false
Time.parseInstant("2026-10-22T14:03:05,5Z").isOk()   // false
Time.parseInstant("2026-10-22T16:03:05+0200").isOk() // false
Time.parseInstant(" 2026-10-22T14:03:05Z").isOk()    // false
```

Trim text from a file or a form first:

```ant
" 2026-10-22T14:03:05Z\n".trim() | Time.parseInstant() // Ok{ value: 2026-10-22T14:03:05Z }
```

A field out of range is an `Err`, and so is a leap second:

```ant
Time.parseInstant("2026-02-30T00:00:00Z") // Err{ error: Error{ message: "no such date: 2026-02-30" } }
Time.parseInstant("2026-12-31T23:59:60Z") // Err{ error: Error{ message: "second out of range: 60" } }
```

The zone name in brackets is not looked up, so it does not confirm the offset. A zone that disagrees
with the offset still parses, by the offset. Use [Time.parseZoned](parseZoned.md) when the zone
matters. The brackets are still read: a critical flag, or any annotation other than
`[u-ca=iso8601]`, is an `Err`:

```ant
let text = "2026-10-22T16:03:05+02:00[Asia/Tokyo]"
Time.parseInstant(text)                                                  // Ok{ value: 2026-10-22T14:03:05Z }
Time.parseZoned(text).isOk()                                             // false
Time.parseInstant("2026-10-22T16:03:05+02:00[!Europe/Stockholm]").isOk() // false
```

The result is a `Result`, not an instant. Unwrap it, or use `try`, before passing it on:

```ant
let t = Time.parseInstant("2026-10-22T14:03:05Z")
let ms = t.epochMillis()
// error: Result[Time.Instant, Error] has no field or function epochMillis
```

## See also

[Time.parseZoned](parseZoned.md) · [Time.parsePlainDateTime](parsePlainDateTime.md) ·
[Time.parsePlainDate](parsePlainDate.md) · [Time.parseDuration](parseDuration.md) ·
[Time.inZone](inZone.md) · [Time.fromEpochMillis](fromEpochMillis.md) ·
[Time.Instant](types/Instant.md)
