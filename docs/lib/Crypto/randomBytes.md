<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Crypto.randomBytes

```ant signature
func Crypto.randomBytes(n: int): Bytes
```

Crypto.randomBytes returns n bytes from the operating system's secure source of randomness. It
panics if n \< 0 or n is above 2^31 - 1.

<!-- end generated -->

## Details

- The bytes come from the operating system's source of randomness, which nothing a program can see
  seeds, so no seed makes a run repeat them.
- Reading the operating system's source of randomness does not fail, so the result is `Bytes`, not a
  `Result`.

## Basic examples

Make 32 random bytes, enough for a secret key:

```ant
let key = Crypto.randomBytes(32)
key.length() // 32
```

The bytes are rarely shown as they are. Encode them as text to print, store, or send them:

```ant
Crypto.randomBytes(16) | Encoding.toHex() | String.length() // 32
```

## Scope

A count of 0 gives empty `Bytes`:

```ant
Crypto.randomBytes(0) // Bytes()
```

The count comes first, so a pipe passes it, and the parameter is named `n`:

```ant
16 | Crypto.randomBytes() | Bytes.length() // 16
Crypto.randomBytes(n: 8).length()          // 8
```

The result is `Bytes`, so every `Bytes` function applies to it, such as `toArray` for the byte
values:

```ant
Crypto.randomBytes(4).toArray().length() // 4
```

## Applications

Make a session token as 64 hex digits:

```ant
func newSessionToken(): string {
  Crypto.randomBytes(32) | Encoding.toHex()
}

let token = newSessionToken()
println(token.length())
println(token.toLower() == token)
```

```text
64
true
```

Make a token for a link, in the URL-safe Base64 alphabet, which has no `+` or `/`. 24 bytes give 32
characters with no padding:

```ant
let code = Crypto.randomBytes(24) | Encoding.toBase64(url: true)
let link = "https://example.com/reset?code=${code}"
println(code.length())
println(["+", "/", "="].every({ !code.contains(x) }))
println(link.length())
```

```text
32
true
63
```

Sign a message with a fresh key, then check a received signature against it with `Crypto.equal`:

```ant
let key = Crypto.randomBytes(32)
let message = "amount=100".toBytes()
let signature = Crypto.hmacSha256(key, message)

func verify(key: Bytes, message: Bytes, received: Bytes): bool {
  Crypto.equal(Crypto.hmacSha256(key, message), received)
}

println(verify(key, message, signature))
println(verify(key, "amount=900".toBytes(), signature))
println(verify(Crypto.randomBytes(32), message, signature))
```

```text
true
false
false
```

## Properties and relations

The result has `n` bytes, and its hex form has two digits per byte:

```ant
let bytes = Crypto.randomBytes(10)
assert(bytes.length() == 10)
assert(Encoding.toHex(bytes).length() == 20)
```

Two calls of 16 bytes give the same bytes with a chance of 1 in 2^128:

```ant
assert(Crypto.randomBytes(16) != Crypto.randomBytes(16))
```

`Crypto.uuid` reads the same source and formats the result as a version 4 UUID:

```ant
let id = Crypto.uuid()
assert(id.length() == 36)
assert(id.split("-")[2].startsWith("4"))
```

## Possible issues

A negative count panics:

```ant
Crypto.randomBytes(-1) // panic: Crypto.randomBytes needs a count of at least 0, got -1
```

A count above 2^31 − 1 panics too:

```ant
Crypto.randomBytes(3000000000) // panic: Crypto.randomBytes needs a count of at most 2147483647, got 3000000000
```

Random bytes are rarely valid UTF-8, so `String.fromBytes` returns `null` for most of them, and the
type checker rejects using its result as a `string`. Encode the bytes with `Encoding.toHex` or
`Encoding.toBase64` instead:

```ant
let token: string = String.fromBytes(Crypto.randomBytes(16))
// error: cannot use string | null where string is expected
```

A `Random` generator built with a seed is not a source for keys or tokens: it gives the same values
on every run, so anyone who knows the seed knows the values. `Random.new()` without a seed and
`Math.random` are not fit for keys or tokens either; use `Crypto.randomBytes`. Two generators with
one seed draw the same values:

```ant
let a = Random.new(42)
let b = Random.new(42)
assert(a.int(0, 1000000) == b.int(0, 1000000))
```

## See also

[Crypto.uuid](uuid.md), [Crypto.equal](equal.md), [Crypto.hmacSha256](hmacSha256.md),
[Encoding.toHex](../Encoding/toHex.md), [Encoding.toBase64](../Encoding/toBase64.md),
[Bytes.length](../Bytes/length.md), [Random.new](../Random/new.md)
