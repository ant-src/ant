<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.abs

```ant signature
func Time.abs(d: Time.Duration): Time.Duration
```

Time.abs returns d with its sign made positive, or d itself when it is zero.

<!-- end generated -->

## Details

- `Time.abs` keeps the fields as given and drops only the sign: `-PT90M` gives `PT90M`, not
  `PT1H30M`.
- It panics when a field is the most negative `int`, whose positive counterpart does not fit.

## Basic examples

```ant
Time.abs(Time.duration(hours: -2))  // PT2H
Time.abs(Time.duration(hours: 2))   // PT2H
Time.abs(Time.duration(seconds: 0)) // PT0S
```

## Scope

Every field loses its sign, the calendar fields included:

```ant
Time.abs(Time.duration(years: -1, months: -2, days: -3)) // P1Y2M3D
Time.abs(Time.duration(milliseconds: -1500))             // PT1.5S
```

Call it with a dot on a duration, such as one that `Time.until` returns:

```ant
let open = Time.plainTime(9, 0, 0)
let close = Time.plainTime(17, 30, 0)
Time.until(close, open)       // -PT8H30M
Time.until(close, open).abs() // PT8H30M
```

Pass `Time.abs` as a function value:

```ant
[Time.duration(hours: -1), Time.duration(minutes: 30)].map(Time.abs) // [PT1H, PT30M]
```

## Applications

Check that two clocks agree within five minutes, whichever one is ahead:

```ant
func inSync(a: Time.Instant, b: Time.Instant): bool {
  let drift = Time.until(a, b).abs()
  Time.compareDuration(drift, Time.duration(minutes: 5)) <= 0
}

let server = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
let client = Time.parseInstant("2026-10-22T14:00:00Z").unwrap()
inSync(server, client) // true
inSync(client, server) // true
```

The length of a gap between two times of day, in hours, in either order:

```ant
func gapHours(a: Time.PlainTime, b: Time.PlainTime): float64 {
  Time.until(a, b).abs().total(Time.Hour{})
}

gapHours(Time.plainTime(9, 0, 0), Time.plainTime(10, 30, 0)) // 1.5
gapHours(Time.plainTime(10, 30, 0), Time.plainTime(9, 0, 0)) // 1.5
```

## Properties and relations

The result is never negative, ignores the sign of its argument, and leaves a nonnegative duration as
it is:

```ant
let ds = [Time.duration(hours: -3), Time.duration(seconds: 0), Time.duration(days: 2, minutes: 5)]
for d in ds {
  assert(d.abs().sign() >= 0)
  assert(d.negate().abs() == d.abs())
  assert(d.abs() == (if d.sign() < 0 { d.negate() } else { d }))
}
```

## Possible issues

`Time.abs` does not balance, so `==` with a balanced duration is `false`. Compare lengths with
`Time.compareDuration`:

```ant
let d = Time.abs(Time.duration(minutes: -90))
d == Time.duration(hours: 1, minutes: 30)                          // false
Time.compareDuration(d, Time.duration(hours: 1, minutes: 30)) == 0 // true
```

A field of the most negative `int` has no positive counterpart:

```ant
Time.abs(Time.duration(nanoseconds: Math.minInt))
// panic: Time.abs: the negated duration has a field that does not fit an int
```

`Math.abs` takes numbers only, not durations:

```ant
let d = Time.duration(hours: -1)
let e = Math.abs(d) // error: T would be Time.Duration, which is not a member of its bound Number
```

## See also

[Time.negate](negate.md), [Time.sign](sign.md), [Time.compareDuration](compareDuration.md),
[Time.until](until.md), [Math.abs](../Math/abs.md), [Time.Duration](types/Duration.md)
