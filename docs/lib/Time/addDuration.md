<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.addDuration

```ant signature
func Time.addDuration(a: Time.Duration, b: Time.Duration): Time.Duration
```

Time.addDuration returns the sum of a and b, with days as 24 hours, balanced up to the larger of
their largest units. It panics if either has years, months or weeks.

<!-- end generated -->

## Details

- Carries into the largest unit either input has: hours never become days unless an input has days,
  and minutes become hours only when an input has hours.
- The inputs may have opposite signs; the result has one sign.
- Panics if either duration has years, months or weeks. Add them to a date with [`Time.add`](add.md)
  instead.

## Basic examples

Add two durations in the same unit. The sum stays in that unit:

```ant
Time.addDuration(Time.duration(minutes: 90), Time.duration(minutes: 30)) // PT120M
```

A sum with days carries 24 hours into a day:

```ant
Time.addDuration(Time.duration(days: 1), Time.duration(hours: 25)) // P2DT1H
```

## Scope

The largest unit of either input sets how far the sum balances. Minutes carry into hours when one
input has hours, and seconds carry into minutes when one has minutes:

```ant
Time.addDuration(Time.duration(hours: 1, minutes: 30), Time.duration(minutes: 45)) // PT2H15M
Time.addDuration(Time.duration(minutes: 1), Time.duration(seconds: 90))            // PT2M30S
Time.addDuration(Time.duration(seconds: 59), Time.duration(milliseconds: 1500))    // PT60.5S
```

Hours never become days on their own. Only an input with days makes 24 hours a day:

```ant
Time.addDuration(Time.duration(hours: 24), Time.duration(hours: 1)) // PT25H
Time.addDuration(Time.duration(days: 1), Time.duration(hours: 1))   // P1DT1H
```

A negative duration subtracts. The result takes the sign of the longer input. In the last line the
day becomes 24 hours before the subtraction, so no day is left in the result:

```ant
Time.addDuration(Time.duration(hours: 1), Time.duration(minutes: -30))     // PT30M
Time.addDuration(Time.duration(hours: 1), Time.duration(minutes: -90))     // -PT30M
Time.addDuration(Time.duration(seconds: -30), Time.duration(seconds: -45)) // -PT75S
Time.addDuration(Time.duration(days: 1), Time.duration(hours: -1))         // PT23H
```

Adding the zero duration, `Time.duration()`, keeps the length of the other input and balances it:

```ant
Time.addDuration(Time.duration(minutes: 25), Time.duration())           // PT25M
Time.addDuration(Time.duration(hours: 1, minutes: 90), Time.duration()) // PT2H30M
```

Call it through `.`, in a pipe, or with named arguments:

```ant
Time.duration(minutes: 90).addDuration(Time.duration(minutes: 30))          // PT120M
Time.duration(hours: 1) | Time.addDuration(Time.duration(minutes: 15))      // PT1H15M
Time.addDuration(b: Time.duration(minutes: 15), a: Time.duration(hours: 1)) // PT1H15M
```

## Applications

Total an array of durations with `reduce`, starting from the zero duration. The total stays in
hours, since no input has days:

```ant
let shifts = [
  Time.duration(hours: 7, minutes: 30),
  Time.duration(hours: 8),
  Time.duration(hours: 6, minutes: 45),
]
shifts.reduce(Time.duration(), Time.addDuration) // PT22H15M
```

Find when a meeting ends from the length of its agenda items:

```ant
type Item = { topic: string, length: Time.Duration }
let agenda = [
  Item{ topic: "Status", length: Time.duration(minutes: 15) },
  Item{ topic: "Design review", length: Time.duration(minutes: 50) },
  Item{ topic: "Planning", length: Time.duration(minutes: 25) },
]
let total = agenda.map({ x.length }).reduce(Time.duration(), Time.addDuration)
total                            // PT90M
Time.plainTime(9, 30).add(total) // 11:00:00
```

There is no subtraction of durations. Add the negation to find how much time is left:

```ant
let budget = Time.duration(hours: 2)
let spent = Time.duration(minutes: 75)
budget.addDuration(spent.negate()) // PT45M
```

## Properties and relations

The order of the inputs does not matter:

```ant
let a = Time.duration(hours: 1, minutes: 20)
let b = Time.duration(minutes: 55)
assert(Time.addDuration(a, b) == Time.addDuration(b, a))
```

A duration plus its negation is zero:

```ant
let d = Time.duration(minutes: 90)
assert(Time.addDuration(d, Time.negate(d)).sign() == 0)
```

The sum is as long as the inputs together, in any unit:

```ant
let a = Time.duration(minutes: 50)
let b = Time.duration(minutes: 25)
let sum = Time.addDuration(a, b)
assert(sum.total(Time.Hour{}) == a.total(Time.Hour{}) + b.total(Time.Hour{}))
```

## Possible issues

Years, months and weeks have no fixed length, so `Time.addDuration` panics on them. Add them to a
date with [`Time.add`](add.md) instead:

```ant
Time.addDuration(Time.duration(weeks: 1), Time.duration(days: 1))
// panic: Time.addDuration: years, months and weeks cannot be added without a starting point; add each to a date
Time.plainDate(2026, 3, 1).add(Time.duration(weeks: 1, days: 1)) // 2026-03-09
```

`==` compares the fields as given, so a sum kept in minutes is not `==` to the same length in hours.
Compare lengths with [`Time.compareDuration`](compareDuration.md), or balance the sum with
[`Time.roundDuration`](roundDuration.md):

```ant
let sum = Time.addDuration(Time.duration(minutes: 90), Time.duration(minutes: 30))
sum == Time.duration(hours: 2)                     // false
Time.compareDuration(sum, Time.duration(hours: 2)) // 0
sum.roundDuration(largestUnit: Time.Hour{})        // PT2H
```

## See also

[Time.duration](duration.md), [Time.add](add.md), [Time.negate](negate.md),
[Time.compareDuration](compareDuration.md), [Time.roundDuration](roundDuration.md),
[Time.total](total.md), [Time.Duration](types/Duration.md)
