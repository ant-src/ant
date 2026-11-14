<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Crypto

Crypto holds cryptographic hashes, HMACs, a constant-time comparison, and secure randomness.

<!-- end generated -->

## Details

- Crypto covers what a program needs to check a webhook, sign a request, or make a token: digests,
  message authentication, and random bytes. It has no encryption, no signatures with key pairs, and
  no key exchange.
- The hashes and HMACs take and return `Bytes`. Convert text with
  [String.toBytes](../String/toBytes.md), and show a digest with
  [Encoding.toHex](../Encoding/toHex.md) or [Encoding.toBase64](../Encoding/toBase64.md).
- [Crypto.sha1](sha1.md) and [Crypto.md5](md5.md) are broken against collisions. Use them only for
  formats that still need them, as git and TOTP do.
- [Crypto.randomBytes](randomBytes.md) and [Crypto.uuid](uuid.md) read the operating system's secure
  source of randomness, which does not fail and which no program can seed. A seeded
  [Random](../Random/index.md) generator is not fit for keys or tokens.

## Basic examples

Hash text with `Crypto.sha256` and show the digest as hex:

```ant
let digest = Crypto.sha256("abc".toBytes())
println(Encoding.toHex(digest))
```

```text
ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
```

Compute a MAC with `Crypto.hmacSha256`, and compare a received one with `Crypto.equal`:

```ant
let key = "secret".toBytes()
let mac = Crypto.hmacSha256(key, "hello".toBytes())
let received = Encoding.fromBase64("iKqz7ejTrflNJquQ07r9SiCDBww7zOnAFO4EpEOEfAs=").unwrap()
Crypto.equal(received, mac) // true
```

Make a secret key with `Crypto.randomBytes` and an id with `Crypto.uuid`:

```ant
Crypto.randomBytes(32).length() // 32
Crypto.uuid().length()          // 36
```

## Scope

Each hash has a fixed digest length, whatever the length of the input:

```ant
let data = "abc".toBytes()
Crypto.sha256(data).length() // 32
Crypto.sha512(data).length() // 64
Crypto.sha1(data).length()   // 20
Crypto.md5(data).length()    // 16
```

Empty input has a digest too:

```ant
let digest = Crypto.sha256("".toBytes())
println(Encoding.toHex(digest))
```

```text
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
```

The data comes first, so a pipe carries text through to its digest:

```ant
"abc".toBytes() | Crypto.md5() | Encoding.toHex() // "900150983cd24fb0d6963f7d28e17f72"
```

`Crypto.equal` gives `false` for sequences of different lengths:

```ant
Crypto.equal("ab".toBytes(), "abc".toBytes()) // false
Crypto.equal("".toBytes(), "".toBytes())      // true
```

## Functions by job

### Hashing

- [Crypto.sha256](sha256.md): the 32-byte SHA-256 digest
- [Crypto.sha512](sha512.md): the 64-byte SHA-512 digest
- [Crypto.sha1](sha1.md): the 20-byte SHA-1 digest, for formats that still use it
- [Crypto.md5](md5.md): the 16-byte MD5 digest, for formats that still use it

### Message authentication

- [Crypto.hmacSha256](hmacSha256.md): the 32-byte HMAC-SHA-256 of a message under a key
- [Crypto.hmacSha512](hmacSha512.md): the 64-byte HMAC-SHA-512
- [Crypto.hmacSha1](hmacSha1.md): the 20-byte HMAC-SHA-1
- [Crypto.equal](equal.md): whether two byte sequences match, in time that does not depend on their
  contents

### Randomness

- [Crypto.randomBytes](randomBytes.md): `n` bytes from the operating system's secure source
- [Crypto.uuid](uuid.md): a random version 4 UUID, in lower case with hyphens

## Applications

Compute the id git gives a file's contents, the SHA-1 of a header and the bytes:

```ant
func gitBlobId(content: string): string {
  let bytes = content.toBytes()
  let header = "blob ${bytes.length()}\u{0}".toBytes()
  Bytes.concat([header, bytes]) | Crypto.sha1() | Encoding.toHex()
}

println(gitBlobId("hello\n"))
```

```text
ce013625030ba8dba906f756967f9e9ca394464a
```

Check a webhook delivery whose header carries the hex HMAC of the body:

```ant
func verify(secret: string, body: Bytes, header: string): bool {
  let received = Encoding.fromHex(header)
  if received is Err { return false }
  Crypto.equal(received.value, Crypto.hmacSha256(secret.toBytes(), body))
}

let body = "{\"action\":\"opened\"}".toBytes()
let header = "2b063a104e1fcedc86cef600a6c042707bac7d6c96f5e864b5e78e966f07aacd"
println(verify("hook-secret", body, header))
println(verify("wrong-secret", body, header))
```

```text
true
false
```

Checksum a file, the most common use of a hash, and name the file when it cannot be read:

```ant
func checksum(path: string): Result[string, Error] {
  let data = try (File.readBytes(path) | Result.wrapErr("checksum ${path}"))
  Result.ok(Crypto.sha256(data) | Encoding.toHex())
}

let path = try File.tempFile("checksum-*")
defer { File.remove(path) }
try File.write(path, "abc")
println(try checksum(path))
println(checksum("missing.bin"))
```

```text
ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
Err{ error: Error{ message: "checksum missing.bin: open missing.bin: no such file or directory" } }
```

Make a session token that is safe to put in a URL or a cookie:

```ant
let token = Crypto.randomBytes(24) | Encoding.toBase64(url: true)
token.length() // 32
```

## Properties and relations

A digest depends only on its input, so equal inputs give equal digests, and here a change of one
byte gives a different digest:

```ant
let a = "hello".toBytes()
assert(Crypto.sha256(a) == Crypto.sha256("hello".toBytes()))
assert(Crypto.sha256(a) != Crypto.sha256("hellp".toBytes()))
```

`Crypto.equal` gives the same answer as `==` on `Bytes`; only its timing differs:

```ant
let mac = Crypto.hmacSha256("k".toBytes(), "m".toBytes())
for other in [mac, Crypto.sha256(mac), "".toBytes()] {
  assert(Crypto.equal(mac, other) == (mac == other))
}
```

## Possible issues

A string is not `Bytes`; convert it first:

```ant
let digest = Crypto.sha256("abc") // error: cannot use string where Bytes is expected
```

The functions are in `Crypto`, not `Bytes`, so a dot call on the data does not find them:

```ant
let digest = "abc".toBytes().sha256() // error: Bytes has no field or function sha256
```

Use a pipe:

```ant
"abc".toBytes() | Crypto.sha256() | Encoding.toHex() // "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
```

A negative count panics:

```ant
Crypto.randomBytes(-1) // panic: Crypto.randomBytes needs a count of at least 0, got -1
```

`==` on a received MAC gives the right answer, but only `Crypto.equal` promises a time that does not
depend on the contents. Use `Crypto.equal` for a signature that comes from outside.

## See also

- [Encoding](../Encoding/index.md): hex and base64, the text forms of digests and keys
- [Bytes](../Bytes/index.md): binary data, the input and output of every hash
- [Random](../Random/index.md): a seeded generator for simulations and tests, not for secrets
