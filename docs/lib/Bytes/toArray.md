<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.toArray

```ant signature
func Bytes.toArray(b: Bytes): Array[byte]
```

Bytes.toArray returns the bytes of b as an array.

<!-- end generated -->

## Details

- The array is new. Changing it leaves `b` as it was, since `Bytes` is immutable.

## Basic examples

Turn bytes into an array of numbers:

```ant
Bytes.toArray(Bytes.fromArray([0x89, 0x50, 0x4e, 0x47])) // [137, 80, 78, 71]
```

Read the UTF-8 bytes of a string as numbers:

```ant
Bytes.toArray(String.toBytes("hi")) // [104, 105]
```

## Scope

Empty bytes give an empty array:

```ant
Bytes.toArray(String.toBytes("")) // []
```

A character outside ASCII takes several bytes, each its own element:

```ant
Bytes.toArray(String.toBytes("é"))  // [195, 169]
Bytes.toArray(String.toBytes("🐜")) // [240, 159, 144, 156]
```

Call it through `.` or in a pipe:

```ant
let b = String.toBytes("Ant")
b.toArray()         // [65, 110, 116]
b | Bytes.toArray() // [65, 110, 116]
```

## Applications

Change one byte and build new bytes from the array. Here the version byte of a small header goes
from 1 to 2:

```ant
let header = Bytes.fromArray([0x41, 0x4e, 0x54, 0x01])
let fields = Bytes.toArray(header)
fields[3] = byte(2)
Bytes.fromArray(fields) // Bytes(41 4e 54 02)
```

Flip the case of ASCII letters with `Bits.xor`, then decode the result:

```ant
let flipped = String.toBytes("Ant").toArray().map({ Bits.xor(x, byte(0x20)) })
String.fromBytes(Bytes.fromArray(flipped)) // "aNT"
```

Add up the bytes as `int`s for a simple checksum, kept in one byte with `%`:

```ant
func checksum(b: Bytes): int {
  (Bytes.toArray(b).map({ int(x) }).sum() ?? 0) % 256
}

checksum(String.toBytes("hello")) // 20
```

Count the bytes that are not ASCII:

```ant
Bytes.toArray(String.toBytes("naïve café")).countWhere({ x > byte(127) }) // 4
```

## Properties and relations

`Bytes.fromArray` turns the array back into the same bytes:

```ant
let b = String.toBytes("ant")
assert(Bytes.fromArray(Bytes.toArray(b)) == b)
```

The array has one element per byte, in order, each equal to indexing `b`:

```ant
let b = String.toBytes("añt")
let xs = Bytes.toArray(b)
assert(len(xs) == len(b))
assert(Array.every(Array.range(0, len(b)), { xs[x] == b[x] }))
```

Changing the array leaves the bytes as they were:

```ant
let b = String.toBytes("ant")
let xs = Bytes.toArray(b)
xs[0] = byte(0)
assert(b == String.toBytes("ant"))
```

## Possible issues

Indexing, `len`, and `for` work on `Bytes` directly; convert with `Bytes.toArray` only to use an
`Array` function or to change bytes:

```ant
let b = String.toBytes("hi")
b[0]   // 104
len(b) // 2
```

```ant
let b = String.toBytes("hi")
for x in b {
  println(x)
}
```

```text
104
105
```

An element is a `byte`, so adding an `int` literal to it is a compile error. Convert one side:

```ant
let xs = Bytes.toArray(String.toBytes("A"))
let next = xs[0] + 1 // error: + needs both operands of one type, got byte and int
```

```ant
let xs = Bytes.toArray(String.toBytes("A"))
xs[0] + byte(1) // 66
int(xs[0]) + 1  // 66
```

Arithmetic on `byte`s panics outside 0 to 255. Compute in `int` instead:

```ant
let xs = Bytes.toArray(Bytes.fromArray([200, 100]))
xs.sum()                      // panic: Array.sum: integer overflow
xs.map({ int(x) }).sum() ?? 0 // 300
xs[1] - xs[0]                 // panic: integer overflow
int(xs[1]) - int(xs[0])       // -100
```

An array of bytes and `Bytes` are different types, so `==` between them is a compile error. Convert
one side first:

```ant
let b = String.toBytes("A")
let same = Bytes.toArray(b) == b
// error: Array[byte] and Bytes cannot be compared with ==: neither is assignable to the other
```

The array renders its bytes in decimal, while `Bytes` renders them in hexadecimal:

```ant
let b = Bytes.fromArray([0x0a, 0xff])
b           // Bytes(0a ff)
b.toArray() // [10, 255]
```

## See also

[Bytes.fromArray](fromArray.md), [Bytes.length](length.md), [Bytes.slice](slice.md),
[String.toBytes](../String/toBytes.md), [String.fromBytes](../String/fromBytes.md),
[Encoding.toHex](../Encoding/toHex.md), [Bits.xor](../Bits/xor.md)
