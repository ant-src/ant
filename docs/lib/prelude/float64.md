<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# float64

```ant signature
func float64[T: Number](x: T): float64
```

float64 converts x to float64, rounding to the nearest value it holds; it never panics. x may also
be a named primitive over a number.

<!-- end generated -->

## Details

- An integer above 2^53 in magnitude may not fit exactly; it rounds to the nearest `float64`, ties
  to even.
- A `float32` widens exactly; `NaN`, the infinities, and `-0.0` are kept.

## Basic examples

Convert an `int` to divide with a fraction:

```ant
float64(3)       // 3.0
float64(7) / 2.0 // 3.5
let n = 7
float64(n) / 2.0 // 3.5
```

The average of an array of integers:

```ant
let scores = [3, 4, 6]
float64(scores.sum() ?? 0) / float64(len(scores)) // 4.333333333333333
```

## Scope

Every numeric type converts, the sized ones included:

```ant
float64(byte(200))    // 200.0
float64(int8(-5))     // -5.0
float64(uint16(500))  // 500.0
float64(float32(1.5)) // 1.5
float64(2.5)          // 2.5
```

The largest integers round to the nearest `float64`: `Math.maxInt` to 2^63 and `Math.maxUint64` to
2^64.

```ant
float64(Math.maxInt)    // 9223372036854776000.0
float64(Math.minInt)    // -9223372036854776000.0
float64(Math.maxUint64) // 18446744073709552000.0
```

A `float32` widens to the same binary value, which is not the decimal it was written as:

```ant
float64(float32(0.1)) // 0.10000000149011612
```

`NaN`, the infinities, and the sign of zero stay as they are:

```ant
float64(float32(Math.NaN))       // NaN
float64(-float32(Math.Infinity)) // -Infinity
float64(float32(-0.0))           // -0.0
```

A named primitive over a number is unwrapped, and the result is a plain `float64`:

```ant
type Meters = float64
type Count = int
float64(Meters(2.5)) // 2.5
float64(Count(4))    // 4.0
```

`float64` is a function value, so it maps an array or ends a pipe:

```ant
[1, 2, 3].map(float64) // [1.0, 2.0, 3.0]
7 | float64()          // 7.0
```

## Applications

Take the square root of an integer; `Math.sqrt` takes a `float64`:

```ant
let area = 50
Math.sqrt(float64(area)) // 7.0710678118654755
```

Show a share as a percentage:

```ant
func percent(part: int, whole: int): string {
  let p = float64(part) / float64(whole) * 100.0
  "${String.fromFloat(p, 1)}%"
}

println(percent(1, 3))
println(percent(7, 8))
```

```text
33.3%
87.5%
```

A generic mean over any numeric array converts each element:

```ant
func mean[T: Number](xs: Array[T]): float64 {
  var total = 0.0
  for x in xs {
    total += float64(x)
  }
  total / float64(len(xs))
}

println(mean([1, 2, 4]))
println(mean([byte(1), byte(2)]))
println(mean([0.5, 1.5]))
```

```text
2.3333333333333335
1.5
1.0
```

## Properties and relations

An integer up to 2^53 in magnitude survives the round trip through `float64`:

```ant
for n in [0, -17, 9007199254740992, -9007199254740992] {
  assert(int(float64(n)) == n)
}
```

A `float32` survives the round trip through `float64`:

```ant
for x in [float32(0.1), float32(-2.75), float32(Math.Infinity)] {
  assert(float32(float64(x)) == x)
}
```

On a `float64`, it is the identity:

```ant
for x in [2.5, -0.0, Math.Infinity, Math.maxFloat64] {
  assert(float64(x) == x)
}
```

## Possible issues

An `int` and a `float64` never mix in arithmetic. Convert one side:

```ant
let n = 7
println(n / 2.0) // error: / needs both operands of one type, got int and float64
```

```ant
let n = 7
float64(n) / 2.0 // 3.5
```

An integer literal is an `int`, never a `float64`. Write `3.0`:

```ant
let x: float64 = 3 // error: cannot use int where float64 is expected
```

Above 2^53, distinct integers can convert to the same `float64`:

```ant
float64(9007199254740993) == float64(9007199254740992) // true
```

`float64(Math.maxInt)` rounds up to 2^63, which no `int` holds, so converting back panics:

```ant
int(float64(Math.maxInt)) // panic: 9223372036854776000.0 out of range for int
```

[`Math.toInt`](../Math/toInt.md) returns `null` instead of panicking:

```ant
Math.toInt(float64(Math.maxInt)) ?? 0 // 0
```

Converting back to an integer is exact, so a fraction panics. Round first with
[`Math.round`](../Math/round.md) or [`Math.trunc`](../Math/trunc.md):

```ant
let ratio = float64(7) / 2.0
int(ratio)             // panic: 3.5 has a fraction, so it does not convert to int
int(Math.trunc(ratio)) // 3
int(Math.round(ratio)) // 4
```

A string is not a number. Parse it with [`String.toFloat`](../String/toFloat.md), which returns
`null` on text that is not a number:

```ant
println(float64("1.5"))
// error: float64 converts a number or a named primitive type over a number, got string
```

```ant
String.toFloat("1.5") ?? 0.0 // 1.5
```

A union such as `int | null` is not a number. Narrow it first:

```ant
let n: int | null = 3
println(float64(n))
// error: float64 converts a number or a named primitive type over a number, got int | null
```

```ant
let n: int | null = 3
float64(n ?? 0) // 3.0
```

## See also

[float32](float32.md) · [int](int.md) · [Math.trunc](../Math/trunc.md) ·
[Math.round](../Math/round.md) · [Math.toInt](../Math/toInt.md) ·
[String.toFloat](../String/toFloat.md) · [String.fromFloat](../String/fromFloat.md) ·
[float types](../../language/types/float.md) · [Number literals](../../language/literals/numbers.md)
· [Arithmetic operators](../../language/operators/arithmetic.md)
