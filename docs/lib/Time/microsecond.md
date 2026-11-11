<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.microsecond

```ant signature
func Time.microsecond(t: Time.Timed): int
```

Time.microsecond returns the microsecond within the millisecond of t, 0 to 999.

<!-- end generated -->

## Details

- The value is the microsecond within the millisecond, 0 to 999, not within the second. The
  microseconds of the second are `t.millisecond() * 1000 + t.microsecond()`.
- On a `Time.ZonedDateTime`, it is the microsecond its zone's wall clock shows.

## Basic examples

```ant
Time.microsecond(Time.plainTime(14, 3, 5, millisecond: 123, microsecond: 456)) // 456
Time.microsecond(Time.plainTime(14, 3, 5))                                     // 0
```

Call it through `.` on any value with a time:

```ant
let t = Time.parsePlainTime("14:03:05.123456789").unwrap()
t.microsecond() // 456
```

## Scope

The digits of a fraction count in threes: milliseconds, then microseconds, then nanoseconds. So `.5`
is 500 milliseconds and no microseconds, and `.000042` is 42 microseconds:

```ant
Time.parsePlainTime("14:03:05.5").unwrap().microsecond()      // 0
Time.parsePlainTime("14:03:05.000042").unwrap().microsecond() // 42
```

A `Time.PlainDateTime` has the microsecond of its time:

```ant
let dt = Time.parsePlainDateTime("2026-10-22T14:03:05.000250").unwrap()
dt.microsecond() // 250
```

A `Time.ZonedDateTime` has the microsecond its wall clock shows. An instant one microsecond before
the epoch shows 999:

```ant
let before = Time.fromEpochNanos(-1000).inZone(Time.utc)
before               // 1969-12-31T23:59:59.999999+00:00[UTC]
before.microsecond() // 999
```

A value typed `Time.Timed` may hold any of the three types with a time. Pass `Time.microsecond` by
name to read each one:

```ant
let time = Time.plainTime(9, microsecond: 1)
let local = Time.parsePlainDateTime("2026-10-22T09:00:00.000002").unwrap()
let zoned = Time.parseZoned("2026-10-22T09:00:00.000003+02:00[Europe/Stockholm]").unwrap()
let times: Array[Time.Timed] = [time, local, zoned]
times.map(Time.microsecond) // [1, 2, 3]
```

In a pipe, the value is the first argument:

```ant
"08:15:00.000777" | Time.parsePlainTime() | Result.map(Time.microsecond) // Ok{ value: 777 }
```

## Named arguments

```ant
Time.microsecond(t: Time.plainTime(microsecond: 5)) // 5
```

## Applications

Write the fraction of a second as six digits, as many log formats do:

```ant
func micros(t: Time.Timed): string {
  let us = t.millisecond() * 1000 + t.microsecond()
  let frac = String.from(us).padStart(6, "0")
  "${t.second()}.${frac}"
}

for s in ["14:03:05.000042", "14:03:07.25"] {
  let t = Time.parsePlainTime(s).unwrap()
  println(micros(t))
}
```

```text
5.000042
7.250000
```

Find the timestamps that carry precision below the millisecond:

```ant
let stamps = ["10:00:00.5", "10:00:00.000300", "10:00:01", "10:00:01.123000001"]
let fine = stamps
  | Array.map({ Time.parsePlainTime(x).unwrap() })
  | Array.filter({ x.microsecond() != 0 || x.nanosecond() != 0 })
println(fine)
```

```text
[10:00:00.0003, 10:00:01.123000001]
```

## Properties and relations

The value a time is made with is the value it reads back:

```ant
for us in [0, 1, 500, 999] {
  assert(Time.plainTime(12, microsecond: us).microsecond() == us)
}
```

Whatever the fraction, the microsecond is between 0 and 999:

```ant
for s in ["12:00:00.000999", "12:00:00.001", "12:00:00.999999999"] {
  let us = Time.parsePlainTime(s).unwrap().microsecond()
  assert(us >= 0 && us <= 999)
}
```

Setting the microsecond with [Time.with](with.md) leaves the millisecond and nanosecond alone:

```ant
let t = Time.plainTime(14, millisecond: 123, microsecond: 456, nanosecond: 789)
let u = t.with(microsecond: 7)
assert(u.microsecond() == 7)
assert(u.millisecond() == t.millisecond())
assert(u.nanosecond() == t.nanosecond())
```

Rounding to the millisecond with [Time.round](round.md) clears the microsecond:

```ant
let t = Time.plainTime(9, 30, millisecond: 1, microsecond: 999, nanosecond: 500)
let r = t.round(Time.Millisecond{})
assert(r.microsecond() == 0)
assert(r.millisecond() == 2)
```

## Possible issues

A `Time.Instant` has no wall clock, so passing one is a compile error:

```ant
let epoch = Time.fromEpochNanos(0)
println(Time.microsecond(epoch)) // error: cannot use Time.Instant where Time.Timed is expected
```

Put it in a zone with [Time.inZone](inZone.md) first.

The microsecond is within the millisecond, so 1.5 milliseconds into a second reads as 500, not 1500:

```ant
let t = Time.parsePlainTime("10:00:00.0015").unwrap()
t.microsecond()                          // 500
t.millisecond() * 1000 + t.microsecond() // 1500
```

`Time.plainTime` panics on a microsecond of 1000 or more, while [Time.with](with.md) constrains it
to 999:

```ant
let t = Time.plainTime(10)
t.with(microsecond: 1000).microsecond() // 999
Time.plainTime(microsecond: 1000)       // panic: Time.plainTime: microsecond out of range: 1000
```

## See also

[Time.millisecond](millisecond.md) · [Time.nanosecond](nanosecond.md) · [Time.second](second.md) ·
[Time.with](with.md) · [Time.round](round.md) · [Time.plainTime](plainTime.md) · [Time](index.md)
