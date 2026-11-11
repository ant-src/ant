<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.parseZoned

```ant signature
func Time.parseZoned(s: string, disambiguation: Time.Disambiguation = Time.Compatible{}): Result[Time.ZonedDateTime, Error]
```

Time.parseZoned parses a zoned date-time from s, such as
2026-10-22T16:03:05+02:00\[Europe/Stockholm\]. The text needs a bracketed zone. Without an offset,
the wall-clock time is placed in the zone by disambiguation; with Z, the moment is taken and shown
in the zone; an offset the zone does not have at that time is an Err, as is text it cannot parse.

<!-- end generated -->

## Details

- The text needs a date and a zone in brackets. The time and the offset are optional: a date alone
  gives the start of that day in the zone, as [Time.startOfDay](startOfDay.md) does.
- With an offset, the offset must be the zone's at that time, so the text names one exact moment.
  Text with `Z` is taken as that moment, shown in the zone.
- Without an offset, `disambiguation` resolves a wall-clock time that a daylight-saving change
  skipped or repeated. The default, `Time.Compatible{}`, takes the later instant in a gap and the
  earlier one in an overlap.
- An unknown zone and text in another spelling, such as a lowercase `t`, are an `Err`.

## Basic examples

```ant
let sent = "2026-10-22T16:03:05+02:00[Europe/Stockholm]"
Time.parseZoned(sent) // Ok{ value: 2026-10-22T16:03:05+02:00[Europe/Stockholm] }
let wall = "2026-10-22T16:03:05[Europe/Stockholm]"
Time.parseZoned(wall) // Ok{ value: 2026-10-22T16:03:05+02:00[Europe/Stockholm] }
```

Text it cannot parse is an `Err` that says why:

```ant
let parsed = Time.parseZoned("2026-10-22T16:03:05+02:00")
if parsed is Err {
  println(parsed.error.message)
}
```

```text
a zoned date-time needs a date and a [zone]: "2026-10-22T16:03:05+02:00"
```

## Scope

With `Z`, the text names a moment, and the result shows it on the zone's wall clock:

```ant
let s = "2026-10-22T14:03:05Z[Europe/Stockholm]"
Time.parseZoned(s) // Ok{ value: 2026-10-22T16:03:05+02:00[Europe/Stockholm] }
```

A date alone is the first instant of that day. On a day whose midnight a daylight-saving change
skipped, that is not midnight:

```ant
Time.parseZoned("2026-10-22[Europe/Stockholm]").unwrap() // 2026-10-22T00:00:00+02:00[Europe/Stockholm]
Time.parseZoned("2024-09-08[America/Santiago]").unwrap() // 2024-09-08T01:00:00-03:00[America/Santiago]
```

The zone can be a fixed offset or `UTC`:

```ant
Time.parseZoned("2026-10-22T16:03:05+05:30[+05:30]") // Ok{ value: 2026-10-22T16:03:05+05:30[+05:30] }
Time.parseZoned("2026-10-22T16:03:05[UTC]")          // Ok{ value: 2026-10-22T16:03:05+00:00[UTC] }
```

A space can stand for the `T`, the fraction keeps up to nine digits, and the ISO calendar annotation
is accepted after the zone:

```ant
let s = "2026-10-22 16:03:05.123456789+02:00[Europe/Stockholm][u-ca=iso8601]"
Time.parseZoned(s).unwrap() // 2026-10-22T16:03:05.123456789+02:00[Europe/Stockholm]
```

A year outside 0 to 9999 takes a sign and six digits:

```ant
Time.parseZoned("+012026-10-22T16:03:05[UTC]") // Ok{ value: +012026-10-22T16:03:05+00:00[UTC] }
```

An unknown zone, another calendar, and other spellings of the format are an `Err`:

```ant
let texts = [
  "2026-10-22T16:03:05[Mars/Base]",
  "2026-10-22T16:03:05+02:00[Europe/Stockholm][u-ca=gregory]",
  "2026-10-22t16:03:05+02:00[Europe/Stockholm]",
]
for text in texts {
  let parsed = Time.parseZoned(text)
  if parsed is Err {
    println(parsed.error.message)
  }
}
```

```text
unknown time zone "Mars/Base"
calendar gregory is not supported; only iso8601 is
not a date or time: "2026-10-22t16:03:05+02:00[Europe/Stockholm]"
```

In a pipe, the text is the first argument:

```ant
"2026-10-22T16:03:05[Europe/Stockholm]" | Time.parseZoned() | Result.map({ x.hour() }) // Ok{ value: 16 }
```

## Named arguments

`disambiguation` picks the instant for a wall-clock time without an offset. On March 10, 2024, New
York skipped from 2:00 to 3:00, so 2:30 does not exist. `Time.Compatible{}` and `Time.Later{}` take
the later instant, 3:30 EDT, and `Time.Earlier{}` takes the earlier one, 1:30 EST:

```ant
let gap = "2024-03-10T02:30[America/New_York]"
Time.parseZoned(gap).unwrap()                                 // 2024-03-10T03:30:00-04:00[America/New_York]
Time.parseZoned(gap, disambiguation: Time.Earlier{}).unwrap() // 2024-03-10T01:30:00-05:00[America/New_York]
Time.parseZoned(gap, disambiguation: Time.Later{}).unwrap()   // 2024-03-10T03:30:00-04:00[America/New_York]
```

On November 3, 2024, New York repeated the hour from 1:00, so 1:30 happened twice. The default takes
the first:

```ant
let overlap = "2024-11-03T01:30[America/New_York]"
Time.parseZoned(overlap).unwrap()                               // 2024-11-03T01:30:00-04:00[America/New_York]
Time.parseZoned(overlap, disambiguation: Time.Later{}).unwrap() // 2024-11-03T01:30:00-05:00[America/New_York]
```

An offset in the text already names one instant, so `disambiguation` has no effect:

```ant
let exact = "2024-11-03T01:30-05:00[America/New_York]"
Time.parseZoned(exact, disambiguation: Time.Earlier{}).unwrap() // 2024-11-03T01:30:00-05:00[America/New_York]
```

## Applications

Read events stored with their zones and list them in the order they happened, not by their wall
clocks:

```ant
let events = [
  "2026-10-22T09:00:00-07:00[America/Los_Angeles]",
  "2026-10-22T17:30:00+02:00[Europe/Stockholm]",
  "2026-10-23T00:15:00+09:00[Asia/Tokyo]",
]

let ordered = events
  | Array.map({ Time.parseZoned(x).unwrap() })
  | Array.sortWith(Time.compare)

for e in ordered {
  println(e)
}
```

```text
2026-10-23T00:15:00+09:00[Asia/Tokyo]
2026-10-22T17:30:00+02:00[Europe/Stockholm]
2026-10-22T09:00:00-07:00[America/Los_Angeles]
```

Parse a meeting time from user input, say where it fails, and show it in the reader's zone:

```ant
func meetingIn(text: string, zone: Time.Zone): Result[string, Error] {
  let meeting = try (Time.parseZoned(text) | Result.wrapErr("meeting time"))
  let local = meeting.withZone(zone)
  Result.ok("${local.toPlainDate()} ${local.toPlainTime()} in ${zone.zoneId()}")
}

let tokyo = Time.zone("Asia/Tokyo").unwrap()
println(meetingIn("2026-10-22T15:00[Europe/Stockholm]", tokyo))
println(meetingIn("2026-10-22T15:00", tokyo))
```

```text
Ok{ value: "2026-10-22 22:00:00 in Asia/Tokyo" }
Err{ error: Error{ message: "meeting time: a zoned date-time needs a date and a [zone]: \"2026-10-22T15:00\"" } }
```

Refuse a wall-clock time that a daylight-saving change made ambiguous, instead of letting
`disambiguation` pick one:

```ant
func strictZoned(date: string, zone: string): Result[Time.ZonedDateTime, Error] {
  let wall = try Time.parsePlainDateTime(date)
  let z = try Time.zone(zone)
  let count = wall.possibleInstants(z).length()
  if count != 1 {
    return Result.fail("${date} has ${count} instants in ${zone}")
  }
  Time.parseZoned("${date}[${zone}]")
}

println(strictZoned("2024-11-03T01:30", "America/New_York"))
println(strictZoned("2024-11-03T03:30", "America/New_York"))
```

```text
Err{ error: Error{ message: "2024-11-03T01:30 has 2 instants in America/New_York" } }
Ok{ value: 2024-11-03T03:30:00-05:00[America/New_York] }
```

## Properties and relations

A `Time.ZonedDateTime` renders as text that parses back to the same value, even in a repeated hour,
since the text keeps the offset:

```ant
let s = "2024-11-03T01:30:00-05:00[America/New_York]"
let t = Time.parseZoned(s).unwrap()
assert(String.from(t) == s)
assert(Time.parseZoned(String.from(t)).unwrap() == t)
```

[Time.parseInstant](parseInstant.md) reads the same text as its instant and ignores the zone:

```ant
let s = "2026-10-22T16:03:05+02:00[Europe/Stockholm]"
assert(Time.parseInstant(s).unwrap() == Time.parseZoned(s).unwrap().toInstant())
```

Without an offset, it places the wall-clock time as [Time.toZoned](toZoned.md) does, with the same
`disambiguation`:

```ant
let ny = Time.zone("America/New_York").unwrap()
let wall = Time.parsePlainDateTime("2024-11-03T01:30").unwrap()
let text = "2024-11-03T01:30[America/New_York]"
assert(Time.parseZoned(text).unwrap() == wall.toZoned(ny))
let later = Time.parseZoned(text, disambiguation: Time.Later{}).unwrap()
assert(later == wall.toZoned(ny, disambiguation: Time.Later{}))
```

## Possible issues

An offset the zone does not have at that time is an `Err`, not a conversion. To convert, parse the
moment with [Time.parseInstant](parseInstant.md) and see it in the zone with
[Time.inZone](inZone.md):

```ant
let s = "2026-10-22T16:03:05+05:00[Europe/Stockholm]"
let stockholm = Time.zone("Europe/Stockholm").unwrap()
println(Time.parseZoned(s))
println(Time.parseInstant(s).unwrap().inZone(stockholm))
```

```text
Err{ error: Error{ message: "the offset +05:00 is not the offset of Europe/Stockholm at 2026-10-22T16:03:05" } }
2026-10-22T13:03:05+02:00[Europe/Stockholm]
```

Text without a bracketed zone is an `Err`, even with `Z` or an offset. Use
[Time.parseInstant](parseInstant.md) for it, and `unwrap` only text you know is valid:

```ant
Time.parseZoned("2026-10-22T14:03:05Z").unwrap()
// panic: Result.unwrap on an Err: Error{ message: "a zoned date-time needs a date and a [zone]: \"2026-10-22T14:03:05Z\"" }
```

A wall-clock time in a daylight-saving gap is moved, not refused, so the result can show another
hour than the text:

```ant
let t = Time.parseZoned("2024-03-10T02:30[America/New_York]").unwrap()
t.toPlainTime() // 03:30:00
```

A historical offset with seconds is rounded to the minute in the text, which still parses back to
the same value:

```ant
let old = Time.parseZoned("1880-01-01T00:00[Europe/Stockholm]").unwrap()
old.offset()                                      // "+00:53:28"
old                                               // 1880-01-01T00:00:00+00:53[Europe/Stockholm]
Time.parseZoned(String.from(old)).unwrap() == old // true
```

The offset in the text is `±HH:MM` only. One with seconds, as `offset` prints it, and one written
`±HH` are an `Err`, so build the text with `String.from(t)`, which rounds the offset to the minute,
not from `offset`:

```ant
let texts = ["1880-01-01T00:00+00:53:28[Europe/Stockholm]", "2026-10-22T16:03:05+02[Europe/Stockholm]"]
for text in texts {
  let parsed = Time.parseZoned(text)
  if parsed is Err {
    println(parsed.error.message)
  }
}
```

```text
not a date or time: "1880-01-01T00:00+00:53:28[Europe/Stockholm]"
not a date or time: "2026-10-22T16:03:05+02[Europe/Stockholm]"
```

## See also

[Time.parseInstant](parseInstant.md) · [Time.parsePlainDateTime](parsePlainDateTime.md) ·
[Time.toZoned](toZoned.md) · [Time.inZone](inZone.md) · [Time.possibleInstants](possibleInstants.md)
· [Time.startOfDay](startOfDay.md) · [Time.ZonedDateTime](types/ZonedDateTime.md) ·
[Time.Disambiguation](types/Disambiguation.md)
