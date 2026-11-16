<!-- status: reviewed -->

# Generate random IDs

Make IDs for records, tokens for links and sessions, and short codes people can type.

1. Call [Crypto.uuid](../../lib/Crypto/uuid.md) for a standard UUID.
2. Call [Crypto.randomBytes](../../lib/Crypto/randomBytes.md) for a token of any length, and turn
   the bytes into text with [Encoding.toHex](../../lib/Encoding/toHex.md) or
   [Encoding.toBase64](../../lib/Encoding/toBase64.md).
3. Use a seeded [Random](../../lib/Random/index.md) generator only for IDs a test must see again,
   never for secrets.

## Make a UUID

`Crypto.uuid` returns a random version 4 UUID, in lower case with hyphens:

<!-- test: check-only the id is random -->

```ant
let id = Crypto.uuid()
println(id)
```

The program prints a line such as `43f692cc-2d9e-4718-94cc-98b1b5aec6d9`. Each call draws a new
random UUID, always 36 characters long:

```ant
let a = Crypto.uuid()
let b = Crypto.uuid()
a == b                           // false
a.length()                       // 36
a.split("-").map({ x.length() }) // [8, 4, 4, 4, 12]
```

## Make a token

For a session token, a password-reset link, or an API key, take random bytes and encode them. 32
bytes give 256 random bits. `Encoding.toHex` writes two digits per byte:

```ant
let token = Crypto.randomBytes(32) | Encoding.toHex()
println(token.length())
```

```text
64
```

For a shorter token that is safe in a URL, encode with `url: true`, which uses `-` and `_` in place
of `+` and `/`. Base64 pads with `=` unless the byte count is a multiple of 3, so 24 or 33 bytes
give a token without padding:

```ant
let token = Crypto.randomBytes(24) | Encoding.toBase64(url: true)
println(token.length())
println(token.contains("="))
```

```text
32
false
```

`Crypto.randomBytes` reads the operating system's source of randomness, which does not fail, so it
returns the bytes and not a `Result`.

## Make a short code from your own alphabet

For a code people read aloud or type, such as an invitation code, map each random byte to a
character of an alphabet. This alphabet has 32 characters and leaves out `i`, `l`, and `o`, which
people misread as `1` and `0`, and `u`, to keep 32 characters. 32 divides 256, so `% 32` picks every
character equally often:

```ant
let alphabet = "0123456789abcdefghjkmnpqrstvwxyz".chars()

func shortCode(n: int): string {
  Crypto.randomBytes(n)
    | Bytes.toArray()
    | Array.map({ alphabet[int(x) % alphabet.length()] })
    | String.join("")
}

let code = shortCode(10)
println(code.length())
println(code.chars().every({ alphabet.contains(x) }))
```

```text
10
true
```

Keep the alphabet at 2, 4, 8, 16, 32, 64, 128, or 256 characters. With another size, such as 36,
`% 36` picks the first characters more often than the rest.

## Make repeatable IDs for tests

A test that compares output needs the same IDs on every run. Draw them from a
[Random.new](../../lib/Random/new.md) generator with a fixed seed, and pass the generator to the
code that makes IDs. [Random.int](../../lib/Random/int.md) gives an index into the alphabet:

```ant
let alphabet = "0123456789abcdefghjkmnpqrstvwxyz".chars()

func testId(r: Random): string {
  Array.range(0, 6)
    | Array.map({ alphabet[r.int(0, alphabet.length())] })
    | String.join("")
}

let r = Random.new(42)
println([testId(r), testId(r)])
```

```text
["py114r", "j7yxh5"]
```

The same seed gives the same IDs on every run and every implementation of Ant, so anyone who knows
the seed can predict them. Use `Random` for test data and simulations only, and `Crypto.randomBytes`
or `Crypto.uuid` for anything a user must not guess. The same holds for `Random.new()` without a
seed and for `Math.random`: neither is fit for tokens, so use `Crypto.randomBytes` or `Crypto.uuid`.

## Check a token you issued

When a request brings a token back, compare it with the one you stored using
[Crypto.equal](../../lib/Crypto/equal.md). Its time depends on the lengths only, not on the
contents, so how long a check takes tells an attacker nothing about the stored token's contents.
[String.toBytes](../../lib/String/toBytes.md) turns the text into bytes:

```ant
let issued = Crypto.randomBytes(32) | Encoding.toHex()

func valid(received: string): bool {
  Crypto.equal(received.toBytes(), issued.toBytes())
}

println(valid(issued))
println(valid("guess"))
```

```text
true
false
```

## See also

[Crypto.uuid](../../lib/Crypto/uuid.md) · [Crypto.randomBytes](../../lib/Crypto/randomBytes.md) ·
[Crypto.equal](../../lib/Crypto/equal.md) · [Encoding.toHex](../../lib/Encoding/toHex.md) ·
[Encoding.toBase64](../../lib/Encoding/toBase64.md) · [Bytes.toArray](../../lib/Bytes/toArray.md) ·
[Random.new](../../lib/Random/new.md) · [Random.int](../../lib/Random/int.md) ·
[String.chars](../../lib/String/chars.md) · [String.join](../../lib/String/join.md) ·
[Hash data and compare hashes](hash.md)
