<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.append

```ant signature
func Bytes.append[T: Number](b: Bytes, t: Reflect.Type[T], x: T, order: Bytes.Order): void
```

Bytes.append appends x to b in the given order.

<!-- end generated -->

## Details

- `Bytes.append` changes `b`: it adds [Bytes.sizeOf(t)](sizeOf.md) bytes at the end, and every name
  for the collection sees them.
- `x` must already have the type that `t` names. An integer literal is an `int`, so convert it
  first, as in `uint16(80)`.
- Integers are written in two's complement, and floats as their IEEE 754 bits.
- A one-byte type, `byte`, `uint8` or `int8`, takes an order and ignores it.

## Basic examples

Append a 16-bit number with the most significant byte first:

```ant
let b = Bytes.new()
Bytes.append(b, type(uint16), uint16(0x1234), Bytes.BigEndian{})
b // Bytes(12 34)
```

The same number with the least significant byte first:

```ant
let b = Bytes.new()
Bytes.append(b, type(uint16), uint16(0x1234), Bytes.LittleEndian{})
b // Bytes(34 12)
```

Each call adds at the end, after the bytes already there:

```ant
let b = "ID".toBytes()
Bytes.append(b, type(uint32), uint32(7), Bytes.BigEndian{})
b      // Bytes(49 44 00 00 00 07)
len(b) // 6
```

## Scope

The type value fixes the width: 1, 2, 4 or 8 bytes:

```ant
let b = Bytes.new()
b.append(type(byte), byte(0xff), Bytes.BigEndian{})
b.append(type(int16), int16(1), Bytes.BigEndian{})
b.append(type(uint32), uint32(1), Bytes.BigEndian{})
b.append(type(int64), int64(1), Bytes.BigEndian{})
b      // Bytes(ff 00 01 00 00 00 01 00 00 00 00 00 00 00 01)
len(b) // 15
```

`int` takes eight bytes:

```ant
let b = Bytes.new()
b.append(type(int), 1, Bytes.LittleEndian{})
b // Bytes(01 00 00 00 00 00 00 00)
```

Negative integers are written in two's complement:

```ant
let b = Bytes.new()
b.append(type(int8), int8(-1), Bytes.BigEndian{})
b.append(type(int32), int32(-2), Bytes.BigEndian{})
b // Bytes(ff ff ff ff fe)
```

Floats are written as their IEEE 754 bits, in four or eight bytes:

```ant
let b = Bytes.new()
b.append(type(float64), 1.5, Bytes.BigEndian{})
b.append(type(float32), float32(1.5), Bytes.LittleEndian{})
b // Bytes(3f f8 00 00 00 00 00 00 00 00 c0 3f)
```

A one-byte type gives the same byte in either order:

```ant
let b = Bytes.new()
b.append(type(byte), byte(5), Bytes.BigEndian{})
b.append(type(byte), byte(5), Bytes.LittleEndian{})
b // Bytes(05 05)
```

Keep the order in a name and pass it to every call:

```ant
let order = Bytes.LittleEndian{}
let b = Bytes.new()
b.append(type(uint16), uint16(1), order)
b.append(type(uint16), uint16(2), order)
b // Bytes(01 00 02 00)
```

Call it through `.`, in a pipe, or with named arguments:

```ant
let b = Bytes.new()
b.append(type(byte), byte(1), Bytes.BigEndian{})
b | Bytes.append(type(byte), byte(2), Bytes.BigEndian{})
Bytes.append(b, t: type(uint16), x: uint16(3), order: Bytes.BigEndian{})
b // Bytes(01 02 00 03)
```

Another name for the same collection sees the new bytes. A copy from [Bytes.clone](clone.md) does
not:

```ant
let b = Bytes.fromArray([1])
let same = b
let copy = b.clone()
b.append(type(byte), byte(2), Bytes.BigEndian{})
same // Bytes(01 02)
copy // Bytes(01)
```

## Applications

Prefix a payload with its length as a big-endian `uint32`. `len` gives an `int`, so convert it:

```ant
func frame(payload: Bytes): Bytes {
  let out = Bytes.new()
  out.append(type(uint32), uint32(len(payload)), Bytes.BigEndian{})
  out.extend(payload)
  out
}

frame("hello".toBytes()) // Bytes(00 00 00 05 68 65 6c 6c 6f)
```

Write a small file header: a magic number, a version, and a count of records, all little-endian:

```ant
type Header = { version: uint16, count: uint32 }

func encode(h: Header): Bytes {
  let order = Bytes.LittleEndian{}
  let out = "ANT1".toBytes()
  out.append(type(uint16), h.version, order)
  out.append(type(uint32), h.count, order)
  out
}

let bytes = encode(Header{ version: uint16(2), count: uint32(300) })
println(Encoding.toHex(bytes))
```

```text
414e543102002c010000
```

Pack points as pairs of `float32` and read them back with [Bytes.read](read.md), moving by
[Bytes.sizeOf](sizeOf.md):

```ant
type Point = { x: float32, y: float32 }

let order = Bytes.LittleEndian{}
let points = [Point{ x: float32(1.5), y: float32(-2) }, Point{ x: float32(0.25), y: float32(4) }]
let packed = Bytes.new()
for p in points {
  packed.append(type(float32), p.x, order)
  packed.append(type(float32), p.y, order)
}
let size = Bytes.sizeOf(type(float32))
var at = 0
while at < len(packed) {
  let x = packed.read(type(float32), at, order)
  let y = packed.read(type(float32), at + size, order)
  println("${x}, ${y}")
  at += 2 * size
}
```

```text
1.5, -2.0
0.25, 4.0
```

## Properties and relations

The length grows by [Bytes.sizeOf(t)](sizeOf.md):

```ant
let b = "abc".toBytes()
let before = len(b)
b.append(type(float64), 3.25, Bytes.BigEndian{})
assert(len(b) == before + Bytes.sizeOf(type(float64)))
```

[Bytes.read](read.md) at the old length, in the same order, gives the number back:

```ant
let order = Bytes.BigEndian{}
let b = "hdr".toBytes()
let at = len(b)
b.append(type(int32), int32(-40000), order)
assert(b.read(type(int32), at, order) == int32(-40000))
```

Little-endian bytes are the big-endian bytes reversed:

```ant
let big = Bytes.new()
let little = Bytes.new()
big.append(type(uint32), uint32(0x01020304), Bytes.BigEndian{})
little.append(type(uint32), uint32(0x01020304), Bytes.LittleEndian{})
assert(big.toArray().reverse() == little.toArray())
```

Appending is the same as making room with [Bytes.new](new.md) and [Bytes.write](write.md):

```ant
let appended = Bytes.new()
appended.append(type(uint16), uint16(513), Bytes.BigEndian{})
let written = Bytes.new(2)
written.write(type(uint16), 0, uint16(513), Bytes.BigEndian{})
assert(appended == written)
```

For one byte, [Bytes.push](push.md) gives the same result without a type value or an order:

```ant
let a = Bytes.new()
a.append(type(byte), byte(7), Bytes.BigEndian{})
let p = Bytes.new()
p.push(byte(7))
assert(a == p)
```

## Possible issues

An integer literal is an `int`, never a narrower type, so it does not match a `uint16` type value.
Convert it:

```ant
let b = Bytes.new()
b.append(type(uint16), 80, Bytes.BigEndian{}) // error: cannot use int where uint16 is expected
```

The same holds for `len` and other `int` results. A conversion that loses the value panics, so a
length above 65535 does not fit a `uint16`:

```ant
let b = Bytes.new()
let n = 70000
b.append(type(uint16), uint16(n), Bytes.BigEndian{}) // panic: 70000 out of range for uint16
```

The type value names one number type. A union is a compile error:

```ant
let b = Bytes.new()
b.append(type(uint16 | uint32), uint16(1), Bytes.BigEndian{})
// error: T would be uint16 | uint32, which is not a single member of its bound Number; narrow first
```

The order is a value, so it takes `{}`:

```ant
let b = Bytes.new()
b.append(type(byte), byte(1), Bytes.BigEndian)
// error: Bytes.BigEndian is a type, not a value; a type appears only as a literal head, after is, or before a dot
```

There is no default order. Every call states one:

```ant
let b = Bytes.new()
b.append(type(uint16), uint16(1)) // error: Bytes.append needs an argument for order
```

[Bytes.write](write.md) never grows the collection, so it cannot stand in for `Bytes.append` on
empty `Bytes`:

```ant
let b = Bytes.new()
b.write(type(uint32), 0, uint32(1), Bytes.BigEndian{})
// panic: Bytes.write: the range 0..4 is not within 0 ..= 0
```

## See also

[Bytes.read](read.md), [Bytes.write](write.md), [Bytes.sizeOf](sizeOf.md),
[Bytes.Order](types/Order.md), [Bytes.push](push.md), [Bytes.extend](extend.md), [Bytes.new](new.md)
