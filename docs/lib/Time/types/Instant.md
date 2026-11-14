<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.Instant

```ant signature
type Time.Instant
```

Time.Instant is a moment on the time line, counted in nanoseconds since the Unix epoch, with no zone
and no calendar. It renders as 2026-10-22T14:03:05.5Z and spans 100 million days each way from the
epoch.

<!-- end generated -->

## Details

- An instant has no zone and no calendar, so it has no year, hour, or other field. Put it in a zone
  with [`Time.inZone`](../inZone.md) to read the wall clock there.
- [`Time.add`](../add.md) and [`Time.subtract`](../subtract.md) take hours and smaller units only; a
  duration with years, months, weeks, or days panics. [`Time.until`](../until.md) between instants
  counts in seconds by default and in units up to hours at most.
- `<` and the other comparison operators do not apply. Order instants with
  [`Time.compare`](../compare.md); `==` tests the same moment.
- An instant renders in UTC with `Z`, its fractional second to the last nonzero digit, and a year
  outside 0 to 9999 as a sign and six digits. An offset in parsed text only locates the moment and
  is not kept.

## Basic examples

Parse an instant from text with an offset, or build one from an epoch count:

```ant
Time.parseInstant("2026-10-22T14:03:05Z")        // Ok{ value: 2026-10-22T14:03:05Z }
Time.parseInstant("2026-10-22T16:03:05.5+02:00") // Ok{ value: 2026-10-22T14:03:05.5Z }
Time.fromEpochMillis(0)                          // 1970-01-01T00:00:00Z
```

Read the current instant with [`Time.now`](../now.md):

<!-- test: check-only reads the clock -->

```ant
let started = Time.now()
println("started at ${started}")
```

Annotate a field with `Time.Instant`:

```ant
type Login = { user: string, at: Time.Instant }

let login = Login{ user: "ann", at: Time.fromEpochMillis(1792677785000) }
login.at // 2026-10-22T14:03:05Z
```

## Scope

### Making an instant

Build an instant from text, an epoch count in milliseconds or nanoseconds, or the moment a zoned
time shows:

```ant
let zoned = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]") | Result.unwrap()
Time.parseInstant("2026-10-22T14:03:05Z") | Result.unwrap() // 2026-10-22T14:03:05Z
Time.fromEpochMillis(1792677785000)                         // 2026-10-22T14:03:05Z
Time.fromEpochNanos(1792677785000000000)                    // 2026-10-22T14:03:05Z
zoned.toInstant()                                           // 2026-10-22T14:03:05Z
```

Text without an offset names no moment, so it is an error. A bracketed zone after the offset is
ignored:

```ant
Time.parseInstant("2026-10-22T14:03:05") is Err             // true
Time.parseInstant("2026-10-22T14:03:05Z[Europe/Stockholm]") // Ok{ value: 2026-10-22T14:03:05Z }
```

### Epoch counts

Read an instant back out as a Unix timestamp in milliseconds or nanoseconds:

```ant
let t = Time.fromEpochMillis(1792677785250)
t.epochMillis() // 1792677785250
t.epochNanos()  // 1792677785250000000
```

### Rendering

String interpolation, `String.from`, and `Json.stringify` all write the same text:

```ant
let t = Time.fromEpochMillis(1792677785250)
"logged at ${t}"  // "logged at 2026-10-22T14:03:05.25Z"
String.from(t)    // "2026-10-22T14:03:05.25Z"
Json.stringify(t) // "\"2026-10-22T14:03:05.25Z\""
```

Times before the epoch and far from it render the same way, with a six-digit year outside 0 to 9999:

```ant
Time.fromEpochMillis(-1)                                       // 1969-12-31T23:59:59.999Z
Time.parseInstant("-001000-01-01T00:00:00Z") | Result.unwrap() // -001000-01-01T00:00:00Z
Time.fromEpochMillis(8640000000000000)                         // +275760-09-13T00:00:00Z
```

### Arithmetic

Add an exact duration, measure the span between two instants, and round:

```ant
let start = Time.parseInstant("2026-10-22T14:03:05Z") | Result.unwrap()
let end = start.add(25h1m)
end                                        // 2026-10-23T15:04:05Z
start.until(end)                           // PT90060S
start.until(end, largestUnit: Time.Hour{}) // PT25H1M
end.round(Time.Hour{})                     // 2026-10-23T15:00:00Z
```

### Zones

An instant shows a different wall clock in each zone, and every zone gives back the same instant:

```ant
let t = Time.parseInstant("2026-10-22T14:03:05Z") | Result.unwrap()
let tokyo = Time.zone("Asia/Tokyo") | Result.unwrap()
t.inZone(tokyo)                // 2026-10-22T23:03:05+09:00[Asia/Tokyo]
t.inZone(tokyo).hour()         // 23
t.inZone(Time.utc).toInstant() // 2026-10-22T14:03:05Z
```

### Unions and collections

An instant is one of the members of `Time.Point`, and an `is` test narrows to it:

```ant
func describe(p: Time.Point): string {
  if p is Time.Instant { "the moment ${p}" } else { "a wall-clock value ${p}" }
}

describe(Time.fromEpochMillis(0))      // "the moment 1970-01-01T00:00:00Z"
describe(Time.plainDate(2026, 10, 22)) // "a wall-clock value 2026-10-22"
```

Instants compare by value, so one moment counts once in a set:

```ant
let a = Time.fromEpochMillis(0)
let b = Time.fromEpochNanos(0)
Set.of([a, b]).size() // 1
```

## Applications

Show one moment to readers in several zones:

```ant
let launch = Time.parseInstant("2026-10-22T14:00:00Z") | Result.unwrap()
for id in ["America/New_York", "Europe/Stockholm", "Asia/Kolkata"] {
  let zone = Time.zone(id) | Result.unwrap()
  let local = launch.inZone(zone)
  println("${id}: ${local.toPlainTime()}")
}
```

```text
America/New_York: 10:00:00
Europe/Stockholm: 16:00:00
Asia/Kolkata: 19:30:00
```

Sort events by when they happened:

```ant
type Event = { name: string, at: Time.Instant }

let events = [
  Event{ name: "deploy", at: Time.parseInstant("2026-10-22T14:05:00Z") | Result.unwrap() },
  Event{ name: "build", at: Time.parseInstant("2026-10-22T15:58:00+02:00") | Result.unwrap() },
  Event{ name: "alert", at: Time.parseInstant("2026-10-22T14:20:00Z") | Result.unwrap() },
]
events.sortWith(func(a: Event, b: Event) { Time.compare(a.at, b.at) }).map({ x.name }) // ["build", "deploy", "alert"]
```

Check whether a token has expired, given the moment it was issued and its lifetime:

```ant
func expired(issued: Time.Instant, lifetime: Time.Duration, now: Time.Instant): bool {
  Time.compare(now, issued.add(lifetime)) >= 0
}

let issued = Time.parseInstant("2026-10-22T14:00:00Z") | Result.unwrap()
let lifetime = 15m
expired(issued, lifetime, Time.parseInstant("2026-10-22T14:10:00Z") | Result.unwrap()) // false
expired(issued, lifetime, Time.parseInstant("2026-10-22T14:15:00Z") | Result.unwrap()) // true
```

Count requests per hour, rounding each instant down to the start of its hour:

```ant
let stamps = ["2026-10-22T14:03:05Z", "2026-10-22T14:59:59Z", "2026-10-22T15:00:00Z"]
let requests = stamps.map({ Time.parseInstant(x) | Result.unwrap() })
requests.groupBy({ x.round(Time.Hour{}, mode: Time.Floor{}) }).map({ x.value.length() }) // Map{ 2026-10-22T14:00:00Z: 2, 2026-10-22T15:00:00Z: 1 }
```

Stamp a log line with the moment it was written:

<!-- test: check-only reads the clock -->

```ant
func log(message: string) {
  println("${Time.now()} ${message}")
}

log("server started")
```

## Properties and relations

An instant renders to text that parses back to it, and its epoch counts round-trip:

```ant
let t = Time.parseInstant("2026-10-22T14:03:05.123456789Z") | Result.unwrap()
let back = Time.parseInstant(String.from(t)) | Result.unwrap()
assert(back == t)
assert(Time.fromEpochNanos(t.epochNanos()) == t)
assert(t.inZone(Time.zone("Asia/Tokyo") | Result.unwrap()).toInstant() == t)
```

`==` agrees with `Time.compare`, and adding the span between two instants gets from one to the
other:

```ant
let a = Time.parseInstant("2026-10-22T14:03:05Z") | Result.unwrap()
let b = Time.parseInstant("2026-10-22T16:03:05+02:00") | Result.unwrap()
let c = Time.parseInstant("2026-10-23T01:00:00Z") | Result.unwrap()
assert(a == b && Time.compare(a, b) == 0)
assert(a.add(a.until(c)) == c)
assert(Time.compare(a, c) == -Time.compare(c, a))
```

## Possible issues

An instant has no calendar, so a day cannot be added to it. Add 24 hours, or put the instant in a
zone, where a day follows the clock:

```ant
let t = Time.fromEpochMillis(0)
t.add(24h) // 1970-01-02T00:00:00Z
t.add(Time.duration(days: 1))
// panic: Time.add: an instant has no calendar, so it takes no days; convert it with Time.inZone first
t.until(t, largestUnit: Time.Day{}) // panic: Time.until: an instant has no days
```

An instant has no fields; read them from a zoned time:

```ant
let t = Time.fromEpochMillis(0)
let year = t.year()
// error: Time.year takes Time.Dated first, so a Time.Instant does not reach it
let hour = Time.hour(t)
// error: cannot use Time.Instant where Time.Timed is expected
```

`<` does not order instants. Use `Time.compare`:

```ant
let a = Time.fromEpochMillis(0)
let b = Time.fromEpochMillis(1)
let earlier = a < b
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.Instant
```

The range is wide but finite, and a result past its edge panics:

```ant
let last = Time.fromEpochMillis(8640000000000000)
last.add(1ns) // panic: Time.add: outside the range of time values
```

`epochMillis` always fits an `int`, but `epochNanos` reaches only 1677-09-21 to 2262-04-11 and
panics outside:

```ant
let last = Time.fromEpochMillis(8640000000000000)
last.epochMillis() // 8640000000000000
last.epochNanos()
// panic: Time.epochNanos: the nanoseconds do not fit an int; the instant is outside 1677 to 2262
```

An instant cannot be rounded to a day, since it has no calendar. Round to 24 hours for a midnight of
UTC:

```ant
let t = Time.fromEpochMillis(1792677785250)
t.round(Time.Day{})                 // panic: Time.round: an instant cannot be rounded to a day
t.round(Time.Hour{}, increment: 24) // 2026-10-23T00:00:00Z
```

## See also

[Time.now](../now.md), [Time.parseInstant](../parseInstant.md),
[Time.fromEpochMillis](../fromEpochMillis.md), [Time.fromEpochNanos](../fromEpochNanos.md),
[Time.epochMillis](../epochMillis.md), [Time.inZone](../inZone.md),
[Time.toInstant](../toInstant.md), [Time.compare](../compare.md), [Time.until](../until.md),
[Time.ZonedDateTime](ZonedDateTime.md), [Time.Duration](Duration.md)
