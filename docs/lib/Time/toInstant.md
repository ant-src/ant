<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.toInstant

```ant signature
func Time.toInstant(t: Time.ZonedDateTime): Time.Instant
```

Time.toInstant returns the instant of t, without its zone.

<!-- end generated -->

## Details

- It never panics: every `Time.ZonedDateTime` has exactly one instant.
- To get the zoned value back, see the instant in its zone again with [Time.inZone](inZone.md) and
  [Time.zoneOf](zoneOf.md).

## Basic examples

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
Time.toInstant(t) // 2026-10-22T14:03:05Z
```

Call it through `.`:

```ant
let t = Time.parseZoned("2026-10-22T23:03:05+09:00[Asia/Tokyo]").unwrap()
t.toInstant() // 2026-10-22T14:03:05Z
```

## Scope

One moment seen in different zones gives one instant:

```ant
let stockholm = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let la = Time.parseZoned("2026-10-22T07:03:05-07:00[America/Los_Angeles]").unwrap()
stockholm.toInstant() // 2026-10-22T14:03:05Z
la.toInstant()        // 2026-10-22T14:03:05Z
```

A fixed-offset zone gives the same instant:

```ant
let offset = Time.parseZoned("2026-10-22T19:33:05+05:30[+05:30]").unwrap()
offset.toInstant() // 2026-10-22T14:03:05Z
```

When a daylight-saving change repeats an hour, one wall-clock time has two instants, and the offset
of the zoned value picks one:

```ant
let first = Time.parseZoned("2024-11-03T01:30:00-04:00[America/New_York]").unwrap()
let second = Time.parseZoned("2024-11-03T01:30:00-05:00[America/New_York]").unwrap()
first.toInstant()  // 2024-11-03T05:30:00Z
second.toInstant() // 2024-11-03T06:30:00Z
```

Nanoseconds are kept:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05.123456789+02:00[Europe/Stockholm]").unwrap()
t.toInstant() // 2026-10-22T14:03:05.123456789Z
```

In a pipe, pass it as a function value:

```ant
"2026-10-22T16:03:05+02:00[Europe/Stockholm]" | Time.parseZoned() | Result.map(Time.toInstant)
// Ok{ value: 2026-10-22T14:03:05Z }
```

## Applications

Store times in UTC, whatever zone they were entered in:

```ant
type Booking = { room: string, start: Time.Instant }

let start = Time.parseZoned("2026-10-22T09:30:00-04:00[America/New_York]").unwrap()
let booking = Booking{ room: "B2", start: start.toInstant() }
println(Json.stringify(booking))
```

```text
{"room":"B2","start":"2026-10-22T13:30:00Z"}
```

Drop duplicate reports of one event that servers in different zones logged:

```ant
let reports = [
  "2026-10-22T16:03:05+02:00[Europe/Stockholm]",
  "2026-10-22T23:03:05+09:00[Asia/Tokyo]",
  "2026-10-22T10:15:00-04:00[America/New_York]",
]

let events = reports
  | Array.map({ Time.parseZoned(x).unwrap() })
  | Array.uniqueBy(Time.toInstant)

for e in events {
  println(e.toInstant())
}
```

```text
2026-10-22T14:03:05Z
2026-10-22T14:15:00Z
```

## Properties and relations

Seeing the instant in the zone of the value gives the value back:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05.5+02:00[Europe/Stockholm]").unwrap()
assert(t.toInstant().inZone(t.zoneOf()) == t)
```

Moving a value to another zone with [Time.withZone](withZone.md) keeps its instant:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
assert(t.withZone(tokyo).toInstant() == t.toInstant())
```

The instant has the same epoch count, and [Time.compare](compare.md) orders zoned values as it
orders their instants:

```ant
let a = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let b = Time.parseZoned("2026-10-22T10:00:00-04:00[America/New_York]").unwrap()
assert(a.toInstant().epochNanos() == a.epochNanos())
assert(Time.compare(a, b) == Time.compare(a.toInstant(), b.toInstant()))
```

## Possible issues

An instant has no zone to drop, so passing one is a compile error:

```ant
let t = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
let i = Time.toInstant(t) // error: cannot use Time.Instant where Time.ZonedDateTime is expected
```

A wall-clock time names no instant until it has a zone. Place it in one with
[Time.toZoned](toZoned.md) first:

```ant
let meeting = Time.parsePlainDateTime("2026-10-22T14:00").unwrap()
let i = Time.toInstant(meeting)
// error: cannot use Time.PlainDateTime where Time.ZonedDateTime is expected
```

```ant
let meeting = Time.parsePlainDateTime("2026-10-22T14:00").unwrap()
let stockholm = Time.zone("Europe/Stockholm").unwrap()
meeting.toZoned(stockholm).toInstant() // 2026-10-22T12:00:00Z
```

`==` on zoned values compares the zone too, so one moment in two zones is not equal. Compare the
instants to ask whether two values are the same moment:

```ant
let stockholm = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let tokyo = Time.parseZoned("2026-10-22T23:03:05+09:00[Asia/Tokyo]").unwrap()
stockholm == tokyo                         // false
stockholm.toInstant() == tokyo.toInstant() // true
```

The instant forgets the zone, so seeing it in another zone does not give the value back:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
t.toInstant().inZone(Time.utc)      // 2026-10-22T14:03:05+00:00[UTC]
t.toInstant().inZone(Time.utc) == t // false
```

## See also

[Time.inZone](inZone.md) · [Time.withZone](withZone.md) · [Time.toZoned](toZoned.md) ·
[Time.toPlainDateTime](toPlainDateTime.md) · [Time.zoneOf](zoneOf.md) ·
[Time.epochNanos](epochNanos.md) · [Time.compare](compare.md) ·
[Time.ZonedDateTime](types/ZonedDateTime.md) · [Time.Instant](types/Instant.md)
