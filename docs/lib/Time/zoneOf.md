<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.zoneOf

```ant signature
func Time.zoneOf(t: Time.ZonedDateTime): Time.Zone
```

Time.zoneOf returns the zone of t.

<!-- end generated -->

## Details

- The zone is the one `t` carries, as it was given: a zone parsed from an alias such as
  `Asia/Calcutta` keeps that name, and an offset zone stays an offset zone.
- The zone is not the offset. A named zone holds every offset it has had and will have, so carry the
  zone to place other times; [Time.offset](offset.md) is the offset at the moment of `t` alone.

## Basic examples

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
Time.zoneOf(t) // Europe/Stockholm
```

Call it through `.`, and take its identifier as a string with [Time.zoneId](zoneId.md):

```ant
let t = Time.parseZoned("2026-10-22T09:30-04:00[America/New_York]").unwrap()
t.zoneOf()          // America/New_York
t.zoneOf().zoneId() // "America/New_York"
```

## Scope

A zone named in any letter case comes back as the database spells it, and an alias keeps its own
name while it equals the zone it stands for:

```ant
let calcutta = Time.zone("asia/calcutta").unwrap()
let t = Time.parsePlainDateTime("2026-10-22T16:03").unwrap().toZoned(calcutta)
t.zoneOf()                                       // Asia/Calcutta
t.zoneOf() == Time.zone("Asia/Kolkata").unwrap() // true
```

A time with a fixed-offset zone gives that offset zone:

```ant
let t = Time.parseZoned("2026-10-22T16:03+05:30[+05:30]").unwrap()
t.zoneOf() // +05:30
```

The zone is the one given to [Time.inZone](inZone.md), [Time.withZone](withZone.md),
[Time.toZoned](toZoned.md) or [Time.nowIn](nowIn.md):

```ant
let at = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
let wall = Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(9, 0, 0))
at.inZone(Time.utc).zoneOf()                            // UTC
wall.toZoned(Time.zone("Asia/Tokyo").unwrap()).zoneOf() // Asia/Tokyo
Time.nowIn(Time.zone("Asia/Tokyo").unwrap()).zoneOf()   // Asia/Tokyo
```

In a pipe, the zoned value from a parse goes through `Result.map`:

```ant
"2026-10-22T16:03:05+11:00[Australia/Sydney]" | Time.parseZoned() | Result.map(Time.zoneOf) // Ok{ value: Australia/Sydney }
```

## Applications

Set a follow-up at 9:00 on another date, in the zone of the first meeting:

```ant
let first = Time.parseZoned("2026-10-22T14:00+02:00[Europe/Stockholm]").unwrap()
let date = Time.plainDate(2026, 11, 5)
let followUp = Time.plainDateTime(date, Time.plainTime(9, 0, 0)).toZoned(first.zoneOf())
println(followUp)
```

```text
2026-11-05T09:00:00+01:00[Europe/Stockholm]
```

Stockholm has left summer time by November 5, and the zone, not the offset of the first meeting,
gives the right `+01:00`.

Store a time in UTC and show it again in the zone it came from:

```ant
let local = Time.parseZoned("2026-10-22T18:45-07:00[America/Los_Angeles]").unwrap()
let zone = local.zoneOf()
let stored = local.withZone(Time.utc)
println(stored)
println(stored.withZone(zone))
```

```text
2026-10-23T01:45:00+00:00[UTC]
2026-10-22T18:45:00-07:00[America/Los_Angeles]
```

Group log entries by the zone they were written in:

```ant
let entries = [
  "2026-10-22T09:15+02:00[Europe/Stockholm]",
  "2026-10-22T03:05-04:00[America/New_York]",
  "2026-10-22T16:40+02:00[Europe/Stockholm]",
]
  | Array.map({ Time.parseZoned(x).unwrap() })

let byZone = entries | Array.groupBy({ x.zoneOf().zoneId() })
for { key, value } in byZone.entries() {
  println("${key}: ${value.length()}")
}
```

```text
Europe/Stockholm: 2
America/New_York: 1
```

## Properties and relations

[Time.withZone](withZone.md) with the zone of `t` gives `t` back, and the result of `withZone`
carries the zone it was given:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
assert(t.withZone(t.zoneOf()) == t)
assert(t.withZone(tokyo).zoneOf() == tokyo)
```

The instant and the zone together rebuild `t`:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05.25-04:00[America/New_York]").unwrap()
assert(t.toInstant().inZone(t.zoneOf()) == t)
```

Calendar arithmetic and transitions keep the zone:

```ant
let t = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
assert(t.add(Time.duration(days: 30)).zoneOf() == t.zoneOf())
assert(t.startOfDay().zoneOf() == t.zoneOf())
assert(t.nextTransition()?.zoneOf() == t.zoneOf())
```

## Possible issues

Two zones can share an offset today and still differ. Stockholm and Berlin show the same time, but
their zones are not equal:

```ant
let at = Time.parseInstant("2026-10-22T14:00:00Z").unwrap()
let stockholm = at.inZone(Time.zone("Europe/Stockholm").unwrap())
let berlin = at.inZone(Time.zone("Europe/Berlin").unwrap())
stockholm.offset() == berlin.offset() // true
stockholm.zoneOf() == berlin.zoneOf() // false
```

A named zone never equals an offset zone, not even `UTC` and `+00:00`:

```ant
let t = Time.parseInstant("2026-10-22T14:00:00Z").unwrap().inZone(Time.utc)
t.zoneOf() == Time.utc                     // true
t.zoneOf() == Time.zone("+00:00").unwrap() // false
```

A `Time.Instant` has no zone. Place it in one with [Time.inZone](inZone.md) first:

```ant
let at = Time.parseInstant("2026-10-22T14:00:00Z").unwrap()
println(Time.zoneOf(at))
// error: cannot use Time.Instant where Time.ZonedDateTime is expected
```

## See also

[Time.zoneId](zoneId.md) · [Time.zone](zone.md) · [Time.offset](offset.md) ·
[Time.withZone](withZone.md) · [Time.inZone](inZone.md) · [Time.toZoned](toZoned.md) ·
[Time.Zone](types/Zone.md) · [Time.ZonedDateTime](types/ZonedDateTime.md)
