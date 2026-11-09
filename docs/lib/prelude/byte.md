<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# byte

```ant signature
func byte[T: Number](x: T): byte
```

byte converts x to byte, which is uint8, exactly. It panics if x is out of range, has a fraction, or
is NaN or infinite; int(Math.trunc(x)) truncates. x may also be a named primitive over a number.

<!-- end generated -->

## Details

- `byte` and `uint8` are one type under two names, and Ant shows it as `byte`.
- An integer literal is always an `int`, so a `byte` comes from a conversion such as `byte(65)` or
  from reading `Bytes`.
- Arithmetic on two `byte` values stays a `byte` and panics when the result leaves 0 to 255.

## Basic examples

```ant
byte(65)  // 65
byte(0)   // 0
byte(255) // 255
```

A float converts when it has no fraction:

```ant
byte(65.0) // 65
```

A value outside 0 to 255 panics:

```ant
byte(256) // panic: 256 out of range for byte
byte(-1)  // panic: -1 out of range for byte
```

## Scope

Every numeric kind converts, as long as the value fits:

```ant
byte(int8(100))     // 100
byte(uint64(7))     // 7
byte(float32(12.0)) // 12
byte(uint8(200))    // 200
```

A float with a fraction panics, and so do `NaN` and the infinities:

```ant
byte(65.5)          // panic: 65.5 has a fraction, so it does not convert to byte
byte(Math.NaN)      // panic: NaN out of range for byte
byte(Math.Infinity) // panic: Infinity out of range for byte
```

A named primitive over a number is unwrapped first:

```ant
type Level = int
byte(Level(3)) // 3
```

`byte` is a function value, so it converts every element of an array:

```ant
Array.map([72, 105], byte)     // [72, 105]
Array.map([72.0, 105.0], byte) // [72, 105]
```

In a generic function, a type parameter bounded by `Number` converts too:

```ant
func toByte[T: Number](x: T): byte {
  byte(x)
}
toByte(int16(7)) // 7
toByte(9.0)      // 9
```

Indexing `Bytes` gives a `byte`, and `int` converts it back:

```ant
let b = String.toBytes("Hi")
b[0]          // 72
int(b[0]) + 1 // 73
```

## Applications

Build `Bytes` from computed values:

```ant
let ramp = Array.map([0, 1, 2, 3], { byte(x * 85) })
Bytes.fromArray(ramp) // Bytes(00 55 aa ff)
```

Sum the bytes of a message into a one-byte checksum. The sum is kept in an `int`, which has room for
it, and only the remainder becomes a `byte`:

```ant
func checksum(data: Bytes): byte {
  var sum = 0
  for b in data {
    sum += int(b)
  }
  byte(sum % 256)
}

println(checksum(String.toBytes("hello")))
```

```text
20
```

Mask each byte with a one-byte key using `Bits.xor`, which keeps the `byte` kind:

```ant
let key = byte(0x2a)
let secret = Array.map(String.toBytes("ant").toArray(), { Bits.xor(x, key) })
println(Bytes.fromArray(secret))
let plain = Array.map(secret, { Bits.xor(x, key) })
println(String.fromBytes(Bytes.fromArray(plain)))
```

```text
Bytes(4b 44 5e)
ant
```

## Properties and relations

Every `int` from 0 to 255 survives the round trip through `byte`:

```ant
for i in Iter.range(0, 256) {
  assert(int(byte(i)) == i)
}
```

`byte` and `uint8` give equal values of one type:

```ant
assert(byte(42) == uint8(42))
```

## Possible issues

An integer literal is an `int`, and Ant converts nothing implicitly:

```ant
let flag: byte = 1 // error: cannot use int where byte is expected
```

Convert the literal:

```ant
let flag = byte(1)
flag // 1
```

A `byte` and an `int` do not mix in arithmetic or comparisons:

```ant
let b = String.toBytes("A")
println(b[0] + 1) // error: + needs both operands of one type, got byte and int
```

Arithmetic on `byte` values panics instead of wrapping around:

```ant
byte(250) + byte(10) // panic: integer overflow
byte(0) - byte(1)    // panic: integer overflow
```

For arithmetic that wraps around, use `Bits.wrapAdd`, `Bits.wrapSub`, or `Bits.wrapMul`:

```ant
Bits.wrapAdd(byte(250), byte(10)) // 4
Bits.wrapSub(byte(0), byte(1))    // 255
```

Or convert to `int` first, and take the remainder before converting back:

```ant
byte((int(byte(250)) + 10) % 256) // 4
```

A float with a fraction does not round. Truncate it first:

```ant
byte(Math.trunc(65.7)) // 65
```

`byte` takes a number, not a character. Get the bytes of a string with `String.toBytes`:

```ant
println(byte("A"))
// error: byte converts a number or a named primitive type over a number, got string
```

```ant
String.toBytes("A")[0] // 65
```

## See also

[uint8](uint8.md) · [int](int.md) · [Bytes](../Bytes/index.md) ·
[Bytes.fromArray](../Bytes/fromArray.md) · [Bytes.toArray](../Bytes/toArray.md) ·
[String.toBytes](../String/toBytes.md) · [Math.trunc](../Math/trunc.md) ·
[Bits.wrapAdd](../Bits/wrapAdd.md) · [byte type](../../language/types/byte.md)
