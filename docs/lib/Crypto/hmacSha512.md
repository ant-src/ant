<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Crypto.hmacSha512

```ant signature
func Crypto.hmacSha512(key: Bytes, message: Bytes): Bytes
```

Crypto.hmacSha512 returns the 64-byte HMAC-SHA-512 of message under key.

<!-- end generated -->

## Details

- The result is always 64 bytes, whatever the lengths of `key` and `message`; either may be empty.
- Compare a computed tag with a received one using [Crypto.equal](equal.md), not `==`.

## Basic examples

Sign a message with a key, and show the tag as hex:

```ant
let key = "Jefe".toBytes()
let message = "what do ya want for nothing?".toBytes()
println(Crypto.hmacSha512(key, message) | Encoding.toHex())
```

```text
164b7a7bfcf819e2e395fbe73b56e0a387bd64222e831fd610270cd7ea2505549758bf75c05a994a6d034f65f8f0e6fdcaeab1a34d4a6b4b636e070a38bce737
```

The tag is `Bytes`, 64 of them:

```ant
let tag = Crypto.hmacSha512("key".toBytes(), "hello".toBytes())
tag.length()    // 64
tag.slice(0, 4) // Bytes(ff 06 ab 36)
```

## Scope

An empty key and an empty message still give a 64-byte tag:

```ant
let empty = "".toBytes()
Crypto.hmacSha512(empty, empty).length()                       // 64
Crypto.hmacSha512("a long key".toBytes(), empty).length()      // 64
Crypto.hmacSha512(empty, "x".repeat(10000).toBytes()).length() // 64
```

The message may be any bytes, not only text. Here the message comes from hex:

```ant
let key = "k".toBytes()
let data = Encoding.fromHex("00ff10").unwrap()
println(Crypto.hmacSha512(key, data) | Encoding.toBase64())
```

```text
2BEldGvID2SA8fwSwsIxuyfVTyMK319GcdpL8DDWTMYsAg78F8V/CsABULfaV58EVQjEEPZhfLrvbLq0SeH4zg==
```

A key from [Crypto.randomBytes](randomBytes.md) is the usual choice for a secret only your program
knows. Its tag changes from run to run, so this program prints only the length:

```ant
let key = Crypto.randomBytes(64)
let tag = Crypto.hmacSha512(key, "session 42".toBytes())
println(tag.length())
```

```text
64
```

## Applications

Verify a webhook: recompute the tag of the body and compare it with the hex signature the sender put
in a header.

```ant
func verify(secret: Bytes, body: string, signature: string): bool {
  let received = Encoding.fromHex(signature)
  if received is Err { return false }
  let expected = Crypto.hmacSha512(secret, body.toBytes())
  Crypto.equal(expected, received.value)
}

let secret = "whsec".toBytes()
let body = "{\"event\":\"paid\"}"
let signature = Crypto.hmacSha512(secret, body.toBytes()) | Encoding.toHex()
println(verify(secret, body, signature))
println(verify(secret, "{\"event\":\"refunded\"}", signature))
println(verify(secret, body, "not hex"))
```

```text
true
false
false
```

Make a tamper-evident token: the payload, a dot, and the URL-safe base64 tag of the payload.

```ant
let key = "server key".toBytes()

func sign(payload: string): string {
  let tag = Crypto.hmacSha512(key, payload.toBytes())
  "${payload}.${Encoding.toBase64(tag, url: true)}"
}

func open(token: string): string | null {
  let parts = token.cut(".")
  if parts == null { return null }
  let tag = Encoding.fromBase64(parts.after, url: true)
  if tag is Err { return null }
  let expected = Crypto.hmacSha512(key, parts.before.toBytes())
  if Crypto.equal(expected, tag.value) { parts.before } else { null }
}

let token = sign("user=ann")
println(open(token))
println(open(token.replace("ann", "bob")))
```

```text
user=ann
null
```

## Properties and relations

The same key and message always give the same tag:

```ant
let key = "k".toBytes()
let message = "m".toBytes()
assert(Crypto.hmacSha512(key, message) == Crypto.hmacSha512(key, message))
```

Changing the key or the message changes the tag:

```ant
let message = "m".toBytes()
let tag = Crypto.hmacSha512("k1".toBytes(), message)
assert(tag != Crypto.hmacSha512("k2".toBytes(), message))
assert(tag != Crypto.hmacSha512("k1".toBytes(), "n".toBytes()))
```

The HMAC functions differ in the hash they use and so in the length of the tag:

```ant
let key = "k".toBytes()
let message = "m".toBytes()
assert(Crypto.hmacSha1(key, message).length() == 20)
assert(Crypto.hmacSha256(key, message).length() == 32)
assert(Crypto.hmacSha512(key, message).length() == 64)
```

An HMAC is not the SHA-512 digest of the key and message joined together:

```ant
let key = "k".toBytes()
let message = "m".toBytes()
assert(Crypto.hmacSha512(key, message) != Crypto.sha512(Bytes.concat([key, message])))
```

## Possible issues

The key comes first. Swapping the arguments compiles, since both are `Bytes`, and gives a different
tag:

```ant
let key = "k".toBytes()
let message = "m".toBytes()
Crypto.hmacSha512(key, message) == Crypto.hmacSha512(message, key) // false
```

Both arguments are `Bytes`. Convert a string with `toBytes` first:

```ant
let tag = Crypto.hmacSha512("secret", "hello")
// error: cannot use string where Bytes is expected
```

`Crypto.hmacSha512` is not a function of the `Bytes` namespace, so a dot call on the key does not
find it:

```ant
let key = "secret".toBytes()
let tag = key.hmacSha512("hello".toBytes()) // error: Bytes has no field or function hmacSha512
```

`==` on a received tag, or on its hex string, makes no promise about how long the comparison takes.
[Crypto.equal](equal.md) takes time that depends on the lengths only, not on the contents, so decode
the received tag and compare with it, as the webhook example above does.

## See also

[Crypto](index.md) · [Crypto.hmacSha256](hmacSha256.md) · [Crypto.hmacSha1](hmacSha1.md) ·
[Crypto.sha512](sha512.md) · [Crypto.equal](equal.md) · [Crypto.randomBytes](randomBytes.md) ·
[Encoding.toHex](../Encoding/toHex.md) · [Encoding.toBase64](../Encoding/toBase64.md)
