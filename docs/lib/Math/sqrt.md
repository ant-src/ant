<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.sqrt

```ant signature
func Math.sqrt(x: float64): float64
```

Math.sqrt returns the square root of x. It is NaN for a negative x.

<!-- end generated -->

## Details

- A negative `x` gives `NaN`; `Math.sqrt` never panics. Check the input first, or the result with
  [Math.isNaN](isNaN.md).

## Basic examples

```ant
Math.sqrt(16.0) // 4.0
Math.sqrt(2.0)  // 1.4142135623730951
Math.sqrt(0.25) // 0.5
```

A negative number has no real square root:

```ant
Math.sqrt(-1.0) // NaN
```

## Scope

Zero keeps its sign:

```ant
Math.sqrt(0.0)  // 0.0
Math.sqrt(-0.0) // -0.0
```

Infinities and `NaN`:

```ant
Math.sqrt(Math.Infinity)  // Infinity
Math.sqrt(-Math.Infinity) // NaN
Math.sqrt(Math.NaN)       // NaN
```

Very large and very small values:

```ant
Math.sqrt(Math.maxFloat64) // 1.3407807929942596e+154
Math.sqrt(1e-300)          // 1e-150
```

An integer goes through `float64` first:

```ant
let n = 49
Math.sqrt(float64(n)) // 7.0
```

Pass `Math.sqrt` as a function value, or reach a field of a record in a lambda:

```ant
[1.0, 4.0, 9.0].map(Math.sqrt) // [1.0, 2.0, 3.0]
type Square = { area: float64 }
let squares = [Square{ area: 25.0 }, Square{ area: 2.25 }]
squares.map({ Math.sqrt(x.area) }) // [5.0, 1.5]
```

In a pipe, the value is the argument:

```ant
81.0 | Math.sqrt() // 9.0
```

The parameter is named `x`:

```ant
Math.sqrt(x: 100.0) // 10.0
```

## Applications

The distance between two points:

```ant
type Point = { x: float64, y: float64 }

func distance(a: Point, b: Point): float64 {
  let dx = b.x - a.x
  let dy = b.y - a.y
  Math.sqrt(dx * dx + dy * dy)
}

distance(Point{ x: 1.0, y: 1.0 }, Point{ x: 4.0, y: 5.0 }) // 5.0
```

The standard deviation of a list of numbers:

```ant
func stddev(xs: Array[float64]): float64 {
  let n = float64(xs.length())
  let mean = (xs.sum() ?? 0.0) / n
  var squares = 0.0
  for x in xs {
    squares += (x - mean) * (x - mean)
  }
  Math.sqrt(squares / n)
}

stddev([2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0]) // 2.0
```

The real roots of a quadratic `ax² + bx + c`, or `null` when there are none:

```ant
func roots(a: float64, b: float64, c: float64): Array[float64] | null {
  let d = b * b - 4.0 * a * c
  if d < 0.0 {
    return null
  }
  let s = Math.sqrt(d)
  [(-b - s) / (2.0 * a), (-b + s) / (2.0 * a)]
}

roots(1.0, -3.0, 2.0) // [1.0, 2.0]
roots(1.0, 0.0, 1.0)  // null
```

Whether an integer is a perfect square:

```ant
func isSquare(n: int): bool {
  if n < 0 {
    return false
  }
  let r = int(Math.round(Math.sqrt(float64(n))))
  r * r == n
}

[16, 17, 144, -4].map(isSquare) // [true, false, true, false]
```

## Properties and relations

The square of the root of a perfect square is the number back:

```ant
for n in [0.0, 1.0, 4.0, 9.0, 1e10, 0.0625] {
  let r = Math.sqrt(n)
  assert(r * r == n)
}
```

The root of `x²` is `Math.abs(x)` while `x * x` stays finite:

```ant
for x in [-3.0, -0.5, 2.0, 10.0] {
  assert(Math.sqrt(x * x) == Math.abs(x))
}
```

The length of a vector is [Math.hypot](hypot.md) written out:

```ant
Math.sqrt(3.0 * 3.0 + 4.0 * 4.0) // 5.0
Math.hypot(3.0, 4.0)             // 5.0
```

## Possible issues

A `NaN` from a negative input is equal to nothing, itself included, so compare with
[Math.isNaN](isNaN.md):

```ant
let r = Math.sqrt(-4.0)
r == r        // false
Math.isNaN(r) // true
```

The result is rounded, so the square of an irrational root misses by a little:

```ant
Math.sqrt(2.0) * Math.sqrt(2.0) // 2.0000000000000004
```

`Math.pow(x, 0.5)` differs from `Math.sqrt` at the edges:

```ant
Math.sqrt(-0.0)               // -0.0
Math.pow(-0.0, 0.5)           // 0.0
Math.sqrt(-Math.Infinity)     // NaN
Math.pow(-Math.Infinity, 0.5) // Infinity
```

Squaring first overflows for large values, where [Math.hypot](hypot.md) does not:

```ant
let x = 1e200
Math.sqrt(x * x + x * x) // Infinity
Math.hypot(x, x)         // 1.414213562373095e+200
```

An `int` argument is a compile error, since numeric types never mix:

```ant
let side = Math.sqrt(16) // error: cannot use int where float64 is expected
```

Converting the result back to `int` panics when it has a fraction. Use [Math.toInt](toInt.md) to get
`null` instead, or round first:

```ant
int(Math.sqrt(16.0))             // 4
int(Math.sqrt(2.0))              // panic: 1.4142135623730951 has a fraction, so it does not convert to int
Math.toInt(Math.sqrt(2.0))       // null
int(Math.floor(Math.sqrt(10.0))) // 3
```

A number has no namespace, so a call through `.` is a compile error. Write `Math.sqrt(x)`, or use a
pipe:

```ant
let x = 9.0
let r = x.sqrt() // error: float64 has no field or function sqrt
```

## See also

[Math.cbrt](cbrt.md) · [Math.pow](pow.md) · [Math.hypot](hypot.md) · [Math.abs](abs.md) ·
[Math.isNaN](isNaN.md) · [Math.toInt](toInt.md) · [Math namespace](index.md)
