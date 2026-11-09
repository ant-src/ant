<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math

Math holds numeric functions and constants. The functions over float64 follow IEEE 754 and never panic: Math.log(0.0) is -Infinity, and Math.sqrt(-1.0) is NaN.

<!-- end generated -->

## Details

- Most functions take and return `float64` only. Numeric types never mix, so an `int` argument is a
  compile error: convert it with `float64(n)`. `Math.abs`, `Math.min`, `Math.max`, and `Math.clamp`
  are generic and keep their argument's type.
- The `float64` functions follow IEEE 754 and never panic: outside its domain a function returns
  `NaN` or an infinity, as `Math.sqrt(-1.0)` is `NaN` and `Math.log(0.0)` is `-Infinity`. Three
  functions do panic: `Math.abs` on the most negative value of an integer type, `Math.clamp` when
  `lo > hi` or a bound is `NaN`, and `Math.randomInt` when `to <= from`.
- `NaN` is not equal to anything, itself included. Test for it with [Math.isNaN](isNaN.md).
- Angles are in radians.
- [Math.random](random.md) and [Math.randomInt](randomInt.md) cannot be seeded. Use
  [Random](../Random/index.md) for numbers you can repeat, and
  [Crypto.randomBytes](../Crypto/randomBytes.md) for keys and tokens.

## Basic examples

Call a function through the namespace:

```ant
Math.sqrt(2.0)      // 1.4142135623730951
Math.pow(2.0, 10.0) // 1024.0
Math.abs(-3)        // 3
Math.max(2, 7)      // 7
```

Round a `float64` four ways:

```ant
Math.floor(-2.5) // -3.0
Math.ceil(-2.5)  // -2.0
Math.round(-2.5) // -3.0
Math.trunc(-2.5) // -2.0
```

The constants are values of the namespace:

```ant
Math.PI     // 3.141592653589793
Math.E      // 2.718281828459045
Math.maxInt // 9223372036854775807
```

## Scope

`Math.abs`, `Math.min`, `Math.max`, and `Math.clamp` take any one numeric type, and the result has
it:

```ant
Math.abs(-2.5)        // 2.5
Math.abs(int8(-5))    // 5
Math.clamp(15, 0, 10) // 10
```

`Math.min`, `Math.max`, and `Math.clamp` also take strings, which they compare as `<` does:

```ant
Math.min("pear", "apple") // "apple"
Math.clamp("m", "a", "f") // "f"
```

An `int` goes through `float64` on its way into a `float64` function, and through
[Math.toInt](toInt.md) or `int` on its way back:

```ant
let total = 7
let parts = 3
Math.ceil(float64(total) / float64(parts))             // 3.0
Math.toInt(Math.ceil(float64(total) / float64(parts))) // 3
int(Math.trunc(-2.7))                                  // -2
```

Results outside a function's domain are `NaN` or an infinity:

```ant
Math.sqrt(-1.0)               // NaN
Math.log(0.0)                 // -Infinity
Math.exp(1000.0)              // Infinity
Math.maxFloat64 * 2.0         // Infinity
Math.Infinity - Math.Infinity // NaN
```

`Math.random` is uniform in `[0, 1)` and cannot be seeded:

```ant
let r = Math.random()
r >= 0.0 && r < 1.0 // true
```

Calls chain through a pipe:

```ant
16.0 | Math.sqrt() | Math.log2() // 2.0
```

## Functions by job

### Sign, size, and bounds

- [Math.abs](abs.md): the absolute value of any number
- [Math.min](min.md), [Math.max](max.md): the smaller or larger of two numbers or strings
- [Math.clamp](clamp.md): a value held within `lo ..= hi`

### Rounding

- [Math.floor](floor.md): the greatest whole number not above `x`
- [Math.ceil](ceil.md): the least whole number not below `x`
- [Math.round](round.md): the nearest whole number, halves away from zero
- [Math.trunc](trunc.md): `x` with its fraction dropped, toward zero
- [Math.toInt](toInt.md): a whole `float64` as an `int`, or `null`

### Powers, roots, and logarithms

- [Math.pow](pow.md): `x` raised to the power `y`
- [Math.sqrt](sqrt.md), [Math.cbrt](cbrt.md): the square and cube root
- [Math.hypot](hypot.md): the square root of `x² + y²`, without overflow in between
- [Math.exp](exp.md): `e` raised to the power `x`
- [Math.log](log.md), [Math.log2](log2.md), [Math.log10](log10.md): the natural, base-2, and base-10
  logarithm

### Trigonometry

- [Math.sin](sin.md), [Math.cos](cos.md), [Math.tan](tan.md): the sine, cosine, and tangent of an
  angle in radians
- [Math.asin](asin.md), [Math.acos](acos.md), [Math.atan](atan.md): the inverse functions, in
  radians
- [Math.atan2](atan2.md): the angle of the point `(x, y)`, from `-PI` to `PI`

### Testing floats

- [Math.isNaN](isNaN.md): whether `x` is `NaN`
- [Math.isFinite](isFinite.md): whether `x` is neither `NaN` nor an infinity

### Random numbers

- [Math.random](random.md): a uniformly random `float64` in `[0, 1)`
- [Math.randomInt](randomInt.md): a uniformly random `int` from `from`, inclusive, to `to`,
  exclusive

### Constants

| Constant          | Type      | Value                                           |
| ----------------- | --------- | ----------------------------------------------- |
| `Math.PI`         | `float64` | π, `3.141592653589793`                          |
| `Math.E`          | `float64` | e, `2.718281828459045`                          |
| `Math.NaN`        | `float64` | not-a-number                                    |
| `Math.Infinity`   | `float64` | positive infinity; `-Math.Infinity` is negative |
| `Math.maxFloat64` | `float64` | the largest finite `float64`                    |
| `Math.minInt`     | `int`     | -2^63, which no literal writes                  |
| `Math.maxInt`     | `int`     | 2^63 - 1                                        |
| `Math.maxUint64`  | `uint64`  | 2^64 - 1, which no literal writes               |

### Type bounds

- [Number](types/Number.md): any integer or floating-point type
- [Integer](types/Integer.md): any integer type
- [Ordered](types/Ordered.md): any number type or `string`, the types `<` compares

## Applications

Convert degrees to radians before a trigonometric call:

```ant
func radians(degrees: float64): float64 {
  degrees * Math.PI / 180.0
}
Math.sin(radians(90.0)) // 1.0
```

Measure the distance between two points:

```ant
type Point = { x: float64, y: float64 }
func distance(a: Point, b: Point): float64 {
  Math.hypot(b.x - a.x, b.y - a.y)
}
distance(Point{ x: 1.0, y: 1.0 }, Point{ x: 4.0, y: 5.0 }) // 5.0
```

Round to two decimal places:

```ant
let x = 3.14159
Math.round(x * 100.0) / 100.0 // 3.14
```

Grow a deposit by compound interest:

```ant
let deposit = 1000.0
let rate = 0.05
deposit * Math.pow(1.0 + rate, 10.0) // 1628.8946267774415
```

Keep a percentage within range:

```ant
let percents = [-12, 40, 130]
percents.map({ Math.clamp(x, 0, 100) }) // [0, 40, 100]
```

## Properties and relations

`Math.floor` and `Math.ceil` bracket `x`, and `Math.trunc` is one of them by sign:

```ant
for x in [2.5, -2.5, 3.0] {
  assert(Math.floor(x) <= x && x <= Math.ceil(x))
  assert(Math.trunc(x) == (if x < 0.0 { Math.ceil(x) } else { Math.floor(x) }))
}
```

`Math.clamp` is `Math.max` then `Math.min`:

```ant
for x in [-5, 3, 99] {
  assert(Math.clamp(x, 0, 10) == Math.min(Math.max(x, 0), 10))
}
```

Outside its domain a function gives `NaN` or an infinity, never a panic:

```ant
assert(Math.isNaN(Math.sqrt(-1.0)))
assert(Math.log(0.0) == -Math.Infinity)
```

## Possible issues

An `int` literal is not a `float64`. Write `2.0`, or convert a variable with `float64`:

```ant
let n = 2
Math.sqrt(n) // error: cannot use int where float64 is expected
```

The two arguments of `Math.max` must share one type:

```ant
Math.max(2, 3.5) // error: cannot use float64 where int is expected
```

`NaN` is not equal to itself, so `==` never finds it:

```ant
let x = Math.sqrt(-1.0)
x == Math.NaN // false
Math.isNaN(x) // true
```

`Math.minInt` has no absolute value in an `int`:

```ant
Math.abs(Math.minInt) // panic: Math.abs: integer overflow
```

The integer constants are limits, so arithmetic past them panics:

```ant
Math.maxInt + 1 // panic: integer overflow
```

`Math.clamp` with the bounds swapped panics:

```ant
Math.clamp(5, 10, 0) // panic: Math.clamp needs lo <= hi, got 10 and 0
```

## See also

- [Bits](../Bits/index.md): bitwise operations and wrapping arithmetic on integers
- [Random](../Random/index.md): a seeded generator of repeatable random numbers
- [String.toFloat](../String/toFloat.md), [String.fromFloat](../String/fromFloat.md): numbers to and
  from text
- [Array.sum](../Array/sum.md), [Array.min](../Array/min.md), [Array.max](../Array/max.md): the same
  jobs over an array
