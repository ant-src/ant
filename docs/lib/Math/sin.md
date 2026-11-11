<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.sin

```ant signature
func Math.sin(x: float64): float64
```

Math.sin returns the sine of x, in radians.

<!-- end generated -->

## Details

- The angle is in radians. Multiply degrees by `Math.PI / 180.0` first.
- `Math.sin` never panics: `NaN` and both infinities give `NaN`.

## Basic examples

The sine of a few angles, in radians:

```ant
Math.sin(0.0)           // 0.0
Math.sin(Math.PI / 2.0) // 1.0
Math.sin(1.0)           // 0.8414709848078965
```

Convert degrees to radians before the call:

```ant
let degrees = 90.0
Math.sin(degrees * Math.PI / 180.0) // 1.0
```

## Scope

A negative angle gives the negated sine, and zero keeps its sign:

```ant
Math.sin(-1.0) // -0.8414709848078965
Math.sin(-0.0) // -0.0
```

Angles beyond a full turn wrap around:

```ant
Math.sin(2.5 * Math.PI) // 1.0
Math.sin(1e300)         // -0.8178819121159087
```

A tiny angle gives itself back, since `sin(x)` is close to `x` near zero:

```ant
Math.sin(1e-10) // 1e-10
```

`NaN` and the infinities give `NaN`:

```ant
Math.sin(Math.NaN)       // NaN
Math.sin(Math.Infinity)  // NaN
Math.sin(-Math.Infinity) // NaN
```

In a pipeline, the angle is the argument:

```ant
Math.PI / 2.0 | Math.sin() // 1.0
```

Pass it as a function value to map a list of angles:

```ant
[0.0, Math.PI / 2.0, -Math.PI / 2.0].map(Math.sin) // [0.0, 1.0, -1.0]
```

## Applications

Sample one period of a sine wave at eight points, as a bar of text per sample:

```ant
let steps = 8
for i in Iter.range(0, steps) {
  let y = Math.sin(2.0 * Math.PI * float64(i) / float64(steps))
  let width = int(Math.round((y + 1.0) * 4.0)) + 1
  println("#".repeat(width))
}
```

```text
#####
########
#########
########
#####
##
#
##
```

Find the height a ladder reaches from its length and the angle it makes with the ground:

```ant
func reach(length: float64, degrees: float64): float64 {
  length * Math.sin(degrees * Math.PI / 180.0)
}

String.fromFloat(reach(5.0, 75.0), 2) // "4.83"
```

Find the distance between two points on the Earth by the haversine formula:

```ant
type Place = { lat: float64, lon: float64 }

func radians(degrees: float64): float64 {
  degrees * Math.PI / 180.0
}

func distanceKm(a: Place, b: Place): float64 {
  let dLat = radians(b.lat - a.lat)
  let dLon = radians(b.lon - a.lon)
  let sLat = Math.sin(dLat / 2.0)
  let sLon = Math.sin(dLon / 2.0)
  let h = sLat * sLat + Math.cos(radians(a.lat)) * Math.cos(radians(b.lat)) * sLon * sLon
  2.0 * 6371.0 * Math.asin(Math.sqrt(h))
}

let paris = Place{ lat: 48.8566, lon: 2.3522 }
let london = Place{ lat: 51.5074, lon: -0.1278 }
Math.round(distanceKm(paris, london)) // 344.0
```

## Properties and relations

Sine is odd: `Math.sin(-x)` is the negation of `Math.sin(x)`:

```ant
for x in [0.5, 1.0, 2.0, 100.0] {
  assert(Math.sin(-x) == -Math.sin(x))
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

The sine of an angle is the cosine of a quarter turn less, within rounding:

```ant
for x in [0.0, 0.5, 1.0, 3.0] {
  assert(Math.abs(Math.sin(x) - Math.cos(Math.PI / 2.0 - x)) < 1e-15)
}
```

`Math.asin` undoes `Math.sin` for angles from `-Math.PI / 2.0` to `Math.PI / 2.0`, within rounding:

```ant
for x in [-1.5, -0.5, 0.0, 0.5, 1.5] {
  assert(Math.abs(Math.asin(Math.sin(x)) - x) < 1e-15)
}
```

## Possible issues

An angle in degrees is read as radians:

```ant
Math.sin(30.0)                   // -0.9880316240928618
Math.sin(30.0 * Math.PI / 180.0) // 0.49999999999999994
```

`Math.PI` is not exactly π, so a sine that is 0 in exact arithmetic comes out as a tiny number.
Compare with a tolerance:

```ant
Math.sin(Math.PI)                   // 1.2246467991473515e-16
Math.sin(Math.PI) == 0.0            // false
Math.abs(Math.sin(Math.PI)) < 1e-12 // true
```

`Math.asin` does not undo `Math.sin` for an angle past a quarter turn. It returns another angle with
the same sine:

```ant
Math.asin(Math.sin(3.0)) // 0.1415926535897932
```

An integer is not a `float64`. Write a float literal, or convert:

```ant
let n = 1
Math.sin(n) // error: cannot use int where float64 is expected
```

```ant
let n = 1
Math.sin(float64(n)) // 0.8414709848078965
```

A `float64` has no namespace of its own, so `x.sin()` does not resolve. Call `Math.sin(x)` or pipe
the value:

```ant
let x = 1.0
x.sin() // error: float64 has no field or function sin
```

## See also

[Math.cos](cos.md) · [Math.tan](tan.md) · [Math.asin](asin.md) · [Math.atan2](atan2.md) ·
[Math](index.md)
