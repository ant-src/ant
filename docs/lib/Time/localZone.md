<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.localZone

```ant signature
func Time.localZone(): Time.Zone
```

Time.localZone returns the host's zone. It is the zone the TZ environment variable names, with a
leading colon dropped, else the one the system's setting names; an offset in TZ, which POSIX reads
as west of UTC, is passed over. When neither names a zone of the database, it is the fixed offset
the host reports at the moment of the call.

<!-- end generated -->

## Details

- The result depends on the machine that runs the program. Code that must give the same answer
  everywhere takes a `Time.Zone` as a parameter instead.
- `TZ` counts only when it names a zone of the database, with a leading colon dropped as POSIX
  allows, so `TZ=:Europe/Stockholm` is Stockholm. An offset such as `TZ=+05:30`, which POSIX reads
  as 5:30 west of UTC, is passed over for the system's setting.

## Basic examples

Print the host's zone:

<!-- test: check-only the host's zone differs between machines -->

```ant
println(Time.localZone())
```

On a machine set to Stockholm it prints `Europe/Stockholm`.

Read the wall-clock time where the program runs:

<!-- test: check-only the clock and the host's zone change between runs -->

```ant
println(Time.nowIn(Time.localZone()))
```

It prints a line such as `2026-11-11T22:26:05.482178+01:00[Europe/Stockholm]`.

## Scope

The result is a `Time.Zone` like any other, so it goes wherever a zone does. Placing an instant in
the host's zone and taking the instant back gives the same instant, whatever the zone:

```ant
let epoch = Time.fromEpochMillis(0)
epoch.inZone(Time.localZone()).toInstant() == epoch // true
```

Its identifier is a name of the database or an offset, and either one parses back with `Time.zone`:

```ant
Time.zone(Time.localZone().zoneId()).isOk() // true
```

Set `TZ` to run a program in another zone:

```sh
TZ=Asia/Tokyo ant run main.ant
```

## Applications

Fall back to the host's zone when a configured zone does not parse:

<!-- test: check-only the host's zone differs between machines -->

```ant
println(Time.zone("Mars/Olympus").unwrapOr(Time.localZone()))
```

`Mars/Olympus` names no zone of the database, so it prints the host's zone, such as
`Europe/Stockholm`.

Stamp log lines with the local wall-clock time:

<!-- test: check-only the clock and the host's zone change between runs -->

```ant
func log(message: string) {
  let at = Time.nowIn(Time.localZone())
  println("${at.toPlainDateTime()} ${message}")
}

log("server started")
```

It prints a line such as `2026-11-11T22:26:05.482178 server started`.

Show a time given in UTC on the reader's clock:

<!-- test: check-only the host's zone differs between machines -->

```ant
let launch = Time.parseInstant("2026-12-01T15:00:00Z").unwrap()
let local = launch.inZone(Time.localZone())
println("Launch at ${local.toPlainTime()} (${local.offset()})")
```

In Stockholm it prints `Launch at 16:00:00 (+01:00)`; in Tokyo, `Launch at 00:00:00 (+09:00)`.

Keep functions testable: take the zone as a parameter, and pass `Time.localZone()` only at the edge
of the program:

```ant
func dayOf(t: Time.Instant, z: Time.Zone): Time.PlainDate {
  t.inZone(z).toPlainDate()
}

let t = Time.parseInstant("2026-11-11T23:30:00Z").unwrap()
dayOf(t, Time.utc)                         // 2026-11-11
dayOf(t, Time.zone("Asia/Tokyo").unwrap()) // 2026-11-12
```

## Properties and relations

`Time.nowIn(Time.localZone())` and `Time.now().inZone(Time.localZone())` show the same zone, and the
zone's identifier names it again:

```ant
let here = Time.localZone()
assert(Time.nowIn(here).zoneOf() == here)
assert(Time.now().inZone(here).zoneOf() == here)
assert(Time.zone(here.zoneId()).unwrap() == here)
```

## Possible issues

A test that calls `Time.localZone()` passes on one machine and fails on another, as on a CI runner
set to UTC. Pass the zone in, as `dayOf` above does.

When the host names no zone of the database, the result is a fixed offset, and an offset zone keeps
one offset all year. The same summer noon gets a different offset from a named zone:

```ant
let summer = Time.plainDateTime(Time.plainDate(2026, 7, 1), Time.plainTime(hour: 12))
summer.toZoned(Time.zone("+01:00").unwrap())           // 2026-07-01T12:00:00+01:00[+01:00]
summer.toZoned(Time.zone("Europe/Stockholm").unwrap()) // 2026-07-01T12:00:00+02:00[Europe/Stockholm]
```

`Time.localZone` is a function. Without the parentheses it is a function value, not a zone:

```ant
let now = Time.nowIn(Time.localZone) // error: cannot use func(): Time.Zone where Time.Zone is expected
```

A zone is not a string. Compare its identifier instead:

```ant
let here = Time.localZone()
println(here == "Europe/Stockholm")
// error: Time.Zone and string cannot be compared with ==: neither is assignable to the other
println(here.zoneId() == "Europe/Stockholm")
```

## See also

[Time.zone](zone.md) · [Time.zoneId](zoneId.md) · [Time.nowIn](nowIn.md) · [Time.inZone](inZone.md)
· [Time.zoneOf](zoneOf.md) · [Time.Zone](types/Zone.md) · [Os.env](../Os/env.md)
