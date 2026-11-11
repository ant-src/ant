<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.wrapSub

```ant signature
func Bits.wrapSub[T: Integer](a: T, b: T): T
```

Bits.wrapSub returns a - b, wrapping around in two's complement instead of panicking on overflow.

<!-- end generated -->

## Details

- The result is `a - b` reduced to the width of `T`: below the type's minimum it wraps around to the
  top, and above its maximum to the bottom. It never panics.

## Basic examples

Subtract below zero in a `uint8` and wrap around to the top:

```ant
Bits.wrapSub(uint8(0), uint8(1)) // 255
```

Without overflow it is ordinary subtraction:

```ant
Bits.wrapSub(10, 3) // 7
```

## Scope

Every integer type works, with both arguments of that type. The width decides where it wraps:

```ant
Bits.wrapSub(uint16(3), uint16(5))   // 65534
Bits.wrapSub(uint32(10), uint32(20)) // 4294967286
Bits.wrapSub(uint64(0), uint64(1))   // 18446744073709551615
```

Signed types wrap from the minimum to the maximum, and back:

```ant
Bits.wrapSub(int8(-128), int8(1)) // 127
Bits.wrapSub(int8(127), int8(-1)) // -128
Bits.wrapSub(Math.minInt, 1)      // 9223372036854775807
Bits.wrapSub(Math.maxInt, -1)     // -9223372036854775808
```

Negating the minimum gives the minimum back, since its positive does not fit:

```ant
Bits.wrapSub(0, Math.minInt) // -9223372036854775808
```

Call it in a pipe, where the value on the left becomes `a`:

```ant
uint8(5) | Bits.wrapSub(uint8(10)) // 251
```

## Applications

Measure the time between two readings of a 32-bit tick counter, even after the counter wrapped
around between them:

```ant
let start = uint32(4294967290)
let now = uint32(4)
Bits.wrapSub(now, start) // 10
```

Compare sequence numbers that wrap: `a` comes after `b` when the distance from `b` forward to `a` is
less than half the range:

```ant
func isAfter(a: uint16, b: uint16): bool {
  let d = Bits.wrapSub(a, b)
  d != uint16(0) && d < uint16(32768)
}
isAfter(uint16(2), uint16(65530)) // true
isAfter(uint16(65530), uint16(2)) // false
isAfter(uint16(7), uint16(7))     // false
```

Compute a checksum byte that makes the bytes of a record sum to zero, as Intel HEX does:

```ant
func checksum(data: Array[byte]): byte {
  let sum = data.reduce(byte(0), Bits.wrapAdd)
  Bits.wrapSub(byte(0), sum)
}
let record = [0x02, 0x00, 0x00, 0x04, 0xFF, 0xFF].map(byte)
checksum(record) // 252
```

## Properties and relations

Adding `b` back with `Bits.wrapAdd` undoes the subtraction:

```ant
for a in [uint8(0), uint8(7), uint8(255)] {
  for b in [uint8(0), uint8(9), uint8(200)] {
    assert(Bits.wrapAdd(Bits.wrapSub(a, b), b) == a)
  }
}
```

Subtracting `b` is adding its wrapped negation, `Bits.wrapSub(0, b)`:

```ant
let a = int8(100)
let b = int8(-100)
assert(Bits.wrapSub(a, b) == Bits.wrapAdd(a, Bits.wrapSub(int8(0), b)))
```

When the difference fits in the type, it equals the `-` operator:

```ant
for a in [-50, 0, 1000] {
  assert(Bits.wrapSub(a, 7) == a - 7)
  assert(Bits.wrapSub(a, a) == 0)
}
```

Negating with `Bits.wrapSub(0, x)` is inverting the bits and adding one, the two's complement:

```ant
for x in [int8(-128), int8(-1), int8(0), int8(42)] {
  assert(Bits.wrapSub(int8(0), x) == Bits.wrapAdd(Bits.not(x), int8(1)))
}
```

## Possible issues

The `-` operator panics on overflow, including an unsigned result below zero. Use `Bits.wrapSub`
when wrapping is what you want:

```ant
uint8(0) - uint8(1) // panic: integer overflow
```

On an unsigned type the result is never negative. Converting it to the signed type of the same width
panics when it is out of range:

```ant
int8(Bits.wrapSub(uint8(3), uint8(5))) // panic: 254 out of range for int8
```

Both arguments must have the same type. An integer literal is an `int`, so convert it:

```ant
let b = uint8(3)
let d = Bits.wrapSub(b, 5) // error: cannot use int where byte is expected
```

A number has no namespace, so a dot call on it finds no function. Call `Bits.wrapSub(n, 1)` or pipe
into it:

```ant
let n = 5
let d = n.wrapSub(1) // error: int has no field or function wrapSub
```

A float is not an integer:

```ant
let d = Bits.wrapSub(1.5, 0.5)
// error: T would be float64, which is not a member of its bound Integer
```

A named primitive type is not a member of `Integer`, so subtract its payload and tag the result:

```ant
type Tick = uint32
let t = Tick(uint32(5))
let d = Bits.wrapSub(t, Tick(uint32(7)))
// error: T would be Tick, which is not a member of its bound Integer; unwrap it with uint32(...)
```

```ant
type Tick = uint32
let t = Tick(uint32(5))
Tick(Bits.wrapSub(uint32(t), uint32(7))) // Tick(4294967294)
```

## See also

- [Bits.wrapAdd](wrapAdd.md) and [Bits.wrapMul](wrapMul.md): the other wrapping operations.
- [Bits.not](not.md): inverts the bits, which with one added negates a value.
- [Bits](index.md): the other bitwise operations.
