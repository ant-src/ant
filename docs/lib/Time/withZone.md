<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.withZone

```ant signature
func Time.withZone(t: Time.ZonedDateTime, z: Time.Zone): Time.ZonedDateTime
```

Time.withZone returns the instant of t seen in zone z instead.

<!-- end generated -->

## Details

- The result is the same instant as `t`, shown in `z`: the wall-clock fields and the offset are
  those of `z` at that instant, and the moment does not change. It never needs a disambiguation.
- The offset is the one `z` has at that instant, so the result can fall on another date than `t`.
- To keep the wall-clock time and change the moment instead, take
  [Time.toPlainDateTime](toPlainDateTime.md) and place it with [Time.toZoned](toZoned.md). For a
  `Time.Instant`, use [Time.inZone](inZone.md).

## Basic examples

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
Time.withZone(t, tokyo)    // 2026-10-22T23:03:05+09:00[Asia/Tokyo]
Time.withZone(t, Time.utc) // 2026-10-22T14:03:05+00:00[UTC]
```

Call it through `.`:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
t.withZone(Time.zone("America/New_York").unwrap()) // 2026-10-22T10:03:05-04:00[America/New_York]
```

## Scope

A fixed-offset zone shows the instant at its offset, and its identifier is the offset:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
t.withZone(Time.zone("-03:30").unwrap()) // 2026-10-22T10:33:05-03:30[-03:30]
```

The result can fall on the next or the previous date:

```ant
let late = Time.parseZoned("2026-10-22T23:30+02:00[Europe/Stockholm]").unwrap()
let kiritimati = Time.zone("Pacific/Kiritimati").unwrap()
late.withZone(kiritimati) // 2026-10-23T11:30:00+14:00[Pacific/Kiritimati]
```

Two instants an hour apart can show the same wall-clock time when a daylight-saving change repeats
an hour. The offset keeps them apart:

```ant
let ny = Time.zone("America/New_York").unwrap()
let first = Time.parseZoned("2024-11-03T05:30:00Z[UTC]").unwrap()
let second = Time.parseZoned("2024-11-03T06:30:00Z[UTC]").unwrap()
first.withZone(ny)  // 2024-11-03T01:30:00-04:00[America/New_York]
second.withZone(ny) // 2024-11-03T01:30:00-05:00[America/New_York]
```

Nanoseconds are kept:

```ant
let t = Time.parseZoned("2026-10-22T23:30:00.000000007+02:00[Europe/Stockholm]").unwrap()
t.withZone(Time.utc) // 2026-10-22T21:30:00.000000007+00:00[UTC]
```

`Time.utc` is the named zone UTC, not the offset zone `+00:00`. Both show the same wall-clock time,
but the identifiers differ, so the two results are not `==`, although [Time.compare](compare.md)
gives 0:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let plus0 = Time.zone("+00:00").unwrap()
t.withZone(Time.utc)                                  // 2026-10-22T14:03:05+00:00[UTC]
t.withZone(plus0)                                     // 2026-10-22T14:03:05+00:00[+00:00]
t.withZone(Time.utc) == t.withZone(plus0)             // false
Time.compare(t.withZone(Time.utc), t.withZone(plus0)) // 0
```

In a pipe, the zoned value from a parse goes through `Result.map`:

```ant
let tokyo = Time.zone("Asia/Tokyo").unwrap()
"2026-10-22T16:03:05+02:00[Europe/Stockholm]"
  | Time.parseZoned()
  | Result.map({ x.withZone(tokyo) }) // Ok{ value: 2026-10-22T23:03:05+09:00[Asia/Tokyo] }
```

## Applications

Show a meeting set in one office at the local time of each attendee:

```ant
type Attendee = { name: string, zone: string }

let meeting = Time.parseZoned("2026-10-22T09:00+02:00[Europe/Stockholm]").unwrap()
let attendees = [
  Attendee{ name: "Ana", zone: "America/Sao_Paulo" },
  Attendee{ name: "Ken", zone: "Asia/Tokyo" },
  Attendee{ name: "Ola", zone: "Europe/Oslo" },
]
for a in attendees {
  let local = meeting.withZone(Time.zone(a.zone).unwrap())
  println("${a.name}: ${local.toPlainDate()} ${local.toPlainTime()}")
}
```

```text
Ana: 2026-10-22 04:00:00
Ken: 2026-10-22 16:00:00
Ola: 2026-10-22 09:00:00
```

Bring timestamps from several zones to UTC, so that their text sorts by time:

```ant
let stamps = [
  "2026-10-22T09:15+02:00[Europe/Stockholm]",
  "2026-10-22T03:05-04:00[America/New_York]",
  "2026-10-22T15:40+09:00[Asia/Tokyo]",
]

let lines = stamps
  | Array.map({ Time.parseZoned(x).unwrap().withZone(Time.utc) })
  | Array.map({ String.from(x) })
  | Array.sort()

for line in lines {
  println(line)
}
```

```text
2026-10-22T06:40:00+00:00[UTC]
2026-10-22T07:05:00+00:00[UTC]
2026-10-22T07:15:00+00:00[UTC]
```

Check whether a support desk is open when a ticket arrives, Monday to Friday, 9 to 17 in its zone:

```ant
func isOpen(at: Time.ZonedDateTime, desk: Time.Zone): bool {
  let local = at.withZone(desk)
  if local.dayOfWeek() > 5 {
    return false
  }
  local.hour() >= 9 && local.hour() < 17
}

let ticket = Time.parseZoned("2026-10-22T18:30-07:00[America/Los_Angeles]").unwrap()
println(isOpen(ticket, Time.zone("Europe/Dublin").unwrap()))
println(isOpen(ticket, Time.zone("Asia/Singapore").unwrap()))
```

```text
false
true
```

## Properties and relations

The result holds the instant of `t` and the zone it was given:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05.5+02:00[Europe/Stockholm]").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
let moved = t.withZone(tokyo)
assert(moved.toInstant() == t.toInstant())
assert(moved.zoneOf() == tokyo)
assert(Time.compare(moved, t) == 0)
```

It is [Time.inZone](inZone.md) applied to the instant of `t`, and moving to the zone `t` already has
gives `t` back:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let ny = Time.zone("America/New_York").unwrap()
assert(t.withZone(ny) == t.toInstant().inZone(ny))
assert(t.withZone(t.zoneOf()) == t)
```

Only the last move counts:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
let ny = Time.zone("America/New_York").unwrap()
assert(t.withZone(tokyo).withZone(ny) == t.withZone(ny))
```

## Possible issues

`==` compares the zone too, so the result is not equal to `t`. Compare the instants, or use
[Time.compare](compare.md):

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let moved = t.withZone(Time.zone("Asia/Tokyo").unwrap())
moved == t                         // false
moved.toInstant() == t.toInstant() // true
Time.compare(moved, t)             // 0
```

`withZone` keeps the moment, not the clock reading: 16:03 in Stockholm becomes 23:03 in Tokyo. To
keep the clock reading in another zone, convert to a plain date-time and place it with
[Time.toZoned](toZoned.md), which gives another moment:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
t.withZone(tokyo)                  // 2026-10-22T23:03:05+09:00[Asia/Tokyo]
t.toPlainDateTime().toZoned(tokyo) // 2026-10-22T16:03:05+09:00[Asia/Tokyo]
```

Calendar arithmetic follows the zone of the value, so adding a day before or after the move can
differ. Stockholm turns its clocks back on October 25, 2026, so its day is 25 hours long, while
Tokyo's is 24:

```ant
let start = Time.parseZoned("2026-10-24T12:00+02:00[Europe/Stockholm]").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
let day = Time.duration(days: 1)
start.add(day).withZone(tokyo) // 2026-10-25T20:00:00+09:00[Asia/Tokyo]
start.withZone(tokyo).add(day) // 2026-10-25T19:00:00+09:00[Asia/Tokyo]
```

A `Time.Instant` has no zone to move from, so passing one is a compile error. Use
[Time.inZone](inZone.md):

```ant
let t = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
let z = Time.withZone(t, Time.utc) // error: cannot use Time.Instant where Time.ZonedDateTime is expected
```

The zone is a `Time.Zone`, not its name. Get one from [Time.zone](zone.md) first:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let u = t.withZone("Asia/Tokyo") // error: cannot use string where Time.Zone is expected
```

## See also

[Time.inZone](inZone.md) · [Time.toZoned](toZoned.md) · [Time.toInstant](toInstant.md) ·
[Time.toPlainDateTime](toPlainDateTime.md) · [Time.zone](zone.md) · [Time.zoneOf](zoneOf.md) ·
[Time.compare](compare.md) · [Time.ZonedDateTime](types/ZonedDateTime.md) ·
[Time.Zone](types/Zone.md) · [Time](index.md)
