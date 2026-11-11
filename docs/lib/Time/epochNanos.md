<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.epochNanos

```ant signature
func Time.epochNanos(t: Time.Instant | Time.ZonedDateTime): int
```

Time.epochNanos returns the nanoseconds from the Unix epoch to t. It panics if t is outside
1677-09-21 to 2262-04-11, beyond which the count does not fit in an int.

<!-- end generated -->

## Details

- The count is exact: no rounding, and every nanosecond of `t` is kept.
- A time before the epoch gives a negative count.
- A `ZonedDateTime` gives the count of its instant; the zone does not change it.

## Basic examples

```ant
let t = Time.parseInstant("2026-10-22T14:03:05.123456789Z") | Result.unwrap()
Time.epochNanos(t)                       // 1792677785123456789
Time.epochNanos(Time.fromEpochNanos(0))  // 0
Time.epochNanos(Time.fromEpochMillis(1)) // 1000000
Time.epochNanos(Time.now()) > 0          // true
```

## Scope

A time before the epoch counts down from zero:

```ant
let before = Time.parseInstant("1969-12-31T23:59:59Z") | Result.unwrap()
before.epochNanos() // -1000000000
```

A `ZonedDateTime` gives the same count as the instant it shows, in any zone:

```ant
let utc = Time.parseZoned("2026-10-22T14:03:05Z[UTC]") | Result.unwrap()
let stockholm = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]") | Result.unwrap()
utc.epochNanos()       // 1792677785000000000
stockholm.epochNanos() // 1792677785000000000
```

The last nanosecond that fits is `Math.maxInt`:

```ant
Time.fromEpochNanos(Math.maxInt)              // 2262-04-11T23:47:16.854775807Z
Time.fromEpochNanos(Math.maxInt).epochNanos() // 9223372036854775807
```

Pass `Time.epochNanos` as a function value:

```ant
[Time.fromEpochNanos(5), Time.fromEpochNanos(-5)].map(Time.epochNanos) // [5, -5]
```

## Named arguments

```ant
Time.epochNanos(t: Time.fromEpochNanos(42)) // 42
```

## Applications

Split an instant into whole seconds and the nanoseconds left over, the shape many wire formats use.
Those formats keep the nanoseconds from 0 to 999999999, so the split rounds the seconds down rather
than truncating toward zero, and a time before the epoch still gets a nanosecond count that is not
negative:

```ant
type Timestamp = { seconds: int, nanos: int }

func timestamp(t: Time.Instant): Timestamp {
  let ns = t.epochNanos()
  let nanos = ((ns % 1000000000) + 1000000000) % 1000000000
  Timestamp{ seconds: (ns - nanos) / 1000000000, nanos: nanos }
}

let at = Time.parseInstant("2026-10-22T14:03:05.25Z") | Result.unwrap()
let before = Time.parseInstant("1969-12-31T23:59:59.5Z") | Result.unwrap()
timestamp(at)     // Timestamp{ seconds: 1792677785, nanos: 250000000 }
timestamp(before) // Timestamp{ seconds: -1, nanos: 500000000 }
```

The gap between two instants as a plain `int` of nanoseconds:

```ant
let start = Time.parseInstant("2026-10-22T14:00:00Z") | Result.unwrap()
let end = Time.parseInstant("2026-10-22T14:00:00.000000250Z") | Result.unwrap()
end.epochNanos() - start.epochNanos() // 250
```

## Properties and relations

`Time.fromEpochNanos` undoes it, a zoned time counts as its instant, and the counts order instants
as `Time.compare` does:

```ant
let a = Time.parseInstant("2026-10-22T14:03:05.123456789Z") | Result.unwrap()
let b = Time.parseInstant("1969-07-20T20:17:40Z") | Result.unwrap()
for t in [a, b] {
  assert(Time.fromEpochNanos(t.epochNanos()) == t)
  assert(t.inZone(Time.utc).epochNanos() == t.epochNanos())
}
assert((a.epochNanos() < b.epochNanos()) == (Time.compare(a, b) < 0))
```

For a time at or after the epoch, dividing by a million gives `Time.epochMillis`:

```ant
let t = Time.parseInstant("2026-10-22T14:03:05.123456789Z") | Result.unwrap()
assert(t.epochNanos() / 1000000 == t.epochMillis())
```

## Possible issues

An `int` of nanoseconds reaches only from 1677 to 2262, while an instant reaches much further. Use
`Time.epochMillis` for a time outside that span:

```ant
let landing = Time.parseInstant("1600-01-01T00:00:00Z") | Result.unwrap()
landing.epochMillis() // -11676096000000
landing.epochNanos()
// panic: Time.epochNanos: the nanoseconds do not fit an int; the instant is outside 1677 to 2262
```

Integer division truncates toward zero, so dividing a negative count gives a different answer than
`Time.epochMillis`, which rounds down:

```ant
let t = Time.fromEpochNanos(-1500000)
t.epochNanos() / 1000000 // -1
t.epochMillis()          // -2
```

The plain types have no zone, so they are not on the time line and have no epoch count. Give them a
zone first with `Time.toZoned`:

```ant
let meeting = Time.parsePlainDateTime("2026-10-22T14:00:00") | Result.unwrap()
let n = Time.epochNanos(meeting)
// error: cannot use Time.PlainDateTime where Time.Instant | Time.ZonedDateTime is expected
```

## See also

[Time.fromEpochNanos](fromEpochNanos.md), [Time.epochMillis](epochMillis.md),
[Time.fromEpochMillis](fromEpochMillis.md), [Time.toInstant](toInstant.md),
[Time.inZone](inZone.md), [Time.Instant](types/Instant.md)
