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

- Years, months and weeks have no fixed length. [`Time.sleep`](../sleep.md),
  [`Time.ticker`](../ticker.md), [`Time.addDuration`](../addDuration.md) and `Context.new` panic on
  them, and [`Time.compareDuration`](../compareDuration.md), [`Time.total`](../total.md) and
  [`Time.roundDuration`](../roundDuration.md) need a `relativeTo` date for them.
- Days count as 24 hours in a wait or a sum, but [`Time.add`](../add.md) on a zoned time adds
  calendar days, and a [`Time.Instant`](Instant.md) takes no days at all.
- `<` and the other comparison operators do not apply to durations;
  [`Time.compareDuration`](../compareDuration.md) orders them by length.

## Basic examples

Build a duration from named fields with [`Time.duration`](../duration.md):

```ant
Time.duration(hours: 1, minutes: 30) // PT1H30M
Time.duration(days: 3)               // P3D
```

Parse one from its ISO 8601 text with [`Time.parseDuration`](../parseDuration.md):

```ant
Time.parseDuration("PT1H30M") // Ok{ value: PT1H30M }
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
Time.duration(milliseconds: 1500)                               // PT1.5S
Time.duration()                                                 // PT0S
```

A negative duration has a sign in front:

```ant
Time.duration(hours: -2)           // -PT2H
Time.duration(hours: 1).negate()   // -PT1H
Time.duration(minutes: -15).sign() // -1
```

Fields are not balanced. Ninety minutes stays ninety minutes, and two weeks has no days, until
[`Time.roundDuration`](../roundDuration.md) balances it:

```ant
let d = Time.duration(minutes: 90)
d                                         // PT90M
d.part(Time.Hour{})                       // 0
d.roundDuration(largestUnit: Time.Hour{}) // PT1H30M
Time.duration(weeks: 2).part(Time.Day{})  // 0
```

String interpolation, `String.from` and `Json.stringify` all write the same text:

```ant
let d = Time.duration(minutes: 5)
"retry in ${d}"     // "retry in PT5M"
String.from(d)      // "PT5M"
Json.stringify([d]) // "[\"PT5M\"]"
```

A duration is a value like any other: a record field, an optional parameter, a map key:

```ant
type Job = { name: string, timeout: Time.Duration | null }
Job{ name: "build", timeout: Time.duration(minutes: 10) } // Job{ name: "build", timeout: PT10M }
Job{ name: "lint", timeout: null }                        // Job{ name: "lint", timeout: null }
let names = Map{ Time.duration(hours: 1): "hour", Time.duration(minutes: 60): "sixty minutes" }
names.size() // 2
```

```ant
func fetch(url: string, timeout: Time.Duration | null = null): string {
  if timeout == null { "${url} without a timeout" } else { "${url} within ${timeout}" }
}

println(fetch("https://example.com"))
println(fetch("https://example.com", timeout: Time.duration(seconds: 2)))
```

```text
https://example.com without a timeout
https://example.com within PT2S
```

[`Time.total`](../total.md) gives the whole length in one unit as a `float64`:

```ant
Time.duration(hours: 1, minutes: 30).total(Time.Minute{}) // 90.0
Time.duration(minutes: 90).total(Time.Hour{})             // 1.5
```

## Applications

Move a time by a duration with [`Time.add`](../add.md):

```ant
Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(9, 0)).add(Time.duration(hours: 1, minutes: 30)) // 2026-10-22T10:30:00
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
let ctx = Context.new(timeout: Time.duration(milliseconds: 20))
Time.sleep(Time.duration(seconds: 5), ctx: ctx)
println("gave up waiting")
```

```text
gave up waiting
```

Sort durations by length with [`Time.compareDuration`](../compareDuration.md):

```ant
let laps = [
  Time.duration(minutes: 90),
  Time.duration(hours: 1),
  Time.duration(hours: 1, minutes: 15),
]
laps.sortWith(Time.compareDuration) // [PT1H, PT1H15M, PT90M]
```

## Properties and relations

`==` compares fields, and [`Time.compareDuration`](../compareDuration.md) compares lengths:

```ant
let a = Time.duration(minutes: 90)
let b = Time.duration(hours: 1, minutes: 30)
assert(a != b)
assert(Time.compareDuration(a, b) == 0)
assert(a.roundDuration(largestUnit: Time.Hour{}) == b)
```

The text of a duration parses back to an equal duration:

```ant
let d = Time.duration(days: 1, hours: 2, seconds: 3, milliseconds: 40)
assert(Time.parseDuration("${d}").unwrap() == d)
```

Negating twice gives the duration back:

```ant
let d = Time.duration(hours: 3, minutes: 5)
assert(d.negate().negate() == d)
assert(d.negate().abs() == d)
```

## Possible issues

`==` is not a test of length. Two durations of one length are `==` only with the same fields:

```ant
Time.duration(minutes: 90) == Time.duration(hours: 1, minutes: 30)                     // false
Time.compareDuration(Time.duration(minutes: 90), Time.duration(hours: 1, minutes: 30)) // 0
```

`<` does not compare durations. Use [`Time.compareDuration`](../compareDuration.md):

```ant
let short = Time.duration(minutes: 90) < Time.duration(hours: 2)
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.Duration; narrow first
```

A wait needs a fixed length, so a duration in months panics:

```ant
Time.sleep(Time.duration(months: 1))
// panic: Time.sleep needs a duration without years, months or weeks, which have no fixed length
```

All fields of a duration have one sign, so 1h30m less is not `hours: 1, minutes: -30`. Write
`Time.duration(minutes: 30)`, or negate the duration to take away and pass it to
[`Time.addDuration`](../addDuration.md):

```ant
Time.duration(hours: 1, minutes: -30)
// panic: Time.duration: a duration's fields must all have one sign
```

```ant
Time.addDuration(Time.duration(hours: 1), Time.duration(minutes: 30).negate()) // PT30M
```

Text that is not ISO 8601 is an `Err`:

```ant
Time.parseDuration("1 hour") // Err{ error: Error{ message: "not a duration: \"1 hour\"" } }
```

## See also

[Time.duration](../duration.md), [Time.parseDuration](../parseDuration.md),
[Time.until](../until.md), [Time.since](../since.md), [Time.add](../add.md),
[Time.addDuration](../addDuration.md), [Time.compareDuration](../compareDuration.md),
[Time.roundDuration](../roundDuration.md), [Time.total](../total.md), [Time.part](../part.md),
[Time.Unit](Unit.md)
