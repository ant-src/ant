<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.wrapMul

```ant signature
func Bits.wrapMul[T: Integer](a: T, b: T): T
```

Bits.wrapMul returns a \* b, wrapping around in two's complement instead of panicking on overflow.

<!-- end generated -->

## Details

- The result is the low bits of the true product, so it never panics.
- On a signed kind, a wrapped product can change sign, even when both arguments are positive.

## Basic examples

```ant
Bits.wrapMul(int8(64), int8(2))    // -128
Bits.wrapMul(uint8(16), uint8(16)) // 0
Bits.wrapMul(uint8(200), uint8(3)) // 88
Bits.wrapMul(int32(7), int32(6))   // 42
```

## Scope

On `int`, the product wraps at 64 bits:

```ant
Bits.wrapMul(Math.maxInt, 2)  // -2
Bits.wrapMul(Math.minInt, -1) // -9223372036854775808
```

Every integer kind works, signed or unsigned:

```ant
Bits.wrapMul(int16(-300), int16(300))        // -24464
Bits.wrapMul(int32(100000), int32(100000))   // 1410065408
Bits.wrapMul(Math.maxUint64, Math.maxUint64) // 1
Bits.wrapMul(byte(255), byte(255))           // 1
```

In a pipe, the value on the left is the first argument:

```ant
uint8(20) | Bits.wrapMul(uint8(13)) // 4
```

## Applications

### A string hash like Java's

Java's `String.hashCode` computes `31 * h + c` in 32-bit signed arithmetic. `Bits.wrapMul` and
[Bits.wrapAdd](wrapAdd.md) give the same values for ASCII text:

```ant
func javaHash(s: string): int32 {
  var h = int32(0)
  for b in s.toBytes() {
    h = h | Bits.wrapMul(int32(31)) | Bits.wrapAdd(int32(b))
  }
  h
}

println(javaHash("hello"))
println(javaHash("polygenelubricants"))
```

```text
99162322
-2147483648
```

### A linear congruential generator

A 64-bit LCG multiplies its state modulo 2^64 on every step:

```ant
type Lcg = { state: uint64 }

func Lcg.next(g: Lcg): uint64 {
  g.state = g.state
    | Bits.wrapMul(uint64(6364136223846793005))
    | Bits.wrapAdd(uint64(1442695040888963407))
  g.state
}

let g = Lcg{ state: uint64(1) }
println(g.next())
println(g.next())
```

```text
7806831264735756412
9396908728118811419
```

### Multiplicative hashing into buckets

Multiplying by a large odd constant mixes the bits of a key, and the top bits of the product pick a
bucket:

```ant
func bucket(key: uint32, bits: int): int {
  let mixed = Bits.wrapMul(key, uint32(2654435769))
  int(Bits.shiftRight(mixed, 32 - bits))
}

println([1, 2, 3, 4, 5].map({ bucket(uint32(x), 4) }))
```

```text
[9, 3, 13, 7, 1]
```

## Properties and relations

When the product fits the kind, `Bits.wrapMul` equals `*`:

```ant
let a = int8(10)
let b = int8(-12)
assert(Bits.wrapMul(a, b) == a * b)
```

On `byte`, the result is the true product modulo 256:

```ant
for a in Array.range(0, 256) {
  for b in [0, 1, 3, 128, 255] {
    assert(int(Bits.wrapMul(byte(a), byte(b))) == a * b % 256)
  }
}
```

The order of the arguments does not matter:

```ant
let a = int16(-1234)
let b = int16(5678)
assert(Bits.wrapMul(a, b) == Bits.wrapMul(b, a))
```

Multiplying by a power of two is a left shift, which also drops the high bits:

```ant
for x in [int8(1), int8(100), int8(-77)] {
  assert(Bits.wrapMul(x, int8(8)) == Bits.shiftLeft(x, 3))
}
```

Multiplying by `-1` is the wrapping negation, [Bits.wrapSub](wrapSub.md) from zero:

```ant
for x in [0, 5, Math.maxInt, Math.minInt] {
  assert(Bits.wrapMul(x, -1) == Bits.wrapSub(0, x))
}
```

## Possible issues

The `*` operator panics where `Bits.wrapMul` wraps:

```ant
int8(64) * int8(2)              // panic: integer overflow
Bits.wrapMul(int8(64), int8(2)) // -128
```

A literal is an `int`, so next to another kind it is a compile error. Convert it:

```ant
let x = int8(64)
let doubled = Bits.wrapMul(x, 2) // error: cannot use int where int8 is expected
```

`Bits.wrapMul` belongs to `Bits`, not to the integer kinds, so it has no dot form:

```ant
let x = int8(64)
let doubled = x.wrapMul(int8(2)) // error: int8 has no field or function wrapMul
```

A named primitive type is not an `Integer`. Unwrap it first, and tag the result again:

```ant
type Hash = uint32
let h = Hash(uint32(7))
let bad = Bits.wrapMul(h, h)
// error: T would be Hash, which is not a member of its bound Integer; unwrap it with uint32(...)
```

```ant
type Hash = uint32
let h = Hash(uint32(7))
Hash(Bits.wrapMul(uint32(h), uint32(h))) // Hash(49)
```

The wrapped product is not the true product, and nothing reports that it wrapped. When an overflow
is a bug, use `*` and let it panic.

## Neat examples

Every odd number has an inverse modulo 2^n, so a wrapping multiplication by an odd number can be
undone. On `byte`, `3 * 171` is `1`:

```ant
Bits.wrapMul(byte(3), byte(171))                         // 1
Bits.wrapMul(Bits.wrapMul(byte(77), byte(3)), byte(171)) // 77
```

## See also

[Bits.wrapAdd](wrapAdd.md) · [Bits.wrapSub](wrapSub.md) · [Bits.shiftLeft](shiftLeft.md) ·
[Bits.xor](xor.md) · [Bits](index.md) · [int](../../language/types/int.md) ·
[byte](../../language/types/byte.md)
