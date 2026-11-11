<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.sign

```ant signature
func Time.sign(d: Time.Duration): int
```

Time.sign returns -1, 0 or 1 as d is negative, zero or positive.

<!-- end generated -->

## Details

- `Time.sign` reads the sign without measuring the duration, so it takes years, months and weeks
  with no `relativeTo` point.

## Basic examples

```ant
Time.sign(Time.duration(hours: 2))     // 1
Time.sign(Time.duration(minutes: -90)) // -1
Time.sign(Time.duration())             // 0
```

## Scope

Calendar fields count like any other, and one nanosecond is enough:

```ant
Time.sign(Time.duration(years: -1, days: -3)) // -1
Time.sign(Time.duration(nanoseconds: 1))      // 1
```

A parsed duration carries its leading sign:

```ant
Time.parseDuration("-P1D") | Result.unwrap() | Time.sign() // -1
```

Call it with a dot on a duration, such as one that `Time.until` returns:

```ant
let due = Time.plainDate(2026, 3, 15)
let today = Time.plainDate(2026, 4, 2)
Time.until(today, due)        // -P18D
Time.until(today, due).sign() // -1
```

Pass `Time.sign` as a function value:

```ant
[Time.duration(hours: 1), Time.duration(), Time.duration(minutes: -5)].map(Time.sign) // [1, 0, -1]
```

## Named arguments

```ant
Time.sign(d: Time.duration(seconds: -5)) // -1
```

## Applications

Label a deadline as overdue, due now, or upcoming:

```ant
func status(now: Time.PlainDateTime, due: Time.PlainDateTime): string {
  let s = Time.until(now, due).sign()
  if s < 0 { "overdue" } else if s == 0 { "due now" } else { "upcoming" }
}

let now = Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(14, 0, 0))
let morning = Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(9, 0, 0))
let friday = Time.plainDateTime(Time.plainDate(2026, 10, 23), Time.plainTime(17, 0, 0))
status(now, morning) // "overdue"
status(now, now)     // "due now"
status(now, friday)  // "upcoming"
```

Keep the offsets in a schedule that point back in time:

```ant
let offsets = [Time.duration(days: -1), Time.duration(hours: 2), Time.duration(minutes: -15)]
offsets.filter({ x.sign() < 0 }) // [-P1D, -PT15M]
```

## Properties and relations

Negating flips the sign, `Time.abs` drops it, and only the zero duration has sign `0`:

```ant
let zero = Time.duration()
let ds = [Time.duration(hours: -3), zero, Time.duration(days: 2, minutes: 5)]
for d in ds {
  assert(d.negate().sign() == -d.sign())
  assert(d.abs().sign() == Math.abs(d.sign()))
  assert((d.sign() == 0) == (d == zero))
  assert(d.sign() == Time.compareDuration(d, zero))
}
```

The sign of `Time.until(a, b)` is `Time.compare(b, a)`:

```ant
let a = Time.plainDate(2026, 3, 15)
let b = Time.plainDate(2026, 1, 31)
Time.until(a, b).sign() == Time.compare(b, a) // true
```

## Possible issues

The comparison operators do not apply to durations, so `d < Time.duration()` does not compile. Use
`Time.sign`:

```ant
let d = Time.duration(hours: -1)
let late = d < Time.duration()
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.Duration
```

`Time.compareDuration` against zero panics on a calendar duration without `relativeTo`, while
`Time.sign` answers:

```ant
let month = Time.duration(months: -1)
month.sign() // -1
Time.compareDuration(month, Time.duration())
// panic: Time.compareDuration: years, months or weeks need a relativeTo point to compare
```

## See also

[Time.negate](negate.md), [Time.abs](abs.md), [Time.compareDuration](compareDuration.md),
[Time.until](until.md), [Time.compare](compare.md), [Time.Duration](types/Duration.md)
