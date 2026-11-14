<!-- status: reviewed -->

# Numbers and math

Ant has ten number types: eight integer types and two float types. `int` and `float64` are the
defaults, and the others are for when the size matters. Numeric types never mix: `+`, `<`, and the
other arithmetic and comparison operators need both operands of one type, so you convert explicitly
by calling the target type's name. The library splits the rest of the work across five places: the
conversion functions of the prelude, `String` for parsing and printing numbers,
[Math](../Math/index.md) for arithmetic on numbers, [Bits](../Bits/index.md) for the bits of
integers, and [Random](../Random/index.md) for random numbers you can repeat.

```ant
let prices = [19.99, 5.25, 0.76]
let total = prices.sum() ?? 0.0
let mean = total / float64(prices.length())
println("total ${String.fromFloat(total, 2)}")
println("mean ${String.fromFloat(mean, 2)}")
```

```text
total 26.00
mean 8.67
```

## Converting between number types

Call a type's name to convert a number to it. A conversion to an integer type must be exact: a
fraction, a value out of range, `NaN`, or an infinity panics. A conversion to a float type rounds to
the nearest value it can hold and never panics.

```ant
float64(7) / 2.0      // 3.5
int8(100)             // 100
int(Math.trunc(-2.7)) // -2
int(Math.round(2.5))  // 3
Math.toInt(2.5)       // null
int(2.5)              // panic: 2.5 has a fraction, so it does not convert to int
```

- [int](../prelude/int.md), [int8](../prelude/int8.md), [int16](../prelude/int16.md),
  [int32](../prelude/int32.md), [int64](../prelude/int64.md): a number as a signed integer, exactly
- [uint8](../prelude/uint8.md), [byte](../prelude/byte.md), [uint16](../prelude/uint16.md),
  [uint32](../prelude/uint32.md), [uint64](../prelude/uint64.md): a number as an unsigned integer,
  exactly
- [float64](../prelude/float64.md), [float32](../prelude/float32.md): a number as a float, rounded
  to the nearest value
- [Math.toInt](../Math/toInt.md): a whole `float64` as an `int`, or `null` where `int` would panic
- [Math.trunc](../Math/trunc.md), [Math.round](../Math/round.md), [Math.floor](../Math/floor.md),
  [Math.ceil](../Math/ceil.md): a `float64` made whole first, so `int` does not panic on a fraction

## Parsing and printing numbers

Parsing returns `null` on text that is not a number, so the caller decides what a bad value means.
It accepts no white space around the number, so trim text read from a line, a file, or stdin first.
Printing chooses the base or the number of decimals. `String.fromFloat` rounds a half to even, while
`Math.round` rounds it away from zero.

```ant
"42".toInt()                 // 42
"ff".toInt(base: 16)         // 255
"2.5".toFloat()              // 2.5
"4 apples".toInt()           // null
" 42\n".trim().toInt()       // 42
String.fromInt(255, base: 2) // "11111111"
String.fromFloat(Math.PI, 3) // "3.142"
String.fromFloat(2.5, 0)     // "2"
String.from(0.1 + 0.2)       // "0.30000000000000004"
```

- [String.toInt](../String/toInt.md): text in a base from 2 to 36 as an `int`, or `null`
- [String.toUint](../String/toUint.md): text with no sign as a `uint64`, or `null`
- [String.toFloat](../String/toFloat.md): a decimal number, `NaN`, or an infinity as a `float64`, or
  `null`
- [String.fromInt](../String/fromInt.md): any integer as digits in a base from 2 to 36
- [String.fromFloat](../String/fromFloat.md): a `float64` in fixed point with a set number of
  decimals, rounded half to even
- [String.from](../String/from.md): any value as text, as string interpolation renders it

## Arithmetic and overflow

Integer `/` truncates toward zero, and `%` keeps the dividend's sign. Integer overflow panics, and
so does integer division by zero; float division follows IEEE 754. When you want wrapping
arithmetic, as a hash function does, ask for it with `Bits`.

```ant
7 / 2                        // 3
-7 % 2                       // -1
1.0 / 0.0                    // Infinity
Bits.wrapAdd(Math.maxInt, 1) // -9223372036854775808
Math.maxInt + 1              // panic: integer overflow
```

- [Arithmetic operators](../../language/operators/arithmetic.md): `+`, `-`, `*`, `/`, and `%`
- [Bits.wrapAdd](../Bits/wrapAdd.md), [Bits.wrapSub](../Bits/wrapSub.md),
  [Bits.wrapMul](../Bits/wrapMul.md): `a + b`, `a - b`, and `a * b`, wrapping around in two's
  complement instead of panicking
- [Array.sum](../Array/sum.md): the sum of an array of numbers, or `null` when it is empty
- [Array.range](../Array/range.md): the `int`s from `from` up to but excluding `to`

## Sign, size, and bounds

`Math.abs`, `Math.min`, `Math.max`, and `Math.clamp` take any one number type and keep it, so they
work on an `int` without a conversion. `Math.abs` of the most negative integer, such as
`Math.minInt`, panics, because its absolute value does not fit the type.

```ant
Math.abs(-3)            // 3
Math.max(2.5, 1.0)      // 2.5
Math.clamp(130, 0, 100) // 100
[4, 9, 1].max()         // 9
Math.abs(Math.minInt)   // panic: Math.abs: integer overflow
```

- [Math.abs](../Math/abs.md): the absolute value of any number
- [Math.min](../Math/min.md), [Math.max](../Math/max.md): the smaller or larger of two numbers or
  strings
- [Math.clamp](../Math/clamp.md): a value held within `lo ..= hi`
- [Array.min](../Array/min.md), [Array.max](../Array/max.md): the least or greatest element, or
  `null` when the array is empty

## Rounding

The four rounding functions take and return a `float64`. They differ on negative numbers and on
halves.

```ant
Math.floor(-2.5) // -3.0
Math.ceil(-2.5)  // -2.0
Math.round(-2.5) // -3.0
Math.trunc(-2.5) // -2.0
```

- [Math.floor](../Math/floor.md): the greatest whole number not above `x`
- [Math.ceil](../Math/ceil.md): the least whole number not below `x`
- [Math.round](../Math/round.md): the nearest whole number, halves away from zero
- [Math.trunc](../Math/trunc.md): `x` with its fraction dropped, toward zero

## Powers, roots, and logarithms

These functions take a `float64`. Convert an `int` with `float64` first.

```ant
let side = 3
Math.sqrt(float64(side * side)) // 3.0
Math.pow(2.0, 10.0)             // 1024.0
Math.hypot(3.0, 4.0)            // 5.0
Math.log2(1024.0)               // 10.0
```

- [Math.pow](../Math/pow.md): `x` raised to the power `y`
- [Math.sqrt](../Math/sqrt.md), [Math.cbrt](../Math/cbrt.md): the square and cube root
- [Math.hypot](../Math/hypot.md): the square root of `x² + y²`, without overflow in between
- [Math.exp](../Math/exp.md): `e` raised to the power `x`
- [Math.log](../Math/log.md), [Math.log2](../Math/log2.md), [Math.log10](../Math/log10.md): the
  natural, base-2, and base-10 logarithm

## Trigonometry

Angles are in radians.

```ant
let degrees = 90.0
Math.sin(degrees * Math.PI / 180.0) // 1.0
Math.atan2(1.0, 0.0)                // 1.5707963267948966
```

- [Math.sin](../Math/sin.md), [Math.cos](../Math/cos.md), [Math.tan](../Math/tan.md): the sine,
  cosine, and tangent of an angle
- [Math.asin](../Math/asin.md), [Math.acos](../Math/acos.md), [Math.atan](../Math/atan.md): the
  inverse functions
- [Math.atan2](../Math/atan2.md): the angle of the point `(x, y)`, from `-PI` to `PI`; it takes `y`
  first

## Special float values

The `float64` functions never panic. Outside its domain a function returns `NaN` or an infinity.
`NaN` is not equal to anything, itself included, so test for it with `Math.isNaN`.

```ant
let x = Math.sqrt(-1.0)
x == Math.NaN                // false
Math.isNaN(x)                // true
Math.isFinite(Math.log(0.0)) // false
```

- [Math.isNaN](../Math/isNaN.md): whether `x` is `NaN`
- [Math.isFinite](../Math/isFinite.md): whether `x` is neither `NaN` nor an infinity
- [Math](../Math/index.md#constants): the constants `Math.PI`, `Math.E`, `Math.NaN`,
  `Math.Infinity`, `Math.maxFloat64`, `Math.minInt`, `Math.maxInt`, and `Math.maxUint64`

## Bitwise operations

Ant has no bitwise operators. The `Bits` functions take any one integer type and work on its two's
complement bits at its width.

```ant
let canRead = 0b100
let canWrite = 0b010
let mode = Bits.or(canRead, canWrite)
Bits.and(mode, canWrite) != 0 // true
Bits.shiftLeft(1, 4)          // 16
Bits.popCount(byte(255))      // 8
Bits.leadingZeros(byte(1))    // 7
```

- [Bits.and](../Bits/and.md), [Bits.or](../Bits/or.md), [Bits.xor](../Bits/xor.md),
  [Bits.not](../Bits/not.md): the logical operations on each bit
- [Bits.shiftLeft](../Bits/shiftLeft.md), [Bits.shiftRight](../Bits/shiftRight.md): the bits moved
  by `n`, arithmetic right for a signed type and logical for an unsigned one
- [Bits.rotateLeft](../Bits/rotateLeft.md), [Bits.rotateRight](../Bits/rotateRight.md): the bits
  turned by `n`, the bits shifted out coming back in at the other end
- [Bits.popCount](../Bits/popCount.md): the number of one bits
- [Bits.leadingZeros](../Bits/leadingZeros.md), [Bits.trailingZeros](../Bits/trailingZeros.md): the
  zero bits above the highest one bit, or below the lowest

## Random numbers

A `Random` generator made from a seed gives the same values in the same order on every run, which
makes a simulation or a test repeatable. `Math.random` and `Math.randomInt` cannot be seeded. None
of them is fit for keys or tokens: use `Crypto.randomBytes` for those.

```ant
let rng = Random.new(42)
let rolls = Array.range(0, 5).map({ rng.int(1, 7) })
println(rolls)
let deck = ["A", "K", "Q", "J", "10"]
rng.shuffle(deck)
println(deck)
```

```text
[2, 2, 5, 1, 3]
["J", "A", "Q", "10", "K"]
```

- [Random.new](../Random/new.md): a generator from a seed, or from the host's secure source without
  one
- [Random.int](../Random/int.md): a uniform `int` from `from`, inclusive, to `to`, exclusive
- [Random.float](../Random/float.md): a uniform `float64` in `[0, 1)`
- [Random.bool](../Random/bool.md): `true` or `false`, each half the time
- [Random.pick](../Random/pick.md): a uniform element of an array, or `null` when it is empty
- [Random.shuffle](../Random/shuffle.md): an array reordered in place
- [Random.fork](../Random/fork.md): a second generator, so each goroutine draws from its own
- [Math.random](../Math/random.md), [Math.randomInt](../Math/randomInt.md): an unseeded `float64` in
  `[0, 1)` or `int` from `from`, inclusive, to `to`, exclusive
- [Crypto.randomBytes](../Crypto/randomBytes.md): bytes from the host's secure source, for keys and
  tokens

## See also

- [int and the sized integers](../../language/types/int.md),
  [float32 and float64](../../language/types/float.md): the number types
- [Number literals](../../language/literals/numbers.md): how to write numbers in code
- [Number](../Math/types/Number.md), [Integer](../Math/types/Integer.md),
  [Ordered](../Math/types/Ordered.md): the bounds for a function generic over numbers
