<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.hypot

```ant signature
func Math.hypot(x: float64, y: float64): float64
```

Math.hypot returns the square root of x² + y², without overflow in the intermediate steps.

<!-- end generated -->

## Details

- The result is never negative: the signs of `x` and `y` do not matter.
- `Math.hypot` never panics. A result too large for a `float64` is `Infinity`.

## Basic examples

```ant
Math.hypot(3.0, 4.0)  // 5.0
Math.hypot(5.0, 12.0) // 13.0
Math.hypot(1.0, 1.0)  // 1.4142135623730951
```

## Scope

Negative arguments give the same result as positive ones:

```ant
Math.hypot(-3.0, 4.0)  // 5.0
Math.hypot(-3.0, -4.0) // 5.0
```

With one argument zero, the result is the absolute value of the other:

```ant
Math.hypot(7.0, 0.0)  // 7.0
Math.hypot(-7.0, 0.0) // 7.0
Math.hypot(0.0, 0.0)  // 0.0
```

Very large and very small arguments, where squaring alone would overflow or underflow:

```ant
Math.hypot(1e200, 1e200)   // 1.414213562373095e+200
Math.hypot(1e-200, 1e-200) // 1.414213562373095e-200
```

Infinities and `NaN`:

```ant
Math.hypot(Math.Infinity, 2.0)      // Infinity
Math.hypot(-Math.Infinity, 2.0)     // Infinity
Math.hypot(Math.NaN, 2.0)           // NaN
Math.hypot(Math.Infinity, Math.NaN) // Infinity
```

A result past the largest `float64` is `Infinity`:

```ant
Math.hypot(Math.maxFloat64, Math.maxFloat64) // Infinity
```

Name the arguments:

```ant
Math.hypot(x: 6.0, y: 8.0) // 10.0
```

In a pipe, the value is `x`:

```ant
3.0 | Math.hypot(4.0) // 5.0
```

## Applications

The distance between two points:

```ant
type Point = { x: float64, y: float64 }

func distance(a: Point, b: Point): float64 {
  Math.hypot(b.x - a.x, b.y - a.y)
}

distance(Point{ x: 1.0, y: 1.0 }, Point{ x: 4.0, y: 5.0 }) // 5.0
```

The length of a path through several points:

```ant
type Point = { x: float64, y: float64 }

let path = [Point{ x: 0.0, y: 0.0 }, Point{ x: 3.0, y: 4.0 }, Point{ x: 3.0, y: 10.0 }]
var total = 0.0
for i in Iter.range(1, path.length()) {
  total += Math.hypot(path[i].x - path[i - 1].x, path[i].y - path[i - 1].y)
}
total // 11.0
```

The length of a vector in three dimensions, by nesting calls:

```ant
Math.hypot(Math.hypot(1.0, 2.0), 2.0) // 3.0
```

The point nearest to the origin:

```ant
type Point = { x: float64, y: float64 }

let points = [Point{ x: 3.0, y: 4.0 }, Point{ x: -1.0, y: 2.0 }, Point{ x: 0.5, y: -2.5 }]
points.minBy({ Math.hypot(x.x, x.y) }) // Point{ x: -1.0, y: 2.0 }
```

Polar coordinates, with [Math.atan2](atan2.md) for the angle:

```ant
let x = 1.0
let y = 1.0
Math.hypot(x, y) // 1.4142135623730951
Math.atan2(y, x) // 0.7853981633974483
```

## Properties and relations

`Math.hypot` is symmetric and ignores signs:

```ant
for [x, y] in [[3.0, 4.0], [-1.5, 2.0], [0.0, -7.0]] {
  assert(Math.hypot(x, y) == Math.hypot(y, x))
  assert(Math.hypot(x, y) == Math.hypot(-x, -y))
}
```

With `y` zero it equals [Math.abs](abs.md):

```ant
for x in [2.5, -2.5, 0.0, 1e300] {
  assert(Math.hypot(x, 0.0) == Math.abs(x))
}
```

On a Pythagorean triple the result is exact:

```ant
for [a, b, c] in [[3.0, 4.0, 5.0], [5.0, 12.0, 13.0], [8.0, 15.0, 17.0], [20.0, 21.0, 29.0]] {
  assert(Math.hypot(a, b) == c)
}
```

## Possible issues

The written-out formula overflows where `Math.hypot` does not:

```ant
let x = 1e200
Math.sqrt(x * x + x * x) // Infinity
Math.hypot(x, x)         // 1.414213562373095e+200
```

It underflows to zero for small arguments too:

```ant
let x = 1e-200
Math.sqrt(x * x + x * x) // 0.0
Math.hypot(x, x)         // 1.414213562373095e-200
```

The arguments are `float64`. An `int` is a compile error; convert it first:

```ant
let dx = 3
let dy = 4
let d = Math.hypot(dx, dy)
// error: cannot use int where float64 is expected
```

```ant
let dx = 3
let dy = 4
Math.hypot(float64(dx), float64(dy)) // 5.0
```

A number has no namespace, so a call through `.` is a compile error. Write `Math.hypot(x, y)`, or
use a pipe:

```ant
let x = 3.0
let d = x.hypot(4.0) // error: float64 has no field or function hypot
```

## See also

[Math.sqrt](sqrt.md) · [Math.abs](abs.md) · [Math.atan2](atan2.md) · [Math.pow](pow.md) ·
[Math.cbrt](cbrt.md) · [Math namespace](index.md)
