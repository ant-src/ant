<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.roundDuration

```ant signature
func Time.roundDuration(d: Time.Duration, smallestUnit: Time.Unit | null = null, largestUnit: Time.Unit | null = null, increment: int = 1, mode: Time.RoundingMode = Time.HalfExpand{}, relativeTo: Time.Dated | null = null): Time.Duration
```

Time.roundDuration balances d so that no field below largestUnit exceeds its next unit, and rounds
it to a multiple of increment of smallestUnit by mode, dropping the smaller fields. largestUnit
defaults to the larger of the largest nonzero unit of d and smallestUnit; years, months and weeks
need relativeTo, as in [`Time.total`](total.md). increment follows [`Time.round`](round.md) for a
time unit, and may be any positive one for a day or calendar unit when largestUnit is smallestUnit.
It panics if neither unit is given, if largestUnit is smaller than smallestUnit, if increment is
below 1, on years, months or weeks without relativeTo, and on an increment that breaks those rules.

<!-- end generated -->

## Details

- `largestUnit` alone keeps the length; only `smallestUnit` changes it.
- A duration with years, months or weeks, or a `largestUnit` or `smallestUnit` of `Year`, `Month` or
  `Week`, needs `relativeTo`. A day is 24 hours unless `relativeTo` is a `ZonedDateTime`, whose days
  have their daylight-saving hours.
- The default `largestUnit` is the largest nonzero field of `d`, or `smallestUnit` when that is
  larger, so rounding `PT59M50S` to minutes gives `PT60M`, not `PT1H`. Pass `largestUnit` to carry
  into the next unit.
- The rounding modes treat a negative duration as a number below zero: `Ceil` moves it toward zero,
  `Floor` away from it.

## Basic examples

```ant
Time.roundDuration(Time.duration(minutes: 90), largestUnit: Time.Hour{})            // PT1H30M
Time.roundDuration(Time.duration(hours: 1, minutes: 29), smallestUnit: Time.Hour{}) // PT1H
Time.roundDuration(Time.duration(hours: 1, minutes: 30), smallestUnit: Time.Hour{}) // PT2H
```

Give both units to balance and round in one call:

```ant
Time.roundDuration(Time.duration(seconds: 7384), largestUnit: Time.Hour{}, smallestUnit: Time.Minute{}) // PT2H3M
```

## Scope

`largestUnit` balances up, carrying seconds into minutes, hours and days:

```ant
let d = Time.duration(seconds: 100000)
d.roundDuration(largestUnit: Time.Day{})  // P1DT3H46M40S
d.roundDuration(largestUnit: Time.Hour{}) // PT27H46M40S
```

It also balances down, spreading larger fields into a smaller unit:

```ant
Time.duration(hours: 1, minutes: 30).roundDuration(largestUnit: Time.Minute{}) // PT90M
Time.duration(days: 1).roundDuration(largestUnit: Time.Hour{})                 // PT24H
```

Fields above `smallestUnit` stay as they are; only the part below it rounds:

```ant
Time.duration(days: 3, hours: 13).roundDuration(smallestUnit: Time.Day{})       // P4D
Time.duration(milliseconds: 1234567).roundDuration(smallestUnit: Time.Second{}) // PT1235S
```

A negative duration rounds by its length and keeps its sign:

```ant
Time.duration(hours: -2, minutes: -30).roundDuration(smallestUnit: Time.Hour{}) // -PT3H
```

A zero duration stays zero:

```ant
Time.duration().roundDuration(smallestUnit: Time.Second{}) // PT0S
```

## Named arguments

### smallestUnit

The fields below `smallestUnit` are dropped after rounding:

```ant
let d = Time.duration(hours: 2, minutes: 37, seconds: 25)
d.roundDuration(smallestUnit: Time.Minute{}) // PT2H37M
d.roundDuration(smallestUnit: Time.Hour{})   // PT3H
```

### largestUnit

`largestUnit` sets the top field of the result. Rounding can carry into it:

```ant
let d = Time.duration(minutes: 59, seconds: 50)
d.roundDuration(smallestUnit: Time.Minute{})                           // PT60M
d.roundDuration(smallestUnit: Time.Minute{}, largestUnit: Time.Hour{}) // PT1H
```

### increment

`increment` rounds to a multiple of several units. For a time unit it must divide the count in the
next unit, 60 minutes or 24 hours, and be smaller than it:

```ant
let d = Time.duration(minutes: 37)
d.roundDuration(smallestUnit: Time.Minute{}, increment: 15)                      // PT30M
Time.duration(hours: 37).roundDuration(smallestUnit: Time.Hour{}, increment: 12) // PT36H
```

For days and calendar units any increment works, as long as `largestUnit` is `smallestUnit`:

```ant
Time.duration(days: 10).roundDuration(smallestUnit: Time.Day{}, increment: 7) // P7D
let jan = Time.plainDate(2026, 1, 1)
Time.duration(months: 5).roundDuration(smallestUnit: Time.Month{}, increment: 2, relativeTo: jan) // P6M
```

### mode

`mode` picks the direction. `Ceil` and `Floor` go toward positive and negative infinity, `Expand`
and `Trunc` away from and toward zero:

```ant
let d = Time.duration(minutes: -37)
d.roundDuration(smallestUnit: Time.Minute{}, increment: 15, mode: Time.Ceil{})   // -PT30M
d.roundDuration(smallestUnit: Time.Minute{}, increment: 15, mode: Time.Floor{})  // -PT45M
d.roundDuration(smallestUnit: Time.Minute{}, increment: 15, mode: Time.Expand{}) // -PT45M
d.roundDuration(smallestUnit: Time.Minute{}, increment: 15, mode: Time.Trunc{})  // -PT30M
```

The `Half` modes go to the nearest multiple and differ only on a tie. `HalfEven` takes the even
multiple:

```ant
let tie = Time.duration(hours: 1, minutes: 30)
tie.roundDuration(smallestUnit: Time.Hour{})                         // PT2H
tie.roundDuration(smallestUnit: Time.Hour{}, mode: Time.HalfTrunc{}) // PT1H
let even = Time.HalfEven{}
Time.duration(milliseconds: 2500).roundDuration(smallestUnit: Time.Second{}, mode: even) // PT2S
Time.duration(milliseconds: 3500).roundDuration(smallestUnit: Time.Second{}, mode: even) // PT4S
```

### relativeTo

Months have different lengths, so the same 45 days balance differently from January and from
February:

```ant
let d = Time.duration(days: 45)
d.roundDuration(largestUnit: Time.Month{}, relativeTo: Time.plainDate(2026, 1, 1)) // P1M14D
d.roundDuration(largestUnit: Time.Month{}, relativeTo: Time.plainDate(2026, 2, 1)) // P1M17D
```

Weeks and years count from `relativeTo` too:

```ant
let jan = Time.plainDate(2026, 1, 1)
Time.duration(days: 10).roundDuration(smallestUnit: Time.Week{}, relativeTo: jan)            // P1W
Time.duration(months: 14).roundDuration(largestUnit: Time.Year{}, relativeTo: jan)           // P1Y2M
Time.duration(years: 1, months: 7).roundDuration(smallestUnit: Time.Year{}, relativeTo: jan) // P2Y
```

A `ZonedDateTime` gives each day its real length. Stockholm moves its clocks forward on 29 March
2026, so that day has 23 hours:

```ant
let z = Time.parseZoned("2026-03-29T00:00:00+01:00[Europe/Stockholm]").unwrap()
let day = Time.duration(days: 1)
day.roundDuration(largestUnit: Time.Hour{}, relativeTo: z) // PT23H
day.roundDuration(largestUnit: Time.Hour{})                // PT24H
```

## Applications

Show how long a job ran, to the minute:

```ant
let start = Time.plainDateTime(Time.plainDate(2026, 5, 4), Time.plainTime(hour: 9, minute: 12, second: 40))
let end = Time.plainDateTime(Time.plainDate(2026, 5, 4), Time.plainTime(hour: 11, minute: 50, second: 5))
let took = Time.until(start, end, largestUnit: Time.Second{})
println(took)
println(took.roundDuration(largestUnit: Time.Hour{}, smallestUnit: Time.Minute{}))
```

```text
PT9445S
PT2H37M
```

Bill each session in started blocks of 15 minutes:

```ant
let sessions = [Time.duration(minutes: 7), Time.duration(minutes: 31), Time.duration(minutes: 45)]
let billed = sessions
  | Array.map({ x.roundDuration(smallestUnit: Time.Minute{}, increment: 15, mode: Time.Ceil{}) })
println(billed)
```

```text
[PT15M, PT45M, PT45M]
```

Turn a count of seconds into days, hours and minutes for a status line:

```ant
func uptime(seconds: int): string {
  let d = Time.duration(seconds: seconds)
  "up ${d.roundDuration(largestUnit: Time.Day{}, smallestUnit: Time.Minute{})}"
}
println(uptime(273_845))
println(uptime(59))
```

```text
up P3DT4H4M
up PT1M
```

## Properties and relations

Balancing alone keeps the length, which [Time.compareDuration](compareDuration.md) and
[Time.total](total.md) measure:

```ant
let d = Time.duration(seconds: 100000)
let balanced = d.roundDuration(largestUnit: Time.Day{})
assert(Time.compareDuration(d, balanced) == 0)
assert(Time.total(d, Time.Second{}) == Time.total(balanced, Time.Second{}))
assert(d != balanced)
```

Rounding a duration and negating it commute for a symmetric mode such as the default:

```ant
let d = Time.duration(minutes: 37)
let unit = Time.Minute{}
let a = Time.negate(d.roundDuration(smallestUnit: unit, increment: 15))
let b = Time.negate(d).roundDuration(smallestUnit: unit, increment: 15)
assert(a == b)
```

`Ceil` never gives less than the input, and `Floor` never more:

```ant
let d = Time.duration(minutes: 37)
let up = d.roundDuration(smallestUnit: Time.Minute{}, increment: 15, mode: Time.Ceil{})
let down = d.roundDuration(smallestUnit: Time.Minute{}, increment: 15, mode: Time.Floor{})
assert(Time.compareDuration(up, d) >= 0)
assert(Time.compareDuration(down, d) <= 0)
```

## Possible issues

Without a unit there is nothing to do, and it panics:

```ant
Time.roundDuration(Time.duration(minutes: 90))
// panic: Time.roundDuration: roundDuration needs a smallest or a largest unit
```

Months, years and weeks need `relativeTo`, in `d` or as a unit:

```ant
Time.duration(months: 1).roundDuration(smallestUnit: Time.Day{})
// panic: Time.roundDuration: balancing months needs a relativeTo point
```

```ant
Time.duration(days: 45).roundDuration(largestUnit: Time.Month{})
// panic: Time.roundDuration: balancing to months needs a relativeTo point
```

An increment of a time unit must divide the next unit and be smaller than it:

```ant
let d = Time.duration(minutes: 37)
d.roundDuration(smallestUnit: Time.Minute{}, increment: 7)
// panic: Time.roundDuration: the increment for minutes must divide 60 and be smaller, not 7
```

The increment cannot be the whole next unit either:

```ant
let d = Time.duration(minutes: 37)
d.roundDuration(smallestUnit: Time.Minute{}, increment: 60)
// panic: Time.roundDuration: the increment for minutes must divide 60 and be smaller, not 60
```

An increment of days or calendar units works only when `largestUnit` is `smallestUnit`:

```ant
let jan = Time.plainDate(2026, 1, 1)
let tenDays = Time.duration(days: 10)
tenDays.roundDuration(smallestUnit: Time.Day{}, increment: 7, largestUnit: Time.Week{}, relativeTo: jan)
// panic: Time.roundDuration: a date unit with an increment above 1 needs the largest unit to be the smallest
```

Balancing a long span into a small unit can overflow a field:

```ant
let jan = Time.plainDate(2026, 1, 1)
Time.duration(years: 300).roundDuration(largestUnit: Time.Nanosecond{}, relativeTo: jan)
// panic: Time.roundDuration: a balanced field does not fit an int
```

## See also

[Time.round](round.md), [Time.total](total.md), [Time.compareDuration](compareDuration.md),
[Time.addDuration](addDuration.md), [Time.until](until.md), [Time.duration](duration.md),
[Time.Unit](types/Unit.md), [Time.RoundingMode](types/RoundingMode.md),
[Time.Duration](types/Duration.md)
