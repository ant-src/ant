<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.now

```ant signature
func Time.now(): Time.Instant
```

Time.now returns the current instant from the wall clock, which the host may set forward or back, so
a later call can return an earlier instant; measure elapsed time with
[`Time.monotonic`](monotonic.md).

<!-- end generated -->

## Details

- The result has no zone and no calendar. Fields such as the year or the hour need a zone: use
  [Time.nowIn](nowIn.md), or [Time.inZone](inZone.md) on the result.
- To measure how long something takes, use [Time.monotonic](monotonic.md), which is meant for that.

## Basic examples

<!-- test: check-only the clock changes from run to run -->

```ant
let now = Time.now()
println(now)
println(now.inZone(Time.zone("Asia/Tokyo").unwrap()))
```

It prints two lines such as `2026-11-11T21:27:07.050436Z` and
`2026-11-12T06:27:07.050436+09:00[Asia/Tokyo]`: the instant, then the same instant seen in Tokyo.

The value changes on every run, but how it compares to a fixed instant does not:

```ant
Time.compare(Time.now(), Time.fromEpochMillis(0)) // 1
```

## Applications

Tell whether a token has expired, by comparing now with the time it was issued plus its lifetime:

```ant
type Token = { value: string, issued: Time.Instant }

func expired(t: Token, ttl: Time.Duration): bool {
  Time.compare(Time.now(), t.issued.add(ttl)) > 0
}

let old = Token{ value: "a1", issued: Time.parseInstant("2020-01-01T00:00:00Z").unwrap() }
let fresh = Token{ value: "b2", issued: Time.now() }
println(expired(old, 1h))
println(expired(fresh, 1h))
```

```text
true
false
```

Stamp each log line with the time, as milliseconds from the Unix epoch:

<!-- test: check-only the clock changes from run to run -->

```ant
type LogLine = { at: int, message: string }

func log(message: string) {
  println(Json.stringify(LogLine{ at: Time.now().epochMillis(), message }))
}

log("started")
```

It prints a line such as `{"at":1794432427050,"message":"started"}`.

Keep trying until a deadline passes:

```ant
let deadline = Time.now().add(50ms)
var tries = 0
while Time.compare(Time.now(), deadline) < 0 {
  tries += 1
  Time.sleep(10ms)
}
println(tries > 0)
```

```text
true
```

## Properties and relations

[Time.inZone](inZone.md) shows the same instant in a zone, and [Time.toInstant](toInstant.md) gives
it back:

```ant
let now = Time.now()
let tokyo = now.inZone(Time.zone("Asia/Tokyo").unwrap())
assert(tokyo.toInstant() == now)
```

The current time lies within the range of [Time.epochNanos](epochNanos.md), so it round-trips
through nanoseconds:

```ant
let now = Time.now()
assert(Time.fromEpochNanos(now.epochNanos()) == now)
```

## Possible issues

The comparison operators do not apply to time values. Use [Time.compare](compare.md):

```ant
let deadline = Time.parseInstant("2030-01-01T00:00:00Z").unwrap()
let early = Time.now() < deadline
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.Instant
```

An instant has no year, month or hour until it is seen in a zone, so asking for one is a compile
error:

```ant
let year = Time.year(Time.now()) // error: cannot use Time.Instant where Time.Dated is expected
```

Ask [Time.nowIn](nowIn.md) for the current time in a zone instead:

<!-- test: check-only the clock changes from run to run -->

```ant
let year = Time.nowIn(Time.utc).year()
println(year)
```

Each call reads the clock again. When several values must agree on one moment, such as the start and
the end of a window, read the clock once and use that value:

```ant
let now = Time.now()
let start = now.subtract(1h)
println(Time.until(start, now, largestUnit: Time.Hour{}))
```

```text
PT1H
```

## See also

[Time.nowIn](nowIn.md) · [Time.monotonic](monotonic.md) · [Time.inZone](inZone.md) ·
[Time.compare](compare.md) · [Time.epochMillis](epochMillis.md) · [Time.since](since.md) ·
[Time.Instant](types/Instant.md)
