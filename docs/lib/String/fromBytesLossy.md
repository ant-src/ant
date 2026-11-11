<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.fromBytesLossy

```ant signature
func String.fromBytesLossy(bytes: Bytes): string
```

String.fromBytesLossy decodes bytes as UTF-8, replacing each maximal subpart of an ill-formed
sequence with U+FFFD. A byte order mark at the start is kept, as U+FEFF.

<!-- end generated -->

## Details

- Each maximal subpart of an ill-formed sequence becomes one U+FFFD: the bytes `FF FE` become two,
  and `E2 82`, a euro sign cut short, one.
- It never fails.
- It decodes the way Ant decodes text from outside the program: a file read with `File.read`,
  standard input, a network peer, a child process.
- To learn whether the bytes are text at all, use [String.fromBytes](fromBytes.md), which returns
  `null` instead.

## Basic examples

Valid UTF-8 decodes to the same text `String.fromBytes` gives:

```ant
String.fromBytesLossy(String.toBytes("héllo")) // "héllo"
```

A byte that cannot start or continue a character becomes U+FFFD, which renders as `�`:

```ant
String.fromBytesLossy(Bytes.fromArray([0x61, 0xFF, 0x62])) // "a�b"
```

## Scope

Empty bytes give the empty string:

```ant
String.fromBytesLossy(String.toBytes("")) // ""
```

Each maximal subpart becomes one U+FFFD. Two bytes that are each invalid on their own give two, and
a sequence that starts well and is cut short gives one:

```ant
let twoBad = String.fromBytesLossy(Bytes.fromArray([0xFF, 0xFE]))
twoBad          // "��"
twoBad.length() // 2
let cutShort = String.fromBytesLossy(Bytes.fromArray([0xE2, 0x82]))
cutShort.length() // 1
```

The bytes after a broken sequence still decode. Here `E2 82` begins a euro sign that never ends, and
`41` is the letter `A`:

```ant
String.fromBytesLossy(Bytes.fromArray([0xE2, 0x82, 0x41])) // "�A"
```

Bytes that look like UTF-8 but encode no scalar value are ill-formed too: an overlong encoding of
`/`, an encoded surrogate, and a value above U+10FFFF. None of them reaches the string:

```ant
String.fromBytesLossy(Bytes.fromArray([0xC0, 0xAF]))             // "��"
String.fromBytesLossy(Bytes.fromArray([0xED, 0xA0, 0x80]))       // "���"
String.fromBytesLossy(Bytes.fromArray([0xF4, 0x90, 0x80, 0x80])) // "����"
```

Characters outside the Basic Multilingual Plane decode whole when all four bytes are there, and give
one U+FFFD when the last is missing:

```ant
String.fromBytesLossy(Bytes.fromArray([0xF0, 0x9F, 0x90, 0x9C])) // "🐜"
String.fromBytesLossy(Bytes.fromArray([0xF0, 0x9F, 0x90]))       // "�"
```

The zero byte is valid UTF-8, the code point U+0000, and stays:

```ant
String.fromBytesLossy(Bytes.fromArray([0x00])) == "\u{0}" // true
```

The parameter is named `bytes`:

```ant
String.fromBytesLossy(bytes: String.toBytes("ant")) // "ant"
```

In a pipe, the bytes are the first argument:

```ant
"Ant 🐜" | String.toBytes() | Bytes.slice(0, 6) | String.fromBytesLossy() // "Ant �"
```

A function value works wherever a `func(Bytes): string` is expected:

```ant
let decode: func(Bytes): string = String.fromBytesLossy
[Bytes.fromArray([0x6f, 0x6b]), Bytes.fromArray([0x80])].map(decode) // ["ok", "�"]
```

## Applications

Show the first bytes of a buffer as text, even when the cut lands inside a character:

```ant
func preview(b: Bytes, max: int): string {
  if len(b) <= max {
    return String.fromBytesLossy(b)
  }
  "${String.fromBytesLossy(Bytes.slice(b, 0, max))}…"
}

println(preview(String.toBytes("naïve café"), 3))
println(preview(String.toBytes("naïve café"), 4))
```

```text
na�…
naï…
```

Read a file as bytes, use the bytes, and still log it as text. Here the file holds the Latin-1 byte
`E9` for `é`, which is not UTF-8:

<!-- test: before
Result.unwrap(File.writeBytes("name.txt", Bytes.fromArray([0x4a, 0x6f, 0x73, 0xe9])))
-->

```ant
let bytes = Result.unwrap(File.readBytes("name.txt"))
println("${len(bytes)} bytes: ${String.fromBytesLossy(bytes)}")
```

```text
4 bytes: Jos�
```

Count how many U+FFFD a decode put in, to report how damaged the input was:

```ant
func damage(b: Bytes): int {
  String.count(String.fromBytesLossy(b), "\u{FFFD}")
}

damage(String.toBytes("fine"))                          // 0
damage(Bytes.fromArray([0x61, 0xFF, 0xE2, 0x82, 0x62])) // 2
```

## Properties and relations

On valid UTF-8, `String.fromBytesLossy` agrees with `String.fromBytes` and undoes `String.toBytes`:

```ant
for s in ["", "ant", "héllo", "🐜 \u{0} \u{10FFFF}"] {
  let b = String.toBytes(s)
  assert(String.fromBytesLossy(b) == s)
  assert(String.fromBytes(b) == s)
}
```

When `String.fromBytes` returns `null`, the lossy result holds at least one U+FFFD that the bytes
did not encode:

```ant
let b = Bytes.fromArray([0x61, 0xFF])
assert(String.fromBytes(b) == null)
assert(String.fromBytesLossy(b) == "a\u{FFFD}")
```

The result is always valid text, so encoding it and decoding again changes nothing:

```ant
let b = Bytes.fromArray([0xC0, 0xAF, 0x61, 0xED, 0xA0, 0x80])
let s = String.fromBytesLossy(b)
assert(String.fromBytes(String.toBytes(s)) == s)
assert(String.fromBytesLossy(String.toBytes(s)) == s)
```

## Possible issues

Decoding is lossy: the bytes that became U+FFFD are gone, and encoding the string again does not
give them back. Keep the `Bytes` when you need them as they were:

```ant
let b = Bytes.fromArray([0xFF])
String.toBytes(String.fromBytesLossy(b))      // Bytes(ef bf bd)
String.toBytes(String.fromBytesLossy(b)) == b // false
```

A U+FFFD in the result does not prove the bytes were broken: valid UTF-8 can encode U+FFFD itself.
Use `String.fromBytes` to tell the two apart:

```ant
let real = String.toBytes("\u{FFFD}")
String.fromBytesLossy(real) == String.fromBytesLossy(Bytes.fromArray([0xFF])) // true
String.fromBytes(real) == null                                                // false
```

The function is in the `String` namespace and its argument is `Bytes`, so a call through `.` on the
bytes is a compile error:

```ant
let b = String.toBytes("ant")
println(b.fromBytesLossy()) // error: Bytes has no field or function fromBytesLossy
```

Write `String.fromBytesLossy(b)`, or pipe the bytes into it: `b | String.fromBytesLossy()`.

A leading byte order mark stays in the result, so the string does not equal the same text without
it:

```ant
let s = String.fromBytesLossy(Bytes.fromArray([0xEF, 0xBB, 0xBF, 0x61]))
s == "a"   // false
s.length() // 2
```

The byte order mark is valid UTF-8 and decodes to U+FEFF. `String.trimPrefix(s, "\u{FEFF}")` removes
it.

## See also

[String.fromBytes](fromBytes.md) · [String.toBytes](toBytes.md) · [Bytes](../Bytes/index.md) ·
[Bytes.fromArray](../Bytes/fromArray.md) · [File.readBytes](../File/readBytes.md) ·
[String.fromCodePoints](fromCodePoints.md)
