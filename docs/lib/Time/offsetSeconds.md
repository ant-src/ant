<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.offsetSeconds

```ant signature
func Time.offsetSeconds(t: Time.ZonedDateTime): int
```

Time.offsetSeconds returns the offset of t from UTC in seconds, positive east of Greenwich.

<!-- end generated -->

## Details

- The offset is the one the zone has at the instant of `t`, so a zone with daylight-saving time
  gives a summer and a winter value across the year.
- The result is exact to the second, also for a historical offset such as Paris's `+00:09:21`. The
  string of `t` rounds its offset to the minute; this does not.
- A fixed-offset zone and UTC give the same value at every instant.

## Basic examples

```ant
let t = Time.parseZoned("2026-07-15T12:00[Europe/Stockholm]").unwrap()
Time.offsetSeconds(t) // 7200
```

West of Greenwich the offset is negative:

```ant
let t = Time.parseZoned("2026-01-15T12:00[America/New_York]").unwrap()
Time.offsetSeconds(t) // -18000
```

## Scope

A named zone uses the offset in force at the instant, including offsets that are not whole hours:

```ant
Time.parseZoned("2026-01-15T12:00[Europe/Stockholm]").unwrap().offsetSeconds()   // 3600
Time.parseZoned("2026-07-15T12:00[Asia/Kathmandu]").unwrap().offsetSeconds()     // 20700
Time.parseZoned("2026-07-15T12:00[Pacific/Kiritimati]").unwrap().offsetSeconds() // 50400
```

UTC and fixed-offset zones:

```ant
Time.parseZoned("2026-07-15T12:00[UTC]").unwrap().offsetSeconds()          // 0
Time.parseZoned("2026-07-15T12:00-03:30[-03:30]").unwrap().offsetSeconds() // -12600
```

Before standard time, a zone kept the local mean time of its city, with seconds in it:

```ant
let paris = Time.parseZoned("1900-01-01T12:00[Europe/Paris]").unwrap()
paris.offsetSeconds() // 561
paris.offset()        // "+00:09:21"
```

When a daylight-saving change repeats an hour, the two instants of `01:30` have different offsets:

```ant
let ny = Time.zone("America/New_York").unwrap()
let first = Time.parseInstant("2024-11-03T05:30:00Z").unwrap().inZone(ny)
let second = Time.parseInstant("2024-11-03T06:30:00Z").unwrap().inZone(ny)
first.offsetSeconds()  // -14400
second.offsetSeconds() // -18000
```

In a pipe, the zoned time is the first argument:

```ant
"2026-07-15T12:00[Asia/Tokyo]" | Time.parseZoned() | Result.map(Time.offsetSeconds) // Ok{ value: 32400 }
```

## Applications

Show how far ahead of a home zone each city is, in hours:

```ant
let at = Time.parseInstant("2026-07-15T12:00:00Z").unwrap()
let home = at.inZone(Time.zone("Europe/London").unwrap())
for city in ["America/Los_Angeles", "Asia/Kathmandu", "Australia/Adelaide"] {
  let local = at.inZone(Time.zone(city).unwrap())
  let hours = float64(local.offsetSeconds() - home.offsetSeconds()) / 3600.0
  println("${city}: ${hours}")
}
```

```text
America/Los_Angeles: -8.0
Asia/Kathmandu: 4.75
Australia/Adelaide: 8.5
```

Write an offset as a label such as `UTC+5:45`:

```ant
func label(t: Time.ZonedDateTime): string {
  let seconds = t.offsetSeconds()
  if seconds == 0 {
    return "UTC"
  }
  let sign = if seconds < 0 { "-" } else { "+" }
  let minutes = Math.abs(seconds) / 60
  let rest = String.from(minutes % 60).padStart(2, pad: "0")
  "UTC${sign}${minutes / 60}:${rest}"
}

let at = Time.parseInstant("2026-07-15T12:00:00Z").unwrap()
for city in ["UTC", "Asia/Kathmandu", "America/St_Johns", "Asia/Tokyo"] {
  println(label(at.inZone(Time.zone(city).unwrap())))
}
```

```text
UTC
UTC+5:45
UTC-2:30
UTC+9:00
```

Measure how far a zone moves its clocks at its next change:

```ant
let t = Time.parseZoned("2026-01-15T12:00[Australia/Lord_Howe]").unwrap()
let change = t.nextTransition()
if change != null {
  println("${change}: ${change.offsetSeconds() - t.offsetSeconds()} seconds")
}
```

```text
2026-04-05T01:30:00+10:30[Australia/Lord_Howe]: -1800 seconds
```

## Properties and relations

The offset is the distance from the instant to the wall-clock time read as UTC:

```ant
let t = Time.parseZoned("2026-07-15T12:00[Asia/Kathmandu]").unwrap()
let wall = Time.toZoned(t.toPlainDateTime(), Time.utc).toInstant()
assert(Time.until(t.toInstant(), wall).total(Time.Second{}) == float64(t.offsetSeconds()))
```

A fixed-offset zone made from [Time.offset](offset.md) has the same offset, when it is whole
minutes:

```ant
let t = Time.parseZoned("2026-07-15T12:00[America/St_Johns]").unwrap()
let fixed = t.withZone(Time.zone(t.offset()).unwrap())
assert(fixed.offsetSeconds() == t.offsetSeconds())
```

## Possible issues

The result is whole seconds, so integer division by 3600 drops the minutes of an offset such as
Kathmandu's. Convert to `float64` first:

```ant
let t = Time.parseZoned("2026-07-15T12:00[Asia/Kathmandu]").unwrap()
t.offsetSeconds() / 3600            // 5
float64(t.offsetSeconds()) / 3600.0 // 5.75
```

The string of a zoned time rounds a historical offset to the minute, so it can disagree with this
function:

```ant
let monrovia = Time.parseZoned("1970-01-01T12:00[Africa/Monrovia]").unwrap()
monrovia.offsetSeconds() // -2670
String.from(monrovia)    // "1970-01-01T12:00:00-00:45[Africa/Monrovia]"
```

An instant has no zone, so it has no offset. See it in a zone first with [Time.inZone](inZone.md):

```ant
let at = Time.parseInstant("2026-07-15T12:00:00Z").unwrap()
println(Time.offsetSeconds(at))
// error: cannot use Time.Instant where Time.ZonedDateTime is expected
```

## See also

[Time.offset](offset.md) · [Time.zoneOf](zoneOf.md) · [Time.zoneId](zoneId.md) ·
[Time.inZone](inZone.md) · [Time.withZone](withZone.md) · [Time.nextTransition](nextTransition.md) ·
[Time.hoursInDay](hoursInDay.md) · [Time.ZonedDateTime](types/ZonedDateTime.md)
