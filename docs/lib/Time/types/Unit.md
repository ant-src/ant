<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.Unit

```ant signature
type Time.Unit = Time.Year | Time.Month | Time.Week | Time.Day | Time.Hour | Time.Minute | Time.Second | Time.Millisecond | Time.Microsecond | Time.Nanosecond
```

Time.Unit names a field of a duration, from years down to nanoseconds. Years, months, weeks and days
are calendar units that move the wall clock; hours and smaller are exact units that move the
instant.

```ant signature
type Time.Year = {}
```

Time.Year is the unit of a calendar year, which has no fixed length.

```ant signature
type Time.Month = {}
```

Time.Month is the unit of a calendar month, which has no fixed length.

```ant signature
type Time.Week = {}
```

Time.Week is the unit of a calendar week, a calendar unit like Time.Year and Time.Month.

```ant signature
type Time.Day = {}
```

Time.Day is the unit of a calendar day, 24 hours except in a zone on the day of a daylight-saving
change.

```ant signature
type Time.Hour = {}
```

Time.Hour is the unit of an hour, an exact unit.

```ant signature
type Time.Minute = {}
```

Time.Minute is the unit of a minute, an exact unit.

```ant signature
type Time.Second = {}
```

Time.Second is the unit of a second, an exact unit.

```ant signature
type Time.Millisecond = {}
```

Time.Millisecond is the unit of a millisecond, an exact unit.

```ant signature
type Time.Microsecond = {}
```

Time.Microsecond is the unit of a microsecond, an exact unit.

```ant signature
type Time.Nanosecond = {}
```

Time.Nanosecond is the unit of a nanosecond, the smallest unit.

<!-- end generated -->

## Details

- Each unit is an empty record, so you write it as a literal with braces: `Time.Hour{}`. Two values
  of the same unit are `==`.
- Years, months and weeks have no fixed length: `Time.total` and `Time.roundDuration` panic on them
  without a `relativeTo` date.
- An `Instant` has no calendar, so `Time.until` and `Time.round` panic on a day or larger for one; a
  `PlainTime` has no units above hours and a `PlainDate` none below days.

## Basic examples

Pass a unit to read one field of a duration, or to measure the whole duration:

```ant
let d = 1h30m
Time.part(d, Time.Minute{})  // 30
Time.total(d, Time.Minute{}) // 90.0
```

Choose the largest unit of a difference, or the unit to round to:

```ant
let start = Time.plainDate(2026, 1, 5)
let end = Time.plainDate(2026, 3, 20)
Time.until(start, end)                           // P74D
Time.until(start, end, largestUnit: Time.Week{}) // P10W4D
Time.round(Time.plainTime(9, 47), Time.Hour{})   // 10:00:00
```

Balance a duration up to a unit, or round it to one:

```ant
Time.roundDuration(90m, largestUnit: Time.Hour{})    // PT1H30M
Time.roundDuration(1h47m, smallestUnit: Time.Hour{}) // PT2H
```

## Scope

All ten units, from the largest to the smallest. Annotate the array so that it holds any unit:

```ant
let units: Array[Time.Unit] = [
  Time.Year{},
  Time.Month{},
  Time.Week{},
  Time.Day{},
  Time.Hour{},
  Time.Minute{},
  Time.Second{},
  Time.Millisecond{},
  Time.Microsecond{},
  Time.Nanosecond{},
]
let d = Time.parseDuration("P1Y2M3W4DT5H6M7.008009010S") | Result.unwrap()
units.map({ d.part(x) }) // [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
```

A unit prints as its literal:

```ant
Time.Hour{}                 // Time.Hour{}
String.from(Time.Minute{})  // "Time.Minute{}"
[Time.Day{}, Time.Second{}] // [Time.Day{}, Time.Second{}]
```

Store a unit in a variable of type `Time.Unit`, and test which one it is with `is` or `==`:

```ant
let u: Time.Unit = Time.Day{}
u is Time.Day    // true
u is Time.Hour   // false
u == Time.Day{}  // true
u == Time.Hour{} // false
```

Group the units with `is`:

```ant
func isCalendar(u: Time.Unit): bool {
  u is Time.Year || u is Time.Month || u is Time.Week || u is Time.Day
}

isCalendar(Time.Week{})   // true
isCalendar(Time.Second{}) // false
```

Units serve as map keys:

```ant
let short: Map[Time.Unit, string] = Map{ Time.Hour{}: "h", Time.Minute{}: "min" }
short[Time.Minute{}] // "min"
short[Time.Day{}]    // null
```

## Applications

Count the gap between two dates in a unit the caller picks. `largestUnit` balances the difference up
to that unit, and `part` reads its field:

```ant
func whole(from: Time.PlainDate, to: Time.PlainDate, unit: Time.Unit): int {
  Time.until(from, to, largestUnit: unit).part(unit)
}

let hired = Time.plainDate(2024, 9, 16)
let today = Time.plainDate(2026, 11, 11)
whole(hired, today, Time.Year{})  // 2
whole(hired, today, Time.Month{}) // 25
whole(hired, today, Time.Week{})  // 112
```

Keep a unit in a setting with a default, here the precision of log timestamps:

```ant
type LogConfig = { precision: Time.Unit = Time.Second{} }

func stamp(t: Time.PlainTime, config: LogConfig): string {
  "${t.round(config.precision, mode: Time.Floor{})}"
}

let t = Time.plainTime(14, 7, 42, 615)
stamp(t, LogConfig{})                           // "14:07:42"
stamp(t, LogConfig{ precision: Time.Minute{} }) // "14:07:00"
```

## Properties and relations

The smaller the unit, the larger the total of the same duration:

```ant
let exact: Array[Time.Unit] = [Time.Hour{}, Time.Minute{}, Time.Second{}, Time.Millisecond{}]
let totals = exact.map({ 2h.total(x) })
assert(totals == [2.0, 120.0, 7200.0, 7200000.0])
```

A day is 24 hours without a `relativeTo` point:

```ant
let d = Time.duration(days: 1)
assert(d.total(Time.Hour{}) == 24.0)
assert(36h.total(Time.Day{}) == 1.5)
```

## Possible issues

A unit is a value, written with braces. The bare name is a type:

```ant
let u: Time.Unit = Time.Hour
// error: Time.Hour is a type, not a value; a type appears only as a literal head, after is, or before a dot
```

`Time.Unit` is a union, not a record, so it has no literal:

```ant
let u = Time.Unit{} // error: Time.Unit is not a record type, so it has no literal
```

A unit is not a string:

```ant
let u: Time.Unit = "hour" // error: cannot use string where Time.Unit is expected
```

An array literal of units has the union of the units it lists as its element type, and array types
are invariant, so it does not fit `Array[Time.Unit]`. Annotate the array where you declare it:

```ant
let units = [Time.Hour{}, Time.Day{}]
let all: Array[Time.Unit] = units
// error: cannot use Array[Time.Hour | Time.Day] where Array[Time.Unit] is expected
```

Two different units have different types, so comparing them directly is a compile error. Compare
through a `Time.Unit` variable, or use `is`:

```ant
let same = Time.Hour{} == Time.Day{}
// error: Time.Hour and Time.Day cannot be compared with ==: neither is assignable to the other
```

A week has no fixed length, so a total in weeks needs a `relativeTo` date:

```ant
Time.duration(days: 14).total(Time.Week{})
// panic: Time.total: a total in weeks needs a relativeTo point
```

```ant
Time.duration(days: 14).total(Time.Week{}, relativeTo: Time.plainDate(2026, 1, 1)) // 2.0
```

An `Instant` has no days, so its difference stops at hours:

```ant
let start = Time.fromEpochMillis(0)
let end = Time.fromEpochMillis(90061000)
Time.until(start, end, largestUnit: Time.Hour{}) // PT25H1M1S
Time.until(start, end, largestUnit: Time.Day{})  // panic: Time.until: an instant has no days
```

## See also

[Time.part](../part.md), [Time.total](../total.md), [Time.until](../until.md),
[Time.since](../since.md), [Time.round](../round.md), [Time.roundDuration](../roundDuration.md),
[Time.Duration](Duration.md), [Time.RoundingMode](RoundingMode.md), [Time](../index.md)
