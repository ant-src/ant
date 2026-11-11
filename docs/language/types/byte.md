<!-- status: reviewed -->

# byte

`byte` is an unsigned 8-bit integer, from 0 to 255. It is another name for `uint8`. A `byte` is the
element of binary data: indexing `Bytes` or looping over it gives a `byte`.

## Details

- `byte` and `uint8` are one type under two names. Ant always shows it as `byte`.
- An integer literal is always an `int`. A `byte` comes from a conversion, `byte(x)`, which panics
  when `x` has a fraction or is outside 0 to 255.
- Arithmetic needs two operands of one type, so a `byte` never mixes with an `int` without a
  conversion. A result outside 0 to 255 panics, subtraction below zero included. `Bits.wrapAdd`,
  `Bits.wrapSub`, and `Bits.wrapMul` wrap instead.
- Bitwise operations are functions in the `Bits` namespace, not operators.

## Basic examples

Convert an `int` to a `byte`, and back:

```ant
let b = byte(65)
b      // 65
int(b) // 65
```

Read the bytes of binary data:

```ant
let data = String.toBytes("Hi")
data[0]        // 72
data.toArray() // [72, 105]
```

## Scope

### `uint8` is the same type

`uint8` names the same type, and it is shown as `byte`:

```ant
type(uint8)             // byte
byte(255) == uint8(255) // true
type(Array[uint8])      // Array[byte]
```

### Conversions

`byte` converts any numeric value that fits exactly, a float without a fraction included:

```ant
byte(0)       // 0
byte(255)     // 255
byte(3.0)     // 3
byte(int8(7)) // 7
```

As a function value, `byte` converts every element of an array:

```ant
Array.map([72, 105], byte) // [72, 105]
```

### Arithmetic and comparison

Two `byte` values add, subtract, and compare like any integers, and the result is a `byte`:

```ant
let a = byte(200)
let b = byte(50)
a + b // 250
a - b // 150
a / b // 4
a > b // true
```

Convert to `int` to compute past 255:

```ant
let a = byte(200)
let b = byte(100)
int(a) + int(b) // 300
```

### Bits

`Bits` functions on a `byte` give a `byte`, except the counts: `Bits.popCount`, `Bits.leadingZeros`,
and `Bits.trailingZeros` return an `int`, and a shift takes its count as an `int`:

```ant
let b = byte(0xa5)
Bits.shiftRight(b, 4)            // 10
Bits.and(b, byte(0x0f))          // 5
Bits.xor(b, byte(0xff))          // 90
Bits.not(b)                      // 90
Bits.popCount(b)                 // 4
Bits.wrapAdd(byte(255), byte(1)) // 0
```

### Bytes and loops

`for` over `Bytes` binds each byte as a `byte`:

```ant
var sum = 0
for b in String.toBytes("ant") {
  sum += int(b)
}
println(sum)
```

```text
323
```

An `Array[byte]` and `Bytes` convert to each other:

```ant
let raw = [byte(0x89), byte(0x50)]
Bytes.fromArray(raw)                  // Bytes(89 50)
Bytes.fromArray(raw).toArray() == raw // true
```

### Rendering

A `byte` renders as a decimal number. `String.fromInt` writes it in another base:

```ant
let b = byte(0xa5)
"${b}"                                             // "165"
String.fromInt(b, base: 16)                        // "a5"
String.fromInt(b, base: 2)                         // "10100101"
String.fromInt(byte(7), base: 16).padStart(2, "0") // "07"
```

## Applications

### A one-byte checksum

`Bits.wrapAdd` keeps the sum in a `byte`, wrapping past 255:

```ant
func checksum(data: Bytes): byte {
  var sum = byte(0)
  for b in data {
    sum = Bits.wrapAdd(sum, b)
  }
  sum
}

println(checksum(String.toBytes("ant")))
```

```text
67
```

### A hex dump

Each byte as two hexadecimal digits:

```ant
let hex = String.toBytes("Ant!").toArray()
  | Array.map({ String.fromInt(x, base: 16).padStart(2, "0") })
  | String.join(" ")
println(hex)
```

```text
41 6e 74 21
```

### High and low nibbles

```ant
func nibbles(b: byte): Array[byte] {
  [Bits.shiftRight(b, 4), Bits.and(b, byte(0x0f))]
}

println(nibbles(byte(0x3c)))
```

```text
[3, 12]
```

### XOR with a key

Compute bytes one at a time as an `Array[byte]`, then convert it to `Bytes`:

```ant
func xorWith(data: Bytes, key: byte): Bytes {
  data.toArray() | Array.map({ Bits.xor(x, key) }) | Bytes.fromArray()
}

let secret = xorWith(String.toBytes("ant"), byte(0x2a))
println(secret)
println(String.fromBytes(xorWith(secret, byte(0x2a))))
```

```text
Bytes(4b 44 5e)
ant
```

## Properties and relations

Every `int` from 0 to 255 converts to a `byte` and back unchanged:

```ant
for n in Array.range(0, 256) {
  assert(int(byte(n)) == n)
}
```

Wrapping arithmetic is arithmetic modulo 256:

```ant
let a = byte(200)
let b = byte(100)
assert(int(Bits.wrapAdd(a, b)) == (int(a) + int(b)) % 256)
assert(int(Bits.wrapSub(b, a)) == (int(b) - int(a) + 256) % 256)
```

## Possible issues

### An integer literal is an `int`

A literal does not become a `byte` from the type around it. Convert it:

```ant
let b: byte = 65 // error: cannot use int where byte is expected
```

```ant
let b: byte = byte(65)
let xs: Array[byte] = [1, 2].map(byte)
```

### Mixing `byte` and `int`

```ant
let b = byte(1)
let c = b + 1 // error: + needs both operands of one type, got byte and int
```

Convert one side: `b + byte(1)` stays a `byte`, `int(b) + 1` is an `int`.

### Overflow and underflow panic

```ant
byte(200) + byte(100) // panic: integer overflow
byte(0) - byte(1)     // panic: integer overflow
```

### A shift of 8 or more panics

A shift count must be below the operand's width, which is 8 for a `byte`:

```ant
Bits.shiftLeft(byte(1), 8) // panic: Bits.shiftLeft needs a count from 0 up to but excluding 8, got 8
```

To combine two bytes into a wider integer, convert first:

```ant
let hi = byte(0x12)
let lo = byte(0x34)
Bits.or(Bits.shiftLeft(int(hi), 8), int(lo)) // 4660
```

### Conversions out of range panic

```ant
byte(256) // panic: 256 out of range for byte
byte(-1)  // panic: -1 out of range for byte
byte(1.5) // panic: 1.5 has a fraction, so it does not convert to byte
```

### A string escape is not a byte

`\u{...}` writes a code point, which UTF-8 may store as more than one byte:

```ant
String.toBytes("\u{89}") // Bytes(c2 89)
Bytes.fromArray([0x89])  // Bytes(89)
```

### `Bytes` is immutable

A `byte` inside `Bytes` cannot be assigned:

```ant
let data = String.toBytes("Hi")
data[0] = byte(104)
// error: Bytes is immutable, so a byte cannot be assigned; build an Array[byte] and convert it with Bytes.fromArray
```

## See also

[int](int.md), [Bytes](../../lib/Bytes/index.md), [Bytes.fromArray](../../lib/Bytes/fromArray.md),
[Bytes.toArray](../../lib/Bytes/toArray.md), [Bits](../../lib/Bits/index.md),
[Bits.wrapAdd](../../lib/Bits/wrapAdd.md), [String.toBytes](../../lib/String/toBytes.md),
[String.fromInt](../../lib/String/fromInt.md), [Number literals](../literals/numbers.md),
[Arithmetic operators](../operators/arithmetic.md)
