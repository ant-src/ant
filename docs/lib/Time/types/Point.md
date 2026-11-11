<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.Point

```ant signature
type Time.Point = Time.Instant | Time.ZonedDateTime | Time.PlainDate | Time.PlainTime | Time.PlainDateTime
```

Time.Point is any of the five time types, the ones [`Time.add`](../add.md),
[`Time.until`](../until.md) and [`Time.compare`](../compare.md) serve.

<!-- end generated -->

## Details

- Its main use is as a bound: `[T: Time.Point]` lets one generic function take every time type. Each
  call fixes `T` to a single member, so all the `T` arguments share one type and a `T` result has it
  too.
- Inside such a function, `Time.add`, `Time.subtract`, `Time.until`, `Time.since` and `Time.compare`
  take `T`. The comparison operators do not.
- A value of type `Time.Point` holds any of the five types, but a function bounded by `Time.Point`
  needs one of them: narrow the value with `is` first.

## Basic examples

A function bounded by `Time.Point` takes any time type and returns that type:

```ant
func later[T: Time.Point](a: T, b: T): T {
  if Time.compare(a, b) >= 0 { a } else { b }
}

later(Time.plainDate(2026, 3, 1), Time.plainDate(2025, 12, 24)) // 2026-03-01
later(Time.plainTime(9), Time.plainTime(17, 30))                // 17:30:00
```

As a value type, `Time.Point` holds time values of different types side by side:

```ant
let points: Array[Time.Point] = [Time.fromEpochMillis(0), Time.plainDate(2026, 5, 4), Time.plainTime(9)]
points // [1970-01-01T00:00:00Z, 2026-05-04, 09:00:00]
```

## Scope

All five members pass, and the result keeps the argument's type:

```ant
func later[T: Time.Point](a: T, b: T): T {
  if Time.compare(a, b) >= 0 { a } else { b }
}

let ny = Time.zone("America/New_York").unwrap()
let morning = Time.plainDateTime(Time.plainDate(2026, 5, 4), Time.plainTime(9))
let lastNight = Time.plainDateTime(Time.plainDate(2026, 5, 3), Time.plainTime(23))
let first = Time.fromEpochMillis(0)
let second = Time.fromEpochMillis(1000)
later(Time.plainDate(2026, 1, 1), Time.plainDate(2026, 1, 2)) // 2026-01-02
later(Time.plainTime(23, 59), Time.plainTime(0, 1))           // 23:59:00
later(morning, lastNight)                                     // 2026-05-04T09:00:00
later(first, second)                                          // 1970-01-01T00:00:01Z
later(first.inZone(ny), second.inZone(ny))                    // 1969-12-31T19:00:01-05:00[America/New_York]
```

A generic record holds two values of one member type:

```ant
type Span[T: Time.Point] = { start: T, end: T }

func Span.length[T: Time.Point](s: Span[T]): Time.Duration {
  Time.until(s.start, s.end)
}

let shift = Span{ start: Time.plainTime(9), end: Time.plainTime(17, 30) }
let trip = Span{ start: Time.plainDate(2026, 7, 1), end: Time.plainDate(2026, 7, 15) }
shift.length() // PT8H30M
trip.length()  // P14D
```

Narrow a `Time.Point` value with `is` to use what only one member has:

```ant
func describe(t: Time.Point): string {
  if t is Time.Instant {
    "instant ${t.epochMillis()}"
  } else if t is Time.PlainDate {
    "date in ${t.year()}"
  } else if t is Time.PlainTime {
    "time at hour ${t.hour()}"
  } else {
    "date and time ${t}"
  }
}

let day = Time.plainDate(2026, 5, 4)
describe(Time.fromEpochMillis(0))                    // "instant 0"
describe(day)                                        // "date in 2026"
describe(Time.plainTime(9))                          // "time at hour 9"
describe(Time.plainDateTime(day, Time.plainTime(9))) // "date and time 2026-05-04T09:00:00"
```

`==` compiles on `Time.Point` values; values of different member types are never equal:

```ant
let a: Time.Point = Time.plainDate(2026, 5, 4)
let b: Time.Point = Time.plainDateTime(Time.plainDate(2026, 5, 4), Time.plainTime(0))
a == Time.plainDate(2026, 5, 4) // true
a == b                          // false
```

## Applications

Test whether a value falls in a half-open range, for any time type:

```ant
func within[T: Time.Point](t: T, start: T, end: T): bool {
  Time.compare(start, t) <= 0 && Time.compare(t, end) < 0
}

let july = Time.plainDate(2026, 7, 1)
let august = Time.plainDate(2026, 8, 1)
within(Time.plainTime(12, 30), Time.plainTime(9), Time.plainTime(17)) // true
within(Time.plainDate(2026, 7, 31), july, august)                     // true
within(august, july, august)                                          // false
```

The earliest of an array of any one time type, or `null` when it is empty:

```ant
func earliest[T: Time.Point](xs: Array[T]): T | null {
  xs.sortWith(Time.compare).get(0)
}

let none: Array[Time.PlainTime] = []
earliest([Time.plainDate(2026, 3, 1), Time.plainDate(2025, 12, 24)]) // 2025-12-24
earliest(none)                                                       // null
```

A schedule of evenly spaced values. `Time.add` takes `T` inside the body, so the steps keep the
start's type:

```ant
func steps[T: Time.Point](start: T, step: Time.Duration, count: int): Array[T] {
  let out: Array[T] = []
  var t = start
  for _ in Iter.range(0, count) {
    out.push(t)
    t = t.add(step)
  }
  out
}

for t in steps(Time.plainTime(22), Time.duration(hours: 2), 3) {
  println(t)
}
println(steps(Time.fromEpochMillis(0), Time.duration(minutes: 15), 3))
```

```text
22:00:00
00:00:00
02:00:00
[1970-01-01T00:00:00Z, 1970-01-01T00:15:00Z, 1970-01-01T00:30:00Z]
```

## Properties and relations

Every member of `Time.Wall`, `Time.Dated` and `Time.Timed` is a member of `Time.Point`, so a value
of those unions is a `Time.Point` too. `Time.Wall` is `Time.Point` without `Time.Instant`:

```ant
let d: Time.Dated = Time.plainDate(2026, 5, 4)
let t: Time.Timed = Time.plainTime(9)
let points: Array[Time.Point] = [d, t]
assert(points.length() == 2)
```

Inside a body bounded by `Time.Point`, `Time.since` is `Time.until` negated, for every member:

```ant
func check[T: Time.Point](a: T, b: T) {
  assert(Time.since(a, b) == Time.until(a, b).negate())
}

check(Time.plainDate(2026, 3, 1), Time.plainDate(2026, 4, 1))
check(Time.plainTime(8), Time.plainTime(17, 45))
check(Time.fromEpochMillis(0), Time.fromEpochMillis(90_000))
```

## Possible issues

A value typed `Time.Point` cannot be passed to `Time.add` and the other functions bounded by it: `T`
would be the whole union, not one member:

```ant
func nextDay(t: Time.Point): Time.Point {
  Time.add(t, Time.duration(days: 1))
  // error: T would be Time.Point, which is not a single member of its bound Time.Point; narrow first
}
```

Narrow it with `is`, one member per branch. An `Instant` has no days, so it takes 24 hours; a
`PlainTime` ignores days, so it comes back unchanged:

```ant
func nextDay(t: Time.Point): Time.Point {
  let day = Time.duration(days: 1)
  if t is Time.Instant {
    t.add(Time.duration(hours: 24))
  } else if t is Time.ZonedDateTime {
    t.add(day)
  } else if t is Time.PlainDate {
    t.add(day)
  } else if t is Time.PlainTime {
    t
  } else {
    t.add(day)
  }
}

nextDay(Time.plainDate(2026, 2, 28)) // 2026-03-01
nextDay(Time.fromEpochMillis(0))     // 1970-01-02T00:00:00Z
```

All the arguments typed `T` share one member type, so a date and a date-time do not mix:

```ant
let day = Time.plainDate(2026, 5, 4)
let gap = Time.until(day, Time.plainDateTime(day, Time.plainTime(9)))
// error: cannot use Time.PlainDateTime where Time.PlainDate is expected
```

Convert one of them, here the date to its midnight:

```ant
let day = Time.plainDate(2026, 5, 4)
let midnight = Time.plainDateTime(day, Time.plainTime(0))
Time.until(midnight, Time.plainDateTime(day, Time.plainTime(9))) // PT9H
```

`<` and the other comparison operators do not apply to `T`, even with this bound. Use
`Time.compare`:

```ant
func later[T: Time.Point](a: T, b: T): T {
  if a > b { a } else { b }
  // error: > applies to a single numeric kind, string, or a named primitive type over one, not to T; narrow first
}
```

`is` cannot test `Time.Point` itself, since an alias has no run-time identity:

```ant
func isTime(x: Time.Point | string): bool {
  x is Time.Point // error: Time.Point is an alias and has no run-time identity; test its members instead
}
```

A narrowed value is a plain member type, no longer a `T`, so it cannot be returned as a `T`:

```ant
func shift[T: Time.Point](t: T): T {
  if t is Time.Instant { t.add(Time.duration(hours: 1)) } else { t }
  // error: cannot use Time.Instant where T is expected
}
```

A `Time.Duration` is not a member. Compare durations with `Time.compareDuration`:

```ant
let d = Time.compare(Time.duration(hours: 1), Time.duration(hours: 2))
// error: T would be Time.Duration, which is not a member of its bound Time.Point
```

```ant
Time.compareDuration(Time.duration(hours: 1), Time.duration(hours: 2)) // -1
```

## See also

- [Time.Wall](Wall.md), [Time.Dated](Dated.md), [Time.Timed](Timed.md): the narrower time unions.
- [Time.Instant](Instant.md), [Time.ZonedDateTime](ZonedDateTime.md),
  [Time.PlainDate](PlainDate.md), [Time.PlainTime](PlainTime.md),
  [Time.PlainDateTime](PlainDateTime.md): the members.
- [Time.add](../add.md), [Time.subtract](../subtract.md), [Time.until](../until.md),
  [Time.since](../since.md), [Time.compare](../compare.md)
- [is](../../../language/operators/is.md)
