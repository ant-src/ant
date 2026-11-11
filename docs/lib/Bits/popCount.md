<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.popCount

```ant signature
func Bits.popCount[T: Integer](a: T): int
```

Bits.popCount returns the number of one bits in a.

<!-- end generated -->

## Details

- `Bits.popCount` counts the bits of `a` at the width of its type, in two's complement, so a
  negative signed value counts its sign bits: `Bits.popCount(-1)` is 64, and
  `Bits.popCount(int8(-1))` is 8.

## Basic examples

```ant
Bits.popCount(0b1011) // 3
Bits.popCount(0)      // 0
Bits.popCount(255)    // 8
```

## Scope

Every integer type works. The same value gives the same count at any width while it is not negative:

```ant
Bits.popCount(byte(200))      // 3
Bits.popCount(uint16(200))    // 3
Bits.popCount(int32(200))     // 3
Bits.popCount(Math.maxUint64) // 64
```

A negative value counts the sign bits that fill its type, so the count depends on the width:

```ant
Bits.popCount(int8(-1))    // 8
Bits.popCount(int16(-1))   // 16
Bits.popCount(-1)          // 64
Bits.popCount(int8(-128))  // 1
Bits.popCount(Math.minInt) // 1
```

In a pipe, the value is the argument:

```ant
0xF0F0 | Bits.popCount() // 8
```

Pass it by name to count the bits of every element:

```ant
[1, 2, 3, 7, 8].map(Bits.popCount) // [1, 1, 2, 3, 1]
```

## Applications

Count the flags set in a permission mask:

```ant
let read = 0b001
let write = 0b010
let exec = 0b100
let mode = Bits.or(read, exec)
println("${Bits.popCount(mode)} of 3 permissions")
```

```text
2 of 3 permissions
```

Measure the Hamming distance between two equal-length byte strings, the number of bits that differ:

```ant
func hamming(a: Bytes, b: Bytes): int {
  var n = 0
  for i in Iter.range(0, a.length()) {
    n += Bits.popCount(Bits.xor(a[i], b[i]))
  }
  n
}

println(hamming("karolin".toBytes(), "kathrin".toBytes()))
```

```text
9
```

Test whether a positive number is a power of two:

```ant
func isPowerOfTwo(n: int): bool {
  n > 0 && Bits.popCount(n) == 1
}

println([1, 6, 64, 0, -8].filter(isPowerOfTwo))
```

```text
[1, 64]
```

Compute an even-parity bit, which makes the total number of one bits even:

```ant
func parity(b: byte): int {
  Bits.popCount(b) % 2
}
parity(byte(0b1011_0001)) // 0
parity(byte(0b1011_0011)) // 1
```

## Properties and relations

A value and its complement together have every bit of the type set:

```ant
let values = [int8(0), int8(5), int8(-3), int8(127)]
for v in values {
  assert(Bits.popCount(v) + Bits.popCount(Bits.not(v)) == 8)
}
```

A rotation moves bits without losing any, so it keeps the count:

```ant
let v = uint32(0x8000_00F1)
for n in [1, 7, 31, 40] {
  assert(Bits.popCount(Bits.rotateLeft(v, n)) == Bits.popCount(v))
}
```

The bits either value has plus the bits both have equal the bits of each counted separately, and the
bits only one has are the difference:

```ant
let a = 0b1100_1010
let b = 0b1010_0110
let either = Bits.popCount(Bits.or(a, b))
let both = Bits.popCount(Bits.and(a, b))
assert(either + both == Bits.popCount(a) + Bits.popCount(b))
assert(Bits.popCount(Bits.xor(a, b)) == either - both)
```

## Possible issues

`Bits.popCount` is not a function of the `int` namespace, so it cannot be called through `.`. Write
the namespace, or use a pipe:

```ant
let mask = 0b0110
mask.popCount() // error: int has no field or function popCount
```

A negative `int` counts all 64 bits of its two's complement form, not the bits of its magnitude.
Count the magnitude with `Math.abs` when that is what you want:

```ant
Bits.popCount(-6)           // 62
Bits.popCount(Math.abs(-6)) // 2
```

## See also

[Bits](index.md) · [Bits.leadingZeros](leadingZeros.md) · [Bits.trailingZeros](trailingZeros.md) ·
[Bits.and](and.md) · [Bits.xor](xor.md) · [Bits.not](not.md)
