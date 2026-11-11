<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.zoneId

```ant signature
func Time.zoneId(z: Time.Zone): string
```

Time.zoneId returns the identifier of z: its name as the database spells it, or its offset as
±HH:MM.

<!-- end generated -->

## Details

- A name comes back as the database spells it, whatever letter case `Time.zone` was given:
  `asia/tokyo` gives `Asia/Tokyo`.
- An alias keeps its own name: `Asia/Calcutta` stays `Asia/Calcutta`, although it equals the zone
  `Asia/Kolkata`.
- An offset comes back as `±HH:MM`, whichever form `Time.zone` was given.
- The identifier is what a zone prints as, and `Time.zone` parses it back to an equal zone.

## Basic examples

```ant
Time.zoneId(Result.unwrap(Time.zone("asia/calcutta"))) // "Asia/Calcutta"
Time.zoneId(Result.unwrap(Time.zone("+0530")))         // "+05:30"
```

The zone of a zoned time:

```ant
let t = Result.unwrap(Time.parseZoned("2026-07-01T12:00[Europe/Stockholm]"))
Time.zoneId(Time.zoneOf(t)) // "Europe/Stockholm"
```

## Scope

Names in any letter case come back as the database spells them:

```ant
Time.zone("EUROPE/STOCKHOLM").unwrap().zoneId() // "Europe/Stockholm"
Time.zone("us/eastern").unwrap().zoneId()       // "US/Eastern"
Time.zone("etc/gmt+5").unwrap().zoneId()        // "Etc/GMT+5"
```

Every form of offset comes back as `±HH:MM`:

```ant
Time.zone("+05:30").unwrap().zoneId() // "+05:30"
Time.zone("-0800").unwrap().zoneId()  // "-08:00"
Time.zone("+14").unwrap().zoneId()    // "+14:00"
```

`Time.utc` is the named zone UTC, not an offset:

```ant
Time.utc.zoneId()                  // "UTC"
Time.zone("utc").unwrap().zoneId() // "UTC"
```

In a pipeline, map it over the result of `Time.zone`:

```ant
"america/new_york" | Time.zone() | Result.map(Time.zoneId) // Ok{ value: "America/New_York" }
```

The identifier of the host's zone is a name such as `Europe/Stockholm`, or an offset such as
`+02:00` when the host names no zone of the database:

<!-- test: check-only output depends on the host -->

```ant
println(Time.localZone().zoneId())
```

Map it over several zones, and join the identifiers:

```ant
["Europe/Stockholm", "asia/tokyo", "-0800"]
  | Array.map({ Time.zone(x).unwrap() })
  | Array.map(Time.zoneId)
  | String.join(", ")
// "Europe/Stockholm, Asia/Tokyo, -08:00"
```

## Applications

Check a zone that a user typed, and store it in the database's spelling:

```ant
func canonicalZone(id: string): Result[string, Error] {
  Time.zone(id) | Result.map(Time.zoneId)
}

println(canonicalZone("asia/kolkata"))
println(canonicalZone("Nowhere"))
```

```text
Ok{ value: "Asia/Kolkata" }
Err{ error: Error{ message: "unknown time zone \"Nowhere\"" } }
```

Save a wall-clock time with its zone as text, and restore it later:

```ant
let meeting = Time.parseZoned("2026-10-22T09:00[America/New_York]").unwrap()
let saved = ["${meeting.toPlainDateTime()}", meeting.zoneOf().zoneId()]

let zone = Time.zone(saved[1]).unwrap()
let restored = Time.parsePlainDateTime(saved[0]).unwrap().toZoned(zone)
println(restored)
assert(restored == meeting)
```

```text
2026-10-22T09:00:00-04:00[America/New_York]
```

Label the clocks of a team's offices:

```ant
let at = Time.parseInstant("2026-07-01T12:00:00Z").unwrap()
for id in ["America/Los_Angeles", "Asia/Tokyo", "+05:45"] {
  let local = at.inZone(Time.zone(id).unwrap())
  println("${local.zoneOf().zoneId()}: ${local.toPlainTime()}")
}
```

```text
America/Los_Angeles: 05:00:00
Asia/Tokyo: 21:00:00
+05:45: 17:45:00
```

## Properties and relations

The identifier parses back to an equal zone, and it is the text a zone prints as:

```ant
for id in ["Europe/Paris", "asia/calcutta", "-0330", "UTC"] {
  let zone = Time.zone(id).unwrap()
  assert(Time.zone(zone.zoneId()).unwrap() == zone)
  assert("${zone}" == zone.zoneId())
}
```

For an offset zone, the identifier is the offset of every time in it:

```ant
let zone = Time.zone("+0530").unwrap()
let t = Time.parseInstant("2026-01-15T12:00:00Z").unwrap().inZone(zone)
assert(t.offset() == zone.zoneId())
```

## Possible issues

Two aliases are equal zones with different identifiers, so compare zones, not identifiers:

```ant
let calcutta = Time.zone("Asia/Calcutta").unwrap()
let kolkata = Time.zone("Asia/Kolkata").unwrap()
calcutta == kolkata                   // true
calcutta.zoneId() == kolkata.zoneId() // false
```

`Time.utc` and the offset zone `+00:00` have the same offset but are different zones, with different
identifiers:

```ant
Time.utc == Time.zone("+00:00").unwrap() // false
Time.zone("+00:00").unwrap().zoneId()    // "+00:00"
```

The identifier of a named zone is not its offset. Use [Time.offset](offset.md) on a zoned time for
that:

```ant
let t = Time.parseZoned("2026-07-01T12:00[Asia/Tokyo]").unwrap()
t.zoneOf().zoneId() // "Asia/Tokyo"
t.offset()          // "+09:00"
```

`Time.zoneId` takes a `Time.Zone`, not a zoned time or a string:

```ant
let t = Time.parseZoned("2026-07-01T12:00[Asia/Tokyo]").unwrap()
println(Time.zoneId(t))
// error: cannot use Time.ZonedDateTime where Time.Zone is expected
println(Time.zoneId("Asia/Tokyo"))
// error: cannot use string where Time.Zone is expected
```

## See also

[Time.zone](zone.md) · [Time.zoneOf](zoneOf.md) · [Time.localZone](localZone.md) ·
[Time.offset](offset.md) · [Time.inZone](inZone.md) · [Time.Zone](types/Zone.md)
