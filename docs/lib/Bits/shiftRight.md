<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.shiftRight

```ant signature
func Bits.shiftRight[T: Integer](a: T, n: int): T
```

Bits.shiftRight returns a shifted right by n bits. The shift is arithmetic for a signed type and
logical for an unsigned one. It panics if n \< 0 or n is at least the bit width of T.

<!-- end generated -->

## Details

- The bits shifted out of the low end are lost. A signed value fills the high end with copies of its
  sign bit, so a negative number stays negative; an unsigned value fills it with zeros.
- On a signed value, the shift rounds toward negative infinity: `Bits.shiftRight(-7, 1)` is `-4`,
  while `-7 / 2` is `-3`.

## Basic examples

```ant
Bits.shiftRight(int8(-8), 1)   // -4
Bits.shiftRight(uint8(248), 1) // 124
Bits.shiftRight(0x1234, 8)     // 18
```

The same bits give different results as signed and as unsigned values:

```ant
Bits.shiftRight(int8(-128), 4) // -8
Bits.shiftRight(uint8(128), 4) // 8
```

## Scope

A count of 0 returns the value unchanged, and a count of one below the width is the largest allowed:

```ant
Bits.shiftRight(5, 0)            // 5
Bits.shiftRight(Math.maxInt, 63) // 0
Bits.shiftRight(uint32(1), 31)   // 0
```

A negative value shifted far enough becomes `-1`, never `0`:

```ant
Bits.shiftRight(-1, 10)          // -1
Bits.shiftRight(Math.minInt, 63) // -1
```

A non-negative value shifted far enough becomes `0`:

```ant
Bits.shiftRight(0xFF, 7) // 1
Bits.shiftRight(0xFF, 8) // 0
Bits.shiftRight(0, 5)    // 0
```

Show the bits with `String.fromInt` and `base: 2`:

```ant
String.fromInt(0b1011_0000, base: 2)                     // "10110000"
String.fromInt(Bits.shiftRight(0b1011_0000, 3), base: 2) // "10110"
```

In a pipe, the value is the first argument:

```ant
0xABCD | Bits.shiftRight(4) // 2748
```

Inside a lambda, shift each element of an array:

```ant
[16, 32, 64].map({ Bits.shiftRight(x, 2) }) // [4, 8, 16]
```

## Named arguments

```ant
Bits.shiftRight(a: 256, n: 4) // 16
```

## Applications

Split a packed `0xRRGGBB` color into its channels:

```ant
func channels(rgb: int): Array[int] {
  [16, 8, 0].map({ Bits.shiftRight(rgb, x) | Bits.and(0xFF) })
}

channels(0x3366CC) // [51, 102, 204]
```

Split a 16-bit word into its high and low bytes:

```ant
let word = uint16(0xBEEF)
Bits.shiftRight(word, 8)     // 190
Bits.and(word, uint16(0xFF)) // 239
```

Read a field of a few bits out of a packed value, here bits 4 to 6 of a byte:

```ant
func field(b: byte, from: int, width: int): byte {
  let mask = Bits.shiftLeft(byte(1), width) - byte(1)
  Bits.and(Bits.shiftRight(b, from), mask)
}

field(byte(0b0101_0000), 4, 3) // 5
```

Count the binary digits of a positive number by shifting until nothing is left:

```ant
func bitLength(n: int): int {
  var rest = n
  var count = 0
  while rest > 0 {
    rest = Bits.shiftRight(rest, 1)
    count += 1
  }
  count
}

bitLength(1)    // 1
bitLength(255)  // 8
bitLength(1000) // 10
```

## Properties and relations

For a non-negative value, a shift by `n` is division by `2` to the power `n`:

```ant
for a in [0, 1, 7, 100, 12345] {
  for n in [0, 1, 3, 8] {
    assert(Bits.shiftRight(a, n) == a / Bits.shiftLeft(1, n))
  }
}
```

For a negative value, the shift rounds down, so it differs from `/` when the division has a
remainder:

```ant
Bits.shiftRight(-7, 1) // -4
-7 / 2                 // -3
Bits.shiftRight(-8, 1) // -4
-8 / 2                 // -4
```

For an unsigned value, a right shift undoes a left shift that dropped no set bits:

```ant
let a = uint8(0b0001_0110)
assert(Bits.shiftRight(Bits.shiftLeft(a, 3), 3) == a)
```

Shifting twice adds the counts, as long as the sum stays below the width:

```ant
let a = 0x7FFF_0000
assert(Bits.shiftRight(Bits.shiftRight(a, 4), 12) == Bits.shiftRight(a, 16))
```

## Possible issues

A count equal to the width or larger panics, and so does a negative count. Check the count before
the call, for example with a guard on `n < 8` for an `int8`:

```ant
Bits.shiftRight(int8(1), 8) // panic: Bits.shiftRight needs a count from 0 up to but excluding 8, got 8
Bits.shiftRight(1, 64)      // panic: Bits.shiftRight needs a count from 0 up to but excluding 64, got 64
Bits.shiftRight(1, -1)      // panic: Bits.shiftRight needs a count from 0 up to but excluding 64, got -1
```

The integer types have no `shiftRight` of their own, so a call through `.` is a compile error. Use
the namespace form or a pipe:

```ant
let word = 0xABCD
let high = word.shiftRight(8) // error: int has no field or function shiftRight
```

The count is always an `int`, whatever the type of the value:

```ant
let b = uint8(0xF0)
let low = Bits.shiftRight(b, uint8(4)) // error: cannot use byte where int is expected
```

Floats have no bits to shift:

```ant
let half = Bits.shiftRight(1.5, 1)
// error: T would be float64, which is not a member of its bound Integer
```

A named primitive type is not an `Integer`. Unwrap it, then tag the result again:

```ant
type Flags = byte
let f = Flags(byte(0xF0))
let high = Bits.shiftRight(f, 4)
// error: T would be Flags, which is not a member of its bound Integer; unwrap it with byte(...)
```

```ant
type Flags = byte
let f = Flags(byte(0xF0))
Flags(Bits.shiftRight(byte(f), 4)) // Flags(15)
```

## See also

[Bits.shiftLeft](shiftLeft.md) · [Bits.rotateRight](rotateRight.md) · [Bits.and](and.md) ·
[Bits.leadingZeros](leadingZeros.md) · [String.fromInt](../String/fromInt.md)
