<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.isNaN

```ant signature
func Math.isNaN(x: float64): bool
```

Math.isNaN reports whether x is NaN.

<!-- end generated -->

## Details

- `NaN` is not equal to anything, itself included, so `==` cannot find it.
- Only `NaN` gives `true`. The infinities are not `NaN`; [Math.isFinite](isFinite.md) rejects both.

## Basic examples

```ant
Math.isNaN(Math.NaN) // true
Math.isNaN(1.5)      // false
```

`NaN` comes out of arithmetic that has no numeric answer:

```ant
Math.isNaN(0.0 / 0.0)                     // true
Math.isNaN(Math.Infinity - Math.Infinity) // true
Math.isNaN(Math.sqrt(-1.0))               // true
```

## Scope

Infinities, zeros, and the extremes are numbers, not `NaN`:

```ant
Math.isNaN(Math.Infinity)   // false
Math.isNaN(-Math.Infinity)  // false
Math.isNaN(-0.0)            // false
Math.isNaN(Math.maxFloat64) // false
```

`NaN` spreads through arithmetic, so one `NaN` input makes the whole result `NaN`:

```ant
let total = 1.0 + Math.NaN * 0.0
Math.isNaN(total) // true
```

`String.toFloat` reads `NaN` as a value, not as `null`:

```ant
String.toFloat("NaN")                    // NaN
Math.isNaN(String.toFloat("nan") ?? 0.0) // true
String.toFloat("abc")                    // null
```

A `float32` goes through `float64`, which keeps `NaN`:

```ant
let f = float32(0.0) / float32(0.0)
Math.isNaN(float64(f)) // true
```

Pass `Math.isNaN` as a function value, or call it in a pipe or with its parameter name `x`:

```ant
[1.0, Math.NaN, 3.0].map(Math.isNaN) // [false, true, false]
Math.NaN | Math.isNaN()              // true
Math.isNaN(x: 2.0)                   // false
```

## Applications

Drop missing readings before averaging them:

```ant
let readings = [12.5, Math.NaN, 13.5, Math.NaN]
let valid = readings.filter({ !Math.isNaN(x) })
valid                                          // [12.5, 13.5]
(valid.sum() ?? 0.0) / float64(valid.length()) // 13.0
```

Fill gaps in a series with a default value:

```ant
func orDefault(x: float64, d: float64): float64 {
  if Math.isNaN(x) { d } else { x }
}

[12.5, Math.NaN, 13.5].map({ orDefault(x, 0.0) }) // [12.5, 0.0, 13.5]
```

Report a computation that went wrong instead of returning `NaN`:

```ant
func ratio(a: float64, b: float64): Result[float64, string] {
  let r = a / b
  if Math.isNaN(r) {
    return Result.err("${a} / ${b} is undefined")
  }
  Result.ok(r)
}

ratio(1.0, 4.0) // Ok{ value: 0.25 }
ratio(0.0, 0.0) // Err{ error: "0.0 / 0.0 is undefined" }
```

## Properties and relations

`NaN` is the only `float64` not equal to itself, so for a `float64` `x`, `Math.isNaN(x)` is
`x != x`:

```ant
for x in [0.0, -0.0, 1.5, Math.Infinity, Math.NaN] {
  assert(Math.isNaN(x) == (x != x))
}
```

A `NaN` is never finite, and a value that is neither finite nor `NaN` is an infinity:

```ant
for x in [0.0, 1e308, -Math.Infinity, Math.NaN] {
  assert(!(Math.isNaN(x) && Math.isFinite(x)))
  if !Math.isNaN(x) && !Math.isFinite(x) {
    assert(Math.abs(x) == Math.Infinity)
  }
}
```

## Possible issues

Comparing with `Math.NaN` never finds it, since `NaN` is not equal to itself:

```ant
let x = 0.0 / 0.0
x == Math.NaN // false
Math.isNaN(x) // true
```

Functions that compare by `==` miss `NaN` the same way:

```ant
let xs = [1.0, Math.NaN]
xs.contains(Math.NaN)      // false
xs.indexOf(Math.NaN)       // null
xs.some({ Math.isNaN(x) }) // true
```

Every ordering comparison with `NaN` (`<`, `<=`, `>`, `>=`) is `false`, so a check for values out of
range lets it through. Test for `NaN` too:

```ant
let x = Math.NaN
x < 0.0 || x > 100.0                  // false
x < 0.0 || x > 100.0 || Math.isNaN(x) // true
```

`NaN` is one input a conversion to `int` panics on, but an infinity or a value out of range panics
too, so ruling out `NaN` does not make the conversion safe. [Math.toInt](toInt.md) returns `null`
for all of them:

```ant
Math.toInt(Math.NaN)      // null
Math.toInt(Math.Infinity) // null
Math.toInt(1e30)          // null
Math.toInt(3.0)           // 3
```

An integer argument is a compile error:

```ant
let bad = Math.isNaN(1) // error: cannot use int where float64 is expected
```

## See also

[Math.isFinite](isFinite.md) · [Math.toInt](toInt.md) · [String.toFloat](../String/toFloat.md) ·
[float64](../prelude/float64.md)
