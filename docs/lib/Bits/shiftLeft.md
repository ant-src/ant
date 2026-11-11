<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.shiftLeft

```ant signature
func Bits.shiftLeft[T: Integer](a: T, n: int): T
```

Bits.shiftLeft returns a shifted left by n bits, filling with zeros. It panics if n \< 0 or n is at
least the bit width of T.

<!-- end generated -->

## Details

- Bits shifted past the top of the type are dropped, so the result never panics on overflow, unlike
  `*`. For a signed type, a bit shifted into the sign position makes the result negative.

## Basic examples

Shift 1 left by 4 bits, which multiplies it by 2⁴:

```ant
Bits.shiftLeft(1, 4) // 16
Bits.shiftLeft(3, 2) // 12
```

The result has the type of `a`:

```ant
Bits.shiftLeft(byte(1), 7) // 128
```

## Scope

A count of zero returns `a` unchanged:

```ant
Bits.shiftLeft(5, 0) // 5
```

Bits that leave the top of the type are lost; zeros come in at the bottom:

```ant
Bits.shiftLeft(byte(255), 4)      // 240
Bits.shiftLeft(uint16(0x1234), 8) // 13312
```

On a signed type the sign bit is an ordinary bit, so a shift can change the sign:

```ant
Bits.shiftLeft(int8(64), 1)  // -128
Bits.shiftLeft(-3, 2)        // -12
Bits.shiftLeft(int32(1), 31) // -2147483648
```

The largest count is one less than the width of the type:

```ant
Bits.shiftLeft(1, 63)         // -9223372036854775808
Bits.shiftLeft(uint64(1), 63) // 9223372036854775808
```

In a pipe, the value is the first argument:

```ant
1 | Bits.shiftLeft(10) // 1024
```

The arguments can be named:

```ant
Bits.shiftLeft(a: 1, n: 3) // 8
```

## Applications

Define flags as single bits and combine them with [`Bits.or`](or.md):

```ant
let READ = Bits.shiftLeft(1, 0)
let WRITE = Bits.shiftLeft(1, 1)
let EXEC = Bits.shiftLeft(1, 2)
Bits.or(READ, EXEC)                       // 5
Bits.and(Bits.or(READ, WRITE), EXEC) == 0 // true
```

Build a mask of the low `n` bits:

```ant
func lowMask(n: int): uint32 {
  Bits.shiftLeft(uint32(1), n) - uint32(1)
}

println(lowMask(4))
println(lowMask(12))
```

```text
15
4095
```

Pack a color into one integer, one byte per channel:

```ant
type Rgb = { r: byte, g: byte, b: byte }

func pack(c: Rgb): uint32 {
  let r = Bits.shiftLeft(uint32(c.r), 16)
  let g = Bits.shiftLeft(uint32(c.g), 8)
  Bits.or(Bits.or(r, g), uint32(c.b))
}

println(String.fromInt(pack(Rgb{ r: byte(255), g: byte(128), b: byte(0) }), base: 16))
```

```text
ff8000
```

Turn an array of bits, most significant first, into a number:

```ant
let bits = [1, 0, 1, 1]
bits.reduce(0, func(n: int, bit: int) { n | Bits.shiftLeft(1) | Bits.or(bit) }) // 11
```

Test whether bit `i` is set, with the index in a variable:

```ant
func hasBit(flags: int, i: int): bool {
  Bits.and(flags, Bits.shiftLeft(1, i)) != 0
}

println(hasBit(5, 2))
println(hasBit(5, 1))
```

```text
true
false
```

## Properties and relations

While the product fits the type, shifting by `n` multiplies by 2ⁿ:

```ant
let a = 7
for n in Iter.range(0, 10) {
  assert(Bits.shiftLeft(a, n) == a * Bits.shiftLeft(1, n))
}
```

On an unsigned type, [`Bits.shiftRight`](shiftRight.md) undoes a shift that lost no bits:

```ant
let a = byte(0b0001_0110)
assert(Bits.shiftRight(Bits.shiftLeft(a, 3), 3) == a)
```

On a signed type `Bits.shiftRight` copies the sign bit, so a shift that moved a 1 into the sign
position does not come back:

```ant
Bits.shiftRight(Bits.shiftLeft(int8(64), 1), 1) // -64
```

Two shifts add their counts, as long as the sum stays below the width:

```ant
let a = uint16(0xBEEF)
assert(Bits.shiftLeft(Bits.shiftLeft(a, 3), 5) == Bits.shiftLeft(a, 8))
```

## Possible issues

A count equal to the width of the type panics; it does not give zero:

```ant
Bits.shiftLeft(byte(1), 8) // panic: Bits.shiftLeft needs a count from 0 up to but excluding 8, got 8
```

A negative count panics instead of shifting right:

```ant
Bits.shiftLeft(1, -1) // panic: Bits.shiftLeft needs a count from 0 up to but excluding 64, got -1
```

Where `*` panics on overflow, the shift drops the high bits without a word:

```ant
Bits.shiftLeft(Math.maxInt, 1) // -2
Math.maxInt * 2                // panic: integer overflow
```

The count must be an `int`, even when `a` is a smaller type:

```ant
let n = byte(3)
Bits.shiftLeft(1, n)
// error: cannot use byte where int is expected
```

Integers have no `shiftLeft` of their own, so a dot call does not compile. Use the namespace form or
a pipe:

```ant
let x = 5
x.shiftLeft(1)
// error: int has no field or function shiftLeft
```

## See also

[Bits.shiftRight](shiftRight.md) · [Bits.rotateLeft](rotateLeft.md) · [Bits.or](or.md) ·
[Bits.and](and.md) · [Bits.leadingZeros](leadingZeros.md) · [Bits.wrapMul](wrapMul.md) ·
[Bits](index.md)
