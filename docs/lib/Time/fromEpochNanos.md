<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.fromEpochNanos

```ant signature
func Time.fromEpochNanos(ns: int): Time.Instant
```

Time.fromEpochNanos returns the instant ns nanoseconds after the Unix epoch, or before it when ns is
negative.

<!-- end generated -->

## Details

- Every `int` gives an instant, so `Time.fromEpochNanos` never panics. The range of an `int` covers
  1677-09-21 to 2262-04-11; for times outside it, count milliseconds with
  [Time.fromEpochMillis](fromEpochMillis.md).
- The instant has no zone. Put it in one with [Time.inZone](inZone.md) to read its date and time of
  day.

## Basic examples

```ant
Time.fromEpochNanos(0)                         // 1970-01-01T00:00:00Z
Time.fromEpochNanos(1_700_000_000_123_456_789) // 2023-11-14T22:13:20.123456789Z
```

A negative count is before the epoch:

```ant
Time.fromEpochNanos(-1) // 1969-12-31T23:59:59.999999999Z
```

## Scope

The smallest and largest `int` are the two ends of what it can name:

```ant
Time.fromEpochNanos(Math.minInt) // 1677-09-21T00:12:43.145224192Z
Time.fromEpochNanos(Math.maxInt) // 2262-04-11T23:47:16.854775807Z
```

Pass it as a function value to convert a list of timestamps:

```ant
[0, 1_000_000_000, 1_500_000_000].map(Time.fromEpochNanos)
// [1970-01-01T00:00:00Z, 1970-01-01T00:00:01Z, 1970-01-01T00:00:01.5Z]
```

Every digit is kept, down to the nanosecond:

```ant
Time.fromEpochNanos(1_700_000_000_123_456_789).inZone(Time.utc).nanosecond() // 789
```

## Named arguments

```ant
Time.fromEpochNanos(ns: 1_000_000_000) // 1970-01-01T00:00:01Z
```

## Applications

Show a nanosecond timestamp from a trace in a local zone:

```ant
let stockholm = Time.zone("Europe/Stockholm").unwrap()
let started = Time.fromEpochNanos(1_762_869_600_000_000_000)
started.inZone(stockholm) // 2025-11-11T15:00:00+01:00[Europe/Stockholm]
```

Measure the time between two spans of a trace:

```ant
let start = Time.fromEpochNanos(1_762_869_600_000_000_000)
let end = Time.fromEpochNanos(1_762_869_600_250_000_000)
Time.until(start, end) // PT0.25S
```

Sort log lines by their timestamp and print them:

```ant
type LogLine = { at: int, message: string }

let lines = [
  LogLine{ at: 1_762_869_601_000_000_000, message: "done" },
  LogLine{ at: 1_762_869_600_000_000_000, message: "start" },
]
for line in lines.sortBy({ x.at }) {
  println("${Time.fromEpochNanos(line.at)} ${line.message}")
}
```

```text
2025-11-11T14:00:00Z start
2025-11-11T14:00:01Z done
```

## Properties and relations

`Time.epochNanos` turns the instant back into the count:

```ant
let ns = 1_700_000_000_123_456_789
assert(Time.fromEpochNanos(ns).epochNanos() == ns)
```

A million nanoseconds are a millisecond:

```ant
let ms = 1_700_000_000_123
assert(Time.fromEpochNanos(ms * 1_000_000) == Time.fromEpochMillis(ms))
```

`Time.epochMillis` rounds down, so a nanosecond before the epoch is millisecond -1:

```ant
Time.fromEpochNanos(-1).epochMillis() // -1
```

## Possible issues

A count in milliseconds or seconds, passed as nanoseconds, lands in January 1970:

```ant
Time.fromEpochNanos(1_700_000_000_123) // 1970-01-01T00:28:20.000000123Z
```

The argument is an `int`; a `float64` does not compile:

```ant
let at = Time.fromEpochNanos(1.7e18)
// error: cannot use float64 where int is expected
```

## See also

[Time.fromEpochMillis](fromEpochMillis.md) · [Time.epochNanos](epochNanos.md) ·
[Time.epochMillis](epochMillis.md) · [Time.now](now.md) · [Time.inZone](inZone.md) ·
[Time.Instant](types/Instant.md)
