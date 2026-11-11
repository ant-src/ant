<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.and

```ant signature
func Bits.and[T: Integer](a: T, b: T): T
```

Bits.and returns the bitwise AND of a and b.

<!-- end generated -->

## Details

- Each bit of the result is set when that bit is set in both `a` and `b`.
- Negative values take part as their two's complement bits, so `-1` has every bit set.

## Basic examples

Keep the bits that are set in both numbers:

```ant
Bits.and(0b1100, 0b1010) // 8
Bits.and(12, 10)         // 8
```

Mask off all but the low byte:

```ant
Bits.and(0x1234, 0xFF) // 52
```

## Scope

Every integer type works, with both arguments of that type:

```ant
Bits.and(uint8(0xF0), uint8(0x3C)) // 48
Bits.and(int8(-1), int8(5))        // 5
Bits.and(uint64(6), uint64(3))     // 2
```

Negative numbers use their two's complement bits. `-1` keeps every bit of the other operand, and
`-8` clears the low three:

```ant
Bits.and(-1, 42)          // 42
Bits.and(-8, 7)           // 0
Bits.and(-8, 29)          // 24
Bits.and(Math.maxInt, -1) // 9223372036854775807
```

Zero clears everything:

```ant
Bits.and(0, 99) // 0
```

Call it in a pipe, where the value on the left becomes `a`:

```ant
0x1234 | Bits.and(0xFF) // 52
```

Call it inside a lambda to test each element:

```ant
[1, 2, 3, 4, 5, 6].filter({ Bits.and(x, 1) == 0 }) // [2, 4, 6]
```

## Named arguments

Name the arguments when the mask should stand out:

```ant
Bits.and(a: 0b0110, b: 0b0011) // 2
```

## Applications

Test whether a flag is set in a set of flags:

```ant
let READ = 0b100
let WRITE = 0b010
let EXEC = 0b001
func has(flags: int, flag: int): bool {
  Bits.and(flags, flag) != 0
}
let mode = Bits.or(READ, EXEC)
has(mode, READ)  // true
has(mode, WRITE) // false
```

Check that all the flags of a mask are set, not only one of them:

```ant
let required = 0b0110
Bits.and(0b0111, required) == required // true
Bits.and(0b0101, required) == required // false
```

Split a color packed as `0xRRGGBB` into its channels:

```ant
type Rgb = { r: int, g: int, b: int }
func unpack(color: int): Rgb {
  Rgb{
    r: color | Bits.shiftRight(16) | Bits.and(0xFF),
    g: color | Bits.shiftRight(8) | Bits.and(0xFF),
    b: color | Bits.and(0xFF),
  }
}
unpack(0x3366CC) // Rgb{ r: 51, g: 102, b: 204 }
```

Round a size down to a multiple of a power of two by clearing its low bits:

```ant
func alignDown(n: int, align: int): int {
  Bits.and(n, Bits.not(align - 1))
}
alignDown(1000, 256) // 768
alignDown(4096, 256) // 4096
```

Tell a power of two from other positive numbers: it has one bit set, which `n - 1` clears:

```ant
func isPowerOfTwo(n: int): bool {
  n > 0 && Bits.and(n, n - 1) == 0
}
[1, 6, 64, 100].map(isPowerOfTwo) // [true, false, true, false]
```

## Properties and relations

The order and grouping of the operands do not matter:

```ant
let a = 0b1101
let b = 0b0111
let c = 0b1011
assert(Bits.and(a, b) == Bits.and(b, a))
assert(Bits.and(Bits.and(a, b), c) == Bits.and(a, Bits.and(b, c)))
```

ANDing a value with itself or with `-1` gives the value back, and with `0` gives `0`:

```ant
for x in [-5, 0, 7, 300] {
  assert(Bits.and(x, x) == x)
  assert(Bits.and(x, -1) == x)
  assert(Bits.and(x, 0) == 0)
}
```

The result has no more set bits than either operand:

```ant
let a = 0b1101
let b = 0b0111
let both = Bits.popCount(Bits.and(a, b))
assert(both <= Bits.popCount(a) && both <= Bits.popCount(b))
```

It is `Bits.or` with every bit inverted on both sides:

```ant
let a = 0b1101
let b = 0b0111
assert(Bits.and(a, b) == Bits.not(Bits.or(Bits.not(a), Bits.not(b))))
```

## Possible issues

Ant has no `&` operator; bitwise AND is this function:

```ant
let low = 0x1234 & 0xFF // error: unexpected character '&'
```

An integer is not reached through `.`, since numbers have no namespace. Call `Bits.and` or pipe into
it:

```ant
let flags = 5
let low = flags.and(1) // error: int has no field or function and
```

Both arguments must have the same type. An integer literal is an `int`, so convert it:

```ant
let b = uint8(0xF0)
let low = Bits.and(b, 0x0F) // error: cannot use int where byte is expected
```

```ant
let b = uint8(0xF0)
Bits.and(b, uint8(0x0F)) // 0
```

A float is not an integer:

```ant
let x = Bits.and(6.0, 3.0) // error: T would be float64, which is not a member of its bound Integer
```

A named primitive type is not a member of `Integer`, so AND its payload and tag the result:

```ant
type Perm = int
let p = Perm(0o755)
let owner = Bits.and(p, Perm(0o700))
// error: T would be Perm, which is not a member of its bound Integer; unwrap it with int(...)
```

```ant
type Perm = int
let p = Perm(0o755)
Perm(Bits.and(int(p), 0o700)) // Perm(448)
```

## See also

- [Bits.or](or.md), [Bits.xor](xor.md), and [Bits.not](not.md): the other bitwise operations.
- [Bits.shiftRight](shiftRight.md): move the bits before masking them.
- [Bits.popCount](popCount.md): count the bits a mask keeps.
