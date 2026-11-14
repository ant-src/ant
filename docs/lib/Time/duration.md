<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.duration

```ant signature
func Time.duration(years: int = 0, months: int = 0, weeks: int = 0, days: int = 0, hours: int = 0, minutes: int = 0, seconds: int = 0, milliseconds: int = 0, microseconds: int = 0, nanoseconds: int = 0): Time.Duration
```

Time.duration returns a duration with the given fields, each 0 when omitted and kept as given. It
panics if the fields have mixed signs, if years, months, weeks or days reach 2^32 in magnitude, or
if the time part, days counted as 24 hours, reaches 2^53 seconds. A constant span of hours and
smaller units is written as a duration literal instead, such as 2h30m or 500ms.

<!-- end generated -->

## Details

- Each field keeps the value you pass: `Time.duration(minutes: 90)` stays `PT90M` until
  [Time.roundDuration](roundDuration.md) balances it.

## Basic examples

An hour and a half, with half a second more:

```ant
Time.duration(hours: 1, minutes: 30, milliseconds: 500) // PT1H30M0.5S
```

Three days, and no time at all:

```ant
Time.duration(days: 3) // P3D
Time.duration()        // PT0S
```

## Scope

Every field at once. The subsecond fields fold into the seconds when the duration renders:

```ant
Time.duration(
  years: 1,
  months: 2,
  weeks: 3,
  days: 4,
  hours: 5,
  minutes: 6,
  seconds: 7,
  milliseconds: 8,
  microseconds: 9,
  nanoseconds: 10,
) // P1Y2M3W4DT5H6M7.00800901S
```

A negative duration has every nonzero field negative, and renders with a sign in front:

```ant
Time.duration(minutes: -15)            // -PT15M
Time.duration(hours: -1, minutes: -30) // -PT1H30M
```

A field is not balanced into a larger one:

```ant
Time.duration(minutes: 90)        // PT90M
Time.duration(milliseconds: 1500) // PT1.5S
Time.duration(hours: 48)          // PT48H
```

Zero fields are the same as omitted ones:

```ant
Time.duration(minutes: 0, seconds: 0) // PT0S
```

## Named arguments

Each field has its own named argument, from `years:` down to `nanoseconds:`. Name the fields you
need, in any order:

```ant
Time.duration(seconds: 30, minutes: 2) // PT2M30S
Time.duration(weeks: 2)                // P2W
Time.duration(microseconds: 250)       // PT0.00025S
```

## Applications

Move a meeting by an hour and a quarter:

```ant
let meeting = Time.plainDateTime(Time.plainDate(2026, 3, 9), Time.plainTime(9, 30))
meeting.add(Time.duration(hours: 1, minutes: 15)) // 2026-03-09T10:45:00
```

Add a month to a date. The day is constrained to the end of February:

```ant
Time.plainDate(2024, 1, 31) | Time.add(Time.duration(months: 1)) // 2024-02-29
```

Turn a list of minutes into one total duration:

```ant
let laps = [15, 45, 30]
laps.map({ Time.duration(minutes: x) }).reduce(Time.duration(), Time.addDuration) // PT90M
```

Wait between attempts, longer each time:

```ant
for attempt in Array.range(1, 4) {
  println("attempt ${attempt}")
  Time.sleep(Time.duration(milliseconds: attempt * 10))
}
```

```text
attempt 1
attempt 2
attempt 3
```

Show a number of seconds as hours, minutes and seconds:

```ant
Time.duration(seconds: 3725) | Time.roundDuration(largestUnit: Time.Hour{}) // PT1H2M5S
```

## Properties and relations

[Time.part](part.md) reads a field back as it was given:

```ant
let d = Time.duration(hours: 2, minutes: 90)
assert(d.part(Time.Hour{}) == 2)
assert(d.part(Time.Minute{}) == 90)
```

A duration renders in ISO 8601 form, and [Time.parseDuration](parseDuration.md) reads that form
back:

```ant
let d = Time.duration(days: 1, hours: 2, seconds: 3)
assert(Time.parseDuration("${d}").unwrap() == d)
```

A negative duration is the [negation](negate.md) of the positive one:

```ant
assert(Time.duration(minutes: -5) == Time.duration(minutes: 5).negate())
assert(Time.duration(minutes: -5).abs() == Time.duration(minutes: 5))
```

## Possible issues

Fields of different signs panic. Write a negative duration with every field negative:

```ant
Time.duration(hours: 1, minutes: -30)
// panic: Time.duration: a duration's fields must all have one sign
Time.duration(hours: -1, minutes: -30) // -PT1H30M
```

The parameters run from `years` down, so the first positional argument is years. Name the field:

```ant
Time.duration(5)          // P5Y
Time.duration(minutes: 5) // PT5M
```

`==` compares the fields as given, so two durations of the same length can differ. Use
[Time.compareDuration](compareDuration.md) to compare lengths:

```ant
Time.duration(minutes: 90) == Time.duration(hours: 1, minutes: 30)                     // false
Time.compareDuration(Time.duration(minutes: 90), Time.duration(hours: 1, minutes: 30)) // 0
```

A day is a calendar unit, so a duration with days does not add to an [Instant](types/Instant.md);
use hours:

```ant
let epoch = Time.fromEpochMillis(0)
epoch.add(Time.duration(days: 1))
// panic: Time.add: an instant has no calendar, so it takes no days; convert it with Time.inZone first
epoch.add(Time.duration(hours: 24)) // 1970-01-02T00:00:00Z
```

Values past the bounds panic:

```ant
Time.duration(days: 4294967295) // P4294967295D
Time.duration(days: 4294967296)
// panic: Time.duration: years, months, weeks and days must be below 2^32
Time.duration(seconds: 9007199254740992)
// panic: Time.duration: the time part of a duration must be below 2^53 seconds
```

## See also

[Time.parseDuration](parseDuration.md), [Time.Duration](types/Duration.md), [Time.part](part.md),
[Time.addDuration](addDuration.md), [Time.roundDuration](roundDuration.md), [Time.total](total.md),
[Time.add](add.md), [Time.sleep](sleep.md)
