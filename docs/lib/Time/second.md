<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.second

```ant signature
func Time.second(t: Time.Timed): int
```

Time.second returns the second of t, 0 to 59.

<!-- end generated -->

## Details

- The second is the whole second the wall clock shows. The fraction below it is dropped, not
  rounded: `10:00:59.7` gives 59.
- A second is never 60: parsing the leap second `23:59:60` is `Err`.
- On a `Time.ZonedDateTime` it is the second the zone's wall clock shows.

## Basic examples

```ant
Time.second(Time.plainTime(14, 3, 5)) // 5
Time.second(Time.plainTime(14, 3))    // 0
Time.second(Time.plainTime(0, 0, 59)) // 59
```

Call it through `.` on any value with a time:

```ant
let start = Time.parsePlainTime("09:30:15").unwrap()
start.second() // 15
```

## Scope

A `Time.PlainDateTime` has the second of its time:

```ant
let saved = Time.parsePlainDateTime("2026-11-11T18:04:42").unwrap()
saved.second() // 42
```

A `Time.ZonedDateTime` has the second its zone's wall clock shows:

```ant
let landing = Time.parseZoned("2026-11-11T16:20:08+01:00[Europe/Stockholm]").unwrap()
landing.second() // 8
```

A fractional second does not count. The digits after the point belong to
[Time.millisecond](millisecond.md) and the fields below it:

```ant
let t = Time.parsePlainTime("12:00:07.999").unwrap()
t.second()      // 7
t.millisecond() // 999
```

A value typed `Time.Timed` may hold any of the three types with a time. Pass `Time.second` by name
to read each one:

```ant
let lunch = Time.parsePlainDateTime("2026-11-11T12:15:30").unwrap()
let launch = Time.parseZoned("2026-11-11T18:45:01Z[UTC]").unwrap()
let times: Array[Time.Timed] = [Time.plainTime(8, 5, 20), lunch, launch]
times.map(Time.second) // [20, 30, 1]
```

In a pipe, the value is the first argument:

```ant
"07:42:10" | Time.parsePlainTime() | Result.map(Time.second) // Ok{ value: 10 }
```

## Named arguments

The parameter is named `t`:

```ant
Time.second(t: Time.plainTime(13, 37, 21)) // 21
```

## Applications

Read the current second from the clock. `Time.now()` is an instant, so put it in a zone first:

```ant
let s = Time.nowIn(Time.utc).second()
assert(s >= 0 && s <= 59)
```

Write a lap time as `M:SS`, padding the second to two digits:

```ant
func lap(t: Time.PlainTime): string {
  let ss = String.from(t.second()).padStart(2, "0")
  "${t.minute()}:${ss}"
}

lap(Time.plainTime(0, 4, 7))   // "4:07"
lap(Time.plainTime(0, 12, 45)) // "12:45"
```

Count the seconds since midnight:

```ant
func secondsOfDay(t: Time.PlainTime): int {
  t.hour() * 3600 + t.minute() * 60 + t.second()
}

secondsOfDay(Time.plainTime(0, 0))       // 0
secondsOfDay(Time.plainTime(1, 2, 3))    // 3723
secondsOfDay(Time.plainTime(23, 59, 59)) // 86399
```

Keep the readings a sensor took on a whole minute:

```ant
let readings = ["10:00:00", "10:00:30", "10:01:00", "10:01:15", "10:02:00"]
  | Array.map({ Time.parsePlainTime(x).unwrap() })
println(readings.filter({ x.second() == 0 }))
```

```text
[10:00:00, 10:01:00, 10:02:00]
```

## Properties and relations

Setting the second with [Time.with](with.md) sets what `Time.second` reads:

```ant
let t = Time.plainTime(10, 15)
for s in Array.range(0, 60) {
  let set = t.with(second: s)
  assert(set.second() == s)
}
```

Adding a second carries into the minute at 59, and adding a minute leaves the second as it was:

```ant
let t = Time.plainTime(9, 30, 59)
assert(t.add(1s).second() == 0)
assert(t.add(1s).minute() == 31)
assert(t.add(1m).second() == 59)
```

Rounding down to the second with [Time.round](round.md) keeps the second and clears the fraction:

```ant
let t = Time.parsePlainTime("14:03:05.875").unwrap()
let floored = t.round(Time.Second{}, mode: Time.Floor{})
assert(floored.second() == t.second())
assert(floored.millisecond() == 0)
```

## Possible issues

A `Time.Instant` has no wall clock, so passing one is a compile error:

```ant
println(Time.second(Time.now())) // error: cannot use Time.Instant where Time.Timed is expected
```

Put the instant in a zone first:

```ant
let instant = Time.fromEpochMillis(83_000)
instant.inZone(Time.utc).second() // 23
```

A `Time.PlainDate` has no time:

```ant
println(Time.plainDate(2026, 11, 11).second())
// error: Time.second takes Time.Timed first, so a Time.PlainDate does not reach it
```

Rounding with the default mode of `Time.round` can carry into the next minute:

```ant
let t = Time.parsePlainTime("10:00:59.7").unwrap()
t.round(Time.Second{})          // 10:01:00
t.round(Time.Second{}).second() // 0
```

There is no second 60. Parsing a leap second is `Err`, and `Time.with` constrains it to 59:

```ant
Time.parsePlainTime("23:59:60").isOk()               // false
Time.plainTime(23, 59, 30).with(second: 60).second() // 59
```

`Time.plainTime` panics on it:

```ant
Time.plainTime(23, 59, 60) // panic: Time.plainTime: second out of range: 60
```

Before 1970, `%` on the epoch seconds is negative, while the second is always 0 to 59:

```ant
let t = Time.fromEpochMillis(-1500)
t.inZone(Time.utc).second() // 58
t.epochMillis() / 1000 % 60 // -1
```

A historical offset can have seconds, so the same instant shows a different second in another zone.
Paris kept its local mean time, 9 minutes 21 seconds ahead of UTC, until 1911:

```ant
let instant = Time.parseInstant("1900-01-01T12:00:00Z").unwrap()
let paris = instant.inZone(Time.zone("Europe/Paris").unwrap())
instant.inZone(Time.utc).second() // 0
paris.second()                    // 21
paris.offset()                    // "+00:09:21"
```

`Time.second` reads the clock, not a span. A duration literal keeps its seconds as written, and
[Time.part](part.md) reads them:

```ant
90s.part(Time.Second{}) // 90
```

## See also

[Time.minute](minute.md) · [Time.millisecond](millisecond.md) · [Time.with](with.md) ·
[Time.round](round.md) · [Time.plainTime](plainTime.md) · [Time.inZone](inZone.md) ·
[Time.part](part.md) · [Time.Timed](types/Timed.md) · [Time](index.md)
