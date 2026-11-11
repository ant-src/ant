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

- It compares lengths, not fields: `PT90M` and `PT1H30M` compare as `0` though they are not `==`.
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
Time.compareDuration(Time.duration(minutes: 30), Time.duration(hours: 1))   // -1
Time.compareDuration(Time.duration(minutes: 90), Time.duration(hours: 1))   // 1
Time.compareDuration(Time.duration(seconds: 3600), Time.duration(hours: 1)) // 0
```

## Scope

Days count as 24 hours:

```ant
Time.compareDuration(Time.duration(days: 1), Time.duration(hours: 24)) // 0
Time.compareDuration(Time.duration(hours: 25), Time.duration(days: 1)) // 1
```

Subsecond fields count at their exact length:

```ant
let a = Time.duration(milliseconds: 1500)
let b = Time.duration(seconds: 1, milliseconds: 500)
Time.compareDuration(a, b)                                           // 0
Time.compareDuration(Time.duration(nanoseconds: 1), Time.duration()) // 1
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
day.compareDuration(Time.duration(hours: 24), relativeTo: z) // -1
day.compareDuration(Time.duration(hours: 23), relativeTo: z) // 0
day.compareDuration(Time.duration(hours: 24))                // 0
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
let budget = Time.duration(hours: 2)

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
let timeouts = [Time.duration(hours: 2), Time.duration(minutes: 30), Time.duration(seconds: 4000)]
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
let a = Time.duration(minutes: 90)
let b = Time.duration(hours: 2)
assert(Time.compareDuration(a, b) == -Time.compareDuration(b, a))
```

The order agrees with [Time.total](total.md) in any unit:

```ant
let a = Time.duration(hours: 1, minutes: 30)
let b = Time.duration(seconds: 5000)
let longer = a.total(Time.Second{}) > b.total(Time.Second{})
assert((Time.compareDuration(a, b) > 0) == longer)
```

`0` means the same length, while `==` compares the fields as given:

```ant
let a = Time.duration(minutes: 90)
let b = Time.duration(hours: 1, minutes: 30)
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
let a = Time.duration(hours: 1)
let b = Time.duration(minutes: 30)
let longer = a > b
// error: > applies to a single numeric kind, string, or a named primitive type over one, not to Time.Duration
```

## See also

[Time.compare](compare.md), [Time.total](total.md), [Time.addDuration](addDuration.md),
[Time.sign](sign.md), [Time.roundDuration](roundDuration.md), [Time.duration](duration.md),
[Array.sortWith](../Array/sortWith.md)
