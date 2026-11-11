<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.nanosecond

```ant signature
func Time.nanosecond(t: Time.Timed): int
```

Time.nanosecond returns the nanosecond within the microsecond of t, 0 to 999.

<!-- end generated -->

## Details

- The value is the nanosecond within the microsecond, 0 to 999, not within the second. The
  nanoseconds of the second are
  `t.millisecond() * 1_000_000 + t.microsecond() * 1000 + t.nanosecond()`.
- The nanosecond is the smallest field: a fraction with more than nine digits does not parse.
- On a `Time.ZonedDateTime`, it is the nanosecond its zone's wall clock shows.
- A `Time.Instant` has no wall clock, so it is a compile error. Put it in a zone with
  [Time.inZone](inZone.md) first.

## Basic examples

```ant
let t = Time.plainTime(14, 3, 5, millisecond: 123, microsecond: 456, nanosecond: 789)
Time.nanosecond(t)                        // 789
Time.nanosecond(Time.plainTime(14, 3, 5)) // 0
```

Call it through `.` on any value with a time:

```ant
let t = Time.parsePlainTime("14:03:05.123456789").unwrap()
t.nanosecond() // 789
```

## Scope

The nanosecond is the last three of the nine digits of the fraction. A shorter fraction is padded
with zeros on the right, so seven digits leave a nanosecond of 700:

```ant
Time.parsePlainTime("14:03:05.000000001").unwrap().nanosecond() // 1
Time.parsePlainTime("14:03:05.1234567").unwrap().nanosecond()   // 700
Time.parsePlainTime("14:03:05.123456").unwrap().nanosecond()    // 0
```

A `Time.ZonedDateTime` has the nanosecond its wall clock shows. An instant one nanosecond before the
epoch shows 999:

```ant
let before = Time.fromEpochNanos(-1).inZone(Time.utc)
before              // 1969-12-31T23:59:59.999999999+00:00[UTC]
before.nanosecond() // 999
```

A value typed `Time.Timed` may hold any of the three types with a time. Pass `Time.nanosecond` by
name to read each one:

```ant
let time = Time.plainTime(9, nanosecond: 1)
let local = Time.parsePlainDateTime("2026-10-22T09:00:00.000000002").unwrap()
let zoned = Time.parseZoned("2026-10-22T09:00:00.000000003+02:00[Europe/Stockholm]").unwrap()
let times: Array[Time.Timed] = [time, local, zoned]
times.map(Time.nanosecond) // [1, 2, 3]
```

In a pipe, the value is the first argument:

```ant
"08:15:00.000000777" | Time.parsePlainTime() | Result.map(Time.nanosecond) // Ok{ value: 777 }
```

## Named arguments

```ant
Time.nanosecond(t: Time.plainTime(nanosecond: 5)) // 5
```

## Applications

Write the fraction of a second as all nine digits:

```ant
func secondsWithNanos(t: Time.Timed): string {
  let ns = t.millisecond() * 1_000_000 + t.microsecond() * 1000 + t.nanosecond()
  let frac = String.from(ns).padStart(9, "0")
  "${t.second()}.${frac}"
}

for text in ["14:03:05.000000042", "14:03:07.25"] {
  let t = Time.parsePlainTime(text).unwrap()
  println(secondsWithNanos(t))
}
```

```text
5.000000042
7.250000000
```

Check whether a log keeps nanosecond precision or only microseconds:

```ant
let stamps = ["10:00:00.000001000", "10:00:00.500000", "10:00:01.123456789"]
let precise = stamps
  | Array.map({ Time.parsePlainTime(x).unwrap() })
  | Array.some({ x.nanosecond() != 0 })
println(precise)
```

```text
true
```

## Properties and relations

The nanosecond is between 0 and 999, and the value a time is made with is the value it reads back:

```ant
for ns in [0, 1, 500, 999] {
  let t = Time.plainTime(12, nanosecond: ns)
  assert(t.nanosecond() == ns)
  assert(t.nanosecond() >= 0 && t.nanosecond() <= 999)
}
```

For a zoned value, the nanosecond matches the last three digits of [Time.epochNanos](epochNanos.md)
when the instant is after the epoch, since every zone's offset is a whole number of seconds:

```ant
let z = Time.fromEpochNanos(1_761_141_600_123_456_789).inZone(Time.zone("Asia/Tokyo").unwrap())
assert(z.nanosecond() == z.toInstant().epochNanos() % 1000)
assert(z.nanosecond() == 789)
```

Adding one nanosecond to 999 carries into the microsecond:

```ant
let t = Time.plainTime(9, nanosecond: 999)
let u = t.add(Time.duration(nanoseconds: 1))
assert(u.nanosecond() == 0)
assert(u.microsecond() == 1)
```

Rounding to the microsecond with [Time.round](round.md) clears the nanosecond:

```ant
let t = Time.parsePlainTime("12:00:00.000000750").unwrap()
assert(t.round(Time.Microsecond{}).nanosecond() == 0)
assert(t.round(Time.Microsecond{}).microsecond() == 1)
```

## Possible issues

A `Time.Instant` has no wall clock, so passing one is a compile error:

```ant
let epoch = Time.fromEpochNanos(5)
println(Time.nanosecond(epoch)) // error: cannot use Time.Instant where Time.Timed is expected
```

The nanosecond is within the microsecond, so 1.5 microseconds into a second reads as 500, not 1500:

```ant
let t = Time.parsePlainTime("10:00:00.0000015").unwrap()
t.nanosecond()                          // 500
t.microsecond() * 1000 + t.nanosecond() // 1500
```

A `Time.Duration` has no wall clock either. Read its nanoseconds with [Time.part](part.md). The
field is kept as given, so it is not limited to 0 to 999:

```ant
Time.part(Time.duration(nanoseconds: 1500), Time.Nanosecond{}) // 1500
```

A time holds nothing below the nanosecond, so a fraction of ten digits does not parse:

```ant
Time.parsePlainTime("14:03:05.123456789").isOk()  // true
Time.parsePlainTime("14:03:05.1234567891").isOk() // false
```

`Time.plainTime` panics on a nanosecond of 1000 or more, while [Time.with](with.md) constrains it to
0 to 999:

```ant
let t = Time.plainTime(10)
t.with(nanosecond: 1000).nanosecond() // 999
t.with(nanosecond: -5).nanosecond()   // 0
Time.plainTime(nanosecond: 1000)      // panic: Time.plainTime: nanosecond out of range: 1000
```

## See also

[Time.microsecond](microsecond.md) · [Time.millisecond](millisecond.md) · [Time.second](second.md) ·
[Time.epochNanos](epochNanos.md) · [Time.part](part.md) · [Time.with](with.md) ·
[Time.round](round.md) · [Time.plainTime](plainTime.md) · [Time](index.md)
