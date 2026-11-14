<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.read

```ant signature
func Bytes.read[T: Number](b: Bytes, t: Reflect.Type[T], at: int, order: Bytes.Order): T
```

Bytes.read returns the number of type t stored in b at index at in the given order. It panics unless
the number's bytes are within 0 ..= Bytes.length(b).

<!-- end generated -->

## Details

- `Bytes.read` takes `Bytes.sizeOf(t)` bytes from index `at`: 1, 2, 4, or 8. It panics unless all of
  them are inside `b`, so `at` may be at most `b.length() - Bytes.sizeOf(t)`.
- The result has the type that `t` names, such as `uint16`, not `int`. Convert it before mixing it
  with `int` values.
- Integers are two's complement and floats are IEEE 754 bits. A one-byte type ignores the order.

## Basic examples

Read the same four bytes as a big-endian and as a little-endian `uint32`:

```ant
let b = Bytes.fromArray([0x12, 0x34, 0x56, 0x78])
Bytes.read(b, type(uint32), 0, Bytes.BigEndian{})    // 305419896
Bytes.read(b, type(uint32), 0, Bytes.LittleEndian{}) // 2018915346
```

`at` picks the index where the number starts:

```ant
let b = Bytes.fromArray([0x12, 0x34, 0x56, 0x78])
Bytes.read(b, type(uint16), 2, Bytes.BigEndian{}) // 22136
```

## Scope

Signed types read the bytes as two's complement, so the same bytes give a different value:

```ant
let b = Bytes.fromArray([0xff, 0xfe])
b.read(type(uint16), 0, Bytes.BigEndian{}) // 65534
b.read(type(int16), 0, Bytes.BigEndian{})  // -2
b.read(type(int8), 0, Bytes.BigEndian{})   // -1
```

Floats read their IEEE 754 bits. `3f f8 00 00 00 00 00 00` is `1.5` as a `float64`, and its first
four bytes are `1.9375` as a `float32`:

```ant
let b = Bytes.fromArray([0x3f, 0xf8, 0, 0, 0, 0, 0, 0])
b.read(type(float64), 0, Bytes.BigEndian{}) // 1.5
b.read(type(float32), 0, Bytes.BigEndian{}) // 1.9375
```

`int` is `int64`, so it takes eight bytes:

```ant
let b = Bytes.fromArray([0, 0, 0, 0, 0, 0, 1, 0])
b.read(type(int), 0, Bytes.BigEndian{}) // 256
```

A one-byte type gives the same value in either order, the same as indexing:

```ant
let b = Bytes.fromArray([7, 200])
b.read(type(byte), 1, Bytes.BigEndian{})    // 200
b.read(type(byte), 1, Bytes.LittleEndian{}) // 200
b[1]                                        // 200
```

Call it with named arguments, or in a pipe, where `b` is the first argument:

```ant
let b = Bytes.fromArray([0x01, 0x02])
Bytes.read(b: b, t: type(uint16), at: 0, order: Bytes.BigEndian{}) // 258
b | Bytes.read(type(uint16), 0, Bytes.LittleEndian{})              // 513
```

## Applications

Read the width and height of a PNG image. They are big-endian `uint32` values at bytes 16 and 20,
after the signature and the start of the header chunk:

```ant
type Size = { width: int, height: int }

func pngSize(data: Bytes): Size | null {
  let signature = Bytes.fromArray([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a])
  if data.length() < 24 || !data.startsWith(signature) { return null }
  let order = Bytes.BigEndian{}
  Size{
    width: int(data.read(type(uint32), 16, order)),
    height: int(data.read(type(uint32), 20, order)),
  }
}

let header = Bytes.concat([
  Bytes.fromArray([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]),
  Bytes.fromArray([0, 0, 0, 13]),
  String.toBytes("IHDR"),
  Bytes.fromArray([0, 0, 0x02, 0x80, 0, 0, 0x01, 0xe0]),
])
println(pngSize(header))
println(pngSize(String.toBytes("GIF89a")))
```

```text
Size{ width: 640, height: 480 }
null
```

A TIFF file states its own order in its first two bytes, `II` for little-endian and `MM` for
big-endian. Pick the order once and pass it to every read:

```ant
func tiffOrder(header: Bytes): Bytes.Order | null {
  if header.startsWith(String.toBytes("II")) { return Bytes.LittleEndian{} }
  if header.startsWith(String.toBytes("MM")) { return Bytes.BigEndian{} }
  null
}

func firstDirectory(header: Bytes): int | null {
  if header.length() < 8 { return null }
  let order = tiffOrder(header) ?? return null
  if header.read(type(uint16), 2, order) != uint16(42) { return null }
  int(header.read(type(uint32), 4, order))
}

println(firstDirectory(Bytes.fromArray([0x49, 0x49, 42, 0, 8, 0, 0, 0])))
println(firstDirectory(Bytes.fromArray([0x4d, 0x4d, 0, 42, 0, 0, 0, 8])))
println(firstDirectory(Bytes.fromArray([0x4d, 0x4d, 42, 0, 0, 0, 0, 8])))
```

```text
8
8
null
```

Walk length-prefixed records without slicing off the rest of the input at each step. The position is
a `var` that the loop moves forward past each length and its bytes:

```ant
func records(data: Bytes): Array[string] {
  let out: Array[string] = []
  var at = 0
  while at + 2 <= data.length() {
    let n = int(data.read(type(uint16), at, Bytes.BigEndian{}))
    at += 2
    if at + n > data.length() { break }
    out.push(String.fromBytesLossy(data.slice(at, at + n)))
    at += n
  }
  out
}

let data = Bytes.concat([
  Bytes.fromArray([0, 3]),
  String.toBytes("ant"),
  Bytes.fromArray([0, 5]),
  String.toBytes("hello"),
])
println(records(data))
```

```text
["ant", "hello"]
```

## Properties and relations

`Bytes.read` gives back what [`Bytes.write`](write.md) wrote, at the same index, type, and order:

```ant
let b = Bytes.new(8)
let order = Bytes.LittleEndian{}
for x in [int32(0), int32(-1), int32(123456), int32(-2147483648)] {
  b.write(type(int32), 4, x, order)
  assert(b.read(type(int32), 4, order) == x)
}
```

Little-endian is big-endian with the bytes reversed:

```ant
let b = Bytes.fromArray([0xca, 0xfe, 0xba, 0xbe])
let reversed = Bytes.fromArray(b.toArray().reverse())
let little = b.read(type(uint32), 0, Bytes.LittleEndian{})
assert(little == reversed.read(type(uint32), 0, Bytes.BigEndian{}))
```

A big-endian read puts the first byte highest:

```ant
let b = Bytes.fromArray([0x12, 0x34])
let manual = int(b[0]) * 256 + int(b[1])
assert(int(b.read(type(uint16), 0, Bytes.BigEndian{})) == manual)
```

## Possible issues

The bytes must all be inside `b`. A read that starts inside but runs past the end panics, and so
does a read at the length:

```ant
let b = Bytes.fromArray([1, 2, 3, 4])
b.read(type(uint32), 1, Bytes.BigEndian{}) // panic: Bytes.read: the range 1..5 is not within 0 ..= 4
b.read(type(byte), 4, Bytes.BigEndian{})   // panic: Bytes.read: the range 4..5 is not within 0 ..= 4
```

`int` takes eight bytes, not four. Name the width the format uses, such as `uint32`:

```ant
let b = Bytes.fromArray([0, 0, 0, 1])
b.read(type(int), 0, Bytes.BigEndian{}) // panic: Bytes.read: the range 0..8 is not within 0 ..= 4
```

The result is not an `int`, and an integer literal is, so arithmetic and comparison with one are
compile errors:

```ant
let b = Bytes.fromArray([0, 5])
let n = b.read(type(uint16), 0, Bytes.BigEndian{})
println(n + 1) // error: + needs both operands of one type, got uint16 and int
```

Convert the result, or the literal:

```ant
let b = Bytes.fromArray([0, 5])
let n = b.read(type(uint16), 0, Bytes.BigEndian{})
int(n) + 1     // 6
n == uint16(5) // true
```

The type argument is a type value, written `type(uint16)`. A type name alone is not a value:

```ant
let b = Bytes.fromArray([1, 2])
println(b.read(uint16, 0, Bytes.BigEndian{}))
// error: uint16 is a type name; write type(uint16) for its value
```

`t` names one number type. A union of number types, or `Number` itself, is a compile error:

```ant
let b = Bytes.fromArray([1, 2, 3, 4])
println(b.read(type(uint16 | uint32), 0, Bytes.BigEndian{}))
// error: T would be uint16 | uint32, which is not a single member of its bound Number; narrow first
```

The order has no default, even for a one-byte type:

```ant
let b = Bytes.fromArray([1, 2])
println(b.read(type(byte), 0)) // error: Bytes.read needs an argument for order
```

## See also

[Bytes.write](write.md) · [Bytes.append](append.md) · [Bytes.sizeOf](sizeOf.md) ·
[Bytes.Order](types/Order.md) · [Bytes.startsWith](startsWith.md) · [Bytes.slice](slice.md) ·
[Bytes.indexOf](indexOf.md)
