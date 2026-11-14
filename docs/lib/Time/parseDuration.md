<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.parseDuration

```ant signature
func Time.parseDuration(s: string): Result[Time.Duration, Error]
```

Time.parseDuration parses a duration from s, in ISO 8601, such as P1Y2M3DT4H5M6.5S, -PT1H or PT0S,
or in Go's notation, such as 2h30m, -500ms or 1.5µs. Text whose first character after an optional
sign is P is ISO 8601: PnYnMnWnDTnHnMnS, where only the last time component may have a fraction.
Other text is read as Go's time.ParseDuration reads it: 0, or components of a number and a unit of
h, m, s, ms, us, µs or ns, in any order and repeating, each adding to its own field. Either way the
whole fields are kept as written, and a fraction is spread into the smaller fields, so PT1.5H and
1.5h are 1 hour and 30 minutes; text it cannot parse is an Err.

<!-- end generated -->

## Details

- The whole fields stay as written: `PT90M` is 90 minutes, not 1 hour 30 minutes. Use
  [Time.roundDuration](roundDuration.md) to balance them. A fraction is spread into the smaller
  fields, so `PT1.5H` is `PT1H30M`.
- Text that starts with `P`, after an optional sign, is ISO 8601. Its letters are capitals, and a
  fraction takes a point and up to nine digits; lowercase letters, a comma and spaces are an `Err`.
- Any other text is Go's notation, as configuration files, flags and environment variables write it:
  `2h30m`, `-500ms`, `1.5µs`. Components come in any order and may repeat, each adding to its own
  field, and digits of a fraction below a nanosecond are dropped.

## Basic examples

Parse a duration and take its value:

```ant
Time.parseDuration("PT1H30M").unwrap() // PT1H30M
Time.parseDuration("P1Y2M3DT4H5M6.5S") // Ok{ value: P1Y2M3DT4H5M6.5S }
```

Go's notation, as a duration literal writes it:

```ant
Time.parseDuration("2h30m").unwrap()          // PT2H30M
Time.parseDuration("500ms").unwrap()          // PT0.5S
Time.parseDuration("2h30m").unwrap() == 2h30m // true
```

Text that is not a duration is an `Err`:

```ant
Time.parseDuration("90 minutes") // Err{ error: Error{ message: "not a duration: \"90 minutes\"" } }
```

## Scope

Date fields alone, time fields alone after `T`, or both:

```ant
Time.parseDuration("P3D").unwrap()     // P3D
Time.parseDuration("PT45S").unwrap()   // PT45S
Time.parseDuration("P1DT12H").unwrap() // P1DT12H
```

`M` before `T` is months, and after it minutes:

```ant
Time.parseDuration("P1M").unwrap().part(Time.Month{})   // 1
Time.parseDuration("PT1M").unwrap().part(Time.Minute{}) // 1
```

Weeks, alone or with other fields:

```ant
Time.parseDuration("P2W").unwrap()   // P2W
Time.parseDuration("P1Y1W").unwrap() // P1Y1W
```

A fraction of a second fills the subsecond fields, up to nine digits:

```ant
let d = Time.parseDuration("PT1.5S").unwrap()
d.part(Time.Second{})                                                 // 1
d.part(Time.Millisecond{})                                            // 500
Time.parseDuration("PT0.000000001S").unwrap().part(Time.Nanosecond{}) // 1
```

A leading sign. A minus makes every field negative:

```ant
Time.parseDuration("-PT1H").unwrap()                     // -PT1H
Time.parseDuration("-P1DT2H").unwrap().part(Time.Hour{}) // -2
Time.parseDuration("+PT1H").unwrap()                     // PT1H
```

Zero, written any way, renders as `PT0S`:

```ant
Time.parseDuration("PT0S").unwrap() // PT0S
Time.parseDuration("P0D").unwrap()  // PT0S
```

Go's notation takes components in any order, repeated, and with a fraction anywhere, as Go does.
Each adds to its own field:

```ant
Time.parseDuration("1m2h").unwrap()    // PT2H1M
Time.parseDuration("1h1h").unwrap()    // PT2H
Time.parseDuration("1.5h30m").unwrap() // PT1H60M
```

Go's spellings of microseconds, a sign, and a bare zero:

```ant
Time.parseDuration("1.5µs").unwrap()  // PT0.0000015S
Time.parseDuration("1.5us").unwrap()  // PT0.0000015S
Time.parseDuration("-500ms").unwrap() // -PT0.5S
Time.parseDuration("0").unwrap()      // PT0S
```

Through a pipe:

```ant
"PT15M" | Time.parseDuration() | Result.unwrap() // PT15M
```

## Applications

Read a timeout from configuration, with context on the error:

```ant
func timeout(text: string): Result[Time.Duration, Error] {
  Time.parseDuration(text) | Result.wrapErr("config timeout")
}

println(timeout("PT30S"))
println(timeout("30s"))
println(timeout("30"))
```

```text
Ok{ value: PT30S }
Ok{ value: PT30S }
Err{ error: Error{ message: "config timeout: not a duration: \"30\"" } }
```

Fall back to a default when the setting is missing or wrong:

```ant
let fallback = 10s
Time.parseDuration("").unwrapOr(fallback)     // PT10S
Time.parseDuration("PT2S").unwrapOr(fallback) // PT2S
```

Total a list of durations, failing on the first bad one:

```ant
let laps = ["PT15M", "PT45M", "PT1H"]
let total = laps
  | Array.map(Time.parseDuration)
  | Result.all()
  | Result.map({ x.reduce(Time.duration(), Time.addDuration) })
println(total)
```

```text
Ok{ value: PT2H }
```

Move a date by a period read from text:

```ant
let period = Time.parseDuration("P1M").unwrap()
Time.plainDate(2026, 1, 31).add(period) // 2026-02-28
```

Wait for a parsed duration:

```ant
let pause = try Time.parseDuration("10ms")
Time.sleep(pause)
println("waited ${pause}")
```

```text
waited PT0.01S
```

## Properties and relations

A duration renders in the form `Time.parseDuration` reads. It parses back to the same fields unless
a subsecond field reaches the next unit:

```ant
let d = Time.duration(days: 1, hours: 2, milliseconds: 500)
assert(Time.parseDuration("${d}").unwrap() == d)
```

Rendering folds the subsecond fields into the seconds, and parsing puts a fraction back in the
subsecond fields, so 1500 milliseconds renders as `PT1.5S` and comes back as 1 second and 500
milliseconds:

```ant
let d = 1500ms
"${d}"                                                       // "PT1.5S"
Time.parseDuration("${d}").unwrap() == d                     // false
Time.parseDuration("${d}").unwrap().part(Time.Millisecond{}) // 500
Time.parseDuration("${d}").unwrap().compareDuration(d)       // 0
```

A parsed string has the same fields as the duration literal or [Time.duration](duration.md) call
that writes them:

```ant
assert(Time.parseDuration("PT90M").unwrap() == 90m)
assert(Time.parseDuration("90m").unwrap() == 90m)
```

A leading minus is the [negation](negate.md) of the positive duration:

```ant
let d = Time.parseDuration("PT1H30M").unwrap()
assert(Time.parseDuration("-PT1H30M").unwrap() == d.negate())
```

## Possible issues

`Time.parseDuration` reads ISO 8601 and Go's notation, not a human form. A number needs a unit, and
there is no unit for days in Go's notation:

```ant
Time.parseDuration("1 hour").isOk() // false
Time.parseDuration("90").isOk()     // false
Time.parseDuration("1d").isOk()     // false
Time.parseDuration("P1D").isOk()    // true
```

Go's notation drops what is below a nanosecond, where a duration literal is an error:

```ant
Time.parseDuration("0.1ns").unwrap() // PT0S
```

Lowercase letters, a comma before the fraction, and surrounding spaces are an `Err`. Trim the text
first:

```ant
Time.parseDuration("pt1h").isOk()          // false
Time.parseDuration("PT1,5S").isOk()        // false
Time.parseDuration(" PT1H ").isOk()        // false
Time.parseDuration(" PT1H ".trim()).isOk() // true
```

Only the last time component may carry a fraction, and a date field never does:

```ant
Time.parseDuration("PT1H30.5M").isOk() // true
Time.parseDuration("PT1.5H30M").isOk() // false
Time.parseDuration("P1.5D").isOk()     // false
```

A sign goes in front, not on a field:

```ant
Time.parseDuration("PT-1H").isOk()  // false
Time.parseDuration("-PT1H").isOk()  // true
Time.parseDuration("1h-30m").isOk() // false
```

Fields stay as written, so `==` tells apart two durations of the same length. Use
[Time.compareDuration](compareDuration.md) to compare lengths:

```ant
let a = Time.parseDuration("PT90M").unwrap()
let b = Time.parseDuration("PT1H30M").unwrap()
a == b               // false
a.compareDuration(b) // 0
```

A parsed duration can hold years, months or weeks, which have no fixed length. `Time.sleep`,
`Time.ticker` and `Time.addDuration` panic on them, and so does `Time.compareDuration` without
`relativeTo`. Check the fields before using a parsed duration as a wait:

```ant
Time.parseDuration("P1W").isOk() // true
Time.parseDuration("P1W").unwrap().addDuration(Time.duration(days: 1))
// panic: Time.addDuration: years, months and weeks cannot be added without a starting point; add each to a date
```

```ant
func wait(text: string): Result[Time.Duration, Error] {
  let d = try Time.parseDuration(text)
  if d.part(Time.Year{}) != 0 || d.part(Time.Month{}) != 0 || d.part(Time.Week{}) != 0 {
    return Result.fail("not a fixed length: ${text}")
  }
  Result.ok(d)
}

println(wait("PT5S"))
println(wait("P1M"))
```

```text
Ok{ value: PT5S }
Err{ error: Error{ message: "not a fixed length: P1M" } }
```

Unwrapping text that does not parse panics:

```ant
Time.parseDuration("1 hour").unwrap()
// panic: Result.unwrap on an Err: Error{ message: "not a duration: \"1 hour\"" }
```

## See also

[Time.duration](duration.md), [Duration literals](../../language/literals/durations.md),
[Time.Duration](types/Duration.md), [Time.part](part.md), [Time.roundDuration](roundDuration.md),
[Time.compareDuration](compareDuration.md), [Time.parsePlainTime](parsePlainTime.md),
[Result.wrapErr](../Result/wrapErr.md)
