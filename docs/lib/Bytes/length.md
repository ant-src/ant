<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.length

```ant signature
func Bytes.length(b: Bytes): int
```

Bytes.length returns the number of bytes in b.

<!-- end generated -->

## Details

- Counts bytes, not characters. Empty `Bytes` has length `0`.
- `len(b)` gives the same count.

## Basic examples

Count the bytes of a string in UTF-8:

```ant
Bytes.length(String.toBytes("ant")) // 3
```

Count the bytes of a constant built from integers:

```ant
let png = Bytes.fromArray([0x89, 0x50, 0x4e, 0x47])
Bytes.length(png) // 4
```

Empty `Bytes` has length `0`:

```ant
Bytes.length(String.toBytes("")) // 0
Bytes.length(Bytes.concat([]))   // 0
```

## Scope

A character outside ASCII takes two to four bytes in UTF-8:

```ant
Bytes.length(String.toBytes("é"))      // 2
Bytes.length(String.toBytes("日本語")) // 9
Bytes.length(String.toBytes("🐜"))     // 4
```

A string escape writes a code point, not a byte, so `"\u{89}"` is two bytes:

```ant
Bytes.length(String.toBytes("\u{89}")) // 2
Bytes.length(Bytes.fromArray([0x89]))  // 1
```

Decoded bytes have one byte per two hexadecimal digits, and three per four Base64 characters:

```ant
Bytes.length(Encoding.fromHex("cafe").unwrap())    // 2
Bytes.length(Encoding.fromBase64("YW50").unwrap()) // 3
```

A slice has as many bytes as its range is wide:

```ant
let b = String.toBytes("anthill")
Bytes.length(Bytes.slice(b, 3, 7)) // 4
Bytes.length(Bytes.slice(b, 3, 3)) // 0
```

Call it through `.`, in a pipe, or pass it as a function value:

```ant
let b = String.toBytes("anthill")
b.length()                                                    // 7
b | Bytes.length()                                            // 7
[String.toBytes("a"), String.toBytes("bc")].map(Bytes.length) // [1, 2]
```

## Applications

Check that a payload fits a limit on bytes, such as a field stored in 16 bytes:

```ant
func fits(name: string): bool {
  Bytes.length(String.toBytes(name)) <= 16
}

fits("anthill")             // true
fits("муравейник")          // false
String.length("муравейник") // 10
```

Take the last four bytes, such as a checksum stored at the end of a block:

```ant
func trailer(b: Bytes): Bytes {
  let n = Bytes.length(b)
  Bytes.slice(b, n - 4, n)
}

trailer(Bytes.fromArray([1, 2, 3, 4, 5, 6])) // Bytes(03 04 05 06)
```

Check the length before reading a byte at a fixed offset:

```ant
func version(b: Bytes): int | null {
  if Bytes.length(b) < 5 {
    return null
  }
  int(b[4])
}

version(Bytes.fromArray([0x89, 0x50, 0x4e, 0x47, 2, 0])) // 2
version(Bytes.fromArray([0x89, 0x50]))                   // null
```

Add up the sizes of several pieces before joining them:

```ant
let parts = ["ant", "hill", "🐜"].map(String.toBytes)
parts.map(Bytes.length).sum()     // 11
Bytes.length(Bytes.concat(parts)) // 11
```

## Properties and relations

`len` and `Bytes.length` give the same count:

```ant
for s in ["", "ant", "日本語"] {
  let b = String.toBytes(s)
  assert(len(b) == Bytes.length(b))
}
```

The length of a concatenation is the sum of the lengths:

```ant
let a = String.toBytes("ant")
let b = Bytes.fromArray([0, 255])
assert(Bytes.length(Bytes.concat([a, b])) == Bytes.length(a) + Bytes.length(b))
```

A slice from `from` to `to` has `to - from` bytes:

```ant
let b = String.toBytes("anthill")
for r in [[0, 0], [1, 4], [0, 7]] {
  assert(Bytes.length(Bytes.slice(b, r[0], r[1])) == r[1] - r[0])
}
```

`Bytes.toArray` and `Bytes.fromArray` keep the length:

```ant
let b = String.toBytes("héllo")
assert(len(Bytes.toArray(b)) == Bytes.length(b))
assert(Bytes.length(Bytes.fromArray([1, 2, 3])) == 3)
```

A string has at least as many bytes as code points:

```ant
for s in ["ant", "héllo", "🐜"] {
  assert(Bytes.length(String.toBytes(s)) >= String.length(s))
}
```

## Possible issues

The byte count of a string is not its length in characters. `len` on a string counts code points
too, not bytes. Convert with `String.toBytes` first:

```ant
let s = "日本語"
String.length(s)                // 3
len(s)                          // 3
Bytes.length(String.toBytes(s)) // 9
len(String.toBytes(s))          // 9
```

The length is one past the last index, so indexing at it panics:

```ant
let b = String.toBytes("ant")
b[Bytes.length(b) - 1] // 116
b[Bytes.length(b)]     // panic: index 3 out of range for Bytes of length 3
```

A string is not `Bytes`, so passing one is a compile error. Convert it with `String.toBytes`:

```ant
println(Bytes.length("ant")) // error: cannot use string where Bytes is expected
```

Arithmetic on the length can go below zero. Taking the last four bytes of shorter data panics, so
check `Bytes.length(b) >= 4` first:

```ant
func trailer(b: Bytes): Bytes {
  let n = Bytes.length(b)
  Bytes.slice(b, n - 4, n)
}

trailer(Bytes.fromArray([1, 2])) // panic: Bytes.slice: the range -2..2 is not within 0..2
```

## See also

[Bytes.slice](slice.md) · [Bytes.concat](concat.md) · [Bytes.toArray](toArray.md) ·
[Bytes.fromArray](fromArray.md) · [String.toBytes](../String/toBytes.md) ·
[String.length](../String/length.md) · [Array.length](../Array/length.md) · [Bytes](index.md)
