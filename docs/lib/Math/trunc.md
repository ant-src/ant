<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.trunc

```ant signature
func Math.trunc(x: float64): float64
```

Math.trunc returns x with its fraction dropped, rounding toward zero. int(Math.trunc(x)) is the
explicit truncating conversion to int.

<!-- end generated -->

## Details

- The result is a `float64`, not an `int`. `int(Math.trunc(x))` turns it into an `int`; a plain
  `int(x)` panics when `x` has a fraction.
- It never panics. `NaN` and the infinities come back unchanged, and a negative number above `-1.0`
  gives `-0.0`, which equals `0.0`.

## Basic examples

Drop the fraction, on either side of zero:

```ant
Math.trunc(2.7)  // 2.0
Math.trunc(-2.7) // -2.0
Math.trunc(5.0)  // 5.0
```

Convert the result to get an `int`:

```ant
int(Math.trunc(9.99)) // 9
```

## Scope

A negative number between `-1.0` and `0.0` gives `-0.0`, which equals `0.0`:

```ant
Math.trunc(-0.5)        // -0.0
Math.trunc(-0.5) == 0.0 // true
```

`NaN` and the infinities come back as they are:

```ant
Math.trunc(Math.NaN)       // NaN
Math.trunc(-Math.Infinity) // -Infinity
```

Every `float64` from 2^52 up is already whole, so a large number comes back unchanged:

```ant
Math.trunc(1e300) // 1e+300
```

The parameter is named `x`, and the function works in a pipe:

```ant
Math.trunc(x: -1.5) // -1.0
3.99 | Math.trunc() // 3.0
```

`Math.trunc` is a function value, so it can be passed to `Array.map`:

```ant
[1.9, -1.9, 0.5].map(Math.trunc) // [1.0, -1.0, 0.0]
```

## Applications

Convert a measured float to an `int`, dropping the fraction:

```ant
let readings = [12.8, 7.2, -3.6]
readings.map({ int(Math.trunc(x)) }) // [12, 7, -3]
```

Split a number into its whole part and its fraction, which keeps the sign of `x`:

```ant
type Parts = { whole: float64, fraction: float64 }

func split(x: float64): Parts {
  let whole = Math.trunc(x)
  Parts{ whole, fraction: x - whole }
}

split(3.25)  // Parts{ whole: 3.0, fraction: 0.25 }
split(-3.25) // Parts{ whole: -3.0, fraction: -0.25 }
```

Show a duration in seconds as whole minutes and seconds:

```ant
func clock(seconds: float64): string {
  let total = int(Math.trunc(seconds))
  "${total / 60}:${String.from(total % 60).padStart(2, "0")}"
}

clock(125.9) // "2:05"
clock(59.99) // "0:59"
```

Cut a value to a number of decimals, toward zero on both sides:

```ant
func cut(x: float64, decimals: int): float64 {
  let scale = Math.pow(10.0, float64(decimals))
  Math.trunc(x * scale) / scale
}

cut(3.14159, 2)  // 3.14
cut(-3.14159, 2) // -3.14
```

## Properties and relations

The result is whole, has the sign of `x`, and is less than one away from it:

```ant
for x in [2.7, -2.7, 0.5, 5.0, -9.99] {
  let t = Math.trunc(x)
  assert(Math.trunc(t) == t)
  assert((t >= 0.0) == (x >= 0.0) || t == 0.0)
  assert(Math.abs(x - t) < 1.0 && Math.abs(t) <= Math.abs(x))
}
```

It is [Math.floor](floor.md) for zero and above and [Math.ceil](ceil.md) below zero:

```ant
for x in [2.5, -2.5, 0.0, 7.0] {
  let expected = if x >= 0.0 { Math.floor(x) } else { Math.ceil(x) }
  assert(Math.trunc(x) == expected)
}
```

It is symmetric around zero:

```ant
for x in [0.5, 1.4, 2.5, 3.6] {
  assert(Math.trunc(-x) == -Math.trunc(x))
}
```

For whole operands, it agrees with integer `/`, which also truncates toward zero:

```ant
for a in [7, -7, 9, -9] {
  assert(int(Math.trunc(float64(a) / 2.0)) == a / 2)
}
```

## Possible issues

`int(x)` does not truncate. It panics when `x` has a fraction:

```ant
int(2.7)             // panic: 2.7 has a fraction, so it does not convert to int
int(Math.trunc(2.7)) // 2
```

The argument must be a `float64`. An `int` is a compile error, since numeric types never mix:

```ant
let n = 7
Math.trunc(n) // error: cannot use int where float64 is expected
```

A product that looks whole can sit just below the whole number in binary, so `Math.trunc` gives one
less than you expect:

```ant
4.35 * 100.0             // 434.99999999999994
Math.trunc(4.35 * 100.0) // 434.0
```

To round to the nearest whole number instead, use [Math.round](round.md):

```ant
Math.trunc(2.9) // 2.0
Math.round(2.9) // 3.0
```

A whole `float64` may still be too large for an `int`, so `int(...)` panics on it. Use
[Math.toInt](toInt.md) to get `null` instead:

```ant
int(Math.trunc(1e20))        // panic: 100000000000000000000.0 out of range for int
Math.toInt(Math.trunc(1e20)) // null
```

## See also

[Math.floor](floor.md) · [Math.ceil](ceil.md) · [Math.round](round.md) · [Math.toInt](toInt.md) ·
[int](../prelude/int.md) · [Arithmetic operators](../../language/operators/arithmetic.md)
