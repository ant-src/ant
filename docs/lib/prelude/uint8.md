<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# uint8

```ant signature
func uint8[T: Number](x: T): uint8
```

uint8 converts x to uint8 exactly. It panics if x is out of range, has a fraction, or is NaN or
infinite; uint8(Math.trunc(x)) truncates. x may also be a named primitive over a number.

<!-- end generated -->

## Details

- `uint8` holds whole numbers from 0 to 255. It is the same type as `byte`, and types and compile
  errors show it as `byte`.
- Arithmetic on two `uint8` values gives a `uint8` and panics when the result leaves 0 to 255,
  including subtraction below zero. For wrapping arithmetic use [Bits.wrapAdd](../Bits/wrapAdd.md)
  and its siblings.

## Basic examples

```ant
uint8(42)  // 42
uint8(0)   // 0
uint8(255) // 255
```

A float converts when it is a whole number:

```ant
uint8(128.0) // 128
uint8(-0.0)  // 0
```

A value outside 0 to 255 panics, and the message names the type `byte`:

```ant
uint8(256) // panic: 256 out of range for byte
uint8(-1)  // panic: -1 out of range for byte
```

## Scope

Any numeric type converts, as long as the value fits:

```ant
uint8(int8(100))     // 100
uint8(uint64(7))     // 7
uint8(int16(255))    // 255
uint8(float32(12.0)) // 12
uint8(int8(-1))      // panic: -1 out of range for byte
```

A float with a fraction, `NaN`, and the infinities panic:

```ant
uint8(2.5)            // panic: 2.5 has a fraction, so it does not convert to byte
uint8(Math.NaN)       // panic: NaN does not convert to byte
uint8(-Math.Infinity) // panic: -Infinity does not convert to byte
```

A named primitive over a number is unwrapped first:

```ant
type Percent = int
uint8(Percent(80)) // 80
```

`uint8` is a function value, so it can be passed where a function is expected. The result is an
`Array[byte]`:

```ant
Array.map([0, 128, 255], uint8) // [0, 128, 255]
```

In a pipe, the value is the argument:

```ant
64.0 | uint8() // 64
```

Indexing `Bytes` gives a `byte`, which is a `uint8`, so it mixes with other `uint8` values:

```ant
let b = String.toBytes("hi")
b[0] + uint8(1) // 105
```

## Applications

Turn a color channel from 0.0 to 1.0 into a byte. Round first, since the conversion does not:

```ant
func channel(f: float64): uint8 {
  uint8(Math.round(Math.clamp(f, 0.0, 1.0) * 255.0))
}

println(channel(0.5))
println(channel(1.0))
println(channel(-0.2))
```

```text
128
255
0
```

Add up the bytes of a `Bytes` value. Each one is a `uint8`, so convert it to `int` before adding, or
the sum overflows past 255:

```ant
let data = String.toBytes("Ant")
var sum = 0
for x in data {
  sum += int(x)
}
println(sum)
```

```text
291
```

Split a byte into its high and low four bits with `Bits`, which keeps the `uint8` type:

```ant
let b = uint8(0xa7)
let high = Bits.shiftRight(b, 4)
let low = Bits.and(b, uint8(0x0f))
println(high)
println(low)
```

```text
10
7
```

Build `Bytes` from `int` values:

```ant
let header = Array.map([0x89, 0x50, 0x4e, 0x47], uint8)
println(Bytes.fromArray(header))
```

```text
Bytes(89 50 4e 47)
```

## Properties and relations

Every `int` from 0 to 255 survives the round trip through `uint8`:

```ant
for n in Iter.range(0, 256) {
  assert(int(uint8(n)) == n)
}
```

`uint8` and `byte` are one type, so their values compare equal and `is` accepts either name:

```ant
let n: Number = uint8(9)
assert(n == byte(9))
assert(n is byte)
assert(n is uint8)
```

An `Array[uint8]` is an `Array[byte]`:

```ant
func total(xs: Array[uint8]): int {
  Array.reduce(xs, 0, func(acc: int, x: byte) { acc + int(x) })
}

let data: Array[byte] = String.toBytes("ab").toArray()
assert(total(data) == 195)
```

`Bits.wrapAdd` agrees with `+` whenever `+` does not overflow:

```ant
let a = uint8(200)
let b = uint8(55)
assert(Bits.wrapAdd(a, b) == a + b)
```

## Possible issues

An integer literal is an `int`, never a `uint8`. The error names the type `byte`:

```ant
let level: uint8 = 5 // error: cannot use int where byte is expected
```

```ant
let level: uint8 = uint8(5)
level // 5
```

`uint8` and `int` do not mix in arithmetic. Convert one side so that both have the same type:

```ant
let level = uint8(3)
let next = level + 1 // error: + needs both operands of one type, got byte and int
```

```ant
let level = uint8(3)
level + uint8(1) // 4
int(level) + 1   // 4
```

Arithmetic overflows quickly in eight bits, and going below zero overflows too:

```ant
uint8(200) + uint8(100)              // panic: integer overflow
uint8(0) - uint8(1)                  // panic: integer overflow
Bits.wrapAdd(uint8(200), uint8(100)) // 44
Bits.wrapSub(uint8(0), uint8(1))     // 255
```

The conversion never rounds. A float with a fraction panics, so round or truncate it first:

```ant
uint8(0.9 * 255.0)             // panic: 229.5 has a fraction, so it does not convert to byte
uint8(Math.round(0.9 * 255.0)) // 230
```

A string does not convert. Parse it with `String.toInt`, which returns `int | null`, and convert the
result:

```ant
let x = uint8("200") // error: byte converts a number or a named primitive type over a number, got string
```

```ant
uint8(String.toInt("200") ?? 0) // 200
```

## See also

[byte](byte.md) · [int8](int8.md) · [uint16](uint16.md) · [int](int.md) ·
[Bytes.fromArray](../Bytes/fromArray.md) · [Math.round](../Math/round.md) ·
[Math.clamp](../Math/clamp.md) · [Bits.wrapAdd](../Bits/wrapAdd.md) ·
[byte type](../../language/types/byte.md) · [Arithmetic](../../language/operators/arithmetic.md)
