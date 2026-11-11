<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.or

```ant signature
func Bits.or[T: Integer](a: T, b: T): T
```

Bits.or returns the bitwise OR of a and b.

<!-- end generated -->

## Details

- A bit of the result is set when it is set in `a`, in `b`, or in both.
- Signed values work on their two's complement bits, so a negative argument sets the high bits.
- Bitwise operations are functions, and `|` is the pipe. Use `Bits.or` for bitwise OR and `||` for
  `bool`.

## Basic examples

```ant
Bits.or(0b1100, 0b1010) // 14
Bits.or(12, 10)         // 14
Bits.or(0, 0)           // 0
```

Show the result in binary with [String.fromInt](../String/fromInt.md):

```ant
String.fromInt(Bits.or(0b1100, 0b1010), base: 2) // "1110"
```

## Scope

Every integer type works, and the result keeps its type:

```ant
Bits.or(byte(0xF0), byte(0x0F))       // 255
Bits.or(uint16(0x1200), uint16(0x34)) // 4660
Bits.or(int8(-128), int8(1))          // -127
```

Negative values are two's complement, so `-1` has every bit set:

```ant
Bits.or(-1, 42)                   // -1
Bits.or(Math.minInt, Math.maxInt) // -1
Bits.or(-8, 3)                    // -5
```

The top bit of an unsigned value is an ordinary bit:

```ant
Bits.or(uint64(1), Bits.shiftLeft(uint64(1), 63)) // 9223372036854775809
```

In a pipe, the value on the left becomes `a`:

```ant
0b0001 | Bits.or(0b0100) // 5
```

Pass `Bits.or` as a function value to combine an array of bits:

```ant
[1, 2, 4, 8].reduce(0, Bits.or) // 15
```

## Named arguments

```ant
Bits.or(a: 1, b: 4) // 5
Bits.or(b: 4, a: 1) // 5
```

## Applications

Combine permission flags, and test one with [Bits.and](and.md):

```ant
let read = 0b100
let write = 0b010
let exec = 0b001
let mode = Bits.or(read, write)
mode                       // 6
Bits.and(mode, write) != 0 // true
Bits.and(mode, exec) != 0  // false
```

Set bit `n` of a value, leaving the others as they are:

```ant
func setBit(x: int, n: int): int {
  Bits.or(x, Bits.shiftLeft(1, n))
}

setBit(0, 3)      // 8
setBit(0b1000, 0) // 9
setBit(9, 3)      // 9
```

Join two bytes into a 16-bit number, high byte first:

```ant
func word(hi: byte, lo: byte): uint16 {
  Bits.or(Bits.shiftLeft(uint16(hi), 8), uint16(lo))
}

word(byte(0x12), byte(0x34)) // 4660
word(byte(0), byte(0xFF))    // 255
```

Pack the channels of a color into one `0xRRGGBB` number:

```ant
type Rgb = { r: int, g: int, b: int }

func Rgb.pack(c: Rgb): int {
  [Bits.shiftLeft(c.r, 16), Bits.shiftLeft(c.g, 8), c.b].reduce(0, Bits.or)
}

let orange = Rgb{ r: 255, g: 165, b: 0 }
String.fromInt(orange.pack(), base: 16) // "ffa500"
```

## Properties and relations

`Bits.or` is commutative and associative, `0` changes nothing, `-1` absorbs everything, and OR with
itself gives the value back:

```ant
for [a, b, c] in [[12, 10, 3], [-8, 3, 0], [0, 255, -1]] {
  assert(Bits.or(a, b) == Bits.or(b, a))
  assert(Bits.or(Bits.or(a, b), c) == Bits.or(a, Bits.or(b, c)))
  assert(Bits.or(a, 0) == a)
  assert(Bits.or(a, -1) == -1)
  assert(Bits.or(a, a) == a)
}
```

It is the [Bits.xor](xor.md) bits plus the [Bits.and](and.md) bits, and by De Morgan's law the
[Bits.not](not.md) of the AND of the complements:

```ant
for [a, b] in [[12, 10], [-8, 3], [Math.minInt, 1]] {
  assert(Bits.or(a, b) == Bits.xor(a, b) + Bits.and(a, b))
  assert(Bits.or(a, b) == Bits.not(Bits.and(Bits.not(a), Bits.not(b))))
}
```

For values with no bit in common, OR is addition:

```ant
let a = 0b1000
let b = 0b0011
assert(Bits.and(a, b) == 0)
assert(Bits.or(a, b) == a + b)
```

## Possible issues

`|` is the pipe, not bitwise OR:

```ant
let flags = 5 | 3 // error: the right side of a pipe must be a call
```

`Bits.or` is not in the namespace of `int`, so a dot call does not reach it. Call it in namespace
form or in a pipe:

```ant
let x = 5
let y = x.or(3) // error: int has no field or function or
```

```ant
let x = 5
x | Bits.or(3) // 7
```

Both arguments need the same type. An integer literal is always an `int`, so convert it:

```ant
let mask = byte(0xF0)
let m = Bits.or(mask, 1) // error: cannot use int where byte is expected
```

```ant
let mask = byte(0xF0)
Bits.or(mask, byte(1)) // 241
```

A named primitive type is not a member of `Integer`. Unwrap it, then tag the result again:

```ant
type Flags = int
let f = Bits.or(Flags(1), Flags(4))
// error: T would be Flags, which is not a member of its bound Integer; unwrap it with int(...)
```

```ant
type Flags = int
let f = Flags(1)
Flags(Bits.or(int(f), 4)) // Flags(5)
```

`bool` is not an integer. Use `||`:

```ant
let both = Bits.or(true, false) // error: T would be bool, which is not a member of its bound Integer
```

## See also

[Bits.and](and.md) · [Bits.xor](xor.md) · [Bits.not](not.md) · [Bits.shiftLeft](shiftLeft.md) ·
[Bits namespace](index.md)
