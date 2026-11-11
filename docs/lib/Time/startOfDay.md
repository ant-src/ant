<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.startOfDay

```ant signature
func Time.startOfDay(t: Time.ZonedDateTime): Time.ZonedDateTime
```

Time.startOfDay returns the first instant of the date of t in its zone. That is midnight, except on
a day whose midnight a daylight-saving change skipped.

<!-- end generated -->

## Details

- The result has the date and the zone of `t`. Only the date of `t` matters, so every time of one
  day gives the same start.
- The offset of the result is the offset at the start of the day, which can differ from the offset
  of `t`.

## Basic examples

```ant
let t = Result.unwrap(Time.parseZoned("2026-10-22T14:03:05.5+02:00[Europe/Stockholm]"))
Time.startOfDay(t) // 2026-10-22T00:00:00+02:00[Europe/Stockholm]
```

Santiago moved its clocks from midnight to 01:00 on September 8, 2024, so that day starts at 01:00:

```ant
let t = Result.unwrap(Time.parseZoned("2024-09-08T12:00[America/Santiago]"))
Time.startOfDay(t) // 2024-09-08T01:00:00-03:00[America/Santiago]
```

## Scope

New York changes its clocks at 02:00, so its days of change still start at midnight:

```ant
let t = Result.unwrap(Time.parseZoned("2026-03-08T12:00[America/New_York]"))
t.startOfDay() // 2026-03-08T00:00:00-05:00[America/New_York]
```

Havana changes its clocks at midnight, as Santiago does:

```ant
let t = Result.unwrap(Time.parseZoned("2026-03-08T12:00[America/Havana]"))
t.startOfDay() // 2026-03-08T01:00:00-04:00[America/Havana]
```

The offset of the start is the one in force at the start. New York moves from `-05:00` to `-04:00`
at 02:00 on March 8, 2026, so noon and the start of that day have different offsets:

```ant
let t = Result.unwrap(Time.parseZoned("2026-03-08T12:00[America/New_York]"))
t.offset()              // "-04:00"
t.startOfDay().offset() // "-05:00"
```

One instant falls on different days in different zones, and so has different starts:

```ant
let at = Result.unwrap(Time.parseInstant("2026-10-21T23:30:00Z"))
let stockholm = Result.unwrap(Time.zone("Europe/Stockholm"))
at.inZone(Time.utc).startOfDay()  // 2026-10-21T00:00:00+00:00[UTC]
at.inZone(stockholm).startOfDay() // 2026-10-22T00:00:00+02:00[Europe/Stockholm]
```

Parse and take the start in one pipeline:

```ant
"2026-10-22T14:03[Asia/Tokyo]" | Time.parseZoned() | Result.map(Time.startOfDay) // Ok{ value: 2026-10-22T00:00:00+09:00[Asia/Tokyo] }
```

## Applications

The start of today in Stockholm:

<!-- test: check-only the current date changes between runs -->

```ant
let stockholm = Result.unwrap(Time.zone("Europe/Stockholm"))
println(Time.nowIn(stockholm).startOfDay())
```

How long since the day started and how long until it ends. On the day New York moves its clocks
back, noon is 13 hours after the start and 12 hours before the end:

```ant
let t = Result.unwrap(Time.parseZoned("2026-11-01T12:00[America/New_York]"))
let tomorrow = t.add(Time.duration(days: 1)).startOfDay()
println("passed: ${t.since(t.startOfDay())}")
println("left: ${t.until(tomorrow)}")
```

```text
passed: PT13H
left: PT12H
```

List the start of each of five days around Santiago's change. The start of the next day is one
calendar day later, taken back to its start:

```ant
let first = Result.unwrap(Time.parseZoned("2024-09-06T00:00[America/Santiago]"))
var day = first
for _ in Iter.range(0, 5) {
  println(day)
  day = day.add(Time.duration(days: 1)).startOfDay()
}
```

```text
2024-09-06T00:00:00-04:00[America/Santiago]
2024-09-07T00:00:00-04:00[America/Santiago]
2024-09-08T01:00:00-03:00[America/Santiago]
2024-09-09T00:00:00-03:00[America/Santiago]
2024-09-10T00:00:00-03:00[America/Santiago]
```

Check whether two times fall on the same day in one zone:

```ant
func sameDay(a: Time.ZonedDateTime, b: Time.ZonedDateTime): bool {
  a.startOfDay() == b.withZone(a.zoneOf()).startOfDay()
}

let call = Result.unwrap(Time.parseZoned("2026-10-22T09:00[Europe/Stockholm]"))
let reply = Result.unwrap(Time.parseZoned("2026-10-22T23:30:00Z[UTC]"))
println(sameDay(call, reply))
```

```text
false
```

At 23:30 UTC it is already October 23 in Stockholm.

## Properties and relations

The start keeps the date and the zone, comes no later than `t`, and is its own start:

```ant
let times = ["2026-10-22T14:03[Europe/Stockholm]", "2024-09-08T12:00[America/Santiago]"]
for text in times {
  let t = Result.unwrap(Time.parseZoned(text))
  let start = t.startOfDay()
  assert(start.toPlainDate() == t.toPlainDate())
  assert(start.zoneOf() == t.zoneOf())
  assert(Time.compare(start, t) <= 0)
  assert(start.startOfDay() == start)
}
```

[Time.parseZoned](parseZoned.md) with a date alone gives the start of that day:

```ant
let t = Result.unwrap(Time.parseZoned("2024-09-08T18:00[America/Santiago]"))
assert(Result.unwrap(Time.parseZoned("2024-09-08[America/Santiago]")) == t.startOfDay())
```

## Possible issues

A day is not always 24 hours long, so subtracting 24 hours from the start of a day does not always
reach the start of the day before. April 6, 2024 was 25 hours long in Santiago, so 24 hours before
the start of April 7 is 01:00 on April 6. Subtract a calendar day instead:

```ant
let start = Result.unwrap(Time.parseZoned("2024-04-07T00:00[America/Santiago]"))
start.subtract(Time.duration(hours: 24)) // 2024-04-06T01:00:00-03:00[America/Santiago]
start.subtract(Time.duration(days: 1))   // 2024-04-06T00:00:00-03:00[America/Santiago]
```

Building midnight by hand with `Time.Earlier{}` lands on the day before when midnight was skipped;
`Time.startOfDay` stays on the date:

```ant
let santiago = Result.unwrap(Time.zone("America/Santiago"))
let midnight = Time.plainDateTime(Time.plainDate(2024, 9, 8), Time.plainTime())
midnight.toZoned(santiago, disambiguation: Time.Earlier{}).toPlainDate() // 2024-09-07
midnight.toZoned(santiago).startOfDay().toPlainDate()                    // 2024-09-08
```

A `Time.Instant` has no zone and so no day. Put it in a zone with [Time.inZone](inZone.md) first:

```ant
println(Time.startOfDay(Time.fromEpochMillis(0)))
// error: cannot use Time.Instant where Time.ZonedDateTime is expected
```

A `Time.PlainDate` has no zone either. Use [Time.toZoned](toZoned.md) to place it in one:

```ant
println(Time.startOfDay(Time.plainDate(2026, 1, 1)))
// error: cannot use Time.PlainDate where Time.ZonedDateTime is expected
```

## See also

[Time.hoursInDay](hoursInDay.md) · [Time.nowIn](nowIn.md) · [Time.toZoned](toZoned.md) ·
[Time.parseZoned](parseZoned.md) · [Time.possibleInstants](possibleInstants.md) ·
[Time.nextTransition](nextTransition.md) · [Time.toPlainDate](toPlainDate.md) · [Time.with](with.md)
· [Time.ZonedDateTime](types/ZonedDateTime.md)
