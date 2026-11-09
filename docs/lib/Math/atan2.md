<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.atan2

```ant signature
func Math.atan2(y: float64, x: float64): float64
```

Math.atan2 returns the angle of the point (x, y), in radians from -PI to PI inclusive. Note the
argument order: y comes first.

<!-- end generated -->

## Details

- The angle is measured counterclockwise from the positive x-axis. Points above the x-axis give a
  positive angle, points below it a negative one.

## Basic examples

The angle of the point (1, 1), a quarter of `PI`:

```ant
Math.atan2(1.0, 1.0) // 0.7853981633974483
```

Straight up and straight left:

```ant
Math.atan2(1.0, 0.0)  // 1.5707963267948966
Math.atan2(0.0, -1.0) // 3.141592653589793
```

A point below the x-axis gives a negative angle:

```ant
Math.atan2(-1.0, -1.0) // -2.356194490192345
```

## Scope

Name the arguments to make the order plain at the call:

```ant
Math.atan2(y: 1.0, x: 0.0) // 1.5707963267948966
```

On the x-axis, the signs of the zeros pick between `0.0`, `-0.0`, `PI`, and `-PI`:

```ant
Math.atan2(0.0, 0.0)   // 0.0
Math.atan2(-0.0, 0.0)  // -0.0
Math.atan2(0.0, -0.0)  // 3.141592653589793
Math.atan2(-0.0, -1.0) // -3.141592653589793
```

Infinite coordinates give the angle of the direction they point in:

```ant
Math.atan2(1.0, Math.Infinity)           // 0.0
Math.atan2(-Math.Infinity, 3.0)          // -1.5707963267948966
Math.atan2(Math.Infinity, Math.Infinity) // 0.7853981633974483
```

`NaN` in either argument gives `NaN`:

```ant
Math.atan2(Math.NaN, 1.0) // NaN
Math.atan2(1.0, Math.NaN) // NaN
```

Coordinates of very different sizes still give the right angle, where `y / x` would underflow:

```ant
Math.atan2(1.0e-300, -1.0e300) // 3.141592653589793
```

Fields of a record work like any other `float64`:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: -3.0, y: 4.0 }
Math.atan2(p.y, p.x) // 2.214297435588181
```

## Applications

A direction in degrees from 0 up to 360, counterclockwise from east:

```ant
type Point = { x: float64, y: float64 }

func heading(from: Point, to: Point): float64 {
  let a = Math.atan2(to.y - from.y, to.x - from.x) * 180.0 / Math.PI
  if a < 0.0 { a + 360.0 } else { a }
}

let home = Point{ x: 0.0, y: 0.0 }
let east = Point{ x: 1.0, y: 0.0 }
let north = Point{ x: 0.0, y: 2.0 }
let west = Point{ x: -3.0, y: 0.0 }
let south = Point{ x: 0.0, y: -1.0 }
for p in [east, north, west, south] {
  println(heading(home, p))
}
```

```text
0.0
90.0
180.0
270.0
```

The signed turn from one direction to another, positive counterclockwise, from the cross and dot
products:

```ant
type Vec = { x: float64, y: float64 }

func turn(a: Vec, b: Vec): float64 {
  Math.atan2(a.x * b.y - a.y * b.x, a.x * b.x + a.y * b.y) * 180.0 / Math.PI
}

println(turn(Vec{ x: 1.0, y: 0.0 }, Vec{ x: 0.0, y: 1.0 }))
println(turn(Vec{ x: 0.0, y: 1.0 }, Vec{ x: 1.0, y: 0.0 }))
println(turn(Vec{ x: 1.0, y: 1.0 }, Vec{ x: -1.0, y: 0.0 }))
```

```text
90.0
-90.0
135.0
```

Convert a point to polar coordinates with [`Math.hypot`](hypot.md):

```ant
type Polar = { r: float64, theta: float64 }

func toPolar(x: float64, y: float64): Polar {
  Polar{ r: Math.hypot(x, y), theta: Math.atan2(y, x) }
}

println(toPolar(-3.0, 4.0))
```

```text
Polar{ r: 5.0, theta: 2.214297435588181 }
```

## Properties and relations

For a positive `x`, it agrees with [`Math.atan`](atan.md) of the ratio:

```ant
assert(Math.abs(Math.atan2(2.0, 4.0) - Math.atan(2.0 / 4.0)) < 1.0e-15)
assert(Math.abs(Math.atan2(-3.0, 5.0) - Math.atan(-3.0 / 5.0)) < 1.0e-15)
```

Scaling the point by a positive number keeps the angle, and mirroring it across the x-axis negates
the angle:

```ant
assert(Math.abs(Math.atan2(2.0, 4.0) - Math.atan2(1.0, 2.0)) < 1.0e-15)
assert(Math.abs(Math.atan2(-3.0, 5.0) + Math.atan2(3.0, 5.0)) < 1.0e-15)
```

With [`Math.hypot`](hypot.md), [`Math.cos`](cos.md), and [`Math.sin`](sin.md) it round-trips a
point, up to rounding:

```ant
let x = -3.0
let y = 4.0
let r = Math.hypot(x, y)
let t = Math.atan2(y, x)
assert(Math.abs(r * Math.cos(t) - x) < 1.0e-12)
assert(Math.abs(r * Math.sin(t) - y) < 1.0e-12)
```

## Possible issues

`y` comes first. Swapping the arguments gives the angle of the mirrored point:

```ant
Math.atan2(1.0, 0.0) // 1.5707963267948966
Math.atan2(0.0, 1.0) // 0.0
```

[`Math.atan`](atan.md) of `y / x` loses the quadrant: (-1, -1) and (1, 1) give the same ratio:

```ant
Math.atan(-1.0 / -1.0) // 0.7853981633974483
Math.atan2(-1.0, -1.0) // -2.356194490192345
```

The result runs from `-PI` to `PI`, not from `0` to `2 * PI`. Add `2 * PI` to a negative angle when
you need the other range:

```ant
let a = Math.atan2(-1.0, 0.0)
a                                           // -1.5707963267948966
if a < 0.0 { a + 2.0 * Math.PI } else { a } // 4.71238898038469
```

An integer literal is an `int`, not a `float64`, so passing one is a compile error:

```ant
println(Math.atan2(1, 1))
// error: cannot use int where float64 is expected
```

Convert integer coordinates with `float64` first:

```ant
let dx = 3
let dy = -3
Math.atan2(float64(dy), float64(dx)) // -0.7853981633974483
```

## See also

[Math.atan](atan.md) · [Math.hypot](hypot.md) · [Math.sin](sin.md) · [Math.cos](cos.md) ·
[Math.tan](tan.md) · [Math.asin](asin.md) · [Math.acos](acos.md) · [Math index](index.md)
