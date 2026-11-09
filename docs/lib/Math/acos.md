<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.acos

```ant signature
func Math.acos(x: float64): float64
```

Math.acos returns the arccosine of x, in radians.

<!-- end generated -->

## Details

- The result is in `0.0 ..= Math.PI`.
- An argument outside `-1.0 ..= 1.0`, or `NaN`, gives `NaN`; `Math.acos` never panics.

## Basic examples

The arccosine of a few common values:

```ant
Math.acos(1.0)  // 0.0
Math.acos(0.5)  // 1.0471975511965976
Math.acos(0.0)  // 1.5707963267948966
Math.acos(-1.0) // 3.141592653589793
```

Convert the result to degrees:

```ant
Math.acos(0.5) * 180.0 / Math.PI // 59.99999999999999
```

## Scope

Negative arguments give angles above a right angle:

```ant
Math.acos(-0.5) // 2.0943951023931957
```

`-0.0` gives the same angle as `0.0`:

```ant
Math.acos(-0.0) // 1.5707963267948966
```

Values outside `-1.0 ..= 1.0`, an infinity among them, and `NaN` give `NaN`:

```ant
Math.acos(2.0)           // NaN
Math.acos(-1.5)          // NaN
Math.acos(Math.Infinity) // NaN
Math.acos(Math.NaN)      // NaN
```

Call it in a pipe, or pass it to another function as a value:

```ant
0.5 | Math.acos()                       // 1.0471975511965976
[1.0, 0.0, -1.0] | Array.map(Math.acos) // [0.0, 1.5707963267948966, 3.141592653589793]
```

## Applications

Find the angle between two vectors from their dot product:

```ant
type Vec = { x: float64, y: float64 }

func angle(a: Vec, b: Vec): float64 {
  let dot = a.x * b.x + a.y * b.y
  Math.acos(dot / (Math.hypot(a.x, a.y) * Math.hypot(b.x, b.y)))
}

println(angle(Vec{ x: 1.0, y: 0.0 }, Vec{ x: 0.0, y: 2.0 }))
println(angle(Vec{ x: 3.0, y: 4.0 }, Vec{ x: -3.0, y: -4.0 }))
```

```text
1.5707963267948966
3.141592653589793
```

Find the angle of a triangle opposite side `c` with the law of cosines:

```ant
func opposite(a: float64, b: float64, c: float64): float64 {
  Math.acos((a * a + b * b - c * c) / (2.0 * a * b))
}

let degrees = opposite(3.0, 4.0, 5.0) * 180.0 / Math.PI
println(degrees)
```

```text
90.0
```

## Properties and relations

The result is never negative and never above `Math.PI`:

```ant
for x in [-1.0, -0.7, 0.0, 0.2, 1.0] {
  let t = Math.acos(x)
  assert(t >= 0.0 && t <= Math.PI)
}
```

`Math.acos` and [`Math.asin`](asin.md) of one argument add up to a right angle, within rounding:

```ant
for x in [-0.9, -0.3, 0.1, 0.5, 0.8] {
  assert(Math.abs(Math.acos(x) + Math.asin(x) - Math.PI / 2.0) < 1e-15)
}
```

[`Math.cos`](cos.md) undoes it, within rounding:

```ant
for x in [-0.9, -0.3, 0.1, 0.5, 0.8] {
  assert(Math.abs(Math.cos(Math.acos(x)) - x) < 1e-15)
}
```

## Possible issues

Ant does not convert an `int` to a `float64`; write a float literal or convert with `float64`:

```ant
let half = Math.acos(1)
// error: cannot use int where float64 is expected
```

```ant
let n = 1
Math.acos(float64(n)) // 0.0
```

`Math.acos` does not undo [`Math.cos`](cos.md) for an angle outside `0.0 ..= Math.PI`. It gives the
angle in that range with the same cosine:

```ant
Math.acos(Math.cos(4.0))  // 2.2831853071795862
Math.acos(Math.cos(-1.0)) // 0.9999999999999999
```

Rounding can push a cosine computed from floats just past `1.0`, which gives `NaN`. Hold the
argument in range with [`Math.clamp`](clamp.md):

```ant
type Vec = { x: float64, y: float64 }

func unit(v: Vec): Vec {
  let n = Math.hypot(v.x, v.y)
  Vec{ x: v.x / n, y: v.y / n }
}

let u = unit(Vec{ x: 0.5, y: 0.3 })
let cos = u.x * u.x + u.y * u.y
println(cos)
println(Math.acos(cos))
println(Math.acos(Math.clamp(cos, -1.0, 1.0)))
```

```text
1.0000000000000004
NaN
0.0
```

`NaN` is not equal to itself, so test for it with [`Math.isNaN`](isNaN.md):

```ant
Math.acos(2.0) == Math.NaN // false
Math.isNaN(Math.acos(2.0)) // true
```

## See also

[Math.cos](cos.md) · [Math.asin](asin.md) · [Math.atan](atan.md) · [Math.atan2](atan2.md) ·
[Math.clamp](clamp.md) · [Math.isNaN](isNaN.md) · [Math.hypot](hypot.md)
