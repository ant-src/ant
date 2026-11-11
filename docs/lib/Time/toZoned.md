<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.toZoned

```ant signature
func Time.toZoned(t: Time.PlainDateTime, z: Time.Zone, disambiguation: Time.Disambiguation = Time.Compatible{}): Time.ZonedDateTime
```

Time.toZoned places the wall-clock time t in zone z. A time that a daylight-saving change skipped or
repeated is resolved by disambiguation, by default Time.Compatible, which moves a time in a gap
forward by the length of the gap and takes the earlier instant in an overlap;
[`Time.possibleInstants`](possibleInstants.md) shows the choices.

<!-- end generated -->

## Details

- A wall-clock time that a daylight-saving change skipped still gets an instant, so the result can
  show a different wall-clock time from `t`.
- In a zone with a fixed offset, such as UTC or `+05:30`, every wall-clock time has exactly one
  instant and `disambiguation` has no effect.
- It panics when the instant falls outside the range of time values.

## Basic examples

Place a wall-clock time in a zone:

```ant
let stockholm = Time.zone("Europe/Stockholm").unwrap()
let meeting = Time.parsePlainDateTime("2026-10-22T14:00").unwrap()
Time.toZoned(meeting, stockholm) // 2026-10-22T14:00:00+02:00[Europe/Stockholm]
```

On March 8, 2026, New York moves its clocks from 02:00 to 03:00, so 02:30 is skipped and the result
is an hour later on the clock:

```ant
let ny = Time.zone("America/New_York").unwrap()
let skipped = Time.parsePlainDateTime("2026-03-08T02:30").unwrap()
Time.toZoned(skipped, ny) // 2026-03-08T03:30:00-04:00[America/New_York]
```

## Scope

Call it through `.`. One wall-clock time in different zones is different moments:

```ant
let t = Time.parsePlainDateTime("2026-10-22T14:00").unwrap()
t.toZoned(Time.zone("Asia/Tokyo").unwrap()).toInstant()          // 2026-10-22T05:00:00Z
t.toZoned(Time.zone("America/Los_Angeles").unwrap()).toInstant() // 2026-10-22T21:00:00Z
t.toZoned(Time.zone("Asia/Kolkata").unwrap()).toInstant()        // 2026-10-22T08:30:00Z
```

On November 1, 2026, New York turns its clocks back from 02:00 to 01:00, so 01:30 happens twice. By
default the result is the first, still on daylight time:

```ant
let ny = Time.zone("America/New_York").unwrap()
let repeated = Time.parsePlainDateTime("2026-11-01T01:30").unwrap()
repeated.toZoned(ny) // 2026-11-01T01:30:00-04:00[America/New_York]
```

UTC and offset zones never skip or repeat a time, so the wall-clock time is kept as given:

```ant
let t = Time.parsePlainDateTime("2026-03-08T02:30").unwrap()
t.toZoned(Time.utc)                     // 2026-03-08T02:30:00+00:00[UTC]
t.toZoned(Time.zone("+05:30").unwrap()) // 2026-03-08T02:30:00+05:30[+05:30]
```

Lord Howe Island moves its clocks by half an hour, so a skipped time moves by half an hour:

```ant
let lordHowe = Time.zone("Australia/Lord_Howe").unwrap()
let t = Time.parsePlainDateTime("2026-10-04T02:15").unwrap()
t.toZoned(lordHowe) // 2026-10-04T02:45:00+11:00[Australia/Lord_Howe]
```

Build the wall-clock time from a date and a time first, and pipe it in:

```ant
let paris = Time.zone("Europe/Paris").unwrap()
Time.plainDateTime(Time.plainDate(2026, 12, 31), Time.plainTime(hour: 23, minute: 59))
  | Time.toZoned(paris) // 2026-12-31T23:59:00+01:00[Europe/Paris]
```

## Named arguments

`disambiguation:` picks the instant for a skipped or repeated time. `Time.Earlier{}` and
`Time.Later{}` take the earlier or the later of the two instants:

```ant
let ny = Time.zone("America/New_York").unwrap()
let repeated = Time.parsePlainDateTime("2026-11-01T01:30").unwrap()
repeated.toZoned(ny, disambiguation: Time.Earlier{}) // 2026-11-01T01:30:00-04:00[America/New_York]
repeated.toZoned(ny, disambiguation: Time.Later{})   // 2026-11-01T01:30:00-05:00[America/New_York]
```

In a gap, `Time.Earlier{}` moves the time back by the length of the gap and `Time.Later{}` moves it
forward. The default, `Time.Compatible{}`, moves it forward in a gap and takes the earlier instant
in an overlap:

```ant
let ny = Time.zone("America/New_York").unwrap()
let skipped = Time.parsePlainDateTime("2026-03-08T02:30").unwrap()
skipped.toZoned(ny, disambiguation: Time.Earlier{})    // 2026-03-08T01:30:00-05:00[America/New_York]
skipped.toZoned(ny, disambiguation: Time.Later{})      // 2026-03-08T03:30:00-04:00[America/New_York]
skipped.toZoned(ny, disambiguation: Time.Compatible{}) // 2026-03-08T03:30:00-04:00[America/New_York]
```

## Applications

Show when a meeting set in Stockholm time starts for each attendee:

```ant
let stockholm = Time.zone("Europe/Stockholm").unwrap()
let start = Time.parsePlainDateTime("2026-10-22T14:00").unwrap().toZoned(stockholm)

for name in ["Asia/Tokyo", "America/New_York", "Asia/Kolkata"] {
  let local = start.withZone(Time.zone(name).unwrap())
  println("${name}: ${local.toPlainTime()}")
}
```

```text
Asia/Tokyo: 21:00:00
America/New_York: 08:00:00
Asia/Kolkata: 17:30:00
```

Run a nightly job at 02:30 local time. On the night the clock skips 02:30, `Time.Later{}` runs it an
hour later on the clock, at 03:30, and on a night that repeats 02:30 it runs once, at the second
02:30:

```ant
let berlin = Time.zone("Europe/Berlin").unwrap()
let at = Time.plainTime(hour: 2, minute: 30)

for day in [28, 29, 30] {
  let date = Time.plainDate(2026, 3, day)
  let run = Time.plainDateTime(date, at).toZoned(berlin, disambiguation: Time.Later{})
  println(run)
}
```

```text
2026-03-28T02:30:00+01:00[Europe/Berlin]
2026-03-29T03:30:00+02:00[Europe/Berlin]
2026-03-30T02:30:00+02:00[Europe/Berlin]
```

Read a form whose date and time come in separate fields, and store the moment in UTC:

```ant
func moment(date: string, time: string, zone: string): Result[Time.Instant, Error] {
  let d = try (Time.parsePlainDate(date) | Result.wrapErr("date"))
  let t = try (Time.parsePlainTime(time) | Result.wrapErr("time"))
  let z = try (Time.zone(zone) | Result.wrapErr("zone"))
  Result.ok(Time.plainDateTime(d, t).toZoned(z).toInstant())
}

println(moment("2026-10-22", "09:30", "America/Chicago"))
println(moment("2026-10-22", "9.30", "America/Chicago"))
```

```text
Ok{ value: 2026-10-22T14:30:00Z }
Err{ error: Error{ message: "time: not a date or time: \"9.30\"" } }
```

## Properties and relations

Outside a gap, the result shows the wall-clock time it was given:

```ant
let ny = Time.zone("America/New_York").unwrap()
let t = Time.parsePlainDateTime("2026-11-01T01:30").unwrap()
assert(t.toZoned(ny).toPlainDateTime() == t)
assert(t.toZoned(ny, disambiguation: Time.Later{}).toPlainDateTime() == t)
```

The instants it can give are the ones [Time.possibleInstants](possibleInstants.md) lists:

```ant
let ny = Time.zone("America/New_York").unwrap()
let repeated = Time.parsePlainDateTime("2026-11-01T01:30").unwrap()
let both = repeated.possibleInstants(ny)
assert(repeated.toZoned(ny, disambiguation: Time.Earlier{}).toInstant() == both[0])
assert(repeated.toZoned(ny, disambiguation: Time.Later{}).toInstant() == both[1])
```

[Time.parseZoned](parseZoned.md) on a string with no offset gives the same value:

```ant
let ny = Time.zone("America/New_York").unwrap()
let t = Time.parsePlainDateTime("2026-03-08T02:30").unwrap()
assert(Time.parseZoned("2026-03-08T02:30[America/New_York]").unwrap() == t.toZoned(ny))
```

## Possible issues

A skipped time does not fail; it moves. Check that [Time.possibleInstants](possibleInstants.md) has
exactly one instant to refuse a skipped or repeated time:

```ant
let ny = Time.zone("America/New_York").unwrap()
let skipped = Time.parsePlainDateTime("2026-03-08T02:30").unwrap()
let repeated = Time.parsePlainDateTime("2026-11-01T01:30").unwrap()
skipped.toZoned(ny).toPlainDateTime() == skipped // false
skipped.possibleInstants(ny).isEmpty()           // true
repeated.possibleInstants(ny).length() != 1      // true
```

`toZoned` takes a wall-clock time. An instant is already a moment; see it in a zone with
[Time.inZone](inZone.md):

```ant
let at = Time.parseInstant("2026-10-22T14:00:00Z").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
let t = Time.toZoned(at, tokyo)
// error: cannot use Time.Instant where Time.PlainDateTime is expected
```

```ant
let at = Time.parseInstant("2026-10-22T14:00:00Z").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
at.inZone(tokyo) // 2026-10-22T23:00:00+09:00[Asia/Tokyo]
```

A date alone has no time of day to place. Pair it with a time first:

```ant
let date = Time.plainDate(2026, 10, 22)
let t = Time.toZoned(date, Time.utc)
// error: cannot use Time.PlainDate where Time.PlainDateTime is expected
```

```ant
let date = Time.plainDate(2026, 10, 22)
Time.plainDateTime(date, Time.plainTime(hour: 9)).toZoned(Time.utc) // 2026-10-22T09:00:00+00:00[UTC]
```

A zoned value moves to another zone with [Time.withZone](withZone.md), which keeps the moment.
Passing it to `toZoned` is a compile error:

```ant
let t = Time.parseZoned("2026-10-22T14:00:00+02:00[Europe/Stockholm]").unwrap()
let moved = Time.toZoned(t, Time.utc)
// error: cannot use Time.ZonedDateTime where Time.PlainDateTime is expected
```

A wall-clock time at the edge of the range can still fall outside it once the zone is applied:

```ant
let last = Time.plainDateTime(Time.plainDate(275760, 9, 13), Time.plainTime(hour: 1))
last.toZoned(Time.utc) // panic: Time.toZoned: outside the range of time values
```

## See also

[Time.possibleInstants](possibleInstants.md) · [Time.Disambiguation](types/Disambiguation.md) ·
[Time.inZone](inZone.md) · [Time.withZone](withZone.md) · [Time.parseZoned](parseZoned.md) ·
[Time.toPlainDateTime](toPlainDateTime.md) · [Time.plainDateTime](plainDateTime.md) ·
[Time.zone](zone.md) · [Time.PlainDateTime](types/PlainDateTime.md) ·
[Time.ZonedDateTime](types/ZonedDateTime.md)
