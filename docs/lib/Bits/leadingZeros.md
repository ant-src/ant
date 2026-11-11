<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.leadingZeros

```ant signature
func Bits.leadingZeros[T: Integer](a: T): int
```

Bits.leadingZeros returns the number of zero bits above the highest one bit of a, counted at the bit
width of T over its two's complement bits. It is the full width for 0.

<!-- end generated -->

## Details

- A negative signed value has its top bit set, so it has no leading zeros.

## Basic examples

```ant
Bits.leadingZeros(uint8(1))    // 7
Bits.leadingZeros(int8(-1))    // 0
Bits.leadingZeros(0b0001_0000) // 59
```

## Scope

Every integer type works, each at its own width:

```ant
Bits.leadingZeros(uint8(1))  // 7
Bits.leadingZeros(int16(1))  // 15
Bits.leadingZeros(uint32(1)) // 31
Bits.leadingZeros(1)         // 63
```

Zero gives the width of its type:

```ant
Bits.leadingZeros(uint8(0)) // 8
Bits.leadingZeros(int32(0)) // 32
Bits.leadingZeros(0)        // 64
```

Negative values count over their two's complement bits, whose top bit is set:

```ant
Bits.leadingZeros(-1)          // 0
Bits.leadingZeros(int8(-128))  // 0
Bits.leadingZeros(Math.minInt) // 0
```

The largest signed value has one leading zero, its sign bit; the largest unsigned value has none:

```ant
Bits.leadingZeros(Math.maxInt)    // 1
Bits.leadingZeros(int8(127))      // 1
Bits.leadingZeros(Math.maxUint64) // 0
```

Pass `Bits.leadingZeros` as a function value:

```ant
[1, 2, 255, 256].map(Bits.leadingZeros) // [63, 62, 56, 55]
```

In a pipe:

```ant
uint8(1) | Bits.leadingZeros() // 7
```

## Named arguments

```ant
Bits.leadingZeros(a: uint16(255)) // 8
```

## Applications

The number of bits a non-negative value needs is the width minus its leading zeros:

```ant
func bitLength(x: int): int {
  64 - Bits.leadingZeros(x)
}

bitLength(0)   // 0
bitLength(1)   // 1
bitLength(255) // 8
bitLength(256) // 9
```

The integer part of the base-2 logarithm, without floats:

```ant
func log2Floor(x: int): int {
  if x <= 0 { panic("log2Floor: ${x} is not positive") }
  63 - Bits.leadingZeros(x)
}

log2Floor(1)    // 0
log2Floor(1000) // 9
log2Floor(1024) // 10
```

Round a size up to the next power of two, as a hash table does for its bucket count:

```ant
func nextPowerOfTwo(x: int): int {
  if x <= 1 { return 1 }
  Bits.shiftLeft(1, 64 - Bits.leadingZeros(x - 1))
}

[1, 2, 3, 5, 17, 64].map(nextPowerOfTwo) // [1, 2, 4, 8, 32, 64]
```

## Properties and relations

A single set bit at position `k` has `width - 1 - k` leading zeros:

```ant
for k in Iter.range(0, 64) {
  assert(Bits.leadingZeros(Bits.shiftLeft(1, k)) == 63 - k)
}
```

Shifting a positive value left by one removes one leading zero, and shifting right adds one:

```ant
let x = 1000
assert(Bits.leadingZeros(Bits.shiftLeft(x, 1)) == Bits.leadingZeros(x) - 1)
assert(Bits.leadingZeros(Bits.shiftRight(x, 1)) == Bits.leadingZeros(x) + 1)
```

Zero is the one value whose leading and trailing zeros are both the full width:

```ant
let width = 32
assert(Bits.leadingZeros(int32(0)) == width)
assert(Bits.trailingZeros(int32(0)) == width)
```

`Bits.not` turns the leading zeros of a non-negative value into leading ones, so the result is
negative and has none:

```ant
let x = 42
assert(Bits.leadingZeros(x) > 0)
assert(Bits.leadingZeros(Bits.not(x)) == 0)
```

## Possible issues

An integer literal is always an `int`, so it counts at 64 bits. Convert it to count at a smaller
width:

```ant
Bits.leadingZeros(255)        // 56
Bits.leadingZeros(uint8(255)) // 0
```

`Bits.leadingZeros` is not a function on the integer types, so a dot call does not find it:

```ant
let mask = uint8(16)
let n = mask.leadingZeros()
// error: byte has no field or function leadingZeros
```

A named primitive type is not an `Integer`. Unwrap it first:

```ant
type Flags = uint8
let flags = Flags(byte(3))
let n = Bits.leadingZeros(flags)
// error: T would be Flags, which is not a member of its bound Integer; unwrap it with byte(...)
```

```ant
type Flags = uint8
let flags = Flags(byte(3))
Bits.leadingZeros(byte(flags)) // 6
```

A value typed `Integer` must be narrowed to one integer type first:

```ant
let n: Integer = 5
let zeros = Bits.leadingZeros(n)
// error: T would be Integer, which is not a single member of its bound Integer; narrow first
```

## See also

[Bits.trailingZeros](trailingZeros.md), [Bits.popCount](popCount.md),
[Bits.shiftLeft](shiftLeft.md), [Bits.shiftRight](shiftRight.md), [Math.log2](../Math/log2.md)
