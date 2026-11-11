<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.ZonedDateTime

```ant signature
type Time.ZonedDateTime
```

Time.ZonedDateTime is an instant seen in a zone, the only type that knows both the moment and the
wall-clock time there. It renders with its offset and zone both, as
2024-11-03T01:00:00-04:00\[America/New\_York\], so the moment survives a round trip even where the
wall-clock time is ambiguous; two values are == when they are the same moment in the same zone.

<!-- end generated -->

## Details

- A `Time.ZonedDateTime` holds an instant and a zone. [Time.toInstant](../toInstant.md) and
  [Time.zoneOf](../zoneOf.md) take them apart; the wall-clock fields and the offset follow from the
  two.
- `==` needs the same moment and the same zone. [Time.compare](../compare.md), not `<`, orders by
  moment alone, so one moment seen in two zones compares as `0` but is not `==`.
- In [Time.add](../add.md), years, months, weeks and days move the wall clock, so a day across a
  daylight-saving change lasts as long as the zone's day does (23 or 25 hours in New York). Hours
  and smaller move the instant. [Time.until](../until.md) counts days in the zone's real days.

## Basic examples

Parse a `Time.ZonedDateTime` from its text with [Time.parseZoned](../parseZoned.md). The text has a
zone in brackets, and here an offset too:

```ant
Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]")
// Ok{ value: 2026-10-22T16:03:05+02:00[Europe/Stockholm] }
```

See an instant in a zone with [Time.inZone](../inZone.md), or place a wall-clock time in a zone with
[Time.toZoned](../toZoned.md):

```ant
let stockholm = Time.zone("Europe/Stockholm").unwrap()
let t = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
let meeting = Time.parsePlainDateTime("2026-10-22T09:30").unwrap()
Time.inZone(t, stockholm)        // 2026-10-22T16:03:05+02:00[Europe/Stockholm]
Time.toZoned(meeting, stockholm) // 2026-10-22T09:30:00+02:00[Europe/Stockholm]
```

[Time.nowIn](../nowIn.md) gives the current time in a zone:

```ant
let ny = Time.zone("America/New_York").unwrap()
let here: Time.ZonedDateTime = Time.nowIn(ny)
println(here.zoneOf())
```

```text
America/New_York
```

## Scope

A zoned value has every field of a date and a time, and an offset:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
t.year()          // 2026
t.dayOfWeek()     // 4
t.hour()          // 16
t.offset()        // "+02:00"
t.offsetSeconds() // 7200
t.zoneOf()        // Europe/Stockholm
```

[Time.startOfDay](../startOfDay.md) gives the first instant of its date, and [Time.with](../with.md)
moves it to another wall-clock time, keeping the offset when the zone still has it there:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
t.startOfDay()                        // 2026-10-22T00:00:00+02:00[Europe/Stockholm]
t.with(hour: 9, minute: 0, second: 0) // 2026-10-22T09:00:00+02:00[Europe/Stockholm]
```

It converts to the instant, to the plain types, and to the same instant in another zone:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let ny = Time.zone("America/New_York").unwrap()
t.toInstant()       // 2026-10-22T14:03:05Z
t.toPlainDate()     // 2026-10-22
t.toPlainTime()     // 16:03:05
t.toPlainDateTime() // 2026-10-22T16:03:05
t.withZone(ny)      // 2026-10-22T10:03:05-04:00[America/New_York]
```

The zone may be a fixed offset, which never changes, so it has no transitions:

```ant
let fixed = Time.parseZoned("2026-10-31T12:00+05:30[+05:30]").unwrap()
fixed                  // 2026-10-31T12:00:00+05:30[+05:30]
fixed.nextTransition() // null
```

String interpolation, `String.from` and `Json.stringify` write the same text:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
"starts ${t}"       // "starts 2026-10-22T16:03:05+02:00[Europe/Stockholm]"
String.from(t)      // "2026-10-22T16:03:05+02:00[Europe/Stockholm]"
Json.stringify([t]) // "[\"2026-10-22T16:03:05+02:00[Europe/Stockholm]\"]"
```

A zoned value is a value like any other, so it serves as a record field or a map key. One moment in
two zones gives two keys:

```ant
type Meeting = { title: string, at: Time.ZonedDateTime }
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let review = Meeting{ title: "Review", at: t }
review.at.hour()                                          // 16
Map{ t: "Stockholm", t.withZone(Time.utc): "UTC" }.size() // 2
```

## Applications

Add a day across a daylight-saving change. New York moves its clocks back on November 1, 2026, so a
calendar day from noon on Saturday is noon on Sunday, 25 hours later, while 24 hours is 11:00:

```ant
let sat = Time.parseZoned("2026-10-31T12:00[America/New_York]").unwrap()
sat.add(Time.duration(days: 1))            // 2026-11-01T12:00:00-05:00[America/New_York]
sat.add(Time.duration(hours: 24))          // 2026-11-01T11:00:00-05:00[America/New_York]
sat.until(sat.add(Time.duration(days: 1))) // PT25H
```

Find how long a local day lasts, and when the clocks change:

```ant
let day = Time.parseZoned("2026-03-08[America/New_York]").unwrap()
day                  // 2026-03-08T00:00:00-05:00[America/New_York]
day.hoursInDay()     // 23.0
day.nextTransition() // 2026-03-08T03:00:00-04:00[America/New_York]
```

Show a call scheduled in one zone to people in others:

```ant
let call = Time.parseZoned("2026-10-22T09:30[America/Los_Angeles]").unwrap()
let cities = ["Europe/London", "Asia/Tokyo"]
for city in cities {
  let local = call.withZone(Time.zone(city).unwrap())
  println("${city}: ${local.toPlainDate()} ${local.toPlainTime()}")
}
```

```text
Europe/London: 2026-10-22 17:30:00
Asia/Tokyo: 2026-10-23 01:30:00
```

## Properties and relations

The text of a zoned value parses back to an equal value:

```ant
let t = Time.parseZoned("2024-11-03T01:30:00-05:00[America/New_York]").unwrap()
assert(Time.parseZoned("${t}").unwrap() == t)
```

Seeing an instant in a zone and taking the instant back gives the same instant:

```ant
let i = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
let ny = Time.zone("America/New_York").unwrap()
assert(i.inZone(ny).toInstant() == i)
assert(i.inZone(ny).zoneOf() == ny)
```

`withZone` keeps the moment, so the two compare as `0`, but they are not `==`:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let utc = t.withZone(Time.utc)
assert(Time.compare(t, utc) == 0)
assert(t.toInstant() == utc.toInstant())
assert(t != utc)
```

## Possible issues

A wall-clock time can happen twice in a zone, or not at all. New York repeats 01:30 on November 1,
2026, and [Time.toZoned](../toZoned.md) takes the earlier one unless
[Time.Disambiguation](Disambiguation.md) says otherwise.
[Time.possibleInstants](../possibleInstants.md) counts the candidates:

```ant
let ny = Time.zone("America/New_York").unwrap()
let t = Time.parsePlainDateTime("2026-11-01T01:30").unwrap()
Time.toZoned(t, ny)                               // 2026-11-01T01:30:00-04:00[America/New_York]
Time.toZoned(t, ny, disambiguation: Time.Later{}) // 2026-11-01T01:30:00-05:00[America/New_York]
Time.possibleInstants(t, ny).length()             // 2
```

New York skips 02:30 on March 8, 2026, so that time moves forward to 03:30:

```ant
let ny = Time.zone("America/New_York").unwrap()
let t = Time.parsePlainDateTime("2026-03-08T02:30").unwrap()
Time.toZoned(t, ny)                   // 2026-03-08T03:30:00-04:00[America/New_York]
Time.possibleInstants(t, ny).length() // 0
```

A zoned value is not an instant. Convert it with [Time.toInstant](../toInstant.md):

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let i: Time.Instant = t // error: cannot use Time.ZonedDateTime where Time.Instant is expected
```

`<` does not order zoned values. Use [Time.compare](../compare.md):

```ant
let a = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let earlier = a < a.withZone(Time.utc)
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.ZonedDateTime; narrow first
```

A difference in days between two zones panics, since their days do not line up. Move one value into
the other's zone first:

```ant
let a = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let b = Time.parseZoned("2026-10-25T09:00[UTC]").unwrap()
a.until(b, largestUnit: Time.Day{})
// panic: Time.until: Europe/Stockholm and UTC are different zones; a difference in days or larger needs one zone
a.withZone(Time.utc).until(b, largestUnit: Time.Day{}) // P2DT18H56M55S
```

Text needs a zone in brackets. An offset is optional, but when the text has one, the zone must have
that offset at that time:

```ant
Time.parseZoned("2026-10-22T14:03:05Z")
// Err{ error: Error{ message: "a zoned date-time needs a date and a [zone]: \"2026-10-22T14:03:05Z\"" } }
Time.parseZoned("2026-10-22T16:03:05+03:00[Europe/Stockholm]")
// Err{ error: Error{ message: "the offset +03:00 is not the offset of Europe/Stockholm at 2026-10-22T16:03:05" } }
```

## See also

[Time.parseZoned](../parseZoned.md), [Time.inZone](../inZone.md), [Time.toZoned](../toZoned.md),
[Time.nowIn](../nowIn.md), [Time.withZone](../withZone.md), [Time.toInstant](../toInstant.md),
[Time.zoneOf](../zoneOf.md), [Time.offset](../offset.md), [Time.startOfDay](../startOfDay.md),
[Time.hoursInDay](../hoursInDay.md), [Time.nextTransition](../nextTransition.md),
[Time.Instant](Instant.md), [Time.PlainDateTime](PlainDateTime.md), [Time.Zone](Zone.md),
[Time.Disambiguation](Disambiguation.md)
