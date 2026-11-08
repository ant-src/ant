<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.fromInt

```ant signature
func String.fromInt[T: Integer](x: T, base: int = 10): string
```

String.fromInt returns x written in base, with lower-case digits and a - in front when negative.
base defaults to 10, and it panics if base is outside 2 ..= 36.

<!-- end generated -->

## Details

- `String.fromInt` writes `x` in `base` with lower-case digits, `0` to `9` and then `a` to `z`, and
  a `-` in front when `x` is negative.
- `base` defaults to `10`. Any base from 2 to 36 works; a base outside `2 ..= 36` panics.
- `x` may be any of the eight integer types, the members of `Integer`. An unsigned value is written
  as the number it is, so the largest `uint64` is `18446744073709551615`
  ([spec: Primitives](https://github.com/ant-src/spec/blob/master/ant-lang.md#primitives),
  [String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)).
- The solution for `T` must be a single member of `Integer`. A float, a union such as `int | null`,
  or a named primitive type over `int` is a compile error until you convert or narrow it
  ([spec: Generics](https://github.com/ant-src/spec/blob/master/ant-lang.md#generics),
  [Named primitive types](https://github.com/ant-src/spec/blob/master/ant-lang.md#named-primitive-types)).

## Basic examples

Write a number in hexadecimal:

```ant
String.fromInt(255, 16) // "ff"
```

Without a base, the result is decimal:

```ant
String.fromInt(42) // "42"
```

A negative number gets a `-` in front of its digits:

```ant
String.fromInt(-255, 16) // "-ff"
```

## Scope

Every base from 2 to 36 works. Above ten, the digits go on with the letters `a` to `z`:

```ant
String.fromInt(10, 2)  // "1010"
String.fromInt(493, 8) // "755"
String.fromInt(35, 36) // "z"
String.fromInt(36, 36) // "10"
```

Zero is `"0"` in every base:

```ant
String.fromInt(0)    // "0"
String.fromInt(0, 2) // "0"
```

The extremes of `int` come out in full, and the most negative one keeps its sign:

```ant
String.fromInt(Math.maxInt)    // "9223372036854775807"
String.fromInt(Math.minInt)    // "-9223372036854775808"
String.fromInt(Math.minInt, 2) // "-1000000000000000000000000000000000000000000000000000000000000000"
```

Every integer kind is accepted. An unsigned value is the number it is, never a negative one:

```ant
String.fromInt(Math.maxUint64)     // "18446744073709551615"
String.fromInt(Math.maxUint64, 16) // "ffffffffffffffff"
String.fromInt(byte(200), 2)       // "11001000"
String.fromInt(int8(-128), 16)     // "-80"
String.fromInt(uint32(4000000000)) // "4000000000"
```

Call it in a pipe, with named arguments, or pass it as a function value:

```ant
42 | String.fromInt(base: 8)                   // "52"
String.fromInt(x: 5, base: 2)                  // "101"
Array.map([10, 11], String.fromInt)            // ["10", "11"]
Array.map([1, 2, 3], { String.fromInt(x, 2) }) // ["1", "10", "11"]
```

## Named arguments

### base

`base` picks the radix, from 2 to 36, and defaults to `10`. Name it to make a call read clearly:

```ant
String.fromInt(511, base: 8)  // "777"
String.fromInt(511, base: 16) // "1ff"
String.fromInt(511)           // "511"
```

## Applications

Build a CSS color from three channels, zero-padded to two hex digits each:

```ant
func hex(c: int): string {
  String.padStart(String.fromInt(c, 16), 2, "0")
}

func color(r: int, g: int, b: int): string {
  "#${hex(r)}${hex(g)}${hex(b)}"
}

println(color(255, 128, 0))
println(color(0, 10, 200))
```

```text
#ff8000
#000ac8
```

Show the bits of a byte, eight wide:

```ant
for b in [byte(5), byte(160), byte(255)] {
  println(String.padStart(String.fromInt(b, 2), 8, "0"))
}
```

```text
00000101
10100000
11111111
```

Make short identifiers from counters in base 36:

```ant
var ids = Array.map([35, 36, 1295, 1296, 1000000], { String.fromInt(x, 36) })
println(ids)
```

```text
["z", "10", "zz", "100", "lfls"]
```

Print a Unix permission mode in octal, the way `chmod` takes it:

```ant
type Mode = { owner: int, group: int, other: int }

func octal(m: Mode): string {
  String.fromInt(m.owner * 64 + m.group * 8 + m.other, 8)
}

println(octal(Mode{ owner: 7, group: 5, other: 5 }))
println(octal(Mode{ owner: 6, group: 4, other: 0 }))
```

```text
755
640
```

## Properties and relations

For an `int`, [`String.toInt`](toInt.md) reads back what `String.fromInt` writes, in the same base:

```ant
for n in [0, 1, -1, 255, -12345, Math.maxInt, Math.minInt] {
  for base in [2, 8, 10, 16, 36] {
    assert(String.toInt(String.fromInt(n, base), base) == n)
  }
}
```

[`String.toUint`](toUint.md) does the same for `uint64`, up to the largest one:

```ant
var big = Math.maxUint64
assert(String.toUint(String.fromInt(big, 16), 16) == big)
```

In base 10, the result is what [`String.from`](from.md) and string interpolation write:

```ant
for n in [0, 7, -42, Math.maxInt] {
  assert(String.fromInt(n) == String.from(n))
  assert(String.fromInt(n) == "${n}")
}
```

A negative number is its absolute value with a `-` in front:

```ant
for n in [1, 255, 4096] {
  assert(String.fromInt(-n, 16) == "-" + String.fromInt(n, 16))
}
```

## Possible issues

A base outside `2 ..= 36` panics:

```ant
String.fromInt(10, 1)  // panic: String.fromInt needs a base from 2 to 36, got 1
String.fromInt(10, 37) // panic: String.fromInt needs a base from 2 to 36, got 37
```

A negative number is written with a sign, not as its two's complement bits, so `-1` in hex is
`"-1"`, not `"ffffffffffffffff"`:

```ant
String.fromInt(-1, 16)      // "-1"
String.fromInt(int8(-1), 2) // "-1"
```

A float is not an `Integer`, so passing one is a compile error. Convert it first: with `int(...)`
when it is whole, otherwise with [`Math.round`](../Math/round.md) or
[`Math.trunc`](../Math/trunc.md) and then `int(...)`:

<!-- test: error which is not a single member of its bound Integer -->

```ant
println(String.fromInt(2.5))
```

```ant
String.fromInt(int(Math.round(2.5))) // "3"
```

A value that may be `null`, such as the result of [`String.toInt`](toInt.md), must be narrowed
first:

<!-- test: error which is not a single member of its bound Integer -->

```ant
var n = String.toInt("42")
println(String.fromInt(n))
```

```ant
var n = String.toInt("42")
if n != null { String.fromInt(n, 2) } // "101010"
```

A named primitive type over `int` is not a member of `Integer`, so passing one is a compile error:

<!-- test: error which is not a single member of its bound Integer -->

```ant
type UserId = int
var id = UserId(42)
println(String.fromInt(id))
```

Unwrap it with `int(...)`:

```ant
type UserId = int
var id = UserId(42)
String.fromInt(int(id)) // "42"
```

A number has no namespace, so a call through `.` fails
([spec: Calls through `.`](https://github.com/ant-src/spec/blob/master/ant-lang.md#calls-through-)).
Call `String.fromInt(n)` or pipe `n | String.fromInt()`:

<!-- test: error int has no field or function fromInt -->

```ant
var n = 42
println(n.fromInt())
```

There are no leading zeros and no width. Pad the result with [`String.padStart`](padStart.md):

```ant
String.fromInt(7, 2)                             // "111"
String.padStart(String.fromInt(7, 2), 8, "0")    // "00000111"
String.padStart(String.fromInt(255, 16), 8, "0") // "000000ff"
```

The letters are always lower case. Upper-case them with [`String.toUpper`](toUpper.md):

```ant
String.toUpper(String.fromInt(48879, 16)) // "BEEF"
```

## See also

[String.toInt](toInt.md) · [String.toUint](toUint.md) · [String.fromFloat](fromFloat.md) ·
[String.from](from.md) · [String.padStart](padStart.md) · [String.toUpper](toUpper.md) ·
[Math.round](../Math/round.md) · [Math.trunc](../Math/trunc.md) ·
[Generics](../../guide/generics.md)
