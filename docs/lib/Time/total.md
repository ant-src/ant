<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.total

```ant signature
func Time.total(d: Time.Duration, unit: Time.Unit, relativeTo: Time.Dated | null = null): float64
```

Time.total returns the length of d in unit, with a fraction: PT1H30M in hours is 1.5. Years, months
and weeks are measured from relativeTo, a PlainDateTime counting as its date, and a day is 24 hours
unless relativeTo is a ZonedDateTime. It panics on years, months or weeks, in d or as unit, when
relativeTo is null.

<!-- end generated -->

## Details

- It measures the whole duration, not one field: `1h30m` in minutes is `90.0`. Use
  [Time.part](part.md) for a single field as given.

## Basic examples

```ant
Time.total(1h30m, Time.Hour{})   // 1.5
Time.total(1h30m, Time.Minute{}) // 90.0
```

Years, months and weeks need a `relativeTo` date to count from. February 2026 has 28 days:

```ant
let feb = Time.plainDate(2026, 2, 1)
Time.total(Time.duration(months: 1), Time.Day{}, relativeTo: feb) // 28.0
```

## Scope

A part of a unit is a fraction:

```ant
20m.total(Time.Hour{}) // 0.3333333333333333
36h.total(Time.Day{})  // 1.5
```

Subsecond fields count at their exact length, in either direction:

```ant
1500ms.total(Time.Second{}) // 1.5
1ns.total(Time.Second{})    // 1e-9
1s.total(Time.Nanosecond{}) // 1000000000.0
```

The empty duration is zero, and a negative duration gives a negative total:

```ant
Time.duration().total(Time.Hour{})    // 0.0
Time.negate(2h15m).total(Time.Hour{}) // -2.25
```

## Named arguments

### relativeTo

A month counts as the length of the month it covers. 45 days from 1 February are the 28 days of
February and 17 of the 31 days of March:

```ant
let days45 = Time.duration(days: 45)
days45.total(Time.Month{}, relativeTo: Time.plainDate(2026, 2, 1)) // 1.5483870967741935
```

A year from a leap year has 366 days:

```ant
let year = Time.duration(years: 1)
year.total(Time.Day{}, relativeTo: Time.plainDate(2027, 1, 1)) // 365.0
year.total(Time.Day{}, relativeTo: Time.plainDate(2028, 1, 1)) // 366.0
```

Weeks also count from `relativeTo`, in `d` or as the unit:

```ant
let feb = Time.plainDate(2026, 2, 1)
Time.duration(weeks: 3).total(Time.Day{}, relativeTo: feb)  // 21.0
Time.duration(days: 10).total(Time.Week{}, relativeTo: feb) // 1.4285714285714286
```

A `ZonedDateTime` gives a day its real length. Stockholm moves its clocks forward on 29 March 2026
and back on 25 October 2026:

```ant
let spring = Time.parseZoned("2026-03-29T00:00:00+01:00[Europe/Stockholm]").unwrap()
let fall = Time.parseZoned("2026-10-25T00:00:00+02:00[Europe/Stockholm]").unwrap()
let day = Time.duration(days: 1)
day.total(Time.Hour{}, relativeTo: spring) // 23.0
day.total(Time.Hour{}, relativeTo: fall)   // 25.0
day.total(Time.Hour{})                     // 24.0
```

A `PlainDateTime` counts as its date, so its time does not change the result:

```ant
let late = Time.plainDateTime(Time.plainDate(2026, 2, 1), Time.plainTime(hour: 23))
Time.duration(months: 1).total(Time.Day{}, relativeTo: late) // 28.0
```

## Applications

Bill a shift by the hour:

```ant
let day = Time.plainDate(2026, 5, 4)
let start = Time.plainDateTime(day, Time.plainTime(hour: 9))
let end = Time.plainDateTime(day, Time.plainTime(hour: 17, minute: 45))
let rate = 40.0

let hours = Time.until(start, end).total(Time.Hour{})
println("${hours} hours, ${hours * rate} due")
```

```text
8.75 hours, 350.0 due
```

Turn lap times into seconds to average them:

```ant
let laps = [
  1m2s300ms,
  59s800ms,
  1m1s,
]
let seconds = laps | Array.map({ x.total(Time.Second{}) })
let sum = seconds.sum() ?? 0.0
println(seconds)
println(sum / float64(seconds.length()))
```

```text
[62.3, 59.8, 61.0]
61.03333333333333
```

Show how far a monthly plan has run, as a percentage of its month:

```ant
let start = Time.plainDate(2026, 4, 1)
let today = Time.plainDate(2026, 4, 19)
let used = Time.until(start, today).total(Time.Month{}, relativeTo: start)
println("${Math.round(used * 100.0)}% of the month used")
```

```text
60.0% of the month used
```

## Properties and relations

The total in a smaller unit is the total in a larger one times the count between them:

```ant
let d = 5h7m3s
assert(d.total(Time.Minute{}) == d.total(Time.Second{}) / 60.0)
```

Negating the duration negates the total:

```ant
let d = 2h30m
assert(Time.negate(d).total(Time.Minute{}) == -d.total(Time.Minute{}))
```

Durations with different fields but one length have the same total, though they are not `==`:

```ant
let a = 90m
let b = 1h30m
assert(a.total(Time.Hour{}) == b.total(Time.Hour{}))
assert(a != b)
```

## Possible issues

Without `relativeTo`, a duration with months panics:

```ant
Time.duration(months: 1).total(Time.Day{})
// panic: Time.total: a total with months needs a relativeTo point
```

So does a total in months, even of a duration that has only days:

```ant
Time.duration(days: 30).total(Time.Month{})
// panic: Time.total: a total in months needs a relativeTo point
```

The result is a `float64`, so compare it with a float literal:

```ant
let hours = 1h.total(Time.Hour{})
let one = hours == 1
// error: float64 and int cannot be compared with ==: neither is assignable to the other
```

## See also

[Time.part](part.md), [Time.compareDuration](compareDuration.md),
[Time.roundDuration](roundDuration.md), [Time.until](until.md), [Time.duration](duration.md),
[Time.Unit](types/Unit.md)
