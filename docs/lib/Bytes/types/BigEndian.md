<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.BigEndian

```ant signature
type Bytes.BigEndian = {}
```

Bytes.BigEndian is the byte order with the most significant byte first, as network formats use.

<!-- end generated -->

## Details

- `Bytes.BigEndian` is an empty record, so you pass the value `Bytes.BigEndian{}`, not the type
  name.
- [Bytes.read](../read.md), [Bytes.write](../write.md), and [Bytes.append](../append.md) take the
  order as their last argument. There is no default order.
- It puts the most significant byte of a number at the lowest index. Network protocols and many file
  formats, PNG among them, store numbers this way.
- A one-byte type, such as `byte` or `int8`, takes the order and ignores it.

## Basic examples

Pass `Bytes.BigEndian{}` as the order of [Bytes.append](../append.md). The high byte comes first:

```ant
let b = Bytes.new()
b.append(type(uint32), uint32(0x12345678), Bytes.BigEndian{})
b // Bytes(12 34 56 78)
```

Read a number stored most significant byte first with [Bytes.read](../read.md):

```ant
let b = Bytes.fromArray([0x00, 0x00, 0x01, 0x00])
b.read(type(uint32), 0, Bytes.BigEndian{}) // 256
```

## Scope

Keep the order in a variable to use it in several calls:

```ant
let be = Bytes.BigEndian{}
let b = Bytes.new()
b.append(type(uint16), uint16(80), be)
b.append(type(uint16), uint16(443), be)
b                           // Bytes(00 50 01 bb)
b.read(type(uint16), 2, be) // 443
```

Signed integers are stored in two's complement, and floats as their IEEE 754 bits, high byte first:

```ant
let b = Bytes.new()
b.append(type(int16), int16(-2), Bytes.BigEndian{})
b.append(type(float64), 1.5, Bytes.BigEndian{})
b // Bytes(ff fe 3f f8 00 00 00 00 00 00)
```

[Bytes.write](../write.md) overwrites bytes in place in the same order:

```ant
let b = Bytes.new(4)
b.write(type(int32), 0, int32(-1), Bytes.BigEndian{})
b // Bytes(ff ff ff ff)
```

A value of type `Bytes.BigEndian` is a [Bytes.Order](Order.md), and `is` tells it apart from
[Bytes.LittleEndian](LittleEndian.md):

```ant
let order: Bytes.Order = Bytes.BigEndian{}
order is Bytes.BigEndian    // true
order is Bytes.LittleEndian // false
```

## Applications

Prefix a message with its length as two big-endian bytes, as many network protocols frame data:

```ant
func frame(payload: Bytes): Bytes {
  let out = Bytes.new()
  out.append(type(uint16), uint16(len(payload)), Bytes.BigEndian{})
  out.extend(payload)
  out
}

println(frame(String.toBytes("hi")))
```

```text
Bytes(00 02 68 69)
```

Read a frame back: the first two bytes, high first, give the payload length:

```ant
func unframe(data: Bytes): Bytes {
  let n = data.read(type(uint16), 0, Bytes.BigEndian{})
  data.slice(2, 2 + int(n))
}

println(String.fromBytes(unframe(Bytes.fromArray([0x00, 0x02, 0x68, 0x69, 0xff]))))
```

```text
hi
```

Format an IPv4 address held in a `uint32`. Its bytes, high first, are the four parts:

```ant
func ipv4(n: uint32): string {
  let b = Bytes.new()
  b.append(type(uint32), n, Bytes.BigEndian{})
  b.toArray() | Array.map({ String.from(x) }) | String.join(".")
}

println(ipv4(uint32(0xc0a80001)))
```

```text
192.168.0.1
```

Pick the order at run time, for a format whose header says which one its numbers use:

```ant
func orderOf(bigEndian: bool): Bytes.Order {
  if bigEndian { Bytes.BigEndian{} } else { Bytes.LittleEndian{} }
}

let b = Bytes.new()
b.append(type(uint16), uint16(0x0102), orderOf(true))
b.append(type(uint16), uint16(0x0102), orderOf(false))
println(b)
```

```text
Bytes(01 02 02 01)
```

## Properties and relations

Reading with the order a number was written in gives the number back:

```ant
let b = Bytes.new()
b.append(type(int64), int64(-123456789), Bytes.BigEndian{})
assert(b.read(type(int64), 0, Bytes.BigEndian{}) == int64(-123456789))
assert(len(b) == Bytes.sizeOf(type(int64)))
```

The big-endian bytes of a number are its little-endian bytes reversed:

```ant
let big = Bytes.new()
big.append(type(uint32), uint32(0xcafe), Bytes.BigEndian{})
let little = Bytes.new()
little.append(type(uint32), uint32(0xcafe), Bytes.LittleEndian{})
assert(big.toArray().reverse() == little.toArray())
```

For a one-byte type, both orders give the same byte:

```ant
let a = Bytes.new()
a.append(type(byte), byte(7), Bytes.BigEndian{})
let b = Bytes.new()
b.append(type(byte), byte(7), Bytes.LittleEndian{})
assert(a == b)
```

All `Bytes.BigEndian` values are equal:

```ant
assert(Bytes.BigEndian{} == Bytes.BigEndian{})
```

## Possible issues

The order is a value, so it needs braces. The bare name is a type:

```ant
let b = Bytes.new()
b.append(type(uint16), uint16(1), Bytes.BigEndian)
// error: Bytes.BigEndian is a type, not a value; a type appears only as a literal head, after is, or before a dot
```

There is no default order. Leaving it out is a compile error:

```ant
let b = Bytes.new()
b.append(type(uint16), uint16(1)) // error: Bytes.append needs an argument for order
```

An integer literal is always an `int`, so convert it to the type the type value names:

```ant
let b = Bytes.new()
b.append(type(uint32), 1, Bytes.BigEndian{}) // error: cannot use int where uint32 is expected
```

Reading with the wrong order gives a different number, not an error:

```ant
let b = Bytes.fromArray([0x00, 0x00, 0x00, 0x0d])
b.read(type(uint32), 0, Bytes.BigEndian{})    // 13
b.read(type(uint32), 0, Bytes.LittleEndian{}) // 218103808
```

A read past the end panics, whatever the order:

```ant
let b = Bytes.fromArray([0, 1])
b.read(type(uint16), 1, Bytes.BigEndian{}) // panic: Bytes.read: the range 1..3 is not within 0 ..= 2
```

## See also

[Bytes.LittleEndian](LittleEndian.md) · [Bytes.Order](Order.md) · [Bytes.read](../read.md) ·
[Bytes.write](../write.md) · [Bytes.append](../append.md) · [Bytes.sizeOf](../sizeOf.md) ·
[Bytes](../index.md)
