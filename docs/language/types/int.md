<!-- status: reviewed -->

# int and the sized integers

`int` is the default integer type: a signed 64-bit two's complement integer. Eight sized integer
types cover the other widths, signed and unsigned.

| Type                               | Values                   |
| ---------------------------------- | ------------------------ |
| `int8` `int16` `int32` `int64`     | −2^(n−1) to 2^(n−1) − 1  |
| `uint8` `uint16` `uint32` `uint64` | 0 to 2^n − 1             |
| `int`                              | another name for `int64` |
| `byte`                             | another name for `uint8` |

## Details

- `int` and `int64` are one type under two names, as are `byte` and `uint8`. Ant always shows the
  short name.
- An integer literal is always an `int`. Every other kind is reached by calling its name as a
  conversion, such as `int8(5)`, which panics unless the value fits exactly.
- Integer kinds never mix, not even to widen: `int32` and `int` need a conversion before `+`, `<`,
  or a call.
- Arithmetic that leaves the kind's range panics with `integer overflow`, including unsigned
  subtraction below zero. Integer `/` and `%` by zero panic. `Bits.wrapAdd`, `Bits.wrapSub`, and
  `Bits.wrapMul` wrap instead.
- Integer `/` truncates toward zero, and `%` takes the sign of the dividend.

## Basic examples

An integer literal is an `int`:

```ant
let count = 42
count * 2   // 84
Math.maxInt // 9223372036854775807
```

Convert to a sized kind by calling its name:

```ant
int8(-128)        // -128
uint16(65535)     // 65535
int8(5) + int8(3) // 8
```

## Scope

### Literal forms

Every integer literal form gives an `int`, and `_` separates digits:

```ant
1_000_000 // 1000000
0xFF      // 255
0b1010    // 10
0o755     // 493
```

### Integer division

`/` drops the fraction toward zero, and `%` keeps the sign of the dividend:

```ant
7 / 2  // 3
-7 / 2 // -3
7 % 3  // 1
-7 % 3 // -1
```

### Every kind keeps its own range

Arithmetic on a sized kind stays in that kind and checks its range:

```ant
int8(100) + int8(27)            // 127
uint8(200) - uint8(100)         // 100
uint64(Math.maxInt) + uint64(1) // 9223372036854775808
Math.maxUint64                  // 18446744073709551615
```

### Two names, one type

`int64` and `int` are interchangeable everywhere, `is` accepts either name, and a type is shown by
its short name:

```ant
func half(n: int): int { n / 2 }

let total: int64 = 10
println(half(total))
let value: int | string = 3
println(value is int64)
println(type(Array[uint8]))
```

```text
5
true
Array[byte]
```

### Conversions are functions

Each integer name is a function value, so it maps over an array:

```ant
[72, 105].map(byte)          // [72, 105]
[int8(-1), int8(2)].map(int) // [-1, 2]
int(Math.trunc(-2.7))        // -2
float64(7) / 2.0             // 3.5
```

### Generic code over every kind

A type parameter bounded by `Integer` accepts all eight kinds, and the result keeps the caller's
kind:

```ant
func double[T: Integer](x: T): T { x + x }

println(double(int8(5)))
println(double(uint64(7)))
println(double(Math.maxUint64 / uint64(2)))
```

```text
10
14
18446744073709551614
```

### Named integer types

A `type` over `int` is a distinct type with the same arithmetic, and the result keeps the tag:

```ant
type UserId = int
let id = UserId(7)
id + UserId(1) // UserId(8)
int(id)        // 7
```

### From and to text

`String.toInt` parses decimal digits with an optional sign, and returns `null` on anything else.
`String.from` renders an integer as text:

```ant
"42".toInt()     // 42
"-17".toInt()    // -17
"4x".toInt()     // null
String.from(255) // "255"
```

## Applications

### A 32-bit hash with wrapping arithmetic

FNV-1a multiplies modulo 2^32. `Bits.wrapMul` wraps on `uint32` where `*` would panic:

```ant
func fnv1a(text: string): uint32 {
  var hash = uint32(2166136261)
  for b in text.toBytes() {
    hash = Bits.wrapMul(Bits.xor(hash, uint32(b)), uint32(16777619))
  }
  hash
}

println(fnv1a(""))
println(fnv1a("ant"))
```

```text
2166136261
522836950
```

### An average as a float

Integer division drops the fraction, so convert before dividing:

```ant
let scores = [7, 8, 10]
let total = scores.sum() ?? 0
total / scores.length()                   // 8
float64(total) / float64(scores.length()) // 8.333333333333334
```

### Clamp before narrowing

A conversion panics outside the target's range, so clamp the value first:

```ant
func toByte(n: int): byte { byte(Math.clamp(n, 0, 255)) }

[-20, 128, 300].map(toByte) // [0, 128, 255]
```

## Properties and relations

For a nonzero divisor, `/` and `%` rebuild the dividend, and the remainder is smaller than the
divisor:

```ant
for [a, b] in [[7, 2], [-7, 2], [7, -2], [-7, -2]] {
  assert((a / b) * b + a % b == a)
  assert(Math.abs(a % b) < Math.abs(b))
}
```

Wrapping arithmetic agrees with checked arithmetic whenever the result fits:

```ant
let a = int16(1200)
let b = int16(25)
assert(Bits.wrapAdd(a, b) == a + b)
assert(Bits.wrapMul(a, b) == a * b)
assert(Bits.wrapAdd(Math.maxInt, 1) == Math.minInt)
```

## Possible issues

### A literal does not adapt to a sized annotation

An integer literal is always an `int`, so a sized annotation needs a conversion:

```ant
let small: int8 = 5 // error: cannot use int where int8 is expected
```

```ant
let small: int8 = int8(5)
small // 5
```

### Kinds do not mix

No integer kind converts to another on its own, not even a narrower one to `int`:

```ant
let width = int32(640)
let area = width * 480 // error: * needs both operands of one type, got int32 and int
```

```ant
let width = int32(640)
int(width) * 480 // 307200
```

### Overflow panics

Arithmetic outside the kind's range panics instead of wrapping, and so does an unsigned result below
zero:

```ant
Math.maxInt + 1      // panic: integer overflow
int8(100) + int8(28) // panic: integer overflow
uint8(1) - uint8(2)  // panic: integer overflow
-Math.minInt         // panic: integer overflow
```

### Division by zero panics

```ant
10 / 0 // panic: division by zero
10 % 0 // panic: division by zero
```

### A conversion must be exact

A value out of the target's range, a fraction, `NaN`, or an infinity panics:

```ant
int8(300)     // panic: 300 out of range for int8
byte(-1)      // panic: -1 out of range for byte
int(2.5)      // panic: 2.5 has a fraction, so it does not convert to int
int(Math.NaN) // panic: NaN does not convert to int
```

To convert a `float64` to `int` without a panic, use `Math.toInt`, which returns `null` instead. To
narrow an integer, clamp it first with `Math.clamp`.

### The smallest int has no literal

`9223372036854775808` does not fit an `int`, so `-9223372036854775808` is a syntax error. Write
`Math.minInt`:

```ant
let lowest = -9223372036854775808 // error: integer literal 9223372036854775808 does not fit an int
```

```ant
Math.minInt // -9223372036854775808
```

### `Integer` is not a kind

A value typed `Integer` can hold any of the eight kinds, so arithmetic on it is a compile error. Use
a type parameter bounded by `Integer`, as in Generic code over every kind above:

```ant
func twice(n: Integer): Integer { n + n }
// error: + applies to a single numeric kind, string, or a named primitive type over one, not to Integer; narrow first
```

### Indexes are `int`

An array index must be an `int`, whatever kind the position was computed in:

```ant
let letters = ["a", "b", "c"]
let i = byte(1)
let second = letters[i] // error: cannot use byte where int is expected
```

```ant
let letters = ["a", "b", "c"]
let i = byte(1)
letters[int(i)] // "b"
```

## See also

- [byte](byte.md)
- [float64 and float32](float.md)
- [Number literals](../literals/numbers.md)
- [Arithmetic operators](../operators/arithmetic.md)
- [int](../../lib/prelude/int.md)
- [Integer](../../lib/Math/types/Integer.md)
- [Math.toInt](../../lib/Math/toInt.md)
- [Math.clamp](../../lib/Math/clamp.md)
- [Bits.wrapAdd](../../lib/Bits/wrapAdd.md)
- [Bits.wrapMul](../../lib/Bits/wrapMul.md)
- [String.toInt](../../lib/String/toInt.md)
- [Generics](../../guide/generics.md)
