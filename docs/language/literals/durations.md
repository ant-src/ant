<!-- status: reviewed -->

# Duration literals

A number followed by time units, such as `500ms`, `2h30m`, `1.5s`, or `90m`, is a duration literal
of type `Time.Duration`, written as in Go. It needs no import.

## Details

- The units are `h`, `m`, `s`, `ms`, `us`, and `ns`, hours through nanoseconds, so `m` is minutes.
  Each unit appears at most once, in descending order, with no white space between components.
- A number is a decimal integer or float, with `_` between digits and no exponent. Only the last
  component may have a fraction; it spreads into the smaller units and must come out as a whole
  number of nanoseconds.
- A literal keeps its fields as written: `90m` is 90 minutes, not 1 hour and 30 minutes.
- There are no literals for days, weeks, months, or years, and `-`, `+`, `*`, and the other
  arithmetic operators do not apply to a duration (`==` compares fields). Write `Time.duration(...)`
  for those spans, for a span computed from a number, and `Time.negate` for a negative one.

## Basic examples

A literal renders as an ISO 8601 duration:

```ant
500ms // PT0.5S
2h30m // PT2H30M
1.5s  // PT1.5S
```

Pass a literal wherever a function takes a `Time.Duration`:

```ant
let ctx = Context.new(50ms)
Time.sleep(1s, ctx: ctx)
println(ctx.reason())
```

```text
timeout
```

## Scope

### Units

Each unit, from hours to nanoseconds:

```ant
2h    // PT2H
90m   // PT90M
45s   // PT45S
250ms // PT0.25S
10us  // PT0.00001S
5ns   // PT0.000000005S
```

Components combine in descending order:

```ant
1h30m      // PT1H30M
1m30s500ms // PT1M30.5S
2h0m5s     // PT2H5S
```

### Zero and digit separators

`0s` is the zero span, and `_` groups digits:

```ant
0s      // PT0S
1_000ms // PT1S
```

### Fractions

A fraction on the last component spreads into the smaller units, exactly:

```ant
1.5h    // PT1H30M
1h30.5m // PT1H30M30S
0.5ms   // PT0.0005S
1.25s   // PT1.25S
```

### The type in any context

A literal is a `Time.Duration` wherever it stands: a record field default, a named-argument default,
a map key:

```ant
type Job = { name: string, timeout: Time.Duration = 10m }
Job{ name: "backup" }.timeout           // PT10M
Map{ 1h: "hour", 60m: "sixty minutes" } // Map{ PT1H: "hour", PT60M: "sixty minutes" }
```

```ant
func retry(attempts: int, wait: Time.Duration = 200ms): string {
  "${attempts} attempts, ${wait} apart"
}

println(retry(3))
println(retry(5, wait: 2s))
```

```text
3 attempts, PT0.2S apart
5 attempts, PT2S apart
```

### Fields as written

A literal keeps its fields as written, and `==` compares fields, so two literals of the same length
are not always equal. `Time.compareDuration` compares lengths:

```ant
let a = 90m
let b = 1h30m
a == b                     // false
Time.compareDuration(a, b) // 0
1.5h == 1h30m              // true
```

### Calls through `.`

A `.` after the unit is member access, so a literal reaches the `Time` functions that take a
duration first:

```ant
1h.negate()                                 // -PT1H
90m.total(Time.Minute{})                    // 90.0
90m.roundDuration(largestUnit: Time.Hour{}) // PT1H30M
```

### Rendering

`String.from` and string interpolation give the ISO 8601 form:

```ant
String.from(500ms) // "PT0.5S"
"${2h30m}"         // "PT2H30M"
```

`ant fmt` prints a literal exactly as written: `90m` stays `90m`, and `1_000ms` stays `1_000ms`.

## Applications

Time out a slow operation with `select`:

```ant
let result: Channel[string] = Channel.new(capacity: 1)
go {
  Time.sleep(1s)
  result.send("done")
}
select {
  let r = Channel.receive(result) {
    println(r ?? "closed")
  }
  Time.sleep(50ms) {
    println("timed out")
  }
}
```

```text
timed out
```

Shift a time by a constant span:

```ant
let start = Time.plainTime(9, 0)
start.add(1h30m)    // 10:30:00
start.subtract(15m) // 08:45:00
```

## Properties and relations

A literal and the constructor with the same fields are equal:

```ant
assert(2h30m == Time.duration(hours: 2, minutes: 30))
assert(1.5h == Time.duration(hours: 1, minutes: 30))
assert(500ms == Time.duration(milliseconds: 500))
```

## Possible issues

### Comparing two literals with `==`

`ant check` rejects `==` or `!=` between two literals of the same length and different fields:

```ant
let same = 90m == 1h30m
// error: 90m == 1h30m compares the fields as written, and == is false: compare the lengths with Time.compareDuration(90m, 1h30m)
```

Compare the lengths:

```ant
Time.compareDuration(90m, 1h30m) == 0 // true
```

### No days, weeks, months, or years

Their length depends on the calendar, so they have no literal:

```ant
let day = 1d
// error: there are no duration literals for days, weeks, months or years: write Time.duration(days: 1) and the like
```

```ant
Time.duration(days: 1) // P1D
```

`24h` is exactly 24 hours, not the next day across a daylight saving change.

### No arithmetic on durations

```ant
let back = -5s // error: unary - needs a number, got Time.Duration
```

```ant
let total = 5s + 1s
// error: + applies to a single numeric kind, string, or a named primitive type over one, not to Time.Duration
```

```ant
let twice = 2 * 1s // error: * needs both operands of one type, got int and Time.Duration
```

Call the `Time` functions instead:

```ant
Time.negate(5s)          // -PT5S
5s.negate()              // -PT5S
Time.addDuration(5s, 1s) // PT6S
```

### A computed span is not a literal

A literal holds constants only. Build a span from a number with `Time.duration`:

```ant
let n = 3
Time.duration(seconds: n * 2) // PT6S
```

### Every number needs a unit

```ant
let wait = 1h30 // error: the number 30 in a duration literal needs a unit
```

### A fraction goes on the last component only

```ant
let wait = 1.5h30m // error: only the last component of a duration literal can have a fraction
```

### Other rejected spellings

The lexer reads every letter, digit, and point after the number as part of the literal, so a wrong
unit is an error rather than a name:

- `min` is not a unit; minutes are `m`:

  ```ant
  let wait = 1min
  // error: malformed number literal: min is not a duration unit, which are h, m, s, ms, us and ns
  ```

- `M` is not a unit:

  ```ant
  let wait = 1M
  // error: malformed number literal: M is not a duration unit, which are h, m, s, ms, us and ns
  ```

- Microseconds are `us`, not `µs`:

  ```ant
  let wait = 1µs // error: malformed duration literal: write microseconds as us
  ```

- A duration has no exponent:

  ```ant
  let wait = 1e3s // error: malformed duration literal: a duration has no exponent
  ```

- The number is decimal, with no prefix:

  ```ant
  let wait = 0x1s // error: malformed duration literal: a duration is decimal, as 1s
  ```

- The value is not a whole number of nanoseconds:

  ```ant
  let wait = 0.1ns // error: duration literal 0.1ns is not a whole number of nanoseconds
  ```

- The units are out of order:

  ```ant
  let wait = 1m2h
  // error: the units of a duration literal come in descending order, each at most once
  ```

- A unit appears twice:

  ```ant
  let wait = 1h1h
  // error: the units of a duration literal come in descending order, each at most once
  ```

- A point after a digit begins a fraction:

  ```ant
  let wait = 1.s // error: a float literal needs a digit after the point
  ```

[Time.parseDuration](../../lib/Time/parseDuration.md) reads text in Go's looser notation, where
`"1µs"` and `"1m2h"` are fine.

### Out of range

The time part of a duration stays below 2^53 seconds:

```ant
let long = 3000000000000h
// error: duration literal out of range: the time part of a duration must be below 2^53 seconds
```

## See also

[Number literals](numbers.md) · [Time.Duration](../../lib/Time/types/Duration.md) ·
[Time.duration](../../lib/Time/duration.md) · [Time.parseDuration](../../lib/Time/parseDuration.md)
· [Time.compareDuration](../../lib/Time/compareDuration.md) ·
[Time.negate](../../lib/Time/negate.md)
