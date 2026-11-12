<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Crypto.hmacSha1

```ant signature
func Crypto.hmacSha1(key: Bytes, message: Bytes): Bytes
```

Crypto.hmacSha1 returns the 20-byte HMAC-SHA-1 of message under key.

<!-- end generated -->

## Details

- Turn text into bytes with `String.toBytes` and a hex key into bytes with `Encoding.fromHex`.
- Any key length works, the empty key included.
- Encode the result with `Encoding.toHex` or `Encoding.toBase64` to print or send it.
- Check a received signature with [Crypto.equal](equal.md), not `==`.
- HMAC-SHA-1 is built on SHA-1, which is broken against collisions; it is here for the formats that
  still use it, such as TOTP. See also [Crypto.hmacSha256](hmacSha256.md).

## Basic examples

Sign a message with a key:

```ant
let key = "key".toBytes()
let message = "The quick brown fox jumps over the lazy dog".toBytes()
Crypto.hmacSha1(key, message) // Bytes(de 7c 9b 85 b8 b7 8a a6 bc 8a 7a 36 f7 0a 90 70 1c 9d b4 d9)
```

Encode the result as hex to print or compare it with a published value:

```ant
let key = "key".toBytes()
let message = "The quick brown fox jumps over the lazy dog".toBytes()
Crypto.hmacSha1(key, message) | Encoding.toHex() // "de7c9b85b8b78aa6bc8a7a36f70a90701c9db4d9"
```

## Scope

An empty key and an empty message are both valid:

```ant
let empty = "".toBytes()
Crypto.hmacSha1(empty, empty) | Encoding.toHex() // "fbdb1d1b18aa6c08324b7d64b71fb76370690e1d"
```

A key can be any bytes, not only text. This one is twenty `0x0b` bytes:

```ant
let key = Bytes.fromArray(Array.filled(20, 0x0b))
Crypto.hmacSha1(key, "Hi There".toBytes()) | Encoding.toHex() // "b617318655057264e28bc0b6fb378c8ef146be00"
```

A key given in hex is decoded with `Encoding.fromHex` first:

```ant
let key = Encoding.fromHex("0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b") | Result.unwrap()
Crypto.hmacSha1(key, "Hi There".toBytes()) | Encoding.toHex() // "b617318655057264e28bc0b6fb378c8ef146be00"
```

Text is signed as its UTF-8 bytes, so `é` is signed as the two bytes `0xc3 0xa9`:

```ant
let key = "key".toBytes()
Crypto.hmacSha1(key, "é".toBytes()) == Crypto.hmacSha1(key, Bytes.fromArray([0xc3, 0xa9])) // true
```

A key longer than 64 bytes works the same way:

```ant
let key = Bytes.fromArray(Array.filled(80, 0xaa))
let message = "Test Using Larger Than Block-Size Key - Hash Key First".toBytes()
Crypto.hmacSha1(key, message) | Encoding.toHex() // "aa4ae5e15272d00e95705637ce8a3b55ed402112"
```

## Applications

### One-time passwords

HOTP, the algorithm behind TOTP authenticator codes, signs an 8-byte big-endian counter and reads
six decimal digits out of the result:

```ant
func hotp(secret: Bytes, counter: int): string {
  let message = Array.range(0, 8)
    | Array.map({ Bits.and(Bits.shiftRight(counter, 8 * (7 - x)), 0xff) })
    | Bytes.fromArray()
  let mac = Crypto.hmacSha1(secret, message).toArray() | Array.map({ int(x) })
  let offset = Bits.and(mac[19], 0x0f)
  let code = mac.slice(offset, offset + 4).reduce(0, func(acc: int, b: int) { acc * 256 + b })
  "${Bits.and(code, 0x7fffffff) % 1_000_000}".padStart(6, "0")
}

let secret = "12345678901234567890".toBytes()
for counter in Iter.range(0, 3) {
  println(hotp(secret, counter))
}
```

```text
755224
287082
359152
```

For TOTP, the counter is the Unix time in seconds divided by 30.

### Verify a webhook signature

A sender puts `sha1=` and the hex HMAC of the body in a header. Recompute it with the shared secret
and compare the two with `Crypto.equal`:

```ant
func verify(secret: Bytes, body: string, header: string): Result[bool, Error] {
  if !header.startsWith("sha1=") {
    return Result.fail("signature: missing sha1= prefix")
  }
  let received = try (Encoding.fromHex(header.trimPrefix("sha1=")) | Result.wrapErr("signature"))
  Result.ok(Crypto.equal(received, Crypto.hmacSha1(secret, body.toBytes())))
}

let secret = "It's a Secret to Everybody".toBytes()
let body = "{\"action\":\"opened\"}"
let header = "sha1=" + (Crypto.hmacSha1(secret, body.toBytes()) | Encoding.toHex())
println(verify(secret, body, header))
println(verify(secret, body + " ", header))
println(verify(secret, body, "md5=00"))
```

```text
Ok{ value: true }
Ok{ value: false }
Err{ error: Error{ message: "signature: missing sha1= prefix" } }
```

### Base64 signatures

Formats such as OAuth 1.0 send the signature in Base64 rather than hex:

```ant
let key = "key".toBytes()
let message = "The quick brown fox jumps over the lazy dog".toBytes()
Crypto.hmacSha1(key, message) | Encoding.toBase64() // "3nybhbi3iqa8ino29wqQcBydtNk="
```

## Properties and relations

The result is always 20 bytes, whatever the input lengths:

```ant
let key = "key".toBytes()
assert(Crypto.hmacSha1(key, "".toBytes()).length() == 20)
assert(Crypto.hmacSha1(key, Bytes.fromArray(Array.filled(1000, 7))).length() == 20)
```

The same key and message always give the same result; changing the key, here from `"a"` to `"b"`,
changes it:

```ant
let message = "hello".toBytes()
assert(Crypto.hmacSha1("a".toBytes(), message) == Crypto.hmacSha1("a".toBytes(), message))
assert(Crypto.hmacSha1("a".toBytes(), message) != Crypto.hmacSha1("b".toBytes(), message))
```

An HMAC is not the hash of the key and message joined together:

```ant
let key = "key".toBytes()
let message = "msg".toBytes()
assert(Crypto.hmacSha1(key, message) != Crypto.sha1(Bytes.concat([key, message])))
```

## Possible issues

Strings are not `Bytes`; passing them directly is a compile error for each argument:

```ant
Crypto.hmacSha1("key", "msg") // error: cannot use string where Bytes is expected
```

Convert text with `.toBytes()` first.

Swapping the arguments signs a different message under a different key; the type checker cannot
catch it, since both are `Bytes`:

```ant
let key = "key".toBytes()
let message = "msg".toBytes()
Crypto.hmacSha1(key, message) == Crypto.hmacSha1(message, key) // false
```

`String.toBytes` on a hex key signs with the hex digits as text, not with the bytes they spell:

```ant
let hexKey = "4b6579"
let wrong = Crypto.hmacSha1(hexKey.toBytes(), "msg".toBytes())
let right = Crypto.hmacSha1(Encoding.fromHex(hexKey) | Result.unwrap(), "msg".toBytes())
wrong == right // false
```

Comparing hex strings is case-sensitive, and `Encoding.toHex` writes lower case. Decode the received
value and compare bytes with `Crypto.equal` instead:

```ant
let mac = Crypto.hmacSha1("key".toBytes(), "msg".toBytes())
let received = (mac | Encoding.toHex()).toUpper()
(mac | Encoding.toHex()) == received                            // false
Crypto.equal(mac, Encoding.fromHex(received) | Result.unwrap()) // true
```

Comparing a received signature with `==` can take less time when the first bytes differ, which leaks
how much of a forged signature matches. `Crypto.equal` runs in time that depends only on the
lengths, as in [the webhook example](#verify-a-webhook-signature).

## Neat examples

HMAC-SHA-1 built from [Crypto.sha1](sha1.md): hash a key longer than 64 bytes with `Crypto.sha1`,
pad it to 64 bytes, hash it XORed with `0x36` followed by the message, then hash it XORed with
`0x5c` followed by that inner hash:

```ant
func hmac(key: Bytes, message: Bytes): Bytes {
  let k = if key.length() > 64 { Crypto.sha1(key) } else { key }
  let padded = k.toArray().concat(Array.filled(64 - k.length(), byte(0)))
  let pad = func(v: byte): Bytes { padded | Array.map({ Bits.xor(x, v) }) | Bytes.fromArray() }
  let inner = Crypto.sha1(Bytes.concat([pad(byte(0x36)), message]))
  Crypto.sha1(Bytes.concat([pad(byte(0x5c)), inner]))
}

let key = "key".toBytes()
let message = "The quick brown fox jumps over the lazy dog".toBytes()
assert(hmac(key, message) == Crypto.hmacSha1(key, message))
println(hmac(key, message) | Encoding.toHex())
```

```text
de7c9b85b8b78aa6bc8a7a36f70a90701c9db4d9
```

## See also

[Crypto.hmacSha256](hmacSha256.md), [Crypto.hmacSha512](hmacSha512.md), [Crypto.sha1](sha1.md),
[Crypto.equal](equal.md), [Encoding.toHex](../Encoding/toHex.md),
[Encoding.toBase64](../Encoding/toBase64.md), [Encoding.fromHex](../Encoding/fromHex.md)
