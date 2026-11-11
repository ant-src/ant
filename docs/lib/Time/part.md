<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.part

```ant signature
func Time.part(d: Time.Duration, unit: Time.Unit): int
```

Time.part returns the field of d for unit, as given and not balanced.

<!-- end generated -->

## Details

- `Time.part` reads one of the ten fields a duration keeps. It does not convert: `PT90M` has `0`
  hours and `90` minutes. Use `Time.total` for the length in a unit.
- A field not given when the duration was made is `0`.
- On a negative duration every nonzero field is negative.

## Basic examples

```ant
let d = Time.duration(hours: 2, minutes: 15)
Time.part(d, Time.Hour{})   // 2
Time.part(d, Time.Minute{}) // 15
Time.part(d, Time.Second{}) // 0
```

Fields are not balanced, so 90 minutes stay 90 minutes:

```ant
Time.part(Time.duration(minutes: 90), Time.Hour{})   // 0
Time.part(Time.duration(minutes: 90), Time.Minute{}) // 90
```

## Scope

Every unit from years down to nanoseconds names a field:

```ant
let d = Time.duration(
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
)
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
units.map({ d.part(x) }) // [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
```

A negative duration has negative fields:

```ant
let back = Time.duration(days: -2, hours: -3)
back.part(Time.Day{})  // -2
back.part(Time.Hour{}) // -3
```

Call it with a dot on a duration that another function returns:

```ant
let shift = Time.until(Time.plainTime(9, 0, 0), Time.plainTime(17, 45, 0))
shift                     // PT8H45M
shift.part(Time.Hour{})   // 8
shift.part(Time.Minute{}) // 45
```

## Named arguments

```ant
Time.part(unit: Time.Minute{}, d: Time.duration(minutes: 90)) // 90
```

## Applications

Format an elapsed time as a clock reading. `Time.roundDuration` balances the seconds into hours,
minutes and seconds first, and `part` reads them:

```ant
func clock(d: Time.Duration): string {
  let b = Time.roundDuration(d, largestUnit: Time.Hour{})
  let mm = "${b.part(Time.Minute{})}".padStart(2, "0")
  let ss = "${b.part(Time.Second{})}".padStart(2, "0")
  "${b.part(Time.Hour{})}:${mm}:${ss}"
}

clock(Time.duration(seconds: 3725)) // "1:02:05"
clock(Time.duration(minutes: 90))   // "1:30:00"
```

Say how long ago something happened in the largest unit that is not zero. A duration between two
instants is all seconds, so `Time.roundDuration` balances it into days, hours and minutes first:

```ant
type Step = { unit: Time.Unit, name: string }

let steps = [
  Step{ unit: Time.Day{}, name: "day" },
  Step{ unit: Time.Hour{}, name: "hour" },
  Step{ unit: Time.Minute{}, name: "minute" },
]

func ago(d: Time.Duration): string {
  let b = Time.roundDuration(d, largestUnit: Time.Day{})
  let step = steps.find({ b.part(x.unit) != 0 })
  if step == null { return "just now" }
  let n = b.part(step.unit)
  let plural = if n == 1 { "" } else { "s" }
  "${n} ${step.name}${plural} ago"
}

let then = Time.parseInstant("2026-10-22T11:40:00Z") | Result.unwrap()
let now = Time.parseInstant("2026-10-22T15:00:00Z") | Result.unwrap()

ago(Time.since(now, then))            // "3 hours ago"
ago(Time.duration(days: 3, hours: 4)) // "3 days ago"
ago(Time.duration(hours: 1))          // "1 hour ago"
ago(Time.duration(seconds: 20))       // "just now"
```

## Properties and relations

`Time.negate` negates every field, and `Time.abs` makes every field nonnegative:

```ant
let d = Time.duration(days: -1, hours: -6, minutes: -30)
let units: Array[Time.Unit] = [Time.Day{}, Time.Hour{}, Time.Minute{}, Time.Second{}]
for u in units {
  assert(Time.negate(d).part(u) == -d.part(u))
  assert(Time.abs(d).part(u) >= 0)
}
```

Two durations are `==` exactly when every field is the same, so `part` tells apart durations of the
same length:

```ant
let a = Time.duration(minutes: 90)
let b = Time.duration(hours: 1, minutes: 30)
Time.compareDuration(a, b)                 // 0
a == b                                     // false
a.part(Time.Hour{}) == b.part(Time.Hour{}) // false
```

## Possible issues

`part` is a field, not a length. A duration between two instants is in seconds, so its hours field
is `0`. Use `Time.total` for the length, or pass `largestUnit:` to get the field you want:

```ant
let start = Time.parseInstant("2026-10-22T14:00:00Z") | Result.unwrap()
let end = Time.parseInstant("2026-10-22T16:30:00Z") | Result.unwrap()
let d = Time.until(start, end)
d                                                                  // PT9000S
d.part(Time.Hour{})                                                // 0
Time.total(d, Time.Hour{})                                         // 2.5
Time.until(start, end, largestUnit: Time.Hour{}).part(Time.Hour{}) // 2
```

Two durations that print the same can hold different fields. `Time.duration(milliseconds: 1500)`
keeps 1500 milliseconds, while the other keeps 1 second and 500 milliseconds:

```ant
let a = Time.duration(milliseconds: 1500)
let b = Time.duration(seconds: 1, milliseconds: 500)
a                     // PT1.5S
b                     // PT1.5S
a.part(Time.Second{}) // 0
b.part(Time.Second{}) // 1
```

The unit is a value, written with braces. The bare name is a type:

```ant
let h = Time.part(Time.duration(hours: 1), Time.Hour)
// error: Time.Hour is a type, not a value; a type appears only as a literal head, after is, or before a dot
```

## See also

[Time.total](total.md), [Time.roundDuration](roundDuration.md), [Time.duration](duration.md),
[Time.parseDuration](parseDuration.md), [Time.sign](sign.md), [Time.until](until.md),
[Time.Unit](types/Unit.md), [Time.Duration](types/Duration.md)
