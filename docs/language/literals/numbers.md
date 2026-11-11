<!-- status: reviewed -->

# Number literals

An integer literal such as `42`, `0xFF`, or `1_000` is an `int`. A float literal such as `3.14` or
`1e9` is a `float64`.

## Details

- An integer literal is always an `int` and a float literal is always a `float64`. There are no
  suffixes: write `1.0` for a float, and call a conversion such as `byte(255)` or `float32(1.5)` for
  any other numeric type.
- A float literal has digits on both sides of the point, an exponent, or both. `.5` and `1.` are
  errors.
- `_` may stand between two digits only, to group them.
- The prefixes `0x`, `0o`, `0b` and the exponent's `e` are lowercase. Hexadecimal digits may be
  either case.
- A decimal integer, or the digits before a float's point, start with `0` only when they are `0`:
  `0`, `0.5`, and `0e3` are fine; `007` and `00.5` are errors. Octal is `0o7`.
- A literal that does not fit its type is a compile error. The most negative `int` has no literal;
  write `Math.minInt`.

## Basic examples

Integers in decimal, hexadecimal, octal, and binary:

```ant
42        // 42
1_000_000 // 1000000
0xFF      // 255
0o755     // 493
0b1010    // 10
```

Floats with a point, an exponent, or both:

```ant
3.14   // 3.14
1e9    // 1000000000.0
2.5e-3 // 0.0025
```

## Scope

### Digit separators

`_` groups digits in any base, and in either part of a float:

```ant
0xFFFF_FFFF // 4294967295
0b1111_0000 // 240
1_000.000_1 // 1000.0001
```

### Hexadecimal digits

The prefix is lowercase; the digits are not:

```ant
0xff == 0xFF // true
0xCafe       // 51966
```

### Zero

A decimal integer, or the digits before a float's point, start with `0` only when they are `0`.
Digits after a prefix may have leading zeros:

```ant
0    // 0
0.5  // 0.5
0e3  // 0.0
0x0F // 15
```

### Negative numbers

A minus sign is the negation operator applied to a literal:

```ant
-5      // -5
-2.5e-3 // -0.0025
-0.0    // -0.0
```

### Integer and float are different types

`1` and `1.0` are values of different types. `is` tells them apart in a union:

```ant
let n: int | float64 = 1.0
n is float64 // true
n is int     // false
```

### Other numeric types

A conversion reaches every other numeric type: the literal is an `int` or a `float64`, and the
conversion turns it into the target type:

```ant
byte(255)    // 255
int8(-128)   // -128
uint64(0xFF) // 255
float32(1.5) // 1.5
```

### Limits

The largest `int` is a literal; the smallest is `Math.minInt`:

```ant
9_223_372_036_854_775_807 == Math.maxInt      // true
-9_223_372_036_854_775_807 - 1 == Math.minInt // true
```

A float literal can be as large as the largest `float64` and as small as the smallest one above
zero:

```ant
1.7976931348623157e308 // 1.7976931348623157e+308
5e-324                 // 5e-324
```

## Applications

File modes read naturally in octal:

```ant
let path = "notes.txt"
Result.unwrap(File.write(path, "hello\n"))
defer { File.remove(path) }
Result.unwrap(File.chmod(path, 0o600))
let info = Result.unwrap(File.stat(path))
println(String.fromInt(info.mode, base: 8))
```

```text
600
```

Flags written in binary show which bit each one sets:

```ant
let read = 0b100
let write = 0b010
let exec = 0b001
let mode = Bits.or(read, write)
Bits.and(mode, write) != 0 // true
Bits.and(mode, exec) != 0  // false
```

Split a hexadecimal color into its channels:

```ant
let orange = 0xFF8800
let red = Bits.shiftRight(orange, 16)
let green = Bits.and(Bits.shiftRight(orange, 8), 0xFF)
let blue = Bits.and(orange, 0xFF)
[red, green, blue] // [255, 136, 0]
```

Spell bytes in hexadecimal:

```ant
Encoding.toHex(Bytes.fromArray([0xCA, 0xFE])) // "cafe"
```

## Properties and relations

A number has the same value in every base:

```ant
assert(255 == 0xFF)
assert(0xFF == 0o377)
assert(0o377 == 0b1111_1111)
```

`_` does not change a value, and an exponent scales by a power of ten:

```ant
assert(1_000_000 == 1000000)
assert(1e6 == 1_000_000.0)
assert(2.5e-3 == 0.0025)
```

## Possible issues

### An integer literal is not a float

`1` is an `int`, so it does not fit where a `float64` is expected:

```ant
let ratio: float64 = 1 // error: cannot use int where float64 is expected
```

Write a float literal instead:

```ant
let ratio: float64 = 1.0
ratio // 1.0
```

### Literals do not adapt to a sized type

An integer literal stays an `int` even where a `byte` or `int8` is expected:

```ant
let b: byte = 255 // error: cannot use int where byte is expected
```

Convert it:

```ant
let b = byte(255)
b // 255
```

### Integer and float do not mix

```ant
let total = 1 + 1.0 // error: + needs both operands of one type, got int and float64
```

Comparison does not mix them either:

```ant
let same = 1 == 1.0 // error: int and float64 cannot be compared with ==: neither is assignable to the other
```

Convert the `int`, or write both as floats:

```ant
float64(1) + 1.0 // 2.0
1.0 + 1.0        // 2.0
```

### Integer division truncates

Two integer literals divide as integers:

```ant
7 / 2     // 3
7.0 / 2.0 // 3.5
```

### Floats are binary fractions

Most decimal fractions have no exact `float64`:

```ant
0.1 + 0.2        // 0.30000000000000004
0.1 + 0.2 == 0.3 // false
```

### A leading zero is not octal

Other languages read `007` as octal. Ant rejects it; write `0o7`:

```ant
let code = 007 // error: a decimal literal other than 0 cannot start with 0
```

### A `uint64` above `Math.maxInt` has no literal

A conversion takes an `int` literal, so a larger `uint64` does not fit it:

```ant
let max = uint64(18446744073709551615) // error: integer literal 18446744073709551615 does not fit an int
```

Write `Math.maxUint64` for the largest one.

### A float needs digits on both sides of the point

```ant
let half = .5 // error: a float literal needs a digit before the point
```

```ant
let one = 1. // error: a float literal needs a digit after the point
```

### Prefixes and exponents are lowercase

```ant
let mask = 0XFF // error: malformed number literal: write the prefix 0x in lowercase
```

```ant
let big = 1E9 // error: malformed number literal: write the exponent with a lowercase e
```

### `_` goes between digits

A doubled, leading, or trailing `_` is an error:

```ant
let n = 1__000 // error: malformed number literal
```

### Out of range

An integer literal past `Math.maxInt` does not compile, even after a minus sign:

```ant
let low = -9223372036854775808 // error: integer literal 9223372036854775808 does not fit an int
```

A float literal that would round to infinity or to zero does not compile either:

```ant
let huge = 1e400 // error: float literal 1e400 is out of range
```

```ant
let tiny = 1e-400 // error: float literal 1e-400 is out of range
```

## See also

[int](../types/int.md) · [float64](../types/float.md) · [byte](../types/byte.md) ·
[Arithmetic operators](../operators/arithmetic.md) · [Math](../../lib/Math/index.md) ·
[Bits](../../lib/Bits/index.md) · [String.toInt](../../lib/String/toInt.md) ·
[String.fromInt](../../lib/String/fromInt.md) · [String literals](strings.md)
