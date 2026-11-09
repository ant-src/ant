<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.pow

```ant signature
func Math.pow(x: float64, y: float64): float64
```

Math.pow returns x raised to the power y.

<!-- end generated -->

## Details

- Ant has no power operator: `Math.pow` is the way to raise a number to a power.
- `Math.pow` never panics, and a negative `x` with a fractional `y` gives `NaN`.
- `int(...)` converts a whole result back and panics on a fraction or a value out of range;
  [Math.toInt](toInt.md) returns `null` instead.

## Basic examples

```ant
Math.pow(2.0, 10.0) // 1024.0
Math.pow(3.0, 2.0)  // 9.0
Math.pow(1.5, 2.0)  // 2.25
```

## Scope

A negative exponent gives the reciprocal, and a fractional one gives a root:

```ant
Math.pow(2.0, -1.0)   // 0.5
Math.pow(10.0, -3.0)  // 0.001
Math.pow(9.0, 0.5)    // 3.0
Math.pow(100.0, -0.5) // 0.1
```

A negative base works with a whole exponent, and the sign follows its parity. With a fractional
exponent the result is `NaN`:

```ant
Math.pow(-2.0, 3.0) // -8.0
Math.pow(-2.0, 2.0) // 4.0
Math.pow(-8.0, 0.5) // NaN
```

Zero, `NaN`, and the infinities follow IEEE 754:

```ant
Math.pow(0.0, 0.0)            // 1.0
Math.pow(Math.NaN, 0.0)       // 1.0
Math.pow(1.0, Math.NaN)       // 1.0
Math.pow(Math.NaN, 1.0)       // NaN
Math.pow(0.0, -1.0)           // Infinity
Math.pow(-0.0, -1.0)          // -Infinity
Math.pow(2.0, Math.Infinity)  // Infinity
Math.pow(0.5, Math.Infinity)  // 0.0
Math.pow(Math.Infinity, -1.0) // 0.0
```

Name the arguments:

```ant
Math.pow(x: 2.0, y: 8.0) // 256.0
```

In a pipe, the value is `x`, the base:

```ant
2.0 | Math.pow(3.0) // 8.0
```

Square every element of an array with a lambda:

```ant
[1.0, 2.0, 3.0].map({ Math.pow(x, 2.0) }) // [1.0, 4.0, 9.0]
```

## Applications

Grow a deposit at 5% interest compounded yearly for ten years:

```ant
let principal = 1000.0
principal * Math.pow(1.05, 10.0) // 1628.8946267774415
```

Round to a number of decimal places:

```ant
func roundTo(x: float64, places: int): float64 {
  let scale = Math.pow(10.0, float64(places))
  Math.round(x * scale) / scale
}

roundTo(3.14159, 2) // 3.14
roundTo(2.71828, 3) // 2.718
```

Format a byte count with a unit, dividing by powers of 1024:

```ant
func size(bytes: int): string {
  let units = ["B", "KiB", "MiB", "GiB"]
  var i = units.length() - 1
  while i > 0 && float64(bytes) < Math.pow(1024.0, float64(i)) {
    i -= 1
  }
  let value = float64(bytes) / Math.pow(1024.0, float64(i))
  "${String.fromFloat(value, 1)} ${units[i]}"
}

println(size(512))
println(size(1536))
println(size(5000000))
```

```text
512.0 B
1.5 KiB
4.8 MiB
```

The geometric mean of some growth rates is the product to the power one over the count:

```ant
let rates = [1.1, 1.2, 0.9]
var product = 1.0
for r in rates {
  product *= r
}
Math.pow(product, 1.0 / float64(rates.length())) // 1.059104500597819
```

## Properties and relations

Exponents add when powers multiply, and a power of a power multiplies the exponents, each up to
rounding:

```ant
func close(a: float64, b: float64): bool {
  Math.abs(a - b) <= 1e-12 * Math.max(Math.abs(a), Math.abs(b))
}

for [x, a, b] in [[2.0, 3.0, 4.0], [1.5, -2.0, 0.5], [10.0, 0.25, 1.75]] {
  assert(close(Math.pow(x, a) * Math.pow(x, b), Math.pow(x, a + b)))
  assert(close(Math.pow(Math.pow(x, a), b), Math.pow(x, a * b)))
  assert(Math.pow(x, 0.0) == 1.0)
  assert(Math.pow(x, 1.0) == x)
}
```

The power `0.5` gives the square root, as [Math.sqrt](sqrt.md) does, up to rounding, and `Math.log`
takes the exponent back out:

```ant
for x in [2.0, 3.0, 10.0, 0.25] {
  assert(Math.abs(Math.pow(x, 0.5) - Math.sqrt(x)) <= 1e-15 * Math.sqrt(x))
  assert(Math.abs(Math.log(Math.pow(x, 3.0)) / Math.log(x) - 3.0) < 1e-12)
}
```

## Possible issues

`1.0 / 3.0` is not exactly a third, so `Math.pow` misses the cube root by the last digit, and gives
`NaN` for a negative base. Use [Math.cbrt](cbrt.md):

```ant
Math.pow(8.0, 1.0 / 3.0)  // 1.9999999999999998
Math.pow(-8.0, 1.0 / 3.0) // NaN
Math.cbrt(-8.0)           // -2.0
```

The result overflows to `Infinity` and underflows to `0.0`:

```ant
Math.pow(10.0, 309.0)  // Infinity
Math.pow(10.0, -400.0) // 0.0
```

A `float64` holds integers exactly only up to 2⁵³, so a large integer power loses its last digits.
Multiply in `int` when you need them all:

```ant
int(Math.pow(3.0, 39.0)) // 4052555153018976256
var p = 1
for _ in Iter.range(0, 39) {
  p *= 3
}
p // 4052555153018976267
```

Converting a result past `Math.maxInt` back to `int` panics:

```ant
int(Math.pow(2.0, 63.0)) // panic: 9223372036854776000.0 out of range for int
```

A fractional result panics too:

```ant
int(Math.pow(10.0, -1.0)) // panic: 0.1 has a fraction, so it does not convert to int
```

Use [Math.toInt](toInt.md) to get `null` instead, or round first:

```ant
Math.toInt(Math.pow(10.0, -1.0))    // null
int(Math.round(Math.pow(2.0, 0.5))) // 1
```

The arguments are `float64`. An `int` is a compile error; convert it first:

```ant
let n = 10
let x = Math.pow(2.0, n) // error: cannot use int where float64 is expected
```

```ant
let n = 10
Math.pow(2.0, float64(n)) // 1024.0
```

Ant has no `**` or `^` operator, and `Math.pow` is not a function on `float64`, so it cannot be
called through `.`:

```ant
let x = 2.0.pow(3.0) // error: float64 has no field or function pow
```

## See also

[Math.sqrt](sqrt.md) · [Math.cbrt](cbrt.md) · [Math.exp](exp.md) · [Math.log](log.md) ·
[Math.hypot](hypot.md) · [Math](index.md)
