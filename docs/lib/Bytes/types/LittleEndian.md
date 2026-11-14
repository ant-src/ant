<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.LittleEndian

```ant signature
type Bytes.LittleEndian = {}
```

Bytes.LittleEndian is the byte order with the least significant byte first.

<!-- end generated -->

## Details

- Bytes.LittleEndian is an empty record, so you pass the value `Bytes.LittleEndian{}`, not the type
  name.
- It is one of the two members of [`Bytes.Order`](Order.md), which [`Bytes.read`](../read.md),
  [`Bytes.write`](../write.md), and [`Bytes.append`](../append.md) take as their last argument.
  There is no default order.
- A one-byte type such as `uint8` takes the order and ignores it.

## Basic examples

Append a `uint32` with the least significant byte first:

```ant
let b = Bytes.new()
b.append(type(uint32), uint32(0x01020304), Bytes.LittleEndian{})
b // Bytes(04 03 02 01)
```

Read the bytes back in the same order:

```ant
let b = Bytes.fromArray([0x04, 0x03, 0x02, 0x01])
b.read(type(uint32), 0, Bytes.LittleEndian{}) == uint32(0x01020304) // true
```

The value renders as an empty record and equals every other `Bytes.LittleEndian{}`:

```ant
Bytes.LittleEndian{}                         // Bytes.LittleEndian{}
Bytes.LittleEndian{} == Bytes.LittleEndian{} // true
```

## Scope

### Integer widths

Every multi-byte integer is stored least significant byte first, a signed one in two's complement:

```ant
let b = Bytes.new()
b.append(type(uint16), uint16(0x0102), Bytes.LittleEndian{})
b.append(type(int16), int16(-2), Bytes.LittleEndian{})
b.append(type(int64), int64(1), Bytes.LittleEndian{})
b // Bytes(02 01 fe ff 01 00 00 00 00 00 00 00)
```

### Floats

A float is stored as its IEEE 754 bits, low byte first:

```ant
let b = Bytes.new()
b.append(type(float32), float32(1.0), Bytes.LittleEndian{})
b.append(type(float64), 1.5, Bytes.LittleEndian{})
b                                              // Bytes(00 00 80 3f 00 00 00 00 00 00 f8 3f)
b.read(type(float64), 4, Bytes.LittleEndian{}) // 1.5
```

### As a `Bytes.Order`

A variable of type `Bytes.Order` holds either order, and `is` tells which:

```ant
let order: Bytes.Order = Bytes.LittleEndian{}
order is Bytes.LittleEndian                           // true
if order is Bytes.LittleEndian { "LE" } else { "BE" } // "LE"
```

### Writing in place

`Bytes.write` overwrites bytes that already exist, here the middle two of four:

```ant
let b = Bytes.new(4)
b.write(type(uint16), 1, uint16(0xabcd), Bytes.LittleEndian{})
b // Bytes(00 cd ab 00)
```

## Applications

### Read a BMP header

Most file formats from PCs, such as BMP, WAV, and ZIP, store their numbers little-endian. A BMP file
keeps its image width and height as `int32` values at offsets 18 and 22:

```ant
func bmpSize(header: Bytes): string {
  let le = Bytes.LittleEndian{}
  let width = header.read(type(int32), 18, le)
  let height = header.read(type(int32), 22, le)
  "${width}x${height}"
}

let header = Bytes.new(26)
header.write(type(int32), 18, int32(640), Bytes.LittleEndian{})
header.write(type(int32), 22, int32(480), Bytes.LittleEndian{})
println(bmpSize(header))
```

```text
640x480
```

### Pick the order from the data

A TIFF file starts with `II` when its numbers are little-endian and `MM` when they are big-endian.
Choose the order once, then pass it to every read:

```ant
func tiffOrder(data: Bytes): Bytes.Order {
  if data.startsWith("II".toBytes()) { Bytes.LittleEndian{} } else { Bytes.BigEndian{} }
}

let data = Bytes.fromArray([0x49, 0x49, 0x2a, 0x00])
let order = tiffOrder(data)
println(order)
println(data.read(type(uint16), 2, order))
```

```text
Bytes.LittleEndian{}
42
```

## Properties and relations

Reading with the order a value was written in gives the value back:

```ant
let b = Bytes.new()
b.append(type(int32), int32(-123456), Bytes.LittleEndian{})
assert(b.read(type(int32), 0, Bytes.LittleEndian{}) == int32(-123456))
```

The little-endian bytes of a number are its big-endian bytes reversed:

```ant
let le = Bytes.new()
let be = Bytes.new()
le.append(type(uint64), uint64(0x0102030405060708), Bytes.LittleEndian{})
be.append(type(uint64), uint64(0x0102030405060708), Bytes.BigEndian{})
assert(le.toArray() == be.toArray().reverse())
```

## Possible issues

### The type name is not a value

Pass `Bytes.LittleEndian{}`; the bare name is a type:

```ant
let b = Bytes.fromArray([0x01, 0x02])
let n = b.read(type(uint16), 0, Bytes.LittleEndian)
// error: Bytes.LittleEndian is a type, not a value; a type appears only as a literal head, after is, or before a dot
```

### The wrong order still reads a number

Nothing in the bytes records their order, so reading little-endian data as big-endian gives a
different number and no error:

```ant
let b = Bytes.fromArray([0x01, 0x02])
b.read(type(uint16), 0, Bytes.LittleEndian{}) // 513
b.read(type(uint16), 0, Bytes.BigEndian{})    // 258
```

## See also

[Bytes.BigEndian](BigEndian.md) · [Bytes.Order](Order.md) · [Bytes.read](../read.md) ·
[Bytes.write](../write.md) · [Bytes.append](../append.md) · [Bytes.sizeOf](../sizeOf.md)
