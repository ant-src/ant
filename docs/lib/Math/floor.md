<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.floor

```ant signature
func Math.floor(x: float64): float64
```

Math.floor returns the greatest whole number not above x.

<!-- end generated -->

## Details

- It rounds toward negative infinity, so a negative fraction moves away from zero: `-2.7` gives
  `-3.0`.
- It never panics. `NaN`, the infinities, and `-0.0` come back unchanged.

## Basic examples

```ant
Math.floor(2.7)  // 2.0
Math.floor(-2.7) // -3.0
Math.floor(5.0)  // 5.0
```

Convert the result to get an `int`:

```ant
int(Math.floor(7.9)) // 7
```

## Scope

A number between `-1.0` and `0.0` goes down to `-1.0`, and `-0.0` keeps its sign:

```ant
Math.floor(-0.5) // -1.0
Math.floor(-0.0) // -0.0
```

`NaN` and the infinities come back as they are:

```ant
Math.floor(Math.NaN)       // NaN
Math.floor(Math.Infinity)  // Infinity
Math.floor(-Math.Infinity) // -Infinity
```

A float beyond 2^53 is already whole, so it comes back unchanged:

```ant
Math.floor(1e300) // 1e+300
```

The parameter is named `x`:

```ant
Math.floor(x: 1.5) // 1.0
```

In a pipe, the value is the argument:

```ant
3.99 | Math.floor() // 3.0
```

`Math.floor` is a function value, so it can be passed to `Array.map`:

```ant
[1.2, -1.2, 3.0].map(Math.floor) // [1.0, -2.0, 3.0]
```

## Applications

Find the grid cell a point falls in, for negative coordinates too:

```ant
type Point = { x: float64, y: float64 }

func cell(p: Point, size: float64): Point {
  Point{ x: Math.floor(p.x / size), y: Math.floor(p.y / size) }
}

cell(Point{ x: 25.0, y: 7.5 }, 10.0)   // Point{ x: 2.0, y: 0.0 }
cell(Point{ x: -2.5, y: -12.0 }, 10.0) // Point{ x: -1.0, y: -2.0 }
```

Sort temperatures into 5-degree bands, below zero too:

```ant
func band(celsius: float64): float64 {
  Math.floor(celsius / 5.0) * 5.0
}

[21.7, 3.2, -3.2, -10.0].map(band) // [20.0, 0.0, -5.0, -10.0]
```

## Properties and relations

`Math.floor(x)` is at most `x`, and `x` is less than one above it:

```ant
for x in [2.7, -2.7, 0.5, -0.5, 5.0] {
  assert(Math.floor(x) <= x && x < Math.floor(x) + 1.0)
}
```

For zero and above it agrees with [Math.trunc](trunc.md); below zero it does not, unless `x` is
whole:

```ant
assert(Math.floor(2.5) == Math.trunc(2.5))
assert(Math.floor(-2.5) != Math.trunc(-2.5))
```

It mirrors [Math.ceil](ceil.md) through negation:

```ant
for x in [2.5, -2.5, 3.0] {
  assert(Math.floor(-x) == -Math.ceil(x))
}
```

## Possible issues

The argument must be a `float64`. An `int` is a compile error, since numeric types never mix:

```ant
let n = 7
Math.floor(n) // error: cannot use int where float64 is expected
```

Convert it first:

```ant
let n = 7
Math.floor(float64(n)) // 7.0
```

The result is a `float64`, so it does not go where an `int` is expected:

```ant
let i: int = Math.floor(2.5) // error: cannot use float64 where int is expected
```

Numbers have no functions through `.`, so `x.floor()` is a compile error. Call `Math.floor(x)`:

```ant
let y = (3.99).floor() // error: float64 has no field or function floor
```

Integer `/` truncates toward zero, so it does not floor a negative quotient:

```ant
-7 / 2                        // -3
Math.floor(float64(-7) / 2.0) // -4.0
```

A product that looks whole may sit just below it in binary, and floors one lower:

```ant
4.35 * 100.0             // 434.99999999999994
Math.floor(4.35 * 100.0) // 434.0
```

A whole `float64` may still be too large for an `int`, so `int(...)` panics on it. Use
[Math.toInt](toInt.md) to get `null` instead:

```ant
int(Math.floor(1e20))        // panic: 100000000000000000000.0 out of range for int
Math.toInt(Math.floor(1e20)) // null
```

## See also

[Math.ceil](ceil.md) · [Math.round](round.md) · [Math.trunc](trunc.md) · [Math.toInt](toInt.md) ·
[Arithmetic operators](../../language/operators/arithmetic.md)
