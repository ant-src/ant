<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.abs

```ant signature
func Math.abs[T: Number](x: T): T
```

Math.abs returns the absolute value of x. It panics on the most negative value of an integer type,
whose absolute value does not fit.

<!-- end generated -->

## Details

- This holds for every signed integer type, such as `int8(-128)`.
- On floats, `NaN` stays `NaN`, both infinities give `Infinity`, and `-0.0` gives `0.0`.

## Basic examples

```ant
Math.abs(-5)   // 5
Math.abs(5)    // 5
Math.abs(-2.5) // 2.5
```

## Scope

Every integer and float type works, and the result keeps its type:

```ant
Math.abs(int8(-100))    // 100
Math.abs(uint8(200))    // 200
Math.abs(float32(-1.5)) // 1.5
Math.abs(0)             // 0
```

Special float values:

```ant
Math.abs(-Math.Infinity) // Infinity
Math.abs(Math.NaN)       // NaN
Math.abs(-0.0)           // 0.0
```

Pass `Math.abs` as a function value:

```ant
[3, -1, -7].map(Math.abs) // [3, 1, 7]
```

## Named arguments

```ant
Math.abs(x: -3) // 3
```

## Applications

The distance between two numbers on a line:

```ant
func distance(a: float64, b: float64): float64 {
  Math.abs(a - b)
}

distance(2.5, 7.0) // 4.5
distance(7.0, 2.5) // 4.5
```

Compare floats within a tolerance instead of with `==`:

```ant
func close(a: float64, b: float64): bool {
  Math.abs(a - b) < 1e-9
}

0.1 + 0.2 == 0.3      // false
close(0.1 + 0.2, 0.3) // true
```

Find the reading nearest a target:

```ant
let readings = [12.5, 9.0, 14.25, 11.75]
let target = 12.0
readings.minBy({ Math.abs(x - target) }) // 11.75
```

## Properties and relations

The absolute value is never negative, and is the larger of `x` and `-x`:

```ant
let xs = [-7, 0, 4, Math.maxInt]
for x in xs {
  assert(Math.abs(x) >= 0)
  assert(Math.abs(x) == Math.max(x, -x))
  assert(Math.abs(-x) == Math.abs(x))
}
```

The absolute value of a product is the product of the absolute values, while neither overflows:

```ant
let a = -3
let b = 4
assert(Math.abs(a * b) == Math.abs(a) * Math.abs(b))
```

## Possible issues

The most negative value of a signed integer type has no positive counterpart, so `Math.abs` panics
on it:

```ant
Math.abs(Math.minInt)     // panic: Math.abs: integer overflow
Math.abs(int8(-128))      // panic: Math.abs: integer overflow
Math.abs(Math.minInt + 1) // 9223372036854775807
```

An unsigned difference panics before `Math.abs` sees it; subtract the smaller from the larger
instead:

```ant
Math.abs(uint8(3) - uint8(5)) // panic: integer overflow
```

```ant
Math.max(uint8(3), uint8(5)) - Math.min(uint8(3), uint8(5)) // 2
```

`Math.abs` takes a single numeric kind. A value typed `Number` must be narrowed first:

```ant
let n: Number = -3
Math.abs(n)
// error: T would be Number, which is not a single member of its bound Number; narrow first
```

A named primitive type is not a `Number`. Unwrap it, then tag the result again:

```ant
type Meters = float64
let drift = Meters(-2.5)
let bad = Math.abs(drift)
// error: T would be Meters, which is not a single member of its bound Number; narrow first
```

```ant
type Meters = float64
let drift = Meters(-2.5)
Meters(Math.abs(float64(drift))) // Meters(2.5)
```

## See also

[Math.min](min.md), [Math.max](max.md), [Math.clamp](clamp.md), [Math.trunc](trunc.md),
[Time.abs](../Time/abs.md)
