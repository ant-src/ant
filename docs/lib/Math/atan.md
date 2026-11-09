<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.atan

```ant signature
func Math.atan(x: float64): float64
```

Math.atan returns the arctangent of x, in radians.

<!-- end generated -->

## Details

- The result lies between `-Math.PI / 2` and `Math.PI / 2`. An infinity gives the bound of its sign.
- `Math.atan` never panics. `NaN` gives `NaN`, and `-0.0` gives `-0.0`.

## Basic examples

```ant
Math.atan(1.0)  // 0.7853981633974483
Math.atan(0.0)  // 0.0
Math.atan(-1.0) // -0.7853981633974483
```

Convert the result to degrees:

```ant
Math.atan(1.0) * 180.0 / Math.PI // 45.0
```

## Scope

Large arguments approach a right angle, and the infinities reach it:

```ant
Math.atan(1e300)          // 1.5707963267948966
Math.atan(Math.Infinity)  // 1.5707963267948966
Math.atan(-Math.Infinity) // -1.5707963267948966
```

`NaN` passes through, and the sign of a zero is kept:

```ant
Math.atan(Math.NaN) // NaN
Math.atan(-0.0)     // -0.0
```

Call it in a pipe, or pass it as a function value:

```ant
0.5 | Math.atan()               // 0.4636476090008061
[0.0, 1.0, -1.0].map(Math.atan) // [0.0, 0.7853981633974483, -0.7853981633974483]
```

## Applications

The angle of a ramp from its rise and run, in degrees:

```ant
func incline(rise: float64, run: float64): float64 {
  Math.atan(rise / run) * 180.0 / Math.PI
}

incline(1.0, 1.0)  // 45.0
incline(1.0, 12.0) // 4.763641690726177
```

Compute π from the arctangent of 1:

```ant
4.0 * Math.atan(1.0) // 3.141592653589793
```

## Properties and relations

`Math.atan` is odd: it maps `-x` to minus its value at `x`:

```ant
for x in [0.25, 1.0, 3.0, 100.0] {
  assert(Math.atan(-x) == -Math.atan(x))
}
```

It undoes [Math.tan](tan.md) on angles between `-Math.PI / 2` and `Math.PI / 2`, up to rounding:

```ant
for a in [-1.2, -0.5, 0.0, 0.5, 1.2] {
  assert(Math.abs(Math.atan(Math.tan(a)) - a) < 1e-12)
}
```

It agrees with [Math.atan2](atan2.md) on the point `(1, y)`, up to rounding:

```ant
for y in [-2.0, 0.5, 3.0] {
  assert(Math.abs(Math.atan(y) - Math.atan2(y, 1.0)) < 1e-15)
}
```

## Possible issues

An integer argument is a compile error. Write a float literal or convert:

```ant
let n = 1
Math.atan(n) // error: cannot use int where float64 is expected
```

```ant
let n = 1
Math.atan(float64(n)) // 0.7853981633974483
```

`Math.atan` of a ratio loses the quadrant. The points `(1, 1)` and `(-1, -1)` have the same ratio,
so they get the same angle. Use [Math.atan2](atan2.md) to tell them apart:

```ant
Math.atan(1.0 / 1.0)   // 0.7853981633974483
Math.atan(-1.0 / -1.0) // 0.7853981633974483
Math.atan2(1.0, 1.0)   // 0.7853981633974483
Math.atan2(-1.0, -1.0) // -2.356194490192345
```

A float ratio with a zero run is an infinity, so `Math.atan` gives a right angle instead of
panicking:

```ant
Math.atan(1.0 / 0.0) // 1.5707963267948966
```

The inverse holds only within `-Math.PI / 2` to `Math.PI / 2`. Outside it, `Math.atan` returns the
angle that differs by a multiple of `Math.PI`:

```ant
Math.atan(Math.tan(3.0)) // -0.14159265358979323
```

## See also

[Math.atan2](atan2.md) · [Math.tan](tan.md) · [Math.asin](asin.md) · [Math.acos](acos.md) ·
[Math.PI](index.md) · [float64](../../language/types/float.md)
