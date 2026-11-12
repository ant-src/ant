<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Crypto.equal

```ant signature
func Crypto.equal(a: Bytes, b: Bytes): bool
```

Crypto.equal reports whether a and b hold the same bytes, in time that depends on their lengths and
not on their contents. It is for checking a received signature against a computed one; sequences of
different lengths are unequal.

<!-- end generated -->

## Details

- It returns the answer `==` returns on two `Bytes` values. Only `Crypto.equal` promises that the
  time it takes does not depend on the contents, so an attacker who times the check learns nothing
  about where a forged signature first differs.
- The time may depend on the lengths. A digest or an HMAC has a fixed, public length, so this gives
  nothing away.

## Basic examples

```ant
let key = String.toBytes("secret")
let mac = Crypto.hmacSha256(key, String.toBytes("hello"))
Crypto.equal(mac, Crypto.hmacSha256(key, String.toBytes("hello"))) // true
Crypto.equal(mac, Crypto.hmacSha256(key, String.toBytes("hellO"))) // false
```

Compare a received digest with one computed from the data:

```ant
let received = Encoding.fromHex("ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad").unwrap()
Crypto.equal(received, Crypto.sha256(String.toBytes("abc"))) // true
```

## Scope

Values of different lengths are unequal, even when one is a prefix of the other:

```ant
let a = Bytes.fromArray([1, 2, 3])
Crypto.equal(a, Bytes.fromArray([1, 2]))       // false
Crypto.equal(a, Bytes.fromArray([1, 2, 3, 0])) // false
Crypto.equal(a, Bytes.fromArray([1, 2, 3]))    // true
```

Two empty values are equal:

```ant
Crypto.equal(String.toBytes(""), Bytes.fromArray([7]).slice(0, 0)) // true
```

Bytes decoded from hex compare by value, whatever the case of the digits:

```ant
let received = Encoding.fromHex("C0FFEE").unwrap()
Crypto.equal(received, Bytes.fromArray([0xc0, 0xff, 0xee])) // true
```

Naming the arguments makes clear which side is received and which is computed; the order does not
change the answer:

```ant
let computed = Crypto.sha256(String.toBytes("data"))
Crypto.equal(a: computed, b: Crypto.sha256(String.toBytes("data"))) // true
Crypto.equal(b: computed, a: Crypto.sha256(String.toBytes("data"))) // true
```

## Applications

Check the signature header of a webhook, `sha256=` followed by the hex HMAC of the body:

```ant
func verify(secret: string, body: string, header: string): bool {
  if !header.startsWith("sha256=") { return false }
  let expected = Crypto.hmacSha256(String.toBytes(secret), String.toBytes(body))
  Encoding.fromHex(header.trimPrefix("sha256="))
    | Result.map({ Crypto.equal(x, expected) })
    | Result.unwrapOr(false)
}

let secret = "whsec"
let body = "{\"event\":\"push\"}"
let signature = Crypto.hmacSha256(String.toBytes(secret), String.toBytes(body)) | Encoding.toHex()

println(verify(secret, body, "sha256=${signature}"))
println(verify(secret, "{\"event\":\"delete\"}", "sha256=${signature}"))
println(verify(secret, body, "sha256=not-hex"))
println(verify(secret, body, signature))
```

```text
true
false
false
false
```

Check an API token against the SHA-256 digest you store instead of the token itself:

```ant
let stored = Crypto.sha256(String.toBytes("tok_8f2a91"))

func allowed(token: string): bool {
  Crypto.equal(Crypto.sha256(String.toBytes(token)), stored)
}

println(allowed("tok_8f2a91"))
println(allowed("tok_8f2a90"))
```

```text
true
false
```

Sign a value for a cookie and read it back only when the signature matches:

```ant
let key = String.toBytes("cookie-key")

func sign(value: string): string {
  let mac = Crypto.hmacSha256(key, String.toBytes(value))
  "${value}.${Encoding.toBase64(mac, url: true)}"
}

func unsign(cookie: string): string | null {
  let cut = cookie.cut(".")
  if cut == null { return null }
  let expected = Crypto.hmacSha256(key, String.toBytes(cut.before))
  let valid = Encoding.fromBase64(cut.after, url: true)
    | Result.map({ Crypto.equal(x, expected) })
    | Result.unwrapOr(false)
  if !valid { return null }
  cut.before
}

let cookie = sign("user=ann")
println(unsign(cookie))
println(unsign(cookie.replace("ann", "bob")))
```

```text
user=ann
null
```

## Properties and relations

`Crypto.equal` gives the same answer as `==` on `Bytes`:

```ant
let values = [
  String.toBytes(""),
  String.toBytes("a"),
  String.toBytes("ab"),
  Bytes.fromArray([0x61, 0x62]),
  Bytes.fromArray([0]),
]
for a in values {
  for b in values {
    assert(Crypto.equal(a, b) == (a == b))
  }
}
```

The order of the arguments does not matter:

```ant
let a = Crypto.sha256(String.toBytes("x"))
let b = Crypto.sha256(String.toBytes("y"))
assert(Crypto.equal(a, b) == Crypto.equal(b, a))
assert(Crypto.equal(a, a))
```

## Possible issues

Both arguments are `Bytes`. A signature that arrives as text must be decoded first; compare the
bytes, not the hex strings:

```ant
let mac = Crypto.hmacSha256(String.toBytes("k"), String.toBytes("m"))
println(Crypto.equal("3d6a...", mac)) // error: cannot use string where Bytes is expected
```

Turning the hex text into bytes with `String.toBytes` does not decode it. It gives the bytes of the
digits, which never equal the digest:

```ant
let mac = Crypto.hmacSha256(String.toBytes("k"), String.toBytes("m"))
let hex = Encoding.toHex(mac)
Crypto.equal(String.toBytes(hex), mac)            // false
Crypto.equal(Encoding.fromHex(hex).unwrap(), mac) // true
```

`Encoding.fromHex` returns a `Result`. Passing it straight in is a compile error:

```ant
let mac = Crypto.hmacSha256(String.toBytes("k"), String.toBytes("m"))
println(Crypto.equal(Encoding.fromHex("00"), mac))
// error: cannot use Result[Bytes, Error] where Bytes is expected
```

Comparing the hex strings with `==` gives the right answer for lower-case hex, but rejects a valid
signature written in upper case, and makes no promise about its timing:

```ant
let mac = Bytes.fromArray([0xc0, 0xff, 0xee])
Encoding.toHex(mac) == "C0FFEE"                        // false
Crypto.equal(Encoding.fromHex("C0FFEE").unwrap(), mac) // true
```

`Crypto.equal` lives in `Crypto`, not in `Bytes`, so it has no dot form on a `Bytes` value:

```ant
let mac = Crypto.sha256(String.toBytes("m"))
println(mac.equal(mac)) // error: Bytes has no field or function equal
```

## See also

[Crypto.hmacSha256](hmacSha256.md) · [Crypto.hmacSha512](hmacSha512.md) · [Crypto.sha256](sha256.md)
· [Encoding.fromHex](../Encoding/fromHex.md) · [Encoding.toHex](../Encoding/toHex.md) ·
[Bytes](../Bytes/index.md)
