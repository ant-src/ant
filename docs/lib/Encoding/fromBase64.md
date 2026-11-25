<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Encoding.fromBase64

```ant signature
func Encoding.fromBase64(s: string, url: bool = false): Result[Bytes, Error]
```

Encoding.fromBase64 decodes s as base64 in the alphabet url names, with the padding present or
absent. It ignores \\r and \\n, and is Err on any other character outside the alphabet or on a
length no encoding produces.

<!-- end generated -->

## Details

- Accepts the text with its `=` padding or without it, and skips `\r` and `\n`, so text wrapped at a
  fixed width decodes as is. Any other character outside the alphabet, a space included, gives an
  `Err`.
- Reads one alphabet only: the standard one with `+` and `/`, or with `url: true` the URL-safe one
  with `-` and `_`.
- Bits past the last byte in the final character are ignored, as in Go, so `aGl=` and `aGk=` both
  decode to the bytes of `hi`.

## Basic examples

```ant
Encoding.fromBase64("aGVsbG8=")                                        // Ok{ value: Bytes(68 65 6c 6c 6f) }
Encoding.fromBase64("aGVsbG8=") | Result.unwrap() | String.fromBytes() // "hello"
Encoding.fromBase64("hello!")                                          // Err{ error: Error{ message: "illegal base64 data at input byte 5", kind: null } }
```

## Scope

The padding may be present or absent:

```ant
Encoding.fromBase64("aGk=") // Ok{ value: Bytes(68 69) }
Encoding.fromBase64("aGk")  // Ok{ value: Bytes(68 69) }
```

The empty string decodes to empty bytes, and so does a string of line breaks alone:

```ant
Encoding.fromBase64("")     // Ok{ value: Bytes() }
Encoding.fromBase64("\r\n") // Ok{ value: Bytes() }
```

Line breaks anywhere in the text are skipped, as in base64 wrapped at 76 columns:

```ant
let wrapped = "SGVsbG8s\r\nIEFudCE=\r\n"
Encoding.fromBase64(wrapped) | Result.unwrap() | String.fromBytes() // "Hello, Ant!"
```

The bytes need not be text. These are the first four bytes of every PNG file:

```ant
Encoding.fromBase64("iVBORw==") // Ok{ value: Bytes(89 50 4e 47) }
```

A string in any script round-trips through its UTF-8 bytes:

```ant
Encoding.fromBase64("8J+QnCBhbnQ=") | Result.unwrap() | String.fromBytes() // "🐜 ant"
```

A length no encoding produces is an `Err`. Four characters carry three bytes, so a single character
left over carries no whole byte, and padding must end the text at a multiple of four:

```ant
Encoding.fromBase64("aGVsb")     // Err{ error: Error{ message: "illegal base64 data at input byte 4", kind: null } }
Encoding.fromBase64("aGVsbG8==") // Err{ error: Error{ message: "illegal base64 data at input byte 8", kind: null } }
```

The text is the first argument, so `fromBase64` ends a pipe:

```ant
"  aGk=\n" | String.trim() | Encoding.fromBase64() // Ok{ value: Bytes(68 69) }
```

## Named arguments

### url

`url: true` reads the URL-safe alphabet, which writes `-` and `_` where the standard one writes `+`
and `/`:

```ant
Encoding.fromBase64("-_8=", url: true) // Ok{ value: Bytes(fb ff) }
Encoding.fromBase64("+/8=")            // Ok{ value: Bytes(fb ff) }
```

## Applications

### Read a Basic authorization header

An HTTP `Authorization: Basic` header carries `user:password` in base64. Decode it, check that it is
UTF-8, and split it at the first colon:

```ant
type Login = { user: string, password: string }

func parseBasic(header: string): Result[Login, Error] {
  if !header.startsWith("Basic ") {
    return Result.fail("not a Basic header")
  }
  let encoded = header.trimPrefix("Basic ")
  let bytes = try (Encoding.fromBase64(encoded) | Result.wrapErr("credentials"))
  let text = String.fromBytes(bytes) ?? return Result.fail("credentials: not UTF-8")
  let cut = text.cut(":") ?? return Result.fail("credentials: no colon")
  Result.ok(Login{ user: cut.before, password: cut.after })
}

println(parseBasic("Basic YW5uOnMzY3JldA=="))
println(parseBasic("Basic YW5u!"))
println(parseBasic("Basic YW5u"))
```

```text
Ok{ value: Login{ user: "ann", password: "s3cret" } }
Err{ error: Error{ message: "credentials: illegal base64 data at input byte 4", kind: null } }
Err{ error: Error{ message: "credentials: no colon", kind: null } }
```

### Read the payload of a JSON Web Token

The three parts of a JSON Web Token are base64 in the URL-safe alphabet without padding. The payload
is the middle one:

```ant
func payload(token: string): Result[Json.Value, Error] {
  let parts = token.split(".")
  if parts.length() != 3 {
    return Result.fail("token: want 3 parts, got ${parts.length()}")
  }
  let bytes = try (Encoding.fromBase64(parts[1], url: true) | Result.wrapErr("token payload"))
  let text = String.fromBytes(bytes) ?? return Result.fail("token payload: not UTF-8")
  Json.parse(text)
}

let token = "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJhbnQiLCJhZG1pbiI6dHJ1ZX0.c2ln"
println(payload(token))
```

```text
Ok{ value: Map{ "sub": "ant", "admin": true } }
```

### Decode a file of base64

A file of base64 text usually ends in a newline. `fromBase64` skips it, so the text needs no trim:

<!-- test: file key.b64
AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=
-->

```ant
let text = Result.unwrap(File.read("key.b64"))
let key = Result.unwrap(Encoding.fromBase64(text))
println("${key.length()} bytes, last ${key[key.length() - 1]}")
```

```text
32 bytes, last 31
```

## Properties and relations

`fromBase64` undoes [Encoding.toBase64](toBase64.md) in either alphabet:

```ant
let data = Bytes.fromArray([0, 1, 254, 255, 62, 63])
assert(Result.unwrap(Encoding.fromBase64(Encoding.toBase64(data))) == data)
assert(Result.unwrap(Encoding.fromBase64(Encoding.toBase64(data, url: true), url: true)) == data)
```

Dropping the padding changes nothing:

```ant
let padded = Encoding.toBase64("ants".toBytes())
assert(padded.endsWith("="))
assert(Encoding.fromBase64(padded) == Encoding.fromBase64(padded.replace("=", "")))
```

Four characters carry three bytes, and each `=` of padding stands for one byte fewer:

```ant
let text = "QW50IGlzIGZ1bg=="
let n = Result.unwrap(Encoding.fromBase64(text)).length()
assert(n == 10)
assert(text.length() / 4 * 3 - 2 == n)
```

## Possible issues

### The wrong alphabet

Text in one alphabet that uses `+`, `/`, `-`, or `_` fails in the other. Pass `url: true` for base64
from URLs, file names, and JSON Web Tokens:

```ant
Encoding.fromBase64("aGk_")            // Err{ error: Error{ message: "illegal base64 data at input byte 3", kind: null } }
Encoding.fromBase64("aGk_", url: true) // Ok{ value: Bytes(68 69 3f) }
```

### Spaces are not skipped

Only `\r` and `\n` are skipped. Text with spaces, such as base64 pasted with indentation, fails;
remove them first:

```ant
let pasted = "aGVs bG8="
Encoding.fromBase64(pasted).isOk()                  // false
Encoding.fromBase64(pasted.replace(" ", "")).isOk() // true
```

### The result is bytes

`fromBase64` gives `Bytes`, so a string operation needs [String.fromBytes](../String/fromBytes.md)
first:

```ant
let bytes = Result.unwrap(Encoding.fromBase64("aGk="))
bytes.toUpper() // error: Bytes has no field or function toUpper
```

## See also

- [Encoding.toBase64](toBase64.md): bytes to base64 text.
- [Encoding.fromHex](fromHex.md): bytes from hex text.
- [String.fromBytes](../String/fromBytes.md): decode the bytes as UTF-8 text.
- [Bytes](../Bytes/index.md): the type of the result.
- [Result](../Result/index.md): working with the `Ok` or `Err`.
