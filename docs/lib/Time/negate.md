<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.negate

```ant signature
func Time.negate(d: Time.Duration): Time.Duration
```

Time.negate returns d with its sign reversed.

<!-- end generated -->

## Details

- `Time.negate` keeps the fields as given and only flips the sign: `-PT90M` gives `PT90M`, not
  `PT1H30M`.
- It panics when a field is the most negative `int`, whose positive counterpart does not fit.

## Basic examples

```ant
Time.negate(Time.duration(hours: 2))     // -PT2H
Time.negate(Time.duration(minutes: -90)) // PT90M
Time.negate(Time.duration(seconds: 0))   // PT0S
```

## Scope

Every field changes sign, the calendar fields included:

```ant
Time.negate(Time.duration(years: 1, months: 2, days: 3, hours: 4)) // -P1Y2M3DT4H
Time.negate(Time.duration(milliseconds: 1500))                     // -PT1.5S
```

Call it with a dot on a duration, such as one that `Time.until` returns:

```ant
let start = Time.plainDate(2026, 1, 31)
let end = Time.plainDate(2026, 3, 15)
Time.until(start, end)          // P43D
Time.until(start, end).negate() // -P43D
```

Pass `Time.negate` as a function value:

```ant
[Time.duration(hours: 1), Time.duration(minutes: -5)].map(Time.negate) // [-PT1H, PT5M]
```

## Applications

Describe how far away a time is, in the past or the future:

```ant
func relative(now: Time.PlainDateTime, at: Time.PlainDateTime): string {
  let d = Time.until(now, at)
  if d.sign() < 0 { "${d.negate()} ago" } else { "in ${d}" }
}

let today = Time.plainDate(2026, 10, 22)
let now = Time.plainDateTime(today, Time.plainTime(14, 0, 0))
let monday = Time.plainDateTime(Time.plainDate(2026, 10, 20), Time.plainTime(9, 0, 0))
let later = Time.plainDateTime(today, Time.plainTime(16, 30, 0))
relative(now, monday) // "P2DT5H ago"
relative(now, later)  // "in PT2H30M"
```

## Properties and relations

Negating twice gives the duration back, the sign flips, and `Time.subtract` adds the negated
duration:

```ant
let t = Time.plainTime(9, 0, 0)
let ds = [Time.duration(hours: -3), Time.duration(seconds: 0), Time.duration(days: 2, minutes: 5)]
for d in ds {
  assert(d.negate().negate() == d)
  assert(d.negate().sign() == -d.sign())
  assert(Time.subtract(t, d) == Time.add(t, Time.negate(d)))
}
```

`Time.since` is `Time.until` negated:

```ant
let a = Time.plainDate(2026, 1, 31)
let b = Time.plainDate(2026, 3, 15)
Time.since(a, b) == Time.negate(Time.until(a, b)) // true
```

## Possible issues

Unary `-` takes numbers only, not durations:

```ant
let d = Time.duration(hours: 1)
let back = -d // error: unary - needs a number, got Time.Duration
```

Adding a calendar duration and then its negation does not always come back, because the day of the
month is constrained on the way:

```ant
let jan31 = Time.plainDate(2026, 1, 31)
let month = Time.duration(months: 1)
jan31.add(month)                     // 2026-02-28
jan31.add(month).add(month.negate()) // 2026-01-28
```

A field of the most negative `int` has no positive counterpart:

```ant
Time.negate(Time.duration(nanoseconds: Math.minInt))
// panic: Time.negate: the negated duration has a field that does not fit an int
```

## See also

[Time.abs](abs.md), [Time.sign](sign.md), [Time.subtract](subtract.md), [Time.since](since.md),
[Time.until](until.md), [Time.Duration](types/Duration.md)
