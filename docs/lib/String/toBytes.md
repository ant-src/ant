<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.toBytes

```ant signature
func String.toBytes(s: string): Bytes
```

String.toBytes returns the UTF-8 encoding of s.

<!-- end generated -->

## Details

- Returns the UTF-8 encoding of `s`, and never fails.
- An escape writes a code point, not a byte: `"\u{89}"` encodes as the two bytes `c2 89`.

## Basic examples

```ant
String.toBytes("Ant") // Bytes(41 6e 74)
String.toBytes("é")   // Bytes(c3 a9)
String.toBytes("")    // Bytes()
```

Call it through `.` on a string:

```ant
"hi".toBytes() // Bytes(68 69)
```

## Scope

ASCII takes one byte per code point; other code points take two, three, or four:

```ant
String.toBytes("a") // Bytes(61)
String.toBytes("ñ") // Bytes(c3 b1)
String.toBytes("€") // Bytes(e2 82 ac)
String.toBytes("🐜") // Bytes(f0 9f 90 9c)
```

The code point U+0000 encodes as a zero byte, and the replacement character U+FFFD as its three
bytes:

```ant
String.toBytes("\u{0}")    // Bytes(00)
String.toBytes("\u{FFFD}") // Bytes(ef bf bd)
```

`String.toBytes` encodes the code points as they are and does not normalize. An `e` followed by
U+0301 and the single code point `é` look alike and encode differently:

```ant
String.toBytes("e\u{301}") // Bytes(65 cc 81)
String.toBytes("\u{e9}")   // Bytes(c3 a9)
```

The parameter is named `s`:

```ant
String.toBytes(s: "go") // Bytes(67 6f)
```

In a pipe, the string is the first argument:

```ant
"naïve" | String.toBytes() | Bytes.length() // 6
```

Index the result to read single bytes as `byte` values:

```ant
var b = String.toBytes("é")
b[0]             // 195
b[1]             // 169
Bytes.toArray(b) // [195, 169]
```

A function value works wherever a `func(string): Bytes` is expected:

```ant
var encode: func(string): Bytes = String.toBytes
encode("x") // Bytes(78)
```

## Applications

### Digests and encodings of text

[Crypto](../Crypto/index.md) and [Encoding](../Encoding/index.md) work on `Bytes`. Turn text into
bytes first:

```ant
var digest = Crypto.sha256(String.toBytes("abc")) | Encoding.toHex()
digest // "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
```

```ant
Encoding.toBase64(String.toBytes("hi?")) // "aGk/"
```

### Fit text into a byte limit

Keep as many whole code points as fit into `limit` bytes, so the cut never splits a code point:

```ant
func fitBytes(s: string, limit: int): string {
  var kept: Array[string] = []
  var used = 0
  for c in s {
    var n = Bytes.length(String.toBytes(c))
    if used + n > limit {
      break
    }
    used += n
    kept.push(c)
  }
  String.join(kept, "")
}

println(fitBytes("naïve café", 4))
println(fitBytes("🐜🐜🐜", 9))
```

```text
naï
🐜🐜
```

### A length-prefixed field

Write a name as one length byte followed by its UTF-8 bytes. The prefix counts bytes, not code
points:

```ant
func field(name: string): Bytes {
  var data = String.toBytes(name)
  assert(Bytes.length(data) <= 255)
  Bytes.concat([Bytes.fromArray([Bytes.length(data)]), data])
}

println(field("Zoë"))
```

```text
Bytes(04 5a 6f c3 ab)
```

A one-byte prefix holds at most 255, and `Bytes.fromArray` panics on a longer name, so `field`
checks the length first.

### Write text as bytes

`File.writeBytes` takes `Bytes`, and `File.readBytes` gives back the same bytes:

```ant
var body = String.toBytes("name,city\nAnn,Zürich\n")
Result.unwrap(File.writeBytes("people.csv", body))
println(Result.unwrap(File.readBytes("people.csv")) == body)
```

```text
true
```

## Properties and relations

[String.fromBytes](fromBytes.md) undoes `String.toBytes` for every string:

```ant
for s in ["", "ant", "straße", "🐜🐜", "e\u{301}", "\u{0}"] {
  assert(String.fromBytes(String.toBytes(s)) == s)
}
```

The bytes of a string are the bytes of its characters, joined:

```ant
var s = "Añ🐜"
var parts = s.chars().map(String.toBytes)
assert(Bytes.concat(parts) == String.toBytes(s))
```

The byte length is at least the code point length and at most four times it:

```ant
for s in ["ant", "naïve", "€uro", "🐜"] {
  var n = Bytes.length(String.toBytes(s))
  assert(n >= String.length(s) && n <= 4 * String.length(s))
}
```

Equal strings give equal bytes, and different strings give different bytes:

```ant
assert(String.toBytes("ab") == String.toBytes("ab"))
assert(String.toBytes("ab") != String.toBytes("ba"))
```

## Possible issues

[String.length](length.md) counts code points, not bytes. Use `Bytes.length` of the encoding for the
byte count. The builtin `len` does the same: it counts code points on a string and bytes on `Bytes`:

```ant
String.length("naïve")                // 5
Bytes.length(String.toBytes("naïve")) // 6
len("naïve")                          // 5
len(String.toBytes("naïve"))          // 6
```

A string escape writes a code point, not a byte. To get the byte `0x89`, use
[Bytes.fromArray](../Bytes/fromArray.md):

```ant
String.toBytes("\u{89}") // Bytes(c2 89)
Bytes.fromArray([0x89])  // Bytes(89)
```

`Bytes` and `string` are different types, so comparing them is a compile error:

```ant
println(String.toBytes("a") == "a")
// error: Bytes and string cannot be compared with ==: neither is assignable to the other
```

`Json.stringify` panics on `Bytes`. Write them as base64 or hex instead:

```ant
Json.stringify(String.toBytes("a"))                    // panic: Json.stringify: Bytes has no JSON form; write Encoding.toBase64 of it
Json.stringify(Encoding.toBase64(String.toBytes("a"))) // "\"YQ==\""
```

## See also

[String.fromBytes](fromBytes.md) · [String.fromBytesLossy](fromBytesLossy.md) ·
[String.toCodePoints](toCodePoints.md) · [String.length](length.md) · [Bytes](../Bytes/index.md) ·
[Bytes.length](../Bytes/length.md) · [Encoding.toHex](../Encoding/toHex.md) ·
[Crypto.sha256](../Crypto/sha256.md) · [File.writeBytes](../File/writeBytes.md)
