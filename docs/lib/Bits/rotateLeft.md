<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.rotateLeft

```ant signature
func Bits.rotateLeft[T: Integer](a: T, n: int): T
```

Bits.rotateLeft returns a rotated left by n bits. The count is taken modulo the bit width of T, so n
may exceed it. It panics if n \< 0.

<!-- end generated -->

## Details

- Bits that leave the top of `a` come back in at the bottom, so no bit is lost: the result has as
  many one bits as `a`.
- The width is that of `T`: 8 for `uint8`, 64 for `int`. A count of the width or more wraps around,
  and a count of 0 returns `a` unchanged.
- A signed value rotates its two's complement bits, so the sign can change.
- Rotate the other way with [`Bits.rotateRight`](rotateRight.md).

## Basic examples

The top bit of `129` (`0b1000_0001`) wraps around to the bottom:

```ant
Bits.rotateLeft(uint8(129), 1) // 3
```

A count larger than the width wraps around: 9 on 8 bits is 1:

```ant
Bits.rotateLeft(uint8(1), 9) // 2
```

## Scope

Rotate by four bits to swap the two halves of a byte:

```ant
Bits.rotateLeft(uint8(0x4F), 4) | String.fromInt(16) // "f4"
```

Rotate by whole bytes on a wider type:

```ant
Bits.rotateLeft(uint32(0x12345678), 8) | String.fromInt(16) // "34567812"
```

A count of 0 or of the width gives the value back:

```ant
Bits.rotateLeft(uint16(0xABCD), 0)  // 43981
Bits.rotateLeft(uint16(0xABCD), 16) // 43981
```

A signed value rotates its two's complement bits:

```ant
Bits.rotateLeft(int8(-128), 1) // 1
Bits.rotateLeft(int8(64), 1)   // -128
Bits.rotateLeft(int8(-1), 3)   // -1
```

An integer literal is an `int`, so it rotates over 64 bits:

```ant
Bits.rotateLeft(1, 63) // -9223372036854775808
Bits.rotateLeft(1, 64) // 1
```

The count can be computed at run time:

```ant
let shifts = [0, 1, 2, 3]
shifts.map({ Bits.rotateLeft(uint8(0b1000_0001), x) }) // [129, 3, 6, 12]
```

## Applications

Run a light around a ring of eight LEDs:

```ant
var lights = uint8(0b0000_0011)
for _ in Iter.range(0, 5) {
  println(String.fromInt(lights, 2).padStart(8, "0"))
  lights = Bits.rotateLeft(lights, 3)
}
```

```text
00000011
00011000
11000000
00000110
00110000
```

Mix the bytes of a string into a small hash, rotating the state before each byte so that order
matters:

```ant
func hash(s: string): uint32 {
  var h = uint32(0)
  for b in s.toBytes() {
    h = Bits.xor(Bits.rotateLeft(h, 5), uint32(b))
  }
  h
}

println(String.fromInt(hash("ant"), 16))
println(String.fromInt(hash("tan"), 16))
```

```text
189b4
1dc4e
```

## Properties and relations

`Bits.rotateRight` by the same count undoes it:

```ant
let a = uint16(0xBEEF)
assert(Bits.rotateRight(Bits.rotateLeft(a, 5), 5) == a)
```

The count is taken modulo the width:

```ant
let a = uint8(0b1011_0001)
assert(Bits.rotateLeft(a, 11) == Bits.rotateLeft(a, 3))
```

A rotation keeps the number of one bits:

```ant
let a = int32(-123456)
assert(Bits.popCount(Bits.rotateLeft(a, 7)) == Bits.popCount(a))
```

For an unsigned value and a count between 1 and the width minus 1, a rotation is two shifts joined
with `Bits.or`:

```ant
let a = uint8(0b1100_1010)
let n = 3
assert(Bits.rotateLeft(a, n) == Bits.or(Bits.shiftLeft(a, n), Bits.shiftRight(a, 8 - n)))
```

Rotating left by `n` is rotating right by the width minus `n`:

```ant
let a = uint32(0xCAFE_F00D)
assert(Bits.rotateLeft(a, 12) == Bits.rotateRight(a, 32 - 12))
```

## Possible issues

A negative count panics instead of rotating right:

```ant
Bits.rotateLeft(uint8(1), -1) // panic: Bits.rotateLeft needs a count of at least 0, got -1
```

`rotateLeft` is in the `Bits` namespace, not in the namespace of the integer, so it cannot be called
with a dot:

```ant
let flags = uint8(129)
let rotated = flags.rotateLeft(1) // error: byte has no field or function rotateLeft
```

The count is an `int`, whatever the type of `a`:

```ant
let rotated = Bits.rotateLeft(uint8(1), uint8(1)) // error: cannot use byte where int is expected
```

A literal without a conversion rotates as a 64-bit `int`, not as a byte:

```ant
Bits.rotateLeft(128, 1)        // 256
Bits.rotateLeft(uint8(128), 1) // 1
```

## See also

[Bits.rotateRight](rotateRight.md) · [Bits.shiftLeft](shiftLeft.md) ·
[Bits.shiftRight](shiftRight.md) · [Bits.popCount](popCount.md) · [Bits.xor](xor.md) ·
[Bits](index.md)
