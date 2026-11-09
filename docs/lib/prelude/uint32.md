<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# uint32

```ant signature
func uint32[T: Number](x: T): uint32
```

uint32 converts x to uint32 exactly. It panics if x is out of range, has a fraction, or is NaN or infinite; uint32(Math.trunc(x)) truncates. x may also be a named primitive over a number.

<!-- end generated -->

## Details

- Every `uint32` value lies between 0 and 4294967295.
- Integer literals are always `int`. Write `uint32(5)`, not `5`, wherever a `uint32` is expected.
- Arithmetic on two `uint32` values gives a `uint32` and panics when the result leaves the range,
  including subtraction below zero. Use [Bits.wrapAdd](../Bits/wrapAdd.md),
  [Bits.wrapSub](../Bits/wrapSub.md), and [Bits.wrapMul](../Bits/wrapMul.md) for wrapping
  arithmetic.

## Basic examples

```ant
uint32(42)         // 42
uint32(4294967295) // 4294967295
uint32(3.0)        // 3
```

The result is a `uint32`, and arithmetic between two of them stays a `uint32`. Integer `/`
truncates:

```ant
let a = uint32(3000000000)
a + uint32(1000000000) // 4000000000
uint32(7) / uint32(2)  // 3
```

## Scope

Any integer kind converts, as long as the value is not negative and fits:

```ant
uint32(byte(200))       // 200
uint32(int16(300))      // 300
uint32(uint64(1000000)) // 1000000
uint32(0x8000_0001)     // 2147483649
```

A float converts when it has no fraction. Negative zero gives `0`:

```ant
uint32(1e9)                 // 1000000000
uint32(float32(16777216.0)) // 16777216
uint32(-0.0)                // 0
```

A value of a named primitive type over a number is unwrapped first:

```ant
type Port = int
let port = Port(8080)
uint32(port) // 8080
```

`uint32` is a function value, so it can be passed to `Array.map`:

```ant
Array.map([1, 2, 3], uint32)        // [1, 2, 3]
[1, 2, 3].map(uint32) | Array.sum() // 6
```

## Applications

Compute the 32-bit FNV-1a hash of a string. The multiplication wraps on purpose:

```ant
func fnv1a(s: string): uint32 {
  var h = uint32(2166136261)
  for b in s.toBytes() {
    h = Bits.xor(h, uint32(b))
    h = Bits.wrapMul(h, uint32(16777619))
  }
  h
}

println(fnv1a(""))
println(fnv1a("a"))
println(fnv1a("foobar"))
```

```text
2166136261
3826002220
3214735720
```

Pack an IPv4 address into a `uint32`, and unpack it again:

```ant
func parseIPv4(s: string): uint32 | null {
  let parts = s.split(".")
  if len(parts) != 4 {
    return null
  }
  var ip = uint32(0)
  for p in parts {
    let n = p.toInt() ?? return null
    if n < 0 || n > 255 {
      return null
    }
    ip = Bits.or(Bits.shiftLeft(ip, 8), uint32(n))
  }
  ip
}

func formatIPv4(ip: uint32): string {
  let octets = [24, 16, 8, 0].map({ String.from(Bits.and(Bits.shiftRight(ip, x), uint32(255))) })
  String.join(octets, ".")
}

let ip = parseIPv4("192.168.1.10") ?? panic("bad address")
println(ip)
println(formatIPv4(ip))
println(parseIPv4("10.0.0.256"))
```

```text
3232235786
192.168.1.10
null
```

Test whether two addresses share a `/24` network with a mask:

```ant
let mask = Bits.shiftLeft(Bits.not(uint32(0)), 8)
let a = uint32(3232235786)
let b = uint32(3232235777)
Bits.and(a, mask) == Bits.and(b, mask) // true
```

## Properties and relations

Every `uint32` survives a round trip through `int` and through `float64`:

```ant
for n in [uint32(0), uint32(2147483648), uint32(4294967295)] {
  assert(uint32(int(n)) == n)
  assert(uint32(float64(n)) == n)
}
```

Wrapping arithmetic works modulo 2^32, so subtracting one from zero gives the largest value:

```ant
assert(Bits.wrapSub(uint32(0), uint32(1)) == uint32(4294967295))
assert(Bits.not(uint32(0)) == uint32(4294967295))
assert(Bits.leadingZeros(uint32(1)) == 31)
```

## Possible issues

An integer literal is an `int`, never a `uint32`, so it does not fit where a `uint32` is expected:

```ant
let n: uint32 = 5     // error: cannot use int where uint32 is expected
let m = uint32(5) - 1 // error: - needs both operands of one type, got uint32 and int
let k = uint32(5) > 3 // error: > needs both operands of one type, got uint32 and int
```

A negative number panics, even a fixed-size one such as `int8(-1)`, and so does a value above
4294967295:

```ant
uint32(-1)         // panic: -1 out of range for uint32
uint32(int8(-1))   // panic: -1 out of range for uint32
uint32(4294967296) // panic: 4294967296 out of range for uint32
```

Subtraction below zero panics, and so does addition past the top:

```ant
uint32(3) - uint32(5)                  // panic: integer overflow
uint32(4000000000) + uint32(400000000) // panic: integer overflow
Bits.wrapSub(uint32(3), uint32(5))     // 4294967294
```

A float with a fraction panics instead of rounding or truncating, and `NaN` and the infinities panic
too:

```ant
uint32(1.5)             // panic: 1.5 has a fraction, so it does not convert to uint32
uint32(-0.5)            // panic: -0.5 has a fraction, so it does not convert to uint32
uint32(Math.trunc(1.5)) // 1
uint32(Math.NaN)        // panic: NaN does not convert to uint32
uint32(-Math.Infinity)  // panic: -Infinity does not convert to uint32
```

A string does not convert:

```ant
let n = uint32("42") // error: uint32 converts a number or a named primitive type over a number, got string
```

Parse it first, then convert the number:

```ant
uint32("42".toInt() ?? panic("bad")) // 42
```

A shift count must be below 32, the width of `uint32`:

```ant
let top = Bits.shiftLeft(uint32(1), 31)
let gone = Bits.shiftLeft(uint32(1), 32)
// panic: Bits.shiftLeft needs a count from 0 up to but excluding 32, got 32
```

## See also

[uint8](uint8.md) · [uint16](uint16.md) · [uint64](uint64.md) · [int32](int32.md) · [int](int.md) ·
[float64](float64.md) · [Math.trunc](../Math/trunc.md) · [Math.toInt](../Math/toInt.md) ·
[String.toInt](../String/toInt.md) · [Bits.wrapAdd](../Bits/wrapAdd.md) ·
[Bits.wrapSub](../Bits/wrapSub.md) · [Bits.wrapMul](../Bits/wrapMul.md) ·
[Bits.shiftLeft](../Bits/shiftLeft.md) ·
[Arithmetic operators](../../language/operators/arithmetic.md)
