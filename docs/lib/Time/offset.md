<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.offset

```ant signature
func Time.offset(t: Time.ZonedDateTime): string
```

Time.offset returns the offset of t from UTC as ±HH:MM, such as +02:00, with seconds as well when a
historical offset has some. The string of t rounds its offset to the minute; this does not.

<!-- end generated -->

## Details

- The offset belongs to the moment, not to the zone: a zone with daylight saving time has one offset
  in winter and another in summer.
- An offset of zero, as in UTC, is `+00:00`, with a plus sign.

## Basic examples

The offset of Stockholm in summer:

```ant
let t = Result.unwrap(Time.parseZoned("2026-07-01T12:00[Europe/Stockholm]"))
Time.offset(t) // "+02:00"
```

The same zone in winter has a different offset:

```ant
let t = Result.unwrap(Time.parseZoned("2026-01-15T12:00[Europe/Stockholm]"))
Time.offset(t) // "+01:00"
```

## Scope

Zones west of Greenwich have a minus sign, and UTC has a plus:

```ant
Result.unwrap(Time.parseZoned("2026-07-01T12:00[America/Los_Angeles]")).offset() // "-07:00"
Result.unwrap(Time.parseZoned("2026-07-01T12:00[UTC]")).offset()                 // "+00:00"
```

Some zones are offset by half an hour or by three quarters:

```ant
Result.unwrap(Time.parseZoned("2026-07-01T12:00[America/St_Johns]")).offset() // "-02:30"
Result.unwrap(Time.parseZoned("2026-07-01T12:00[Asia/Kathmandu]")).offset()   // "+05:45"
Result.unwrap(Time.parseZoned("2026-07-01T12:00[Pacific/Chatham]")).offset()  // "+12:45"
```

A zone with a fixed offset always gives that offset:

```ant
let zone = Result.unwrap(Time.zone("+0530"))
let at = Result.unwrap(Time.parseInstant("2026-07-01T12:00:00Z"))
at.inZone(zone).offset() // "+05:30"
```

Before standard time, Paris kept its local mean time, nine minutes and 21 seconds ahead of
Greenwich. The offset keeps the seconds, and the string of `t` rounds them away:

```ant
let t = Result.unwrap(Time.parseZoned("1800-07-01T12:00[Europe/Paris]"))
t.offset() // "+00:09:21"
"${t}"     // "1800-07-01T12:00:00+00:09[Europe/Paris]"
```

One instant has a different offset in each zone:

```ant
let at = Result.unwrap(Time.parseInstant("2026-07-01T12:00:00Z"))
at.inZone(Result.unwrap(Time.zone("Asia/Tokyo"))).offset()       // "+09:00"
at.inZone(Result.unwrap(Time.zone("America/New_York"))).offset() // "-04:00"
```

In a pipeline, the zoned time is the first argument:

```ant
"2026-07-01T12:00[Australia/Sydney]" | Time.parseZoned() | Result.map(Time.offset)
// Ok{ value: "+10:00" }
```

## Applications

Label a time with its offset from UTC:

```ant
let t = Result.unwrap(Time.parseZoned("2026-07-01T14:30[Asia/Kolkata]"))
println("${t.toPlainTime()} (UTC${t.offset()})")
```

```text
14:30:00 (UTC+05:30)
```

Show the offset on each side of a daylight-saving change:

```ant
let t = Result.unwrap(Time.parseZoned("2026-03-01T12:00[America/New_York]"))
let change = t.nextTransition()
println(if change == null { "no change" } else { "${t.offset()} until ${change.toPlainDate()}, then ${change.offset()}" })
```

```text
-05:00 until 2026-03-08, then -04:00
```

Print a meeting time for the offices of a team:

```ant
let meeting = Result.unwrap(Time.parseInstant("2026-10-22T15:00:00Z"))
for id in ["America/Los_Angeles", "Europe/London", "Asia/Kolkata"] {
  let local = meeting.inZone(Result.unwrap(Time.zone(id)))
  println("${id}: ${local.toPlainTime()} ${local.offset()}")
}
```

```text
America/Los_Angeles: 08:00:00 -07:00
Europe/London: 16:00:00 +01:00
Asia/Kolkata: 20:30:00 +05:30
```

## Properties and relations

An offset zone gives its own identifier as its offset:

```ant
let zone = Result.unwrap(Time.zone("-0330"))
let t = Result.unwrap(Time.parseInstant("2026-07-01T12:00:00Z")).inZone(zone)
assert(t.offset() == zone.zoneId())
```

The offset says the same as [Time.offsetSeconds](offsetSeconds.md), written as text:

```ant
let t = Result.unwrap(Time.parseZoned("2026-07-01T12:00[America/St_Johns]"))
assert(t.offset() == "-02:30")
assert(t.offsetSeconds() == -(2 * 3600 + 30 * 60))
```

When the offset has no seconds, the string of `t` holds it as it is:

```ant
let t = Result.unwrap(Time.parseZoned("2026-07-01T12:00[Asia/Kathmandu]"))
assert("${t}".contains(t.offset()))
```

## Possible issues

The offset is not the zone. Two zones can share an offset today and differ tomorrow, so keep the
zone, not the offset, to place later times:

```ant
let at = Result.unwrap(Time.parseInstant("2026-07-01T12:00:00Z"))
let london = at.inZone(Result.unwrap(Time.zone("Europe/London")))
let lagos = at.inZone(Result.unwrap(Time.zone("Africa/Lagos")))
london.offset() == lagos.offset() // true
london.zoneOf() == lagos.zoneOf() // false
```

A `Time.Zone` has no single offset, so `Time.offset` takes a zoned time, not a zone:

```ant
let zone = Result.unwrap(Time.zone("Europe/Stockholm"))
println(Time.offset(zone))
// error: cannot use Time.Zone where Time.ZonedDateTime is expected
```

An offset with seconds is not one the parsers accept, so text built from it does not parse back:

```ant
let t = Result.unwrap(Time.parseZoned("1800-07-01T12:00[Europe/Paris]"))
Time.parseInstant("${t.toPlainDateTime()}${t.offset()}")
// Err{ error: Error{ message: "not a date or time: \"1800-07-01T12:00:00+00:09:21\"" } }
```

Use the string of `t` instead, which rounds the offset to the minute and parses back to the same
moment.

## See also

[Time.offsetSeconds](offsetSeconds.md) · [Time.zoneOf](zoneOf.md) · [Time.zoneId](zoneId.md) ·
[Time.nextTransition](nextTransition.md) · [Time.inZone](inZone.md) · [Time.zone](zone.md) ·
[Time.ZonedDateTime](types/ZonedDateTime.md)
