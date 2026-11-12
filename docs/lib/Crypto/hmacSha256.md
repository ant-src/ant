<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Crypto.hmacSha256

```ant signature
func Crypto.hmacSha256(key: Bytes, message: Bytes): Bytes
```

Crypto.hmacSha256 returns the 32-byte HMAC-SHA-256 of message under key. Compare it with a received
signature using [`Crypto.equal`](equal.md).

<!-- end generated -->

## Details

- Both arguments are `Bytes`. Convert text with [`String.toBytes`](../String/toBytes.md) and a hex
  or base64 secret with [`Encoding.fromHex`](../Encoding/fromHex.md) or
  [`Encoding.fromBase64`](../Encoding/fromBase64.md).

## Basic examples

Compute the MAC of a message under a key, and show it as hex:

```ant
let mac = Crypto.hmacSha256("Jefe".toBytes(), "what do ya want for nothing?".toBytes())
Encoding.toHex(mac) // "5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843"
mac.length()        // 32
```

Base64 is the other common text form of a MAC:

```ant
let mac = Crypto.hmacSha256("secret".toBytes(), "hello".toBytes())
Encoding.toBase64(mac) // "iKqz7ejTrflNJquQ07r9SiCDBww7zOnAFO4EpEOEfAs="
```

## Scope

An empty key and an empty message still give 32 bytes:

```ant
let empty = "".toBytes()
let mac = Crypto.hmacSha256(empty, empty)
mac.length()        // 32
Encoding.toHex(mac) // "b613679a0814d9ec772f95d778c35fc5ff1697c493715653c6c712144292c5ad"
```

A binary key, written in hex:

```ant
let key = Encoding.fromHex("0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b").unwrap()
let mac = Crypto.hmacSha256(key, "Hi There".toBytes())
Encoding.toHex(mac) // "b0344c61d8db38535ca8afceaf0bf12b881dc200c9833da726e9376c2e32cff7"
```

A key of any length works, here 131 bytes built from an array:

```ant
let key = Bytes.fromArray(Array.filled(131, 0xaa))
let message = "Test Using Larger Than Block-Size Key - Hash Key First".toBytes()
let mac = Crypto.hmacSha256(key, message)
Encoding.toHex(mac) // "60e431591ee0b67f0d8a26aacbf5b77f8e0bc6213728c5140546040f0ee37f54"
```

Text in any script is hashed as its UTF-8 bytes:

```ant
let mac = Crypto.hmacSha256("ключ".toBytes(), "héllo 🐜".toBytes())
Encoding.toHex(mac) // "cfbede78711f9ce8f95bd139b0f16a28f89faadd3acb33462a6bfb15f06a6044"
```

Since the key comes first, a pipe chains MACs where each result keys the next one. This derives an
AWS Signature Version 4 signing key:

```ant
let secret = "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY"
let signingKey = "AWS4${secret}".toBytes()
  | Crypto.hmacSha256("20120215".toBytes())
  | Crypto.hmacSha256("us-east-1".toBytes())
  | Crypto.hmacSha256("iam".toBytes())
  | Crypto.hmacSha256("aws4_request".toBytes())
println(Encoding.toHex(signingKey))
```

```text
f4780e2d9f65fa895f9c67b32ce1baf0b0d8a43505a000a1a9e090d414db404d
```

## Applications

Check a webhook delivery whose header carries `sha256=` and the hex MAC of the body:

```ant
func verifyWebhook(secret: string, body: Bytes, header: string): bool {
  if !header.startsWith("sha256=") { return false }
  let received = Encoding.fromHex(header.trimPrefix("sha256="))
  if received is Err { return false }
  let expected = Crypto.hmacSha256(secret.toBytes(), body)
  Crypto.equal(received.value, expected)
}

let body = "{\"action\":\"opened\"}".toBytes()
let header = "sha256=2b063a104e1fcedc86cef600a6c042707bac7d6c96f5e864b5e78e966f07aacd"
println(verifyWebhook("hook-secret", body, header))
println(verifyWebhook("wrong-secret", body, header))
println(verifyWebhook("hook-secret", body, "sha256=zz"))
```

```text
true
false
false
```

Sign a cookie value so that a client cannot change it, and read it back only if the MAC matches. The
URL-safe base64 MAC has no dot, so the last dot separates it from a value that may have dots:

```ant
let key = "cookie-secret".toBytes()

func sign(value: string): string {
  let mac = Crypto.hmacSha256(key, value.toBytes())
  "${value}.${Encoding.toBase64(mac, url: true)}"
}

func verify(cookie: string): string | null {
  let dot = cookie.lastIndexOf(".") ?? return null
  let value = cookie.slice(0, dot)
  let mac = cookie.slice(dot + 1, cookie.length())
  let received = Encoding.fromBase64(mac, url: true).unwrapOr("".toBytes())
  let expected = Crypto.hmacSha256(key, value.toBytes())
  if Crypto.equal(received, expected) { value } else { null }
}

let cookie = sign("user=ann")
println(cookie)
println(verify(cookie))
println(verify(cookie.replace("ann", "bob")))
println(verify(sign("v=1.2")))
```

```text
user=ann.dPbLglzqr8Jtf5P0iqcLWKVgiSaCmY3sA2Fw1nSaLd0=
user=ann
null
v=1.2
```

## Properties and relations

The MAC is 32 bytes whatever the length of the message:

```ant
let key = "k".toBytes()
for message in ["", "a", "a much longer message than the key"] {
  assert(Crypto.hmacSha256(key, message.toBytes()).length() == 32)
}
```

The same key and message always give the same MAC, and another key gives another one:

```ant
let message = "hello".toBytes()
let mac = Crypto.hmacSha256("k1".toBytes(), message)
assert(Crypto.hmacSha256("k1".toBytes(), message) == mac)
assert(Crypto.hmacSha256("k2".toBytes(), message) != mac)
```

An HMAC is not the SHA-256 digest of the key and message joined:

```ant
let key = "k".toBytes()
let message = "m".toBytes()
let joined = Crypto.sha256(Bytes.concat([key, message]))
assert(Crypto.hmacSha256(key, message) != joined)
```

For a key of at most 64 bytes, it is SHA-256 applied twice, with the key padded and mixed with two
constants:

```ant
func hmac(key: Bytes, message: Bytes): Bytes {
  let block = Bytes.concat([key, Bytes.fromArray(Array.filled(64 - len(key), 0))]).toArray()
  let inner = Bytes.fromArray(block.map({ Bits.xor(x, byte(0x36)) }))
  let outer = Bytes.fromArray(block.map({ Bits.xor(x, byte(0x5c)) }))
  let innerHash = Crypto.sha256(Bytes.concat([inner, message]))
  Crypto.sha256(Bytes.concat([outer, innerHash]))
}

let key = "secret".toBytes()
let message = "hello".toBytes()
assert(hmac(key, message) == Crypto.hmacSha256(key, message))
```

## Possible issues

A string is not `Bytes`; convert it first:

```ant
let mac = Crypto.hmacSha256("key".toBytes(), "message")
// error: cannot use string where Bytes is expected
```

`hmacSha256` is in `Crypto`, not `Bytes`, so a dot call on the key does not find it:

```ant
let key = "key".toBytes()
let mac = key.hmacSha256("message".toBytes())
// error: Bytes has no field or function hmacSha256
```

Piping the message makes it the key. The call type checks, and gives a different MAC:

```ant
let key = "key".toBytes()
let body = "body".toBytes()
assert((body | Crypto.hmacSha256(key)) != Crypto.hmacSha256(key, body))
```

Comparing hex strings fails when the sender writes upper case. Decode the received hex and compare
bytes:

```ant
let mac = Crypto.hmacSha256("k".toBytes(), "m".toBytes())
let received = Encoding.toHex(mac).toUpper()
assert(received != Encoding.toHex(mac))
assert(Crypto.equal(Encoding.fromHex(received).unwrap(), mac))
```

`==` on two MACs gives the same answer as `Crypto.equal`, but only `Crypto.equal` promises a time
that does not depend on the contents. Use it for a MAC that comes from outside.

## See also

[Crypto.equal](equal.md) · [Crypto.hmacSha1](hmacSha1.md) · [Crypto.hmacSha512](hmacSha512.md) ·
[Crypto.sha256](sha256.md) · [Crypto.randomBytes](randomBytes.md) ·
[Encoding.toHex](../Encoding/toHex.md) · [Encoding.toBase64](../Encoding/toBase64.md) ·
[String.toBytes](../String/toBytes.md)
