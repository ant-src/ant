<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.fromBytes

```ant signature
func String.fromBytes(bytes: Bytes): string | null
```

String.fromBytes decodes bytes as UTF-8, or returns null when they are not valid UTF-8.

<!-- end generated -->

## Details

- `String.fromBytes` returns the text that `bytes` encode as UTF-8, or `null` when they are not
  valid UTF-8 ([spec: String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)).
- It replaces nothing. Text that reaches a program as a `string` from outside has each maximal
  subpart of an ill-formed sequence replaced by U+FFFD; a program that reads `Bytes` instead uses
  `String.fromBytes` to learn whether they are text at all, and
  [String.fromBytesLossy](fromBytesLossy.md) to decode them the way outside text is decoded.
- UTF-8 does not encode surrogates, so bytes that encode one give `null`; every string holds only
  scalar values ([spec: String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)).
- The result is `string | null`. Narrow it with `!= null`, or give a fallback with `??`
  ([spec: Nullability](https://github.com/ant-src/spec/blob/master/ant-lang.md#7-nullability)).
- `String.fromBytes` need not copy the bytes, since neither the `Bytes` nor the `string` can change
  ([spec: Bytes](https://github.com/ant-src/spec/blob/master/ant-lang.md#bytes)).

## Basic examples

```ant
String.fromBytes(String.toBytes("héllo"))             // "héllo"
String.fromBytes(Bytes.fromArray([0x41, 0x6e, 0x74])) // "Ant"
String.fromBytes(Bytes.fromArray([0xff, 0xfe]))       // null
```

## Scope

Empty bytes decode to the empty string:

```ant
var empty: Array[byte] = []
String.fromBytes(Bytes.fromArray(empty)) // ""
```

Multi-byte code points decode when all their bytes are there:

```ant
String.fromBytes(Bytes.fromArray([0xe2, 0x82, 0xac]))       // "€"
String.fromBytes(Bytes.fromArray([0xf0, 0x9f, 0x90, 0x9c])) // "🐜"
```

A code point cut short is not valid UTF-8, and neither is a byte that continues nothing:

```ant
String.fromBytes(Bytes.fromArray([0xe2, 0x82])) // null
String.fromBytes(Bytes.fromArray([0x82, 0xac])) // null
```

A slice of the bytes of a string decodes only when it starts and ends between code points:

```ant
var b = String.toBytes("€uro")
String.fromBytes(Bytes.slice(b, 0, 2)) // null
String.fromBytes(Bytes.slice(b, 3, 6)) // "uro"
```

An overlong encoding, a surrogate, and a value above U+10FFFF are all rejected:

```ant
String.fromBytes(Bytes.fromArray([0xc0, 0xaf]))             // null
String.fromBytes(Bytes.fromArray([0xed, 0xa0, 0x80]))       // null
String.fromBytes(Bytes.fromArray([0xf4, 0x90, 0x80, 0x80])) // null
```

A zero byte is the code point U+0000. The result holds the code points the bytes encode, so `e`
followed by U+0301 is not the single code point `é`; [String.normalize](normalize.md) brings the two
spellings to one:

```ant
String.fromBytes(Bytes.fromArray([0x61, 0x00, 0x62])) == "a\u{0}b"  // true
String.fromBytes(Bytes.fromArray([0x65, 0xcc, 0x81])) == "e\u{301}" // true
String.fromBytes(Bytes.fromArray([0x65, 0xcc, 0x81])) == "\u{e9}"   // false
```

The parameter is named `bytes`:

```ant
String.fromBytes(bytes: String.toBytes("go")) // "go"
```

In a pipe, the bytes are the first argument:

```ant
"naïve" | String.toBytes() | String.fromBytes() // "naïve"
```

Bytes from [Encoding.fromHex](../Encoding/fromHex.md) decode like any others:

```ant
Encoding.fromHex("48656c6c6f") | Result.unwrap() | String.fromBytes() // "Hello"
```

A function value works wherever a `func(Bytes): string | null` is expected:

```ant
var decode: func(Bytes): string | null = String.fromBytes
decode(Bytes.fromArray([0x6f, 0x6b])) // "ok"
```

## Applications

### Text or binary

Read a file as `Bytes` and report whether it is text or binary:

<!-- test: file notes.txt
café
-->

```ant
func describe(path: string): string {
  var bytes = Result.unwrap(File.readBytes(path))
  var text = String.fromBytes(bytes)
  if text == null {
    return "${path}: binary, ${Bytes.length(bytes)} bytes"
  }
  "${path}: ${String.length(String.trim(text))} characters of text"
}

Result.unwrap(File.writeBytes("logo.png", Bytes.fromArray([0x89, 0x50, 0x4e, 0x47])))
println(describe("notes.txt"))
println(describe("logo.png"))
```

```text
notes.txt: 4 characters of text
logo.png: binary, 4 bytes
```

### Turn null into an error

`String.fromBytes` returns `null`, not a `Result`. Use `??` with `return` to turn the `null` into an
`Err` in a function that returns a `Result`:

```ant
func decodeUtf8(bytes: Bytes): Result[string, Error] {
  var text = String.fromBytes(bytes) ?? return Result.err(Error{ message: "not UTF-8" })
  Result.ok(text)
}

println(decodeUtf8(String.toBytes("ok")))
println(decodeUtf8(Bytes.fromArray([0xff])))
```

```text
Ok{ value: "ok" }
Err{ error: Error{ message: "not UTF-8" } }
```

### Warn before a lossy decode

Keep the text when it decodes; otherwise say so, then fall back to U+FFFD replacement, which prints
as `�`:

```ant
func decode(raw: Bytes): string {
  var text = String.fromBytes(raw)
  if text != null {
    return text
  }
  println("warning: not UTF-8, decoding with U+FFFD")
  String.fromBytesLossy(raw)
}

println(decode(String.toBytes("ab")))
println(decode(Bytes.fromArray([0x61, 0xff, 0x62])))
```

```text
ab
warning: not UTF-8, decoding with U+FFFD
a�b
```

## Properties and relations

`String.fromBytes` undoes [String.toBytes](toBytes.md) for every string:

```ant
for s in ["", "ant", "straße", "🐜🐜", "e\u{301}"] {
  assert(String.fromBytes(String.toBytes(s)) == s)
}
```

When the bytes are valid UTF-8, [String.fromBytesLossy](fromBytesLossy.md) gives the same string;
when they are not, the lossy result holds at least one U+FFFD:

```ant
for b in [String.toBytes("ok"), Bytes.fromArray([0x61, 0xff]), Bytes.fromArray([0xe2, 0x82])] {
  var strict = String.fromBytes(b)
  if strict != null {
    assert(String.fromBytesLossy(b) == strict)
  } else {
    assert(String.contains(String.fromBytesLossy(b), "\u{FFFD}"))
  }
}
```

The bytes of the result are the input bytes:

```ant
var b = Bytes.fromArray([0xc3, 0xa9, 0x21])
var s = String.fromBytes(b)
assert(s != null && String.toBytes(s) == b)
```

## Possible issues

The result may be `null`, so it cannot be used as a `string` before it is narrowed:

<!-- test: error cannot use string | null where string is expected -->

```ant
var text: string = String.fromBytes(String.toBytes("hi"))
```

A byte order mark is valid UTF-8 and stays in the result as U+FEFF:

```ant
String.fromBytes(Bytes.fromArray([0xef, 0xbb, 0xbf, 0x41])) == "\u{FEFF}A" // true
```

`Bytes` has no `fromBytes` function, so a call through `.` on bytes is a compile error. Use a pipe:

<!-- test: error Bytes has no field or function fromBytes -->

```ant
var b = String.toBytes("hi")
println(b.fromBytes())
```

A pipe binds looser than `??`, so a pipeline before `??` needs parentheses:

<!-- test: error parenthesize the pipe -->

```ant
var b = String.toBytes("hi")
println(b | String.fromBytes() ?? "binary")
```

```ant
var b = String.toBytes("hi")
(b | String.fromBytes()) ?? "binary" // "hi"
```

## See also

[String.fromBytesLossy](fromBytesLossy.md), [String.toBytes](toBytes.md),
[String.fromCodePoints](fromCodePoints.md), [Bytes](../Bytes/index.md),
[Bytes.fromArray](../Bytes/fromArray.md), [File.readBytes](../File/readBytes.md),
[Encoding.fromHex](../Encoding/fromHex.md)
