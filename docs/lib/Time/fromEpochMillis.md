<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.fromEpochMillis

```ant signature
func Time.fromEpochMillis(ms: int): Time.Instant
```

Time.fromEpochMillis returns the instant ms milliseconds after the Unix epoch, or before it when ms
is negative. It panics if the result is beyond the range of an instant, 100 million days each way.

<!-- end generated -->

## Details

- The Unix epoch is 1970-01-01T00:00:00Z.
- The result is an instant, with no zone. Place it in a zone with `Time.inZone` to read its date and
  clock time.
- The range is ±8,640,000,000,000,000 milliseconds; a count beyond it panics.

## Basic examples

```ant
Time.fromEpochMillis(0)                 // 1970-01-01T00:00:00Z
Time.fromEpochMillis(1_793_966_400_000) // 2026-11-06T12:00:00Z
```

A count that is not a whole second keeps its milliseconds:

```ant
Time.fromEpochMillis(1_500) // 1970-01-01T00:00:01.5Z
```

## Scope

A negative count counts back from the epoch:

```ant
Time.fromEpochMillis(-1)          // 1969-12-31T23:59:59.999Z
Time.fromEpochMillis(-86_400_000) // 1969-12-31T00:00:00Z
```

The ends of the range are 100 million days from the epoch:

```ant
Time.fromEpochMillis(8_640_000_000_000_000)  // +275760-09-13T00:00:00Z
Time.fromEpochMillis(-8_640_000_000_000_000) // -271821-04-20T00:00:00Z
```

Pass it as a function value:

```ant
[0, 60_000, 3_600_000].map(Time.fromEpochMillis)
// [1970-01-01T00:00:00Z, 1970-01-01T00:01:00Z, 1970-01-01T01:00:00Z]
```

## Named arguments

The parameter is named `ms`:

```ant
Time.fromEpochMillis(ms: 86_400_000) // 1970-01-02T00:00:00Z
```

## Applications

Show a timestamp from a JavaScript `Date.now()` as local time in a zone:

```ant
let stockholm = Result.unwrap(Time.zone("Europe/Stockholm"))
Time.fromEpochMillis(1_793_966_400_000).inZone(stockholm)
// 2026-11-06T13:00:00+01:00[Europe/Stockholm]
```

Measure the time between two stored timestamps:

```ant
let started = Time.fromEpochMillis(1_793_966_400_000)
let finished = Time.fromEpochMillis(1_793_966_490_500)
Time.until(started, finished) // PT90.5S
```

Print a log of events stored with millisecond timestamps, oldest first:

```ant
type Event = { name: string, at: int }

let events = [
  Event{ name: "deploy", at: 1_793_966_400_000 },
  Event{ name: "build", at: 1_793_962_800_000 },
]
for e in events.sortBy({ x.at }) {
  println("${Time.fromEpochMillis(e.at)} ${e.name}")
}
```

```text
2026-11-06T11:00:00Z build
2026-11-06T12:00:00Z deploy
```

## Properties and relations

`Time.epochMillis` reverses it, and `Time.fromEpochNanos` gives the same instant for a million times
the count:

```ant
for ms in [0, 1, -1, 1_793_966_400_123, -8_640_000_000_000_000] {
  let t = Time.fromEpochMillis(ms)
  assert(t.epochMillis() == ms)
}
for ms in [0, -1, 1_793_966_400_123] {
  assert(Time.fromEpochMillis(ms) == Time.fromEpochNanos(ms * 1_000_000))
}
```

A larger count is a later instant:

```ant
Time.compare(Time.fromEpochMillis(1), Time.fromEpochMillis(2)) // -1
```

## Possible issues

A count in seconds, as Unix tools print it, lands in January 1970. Multiply by 1000 first:

```ant
Time.fromEpochMillis(1_793_966_400)         // 1970-01-21T18:19:26.4Z
Time.fromEpochMillis(1_793_966_400 * 1_000) // 2026-11-06T12:00:00Z
```

A count in nanoseconds is far beyond the range, and panics. Use `Time.fromEpochNanos` for it:

```ant
Time.fromEpochMillis(1_793_966_400_000_000_000)
// panic: Time.fromEpochMillis: outside the range of time values
Time.fromEpochNanos(1_793_966_400_000_000_000) // 2026-11-06T12:00:00Z
```

The count is an `int`, and a `float64` does not convert to it on its own:

```ant
let t = Time.fromEpochMillis(1.5e12) // error: cannot use float64 where int is expected
```

Convert with `int`, which panics on a fraction; truncate first when the count may have one:

```ant
Time.fromEpochMillis(int(1.5e12))                          // 2017-07-14T02:40:00Z
Time.fromEpochMillis(int(Math.trunc(1_700_000_000_123.7))) // 2023-11-14T22:13:20.123Z
```

## See also

[Time.epochMillis](epochMillis.md), [Time.fromEpochNanos](fromEpochNanos.md),
[Time.epochNanos](epochNanos.md), [Time.now](now.md), [Time.inZone](inZone.md),
[Time.Instant](types/Instant.md)
