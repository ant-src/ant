<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.RoundingMode

```ant signature
type Time.RoundingMode = Time.Ceil | Time.Floor | Time.Expand | Time.Trunc | Time.HalfCeil | Time.HalfFloor | Time.HalfExpand | Time.HalfTrunc | Time.HalfEven
```

Time.RoundingMode says how [`Time.round`](../round.md) and
[`Time.roundDuration`](../roundDuration.md) round. The default is Time.HalfExpand.

```ant signature
type Time.Ceil = {}
```

Time.Ceil is the rounding mode toward positive infinity.

```ant signature
type Time.Floor = {}
```

Time.Floor is the rounding mode toward negative infinity.

```ant signature
type Time.Expand = {}
```

Time.Expand is the rounding mode away from zero.

```ant signature
type Time.Trunc = {}
```

Time.Trunc is the rounding mode toward zero.

```ant signature
type Time.HalfCeil = {}
```

Time.HalfCeil is the rounding mode to the nearest multiple, with a tie toward positive infinity.

```ant signature
type Time.HalfFloor = {}
```

Time.HalfFloor is the rounding mode to the nearest multiple, with a tie toward negative infinity.

```ant signature
type Time.HalfExpand = {}
```

Time.HalfExpand is the rounding mode to the nearest multiple, with a tie away from zero, as
Math.round does. It is the default.

```ant signature
type Time.HalfTrunc = {}
```

Time.HalfTrunc is the rounding mode to the nearest multiple, with a tie toward zero.

```ant signature
type Time.HalfEven = {}
```

Time.HalfEven is the rounding mode to the nearest multiple, with a tie to the even multiple.

<!-- end generated -->

## Details

- Each mode is an empty record, so you pass a value such as `Time.Floor{}`, not the type name.
- The modes treat a negative value as a number below zero: `Ceil` and `Trunc` move `-PT90M` toward
  zero, `Floor` and `Expand` away from it. An `Instant` before 1970 counts as negative.

## Basic examples

Pass a mode as the `mode:` argument of `Time.round` or `Time.roundDuration`:

```ant
let t = Time.plainTime(14, 37)
t.round(Time.Minute{}, increment: 15, mode: Time.Floor{}) // 14:30:00
t.round(Time.Minute{}, increment: 15, mode: Time.Ceil{})  // 14:45:00
```

Without `mode:`, both functions use `Time.HalfExpand{}`, which takes a tie away from zero:

```ant
let tie = 90m
tie.roundDuration(smallestUnit: Time.Hour{})                         // PT2H
tie.roundDuration(smallestUnit: Time.Hour{}, mode: Time.HalfEven{})  // PT2H
tie.roundDuration(smallestUnit: Time.Hour{}, mode: Time.HalfTrunc{}) // PT1H
```

## Scope

### All nine modes

Each duration below rounds to whole hours. The first is off the halfway point; the other three are
ties, one of them negative:

```ant
let modes: Array[Time.RoundingMode] = [
  Time.Ceil{},
  Time.Floor{},
  Time.Expand{},
  Time.Trunc{},
  Time.HalfCeil{},
  Time.HalfFloor{},
  Time.HalfExpand{},
  Time.HalfTrunc{},
  Time.HalfEven{},
]
let inputs = [100, 90, 150, -90].map({ Time.duration(minutes: x) })
for mode in modes {
  let row = inputs
    | Array.map({ "${x.roundDuration(smallestUnit: Time.Hour{}, mode: mode)}".padStart(6) })
    | String.join("")
  println("${mode}".padEnd(18) + row)
}
```

```text
Time.Ceil{}         PT2H  PT2H  PT3H -PT1H
Time.Floor{}        PT1H  PT1H  PT2H -PT2H
Time.Expand{}       PT2H  PT2H  PT3H -PT2H
Time.Trunc{}        PT1H  PT1H  PT2H -PT1H
Time.HalfCeil{}     PT2H  PT2H  PT3H -PT1H
Time.HalfFloor{}    PT2H  PT1H  PT2H -PT2H
Time.HalfExpand{}   PT2H  PT2H  PT3H -PT2H
Time.HalfTrunc{}    PT2H  PT1H  PT2H -PT1H
Time.HalfEven{}     PT2H  PT2H  PT2H -PT2H
```

The first four always go one way. The five `Half` modes go to the nearest multiple, so they agree on
100 minutes. They differ only on a tie: the four with a plain form settle it as that form would, and
`HalfEven` sends 90 and 150 minutes to 2 hours, the even multiple.

### A mode as a value

A mode is a value like any other: keep it in a variable, take it as a parameter, and test it with
`is`:

```ant
func describe(mode: Time.RoundingMode): string {
  if mode is Time.Ceil || mode is Time.Floor { "always one way" } else { "depends" }
}
let policy: Time.RoundingMode = Time.Floor{}
describe(policy)            // "always one way"
describe(Time.HalfExpand{}) // "depends"
policy is Time.Floor        // true
"${policy}"                 // "Time.Floor{}"
```

## Applications

Read a rounding policy from a setting, and reject an unknown name:

```ant
func parseMode(name: string): Result[Time.RoundingMode, Error] {
  if name == "up" {
    Result.ok(Time.Ceil{})
  } else if name == "down" {
    Result.ok(Time.Floor{})
  } else if name == "nearest" {
    Result.ok(Time.HalfExpand{})
  } else {
    Result.fail("unknown rounding policy: ${name}")
  }
}

func bill(worked: Time.Duration, policy: string): Result[Time.Duration, Error] {
  let mode = try parseMode(policy)
  Result.ok(worked.roundDuration(smallestUnit: Time.Minute{}, increment: 15, mode: mode))
}

let worked = 1h20m
bill(worked, "up")      // Ok{ value: PT1H30M }
bill(worked, "down")    // Ok{ value: PT1H15M }
bill(worked, "nearest") // Ok{ value: PT1H15M }
bill(worked, "weekly")  // Err{ error: Error{ message: "unknown rounding policy: weekly" } }
```

## Properties and relations

`HalfExpand`, the default, rounds as `Math.round` does, halves away from zero:

```ant
let half = Time.negate(2500ms)
let rounded = half.roundDuration(smallestUnit: Time.Second{})
assert(rounded.total(Time.Second{}) == Math.round(-2.5))
```

`Floor` and `Ceil` bracket the value, and every mode leaves a multiple where it is:

```ant
let t = Time.plainTime(8, 17)
let down = t.round(Time.Minute{}, increment: 10, mode: Time.Floor{})
let up = t.round(Time.Minute{}, increment: 10, mode: Time.Ceil{})
assert(Time.compare(down, t) <= 0 && Time.compare(t, up) <= 0)
let modes: Array[Time.RoundingMode] = [Time.Ceil{}, Time.Trunc{}, Time.HalfEven{}]
assert(modes.every({ down.round(Time.Minute{}, increment: 10, mode: x) == down }))
```

## Possible issues

A mode is a value, so it needs the braces of an empty record:

```ant
let t = Time.plainTime(14, 37).round(Time.Minute{}, mode: Time.Floor)
// error: Time.Floor is a type, not a value; a type appears only as a literal head, after is, or before a dot
```

Two different modes are different types, so `==` between their literals is a compile error. Test a
mode with `is` instead:

```ant
let same = Time.HalfEven{} == Time.HalfExpand{}
// error: Time.HalfEven and Time.HalfExpand cannot be compared with ==: neither is assignable to the other
```

`Trunc` goes toward zero, which is up for a negative value. Use `Floor` to always round down:

```ant
let d = Time.negate(100m)
d.roundDuration(smallestUnit: Time.Hour{}, mode: Time.Trunc{}) // -PT1H
d.roundDuration(smallestUnit: Time.Hour{}, mode: Time.Floor{}) // -PT2H
```

## See also

[Time.round](../round.md), [Time.roundDuration](../roundDuration.md), [Time.Unit](Unit.md),
[Math.round](../../Math/round.md), [Math.trunc](../../Math/trunc.md)
