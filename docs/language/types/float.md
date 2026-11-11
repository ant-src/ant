<!-- status: reviewed -->

# float32 and float64

`float64` and `float32` are the IEEE 754 floating-point types, of 64 and 32 bits. `float64` is the
float type of literals and of the `Math` functions; `float32` is for when the size matters.

## Details

- A float literal has digits on both sides of the point, an exponent, or both, and is always a
  `float64`. `1` is an `int`; write `1.0`.
- Numeric types never mix. `1 + 1.0` and `float32(1.5) + 2.0` are compile errors; convert with
  `float64(x)` or `float32(x)`.
- Float arithmetic never panics: dividing by zero or overflowing gives an infinity or `NaN`.
- A conversion to a float rounds to the nearest value it can hold and never panics. A conversion
  from a float to an integer type must be exact: a fraction, a value out of range, `NaN`, or an
  infinity panics.
- `NaN` is not equal to anything, itself included, and every `<`, `<=`, `>`, `>=` with a `NaN`
  operand is `false`. `-0.0 == 0.0` is `true`.

## Basic examples

Float literals are `float64`:

```ant
3.14      // 3.14
1e9       // 1000000000.0
2.5e-3    // 0.0025
7.0 / 2.0 // 3.5
```

Convert an integer to divide it as a float:

```ant
let n = 7
float64(n) / 2.0 // 3.5
```

A `float32` comes from a conversion:

```ant
float32(1.5) * float32(2.0) // 3.0
```

## Scope

### Rendering

A float renders as the shortest decimal that reads back to the same value, with `.0` added when it
is integral. A nonzero magnitude at or above `1e21` or below `1e-6` takes the exponent form, with no
`.0`:

```ant
100.0     // 100.0
0.000001  // 0.000001
1e-7      // 1e-7
1e21      // 1e+21
2.0 / 3.0 // 0.6666666666666666
```

The same rule applies to a `float32`: it renders as the shortest decimal that reads back to the same
`float32`:

```ant
float32(2.0) / float32(3.0) // 0.6666667
float32(0.1)                // 0.1
float64(float32(0.1))       // 0.10000000149011612
```

### Conversions between floats

`float32(x)` rounds to the nearest `float32`. A value beyond its range becomes an infinity, and one
below its smallest becomes a zero:

```ant
float32(16777217.0) // 16777216.0
float32(1e300)      // Infinity
float32(1e-300)     // 0.0
```

### Conversions to integers

`int(x)` panics unless `x` is a whole number in range. Truncate first to drop the fraction, or use
`Math.toInt` to get `null` instead of a panic:

```ant
int(4.0)             // 4
int(Math.trunc(3.7)) // 3
int(Math.round(3.7)) // 4
Math.toInt(3.7)      // null
int(3.7)             // panic: 3.7 has a fraction, so it does not convert to int
```

### Infinities and `NaN`

`Math.Infinity` and `Math.NaN` name the special values, and `-Math.Infinity` is the negative
infinity; `Math.maxFloat64` is the largest finite `float64`. Arithmetic reaches the special values
without a panic:

```ant
1.0 / 0.0                              // Infinity
-1.0 / 0.0                             // -Infinity
0.0 / 0.0                              // NaN
Math.maxFloat64 * 2.0                  // Infinity
Math.maxFloat64 * 2.0 == Math.Infinity // true
-Math.Infinity                         // -Infinity
Math.maxFloat64                        // 1.7976931348623157e+308
Math.isNaN(Math.NaN)                   // true
Math.isFinite(1.0 / 0.0)               // false
```

`Array.sort` puts `NaN` after every number, and `Math.min` and `Math.max` return `NaN` when either
operand is `NaN`:

```ant
[3.0, Math.NaN, 1.0].sort() // [1.0, 3.0, NaN]
Math.min(1.0, Math.NaN)     // NaN
```

### Remainder

`%` works on floats too. The remainder has the sign of the dividend:

```ant
7.5 % 2.0  // 1.5
-7.5 % 2.0 // -1.5
```

### Negative zero

`-0.0` keeps its sign when rendered and equals `0.0`. `Math.min` and `Math.max` order it below
`0.0`:

```ant
-0.0                // -0.0
-0.0 == 0.0         // true
Math.min(0.0, -0.0) // -0.0
```

### Parsing and formatting

`String.toFloat` parses a decimal, and `String.fromFloat` writes a fixed number of decimals:

```ant
"1.5".toFloat()                // 1.5
"abc".toFloat()                // null
"-0".toFloat()                 // -0.0
String.fromFloat(2.0 / 3.0, 3) // "0.667"
```

### Generic functions

A type parameter bounded by `Number` accepts both float types, and the integers too:

```ant
func double[T: Number](x: T): T { x + x }
double(1.25)          // 2.5
double(float32(1.25)) // 2.5
double(4)             // 8
```

## Applications

### Average of a list

Convert the count to divide the sum. An empty list would give `0.0 / 0.0`, which is `NaN`, so guard
it first:

```ant
func average(xs: Array[float64]): float64 | null {
  if xs.isEmpty() {
    return null
  }
  (xs.sum() ?? 0.0) / float64(xs.length())
}

println(average([7.5, 9.0, 8.25]))
println(average([]))
```

```text
8.25
null
```

### Money for display

Round only when you render, with `String.fromFloat`:

```ant
let price = 19.99
let total = price * 3.0
println("Total: ${String.fromFloat(total, 2)}")
```

```text
Total: 59.97
```

### Distance between points

```ant
type Point = { x: float64, y: float64 }

func Point.distance(a: Point, b: Point): float64 {
  Math.hypot(b.x - a.x, b.y - a.y)
}

let start = Point{ x: 0.0, y: 0.0 }
let end = Point{ x: 3.0, y: 4.0 }
println(start.distance(end))
```

```text
5.0
```

## Properties and relations

A conversion from `int` to `float64` and back keeps every integer up to 2^53 in magnitude. Above it,
neighboring integers round to the same float:

```ant
let n = 9007199254740992
assert(int(float64(n)) == n)
assert(float64(n + 1) == float64(n))
```

`NaN` is the one value that does not equal itself:

```ant
let x = Math.NaN
assert(x != x)
assert(Math.isNaN(x))
```

## Possible issues

### Mixing integers and floats

An integer where a float is expected, or two different float types in one operation, is a compile
error:

```ant
let price: float64 = 3
// error: cannot use int where float64 is expected
let half = 7 / 2.0
// error: / needs both operands of one type, got int and float64
let total = float32(1.5) + 2.0
// error: + needs both operands of one type, got float32 and float64
```

The `Math` functions take a `float64`, so a `float32` is converted on the way in and out:

```ant
let side = float32(2.0)
float32(Math.sqrt(float64(side))) // 1.4142135
```

### Rounding error

Most decimal fractions have no exact binary form, so arithmetic on them can be off in the last
digits. Compare with a tolerance rather than `==`:

```ant
0.1 + 0.2                        // 0.30000000000000004
0.1 + 0.2 == 0.3                 // false
Math.abs(0.1 + 0.2 - 0.3) < 1e-9 // true
```

`String.fromFloat` rounds the stored binary value, half to even, not the decimal you wrote. The
`float64` nearest `1.005` is a little below it, and `2.5` rounds to the even `2`, while `Math.round`
rounds half away from zero:

```ant
String.fromFloat(1.005, 2) // "1.00"
String.fromFloat(2.5, 0)   // "2"
Math.round(2.5)            // 3.0
```

### Testing for `NaN` with `==`

`x == Math.NaN` is `false` even when `x` is `NaN`. Use `Math.isNaN`:

```ant
let x = 0.0 / 0.0
x == Math.NaN // false
Math.isNaN(x) // true
```

### Converting a fraction to an integer

`int(x)` panics on a fraction. Say which rounding you want with `int(Math.round(x))` or
`int(Math.trunc(x))`, or use `Math.toInt` to get `null`:

```ant
let ratio = 2.5
let n = int(ratio) // panic: 2.5 has a fraction, so it does not convert to int
```

## See also

[int](int.md), [Number](../../lib/Math/types/Number.md), [float64](../../lib/prelude/float64.md),
[float32](../../lib/prelude/float32.md), [Math](../../lib/Math/index.md),
[Math.toInt](../../lib/Math/toInt.md), [Math.isNaN](../../lib/Math/isNaN.md),
[String.toFloat](../../lib/String/toFloat.md), [String.fromFloat](../../lib/String/fromFloat.md),
[number literals](../literals/numbers.md), [arithmetic operators](../operators/arithmetic.md)
