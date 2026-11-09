<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.tan

```ant signature
func Math.tan(x: float64): float64
```

Math.tan returns the tangent of x, in radians.

<!-- end generated -->

## Details

- The angle is in radians. Multiply degrees by `Math.PI / 180.0` first.
- `Math.tan` never panics: `NaN` and both infinities give `NaN`, and `-0.0` gives `-0.0`.

## Basic examples

The tangent of a few angles, in radians:

```ant
Math.tan(0.0)           // 0.0
Math.tan(1.0)           // 1.557407724654902
Math.tan(Math.PI / 3.0) // 1.7320508075688767
```

Convert degrees to radians before the call:

```ant
let degrees = 30.0
Math.tan(degrees * Math.PI / 180.0) // 0.5773502691896257
```

## Scope

Negative angles give the negated tangent, and the sign of a zero is kept:

```ant
Math.tan(-1.0) // -1.557407724654902
Math.tan(-0.0) // -0.0
```

The tangent repeats every π, so adding `Math.PI` gives the same value within rounding:

```ant
Math.tan(1.0 + Math.PI) // 1.5574077246549016
```

`NaN` and the infinities give `NaN`:

```ant
Math.tan(Math.NaN)       // NaN
Math.tan(Math.Infinity)  // NaN
Math.tan(-Math.Infinity) // NaN
```

Call it in a pipe, or pass it as a function value:

```ant
Math.PI / 4.0 | Math.tan() // 0.9999999999999998
[0.0, 1.0].map(Math.tan)   // [0.0, 1.557407724654902]
```

## Applications

The height of a tree from the distance to its base and the angle up to its top:

```ant
func treeHeight(distance: float64, degrees: float64): float64 {
  distance * Math.tan(degrees * Math.PI / 180.0)
}

treeHeight(20.0, 30.0) // 11.547005383792515
```

The grade of a road, in percent, from its angle in degrees:

```ant
type Road = { name: string, degrees: float64 }

func grade(road: Road): float64 {
  Math.round(1000.0 * Math.tan(road.degrees * Math.PI / 180.0)) / 10.0
}

let roads = [Road{ name: "Main", degrees: 2.0 }, Road{ name: "Hill", degrees: 5.0 }]
for road in roads {
  println("${road.name}: ${grade(road)}%")
}
```

```text
Main: 3.5%
Hill: 8.7%
```

## Properties and relations

`Math.tan` is odd: it maps `-x` to minus its value at `x`:

```ant
for x in [0.25, 1.0, 1.5, 100.0] {
  assert(Math.tan(-x) == -Math.tan(x))
}
```

The tangent is the sine over the cosine, within rounding:

```ant
for x in [0.1, 1.0, 2.5, 10.0] {
  let ratio = Math.sin(x) / Math.cos(x)
  assert(Math.abs(Math.tan(x) - ratio) < 1e-12 * Math.abs(ratio))
}
```

[Math.atan](atan.md) undoes `Math.tan` for angles between `-Math.PI / 2` and `Math.PI / 2`, within
rounding:

```ant
for x in [-1.2, -0.5, 0.0, 0.5, 1.2] {
  assert(Math.abs(Math.atan(Math.tan(x)) - x) < 1e-15)
}
```

## Possible issues

An angle in degrees is read as radians:

```ant
Math.tan(45.0)                   // 1.6197751905438615
Math.tan(45.0 * Math.PI / 180.0) // 0.9999999999999998
```

`Math.PI` is not exactly π, so a tangent that is exact in arithmetic comes out a little off. Compare
with a tolerance:

```ant
Math.tan(Math.PI / 4.0) == 1.0                  // false
Math.abs(Math.tan(Math.PI / 4.0) - 1.0) < 1e-12 // true
Math.tan(Math.PI)                               // -1.2246467991473515e-16
```

The tangent at a right angle is not an infinity but a large finite number, so a test with
[Math.isFinite](isFinite.md) does not catch it:

```ant
Math.tan(Math.PI / 2.0)                // 16331239353195392.0
Math.isFinite(Math.tan(Math.PI / 2.0)) // true
```

[Math.atan](atan.md) returns an angle between `-Math.PI / 2` and `Math.PI / 2`, so it does not undo
`Math.tan` outside it:

```ant
Math.atan(Math.tan(2.0)) // -1.1415926535897931
```

An integer is not a `float64`. Write a float literal, or convert:

```ant
let n = 1
Math.tan(n) // error: cannot use int where float64 is expected
```

```ant
let n = 1
Math.tan(float64(n)) // 1.557407724654902
```

A `float64` has no namespace of its own, so `x.tan()` does not resolve. Call `Math.tan(x)` or pipe
the value:

```ant
let x = 1.0
x.tan() // error: float64 has no field or function tan
```

## See also

[Math.sin](sin.md) · [Math.cos](cos.md) · [Math.atan](atan.md) · [Math.atan2](atan2.md) ·
[Math.PI](index.md) · [float64](../../language/types/float.md)
