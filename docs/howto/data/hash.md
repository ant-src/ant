<!-- status: reviewed -->

# Hash data and compare hashes

Compute a digest of a string or a file, check it against a published checksum, and verify a signed
message.

## Hash a string

Turn the string into bytes with [String.toBytes](../../lib/String/toBytes.md), hash them with
[Crypto.sha256](../../lib/Crypto/sha256.md), and write the digest as hex with
[Encoding.toHex](../../lib/Encoding/toHex.md):

```ant
let digest = "hello".toBytes() | Crypto.sha256() | Encoding.toHex()
println(digest)
```

```text
2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
```

A digest is `Bytes`, 32 bytes long for SHA-256. `toHex` writes two lower-case digits per byte. Use
[Crypto.sha512](../../lib/Crypto/sha512.md) the same way for a 64-byte digest.

## Hash a file

Read the file as bytes with [File.readBytes](../../lib/File/readBytes.md), so that the digest covers
exactly what the file holds:

<!-- test: file notes.txt
hello
-->

```ant
let data = try File.readBytes("notes.txt")
println(data | Crypto.sha256() | Encoding.toHex())
```

```text
5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03
```

## Compare with a published checksum

Decode the expected checksum with [Encoding.fromHex](../../lib/Encoding/fromHex.md) and compare the
bytes with `==`. `fromHex` reads either case, so a checksum written in upper case matches too, and
it returns `Err` for text that is not hex:

<!-- test: file notes.txt
hello
-->

```ant
let published = "5891B5B522D5DF086D0FF0B110FBD9D21BB4FC7163AF34D08286A2E846F6BE03"
let expected = try Encoding.fromHex(published)
let actual = Crypto.sha256(try File.readBytes("notes.txt"))
println(if actual == expected { "OK" } else { "FAILED" })
```

```text
OK
```

A checksum guards against a damaged download, not against an attacker who can change both the file
and its checksum. An HMAC, below, guards against tampering when both sides share a secret, such as a
webhook sender and its receiver.

## Verify a signed message

A webhook sender signs the request body with a shared secret and sends the HMAC in a header. Compute
the HMAC of the body with [Crypto.hmacSha256](../../lib/Crypto/hmacSha256.md) and compare it with
the received one using [Crypto.equal](../../lib/Crypto/equal.md), never `==`: `Crypto.equal` takes
time that depends on the lengths of the two values, not on their contents, so timing the check does
not show where a forged signature goes wrong.

```ant
func verify(secret: string, body: string, signature: string): bool {
  let expected = Crypto.hmacSha256(secret.toBytes(), body.toBytes())
  let received = Encoding.fromHex(signature).unwrapOr(Bytes.new())
  Crypto.equal(received, expected)
}

let body = "{\"event\":\"push\"}"
let signature = "7eaf83970d7ec8ac8821fb679caa6197a556926ca35aaee7893bc5f4461a6bba"
println(verify("webhook-secret", body, signature))
println(verify("webhook-secret", "{\"event\":\"pull\"}", signature))
println(verify("webhook-secret", body, "not hex"))
```

```text
true
false
false
```

[Result.unwrapOr](../../lib/Result/unwrapOr.md) turns a signature that is not hex into empty bytes,
and `Crypto.equal` finds sequences of different lengths unequal, so a malformed header fails the
check like a wrong one.

Some senders put a prefix such as `sha256=` before the hex; remove it with
[String.trimPrefix](../../lib/String/trimPrefix.md) before decoding. Others write the HMAC in
base64; decode it with [Encoding.fromBase64](../../lib/Encoding/fromBase64.md) instead of `fromHex`:

```ant
let header = "sha256=7eaf83970d7ec8ac8821fb679caa6197a556926ca35aaee7893bc5f4461a6bba"
let fromPrefixed = Encoding.fromHex(header.trimPrefix("sha256=")).unwrapOr(Bytes.new())
let encoded = "fq+Dlw1+yKyIIftnnKphl6VWkmyjWq7niTvF9EYaa7o="
let fromBase64 = Encoding.fromBase64(encoded).unwrapOr(Bytes.new())
println(Crypto.equal(fromPrefixed, fromBase64))
```

```text
true
```

To sign a message you send, compute the same HMAC and write it as hex:

```ant
let signature = Crypto.hmacSha256("webhook-secret".toBytes(), "{\"event\":\"push\"}".toBytes())
  | Encoding.toHex()
println(signature)
```

```text
7eaf83970d7ec8ac8821fb679caa6197a556926ca35aaee7893bc5f4461a6bba
```

Use [Crypto.hmacSha512](../../lib/Crypto/hmacSha512.md) when the other side asks for SHA-512.

## Match an older format

[Crypto.sha1](../../lib/Crypto/sha1.md), [Crypto.md5](../../lib/Crypto/md5.md), and
[Crypto.hmacSha1](../../lib/Crypto/hmacSha1.md) are there for formats that still use them, such as
git object names and TOTP codes. Collisions can be made for both digests, so do not pick them for
anything new:

```ant
let data = "abc".toBytes()
println(Crypto.sha1(data) | Encoding.toHex())
println(Crypto.md5(data) | Encoding.toHex())
```

```text
a9993e364706816aba3e25717850c26c9cd0d89d
900150983cd24fb0d6963f7d28e17f72
```

## The whole program

This script checks files against a checksum list in the format `sha256sum` writes: a hex digest, two
spaces, and a file name on each line. It prints `OK` or `FAILED` for each file, and exits with
status 1 when any file fails, as `sha256sum -c` does:

<!-- test: file SHA256SUMS
5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03  notes.txt
0000000000000000000000000000000000000000000000000000000000000000  logo.txt
-->
<!-- test: file notes.txt
hello
-->
<!-- test: file logo.txt
Ant
-->
<!-- test: exit 1 -->

```ant
let list = try File.read("SHA256SUMS")
var failed = 0
for line in list.lines() {
  let entry = line.cut("  ") ?? return Result.fail("malformed line: ${line}")
  let expected = try (Encoding.fromHex(entry.before) | Result.wrapErr("checksum of ${entry.after}"))
  let actual = Crypto.sha256(try File.readBytes(entry.after))
  let ok = actual == expected
  if !ok {
    failed += 1
  }
  let status = if ok { "OK" } else { "FAILED" }
  println("${entry.after}: ${status}")
}
if failed > 0 {
  return Result.fail("files that failed: ${failed}")
}
```

```text
notes.txt: OK
logo.txt: FAILED
```

The error goes to standard error:

```sh
$ ant run check.ant; echo $?
notes.txt: OK
logo.txt: FAILED
files that failed: 1
1
```

1. Read the list with [File.read](../../lib/File/read.md) and split it with
   [String.lines](../../lib/String/lines.md).
2. Split each line at the two spaces with [String.cut](../../lib/String/cut.md). Stop the script
   with [Result.fail](../../lib/Result/fail.md) on a line without them.
3. Decode the digest with `Encoding.fromHex`, and name the file in the error with
   [Result.wrapErr](../../lib/Result/wrapErr.md).
4. Hash the file with `File.readBytes` and `Crypto.sha256`, compare the bytes with `==`, and count
   the files that differ.
5. After the loop, return `Result.fail` when any file failed. The script writes the message to
   standard error and exits with status 1, so a CI job or another script that runs it sees the
   failure.
