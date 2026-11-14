<!-- status: reviewed -->

# Encode as base64 or hex

Turn bytes into text that survives a URL, a header, or a JSON field, and turn it back.

1. Get the bytes: [String.toBytes](../../lib/String/toBytes.md) for text,
   [File.readBytes](../../lib/File/readBytes.md) for a file.
2. Encode them with [Encoding.toBase64](../../lib/Encoding/toBase64.md) or
   [Encoding.toHex](../../lib/Encoding/toHex.md).
3. Decode with [Encoding.fromBase64](../../lib/Encoding/fromBase64.md) or
   [Encoding.fromHex](../../lib/Encoding/fromHex.md), which return a `Result`, and use
   [`try`](../../language/keywords/try.md) to return the `Err`.
4. Turn decoded bytes back into text with [String.fromBytes](../../lib/String/fromBytes.md), which
   returns `null` when they are not UTF-8.

## Encode text as base64

The encoders take bytes, so convert the string first and pipe the bytes on:

```ant
let encoded = "Hello, Ant!".toBytes() | Encoding.toBase64()
println(encoded)
```

```text
SGVsbG8sIEFudCE=
```

An HTTP Basic authentication header encodes `user:password` this way:

```ant
let auth = "Basic ${"ann:secret".toBytes() | Encoding.toBase64()}"
println(auth)
```

```text
Basic YW5uOnNlY3JldA==
```

The result uses the standard alphabet with `+` and `/`, padded with `=`.

## Encode for a URL or a file name

Pass `url: true` for the URL-safe alphabet, which has `-` and `_` in place of `+` and `/`. These
bytes, made with [Bytes.fromArray](../../lib/Bytes/fromArray.md), show the difference:

```ant
let bytes = Bytes.fromArray([0xfb, 0xff, 0xfe])
println(Encoding.toBase64(bytes))
println(Encoding.toBase64(bytes, url: true))
```

```text
+//+
-__-
```

The output is padded either way. Where a format wants no padding, as a JSON Web Token does, remove
the `=` with [String.replace](../../lib/String/replace.md); it appears nowhere else in base64. The
decoder reads the text with the padding or without it:

```ant
let token = "hi".toBytes() | Encoding.toBase64(url: true) | String.replace("=", "")
println(token)
println(try Encoding.fromBase64(token, url: true))
```

```text
aGk
Bytes(68 69)
```

## Decode base64 back to text

Decode with `Encoding.fromBase64`, then convert the bytes with `String.fromBytes`. Give the `null`
for bytes that are not UTF-8 its own error with [Result.fail](../../lib/Result/fail.md):

```ant
func decodeText(s: string): Result[string, Error] {
  let bytes = try (Encoding.fromBase64(s) | Result.wrapErr("decode ${s}"))
  let text = String.fromBytes(bytes) ?? return Result.fail("decode ${s}: not UTF-8 text")
  Result.ok(text)
}

println(decodeText("SGVsbG8sIEFudCE="))
println(decodeText("Hello!"))
println(decodeText("/w=="))
```

```text
Ok{ value: "Hello, Ant!" }
Err{ error: Error{ message: "decode Hello!: illegal base64 data at input byte 5" } }
Err{ error: Error{ message: "decode /w==: not UTF-8 text" } }
```

The decoder skips line breaks, so base64 wrapped across lines decodes as it is:

```ant
println(try Encoding.fromBase64("SGVsbG8s\nIEFudCE=") | String.fromBytes())
```

```text
Hello, Ant!
```

Decode with the alphabet the text was encoded in. `url: true` text with `-` or `_` is an `Err` for
the standard decoder:

```ant
println(Encoding.fromBase64("-__-"))
println(Encoding.fromBase64("-__-", url: true))
```

```text
Err{ error: Error{ message: "illegal base64 data at input byte 0" } }
Ok{ value: Bytes(fb ff fe) }
```

## Encode as hex

`Encoding.toHex` writes two lower-case digits per byte. Use it to print a digest from
[Crypto.sha256](../../lib/Crypto/sha256.md):

```ant
println("ant".toBytes() | Encoding.toHex())
println("ant".toBytes() | Crypto.sha256() | Encoding.toHex())
```

```text
616e74
67a333356cdc566e6e346b5718447308ec0e25f47e623161fb03962b327a651f
```

## Decode hex

`Encoding.fromHex` reads digits in either case. An odd number of digits or a character that is not a
digit is an `Err`:

```ant
println(Encoding.fromHex("DEADbeef"))
println(Encoding.fromHex("abc"))
println(Encoding.fromHex("0x1f"))
```

```text
Ok{ value: Bytes(de ad be ef) }
Err{ error: Error{ message: "odd length hex string" } }
Err{ error: Error{ message: "invalid byte: U+0078 'x'" } }
```

Remove a `0x` prefix with [String.trimPrefix](../../lib/String/trimPrefix.md) before you decode:

```ant
println(try Encoding.fromHex("0x1f".trimPrefix("0x")))
```

```text
Bytes(1f)
```

## The whole program

This script reads a file as bytes and prints it as a `data:` URL, the form a browser accepts in
place of a link to the file, followed by the hex SHA-256 digest that identifies its content:

<!-- test: file note.txt
Hi, Ant
-->

```ant
let bytes = try (File.readBytes("note.txt") | Result.wrapErr("read note"))
let prefix = "data:text/plain;base64,"
let url = "${prefix}${Encoding.toBase64(bytes)}"
println(url)
println(bytes | Crypto.sha256() | Encoding.toHex())

let back = try Encoding.fromBase64(url.trimPrefix(prefix))
assert(back == bytes)
```

```text
data:text/plain;base64,SGksIEFudAo=
a1c6417c46014cca024afb1900315ab950d1cc0aa6b782b1bc6e0739b333b19b
```

Besides the functions above, it uses [Result.wrapErr](../../lib/Result/wrapErr.md),
[String.trimPrefix](../../lib/String/trimPrefix.md), and [`assert`](../../lib/prelude/assert.md).

## See also

[Encoding.toBase64](../../lib/Encoding/toBase64.md) ·
[Encoding.fromBase64](../../lib/Encoding/fromBase64.md) ·
[Encoding.toHex](../../lib/Encoding/toHex.md) · [Encoding.fromHex](../../lib/Encoding/fromHex.md) ·
[Encoding](../../lib/Encoding/index.md) · [Bytes](../../lib/Bytes/index.md) ·
[String.toBytes](../../lib/String/toBytes.md) · [String.fromBytes](../../lib/String/fromBytes.md) ·
[File.readBytes](../../lib/File/readBytes.md) · [Crypto.sha256](../../lib/Crypto/sha256.md) ·
[Hash data and compare hashes](hash.md) · [Parse and build URLs](urls.md)
