<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.xor

```ant signature
func Bits.xor[T: Integer](a: T, b: T): T
```

Bits.xor returns the bitwise exclusive OR of a and b.

<!-- end generated -->

## Details

- A bit of the result is set when it is set in exactly one of `a` and `b`.
- Negative values take part as their two's complement bits, so XOR with `-1` inverts every bit.
- It never panics.

## Basic examples

Keep the bits that differ between two numbers:

```ant
Bits.xor(0b1100, 0b1010) // 6
Bits.xor(12, 10)         // 6
Bits.xor(7, 7)           // 0
```

Show the result in binary with [String.fromInt](../String/fromInt.md):

```ant
String.fromInt(Bits.xor(0b1100, 0b1010), base: 2) // "110"
```

## Scope

Every integer type works, and the result keeps its type:

```ant
Bits.xor(byte(0xFF), byte(0x0F))       // 240
Bits.xor(uint16(0x1234), uint16(0xFF)) // 4811
Bits.xor(int8(-128), int8(127))        // -1
Bits.xor(uint64(1), Math.maxUint64)    // 18446744073709551614
```

Negative values are two's complement. XOR with `-1` flips every bit, and the sign bit flips like any
other:

```ant
Bits.xor(-1, 5)                    // -6
Bits.xor(Math.minInt, Math.maxInt) // -1
Bits.xor(-8, 3)                    // -5
```

In a pipe, the value on the left becomes `a`:

```ant
0b0110 | Bits.xor(0b0011) // 5
```

Pass `Bits.xor` as a function value to combine an array:

```ant
[0b0001, 0b0011, 0b0111].reduce(0, Bits.xor) // 5
```

## Applications

Toggle a flag: the first call sets it, the second clears it, and the other bits stay:

```ant
let bold = 0b001
let italic = 0b010
var style = bold
style = Bits.xor(style, italic)
println(style)
style = Bits.xor(style, italic)
println(style)
```

```text
3
1
```

Find the one value without a pair, since each pair cancels out:

```ant
[4, 9, 7, 4, 7].reduce(0, Bits.xor) // 9
```

Count the bits in which two values differ, their Hamming distance, with
[Bits.popCount](popCount.md):

```ant
func distance(a: int, b: int): int {
  Bits.popCount(Bits.xor(a, b))
}

distance(0b1011, 0b1001) // 1
distance(0, 255)         // 8
distance(42, 42)         // 0
```

Tell whether two integers have opposite signs: the sign bit of their XOR is set exactly then:

```ant
func oppositeSigns(a: int, b: int): bool {
  Bits.xor(a, b) < 0
}

oppositeSigns(5, -3)  // true
oppositeSigns(-5, -3) // false
oppositeSigns(0, 3)   // false
```

Scramble bytes with a key and get them back with the same key:

```ant
func scramble(data: Bytes, key: byte): Bytes {
  data.toArray().map({ Bits.xor(x, key) }) | Bytes.fromArray()
}

let key = byte(0x5A)
let secret = "Ant".toBytes() | scramble(key)
secret.toArray()                            // [27, 52, 46]
secret | scramble(key) | String.fromBytes() // "Ant"
```

## Properties and relations

`Bits.xor` is commutative and associative, `0` changes nothing, and a value XORed with itself is
`0`:

```ant
for [a, b, c] in [[12, 10, 3], [-8, 3, 0], [0, 255, -1]] {
  assert(Bits.xor(a, b) == Bits.xor(b, a))
  assert(Bits.xor(Bits.xor(a, b), c) == Bits.xor(a, Bits.xor(b, c)))
  assert(Bits.xor(a, 0) == a)
  assert(Bits.xor(a, a) == 0)
}
```

XOR with `-1` is [Bits.not](not.md), and XOR is the [Bits.or](or.md) bits without the
[Bits.and](and.md) bits:

```ant
for [a, b] in [[12, 10], [-8, 3], [Math.minInt, 1]] {
  assert(Bits.xor(a, -1) == Bits.not(a))
  assert(Bits.xor(a, b) == Bits.and(Bits.or(a, b), Bits.not(Bits.and(a, b))))
}
```

Applying the same key twice undoes it, at every integer type:

```ant
let x = uint8(0b1010_0110)
let k = uint8(0b0011_1100)
assert(Bits.xor(Bits.xor(x, k), k) == x)
```

## Possible issues

Ant has no `^` operator; bitwise XOR is this function:

```ant
let x = 5 ^ 3 // error: unexpected character '^'
```

`Bits.xor` is not in the namespace of `int`, so a dot call does not reach it. Call it in namespace
form or in a pipe:

```ant
let x = 5
let y = x.xor(3) // error: int has no field or function xor
```

```ant
let x = 5
x | Bits.xor(3) // 6
```

Both arguments need the same type. An integer literal is always an `int`, so convert it:

```ant
let b = byte(0xF0)
let y = Bits.xor(b, 1) // error: cannot use int where byte is expected
```

```ant
let b = byte(0xF0)
Bits.xor(b, byte(1)) // 241
```

A named primitive type is not a member of `Integer`. Unwrap it, then tag the result again:

```ant
type Flags = int
let f = Bits.xor(Flags(1), Flags(4))
// error: T would be Flags, which is not a member of its bound Integer; unwrap it with int(...)
```

```ant
type Flags = int
let f = Flags(5)
Flags(Bits.xor(int(f), 4)) // Flags(1)
```

`bool` is not an integer. For exclusive OR of two `bool` values, compare them with `!=`:

```ant
let t = Bits.xor(true, false) // error: T would be bool, which is not a member of its bound Integer
```

```ant
true != false // true
true != true  // false
```

## See also

[Bits.and](and.md) · [Bits.or](or.md) · [Bits.not](not.md) · [Bits.popCount](popCount.md) ·
[Bits namespace](index.md)
