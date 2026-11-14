<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.startsWith

```ant signature
func Bytes.startsWith(b: Bytes, prefix: Bytes, at: int = 0): bool
```

Bytes.startsWith reports whether prefix occurs in b at index at. It panics if at is outside 0 ..=
Bytes.length(b).

<!-- end generated -->

## Details

- Every `Bytes` value starts with the empty bytes.
- The match is byte by byte. It knows nothing of text, so a prefix may end in the middle of a UTF-8
  character.

## Basic examples

```ant
let png = Bytes.fromArray([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a])
Bytes.startsWith(png, Bytes.fromArray([0x89, 0x50, 0x4e, 0x47])) // true
Bytes.startsWith(png, Bytes.fromArray([0xff, 0xd8]))             // false
```

Call it through `.` on any `Bytes` value:

```ant
let data = String.toBytes("GIF89a")
data.startsWith(String.toBytes("GIF"))   // true
data.startsWith(String.toBytes("GIF87")) // false
```

## Scope

Every value starts with the empty bytes, and the empty bytes start with nothing else:

```ant
let empty = String.toBytes("")
String.toBytes("ant").startsWith(empty) // true
empty.startsWith(empty)                 // true
empty.startsWith(String.toBytes("a"))   // false
```

A prefix longer than the bytes never matches, and a value starts with itself:

```ant
let ant = String.toBytes("ant")
ant.startsWith(String.toBytes("anthill")) // false
ant.startsWith(ant)                       // true
```

Any byte value matches, zero and `0xff` included:

```ant
let b = Bytes.fromArray([0, 0xff, 7])
b.startsWith(Bytes.fromArray([0]))       // true
b.startsWith(Bytes.fromArray([0, 0xff])) // true
b.startsWith(Bytes.fromArray([0xff]))    // false
```

In a pipeline, the bytes are the first argument:

```ant
String.toBytes("PK\u{3}\u{4}") | Bytes.startsWith(String.toBytes("PK")) // true
```

Naming the arguments makes clear which value is the prefix, in either order:

```ant
let header = String.toBytes("%PDF-1.7")
Bytes.startsWith(b: header, prefix: String.toBytes("%PDF")) // true
Bytes.startsWith(prefix: String.toBytes("%PDF"), b: header) // true
```

On bytes that can be null, use `?.` and supply a default with `??`:

```ant
let body: Bytes | null = null
body?.startsWith(String.toBytes("{")) ?? false // false
```

## Applications

Guess a file type from its magic number:

```ant
func kind(data: Bytes): string {
  if data.startsWith(Bytes.fromArray([0x89, 0x50, 0x4e, 0x47])) {
    "png"
  } else if data.startsWith(Bytes.fromArray([0xff, 0xd8, 0xff])) {
    "jpeg"
  } else if data.startsWith(String.toBytes("%PDF")) {
    "pdf"
  } else {
    "unknown"
  }
}

println(kind(Bytes.fromArray([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a])))
println(kind(Bytes.fromArray([0xff, 0xd8, 0xff, 0xe0])))
println(kind(String.toBytes("%PDF-1.7")))
println(kind(String.toBytes("hello")))
```

```text
png
jpeg
pdf
unknown
```

Drop the UTF-8 byte order mark that some editors write at the start of a file:

```ant
let bom = Bytes.fromArray([0xef, 0xbb, 0xbf])

func withoutBom(data: Bytes): Bytes {
  if data.startsWith(bom) {
    data.slice(bom.length(), data.length())
  } else {
    data
  }
}

let raw = Bytes.concat([bom, String.toBytes("name,age")])
println(String.fromBytes(withoutBom(raw)))
println(String.fromBytes(withoutBom(String.toBytes("name,age"))))
```

```text
name,age
name,age
```

## Properties and relations

`Bytes.startsWith` agrees with slicing the first bytes of `b`:

```ant
let b = String.toBytes("anthill")
for t in ["ant", "", "anthill", "anthills", "hill"] {
  let prefix = String.toBytes(t)
  let m = prefix.length()
  assert(b.startsWith(prefix) == (m <= b.length() && b.slice(0, m) == prefix))
}
```

It holds exactly when [`Bytes.indexOf`](indexOf.md) finds the prefix at index 0:

```ant
let b = String.toBytes("banana")
for t in ["ban", "an", "", "x"] {
  let prefix = String.toBytes(t)
  assert(b.startsWith(prefix) == (b.indexOf(prefix) == 0))
}
```

Bytes always start with what is joined in front of them with [`Bytes.concat`](concat.md):

```ant
let head = Bytes.fromArray([1, 2])
let tail = Bytes.fromArray([3])
assert(Bytes.concat([head, tail]).startsWith(head))
```

On UTF-8 text, it gives the answer [`String.startsWith`](../String/startsWith.md) gives:

```ant
let s = "año"
for t in ["a", "añ", "ñ", ""] {
  assert(Bytes.startsWith(String.toBytes(s), String.toBytes(t)) == s.startsWith(t))
}
```

## Possible issues

The bytes come first and the prefix second. Swapped, the call asks whether the prefix begins with
the bytes:

```ant
let data = String.toBytes("GIF89a")
let magic = String.toBytes("GIF")
Bytes.startsWith(magic, data) // false
Bytes.startsWith(data, magic) // true
```

The prefix must be `Bytes`. Convert a string with `String.toBytes`, and integers with
[`Bytes.fromArray`](fromArray.md):

```ant
let data = String.toBytes("GIF89a")
println(data.startsWith("GIF")) // error: cannot use string where Bytes is expected
```

```ant
let data = String.toBytes("GIF89a")
println(data.startsWith([0x47])) // error: cannot use Array[int] where Bytes is expected
```

`Encoding.fromHex` returns a `Result`, not `Bytes`. Unwrap it first:

```ant
let data = Bytes.fromArray([0xef, 0xbb, 0xbf, 0x61])
println(data.startsWith(Encoding.fromHex("efbbbf")))
// error: cannot use Result[Bytes, Error] where Bytes is expected
```

```ant
let data = Bytes.fromArray([0xef, 0xbb, 0xbf, 0x61])
data.startsWith(Encoding.fromHex("efbbbf").unwrap()) // true
```

The match works on bytes, not characters. "é" is two bytes in UTF-8, so its first byte alone is a
prefix of the encoded text, though no string ends halfway through a character:

```ant
let word = String.toBytes("école")
word.startsWith(Bytes.fromArray([0xc3])) // true
```

The same text in two spellings is two byte sequences. The precomposed "é" and "e" followed by a
combining accent do not match:

```ant
let word = String.toBytes("e\u{301}cole")
word.startsWith(String.toBytes("é"))                          // false
word.startsWith(String.toBytes(String.normalize("é", "NFD"))) // true
```

## See also

[Bytes.endsWith](endsWith.md) · [Bytes.indexOf](indexOf.md) · [Bytes.slice](slice.md) ·
[Bytes.concat](concat.md) · [String.startsWith](../String/startsWith.md)
