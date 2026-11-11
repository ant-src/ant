<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.zone

```ant signature
func Time.zone(id: string): Result[Time.Zone, Error]
```

Time.zone returns the zone named by id, a name of the IANA database in any letter case or a fixed
offset. An alias such as Asia/Calcutta is accepted, and an offset may be written ±HH:MM, ±HHMM or
±HH, below 24 hours; anything else, the empty string included, is an Err.

<!-- end generated -->

## Details

- A name matches in any letter case, and the zone keeps the spelling of the database: `"utc"` gives
  `UTC`. An alias keeps its own name, so `Asia/Calcutta` stays `Asia/Calcutta`.
- An offset is a sign and two-digit hours, with or without two-digit minutes: `+05:30`, `+0530` and
  `+05` are accepted; `+5:30`, `Z` and text with spaces around it are not.
- Two named zones are equal when they are aliases of one zone, two offset zones when their offsets
  are; a named zone never equals an offset zone, so `Time.zone("+00:00")` is not `Time.utc`.
- An offset zone keeps one offset all year. Use a named zone when daylight saving time matters.

## Basic examples

```ant
Time.zone("Europe/Stockholm") // Ok{ value: Europe/Stockholm }
Time.zone("+05:30")           // Ok{ value: +05:30 }
Time.zone("Mars/Base")        // Err{ error: Error{ message: "unknown time zone \"Mars/Base\"" } }
```

Unwrap the zone to place an instant in it:

```ant
let stockholm = Time.zone("Europe/Stockholm").unwrap()
Time.fromEpochMillis(0).inZone(stockholm) // 1970-01-01T01:00:00+01:00[Europe/Stockholm]
```

## Scope

Letter case does not matter, and the zone takes the name as the database spells it:

```ant
Time.zone("europe/stockholm") // Ok{ value: Europe/Stockholm }
Time.zone("utc")              // Ok{ value: UTC }
```

An alias is accepted and keeps its own name:

```ant
Time.zone("Asia/Calcutta") // Ok{ value: Asia/Calcutta }
Time.zone("US/Eastern")    // Ok{ value: US/Eastern }
```

An offset may be written with or without the colon, or as hours alone. The zone always renders as
`±HH:MM`:

```ant
Time.zone("+05:30") // Ok{ value: +05:30 }
Time.zone("+0530")  // Ok{ value: +05:30 }
Time.zone("-08")    // Ok{ value: -08:00 }
```

Anything else is an `Err`, the empty string included:

```ant
Time.zone("")      // Err{ error: Error{ message: "unknown time zone \"\"" } }
Time.zone("+5:30") // Err{ error: Error{ message: "unknown time zone \"+5:30\"" } }
Time.zone("Z")     // Err{ error: Error{ message: "unknown time zone \"Z\"" } }
Time.zone(" UTC")  // Err{ error: Error{ message: "unknown time zone \" UTC\"" } }
```

## Applications

Read a zone from user input inside a function that returns a `Result`, and say which setting was
wrong:

```ant
func meetingIn(city: string): Result[Time.ZonedDateTime, Error] {
  let zone = try (Time.zone(city) | Result.wrapErr("meeting zone"))
  let start = Time.parsePlainDateTime("2026-12-01T09:00").unwrap()
  Result.ok(start.toZoned(zone))
}

println(meetingIn("America/Chicago"))
println(meetingIn("Chicago"))
```

```text
Ok{ value: 2026-12-01T09:00:00-06:00[America/Chicago] }
Err{ error: Error{ message: "meeting zone: unknown time zone \"Chicago\"" } }
```

Fall back to UTC when a configured zone is missing or does not parse:

```ant
Time.zone("Europe/Lisbon").unwrapOr(Time.utc) // Europe/Lisbon
Time.zone("Lisbon").unwrapOr(Time.utc)        // UTC
```

Find the names in a list that are not zones:

```ant
let names = ["Asia/Tokyo", "Europe/Paris", "Pacific/Atlantis", "America/Sao_Paulo"]
names.filter({ !Time.zone(x).isOk() }) // ["Pacific/Atlantis"]
```

Show one moment on a clock in each office:

```ant
let launch = Time.parseInstant("2026-10-22T14:00:00Z").unwrap()
for city in ["America/Los_Angeles", "Asia/Kolkata", "+03:00"] {
  let zone = Time.zone(city).unwrap()
  println("${zone.zoneId()}: ${launch.inZone(zone).toPlainTime()}")
}
```

```text
America/Los_Angeles: 07:00:00
Asia/Kolkata: 19:30:00
+03:00: 17:00:00
```

## Properties and relations

The identifier of a zone parses back to an equal zone:

```ant
let zones = ["Europe/Stockholm", "asia/calcutta", "+0530", "-08"].map({ Time.zone(x).unwrap() })
assert(zones.every({ Time.zone(x.zoneId()).unwrap() == x }))
```

Aliases of one zone are equal, and so are spellings of one offset:

```ant
assert(Time.zone("Asia/Calcutta").unwrap() == Time.zone("Asia/Kolkata").unwrap())
assert(Time.zone("US/Eastern").unwrap() == Time.zone("America/New_York").unwrap())
assert(Time.zone("+01:00").unwrap() == Time.zone("+0100").unwrap())
assert(Time.zone("UTC").unwrap() == Time.utc)
```

## Possible issues

`Time.zone` returns a `Result`, not a zone. Unwrap it, or use `try`, before passing it on:

```ant
let t = Time.fromEpochMillis(0)
let local = t.inZone(Time.zone("Asia/Tokyo"))
// error: cannot use Result[Time.Zone, Error] where Time.Zone is expected
```

A named zone never equals an offset zone, even one with the same offset at every instant:

```ant
Time.zone("+00:00").unwrap() == Time.utc // false
```

Zones with the same rules today are not equal unless they are aliases of one zone:

```ant
Time.zone("Europe/Stockholm").unwrap() == Time.zone("Europe/Berlin").unwrap() // false
```

An offset zone has no daylight saving time. The same summer noon gets a different offset from the
named zone:

```ant
let noon = Time.parsePlainDateTime("2026-07-01T12:00").unwrap()
noon.toZoned(Time.zone("+01:00").unwrap())           // 2026-07-01T12:00:00+01:00[+01:00]
noon.toZoned(Time.zone("Europe/Stockholm").unwrap()) // 2026-07-01T12:00:00+02:00[Europe/Stockholm]
```

## See also

[Time.localZone](localZone.md) · [Time.zoneId](zoneId.md) · [Time.zoneOf](zoneOf.md) ·
[Time.inZone](inZone.md) · [Time.toZoned](toZoned.md) · [Time.Zone](types/Zone.md)
