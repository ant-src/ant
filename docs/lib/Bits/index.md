<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits

Bits holds the bitwise operations and wrapping arithmetic over the integer types, which have no
operators for them. A signed value is taken as its two's complement bits.

<!-- end generated -->

## Details

- Where other languages write `a & b`, `a | b`, `a ^ b`, `~a`, `a << n`, and `a >> n`, Ant calls
  [Bits.and](and.md), [Bits.or](or.md), [Bits.xor](xor.md), [Bits.not](not.md),
  [Bits.shiftLeft](shiftLeft.md), and [Bits.shiftRight](shiftRight.md). In Ant, `|` is the pipe.
- The two operands of `Bits.and`, `Bits.or`, `Bits.xor`, and the wrap functions share one integer
  type. The count of a shift or rotation is always an `int`, so `Bits.shiftLeft(uint8(1), 3)` needs
  no conversion.
- Negative values take part as their two's complement bits at the width of their type, so `-1` has
  every bit set.
- A shift panics when its count is negative or not below the width of the type. A rotation takes its
  count modulo the width and panics only on a negative count.
- `+`, `-`, and `*` panic on overflow. [Bits.wrapAdd](wrapAdd.md), [Bits.wrapSub](wrapSub.md), and
  [Bits.wrapMul](wrapMul.md) wrap around instead and never panic.

## Basic examples

Combine, mask, and flip bits:

```ant
Bits.and(0b1100, 0b1010) // 8
Bits.or(0b1100, 0b1010)  // 14
Bits.xor(0b1100, 0b1010) // 6
Bits.not(uint8(0))       // 255
```

Shift and rotate:

```ant
Bits.shiftLeft(1, 10)          // 1024
Bits.shiftRight(0x1234, 8)     // 18
Bits.rotateLeft(uint8(129), 1) // 3
Bits.rotateRight(uint8(1), 1)  // 128
```

Count bits:

```ant
Bits.popCount(0b1011)       // 3
Bits.leadingZeros(uint8(1)) // 7
Bits.trailingZeros(0b1000)  // 3
```

Add past the maximum without a panic:

```ant
Bits.wrapAdd(uint8(255), uint8(1)) // 0
uint8(255) + uint8(1)              // panic: integer overflow
```

## Scope

The type of the operands sets the width. The same bits give different answers at different widths:

```ant
Bits.not(uint8(0))          // 255
Bits.not(uint16(0))         // 65535
Bits.not(0)                 // -1
Bits.leadingZeros(uint8(1)) // 7
Bits.leadingZeros(1)        // 63
```

A right shift copies the sign bit of a signed type and fills an unsigned one with zeros:

```ant
Bits.shiftRight(int8(-8), 1)   // -4
Bits.shiftRight(uint8(248), 1) // 124
```

Negative values count their two's complement bits:

```ant
Bits.popCount(int8(-1))     // 8
Bits.popCount(-1)           // 64
Bits.trailingZeros(-8)      // 3
Bits.leadingZeros(int8(-1)) // 0
```

A left shift drops the bits pushed past the width; only the count can make it panic:

```ant
Bits.shiftLeft(uint8(200), 1) // 144
Bits.shiftLeft(1, 63)         // -9223372036854775808
```

Write the bits out with `String.fromInt` in base 2:

```ant
String.fromInt(0b1011, base: 2) // "1011"
```

Zero has no set bit, so both zero counts are the full width:

```ant
Bits.leadingZeros(uint16(0))  // 16
Bits.trailingZeros(uint16(0)) // 16
```

The functions chain through a pipe, where the value on the left is the first argument:

```ant
0x3366CC | Bits.shiftRight(8) | Bits.and(0xFF) // 102
```

## Functions by job

### Logic

- [Bits.and](and.md): the bits set in both `a` and `b`
- [Bits.or](or.md): the bits set in either `a` or `b`
- [Bits.xor](xor.md): the bits set in exactly one of `a` and `b`
- [Bits.not](not.md): `a` with every bit inverted

### Shifts and rotations

- [Bits.shiftLeft](shiftLeft.md): `a` moved left by `n` bits, zeros shifted in
- [Bits.shiftRight](shiftRight.md): `a` moved right by `n` bits, arithmetic for a signed type and
  logical for an unsigned one
- [Bits.rotateLeft](rotateLeft.md), [Bits.rotateRight](rotateRight.md): `a` turned by `n` bits, the
  bits shifted out coming back in at the other end

### Counting

- [Bits.popCount](popCount.md): the number of one bits
- [Bits.leadingZeros](leadingZeros.md): the zero bits above the highest one bit
- [Bits.trailingZeros](trailingZeros.md): the zero bits below the lowest one bit

### Wrapping arithmetic

- [Bits.wrapAdd](wrapAdd.md), [Bits.wrapSub](wrapSub.md), [Bits.wrapMul](wrapMul.md): `a + b`,
  `a - b`, and `a * b`, wrapping around in two's complement instead of panicking on overflow

## Applications

Keep a set of permissions as flags in one integer:

```ant
let READ = 0b100
let WRITE = 0b010
let EXEC = 0b001
func has(mode: int, flag: int): bool {
  Bits.and(mode, flag) != 0
}
let mode = Bits.or(READ, EXEC)
has(mode, READ)  // true
has(mode, WRITE) // false
```

Clear a flag by ANDing with its inverse, and toggle one with `Bits.xor`:

```ant
let READ = 0b100
let EXEC = 0b001
let mode = 0b101
Bits.and(mode, Bits.not(EXEC)) // 4
Bits.xor(mode, READ)           // 1
```

Pack a color into one integer and take it apart again:

```ant
type Rgb = { r: int, g: int, b: int }
func pack(c: Rgb): int {
  Bits.shiftLeft(c.r, 16) | Bits.or(Bits.shiftLeft(c.g, 8)) | Bits.or(c.b)
}
func unpack(color: int): Rgb {
  Rgb{
    r: color | Bits.shiftRight(16) | Bits.and(0xFF),
    g: color | Bits.shiftRight(8) | Bits.and(0xFF),
    b: color | Bits.and(0xFF),
  }
}
pack(Rgb{ r: 51, g: 102, b: 204 }) // 3368652
unpack(0x3366CC)                   // Rgb{ r: 51, g: 102, b: 204 }
```

Count the bits that differ between two values, their Hamming distance:

```ant
func distance(a: int, b: int): int {
  Bits.popCount(Bits.xor(a, b))
}
distance(0b1011, 0b1110) // 2
```

Hash a string with FNV-1a, whose multiplication is meant to overflow:

```ant
func fnv1a(s: string): uint32 {
  var hash = uint32(2166136261)
  for b in s.toBytes() {
    hash = Bits.wrapMul(Bits.xor(hash, uint32(b)), uint32(16777619))
  }
  hash
}
fnv1a("")    // 2166136261
fnv1a("ant") // 522836950
```

## Properties and relations

A left shift by `n` multiplies by `2^n` while the result fits, and a right shift of a non-negative
value divides by it:

```ant
var power = 1
for n in Iter.range(0, 8) {
  assert(Bits.shiftLeft(3, n) == 3 * power)
  assert(Bits.shiftRight(3 * power, n) == 3)
  power *= 2
}
```

`Bits.xor` with the same value twice gives the value back:

```ant
let key = 0b1010_0101
for x in [0, 7, 200] {
  assert(Bits.xor(Bits.xor(x, key), key) == x)
}
```

A rotation keeps every bit, so it keeps the count of one bits:

```ant
let x = uint8(0b1001_0110)
for n in Iter.range(0, 8) {
  assert(Bits.popCount(Bits.rotateLeft(x, n)) == Bits.popCount(x))
  assert(Bits.rotateRight(Bits.rotateLeft(x, n), n) == x)
}
```

`Bits.wrapSub` undoes `Bits.wrapAdd`, even across the wrap:

```ant
let a = uint8(250)
let b = uint8(10)
assert(Bits.wrapSub(Bits.wrapAdd(a, b), b) == a)
```

## Possible issues

Ant has no `&` operator, and `|` is the pipe, not bitwise OR:

```ant
let low = 0x1234 & 0xFF // error: unexpected character '&'
```

Integers have no namespace, so a dot call does not reach `Bits`. Use the namespace form or a pipe:

```ant
let flags = 5
let low = flags.and(1) // error: int has no field or function and
```

Both operands share one type. An integer literal is an `int`, so convert it:

```ant
let b = uint8(0xF0)
let low = Bits.and(b, 0x0F) // error: cannot use int where byte is expected
```

A shift by the full width panics instead of giving zero:

```ant
Bits.shiftLeft(uint8(1), 8)
// panic: Bits.shiftLeft needs a count from 0 up to but excluding 8, got 8
```

## See also

- [Math](../Math/index.md): numeric functions, the integer limits, and the `Integer` bound
- [Bytes](../Bytes/index.md): binary data, a mutable sequence of `byte`
- [String.fromInt](../String/fromInt.md): an integer's digits in any base, such as base 2 to show
  its bits
