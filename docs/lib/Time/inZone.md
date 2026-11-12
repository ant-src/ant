<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.inZone

```ant signature
func Time.inZone(t: Time.Instant, z: Time.Zone): Time.ZonedDateTime
```

Time.inZone returns the instant t seen in zone z.

<!-- end generated -->

## Details

- The result is the same instant: only the wall-clock fields and the offset depend on the zone. It
  never needs a disambiguation, since every instant has exactly one wall-clock time in a zone.
- The offset is the one the zone has at that instant, so a zone with daylight saving gives different
  offsets in summer and winter.
- To move a `Time.ZonedDateTime` to another zone, use [Time.withZone](withZone.md); to place a
  wall-clock time in a zone, use [Time.toZoned](toZoned.md).

## Basic examples

```ant
let t = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
let stockholm = Time.zone("Europe/Stockholm").unwrap()
Time.inZone(t, stockholm) // 2026-10-22T16:03:05+02:00[Europe/Stockholm]
Time.inZone(t, Time.utc)  // 2026-10-22T14:03:05+00:00[UTC]
```

Call it through `.`:

```ant
let epoch = Time.fromEpochMillis(0)
let ny = Time.zone("America/New_York").unwrap()
epoch.inZone(ny) // 1969-12-31T19:00:00-05:00[America/New_York]
```

## Scope

A fixed-offset zone shows every instant at the same offset, and its identifier is the offset:

```ant
let t = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
t.inZone(Time.zone("+05:30").unwrap()) // 2026-10-22T19:33:05+05:30[+05:30]
```

A named zone uses the offset in force at the instant, including offsets that are not whole hours:

```ant
let stockholm = Time.zone("Europe/Stockholm").unwrap()
let winter = Time.parseInstant("2026-01-15T12:00:00Z").unwrap()
let summer = Time.parseInstant("2026-07-15T12:00:00Z").unwrap()
let kathmandu = Time.zone("Asia/Kathmandu").unwrap()
winter.inZone(stockholm) // 2026-01-15T13:00:00+01:00[Europe/Stockholm]
summer.inZone(stockholm) // 2026-07-15T14:00:00+02:00[Europe/Stockholm]
summer.inZone(kathmandu) // 2026-07-15T17:45:00+05:45[Asia/Kathmandu]
```

When a daylight-saving change repeats an hour, two instants show the same wall-clock time. The
offset tells them apart:

```ant
let ny = Time.zone("America/New_York").unwrap()
let first = Time.parseInstant("2024-11-03T05:30:00Z").unwrap()
let second = Time.parseInstant("2024-11-03T06:30:00Z").unwrap()
first.inZone(ny)  // 2024-11-03T01:30:00-04:00[America/New_York]
second.inZone(ny) // 2024-11-03T01:30:00-05:00[America/New_York]
```

Nanoseconds are kept:

```ant
Time.fromEpochNanos(1).inZone(Time.utc) // 1970-01-01T00:00:00.000000001+00:00[UTC]
```

After a parse in a pipe, map over the Result:

```ant
let tokyo = Time.zone("Asia/Tokyo").unwrap()
"2026-10-22T14:03:05Z" | Time.parseInstant() | Result.map({ x.inZone(tokyo) }) // Ok{ value: 2026-10-22T23:03:05+09:00[Asia/Tokyo] }
```

In the host's own zone, from [Time.localZone](localZone.md), the result is still the same instant:

```ant
let t = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
t.inZone(Time.localZone()).toInstant() == t // true
```

## Applications

Show one moment on a world clock:

```ant
let launch = Time.parseInstant("2026-10-22T14:00:00Z").unwrap()
let cities = ["America/Los_Angeles", "Europe/Stockholm", "Asia/Tokyo", "Australia/Sydney"]
for city in cities {
  let local = launch.inZone(Time.zone(city).unwrap())
  println("${city}: ${local.toPlainDate()} ${local.toPlainTime()}")
}
```

```text
America/Los_Angeles: 2026-10-22 07:00:00
Europe/Stockholm: 2026-10-22 16:00:00
Asia/Tokyo: 2026-10-22 23:00:00
Australia/Sydney: 2026-10-23 01:00:00
```

Count events by the day they happened on for a user in another zone:

```ant
let user = Time.zone("America/Chicago").unwrap()
let events = [
  "2026-10-22T03:15:00Z",
  "2026-10-22T09:40:00Z",
  "2026-10-22T23:05:00Z",
  "2026-10-23T04:50:00Z",
]

let byDay = events
  | Array.map({ Time.parseInstant(x).unwrap() })
  | Array.groupBy({ x.inZone(user).toPlainDate() })

for { key, value } in byDay {
  println("${key}: ${value.length()}")
}
```

```text
2026-10-21: 1
2026-10-22: 3
```

Check whether an instant falls in an office's working hours, Monday to Friday, 9 to 17:

```ant
func isOpen(at: Time.Instant, office: Time.Zone): bool {
  let local = at.inZone(office)
  if local.dayOfWeek() > 5 {
    return false
  }
  local.hour() >= 9 && local.hour() < 17
}

let london = Time.zone("Europe/London").unwrap()
let singapore = Time.zone("Asia/Singapore").unwrap()
let call = Time.parseInstant("2026-10-22T07:30:00Z").unwrap()
println(isOpen(call, london))
println(isOpen(call, singapore))
```

```text
false
true
```

## Properties and relations

The result holds the instant and the zone it was given:

```ant
let t = Time.parseInstant("2026-10-22T14:03:05.5Z").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
let local = t.inZone(tokyo)
assert(local.toInstant() == t)
assert(local.zoneOf() == tokyo)
```

Seeing an instant in one zone and moving it to another with [Time.withZone](withZone.md) is the same
as seeing it in the second zone:

```ant
let t = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
let stockholm = Time.zone("Europe/Stockholm").unwrap()
assert(t.inZone(tokyo).withZone(stockholm) == t.inZone(stockholm))
```

The text of the result parses back to the same value, even in a repeated hour, since the text keeps
the offset:

```ant
let ny = Time.zone("America/New_York").unwrap()
let late = Time.parseInstant("2024-11-03T06:30:00Z").unwrap().inZone(ny)
assert(Time.parseZoned(String.from(late)).unwrap() == late)
```

## Possible issues

A wall-clock time is not an instant, so passing one is a compile error. Use
[Time.toZoned](toZoned.md) to place it in a zone:

```ant
let meeting = Time.parsePlainDateTime("2026-10-22T14:00").unwrap()
let z = Time.inZone(meeting, Time.utc)
// error: cannot use Time.PlainDateTime where Time.Instant is expected
```

The zone is a `Time.Zone`, not its name. Get one from [Time.zone](zone.md) first:

```ant
let t = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
let z = t.inZone("Europe/Stockholm") // error: cannot use string where Time.Zone is expected
```

`==` compares the zone too, so the same instant in two zones is not equal. Compare the instants, or
use [Time.compare](compare.md), which orders by instant:

```ant
let t = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
let tokyo = t.inZone(Time.zone("Asia/Tokyo").unwrap())
let stockholm = t.inZone(Time.zone("Europe/Stockholm").unwrap())
tokyo == stockholm                         // false
tokyo.toInstant() == stockholm.toInstant() // true
Time.compare(tokyo, stockholm)             // 0
```

`Time.utc` is the named zone UTC, not the offset `+00:00`, so the two results differ though they
show the same time:

```ant
let t = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
let zero = Time.zone("+00:00").unwrap()
t.inZone(Time.utc)                   // 2026-10-22T14:03:05+00:00[UTC]
t.inZone(zero)                       // 2026-10-22T14:03:05+00:00[+00:00]
t.inZone(Time.utc) == t.inZone(zero) // false
```

## See also

[Time.withZone](withZone.md) · [Time.toZoned](toZoned.md) · [Time.toInstant](toInstant.md) ·
[Time.nowIn](nowIn.md) · [Time.zone](zone.md) · [Time.localZone](localZone.md) ·
[Time.zoneOf](zoneOf.md) · [Time.ZonedDateTime](types/ZonedDateTime.md) ·
[Time.Instant](types/Instant.md)
