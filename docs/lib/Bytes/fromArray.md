<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.fromArray

```ant signature
func Bytes.fromArray[T: Integer](xs: Array[T]): Bytes
```

Bytes.fromArray returns the bytes holding the integers in xs, one byte each. It panics if an element
is outside 0 ..= 255.

<!-- end generated -->

## Details

- A negative element panics like any other out-of-range one: `-1` does not wrap to `ff`.
- The result is a new `Bytes` value. Changing `xs` afterwards does not change it.

## Basic examples

Write a byte constant with hexadecimal literals:

```ant
Bytes.fromArray([0x89, 0x50, 0x4e, 0x47]) // Bytes(89 50 4e 47)
```

Any integer literal works, in any base:

```ant
Bytes.fromArray([72, 105])            // Bytes(48 69)
Bytes.fromArray([0b1010, 0o17, 0xff]) // Bytes(0a 0f ff)
```

## Scope

An empty array gives empty bytes. Its element type must be known:

```ant
let none: Array[int] = []
Bytes.fromArray(none) // Bytes()
```

An `Array[byte]` converts as it is, and so does an array of any other integer type:

```ant
let raw = Array.map([1, 2, 255], byte)
Bytes.fromArray(raw)                      // Bytes(01 02 ff)
Bytes.fromArray([int8(1), int8(127)])     // Bytes(01 7f)
Bytes.fromArray([uint64(0), uint64(255)]) // Bytes(00 ff)
```

Call it with a pipe:

```ant
[65, 66] | Bytes.fromArray() // Bytes(41 42)
```

The result reads like any `Bytes`: `len` counts the bytes and an index reads one as a `byte`:

```ant
let b = Bytes.fromArray([10, 20, 30])
len(b) // 3
b[2]   // 30
```

## Applications

Check a file header against a known signature:

```ant
let png = Bytes.fromArray([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a])
let file = Bytes.concat([png, String.toBytes("IHDR")])
Bytes.startsWith(file, png) // true
```

Compute bytes one at a time into an array, then convert it once:

```ant
let ramp: Array[int] = []
for i in Iter.range(0, 4) {
  ramp.push(i * 64)
}
Bytes.fromArray(ramp) // Bytes(00 40 80 c0)
```

Encode an integer as four bytes, most significant first:

```ant
func bigEndian(n: int): Bytes {
  Bytes.fromArray([24, 16, 8, 0].map({ Bits.and(Bits.shiftRight(n, x), 0xff) }))
}

bigEndian(0x12345678) // Bytes(12 34 56 78)
```

Decode UTF-8 bytes given as numbers:

```ant
String.fromBytes(Bytes.fromArray([0xe6, 0x97, 0xa5])) // "日"
String.fromBytes(Bytes.fromArray([0xff]))             // null
```

## Properties and relations

[Bytes.toArray](toArray.md) undoes it:

```ant
let b = String.toBytes("ant")
assert(Bytes.fromArray(Bytes.toArray(b)) == b)
assert(Bytes.toArray(Bytes.fromArray([1, 2, 3])) == Array.map([1, 2, 3], byte))
```

The bytes of an ASCII string are its character codes:

```ant
assert(Bytes.fromArray([104, 105]) == String.toBytes("hi"))
```

The length is the length of the array:

```ant
let xs = [7, 8, 9]
assert(len(Bytes.fromArray(xs)) == len(xs))
```

## Possible issues

A value outside `0 ..= 255` panics, a negative one included:

```ant
let low = [int8(-1)]
Bytes.fromArray(low)      // panic: Bytes.fromArray needs integers from 0 to 255, got -1 at index 0
Bytes.fromArray([1, 256]) // panic: Bytes.fromArray needs integers from 0 to 255, got 256 at index 1
```

To keep only the low eight bits of each value, mask them first:

```ant
Bytes.fromArray([300, 1].map({ Bits.and(x, 0xff) })) // Bytes(2c 01)
```

The result is a copy. Pushing to the array afterwards leaves it as it was:

```ant
let xs = [1, 2]
let b = Bytes.fromArray(xs)
xs.push(3)
b // Bytes(01 02)
```

An array literal is not `Bytes`, and floats are not integers. Both are compile errors:

```ant
let header: Bytes = [1, 2] // error: cannot use Array[int] where Bytes is expected
let b = Bytes.fromArray([1.0])
// error: T would be float64, which is not a member of its bound Integer
```

A string escape does not write a byte. `"\u{89}"` is a character, which UTF-8 stores as two bytes;
write the byte with `Bytes.fromArray`:

```ant
String.toBytes("\u{89}") // Bytes(c2 89)
Bytes.fromArray([0x89])  // Bytes(89)
```

## See also

[Bytes.toArray](toArray.md) · [Bytes.concat](concat.md) · [Bytes.startsWith](startsWith.md) ·
[Bytes](index.md) · [String.toBytes](../String/toBytes.md) ·
[Encoding.fromHex](../Encoding/fromHex.md)
