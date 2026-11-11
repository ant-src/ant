<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# uint16

```ant signature
func uint16[T: Number](x: T): uint16
```

uint16 converts x to uint16 exactly. It panics if x is out of range, has a fraction, or is NaN or
infinite; uint16(Math.trunc(x)) truncates. x may also be a named primitive over a number.

<!-- end generated -->

## Details

- `uint16` holds the whole numbers from `0` to `65535`.
- The conversion never rounds or wraps.
- Arithmetic on two `uint16` values gives a `uint16`, and panics when the result leaves the range,
  below zero included.

## Basic examples

Convert an `int` or a whole `float64`:

```ant
uint16(8080)    // 8080
uint16(65535.0) // 65535
uint16(0)       // 0
```

Two `uint16` values add to a `uint16`:

```ant
let a = uint16(60000)
let b = uint16(5535)
a + b // 65535
```

## Scope

Every numeric type converts, the sized ones included:

```ant
uint16(byte(200))     // 200
uint16(int64(65535))  // 65535
uint16(float32(12.0)) // 12
uint16(int8(127))     // 127
uint16(uint16(5))     // 5
```

`uint16` is a function value, so it converts every element of an array:

```ant
Array.map([80, 443, 8080], uint16) // [80, 443, 8080]
```

It unwraps a named primitive type over a number:

```ant
type Port = uint16
let https = Port(uint16(443))
https         // Port(443)
uint16(https) // 443
```

In a pipe, the value is the argument:

```ant
(String.toInt("3000") ?? 0) | uint16() // 3000
```

## Applications

Split a `uint16` into its two bytes, high byte first:

```ant
func toBytes(n: uint16): Bytes {
  let hi = byte(Bits.shiftRight(n, 8))
  let lo = byte(Bits.and(n, uint16(0xff)))
  Bytes.fromArray([hi, lo])
}

toBytes(uint16(0x1234)) // Bytes(12 34)
```

Pack a color into 16 bits as RGB565: five bits of red, six of green, five of blue:

```ant
func rgb565(r: int, g: int, b: int): uint16 {
  let r5 = Bits.shiftLeft(uint16(r / 8), 11)
  let g6 = Bits.shiftLeft(uint16(g / 4), 5)
  Bits.or(Bits.or(r5, g6), uint16(b / 8))
}

rgb565(255, 0, 0)     // 63488
rgb565(255, 255, 255) // 65535
```

Check a port number read as text before converting it, since `uint16` panics on a value out of
range:

```ant
func parsePort(s: string): uint16 | null {
  let n = String.toInt(s) ?? -1
  if n < 1 || n > 65535 {
    return null
  }
  uint16(n)
}

parsePort("8080")  // 8080
parsePort("70000") // null
parsePort("http")  // null
```

## Properties and relations

Every `int` in range survives the round trip through `uint16`:

```ant
for n in [0, 1, 255, 256, 65535] {
  assert(int(uint16(n)) == n)
}
```

`Bits.wrapAdd` and `Bits.wrapSub` wrap around where `+` and `-` panic:

```ant
Bits.wrapAdd(uint16(65535), uint16(1)) // 0
Bits.wrapSub(uint16(0), uint16(1))     // 65535
```

## Possible issues

A value outside the range panics; it does not wrap:

```ant
uint16(65536) // panic: 65536 out of range for uint16
uint16(-1)    // panic: -1 out of range for uint16
```

A fraction panics instead of truncating:

```ant
uint16(2.5) // panic: 2.5 has a fraction, so it does not convert to uint16
```

Round explicitly first; `Math.trunc` rounds toward zero:

```ant
uint16(Math.trunc(2.7)) // 2
uint16(Math.round(2.5)) // 3
```

`NaN` and the infinities panic:

```ant
uint16(Math.NaN)       // panic: NaN does not convert to uint16
uint16(-Math.Infinity) // panic: -Infinity does not convert to uint16
```

An integer literal is an `int`, not a `uint16`:

```ant
let port: uint16 = 80 // error: cannot use int where uint16 is expected
```

Convert the literal:

```ant
let port: uint16 = uint16(80)
port // 80
```

A `uint16` and an `int` do not mix in arithmetic or comparisons:

```ant
let a = uint16(1)
println(a < 2) // error: < needs both operands of one type, got uint16 and int
```

Convert the other operand:

```ant
let a = uint16(1)
a < uint16(2) // true
```

Subtraction below zero panics, as does any result above `65535`:

```ant
uint16(0) - uint16(1)         // panic: integer overflow
uint16(300) * uint16(300)     // panic: integer overflow
uint16(60000) + uint16(10000) // panic: integer overflow
```

Widen to `int` before a computation that may leave the range:

```ant
int(uint16(300)) * int(uint16(300)) // 90000
```

## See also

[uint8](uint8.md) · [uint32](uint32.md) · [uint64](uint64.md) · [int16](int16.md) · [int](int.md) ·
[Math.trunc](../Math/trunc.md) · [Math.round](../Math/round.md) · [Bits.wrapAdd](../Bits/wrapAdd.md)
· [Bits.wrapSub](../Bits/wrapSub.md) · [Bits.shiftLeft](../Bits/shiftLeft.md) ·
[Bits.shiftRight](../Bits/shiftRight.md) ·
[Arithmetic operators](../../language/operators/arithmetic.md)
