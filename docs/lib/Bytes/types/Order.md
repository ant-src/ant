<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.Order

```ant signature
type Bytes.Order = Bytes.BigEndian | Bytes.LittleEndian
```

Bytes.Order is the byte order of a multi-byte number in [`Bytes.read`](../read.md),
[`Bytes.write`](../write.md) and [`Bytes.append`](../append.md).

<!-- end generated -->

## Details

- The two members are records with no fields, so a value is written with braces:
  `Bytes.BigEndian{}`, `Bytes.LittleEndian{}`.
- A one-byte type, `byte`, `uint8` or `int8`, takes an order and ignores it.

## Basic examples

The same `uint32` in each order. `Bytes.BigEndian` puts the most significant byte first,
`Bytes.LittleEndian` the least significant:

```ant
let big = Bytes.new()
big.append(type(uint32), uint32(0x0a0b0c0d), Bytes.BigEndian{})
big // Bytes(0a 0b 0c 0d)
let little = Bytes.new()
little.append(type(uint32), uint32(0x0a0b0c0d), Bytes.LittleEndian{})
little // Bytes(0d 0c 0b 0a)
```

The order decides which number the same bytes hold:

```ant
let b = Bytes.fromArray([0x01, 0x02])
b.read(type(uint16), 0, Bytes.BigEndian{})    // 258
b.read(type(uint16), 0, Bytes.LittleEndian{}) // 513
```

Take `Bytes.Order` as a parameter to write code that works for either order:

```ant
func encodeAll(xs: Array[uint16], order: Bytes.Order): Bytes {
  let out = Bytes.new()
  for x in xs {
    out.append(type(uint16), x, order)
  }
  out
}

encodeAll([uint16(1), uint16(2)], Bytes.BigEndian{})    // Bytes(00 01 00 02)
encodeAll([uint16(1), uint16(2)], Bytes.LittleEndian{}) // Bytes(01 00 02 00)
```

## Scope

A value of the union renders with its member's name, and compares equal to another of the same
member:

```ant
let order: Bytes.Order = Bytes.BigEndian{}
order                                        // Bytes.BigEndian{}
order is Bytes.BigEndian                     // true
order == Bytes.LittleEndian{}                // false
Bytes.LittleEndian{} == Bytes.LittleEndian{} // true
String.from(Bytes.LittleEndian{})            // "Bytes.LittleEndian{}"
```

Keep the order in a name and pass it to every call of one format:

```ant
let order = Bytes.LittleEndian{}
let b = Bytes.new()
b.append(type(uint16), uint16(1), order)
b.append(type(int32), int32(-1), order)
b                             // Bytes(01 00 ff ff ff ff)
b.read(type(int32), 2, order) // -1
```

Choose an order with an `if` expression:

```ant
let littleEndian = true
let order: Bytes.Order = if littleEndian { Bytes.LittleEndian{} } else { Bytes.BigEndian{} }
order // Bytes.LittleEndian{}
```

Test the member with `is` to branch on it:

```ant
func describe(order: Bytes.Order): string {
  if order is Bytes.BigEndian {
    "most significant byte first"
  } else {
    "least significant byte first"
  }
}

describe(Bytes.BigEndian{})    // "most significant byte first"
describe(Bytes.LittleEndian{}) // "least significant byte first"
```

A record field can hold an order:

```ant
type Format = { name: string, order: Bytes.Order }

let formats = [
  Format{ name: "network", order: Bytes.BigEndian{} },
  Format{ name: "x86", order: Bytes.LittleEndian{} },
]
let big = formats | Array.filter({ x.order is Bytes.BigEndian }) | Array.map({ x.name })
big // ["network"]
```

## Applications

A TIFF file states its own byte order in its first two bytes, `II` for little-endian and `MM` for
big-endian. Read the order from the header, then use it for every number after it. A header shorter
than four bytes fails before the read:

```ant
func byteOrder(header: Bytes): Result[Bytes.Order, Error] {
  if header.startsWith("II".toBytes()) {
    return Result.ok(Bytes.LittleEndian{})
  }
  if header.startsWith("MM".toBytes()) {
    return Result.ok(Bytes.BigEndian{})
  }
  Result.fail("not a TIFF file")
}

func magic(header: Bytes): Result[uint16, Error] {
  if header.length() < 4 {
    return Result.fail("not a TIFF file")
  }
  let order = try byteOrder(header)
  Result.ok(header.read(type(uint16), 2, order))
}

println(magic(Bytes.fromArray([0x49, 0x49, 0x2a, 0x00])))
println(magic(Bytes.fromArray([0x4d, 0x4d, 0x00, 0x2a])))
println(magic(Bytes.fromArray([0x50, 0x4b, 0x03, 0x04])))
println(magic(Bytes.fromArray([0x49, 0x49])))
```

```text
Ok{ value: 42 }
Ok{ value: 42 }
Err{ error: Error{ message: "not a TIFF file" } }
Err{ error: Error{ message: "not a TIFF file" } }
```

Convert a number from one order to the other by writing it in one and reading it in the other:

```ant
func swap32(x: uint32): uint32 {
  let b = Bytes.new()
  b.append(type(uint32), x, Bytes.BigEndian{})
  b.read(type(uint32), 0, Bytes.LittleEndian{})
}

swap32(uint32(1))                                // 16777216
swap32(uint32(0x11223344)) == uint32(0x44332211) // true
```

## Properties and relations

The bytes of one order are the bytes of the other reversed:

```ant
let big = Bytes.new()
let little = Bytes.new()
big.append(type(float64), 1.5, Bytes.BigEndian{})
little.append(type(float64), 1.5, Bytes.LittleEndian{})
assert(big.toArray().reverse() == little.toArray())
```

Reading in the order that wrote gives the number back:

```ant
let orders: Array[Bytes.Order] = [Bytes.BigEndian{}, Bytes.LittleEndian{}]
for order in orders {
  let b = Bytes.new()
  b.append(type(int64), -123456789, order)
  assert(b.read(type(int64), 0, order) == -123456789)
}
```

A one-byte type gives the same byte in either order:

```ant
let big = Bytes.new()
let little = Bytes.new()
big.append(type(int8), int8(-2), Bytes.BigEndian{})
little.append(type(int8), int8(-2), Bytes.LittleEndian{})
assert(big == little)
```

## Possible issues

Reading in the order the data was not written in gives a different number, with no error:

```ant
let b = Bytes.new()
b.append(type(uint32), uint32(1), Bytes.LittleEndian{})
b.read(type(uint32), 0, Bytes.BigEndian{}) // 16777216
```

The order is a value, so it takes `{}`:

```ant
let b = Bytes.fromArray([1, 2])
println(b.read(type(uint16), 0, Bytes.BigEndian))
// error: Bytes.BigEndian is a type, not a value; a type appears only as a literal head, after is, or before a dot
```

There is no default order. Every call states one:

```ant
let b = Bytes.fromArray([1, 2])
println(b.read(type(uint16), 0)) // error: Bytes.read needs an argument for order
```

## See also

[Bytes.BigEndian](BigEndian.md) · [Bytes.LittleEndian](LittleEndian.md) · [Bytes.read](../read.md) ·
[Bytes.write](../write.md) · [Bytes.append](../append.md) · [Bytes.sizeOf](../sizeOf.md) ·
[Bytes](../index.md)
