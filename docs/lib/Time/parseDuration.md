<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.parseDuration

```ant signature
func Time.parseDuration(s: string): Result[Time.Duration, Error]
```

Time.parseDuration parses an ISO 8601 duration from s, such as P1Y2M3DT4H5M6.5S, -PT1H or PT0S. It
takes PnYnMnWnDTnHnMnS with an optional leading sign, where only the last time component may have a
fraction, and keeps the fields as written; text it cannot parse is an Err.

<!-- end generated -->

## Details

- The fields stay as written: `PT90M` is 90 minutes, not 1 hour 30 minutes. Use
  [Time.roundDuration](roundDuration.md) to balance them.
- Letters are capitals, and a fraction takes a point and up to nine digits. Lowercase letters, a
  comma, spaces, and Go-style text such as `1h30m` are an `Err`.

## Basic examples

Parse a duration and take its value:

```ant
Time.parseDuration("PT1H30M").unwrap() // PT1H30M
Time.parseDuration("P1Y2M3DT4H5M6.5S") // Ok{ value: P1Y2M3DT4H5M6.5S }
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
```

```text
Ok{ value: PT30S }
Err{ error: Error{ message: "config timeout: not a duration: \"30s\"" } }
```

Fall back to a default when the setting is missing or wrong:

```ant
let fallback = Time.duration(seconds: 10)
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
let pause = try Time.parseDuration("PT0.01S")
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
let d = Time.duration(milliseconds: 1500)
"${d}"                                                       // "PT1.5S"
Time.parseDuration("${d}").unwrap() == d                     // false
Time.parseDuration("${d}").unwrap().part(Time.Millisecond{}) // 500
Time.parseDuration("${d}").unwrap().compareDuration(d)       // 0
```

A parsed string has the same fields as [Time.duration](duration.md) given them:

```ant
assert(Time.parseDuration("PT90M").unwrap() == Time.duration(minutes: 90))
```

A leading minus is the [negation](negate.md) of the positive duration:

```ant
let d = Time.parseDuration("PT1H30M").unwrap()
assert(Time.parseDuration("-PT1H30M").unwrap() == d.negate())
```

## Possible issues

`Time.parseDuration` reads ISO 8601, not the Go or human form:

```ant
Time.parseDuration("1h30m").isOk()   // false
Time.parseDuration("PT1H30M").isOk() // true
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

A sign goes in front of the `P`, not on a field:

```ant
Time.parseDuration("PT-1H").isOk() // false
Time.parseDuration("-PT1H").isOk() // true
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

[Time.duration](duration.md), [Time.Duration](types/Duration.md), [Time.part](part.md),
[Time.roundDuration](roundDuration.md), [Time.compareDuration](compareDuration.md),
[Time.parsePlainTime](parsePlainTime.md), [Result.wrapErr](../Result/wrapErr.md)
