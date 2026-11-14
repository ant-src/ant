<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.write

```ant signature
func Bytes.write[T: Number](b: Bytes, t: Reflect.Type[T], at: int, x: T, order: Bytes.Order): void
```

Bytes.write overwrites the bytes of b at index at with x in the given order. It never grows b, and
panics unless the number's bytes are within 0 ..= Bytes.length(b).

<!-- end generated -->

## Details

- `Bytes.write` overwrites [Bytes.sizeOf(t)](sizeOf.md) bytes of `b` in place, starting at `at`.
- `at` may be at most `b.length() - Bytes.sizeOf(t)`. Use [Bytes.append](append.md) to add a number
  at the end.
- `x` must already have the type that `t` names. An integer literal is an `int`, so convert it
  first, as in `uint16(80)`.
- Integers are written in two's complement, and floats as their IEEE 754 bits. A one-byte type takes
  an order and ignores it.

## Basic examples

Write a 16-bit number into two zero bytes, most significant byte first:

```ant
let b = Bytes.new(2)
Bytes.write(b, type(uint16), 0, uint16(0x1234), Bytes.BigEndian{})
b // Bytes(12 34)
```

The same number with the least significant byte first:

```ant
let b = Bytes.new(2)
Bytes.write(b, type(uint16), 0, uint16(0x1234), Bytes.LittleEndian{})
b // Bytes(34 12)
```

`at` picks where the number starts. The bytes around it stay as they were, and the length does not
change:

```ant
let b = Bytes.fromArray([1, 2, 3, 4, 5, 6])
b.write(type(uint16), 2, uint16(0xffff), Bytes.BigEndian{})
b          // Bytes(01 02 ff ff 05 06)
b.length() // 6
```

## Scope

The type value fixes how many bytes are overwritten: 1, 2, 4 or 8:

```ant
let b = Bytes.new(15)
b.write(type(byte), 0, byte(0xff), Bytes.BigEndian{})
b.write(type(int16), 1, int16(1), Bytes.BigEndian{})
b.write(type(uint32), 3, uint32(1), Bytes.BigEndian{})
b.write(type(int64), 7, int64(1), Bytes.BigEndian{})
b // Bytes(ff 00 01 00 00 00 01 00 00 00 00 00 00 00 01)
```

`int` takes eight bytes:

```ant
let b = Bytes.new(8)
b.write(type(int), 0, 256, Bytes.LittleEndian{})
b // Bytes(00 01 00 00 00 00 00 00)
```

Negative integers are written in two's complement:

```ant
let b = Bytes.new(5)
b.write(type(int8), 0, int8(-1), Bytes.BigEndian{})
b.write(type(int32), 1, int32(-2), Bytes.BigEndian{})
b // Bytes(ff ff ff ff fe)
```

Floats are written as their IEEE 754 bits. `-0.0` differs from `0.0` only in its sign bit:

```ant
let b = Bytes.new(12)
b.write(type(float64), 0, -0.0, Bytes.BigEndian{})
b.write(type(float32), 8, float32(1.5), Bytes.LittleEndian{})
b // Bytes(80 00 00 00 00 00 00 00 00 00 c0 3f)
```

A one-byte type gives the same byte in either order, the same as assigning by index:

```ant
let b = Bytes.new(3)
b.write(type(byte), 0, byte(7), Bytes.BigEndian{})
b.write(type(byte), 1, byte(7), Bytes.LittleEndian{})
b[2] = byte(7)
b // Bytes(07 07 07)
```

The last position a number fits at is the length minus its size:

```ant
let b = Bytes.new(6)
let at = b.length() - Bytes.sizeOf(type(uint32))
b.write(type(uint32), at, uint32(0xcafebabe), Bytes.BigEndian{})
b // Bytes(00 00 ca fe ba be)
```

The order is a value of type [Bytes.Order](types/Order.md), so a program can pick it at run time:

```ant
let little = true
let order: Bytes.Order = if little { Bytes.LittleEndian{} } else { Bytes.BigEndian{} }
let b = Bytes.new(4)
b.write(type(uint32), 0, uint32(1), order)
b // Bytes(01 00 00 00)
```

Call it through `.`, in a pipe, or with named arguments:

```ant
let b = Bytes.new(4)
b.write(type(byte), 0, byte(1), Bytes.BigEndian{})
b | Bytes.write(type(byte), 1, byte(2), Bytes.BigEndian{})
Bytes.write(b, t: type(uint16), at: 2, x: uint16(3), order: Bytes.BigEndian{})
b // Bytes(01 02 00 03)
```

Another name for the same collection sees the change. A copy from [Bytes.clone](clone.md) does not:

```ant
let b = Bytes.fromArray([0, 0])
let same = b
let copy = b.clone()
b.write(type(uint16), 0, uint16(0x0102), Bytes.BigEndian{})
same // Bytes(01 02)
copy // Bytes(00 00)
```

## Applications

Build a header of a known size with [Bytes.new](new.md), then write each field at its offset:

```ant
type Header = { version: uint16, count: uint32 }

func encode(h: Header): Bytes {
  let order = Bytes.LittleEndian{}
  let out = Bytes.new(8)
  out.copy(0, "AN".toBytes())
  out.write(type(uint16), 2, h.version, order)
  out.write(type(uint32), 4, h.count, order)
  out
}

let h = Header{ version: uint16(2), count: uint32(300) }
println(Encoding.toHex(encode(h)))
```

```text
414e02002c010000
```

Frame a message whose length is known only at the end. Start with four zero bytes from
[Bytes.new](new.md) as a placeholder, extend with the body, and then write the length over the
placeholder:

```ant
func frame(parts: Array[string]): Bytes {
  let out = Bytes.new(4)
  for part in parts {
    out.extend(part.toBytes())
  }
  out.write(type(uint32), 0, uint32(out.length() - 4), Bytes.BigEndian{})
  out
}

println(frame(["hello", ", ", "ant"]))
```

```text
Bytes(00 00 00 0a 68 65 6c 6c 6f 2c 20 61 6e 74)
```

Update a counter in place with [Bytes.read](read.md) and `Bytes.write` at the same offset:

```ant
func bump(b: Bytes, at: int) {
  let order = Bytes.BigEndian{}
  let n = b.read(type(uint16), at, order)
  b.write(type(uint16), at, n + uint16(1), order)
}

let counters = Bytes.fromArray([0, 0, 0, 0xff])
bump(counters, 2)
bump(counters, 2)
println(counters)
```

```text
Bytes(00 00 01 01)
```

## Properties and relations

The length never changes:

```ant
let b = "abcdefgh".toBytes()
let before = b.length()
b.write(type(float64), 0, 3.25, Bytes.BigEndian{})
assert(b.length() == before)
```

[Bytes.read](read.md) at the same index, type and order gives the number back:

```ant
let b = Bytes.new(8)
let order = Bytes.BigEndian{}
for x in [-1.5, 0.0, 1e300] {
  b.write(type(float64), 0, x, order)
  assert(b.read(type(float64), 0, order) == x)
}
```

Little-endian bytes are the big-endian bytes reversed:

```ant
let big = Bytes.new(4)
let little = Bytes.new(4)
big.write(type(uint32), 0, uint32(0x01020304), Bytes.BigEndian{})
little.write(type(uint32), 0, uint32(0x01020304), Bytes.LittleEndian{})
assert(big.toArray().reverse() == little.toArray())
```

Writing into zero bytes from [Bytes.new](new.md) gives the same bytes as [Bytes.append](append.md)
on empty `Bytes`:

```ant
let written = Bytes.new(2)
written.write(type(int16), 0, int16(-300), Bytes.LittleEndian{})
let appended = Bytes.new()
appended.append(type(int16), int16(-300), Bytes.LittleEndian{})
assert(written == appended)
```

## Possible issues

`Bytes.write` never grows `b`. A number that starts inside but runs past the end panics, and so does
a write to empty `Bytes`:

```ant
let b = Bytes.fromArray([1, 2, 3, 4])
b.write(type(uint32), 1, uint32(0), Bytes.BigEndian{})
// panic: Bytes.write: the range 1..5 is not within 0 ..= 4
```

```ant
Bytes.new().write(type(byte), 0, byte(0), Bytes.BigEndian{})
// panic: Bytes.write: the range 0..1 is not within 0 ..= 0
```

A negative index panics too:

```ant
let b = Bytes.new(4)
b.write(type(uint16), -1, uint16(0), Bytes.BigEndian{})
// panic: Bytes.write: the range -1..1 is not within 0 ..= 4
```

`int` takes eight bytes, not four. Name the width the format uses, such as `uint32`:

```ant
let b = Bytes.new(4)
b.write(type(int), 0, 1, Bytes.BigEndian{})
// panic: Bytes.write: the range 0..8 is not within 0 ..= 4
```

An integer literal is an `int`, never a narrower type, so it does not match a `uint16` type value:

```ant
let b = Bytes.new(2)
b.write(type(uint16), 0, 80, Bytes.BigEndian{}) // error: cannot use int where uint16 is expected
```

Converting a value that does not fit the type panics before `Bytes.write` runs:

```ant
let b = Bytes.new(2)
let n = 70000
b.write(type(uint16), 0, uint16(n), Bytes.BigEndian{}) // panic: 70000 out of range for uint16
```

[Bytes.slice](slice.md) returns a copy, so writing into a slice leaves the original as it was:

```ant
let b = Bytes.fromArray([1, 2, 3, 4])
let part = b.slice(0, 2)
part.write(type(uint16), 0, uint16(0), Bytes.BigEndian{})
part // Bytes(00 00)
b    // Bytes(01 02 03 04)
```

Writes that overlap overwrite each other. The `uint16` at 2 takes bytes 2 and 3, so a byte written
at 3 afterwards replaces half of it:

```ant
let b = Bytes.new(4)
b.write(type(uint16), 2, uint16(0x0102), Bytes.BigEndian{})
b.write(type(byte), 3, byte(0xff), Bytes.BigEndian{})
b // Bytes(00 00 01 ff)
```

There is no default order. Every call states one:

```ant
let b = Bytes.new(2)
b.write(type(uint16), 0, uint16(1)) // error: Bytes.write needs an argument for order
```

## See also

[Bytes.read](read.md) · [Bytes.append](append.md) · [Bytes.sizeOf](sizeOf.md) ·
[Bytes.Order](types/Order.md) · [Bytes.copy](copy.md) · [Bytes.fill](fill.md) · [Bytes.new](new.md)
