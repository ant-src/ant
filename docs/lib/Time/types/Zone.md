<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.Zone

```ant signature
type Time.Zone
```

Time.Zone is a time zone: a zone of the IANA database, such as Europe/Stockholm, or a fixed offset,
such as +05:30. Two named zones are equal when they are aliases of one zone, two offset zones when
their offsets are, and a named zone never equals an offset zone. See [`Time.zone`](../zone.md).

<!-- end generated -->

## Details

- There is no literal for a zone. [Time.zone](../zone.md) makes one from a name or an offset and
  returns a `Result`; `Time.utc` and [Time.localZone](../localZone.md) give one directly.
- A zone renders as its identifier, the same text [Time.zoneId](../zoneId.md) returns, in
  `String.from`, interpolation and `Json.stringify`.
- `Time.utc` is a named zone, so it never equals `+00:00`.
- A named zone follows its daylight-saving rules; an offset zone keeps one offset all year.
- A zone is immutable and works as a map key.

## Basic examples

Look up a zone of the IANA database by name with [Time.zone](../zone.md):

```ant
Time.zone("Europe/Stockholm") // Ok{ value: Europe/Stockholm }
Time.zone("Mars/Olympus")     // Err{ error: Error{ message: "unknown time zone \"Mars/Olympus\"" } }
```

Make a fixed-offset zone from an offset:

```ant
Time.zone("+05:30") // Ok{ value: +05:30 }
```

Use a zone to see an instant on a wall clock:

```ant
let stockholm = Time.zone("Europe/Stockholm").unwrap()
let launch = Time.parseInstant("2026-07-01T12:00:00Z").unwrap()
launch.inZone(stockholm) // 2026-07-01T14:00:00+02:00[Europe/Stockholm]
```

Get the host's zone with [Time.localZone](../localZone.md), and the current time in a zone with
[Time.nowIn](../nowIn.md):

<!-- test: check-only the local zone and the current time depend on the host -->

```ant
let here = Time.localZone()
println("local zone: ${here}")
let stockholm = Time.zone("Europe/Stockholm").unwrap()
println(Time.nowIn(stockholm))
```

## Scope

A name is accepted in any letter case, and an offset as `±HH:MM`, `±HHMM` or `±HH`. The zone renders
as the database spells the name, and an offset always as `±HH:MM`:

```ant
Time.zone("europe/stockholm") // Ok{ value: Europe/Stockholm }
Time.zone("+0530")            // Ok{ value: +05:30 }
Time.zone("-03")              // Ok{ value: -03:00 }
Time.zone("")                 // Err{ error: Error{ message: "unknown time zone \"\"" } }
```

`Time.utc` is the named zone UTC:

```ant
Time.utc                              // UTC
Time.utc == Time.zone("utc").unwrap() // true
```

A zone turns into text wherever a string is needed:

```ant
let ny = Time.zone("America/New_York").unwrap()
ny.zoneId()        // "America/New_York"
String.from(ny)    // "America/New_York"
"meeting in ${ny}" // "meeting in America/New_York"
Json.stringify(ny) // "\"America/New_York\""
```

An alias keeps its own name, as the database spells it, but equals the zone it names:

```ant
let calcutta = Time.zone("Asia/Calcutta").unwrap()
let kolkata = Time.zone("Asia/Kolkata").unwrap()
calcutta                                                                   // Asia/Calcutta
calcutta == kolkata                                                        // true
Time.zone("US/Eastern").unwrap() == Time.zone("America/New_York").unwrap() // true
```

Two offset zones are equal when their offsets are, however they were written. A named zone and an
offset zone are never equal, even when they show the same offset now:

```ant
Time.zone("+0530").unwrap() == Time.zone("+05:30").unwrap()  // true
Time.zone("+00:00").unwrap() == Time.zone("-00:00").unwrap() // true
Time.utc == Time.zone("+00:00").unwrap()                     // false
```

A named zone changes its offset with daylight saving time. An offset zone does not, and has no
transitions:

```ant
let ny = Time.zone("America/New_York").unwrap()
let fixed = Time.zone("-05:00").unwrap()
let july = Time.parseInstant("2026-07-15T12:00:00Z").unwrap()
july.inZone(ny)                     // 2026-07-15T08:00:00-04:00[America/New_York]
july.inZone(fixed)                  // 2026-07-15T07:00:00-05:00[-05:00]
july.inZone(ny).nextTransition()    // 2026-11-01T01:00:00-05:00[America/New_York]
july.inZone(fixed).nextTransition() // null
```

Zones are keys of a map. Aliases of one zone land on one entry:

```ant
let teams: Map[Time.Zone, string] = Map{}
teams[Time.zone("Asia/Calcutta").unwrap()] = "support"
teams[Time.zone("Asia/Kolkata").unwrap()] = "platform"
teams.size()                              // 1
teams[Time.zone("Asia/Kolkata").unwrap()] // "platform"
```

## Applications

Show one meeting on the clocks of everyone invited:

```ant
let meeting = Time.parseZoned("2026-10-22T16:00[Europe/Stockholm]").unwrap()
let ids = ["America/New_York", "Asia/Kolkata", "Australia/Sydney"]
for id in ids {
  let zone = Time.zone(id).unwrap()
  let local = meeting.withZone(zone)
  println("${zone}: ${local.toPlainDateTime()}")
}
```

```text
America/New_York: 2026-10-22T10:00:00
Asia/Kolkata: 2026-10-22T19:30:00
Australia/Sydney: 2026-10-23T01:00:00
```

Read a zone from configuration, and say which setting was wrong when it is not one:

```ant
type Settings = { zone: Time.Zone }

func parseSettings(fields: Map[string, string]): Result[Settings, Error] {
  let id = fields.get("zone") ?? "UTC"
  let zone = try (Time.zone(id) | Result.wrapErr("setting zone"))
  Result.ok(Settings{ zone })
}

println(parseSettings(Map{ zone: "Europe/Stockholm" }))
println(parseSettings(Map{}))
println(parseSettings(Map{ zone: "CEST" }))
```

```text
Ok{ value: Settings{ zone: Europe/Stockholm } }
Ok{ value: Settings{ zone: UTC } }
Err{ error: Error{ message: "setting zone: unknown time zone \"CEST\"" } }
```

## Properties and relations

The identifier of a zone parses back to the same zone, and is what the zone renders as:

```ant
for id in ["Asia/Calcutta", "europe/stockholm", "+0530", "UTC"] {
  let z = Time.zone(id).unwrap()
  assert(Time.zone(z.zoneId()).unwrap() == z)
  assert(String.from(z) == z.zoneId())
}
```

[Time.zoneOf](../zoneOf.md) gives back the zone an instant was seen in:

```ant
let z = Time.zone("Pacific/Auckland").unwrap()
let i = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
assert(i.inZone(z).zoneOf() == z)
assert(i.inZone(z).withZone(Time.utc).zoneOf() == Time.utc)
```

## Possible issues

`Time.zone` returns a `Result`, not a zone. Unwrap it, or handle the error with `try`:

```ant
let ny = Time.zone("America/New_York")
let now = Time.nowIn(ny) // error: cannot use Result[Time.Zone, Error] where Time.Zone is expected
```

A zone has no literal, and a string is not a zone:

```ant
let a = Time.Zone{}                   // error: Time.Zone is not a record type, so it has no literal
let b: Time.Zone = "Europe/Stockholm" // error: cannot use string where Time.Zone is expected
```

An abbreviation such as `CEST` is not a name of the database; write the full name, or an offset:

```ant
Time.zone("CEST") // Err{ error: Error{ message: "unknown time zone \"CEST\"" } }
```

Two named zones with the same rules today are still different zones:

```ant
Time.zone("Europe/Stockholm").unwrap() == Time.zone("Europe/Berlin").unwrap() // false
```

A fixed offset does not follow daylight saving time. In January `-05:00` matches New York, in July
it is an hour behind; keep the named zone when you mean a place:

```ant
let ny = Time.zone("America/New_York").unwrap()
let fixed = Time.zone("-05:00").unwrap()
let july = Time.parsePlainDateTime("2026-07-15T09:00").unwrap()
july.toZoned(ny).toInstant() == july.toZoned(fixed).toInstant() // false
```

Zones have no order, so `<` does not apply to them:

```ant
let a = Time.utc
let b = Time.zone("+01:00").unwrap()
let before = a < b
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.Zone; narrow first
```

## See also

[Time.zone](../zone.md) · [Time.zoneId](../zoneId.md) · [Time.localZone](../localZone.md) ·
[Time.zoneOf](../zoneOf.md) · [Time.inZone](../inZone.md) · [Time.withZone](../withZone.md) ·
[Time.toZoned](../toZoned.md) · [Time.nowIn](../nowIn.md) · [Time.ZonedDateTime](ZonedDateTime.md)
