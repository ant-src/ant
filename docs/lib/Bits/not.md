<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.not

```ant signature
func Bits.not[T: Integer](a: T): T
```

Bits.not returns a with every bit inverted.

<!-- end generated -->

## Details

- Every bit is inverted at the width of the argument's type, so the result depends on the type:
  `Bits.not(uint8(0))` is `255`, and `Bits.not(uint16(0))` is `65535`.
- On a signed type the bits are the two's complement ones, so the result is `-a - 1`, and
  `Bits.not(Math.minInt)` is `Math.maxInt`.

## Basic examples

```ant
Bits.not(uint8(0)) // 255
Bits.not(0)        // -1
Bits.not(5)        // -6
```

## Scope

Every integer type works, and the result has the argument's type:

```ant
Bits.not(int8(0))           // -1
Bits.not(int32(100))        // -101
Bits.not(byte(0b0000_1111)) // 240
Bits.not(uint16(0x00FF))    // 65280
Bits.not(uint32(0))         // 4294967295
```

The largest and smallest values swap with their opposites:

```ant
Bits.not(Math.maxUint64) // 0
Bits.not(Math.minInt)    // 9223372036854775807
Bits.not(Math.maxInt)    // -9223372036854775808
```

The argument is a named parameter, `a`:

```ant
Bits.not(a: 3) // -4
```

In a pipe, the value is the argument:

```ant
byte(0xF0) | Bits.not() // 15
```

## Applications

Clear a flag by combining the inverted flag with [Bits.and](and.md):

```ant
let READ = 0b100
let WRITE = 0b010
let mode = Bits.or(READ, WRITE)
let readOnly = Bits.and(mode, Bits.not(WRITE))
readOnly                       // 4
Bits.and(readOnly, WRITE) == 0 // true
```

Round down to a multiple of a power of two by clearing its low bits:

```ant
func alignDown(n: int, size: int): int {
  Bits.and(n, Bits.not(size - 1))
}

[0, 7, 8, 13, 4096].map({ alignDown(x, 8) }) // [0, 0, 8, 8, 4096]
```

Compute the ones' complement checksum of an IPv4 header. The sum folds its carry back in, and
`Bits.not` inverts it at 16 bits. A header with its checksum appended sums to zero:

```ant
func checksum(words: Array[uint16]): uint16 {
  var sum = 0
  for w in words {
    sum += int(w)
    if sum > 0xFFFF {
      sum = Bits.and(sum, 0xFFFF) + 1
    }
  }
  Bits.not(uint16(sum))
}

let header = [0x4500, 0x0073, 0x0000, 0x4000, 0x4011, 0x0000, 0xC0A8, 0x0001, 0xC0A8, 0x00C7]
  .map(uint16)
let sum = checksum(header)
println(sum == uint16(0xB861))
println(checksum(header.concat([sum])))
```

```text
true
0
```

## Properties and relations

Inverting twice gives back the argument:

```ant
for x in [0, 1, -1, 42, Math.minInt, Math.maxInt] {
  assert(Bits.not(Bits.not(x)) == x)
}
```

On a signed type, `Bits.not(x)` is `-x - 1`, and the same as [Bits.xor](xor.md) with `-1`, whose
bits are all set:

```ant
for x in [0, 1, -1, 42, -1000] {
  assert(Bits.not(x) == -x - 1)
  assert(Bits.not(x) == Bits.xor(x, -1))
}
```

On an unsigned type, `Bits.not(x)` is the largest value minus `x`:

```ant
for x in [byte(0), byte(1), byte(200), byte(255)] {
  assert(Bits.not(x) == byte(255) - x)
}
```

A value and its inverse share no set bits, and together set them all:

```ant
for x in [byte(0), byte(0b1011), byte(0xF0), byte(255)] {
  assert(Bits.and(x, Bits.not(x)) == byte(0))
  assert(Bits.or(x, Bits.not(x)) == byte(255))
  assert(Bits.popCount(x) + Bits.popCount(Bits.not(x)) == 8)
}
```

## Possible issues

On a signed type the sign bit flips too, so a non-negative argument gives a negative result. To
invert only the low bits, use [Bits.xor](xor.md) with a mask:

```ant
Bits.not(0b101)        // -6
Bits.xor(0b101, 0b111) // 2
```

An integer literal is always an `int`, so an inverted literal used as a mask for a smaller type is a
compile error. Convert the literal first:

```ant
let b = byte(0xF5)
let low = Bits.and(b, Bits.not(0x0F)) // error: cannot use int where byte is expected
```

```ant
let b = byte(0xF5)
Bits.and(b, Bits.not(byte(0x0F))) // 240
```

`Bits.not` takes only integers, so a float is a compile error:

```ant
let x = Bits.not(1.5) // error: T would be float64, which is not a member of its bound Integer
```

`Bits` is not the namespace of the integer types, so `x.not()` is a compile error. Write
`Bits.not(x)`, or pipe the value into it:

```ant
let x = 5
let y = x.not() // error: int has no field or function not
```

Ant has no `~` operator, and `!` takes only a `bool`:

```ant
let x = 5
let y = !x // error: cannot use int where bool is expected
```

## See also

[Bits.and](and.md) · [Bits.or](or.md) · [Bits.xor](xor.md) · [Bits.popCount](popCount.md) ·
[Bits](index.md) · [Logical operators](../../language/operators/logical.md)
