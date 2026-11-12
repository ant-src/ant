<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Crypto.sha1

```ant signature
func Crypto.sha1(bytes: Bytes): Bytes
```

Crypto.sha1 returns the 20-byte SHA-1 digest of bytes. SHA-1 is broken against collisions and is
here for the formats that still use it, as git and TOTP do.

<!-- end generated -->

## Details

- The digest is `Bytes`, not text. Pipe it to [Encoding.toHex](../Encoding/toHex.md) for the usual
  40-character form, or to [Encoding.toBase64](../Encoding/toBase64.md).
- Text is hashed as its UTF-8 bytes: convert a string with [String.toBytes](../String/toBytes.md)
  first.

## Basic examples

Hash a string's bytes:

```ant
let abc = String.toBytes("abc")
Crypto.sha1(abc) // Bytes(a9 99 3e 36 47 06 81 6a ba 3e 25 71 78 50 c2 6c 9c d0 d8 9d)
```

Write the digest as hex:

```ant
let abc = String.toBytes("abc")
Crypto.sha1(abc) | Encoding.toHex() // "a9993e364706816aba3e25717850c26c9cd0d89d"
```

## Scope

Empty input has a digest too:

```ant
let empty = String.toBytes("")
Crypto.sha1(empty) | Encoding.toHex() // "da39a3ee5e6b4b0d3255bfef95601890afd80709"
```

Bytes that are not text, built from integers:

```ant
let raw = Bytes.fromArray([0xff, 0x00])
Crypto.sha1(raw) | Encoding.toHex() // "da33e41954997c24ad31f6c1845e18510b384f67"
```

The contents of a file, read as bytes:

<!-- test: file hello.txt
hello
-->

```ant
let data = try File.readBytes("hello.txt")
println(Crypto.sha1(data) | Encoding.toHex())
```

```text
f572d396fae9206628714fb2ce00f72e94f2258f
```

## Applications

Compute the object ID git gives a file: the digest of a `blob` header, a zero byte, and the
contents.

```ant
func gitBlobId(content: Bytes): string {
  let header = String.toBytes("blob ${content.length()}")
  Bytes.concat([header, Bytes.fromArray([0]), content]) | Crypto.sha1() | Encoding.toHex()
}

println(gitBlobId(String.toBytes("hello\n")))
```

```text
ce013625030ba8dba906f756967f9e9ca394464a
```

Answer a WebSocket handshake: the server hashes the client's key with a fixed GUID and sends the
digest back in Base64.

```ant
let websocketGuid = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func acceptKey(key: string): string {
  String.toBytes(key + websocketGuid) | Crypto.sha1() | Encoding.toBase64()
}

println(acceptKey("dGhlIHNhbXBsZSBub25jZQ=="))
```

```text
s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```

## Properties and relations

The digest is always 20 bytes, and its hex form 40 characters:

```ant
let short = Crypto.sha1(String.toBytes("a"))
let long = Crypto.sha1(String.toBytes("a".repeat(10000)))
assert(short.length() == 20 && long.length() == 20)
assert(Encoding.toHex(long).length() == 40)
```

Equal bytes give equal digests, and `Bytes` compares by contents:

```ant
let a = String.toBytes("ant")
let b = Bytes.fromArray([0x61, 0x6e, 0x74])
assert(Crypto.sha1(a) == Crypto.sha1(b))
```

## Possible issues

A string is not `Bytes`; convert it first:

```ant
let digest = Crypto.sha1("abc") // error: cannot use string where Bytes is expected
```

A digest does not compare with a hex string. Compare hex with hex, or decode the string with
[Encoding.fromHex](../Encoding/fromHex.md):

```ant
let expected = "a9993e364706816aba3e25717850c26c9cd0d89d"
let ok = Crypto.sha1(String.toBytes("abc")) == expected
// error: Bytes and string cannot be compared with ==: neither is assignable to the other
```

Strings that look the same can hold different bytes. A precomposed `é` and an `e` with a combining
accent hash differently; normalize them first with [String.normalize](../String/normalize.md):

```ant
func digest(s: string): Bytes { Crypto.sha1(String.toBytes(s)) }
let composed = "caf\u{e9}"
let decomposed = "cafe\u{301}"
digest(composed) == digest(decomposed)                                   // false
digest(composed.normalize("NFC")) == digest(decomposed.normalize("NFC")) // true
```

## See also

- [Crypto.sha256](sha256.md), [Crypto.sha512](sha512.md), [Crypto.md5](md5.md): other digests.
- [Crypto.hmacSha1](hmacSha1.md): SHA-1 with a key, as TOTP uses it.
- [Crypto.equal](equal.md): compare a digest with a received one.
- [Encoding.toHex](../Encoding/toHex.md), [Encoding.toBase64](../Encoding/toBase64.md): write a
  digest as text.
