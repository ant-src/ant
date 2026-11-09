<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.cbrt

```ant signature
func Math.cbrt(x: float64): float64
```

Math.cbrt returns the cube root of x.

<!-- end generated -->

## Details

- A negative `x` has a negative cube root, so `Math.cbrt(-8.0)` is `-2.0`, not `NaN`.
- `Math.cbrt` never panics: `NaN` gives `NaN`, and an infinity gives the infinity of the same sign.

## Basic examples

```ant
Math.cbrt(27.0) // 3.0
Math.cbrt(2.0)  // 1.2599210498948732
```

A negative number has a real cube root:

```ant
Math.cbrt(-8.0) // -2.0
```

## Scope

Zero keeps its sign, and a positive fraction below 1 gives a larger result:

```ant
Math.cbrt(0.0)   // 0.0
Math.cbrt(-0.0)  // -0.0
Math.cbrt(0.001) // 0.1
```

Infinities and `NaN` pass through:

```ant
Math.cbrt(Math.Infinity)  // Infinity
Math.cbrt(-Math.Infinity) // -Infinity
Math.cbrt(Math.NaN)       // NaN
```

The largest `float64` has a finite cube root:

```ant
Math.cbrt(Math.maxFloat64) // 5.643803094122362e+102
```

An integer goes through `float64` first:

```ant
let n = 125
Math.cbrt(float64(n)) // 5.0
```

Pass `Math.cbrt` as a function value, or reach a field of a record in a lambda:

```ant
[1.0, 8.0, 27.0].map(Math.cbrt) // [1.0, 2.0, 3.0]
type Box = { volume: float64 }
let boxes = [Box{ volume: 8.0 }, Box{ volume: 1000.0 }]
boxes.map({ Math.cbrt(x.volume) }) // [2.0, 10.0]
```

In a pipe, the value is the argument:

```ant
64.0 | Math.cbrt() // 4.0
```

## Applications

The radius of a sphere from its volume, `r = ∛(3V / 4π)`:

```ant
func sphereRadius(volume: float64): float64 {
  Math.cbrt(3.0 * volume / (4.0 * Math.PI))
}

sphereRadius(4.0 / 3.0 * Math.PI * 8.0) // 2.0
sphereRadius(1.0)                       // 0.6203504908994
```

The geometric mean of three numbers:

```ant
func geometricMean(a: float64, b: float64, c: float64): float64 {
  Math.cbrt(a * b * c)
}

geometricMean(2.0, 4.0, 8.0)    // 4.0
geometricMean(1.0, 10.0, 100.0) // 10.0
```

## Properties and relations

For these perfect cubes, the cube of the cube root gives the number back, and negating the number
negates the root:

```ant
for n in [1.0, 8.0, 27.0, 64.0, -125.0] {
  let r = Math.cbrt(n)
  assert(r * r * r == n)
  assert(Math.cbrt(-n) == -r)
}
```

The cube root of the square root is the sixth root:

```ant
Math.cbrt(Math.sqrt(64.0)) // 2.0
Math.sqrt(Math.cbrt(64.0)) // 2.0
```

## Possible issues

`Math.pow(x, 1.0 / 3.0)` is not a cube root: it gives `NaN` for a negative `x`, and `1.0 / 3.0` is
not exactly a third, so a perfect cube can miss:

```ant
Math.pow(-8.0, 1.0 / 3.0) // NaN
Math.pow(27.0, 1.0 / 3.0) // 2.9999999999999996
Math.cbrt(27.0)           // 3.0
```

An `int` argument is a compile error, since numeric types never mix:

```ant
let side = Math.cbrt(27) // error: cannot use int where float64 is expected
```

Converting the result back to `int` panics when it has a fraction. Use [Math.toInt](toInt.md) to get
`null` instead, or round first:

```ant
int(Math.cbrt(64.0))             // 4
int(Math.cbrt(10.0))             // panic: 2.154434690031884 has a fraction, so it does not convert to int
Math.toInt(Math.cbrt(10.0))      // null
int(Math.round(Math.cbrt(10.0))) // 2
```

## See also

[Math.sqrt](sqrt.md) · [Math.pow](pow.md) · [Math.hypot](hypot.md) · [Math.exp](exp.md) ·
[Math.log](log.md) · [Math.toInt](toInt.md) · [float](../../language/types/float.md)
