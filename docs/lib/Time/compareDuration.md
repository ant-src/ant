<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.compareDuration

```ant signature
func Time.compareDuration(a: Time.Duration, b: Time.Duration, relativeTo: Time.Dated | null = null): int
```

Time.compareDuration returns -1, 0 or 1 as a is shorter than, as long as or longer than b, by signed
length, so a negative duration is shorter than any positive one. Years, months and weeks are
measured from relativeTo, a week being 7 days, and a day is 24 hours unless relativeTo is a
ZonedDateTime, whose days have their daylight-saving hours. It panics if either has years, months or
weeks and relativeTo is null.

<!-- end generated -->

## Details

- It compares lengths, not fields: `90m` and `1h30m` compare as `0` though they are not `==`.
- Two durations with the same fields compare as `0` without `relativeTo`, even with years, months or
  weeks.
- A `PlainDateTime` given as `relativeTo` counts as its date; its time is ignored.
- The comparison operators do not apply to durations; `Time.compareDuration` is their ordering.

## Basic examples

```ant
let month = Time.duration(months: 1)
let feb = Time.plainDate(2026, 2, 1)
Time.compareDuration(month, Time.duration(days: 30), relativeTo: feb) // -1
```

Durations without years, months or weeks need no `relativeTo`:

```ant
Time.compareDuration(30m, 1h)   // -1
Time.compareDuration(90m, 1h)   // 1
Time.compareDuration(3600s, 1h) // 0
```

## Scope

Days count as 24 hours:

```ant
Time.compareDuration(Time.duration(days: 1), 24h) // 0
Time.compareDuration(25h, Time.duration(days: 1)) // 1
```

Subsecond fields count at their exact length:

```ant
let a = 1500ms
let b = 1s500ms
Time.compareDuration(a, b)                 // 0
Time.compareDuration(1ns, Time.duration()) // 1
```

## Named arguments

### relativeTo

A month has the length of the month it starts in, so the same pair can compare either way:

```ant
let month = Time.duration(months: 1)
let days30 = Time.duration(days: 30)
month.compareDuration(days30, relativeTo: Time.plainDate(2026, 2, 1)) // -1
month.compareDuration(days30, relativeTo: Time.plainDate(2026, 3, 1)) // 1
```

A year from a leap year has 366 days:

```ant
let year = Time.duration(years: 1)
let days365 = Time.duration(days: 365)
year.compareDuration(days365, relativeTo: Time.plainDate(2027, 1, 1)) // 0
year.compareDuration(days365, relativeTo: Time.plainDate(2028, 1, 1)) // 1
```

A `ZonedDateTime` gives the day its real length. Stockholm moves its clocks forward on 29 March
2026, so that day has 23 hours:

```ant
let z = Time.parseZoned("2026-03-29T00:00:00+01:00[Europe/Stockholm]").unwrap()
let day = Time.duration(days: 1)
day.compareDuration(24h, relativeTo: z) // -1
day.compareDuration(23h, relativeTo: z) // 0
day.compareDuration(24h)                // 0
```

A `PlainDateTime` counts as its date; its time does not change the result:

```ant
let start = Time.plainDateTime(Time.plainDate(2026, 2, 1), Time.plainTime(hour: 23))
let month = Time.duration(months: 1)
month.compareDuration(Time.duration(days: 28), relativeTo: start) // 0
```

## Applications

Check how long a job ran against its time budget:

```ant
let started = Time.plainDateTime(Time.plainDate(2026, 5, 4), Time.plainTime(hour: 9))
let finished = Time.plainDateTime(Time.plainDate(2026, 5, 4), Time.plainTime(hour: 11, minute: 15))
let budget = 2h

let took = Time.until(started, finished)
if took.compareDuration(budget) > 0 {
  println("over budget: ${took}")
} else {
  println("on time: ${took}")
}
```

```text
over budget: PT2H15M
```

Sort durations by length. `Time.compareDuration` is a comparator, so it passes to `Array.sortWith`
directly:

```ant
let timeouts = [2h, 30m, 4000s]
timeouts | Array.sortWith(Time.compareDuration) // [PT30M, PT4000S, PT2H]
```

Find the months of a year in which a one-month plan is shorter than a 30-day trial:

```ant
let month = Time.duration(months: 1)
let trial = Time.duration(days: 30)
let shorter = Array.range(1, 13)
  | Array.filter({ month.compareDuration(trial, relativeTo: Time.plainDate(2026, x, 1)) < 0 })
println(shorter)
```

```text
[2]
```

## Properties and relations

Swapping the arguments negates the result:

```ant
let a = 90m
let b = 2h
assert(Time.compareDuration(a, b) == -Time.compareDuration(b, a))
```

The order agrees with [Time.total](total.md) in any unit:

```ant
let a = 1h30m
let b = 5000s
let longer = a.total(Time.Second{}) > b.total(Time.Second{})
assert((Time.compareDuration(a, b) > 0) == longer)
```

`0` means the same length, while `==` compares the fields as given:

```ant
let a = 90m
let b = 1h30m
assert(Time.compareDuration(a, b) == 0)
assert(a != b)
```

## Possible issues

Without `relativeTo`, a duration with years, months or weeks panics:

```ant
let month = Time.duration(months: 1)
month.compareDuration(Time.duration(days: 30))
// panic: Time.compareDuration: years, months or weeks need a relativeTo point to compare
```

Two durations with the same fields are the exception, so this does not panic:

```ant
Time.compareDuration(Time.duration(months: 1), Time.duration(months: 1)) // 0
```

`relativeTo` takes a date, so an `Instant` such as `Time.now()` does not fit:

```ant
let month = Time.duration(months: 1)
month.compareDuration(Time.duration(days: 30), relativeTo: Time.now())
// error: cannot use Time.Instant where Time.Dated | null is expected
```

Place the instant in a zone first:

```ant
let zone = Time.zone("Europe/Stockholm").unwrap()
let month = Time.duration(months: 1)
let now = Time.now().inZone(zone)
let shorter = month.compareDuration(Time.duration(days: 32), relativeTo: now) < 0
assert(shorter)
```

The comparison operators do not apply to durations:

```ant
let a = 1h
let b = 30m
let longer = a > b
// error: > applies to a single numeric kind, string, or a named primitive type over one, not to Time.Duration
```

`==` between two duration literals of the same length with different fields does not compile:

```ant
let same = 90m == 1h30m
// error: 90m == 1h30m compares the fields as written, and == is false: compare the lengths with Time.compareDuration(90m, 1h30m)
```

Compare their lengths instead:

```ant
Time.compareDuration(90m, 1h30m) // 0
```

## See also

[Time.compare](compare.md), [Time.total](total.md), [Time.addDuration](addDuration.md),
[Time.sign](sign.md), [Time.roundDuration](roundDuration.md), [Time.duration](duration.md),
[Array.sortWith](../Array/sortWith.md), [Duration literals](../../language/literals/durations.md)
