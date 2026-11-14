<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.Duration

```ant signature
type Time.Duration
```

Time.Duration is an amount of time in ten fields, years down to nanoseconds, all of one sign and
kept as given. 90 minutes stays 90 minutes until [`Time.roundDuration`](../roundDuration.md)
balances it, and two durations are == only when all ten fields are, so PT90M and PT1H30M are not. It
renders in ISO 8601 form, as P1Y2M3DT4H5M6.5S, -PT1H or PT0S.

<!-- end generated -->

## Details

- A constant span of hours and smaller units is a
  [duration literal](../../../language/literals/durations.md), written as in Go: `500ms`, `2h30m`,
  `1.5s`, `90m`. The units are `h`, `m`, `s`, `ms`, `us` and `ns`, so `m` is minutes; they come in
  descending order, each at most once, and only the last component may have a fraction. A literal
  has type Time.Duration and needs no import. There are no literals for days, weeks, months or
  years, which are `Time.duration(days: 1)` and the like, and a span computed from a number is
  `Time.duration(seconds: n)`. A literal keeps its fields as written, so `90m` is PT90M.
- Durations take no arithmetic or ordering operators, so `-5s`, `5s + 1s` and `90m < 2h` are compile
  errors. Negate with [`Time.negate`](../negate.md), add with
  [`Time.addDuration`](../addDuration.md), and order with
  [`Time.compareDuration`](../compareDuration.md).
- Years, months and weeks have no fixed length. [`Time.sleep`](../sleep.md),
  [`Time.ticker`](../ticker.md), [`Time.addDuration`](../addDuration.md) and `Context.new` panic on
  them, and [`Time.compareDuration`](../compareDuration.md), [`Time.total`](../total.md) and
  [`Time.roundDuration`](../roundDuration.md) need a `relativeTo` date for them.
- Days count as 24 hours in a wait or a sum, but [`Time.add`](../add.md) on a zoned time adds
  calendar days, and a [`Time.Instant`](Instant.md) takes no days at all.

## Basic examples

Write a constant duration as a literal:

```ant
1h30m // PT1H30M
500ms // PT0.5S
```

Build a calendar or computed duration with [`Time.duration`](../duration.md):

```ant
Time.duration(days: 3) // P3D
```

```ant
let n = 90
Time.duration(seconds: n) // PT90S
```

Parse one from its ISO 8601 text or Go's notation with [`Time.parseDuration`](../parseDuration.md):

```ant
Time.parseDuration("PT1H30M") // Ok{ value: PT1H30M }
Time.parseDuration("1h30m")   // Ok{ value: PT1H30M }
```

Measure the time between two points with [`Time.until`](../until.md):

```ant
Time.until(Time.plainTime(9, 0), Time.plainTime(17, 45)) // PT8H45M
```

## Scope

A duration renders as ISO 8601 text: the date part after `P`, the time part after `T`, the subsecond
fields folded into the seconds, and `PT0S` when it is zero:

```ant
Time.duration(years: 1, months: 2, days: 3)                     // P1Y2M3D
Time.duration(days: 3, hours: 4, seconds: 6, milliseconds: 500) // P3DT4H6.5S
Time.duration(weeks: 2)                                         // P2W
1500ms                                                          // PT1.5S
0s                                                              // PT0S
```

A negative duration has a sign in front:

```ant
Time.duration(hours: -2) // -PT2H
1h.negate()              // -PT1H
Time.negate(5s)          // -PT5S
Time.negate(15m).sign()  // -1
```

Fields are not balanced. Ninety minutes stays ninety minutes, and two weeks has no days, until
[`Time.roundDuration`](../roundDuration.md) balances it:

```ant
let d = 90m
d                                         // PT90M
d.part(Time.Hour{})                       // 0
d.roundDuration(largestUnit: Time.Hour{}) // PT1H30M
Time.duration(weeks: 2).part(Time.Day{})  // 0
```

String interpolation, `String.from` and `Json.stringify` all write the same text:

```ant
let d = 5m
"retry in ${d}"     // "retry in PT5M"
String.from(d)      // "PT5M"
Json.stringify([d]) // "[\"PT5M\"]"
```

A duration is a value like any other: a record field, a field default, an optional parameter, a map
key:

```ant
type Job = { name: string, timeout: Time.Duration | null }
Job{ name: "build", timeout: 10m } // Job{ name: "build", timeout: PT10M }
Job{ name: "lint", timeout: null } // Job{ name: "lint", timeout: null }
type Task = { name: string, timeout: Time.Duration = 10m }
Task{ name: "build" } // Task{ name: "build", timeout: PT10M }
let names = Map{ 1h: "hour", 60m: "sixty minutes" }
names.size() // 2
```

```ant
func fetch(url: string, timeout: Time.Duration | null = null): string {
  if timeout == null { "${url} without a timeout" } else { "${url} within ${timeout}" }
}

println(fetch("https://example.com"))
println(fetch("https://example.com", timeout: 2s))
```

```text
https://example.com without a timeout
https://example.com within PT2S
```

[`Time.total`](../total.md) gives the whole length in one unit as a `float64`:

```ant
1h30m.total(Time.Minute{}) // 90.0
90m.total(Time.Hour{})     // 1.5
```

## Applications

Move a time by a duration with [`Time.add`](../add.md):

```ant
Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(9, 0)).add(1h30m) // 2026-10-22T10:30:00
```

Show the time between two instants as hours, minutes and seconds. Balance it to hours and round to
the second first, then read each field:

```ant
func label(d: Time.Duration): string {
  let b = d.roundDuration(largestUnit: Time.Hour{}, smallestUnit: Time.Second{})
  "${b.part(Time.Hour{})}h ${b.part(Time.Minute{})}m ${b.part(Time.Second{})}s"
}

let start = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
let end = Time.parseInstant("2026-10-22T16:20:00.75Z").unwrap()
println(label(start.until(end)))
```

```text
2h 16m 56s
```

Give up waiting after a timeout. Every wait in the library takes a duration:

```ant
let ctx = Context.new(timeout: 20ms)
Time.sleep(5s, ctx: ctx)
println("gave up waiting")
```

```text
gave up waiting
```

Read a timeout from configuration, in either notation, with a default for text that is not a
duration:

```ant
func timeoutFrom(text: string): Time.Duration {
  Time.parseDuration(text).unwrapOr(30s)
}

println(timeoutFrom("500ms"))
println(timeoutFrom("PT2M"))
println(timeoutFrom("soon"))
```

```text
PT0.5S
PT2M
PT30S
```

Sort durations by length with [`Time.compareDuration`](../compareDuration.md):

```ant
let laps = [90m, 1h, 1h15m]
laps.sortWith(Time.compareDuration) // [PT1H, PT1H15M, PT90M]
```

## Properties and relations

`==` compares fields, and [`Time.compareDuration`](../compareDuration.md) compares lengths:

```ant
let a = 90m
let b = 1h30m
assert(a != b)
assert(Time.compareDuration(a, b) == 0)
assert(a.roundDuration(largestUnit: Time.Hour{}) == b)
```

A fraction in a literal spreads into the smaller fields, so `1.5h` has the fields of `1h30m`:

```ant
assert(1.5h == 1h30m)
assert(String.from(1.5h) == "PT1H30M")
```

The text of a duration parses back to an equal duration:

```ant
let d = Time.duration(days: 1, hours: 2, seconds: 3, milliseconds: 40)
assert(Time.parseDuration("${d}").unwrap() == d)
```

Negating twice gives the duration back:

```ant
let d = 3h5m
assert(d.negate().negate() == d)
assert(d.negate().abs() == d)
```

## Possible issues

`==` is not a test of length. Two durations of one length are `==` only with the same fields, and
`ant check` rejects `==` between two such literals:

```ant
let same = 90m == 1h30m
// error: 90m == 1h30m compares the fields as written, and == is false: compare the lengths with Time.compareDuration(90m, 1h30m)
```

```ant
Time.compareDuration(90m, 1h30m) // 0
```

`<` does not compare durations. Use [`Time.compareDuration`](../compareDuration.md):

```ant
let short = 90m < 2h
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.Duration
```

Durations take no operators. Negate with [`Time.negate`](../negate.md), as in `Time.negate(5s)`:

```ant
let d = -5s // error: unary - needs a number, got Time.Duration
```

Add with [`Time.addDuration`](../addDuration.md), as in `Time.addDuration(5s, 1s)`:

```ant
let d = 5s + 1s
// error: + applies to a single numeric kind, string, or a named primitive type over one, not to Time.Duration
```

There are no literals for days, weeks, months or years:

```ant
let d = 1d
// error: there are no duration literals for days, weeks, months or years: write Time.duration(days: 1) and the like
```

A wait needs a fixed length, so a duration in months panics:

```ant
Time.sleep(Time.duration(months: 1))
// panic: Time.sleep needs a duration without years, months or weeks, which have no fixed length
```

All fields of a duration have one sign, so an hour less 30 minutes is not `hours: 1, minutes: -30`.
Write `30m`, or negate the duration to take away and pass it to
[`Time.addDuration`](../addDuration.md):

```ant
Time.duration(hours: 1, minutes: -30)
// panic: Time.duration: a duration's fields must all have one sign
```

```ant
Time.addDuration(1h, Time.negate(30m)) // PT30M
```

Text in neither ISO 8601 nor Go's notation is an `Err`, and so is a bare number other than `0`,
which has no unit:

```ant
Time.parseDuration("1 hour") // Err{ error: Error{ message: "not a duration: \"1 hour\"" } }
Time.parseDuration("5")      // Err{ error: Error{ message: "not a duration: \"5\"" } }
```

## See also

[Time.duration](../duration.md), [Time.parseDuration](../parseDuration.md),
[Time.until](../until.md), [Time.since](../since.md), [Time.add](../add.md),
[Time.addDuration](../addDuration.md), [Time.negate](../negate.md),
[Time.compareDuration](../compareDuration.md), [Time.roundDuration](../roundDuration.md),
[Time.total](../total.md), [Time.part](../part.md), [Time.Unit](Unit.md),
[Duration literals](../../../language/literals/durations.md)
