<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.asin

```ant signature
func Math.asin(x: float64): float64
```

Math.asin returns the arcsine of x, in radians.

<!-- end generated -->

## Details

- The result is an angle in radians, from `-Math.PI / 2.0` to `Math.PI / 2.0`.
- An `x` outside `-1.0 ..= 1.0`, an infinity, or `NaN` gives `NaN`. `Math.asin` never panics.

## Basic examples

```ant
Math.asin(0.5)  // 0.5235987755982989
Math.asin(1.0)  // 1.5707963267948966
Math.asin(-1.0) // -1.5707963267948966
```

The arcsine of `0.5` is 30 degrees:

```ant
Math.asin(0.5) * 180.0 / Math.PI // 30.000000000000004
```

## Scope

Zero keeps its sign:

```ant
Math.asin(0.0)  // 0.0
Math.asin(-0.0) // -0.0
```

A value outside the domain gives `NaN`, as does `NaN` itself:

```ant
Math.asin(2.0)           // NaN
Math.asin(1.0000001)     // NaN
Math.asin(Math.Infinity) // NaN
Math.asin(Math.NaN)      // NaN
```

Pass it as a function value, or call it through a pipe:

```ant
[0.0, 0.5, 1.0].map(Math.asin) // [0.0, 0.5235987755982989, 1.5707963267948966]
0.5 | Math.asin()              // 0.5235987755982989
```

## Applications

The angle of a ramp from its rise and its length, in degrees:

```ant
func slope(rise: float64, length: float64): float64 {
  Math.asin(rise / length) * 180.0 / Math.PI
}

String.fromFloat(slope(1.0, 2.0), 1)  // "30.0"
String.fromFloat(slope(1.0, 12.0), 2) // "4.78"
```

Snell's law gives the angle of a ray of light after it crosses into another medium. When the sine
comes out above 1, there is no refracted ray, and `NaN` signals total internal reflection:

```ant
func refract(angle: float64, from: float64, to: float64): float64 | null {
  let r = Math.asin(from / to * Math.sin(angle))
  if Math.isNaN(r) { null } else { r }
}

refract(0.5, 1.0, 1.33) // 0.3687720730730495
refract(1.2, 1.33, 1.0) // null
```

## Properties and relations

`Math.asin` is odd: it maps `-x` to the negation of `Math.asin(x)`:

```ant
for x in [0.0, 0.3, 0.6, 1.0] {
  assert(Math.asin(-x) == -Math.asin(x))
}
```

`Math.sin` undoes `Math.asin` up to rounding, and `Math.asin(x) + Math.acos(x)` is a right angle.
Compare floats with a tolerance:

```ant
let eps = 1e-15
for x in [-1.0, -0.6, 0.0, 0.25, 0.9] {
  assert(Math.abs(Math.sin(Math.asin(x)) - x) < eps)
  assert(Math.abs(Math.asin(x) + Math.acos(x) - Math.PI / 2.0) < eps)
}
```

## Possible issues

`Math.asin` undoes `Math.sin` only for angles from `-Math.PI / 2.0` to `Math.PI / 2.0`. For any
other angle it returns the angle in that range with the same sine, up to rounding:

```ant
Math.asin(Math.sin(2.0)) // 1.141592653589793
Math.PI - 2.0            // 1.1415926535897931
```

A sine computed with rounding can land just past 1. Clamp it first, or the result is `NaN`:

```ant
let s = 0.1 * 3.0 / 0.3
s                                   // 1.0000000000000002
Math.asin(s)                        // NaN
Math.asin(Math.clamp(s, -1.0, 1.0)) // 1.5707963267948966
```

The argument is a `float64`. An integer literal is an `int`, so convert it or write `1.0`:

```ant
Math.asin(1) // error: cannot use int where float64 is expected
```

## See also

[Math.acos](acos.md), [Math.atan](atan.md), [Math.atan2](atan2.md), [Math.sin](sin.md),
[Math.isNaN](isNaN.md), [Math.clamp](clamp.md)
