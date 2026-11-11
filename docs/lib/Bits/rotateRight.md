<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.rotateRight

```ant signature
func Bits.rotateRight[T: Integer](a: T, n: int): T
```

Bits.rotateRight returns a rotated right by n bits. The count is taken modulo the bit width of T, so
n may exceed it. It panics if n \< 0.

<!-- end generated -->

## Details

- The bits that fall off the low end come back in at the high end, so no bit is lost.
- The width is the width of `T`: 8 for `uint8`, 64 for `int`.
- For a signed type the rotation works on the two's complement bits, so a bit can move into or out
  of the sign bit.

## Basic examples

```ant
Bits.rotateRight(uint8(1), 1)    // 128
Bits.rotateRight(uint8(3), 1)    // 129
Bits.rotateRight(uint8(0xab), 4) // 186
```

Written in base 2, the low bit moves to the top:

```ant
String.fromInt(uint8(0b00000110), 2)                      // "110"
String.fromInt(Bits.rotateRight(uint8(0b00000110), 2), 2) // "10000001"
```

## Scope

Every integer type, each at its own width:

```ant
Bits.rotateRight(uint8(1), 1)  // 128
Bits.rotateRight(uint16(1), 1) // 32768
Bits.rotateRight(uint32(1), 1) // 2147483648
Bits.rotateRight(uint64(1), 1) // 9223372036854775808
```

A signed type rotates its two's complement bits. A low bit that comes around lands in the sign bit:

```ant
Bits.rotateRight(int8(1), 1)    // -128
Bits.rotateRight(int8(-128), 7) // 1
Bits.rotateRight(int8(-1), 3)   // -1
```

A count at or beyond the width wraps around:

```ant
Bits.rotateRight(uint8(2), 9)   // 1
Bits.rotateRight(uint8(2), 8)   // 2
Bits.rotateRight(uint64(5), 64) // 5
```

A count of 0 returns the value unchanged:

```ant
Bits.rotateRight(uint8(5), 0) // 5
```

In a pipe, the value is the first argument:

```ant
uint16(0x1234) | Bits.rotateRight(4) // 16675
```

The result has the type of `a`, so a `uint32` stays a `uint32`:

```ant
let word = uint32(0x12345678)
let rotated: uint32 = Bits.rotateRight(word, 8)
String.fromInt(rotated, 16) // "78123456"
```

## Applications

Swap the two halves of a byte, its high and low nibbles:

```ant
let b = uint8(0x3c)
String.fromInt(Bits.rotateRight(b, 4), 16) // "c3"
```

Mix the bits of a word the way SHA-256 does in its round function, with three rotations combined by
`xor`:

```ant
func bigSigma0(x: uint32): uint32 {
  let a = Bits.rotateRight(x, 2)
  let b = Bits.rotateRight(x, 13)
  let c = Bits.rotateRight(x, 22)
  Bits.xor(Bits.xor(a, b), c)
}

println(String.fromInt(bigSigma0(uint32(0x6a09e667)), 16))
```

```text
ce20b47e
```

Step through a ring of eight flags, one position to the right on each turn:

```ant
var lamp = uint8(0b10000000)
for _ in Iter.range(0, 4) {
  println(String.fromInt(lamp, 2))
  lamp = Bits.rotateRight(lamp, 1)
}
```

```text
10000000
1000000
100000
10000
```

## Properties and relations

<!-- test: before
let x = uint8(0b10110010)
let n = 3
-->

```ant
assert(Bits.rotateRight(Bits.rotateLeft(x, n), n) == x)
assert(Bits.rotateRight(x, n) == Bits.rotateLeft(x, 8 - n))
assert(Bits.rotateRight(x, n + 8) == Bits.rotateRight(x, n))
assert(Bits.popCount(Bits.rotateRight(x, n)) == Bits.popCount(x))
```

For an unsigned value and a count between 1 and the width minus 1, a rotation is two shifts joined
by `or`:

<!-- test: before
let x = uint8(0b10110010)
let n = 3
-->

```ant
assert(Bits.rotateRight(x, n) == Bits.or(Bits.shiftRight(x, n), Bits.shiftLeft(x, 8 - n)))
```

## Possible issues

A negative count panics. Use [Bits.rotateLeft](rotateLeft.md) to rotate the other way:

```ant
Bits.rotateRight(uint8(5), -1) // panic: Bits.rotateRight needs a count of at least 0, got -1
Bits.rotateLeft(uint8(5), 1)   // 10
```

A literal is an `int`, 64 bits wide, so its low bit comes around at bit 63, not bit 7:

```ant
Bits.rotateRight(1, 1)        // -9223372036854775808
Bits.rotateRight(uint8(1), 1) // 128
```

`Bits.rotateRight` is not a function on the integer types, so a dot call does not compile. Use the
namespace or a pipe:

```ant
let x = uint8(5)
let y = x.rotateRight(1) // error: byte has no field or function rotateRight
```

## See also

[Bits.rotateLeft](rotateLeft.md), [Bits.shiftRight](shiftRight.md), [Bits.shiftLeft](shiftLeft.md),
[Bits.popCount](popCount.md), [Bits.xor](xor.md), [String.fromInt](../String/fromInt.md)
