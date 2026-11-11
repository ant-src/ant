<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.trailingZeros

```ant signature
func Bits.trailingZeros[T: Integer](a: T): int
```

Bits.trailingZeros returns the number of zero bits below the lowest one bit of a, over its two's
complement bits. It is the bit width of T for 0.

<!-- end generated -->

## Details

- Every value other than `0` gives less than the width.
- The sign does not matter: a value and its negation have the same lowest one bit.

## Basic examples

Count the zero bits at the low end of a number:

```ant
Bits.trailingZeros(8)  // 3
Bits.trailingZeros(12) // 2
Bits.trailingZeros(1)  // 0
```

An odd number ends in a one bit, so it gives `0`:

```ant
Bits.trailingZeros(7)   // 0
Bits.trailingZeros(101) // 0
```

## Scope

`0` has no one bit, so the result is the width of its type:

```ant
Bits.trailingZeros(0)         // 64
Bits.trailingZeros(uint8(0))  // 8
Bits.trailingZeros(int16(0))  // 16
Bits.trailingZeros(uint32(0)) // 32
```

A negative number counts over its two's complement bits, which end like those of its absolute value:

```ant
Bits.trailingZeros(-8)          // 3
Bits.trailingZeros(-1)          // 0
Bits.trailingZeros(int8(-128))  // 7
Bits.trailingZeros(Math.minInt) // 63
```

Only the lowest one bit counts. The highest bit of a width gives one less than the width, and a
value with its lowest bit set gives `0` however many bits it has:

```ant
Bits.trailingZeros(byte(0x80))                    // 7
Bits.trailingZeros(Bits.shiftLeft(uint64(1), 63)) // 63
Bits.trailingZeros(Math.maxUint64)                // 0
Bits.trailingZeros(0x1_0000_0000)                 // 32
```

Pass it by name to `Array.map`, call it in a pipe, or name its parameter `a`:

```ant
[1, 2, 4, 6, 0].map(Bits.trailingZeros) // [0, 1, 2, 1, 64]
40 | Bits.trailingZeros()               // 3
Bits.trailingZeros(a: 40)               // 3
```

## Applications

Find the exponent of a power of two, such as the shift for a buffer size:

```ant
Bits.trailingZeros(1024) // 10
Bits.trailingZeros(4096) // 12
```

Split a number into a power of two and an odd part:

```ant
let n = 96
let k = Bits.trailingZeros(n)
k                     // 5
Bits.shiftRight(n, k) // 3
Bits.shiftLeft(1, k)  // 32
```

List the positions of the set bits of a mask, lowest first. `Bits.and(rest, rest - 1)` clears the
lowest one bit on each step:

```ant
func setBits(mask: int): Array[int] {
  var rest = mask
  let out: Array[int] = []
  while rest != 0 {
    out.push(Bits.trailingZeros(rest))
    rest = Bits.and(rest, rest - 1)
  }
  out
}

println(setBits(0b1010_0110))
```

```text
[1, 2, 5, 7]
```

Compute a greatest common divisor without division, as the binary GCD algorithm does. The common
factor of two is the trailing zeros of `Bits.or(a, b)`:

```ant
func gcd(a: int, b: int): int {
  if a == 0 { return b }
  if b == 0 { return a }
  let shift = Bits.trailingZeros(Bits.or(a, b))
  var x = Bits.shiftRight(a, Bits.trailingZeros(a))
  var y = b
  while y != 0 {
    y = Bits.shiftRight(y, Bits.trailingZeros(y))
    if x > y {
      let t = x
      x = y
      y = t
    }
    y -= x
  }
  Bits.shiftLeft(x, shift)
}

println(gcd(48, 180))
println(gcd(0, 7))
```

```text
12
7
```

## Properties and relations

The trailing zeros are the one bits of `Bits.and(Bits.not(x), x - 1)`, for every value, `0`
included:

```ant
for x in [1, 6, 40, 96, -12, 0, Math.maxInt] {
  assert(Bits.trailingZeros(x) == Bits.popCount(Bits.and(Bits.not(x), Bits.wrapSub(x, 1))))
}
```

A nonzero `x` divides by `2` to the power of its trailing zeros, and the quotient is odd:

```ant
for x in [2, 12, 40, -96, Bits.shiftLeft(3, 40)] {
  let k = Bits.trailingZeros(x)
  let odd = Bits.shiftRight(x, k)
  assert(Bits.shiftLeft(odd, k) == x)
  assert(odd % 2 != 0)
}
```

For a power of two, `Bits.trailingZeros` and `Bits.leadingZeros` add up to one less than the width:

```ant
for k in Iter.range(0, 63) {
  let p = Bits.shiftLeft(1, k)
  assert(Bits.trailingZeros(p) == k)
  assert(Bits.trailingZeros(p) + Bits.leadingZeros(p) == 63)
}
```

## Possible issues

`0` gives the width, not `0`, and the width depends on the type. Guard for `0` when the count feeds
a shift:

```ant
Bits.trailingZeros(0)        // 64
Bits.trailingZeros(uint8(0)) // 8
Bits.shiftRight(0, Bits.trailingZeros(0))
// panic: Bits.shiftRight needs a count from 0 up to but excluding 64, got 64
```

Choose the result for `0` before the shift with an `if` expression:

```ant
let n = 0
let odd = if n == 0 { 0 } else { Bits.shiftRight(n, Bits.trailingZeros(n)) }
odd // 0
```

A dot call looks in the namespace of the receiver's type, and `Bits` is not the namespace of `int`.
Call it by its namespace:

```ant
let n = 48
n.trailingZeros() // error: int has no field or function trailingZeros
```

A named primitive type is not an `Integer`; unwrap it first:

```ant
type Flags = uint16

let f = Flags(uint16(8))
let count = Bits.trailingZeros(f)
// error: T would be Flags, which is not a member of its bound Integer; unwrap it with uint16(...)
let unwrapped = Bits.trailingZeros(uint16(f))
```

## See also

[Bits.leadingZeros](leadingZeros.md) · [Bits.popCount](popCount.md) ·
[Bits.shiftRight](shiftRight.md) · [Bits.and](and.md) · [Math.log2](../Math/log2.md) ·
[Bits](index.md)
