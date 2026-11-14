<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Encoding.toBase64

```ant signature
func Encoding.toBase64(bytes: Bytes, url: bool = false): string
```

Encoding.toBase64 encodes bytes as padded base64 in the standard alphabet, with + and /, or when url
is true in the URL-safe one, with - and \_.

<!-- end generated -->

## Details

- Every 3 bytes become 4 characters, and `=` pads the last group, so the result is always a multiple
  of 4 characters long. Empty `Bytes` give `""`.
- `url: true` changes only `+` and `/` to `-` and `_`. The padding stays.

## Basic examples

Encode the bytes of a string:

```ant
Encoding.toBase64(String.toBytes("hello")) // "aGVsbG8="
```

Encode for a URL or a file name with the URL-safe alphabet:

```ant
let b = Bytes.fromArray([0xfb, 0xff])
Encoding.toBase64(b)            // "+/8="
Encoding.toBase64(b, url: true) // "-_8="
```

## Scope

Empty `Bytes` encode as the empty string:

```ant
Encoding.toBase64(Bytes.new()) // ""
```

One or two bytes left over after the last full group of 3 get two or one `=`:

```ant
Encoding.toBase64(String.toBytes("a"))   // "YQ=="
Encoding.toBase64(String.toBytes("ab"))  // "YWI="
Encoding.toBase64(String.toBytes("abc")) // "YWJj"
```

Text outside ASCII encodes as its UTF-8 bytes:

```ant
"héllo 🐜" | String.toBytes() | Encoding.toBase64() // "aMOpbGxvIPCfkJw="
```

Any bytes encode, including ones that are not text, such as the start of a PNG file:

```ant
Bytes.fromArray([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]) | Encoding.toBase64() // "iVBORw0KGgo="
```

End a pipe with it to encode a digest:

```ant
"abc" | String.toBytes() | Crypto.sha256() | Encoding.toBase64() // "ungWv48Bz+pBQUDeXa4iI7ADYaOWF3qctBD/YfIAFa0="
```

## Named arguments

`url` picks the alphabet. With `url: true`, the result holds no `+` or `/`, which a URL or a path
would read as something else:

```ant
let digest = Crypto.sha256(String.toBytes("abc"))
Encoding.toBase64(digest, url: true) // "ungWv48Bz-pBQUDeXa4iI7ADYaOWF3qctBD_YfIAFa0="
```

Both alphabets pad with `=`. Where a token must not hold `=` either, remove it;
`Encoding.fromBase64` reads the result without it:

```ant
let padded = Encoding.toBase64(String.toBytes("ants"), url: true)
padded                                                                 // "YW50cw=="
padded.replace("=", "")                                                // "YW50cw"
Encoding.fromBase64("YW50cw", url: true).unwrap() | String.fromBytes() // "ants"
```

## Applications

Build an HTTP Basic authorization header from a user name and a password:

```ant
func basicAuth(user: string, password: string): string {
  let credentials = "${user}:${password}" | String.toBytes() | Encoding.toBase64()
  "Basic ${credentials}"
}

println(basicAuth("ann", "s3cret"))
```

```text
Basic YW5uOnMzY3JldA==
```

Embed a small image in HTML as a data URL:

```ant
let gif = Bytes.fromArray([
  0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x3b,
])
let src = "data:image/gif;base64,${Encoding.toBase64(gif)}"
println(src)
```

```text
data:image/gif;base64,R0lGODlhAQABAAAAADs=
```

Put binary data in JSON. `Json.stringify` has no form for `Bytes`, so a record holds the encoded
string instead:

```ant
type Attachment = { name: string, data: string }

let file = Bytes.fromArray([0x00, 0x01, 0xfe, 0xff])
let attachment = Attachment{ name: "blob.bin", data: Encoding.toBase64(file) }
println(Json.stringify(attachment))
```

```text
{"name":"blob.bin","data":"AAH+/w=="}
```

Make a random session token that is safe in a URL. 24 random bytes give 32 characters with no
padding:

```ant
let token = Crypto.randomBytes(24) | Encoding.toBase64(url: true)
println(token.length())
println(["+", "/", "="].some({ token.contains(x) }))
```

```text
32
false
```

## Properties and relations

`Encoding.fromBase64` reverses it, in either alphabet:

```ant
let data = Bytes.fromArray([0x00, 0x7f, 0x80, 0xfb, 0xff])
assert((data | Encoding.toBase64() | Encoding.fromBase64() | Result.unwrap()) == data)
let safe = Encoding.toBase64(data, url: true)
assert(Encoding.fromBase64(safe, url: true).unwrap() == data)
```

The length is 4 characters for every 3 bytes, rounded up:

```ant
for n in Iter.range(0, 20) {
  let encoded = Encoding.toBase64(Bytes.new(n))
  assert(encoded.length() == (n + 2) / 3 * 4)
}
```

The URL-safe result is the standard one with `+` and `/` replaced:

```ant
let data = Crypto.sha256(String.toBytes("abc"))
let standard = Encoding.toBase64(data)
assert(Encoding.toBase64(data, url: true) == standard.replace("+", "-").replace("/", "_"))
```

## Possible issues

A string is not `Bytes`; convert it first:

```ant
let encoded = Encoding.toBase64("hello") // error: cannot use string where Bytes is expected
```

`toBase64` belongs to `Encoding`, not `Bytes`, so a dot call on `Bytes` does not find it. Use a
pipe:

```ant
let b = String.toBytes("hello")
let dotted = b.toBase64() // error: Bytes has no field or function toBase64
let piped = b | Encoding.toBase64()
```

The result is a snapshot. `Bytes` can change after it is encoded, and the string does not follow:

```ant
let b = String.toBytes("hi")
let before = Encoding.toBase64(b)
b.push(byte(0x21))
before               // "aGk="
Encoding.toBase64(b) // "aGkh"
```

Base64 is an encoding, not encryption. Anyone can decode it, so it hides nothing:

```ant
let hidden = Encoding.toBase64(String.toBytes("s3cret"))
hidden | Encoding.fromBase64() | Result.unwrap() | String.fromBytes() // "s3cret"
```

Decoding needs the same alphabet. A URL-safe string that holds `-` or `_` is `Err` read as standard:

```ant
let encoded = Encoding.toBase64(Bytes.fromArray([0xfb, 0xff]), url: true)
Encoding.fromBase64(encoded).isOk()            // false
Encoding.fromBase64(encoded, url: true).isOk() // true
```

## See also

[Encoding.fromBase64](fromBase64.md) · [Encoding.toHex](toHex.md) ·
[String.toBytes](../String/toBytes.md) · [Crypto.randomBytes](../Crypto/randomBytes.md) ·
[Crypto.sha256](../Crypto/sha256.md) · [Json.stringify](../Json/stringify.md)
