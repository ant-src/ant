<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.round

```ant signature
func Math.round(x: float64): float64
```

Math.round returns x rounded to the nearest whole number, with halves rounded away from zero.

<!-- end generated -->

## Details

- The result is a `float64`, not an `int`. Convert it with [int](../prelude/int.md), which panics
  when the value is out of range, or with [Math.toInt](toInt.md), which returns `null` instead.
- A half goes away from zero: `2.5` gives `3.0` and `-2.5` gives `-3.0`. It is not rounded to even.

## Basic examples

Round to the nearest whole number:

```ant
Math.round(2.4) // 2.0
Math.round(2.6) // 3.0
```

A half goes away from zero, on both sides:

```ant
Math.round(2.5)  // 3.0
Math.round(-2.5) // -3.0
```

Convert the result to get an `int`:

```ant
int(Math.round(7.5)) // 8
```

## Scope

Halves go away from zero at every size, even when the even neighbor is closer to zero:

```ant
[0.5, 1.5, 2.5, -0.5].map(Math.round) // [1.0, 2.0, 3.0, -1.0]
```

A negative number that rounds to zero keeps its sign, and `-0.0` equals `0.0`:

```ant
Math.round(-0.4)        // -0.0
Math.round(-0.4) == 0.0 // true
```

The largest `float64` below `0.5` rounds down, as it is not a half:

```ant
Math.round(0.49999999999999994) // 0.0
```

`NaN` and the infinities come back unchanged:

```ant
Math.round(Math.NaN)       // NaN
Math.round(-Math.Infinity) // -Infinity
```

Every `float64` from 2^52 up is already whole, so a large number comes back as it is:

```ant
Math.round(4503599627370497.0) // 4503599627370497.0
Math.round(1e300)              // 1e+300
```

The parameter is named `x`, and the function works in a pipe:

```ant
Math.round(x: 1.5) // 2.0
3.7 | Math.round() // 4.0
```

## Applications

Round ratings to whole stars:

```ant
let ratings = [4.2, 3.5, 4.8, 2.49]
ratings.map({ int(Math.round(x)) }) // [4, 4, 5, 2]
```

Round to a number of decimals by scaling first:

```ant
func roundTo(x: float64, decimals: int): float64 {
  let scale = Math.pow(10.0, float64(decimals))
  Math.round(x * scale) / scale
}

roundTo(3.14159, 2) // 3.14
roundTo(2.675, 2)   // 2.68
roundTo(1234.5, -2) // 1200.0
```

Report a percentage as a whole number:

```ant
type Poll = { yes: int, total: int }

func percent(p: Poll): int {
  int(Math.round(float64(p.yes) * 100.0 / float64(p.total)))
}

let polls = [Poll{ yes: 1, total: 3 }, Poll{ yes: 2, total: 3 }, Poll{ yes: 1, total: 8 }]
for p in polls {
  println("${percent(p)}%")
}
```

```text
33%
67%
13%
```

## Properties and relations

The result is never more than one half away from `x`:

```ant
for x in [2.4, 2.5, -2.5, -0.4, 7.0] {
  assert(Math.abs(Math.round(x) - x) <= 0.5)
}
```

Rounding is symmetric around zero:

```ant
for x in [0.5, 1.4, 2.5, 3.6] {
  assert(Math.round(-x) == -Math.round(x))
}
```

The result is always between [Math.floor](floor.md) and [Math.ceil](ceil.md):

```ant
let x = -3.7
assert(Math.floor(x) <= Math.round(x) && Math.round(x) <= Math.ceil(x))
```

## Possible issues

The argument must be a `float64`. An `int` is a compile error, since numeric types never mix:

```ant
let n = 7
Math.round(n) // error: cannot use int where float64 is expected
```

Integer division has already dropped the fraction, so convert before dividing:

```ant
7 / 2                        // 3
Math.round(float64(7) / 2.0) // 4.0
```

A decimal that looks like a half may sit just below it in binary, and rounds down:

```ant
1.005 * 100.0             // 100.49999999999999
Math.round(1.005 * 100.0) // 100.0
```

Adding `0.5` and taking [Math.floor](floor.md) is not the same: the sum can round up before
`Math.floor` sees it.

```ant
Math.floor(0.49999999999999994 + 0.5) // 1.0
Math.round(0.49999999999999994)       // 0.0
```

`Math.round` and [String.fromFloat](../String/fromFloat.md) round halves differently: `fromFloat`
rounds half to even.

```ant
Math.round(2.5)          // 3.0
String.fromFloat(2.5, 0) // "2"
```

A whole `float64` may still be too large for an `int`, so `int(...)` panics on it. Use
[Math.toInt](toInt.md) to get `null` instead:

```ant
int(Math.round(1e20))        // panic: 100000000000000000000.0 out of range for int
Math.toInt(Math.round(1e20)) // null
```

## See also

[Math.floor](floor.md) · [Math.ceil](ceil.md) · [Math.trunc](trunc.md) · [Math.toInt](toInt.md) ·
[String.fromFloat](../String/fromFloat.md) · [int](../prelude/int.md)
