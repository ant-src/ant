<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Encoding

Encoding converts bytes to and from text in base64 and hex.

<!-- end generated -->

## Details

- Encoding works on [Bytes](../Bytes/index.md). Text goes through
  [String.toBytes](../String/toBytes.md) before encoding and through
  [String.fromBytes](../String/fromBytes.md) after decoding.
- A decoder returns `Err` on text that is not valid in its format.
- Hex writes two lower-case digits per byte and reads either case. Base64 writes four characters per
  three bytes, padded with `=`, and reads with or without the padding.
- Base64 has two alphabets: the standard one with `+` and `/`, and with `url: true` the URL-safe one
  with `-` and `_`. A decoder reads only the alphabet it is given.

## Basic examples

Encode bytes as hex and base64:

```ant
let data = String.toBytes("hi?")
Encoding.toHex(data)               // "68693f"
Encoding.toBase64(data)            // "aGk/"
Encoding.toBase64(data, url: true) // "aGk_"
```

Decode them back into bytes:

```ant
Encoding.fromHex("68693f")  // Ok{ value: Bytes(68 69 3f) }
Encoding.fromBase64("aGk/") // Ok{ value: Bytes(68 69 3f) }
Encoding.fromHex("6g")      // Err{ error: Error{ message: "invalid byte: U+0067 'g'", kind: null } }
```

## Scope

Empty bytes encode as the empty string, and the empty string decodes as empty bytes:

```ant
let empty = Bytes.new()
Encoding.toHex(empty)    // ""
Encoding.toBase64(empty) // ""
Encoding.fromHex("")     // Ok{ value: Bytes() }
Encoding.fromBase64("")  // Ok{ value: Bytes() }
```

The bytes need not be text:

```ant
let png = Bytes.fromArray([0x89, 0x50, 0x4e, 0x47])
Encoding.toHex(png)    // "89504e47"
Encoding.toBase64(png) // "iVBORw=="
```

Hex is twice the length of the bytes; base64 is four characters for every three bytes, rounded up:

```ant
let data = Bytes.new(10)
Encoding.toHex(data).length()    // 20
Encoding.toBase64(data).length() // 16
```

The bytes are the first argument, so an encoder ends a pipe and a decoder starts one:

```ant
"abc".toBytes() | Crypto.md5() | Encoding.toHex()             // "900150983cd24fb0d6963f7d28e17f72"
"QW50" | Encoding.fromBase64() | Result.map(String.fromBytes) // Ok{ value: "Ant" }
```

The decoders read hex of either case, base64 without its padding, and base64 broken into lines.
Base64 whose length no bytes encode to is `Err`:

```ant
Encoding.fromHex("6A3F")      // Ok{ value: Bytes(6a 3f) }
Encoding.fromBase64("aGk")    // Ok{ value: Bytes(68 69) }
Encoding.fromBase64("aGk\n/") // Ok{ value: Bytes(68 69 3f) }
Encoding.fromBase64("abcde")  // Err{ error: Error{ message: "illegal base64 data at input byte 4", kind: null } }
```

## Functions by job

### Hex

- [Encoding.toHex](toHex.md): bytes as two lower-case hex digits each
- [Encoding.fromHex](fromHex.md): hex digits of either case back into bytes; `Err` on an odd length
  or a character that is not a hex digit

### Base64

- [Encoding.toBase64](toBase64.md): bytes as padded base64, in the standard or the URL-safe alphabet
- [Encoding.fromBase64](fromBase64.md): base64 back into bytes, with or without padding, skipping
  `\r` and `\n`

## Applications

Show a digest as hex, the form checksums are published in:

```ant
let digest = "hello".toBytes() | Crypto.sha256() | Encoding.toHex()
println(digest)
```

```text
2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
```

Make a random token that is safe in a URL, a file name, or a cookie:

```ant
let token = Crypto.randomBytes(18) | Encoding.toBase64(url: true)
token.length() // 24
```

18 bytes is a multiple of three, so the token has no `=` padding. The URL-safe alphabet still pads,
and `=` needs escaping in a query string.

Write the value of an HTTP Basic authorization header, which carries `user:password` in base64:

```ant
func basicAuth(user: string, password: string): string {
  let credentials = "${user}:${password}".toBytes() | Encoding.toBase64()
  "Basic ${credentials}"
}

println(basicAuth("ant", "s3cret"))
```

```text
Basic YW50OnMzY3JldA==
```

Read the credentials back out of that header. A decoder's `Err` goes up through `try`, with context
added by `Result.wrapErr`:

```ant
func parseBasicAuth(header: string): Result[string, Error] {
  if !header.startsWith("Basic ") {
    return Result.fail("not Basic authorization")
  }
  let encoded = header.trimPrefix("Basic ")
  let raw = try (Encoding.fromBase64(encoded) | Result.wrapErr("bad credentials"))
  let credentials = String.fromBytes(raw) ?? return Result.fail("credentials are not UTF-8")
  Result.ok(credentials)
}

println(parseBasicAuth("Basic YW50OnMzY3JldA=="))
println(parseBasicAuth("Basic YW50OnMzY3JldA!"))
```

```text
Ok{ value: "ant:s3cret" }
Err{ error: Error{ message: "bad credentials: illegal base64 data at input byte 14", kind: null } }
```

Put binary data in JSON, which has no form for `Bytes`:

```ant
type Attachment = { name: string, data: string }

let attachment = Attachment{
  name: "logo.png",
  data: Bytes.fromArray([0x89, 0x50, 0x4e, 0x47]) | Encoding.toBase64(),
}
println(Json.stringify(attachment))
```

```text
{"name":"logo.png","data":"iVBORw=="}
```

## Properties and relations

Each decoder gives back the bytes its encoder took:

```ant
let samples = [Bytes.new(), "ant".toBytes(), "🐜".toBytes(), Bytes.fromArray([0, 255, 251])]
for b in samples {
  assert((b | Encoding.toHex() | Encoding.fromHex() | Result.unwrap()) == b)
  assert((b | Encoding.toBase64() | Encoding.fromBase64() | Result.unwrap()) == b)
  assert((b | Encoding.toBase64(url: true) | Encoding.fromBase64(url: true) | Result.unwrap()) == b)
}
```

The two base64 alphabets differ only in `+` and `/`, which the URL-safe one writes as `-` and `_`:

```ant
let b = Bytes.fromArray([0xfb, 0xff, 0xbf])
let standard = Encoding.toBase64(b)
assert(Encoding.toBase64(b, url: true) == standard.replace("+", "-").replace("/", "_"))
```

## Possible issues

A string is not `Bytes`; convert it first:

```ant
let hex = Encoding.toHex("hi") // error: cannot use string where Bytes is expected
```

The functions are in `Encoding`, not `Bytes`, so a dot call does not find them. Use a pipe:

```ant
let b = "hi".toBytes()
let dotted = b.toHex() // error: Bytes has no field or function toHex
let piped = b | Encoding.toHex()
```

A decoder returns a `Result`. Unwrap it, or use `try` in a function that returns a `Result`:

```ant
let b: Bytes = Encoding.fromBase64("aGk=")
// error: cannot use Result[Bytes, Error] where Bytes is expected
```

A decoder reads one alphabet. URL-safe base64 with `-` or `_` is `Err` read as standard:

```ant
let encoded = Bytes.fromArray([0xfb, 0xff]) | Encoding.toBase64(url: true)
Encoding.fromBase64(encoded).isOk()            // false
Encoding.fromBase64(encoded, url: true).isOk() // true
```

Base64 and hex hide nothing. Anyone can decode them; they are not encryption.

## See also

- [Bytes](../Bytes/index.md): binary data, the input of every encoder
- [String.toBytes](../String/toBytes.md), [String.fromBytes](../String/fromBytes.md): between
  strings and bytes
- [Crypto](../Crypto/index.md): digests and random bytes, which are shown as hex or base64
- [Url.escape](../Url/escape.md): escaping text for a query string, not encoding bytes
