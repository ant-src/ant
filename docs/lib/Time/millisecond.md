<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.millisecond

```ant signature
func Time.millisecond(t: Time.Timed): int
```

Time.millisecond returns the millisecond within the second of t, 0 to 999.

<!-- end generated -->

## Details

- The result is the first three digits of the fractional second. The microseconds and nanoseconds
  below it are dropped, not rounded: `14:03:05.9996` gives 999.
- On a `Time.ZonedDateTime`, it is the millisecond its wall clock shows. A zone offset is a whole
  number of seconds, so every zone shows the same millisecond for one instant.

## Basic examples

```ant
Time.millisecond(Time.plainTime(14, 3, 5, millisecond: 123)) // 123
Time.millisecond(Time.plainTime(14, 3, 5))                   // 0
```

Call it through `.` on any value with a time:

```ant
let t = Time.parsePlainTime("14:03:05.5").unwrap()
t.millisecond() // 500
```

## Scope

A `Time.PlainDateTime` has the millisecond of its time:

```ant
let pdt = Time.parsePlainDateTime("2026-10-22T14:03:05.999").unwrap()
pdt.millisecond() // 999
```

A `Time.ZonedDateTime` has the millisecond its wall clock shows:

```ant
let z = Time.parseZoned("2026-10-22T16:03:05.250+02:00[Europe/Stockholm]").unwrap()
z.millisecond() // 250
```

Digits below a millisecond belong to [Time.microsecond](microsecond.md) and
[Time.nanosecond](nanosecond.md):

```ant
let t = Time.parsePlainTime("14:03:05.123456789").unwrap()
t.millisecond() // 123
t.microsecond() // 456
t.nanosecond()  // 789
```

A value typed `Time.Timed` may hold any of the three types with a time. Pass `Time.millisecond` by
name to read each one:

```ant
let alarm = Time.plainTime(9, millisecond: 7)
let meeting = Time.parsePlainDateTime("2026-01-06T09:00:00.08").unwrap()
let launch = Time.parseZoned("2026-01-07T12:00:00.9Z[UTC]").unwrap()
let times: Array[Time.Timed] = [alarm, meeting, launch]
times.map(Time.millisecond) // [7, 80, 900]
```

In a pipe, the value is the first argument:

```ant
"14:03:05.75" | Time.parsePlainTime() | Result.map(Time.millisecond) // Ok{ value: 750 }
```

## Applications

A time renders its fractional second to the last nonzero digit, so `14:03:05.5` and `14:03:05.12`
have different widths. Write log stamps with exactly three digits of milliseconds:

```ant
func stamp(t: Time.PlainTime): string {
  let ms = String.from(t.millisecond()).padStart(3, "0")
  "${t.with(millisecond: 0, microsecond: 0, nanosecond: 0)}.${ms}"
}

println(stamp(Time.plainTime(9, 5, 7, millisecond: 42)))
println(stamp(Time.parsePlainTime("14:03:05.5").unwrap()))
println(stamp(Time.plainTime(23, 59, 59, millisecond: 999, microsecond: 999)))
```

```text
09:05:07.042
14:03:05.500
23:59:59.999
```

Keep the events that arrived in the second half of their second:

```ant
let hits = ["14:03:05.120", "14:03:05.480", "14:03:06.010", "14:03:06.990"]
  | Array.map({ Time.parsePlainTime(x).unwrap() })
println(hits.filter({ x.millisecond() >= 500 }))
```

```text
[14:03:06.99]
```

## Properties and relations

The millisecond, microsecond, and nanosecond together are the nanoseconds within the second:

```ant
let t = Time.parsePlainTime("14:03:05.123456789").unwrap()
assert(t.millisecond() * 1_000_000 + t.microsecond() * 1000 + t.nanosecond() == 123_456_789)
```

Rounding down to the millisecond with [Time.round](round.md) keeps the millisecond and clears the
digits below it:

```ant
let t = Time.parsePlainTime("14:03:05.123456789").unwrap()
let floored = t.round(Time.Millisecond{}, mode: Time.Floor{})
assert(floored.millisecond() == t.millisecond())
assert(floored.microsecond() == 0)
assert(floored.nanosecond() == 0)
```

Setting the millisecond with [Time.with](with.md) sets what `Time.millisecond` reads:

```ant
let t = Time.plainTime(14, 3, 5, millisecond: 123)
assert(t.with(millisecond: 0).millisecond() == 0)
assert(t.with(millisecond: 999).millisecond() == 999)
```

The same instant shows the same millisecond in every zone:

```ant
let z = Time.parseZoned("2026-10-22T16:03:05.250+02:00[Europe/Stockholm]").unwrap()
let tokyo = z.withZone(Time.zone("Asia/Tokyo").unwrap())
assert(tokyo.millisecond() == z.millisecond())
```

## Possible issues

A `Time.Instant` has no wall clock, so passing one is a compile error:

```ant
let instant = Time.fromEpochMillis(1234)
println(Time.millisecond(instant)) // error: cannot use Time.Instant where Time.Timed is expected
```

Put the instant in a zone first:

```ant
let instant = Time.fromEpochMillis(1234)
instant.inZone(Time.utc).millisecond() // 234
```

`Time.millisecond` drops the digits below it, while the default mode of `Time.round` rounds half
away from zero. Rounding `14:03:05.9996` to the millisecond moves it into the next second:

```ant
let t = Time.parsePlainTime("14:03:05.9996").unwrap()
t.millisecond()                           // 999
t.round(Time.Millisecond{})               // 14:03:06
t.round(Time.Millisecond{}).millisecond() // 0
```

Before 1970, `%` on the epoch milliseconds is negative, while the millisecond is always 0 to 999:

```ant
let t = Time.fromEpochMillis(-1500).inZone(Time.utc)
t.millisecond()        // 500
t.epochMillis() % 1000 // -500
```

`Time.plainTime` panics on a millisecond of 1000 or more, while `Time.with` constrains it to 999:

```ant
let t = Time.plainTime(14, 3, 5)
t.with(millisecond: 1500).millisecond()     // 999
Time.plainTime(14, 3, 5, millisecond: 1000) // panic: Time.plainTime: millisecond out of range: 1000
```

## See also

[Time.second](second.md) · [Time.microsecond](microsecond.md) · [Time.nanosecond](nanosecond.md) ·
[Time.plainTime](plainTime.md) · [Time.with](with.md) · [Time.round](round.md) ·
[Time.epochMillis](epochMillis.md) · [Time.Timed](types/Timed.md) · [Time](index.md)
