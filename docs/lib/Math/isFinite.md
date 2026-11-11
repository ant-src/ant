<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.isFinite

```ant signature
func Math.isFinite(x: float64): bool
```

Math.isFinite reports whether x is neither NaN nor an infinity.

<!-- end generated -->

## Details

- Float arithmetic never panics: an overflow gives an infinity and `0.0 / 0.0` gives `NaN`, so
  `Math.isFinite` is how you check a result before you use it.

## Basic examples

```ant
Math.isFinite(1.5)           // true
Math.isFinite(Math.Infinity) // false
Math.isFinite(Math.NaN)      // false
```

## Scope

Both infinities are not finite:

```ant
Math.isFinite(Math.Infinity)  // false
Math.isFinite(-Math.Infinity) // false
```

Zero of either sign, the largest finite value, and the smallest positive value are finite:

```ant
Math.isFinite(0.0)             // true
Math.isFinite(-0.0)            // true
Math.isFinite(Math.maxFloat64) // true
Math.isFinite(5e-324)          // true
```

Arithmetic and `Math` functions that give an infinity or `NaN`:

```ant
Math.isFinite(Math.maxFloat64 * 2.0) // false
Math.isFinite(1.0 / 0.0)             // false
Math.isFinite(Math.log(0.0))         // false
Math.isFinite(Math.sqrt(-1.0))       // false
```

A `float32` that overflows becomes an infinity, which stays one when converted to `float64`:

```ant
let big = float32(1e39)
Math.isFinite(float64(big)) // false
```

Pass `Math.isFinite` as a function value, or call it in a lambda on a record field:

```ant
[1.0, Math.NaN, 2.0, -Math.Infinity].filter(Math.isFinite) // [1.0, 2.0]
type Reading = { sensor: string, value: float64 }
let readings = [Reading{ sensor: "a", value: 20.5 }, Reading{ sensor: "b", value: Math.NaN }]
readings.filter({ !Math.isFinite(x.value) }).map({ x.sensor }) // ["b"]
```

In a pipe, the value is the argument:

```ant
2.0 | Math.isFinite() // true
```

## Applications

Reject a parsed number that is not finite. `String.toFloat` reads `"inf"` and `"NaN"` as numbers:

```ant
func parseAmount(s: string): float64 | null {
  let n = s.toFloat()
  if n == null || !Math.isFinite(n) {
    return null
  }
  n
}

parseAmount("12.5") // 12.5
parseAmount("inf")  // null
parseAmount("NaN")  // null
parseAmount("abc")  // null
```

Average only the usable values of a series:

```ant
func mean(xs: Array[float64]): float64 | null {
  let ok = xs.filter(Math.isFinite)
  let total = ok.sum()
  if total == null {
    return null
  }
  total / float64(ok.length())
}

mean([1.0, Math.NaN, 3.0, Math.Infinity]) // 2.0
mean([Math.NaN])                          // null
```

Write `null` in place of a value that is not finite before `Json.stringify`, which panics on `NaN`
and the infinities:

```ant
func toJson(xs: Array[float64]): string {
  let safe: Array[float64 | null] = xs.map({ if Math.isFinite(x) { x } else { null } })
  Json.stringify(safe)
}

toJson([1.5, Math.NaN, Math.Infinity]) // "[1.5,null,null]"
Json.stringify([1.5, Math.NaN])        // panic: Json.stringify: NaN has no JSON form
```

## Properties and relations

A value is finite exactly when it is neither `NaN` nor an infinity, and then subtracting it from
itself gives `0.0`:

```ant
for v in [0.0, -0.0, 1.5, Math.maxFloat64, Math.NaN, Math.Infinity, -Math.Infinity] {
  let viaNaN = !Math.isNaN(v) && v != Math.Infinity && v != -Math.Infinity
  assert(Math.isFinite(v) == viaNaN)
  assert(Math.isFinite(v) == (v - v == 0.0))
}
```

A finite value with no fraction converts to `int` when it is in range; [Math.toInt](toInt.md) gives
`null` for every value that is not finite:

```ant
Math.toInt(Math.Infinity) // null
Math.toInt(Math.NaN)      // null
Math.toInt(42.0)          // 42
```

## Possible issues

Comparing with `Math.NaN` does not find `NaN`, since `NaN` equals nothing, itself included. Use
`Math.isFinite` or [Math.isNaN](isNaN.md):

```ant
let v = 0.0 / 0.0
v == Math.NaN    // false
Math.isNaN(v)    // true
Math.isFinite(v) // false
```

An `int` argument is a compile error, since numeric types never mix:

```ant
let ok = Math.isFinite(1) // error: cannot use int where float64 is expected
```

So is a `float32`; convert it with `float64` first:

```ant
let f = float32(1.5)
let ok = Math.isFinite(f) // error: cannot use float32 where float64 is expected
```

Not finite is not the same as infinite: `!Math.isFinite(v)` is also `true` for `NaN`. To test only
for an infinity, compare [Math.abs](abs.md) with `Math.Infinity`:

```ant
!Math.isFinite(Math.NaN)                  // true
Math.abs(Math.NaN) == Math.Infinity       // false
Math.abs(-Math.Infinity) == Math.Infinity // true
```

Converting a value that is not finite to an integer panics:

```ant
int(Math.Infinity) // panic: Infinity does not convert to int
```

## See also

[Math.isNaN](isNaN.md) · [Math.toInt](toInt.md) · [Math.clamp](clamp.md) · [Math.abs](abs.md) ·
[String.toFloat](../String/toFloat.md) · [Array.sum](../Array/sum.md) ·
[Json.stringify](../Json/stringify.md) · [float](../../language/types/float.md)
