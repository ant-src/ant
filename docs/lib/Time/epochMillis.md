<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.epochMillis

```ant signature
func Time.epochMillis(t: Time.Instant | Time.ZonedDateTime): int
```

Time.epochMillis returns the milliseconds from the Unix epoch to t, rounded down. Every instant
fits.

<!-- end generated -->

## Details

- The result rounds down, toward the past, so a time a fraction of a millisecond before the epoch
  gives `-1`, not `0`.
- It never panics: every instant, 100 million days each way from the epoch, fits in an `int` of
  milliseconds. [Time.epochNanos](epochNanos.md) reaches only 1677 to 2262.
- A `ZonedDateTime` gives the count for its instant; the zone does not change it.

## Basic examples

```ant
let t = Time.parseInstant("2026-10-22T14:03:05.123Z").unwrap()
Time.epochMillis(t)                       // 1792677785123
Time.epochMillis(Time.fromEpochMillis(0)) // 0
```

Call it through `.`:

```ant
let t = Time.parseInstant("1970-01-01T00:00:01Z").unwrap()
t.epochMillis() // 1000
```

## Scope

A `ZonedDateTime` gives the milliseconds of the instant it shows, in any zone:

```ant
let z = Time.parseZoned("2026-10-22T16:03:05.123+02:00[Europe/Stockholm]").unwrap()
z.epochMillis() // 1792677785123
```

Nanoseconds below a millisecond are dropped, rounding toward the past:

```ant
Time.fromEpochNanos(1_999_999).epochMillis()  // 1
Time.fromEpochNanos(-1).epochMillis()         // -1
Time.fromEpochNanos(-1_500_000).epochMillis() // -2
```

Instants before 1970 give negative counts:

```ant
let landing = Time.parseInstant("1969-07-20T20:17:40Z").unwrap()
landing.epochMillis() // -14182940000
```

The whole range of an instant fits, far beyond what [Time.epochNanos](epochNanos.md) reaches:

```ant
Time.parseInstant("+275760-09-13T00:00:00Z").unwrap().epochMillis() // 8640000000000000
Time.parseInstant("-271821-04-20T00:00:00Z").unwrap().epochMillis() // -8640000000000000
```

An array of both types, typed as the union, takes it as a function value:

```ant
let points: Array[Time.Instant | Time.ZonedDateTime] = [
  Time.fromEpochMillis(5),
  Time.fromEpochMillis(7).inZone(Time.utc),
]
points.map(Time.epochMillis) // [5, 7]
```

## Applications

Stamp a record with the current time:

```ant
let stamp = Time.now().epochMillis()
assert(stamp > 0)
```

Write a time as a JavaScript timestamp, the milliseconds that `Date.now()` and `new Date(ms)` use:

```ant
type Event = { name: string, at: int }

let at = Time.parseInstant("2026-10-22T14:03:05.123Z").unwrap()
println(Json.stringify(Event{ name: "deploy", at: at.epochMillis() }))
```

```text
{"name":"deploy","at":1792677785123}
```

Group events into five-minute windows by the window each one starts in:

```ant
let window = 5 * 60 * 1000
let events = [
  "2026-10-22T14:01:10Z",
  "2026-10-22T14:04:59Z",
  "2026-10-22T14:05:00Z",
  "2026-10-22T14:12:30Z",
]

let windows = events
  | Array.map({ Time.parseInstant(x).unwrap() })
  | Array.groupBy({ x.epochMillis() / window * window })

for { key, value } in windows {
  println("${Time.fromEpochMillis(key)}: ${value.length()}")
}
```

```text
2026-10-22T14:00:00Z: 2
2026-10-22T14:05:00Z: 1
2026-10-22T14:10:00Z: 1
```

## Properties and relations

[Time.fromEpochMillis](fromEpochMillis.md) gives back the same instant when it has no part below a
millisecond:

```ant
let t = Time.parseInstant("2026-10-22T14:03:05.123Z").unwrap()
assert(Time.fromEpochMillis(t.epochMillis()) == t)
```

Otherwise the round trip is the instant rounded down to the millisecond, as [Time.round](round.md)
with `Time.Floor{}` gives it:

```ant
let t = Time.parseInstant("2026-10-22T14:03:05.123456789Z").unwrap()
let floored = Time.round(t, Time.Millisecond{}, mode: Time.Floor{})
assert(Time.fromEpochMillis(t.epochMillis()) == floored)
```

A `ZonedDateTime` gives the same count as its instant:

```ant
let t = Time.parseInstant("2026-10-22T14:03:05.123Z").unwrap()
let tokyo = t.inZone(Time.zone("Asia/Tokyo").unwrap())
assert(tokyo.epochMillis() == t.epochMillis())
assert(tokyo.epochMillis() == tokyo.toInstant().epochMillis())
```

## Possible issues

A wall-clock time names no instant until it has a zone, so a plain type is a compile error:

```ant
let meeting = Time.parsePlainDateTime("2026-10-22T14:00").unwrap()
let ms = Time.epochMillis(meeting)
// error: cannot use Time.PlainDateTime where Time.Instant | Time.ZonedDateTime is expected
```

Place it in a zone first:

```ant
let meeting = Time.parsePlainDateTime("2026-10-22T14:00").unwrap()
meeting.toZoned(Time.utc).epochMillis() // 1792677600000
```

Dividing [Time.epochNanos](epochNanos.md) by a million is not the same before 1970: `/` on an `int`
truncates toward zero, while `Time.epochMillis` rounds down:

```ant
let t = Time.fromEpochNanos(-1_500_000)
t.epochNanos() / 1_000_000 // -1
t.epochMillis()            // -2
```

The same truncation applies to dividing the result. `t.epochMillis() / 1000` rounds a time before
1970 toward the epoch, not down to its second:

```ant
let t = Time.fromEpochMillis(-1_500)
t.epochMillis() / 1000 // -1
```

`Time.epochMillis` drops the nanoseconds below a millisecond, so two different instants can give the
same count:

```ant
let a = Time.fromEpochNanos(1_000_000)
let b = Time.fromEpochNanos(1_999_999)
a == b                             // false
a.epochMillis() == b.epochMillis() // true
```

To time code, do not subtract two `Time.now().epochMillis()` readings: they read the time of day.
[Time.monotonic](monotonic.md) is the reading for measuring, milliseconds from an arbitrary origin:

```ant
let start = Time.monotonic()
var total = 0
for x in Iter.range(1, 1001) {
  total += x
}
let elapsed = Time.monotonic() - start
println("summed to ${total} in ${elapsed} ms")
```

## See also

[Time.fromEpochMillis](fromEpochMillis.md) · [Time.epochNanos](epochNanos.md) ·
[Time.fromEpochNanos](fromEpochNanos.md) · [Time.now](now.md) · [Time.monotonic](monotonic.md) ·
[Time.round](round.md) · [Time.Instant](types/Instant.md)
