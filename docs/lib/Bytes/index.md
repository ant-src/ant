<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes

```ant signature
type Bytes
```

Bytes is an immutable sequence of bytes, for binary data. It compares by contents with ==, so it
serves as a map key; b\[i\] reads one byte and for x in b yields each.

<!-- end generated -->

## Details

- `Bytes` has no literal. Build it with [Bytes.fromArray](fromArray.md), from text with
  [String.toBytes](../String/toBytes.md), from hex or Base64 with
  [Encoding.fromHex](../Encoding/fromHex.md) and [Encoding.fromBase64](../Encoding/fromBase64.md),
  or read it with [File.readBytes](../File/readBytes.md).
- An index outside `0 ..= len(b) - 1` panics.
- `Bytes` is immutable: there is no `b[i] = x`. Every function returns a new value.
- `Json.stringify` panics on `Bytes`. Put `Encoding.toBase64(b)` or `Encoding.toHex(b)` in the JSON
  instead.

## Basic examples

`String.toBytes` gives the UTF-8 bytes of a string. `Bytes` renders as hexadecimal pairs:

```ant
let b = String.toBytes("hi!")
b      // Bytes(68 69 21)
len(b) // 3
b[0]   // 104
```

`Bytes.fromArray` writes a constant from integer literals:

```ant
Bytes.fromArray([0x89, 0x50, 0x4e, 0x47]) // Bytes(89 50 4e 47)
```

A `for` loop visits each byte:

```ant
for x in String.toBytes("Ant") {
  println(x)
}
```

```text
65
110
116
```

## Scope

Empty bytes come from an empty string or an empty array:

```ant
let none: Array[byte] = []
String.toBytes("")    // Bytes()
Bytes.fromArray(none) // Bytes()
```

A byte is a `byte`, not a one-byte `Bytes`. Slice to keep `Bytes`:

```ant
let b = String.toBytes("ant")
b[1]          // 110
b.slice(1, 2) // Bytes(6e)
```

Calls chain through `.`, or the bytes go down a pipe:

```ant
let data = String.toBytes("key=value")
data.indexOf(String.toBytes("="))             // 3
data | Bytes.slice(4, 9) | String.fromBytes() // "value"
```

`Bytes` is a map key by contents:

```ant
let names: Map[Bytes, string] = Map{}
names[Bytes.fromArray([0x89, 0x50])] = "png"
names[Bytes.fromArray([0x89, 0x50])] // "png"
```

## Functions by job

### Building bytes

- [Bytes.fromArray](fromArray.md): the bytes holding an array of integers, one byte each
- [Bytes.concat](concat.md): an array of `Bytes` joined in order

### Size and parts

- [Bytes.length](length.md): the number of bytes; the same as `len(b)`
- [Bytes.slice](slice.md): the bytes from one index up to but excluding another

### Searching

- [Bytes.indexOf](indexOf.md): the index of the first occurrence of `sub`, or `null`
- [Bytes.startsWith](startsWith.md): whether `b` begins with a prefix
- [Bytes.endsWith](endsWith.md): whether `b` ends with a suffix

### Bytes as an array

- [Bytes.toArray](toArray.md): the bytes as an `Array[byte]`, to change or compute with

## Applications

Check a file signature:

```ant
let pngMagic = Bytes.fromArray([0x89, 0x50, 0x4e, 0x47])
let header = Bytes.concat([pngMagic, String.toBytes("rest")])
header.startsWith(pngMagic) // true
```

Change one byte: convert to an array, change it, and convert back:

```ant
let b = String.toBytes("cat")
let xs = b.toArray()
xs[0] = byte(0x62)
Bytes.fromArray(xs) // Bytes(62 61 74)
```

Sum the bytes into a simple checksum:

```ant
var sum = 0
for x in String.toBytes("abc") {
  sum += int(x)
}
sum // 294
```

Decode bytes as text, which gives `null` when they are not valid UTF-8:

```ant
String.fromBytes(Bytes.fromArray([0xe6, 0x97, 0xa5])) // "日"
String.fromBytes(Bytes.fromArray([0xff]))             // null
```

`String.fromBytesLossy` decodes bytes that may be invalid, such as file or process output, and
writes U+FFFD for each bad sequence:

```ant
String.fromBytesLossy(Bytes.fromArray([0x61, 0xff])) // "a�"
```

Hash a string and print the hex digest:

```ant
Crypto.sha256(String.toBytes("abc")) | Encoding.toHex() // "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
```

Write bytes as hex text and read them back:

```ant
let b = Bytes.fromArray([0xca, 0xfe])
Encoding.toHex(b)        // "cafe"
Encoding.fromHex("cafe") // Ok{ value: Bytes(ca fe) }
```

## Properties and relations

`len`, `Bytes.length`, and `Bytes.toArray` agree:

```ant
let b = String.toBytes("héllo")
assert(len(b) == Bytes.length(b))
assert(len(b.toArray()) == len(b))
```

`Bytes.fromArray` and `Bytes.toArray` undo each other:

```ant
let b = String.toBytes("ant")
assert(Bytes.fromArray(b.toArray()) == b)
```

Slices joined back give the whole:

```ant
let b = String.toBytes("binary")
assert(Bytes.concat([b.slice(0, 3), b.slice(3, len(b))]) == b)
```

## Possible issues

There is no `b[i] = x`. Assigning a byte is a compile error:

```ant
let b = String.toBytes("hi")
b[0] = 1
// error: Bytes is immutable, so a byte cannot be assigned; build an Array[byte] and convert it with Bytes.fromArray
```

An index past the end panics:

```ant
let b = String.toBytes("hi")
b[2] // panic: index 2 out of range for Bytes of length 2
```

`Json.stringify` has no form for `Bytes`. Encode them first:

```ant
let b = Bytes.fromArray([1, 2])
Json.stringify(b) // panic: Json.stringify: Bytes has no JSON form; write Encoding.toBase64 of it
```

```ant
let b = Bytes.fromArray([1, 2])
Json.stringify(Encoding.toBase64(b)) // "\"AQI=\""
```

`Bytes.fromArray` panics on an element outside `0 ..= 255`, which int arithmetic can produce:

```ant
Bytes.fromArray([0x89, 256]) // panic: Bytes.fromArray needs integers from 0 to 255, got 256 at index 1
```

A string escape writes a character, not a byte. `"\u{89}"` is two bytes in UTF-8:

```ant
String.toBytes("\u{89}") // Bytes(c2 89)
Bytes.fromArray([0x89])  // Bytes(89)
```

## See also

- [String.toBytes](../String/toBytes.md), [String.fromBytes](../String/fromBytes.md),
  [String.fromBytesLossy](../String/fromBytesLossy.md): text to bytes and back
- [Encoding.toHex](../Encoding/toHex.md), [Encoding.toBase64](../Encoding/toBase64.md): bytes as
  text
- [Crypto.sha256](../Crypto/sha256.md): the SHA-256 digest of bytes
- [File.readBytes](../File/readBytes.md), [File.writeBytes](../File/writeBytes.md): bytes from and
  to files
- [byte](../../language/types/byte.md): the type of one byte
