<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.cos

```ant signature
func Math.cos(x: float64): float64
```

Math.cos returns the cosine of x, in radians.

<!-- end generated -->

## Details

- `Math.cos` never panics: `NaN` and both infinities give `NaN`.

## Basic examples

The cosine of a few angles, in radians:

```ant
Math.cos(0.0)     // 1.0
Math.cos(Math.PI) // -1.0
Math.cos(1.0)     // 0.5403023058681398
```

Convert degrees to radians before the call:

```ant
let degrees = 180.0
Math.cos(degrees * Math.PI / 180.0) // -1.0
```

## Scope

Negative angles give the same cosine as positive ones:

```ant
Math.cos(-1.0) // 0.5403023058681398
Math.cos(-0.0) // 1.0
```

Angles beyond a full turn wrap around:

```ant
Math.cos(2.0 * Math.PI) // 1.0
Math.cos(1e300)         // -0.5753861119575491
```

`NaN` and the infinities give `NaN`:

```ant
Math.cos(Math.NaN)       // NaN
Math.cos(Math.Infinity)  // NaN
Math.cos(-Math.Infinity) // NaN
```

In a pipeline, the angle is the argument:

```ant
Math.PI | Math.cos() // -1.0
```

Pass it as a function value to map a list of angles:

```ant
[0.0, Math.PI].map(Math.cos) // [1.0, -1.0]
```

## Applications

Place points evenly on a circle of radius 2:

```ant
type Point = { x: float64, y: float64 }

func onCircle(radius: float64, angle: float64): Point {
  Point{ x: radius * Math.cos(angle), y: radius * Math.sin(angle) }
}

for i in Iter.range(0, 4) {
  let p = onCircle(2.0, float64(i) * Math.PI / 2.0)
  println("${Math.round(p.x)}, ${Math.round(p.y)}")
}
```

```text
2.0, 0.0
0.0, 2.0
-2.0, 0.0
-0.0, -2.0
```

The last `x` is a tiny negative number left by rounding in `Math.PI`, so `Math.round` gives `-0.0`.

Rotate a point about the origin by an angle:

```ant
type Point = { x: float64, y: float64 }

func rotate(p: Point, angle: float64): Point {
  let c = Math.cos(angle)
  let s = Math.sin(angle)
  Point{ x: p.x * c - p.y * s, y: p.x * s + p.y * c }
}

let r = rotate(Point{ x: 1.0, y: 0.0 }, Math.PI)
r.x                   // -1.0
Math.abs(r.y) < 1e-12 // true
```

Find the third side of a triangle from two sides and the angle between them, by the law of cosines:

```ant
func thirdSide(a: float64, b: float64, angle: float64): float64 {
  Math.sqrt(a * a + b * b - 2.0 * a * b * Math.cos(angle))
}

thirdSide(3.0, 4.0, Math.PI / 2.0) // 5.0
```

## Properties and relations

Cosine is even: `Math.cos(-x)` equals `Math.cos(x)`:

```ant
for x in [0.5, 1.0, 2.0, 100.0] {
  assert(Math.cos(-x) == Math.cos(x))
}
```

The squares of the sine and the cosine add to 1, within rounding:

```ant
for x in [0.1, 1.0, 2.5, 10.0] {
  let s = Math.sin(x)
  let c = Math.cos(x)
  assert(Math.abs(s * s + c * c - 1.0) < 1e-15)
}
```

`Math.acos` undoes `Math.cos` for angles in `0.0 ..= Math.PI`, within rounding:

```ant
for x in [0.0, 0.5, 1.0, 3.0] {
  assert(Math.abs(Math.acos(Math.cos(x)) - x) < 1e-15)
}
```

## Possible issues

An angle in degrees is read as radians:

```ant
Math.cos(60.0)                   // -0.9524129804151563
Math.cos(60.0 * Math.PI / 180.0) // 0.5000000000000001
```

`Math.PI` is not exactly π, so a cosine that is 0 in exact arithmetic comes out as a tiny number.
Compare with a tolerance:

```ant
Math.cos(Math.PI / 2.0)                   // 6.123233995736757e-17
Math.cos(Math.PI / 2.0) == 0.0            // false
Math.abs(Math.cos(Math.PI / 2.0)) < 1e-12 // true
```

`Math.acos` returns an angle in `0.0 ..= Math.PI`, so it does not undo `Math.cos` outside it:

```ant
Math.acos(Math.cos(4.0)) // 2.2831853071795862
```

An integer is not a `float64`. An integer literal is always an `int`, so write `2.0`:

```ant
Math.cos(2) // error: cannot use int where float64 is expected
```

For an `int` variable, convert:

```ant
let n = 2
Math.cos(float64(n)) // -0.4161468365471424
```

A `float64` has no namespace of its own, so `x.cos()` does not resolve. Call `Math.cos(x)` or pipe
the value:

```ant
let x = 1.0
x.cos() // error: float64 has no field or function cos
```

## See also

[Math.sin](sin.md) · [Math.tan](tan.md) · [Math.acos](acos.md) · [Math.atan2](atan2.md) ·
[Math.hypot](hypot.md) · [Math](index.md)
