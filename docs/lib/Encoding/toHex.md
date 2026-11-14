<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Encoding.toHex

```ant signature
func Encoding.toHex(bytes: Bytes): string
```

Encoding.toHex encodes bytes as two lower-case hex digits per byte.

<!-- end generated -->

## Details

- The result has exactly twice as many characters as `bytes` has bytes, with no separators and no
  `0x` prefix. Empty `Bytes` give `""`.
- The string is a snapshot: changing the `Bytes` afterwards does not change it.

## Basic examples

```ant
Encoding.toHex(String.toBytes("hi"))          // "6869"
Encoding.toHex(Bytes.fromArray([0, 15, 255])) // "000fff"
Encoding.toHex(Bytes.new())                   // ""
```

## Scope

Every byte takes two digits, so a byte below 16 keeps its leading zero:

```ant
Encoding.toHex(Bytes.fromArray([1, 10, 16, 171])) // "010a10ab"
Encoding.toHex(Bytes.new(3))                      // "000000"
```

Text goes through [String.toBytes](../String/toBytes.md) first, and the hex shows its UTF-8
encoding:

```ant
Encoding.toHex(String.toBytes("é"))  // "c3a9"
Encoding.toHex(String.toBytes("🐜")) // "f09f909c"
```

The parameter is named `bytes`:

```ant
Encoding.toHex(bytes: String.toBytes("Ant")) // "416e74"
```

In a pipe, the bytes are the first argument:

```ant
"abc" | String.toBytes() | Encoding.toHex() // "616263"
```

A function value works wherever a `func(Bytes): string` is expected:

```ant
let hex: func(Bytes): string = Encoding.toHex
[String.toBytes("a"), String.toBytes("bc")].map(hex) // ["61", "6263"]
```

## Applications

### Hex digests

[Crypto](../Crypto/index.md) hashes return `Bytes`. Encode the digest to print or store it:

```ant
let digest = "abc" | String.toBytes() | Crypto.sha256()
Encoding.toHex(digest) // "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
```

### Random identifiers

Sixteen random bytes make a 32-character token:

```ant
let token = Crypto.randomBytes(16) | Encoding.toHex()
println(token.length())
```

```text
32
```

### CSS colors

Three bytes for red, green, and blue give a `#rrggbb` color:

```ant
func cssColor(r: int, g: int, b: int): string {
  "#" + Encoding.toHex(Bytes.fromArray([r, g, b]))
}

println(cssColor(255, 128, 0))
println(cssColor(10, 20, 30))
```

```text
#ff8000
#0a141e
```

### A number with a fixed width

[String.fromInt](../String/fromInt.md) with `base: 16` drops leading zeros. To write all the bytes
of a `uint32`, append it to `Bytes` in big-endian order and encode those:

```ant
let b = Bytes.new()
b.append(type(uint32), uint32(0xbeef), Bytes.BigEndian{})
println(Encoding.toHex(b))
println(String.fromInt(0xbeef, base: 16))
```

```text
0000beef
beef
```

### A hex dump

Print the offset and the hex of each row of bytes:

```ant
func hexDump(data: Bytes, width: int): void {
  for row in Iter.range(0, (data.length() + width - 1) / width) {
    let at = row * width
    let line = data.slice(at, Math.min(at + width, data.length())) | Encoding.toHex()
    println("${String.fromInt(at, base: 16).padStart(4, "0")}  ${line}")
  }
}

hexDump(String.toBytes("Hello, hex dump!!"), 8)
```

```text
0000  48656c6c6f2c2068
0008  65782064756d7021
0010  21
```

## Properties and relations

[Encoding.fromHex](fromHex.md) undoes `Encoding.toHex` for any bytes:

```ant
for b in [Bytes.new(), Bytes.fromArray([0, 127, 255]), String.toBytes("🐜")] {
  assert(Result.unwrap(Encoding.fromHex(Encoding.toHex(b))) == b)
}
```

The result has two characters per byte and is already lower case:

```ant
let b = Crypto.randomBytes(20)
let hex = Encoding.toHex(b)
assert(hex.length() == 2 * b.length())
assert(hex == hex.toLower())
```

Encoding a concatenation joins the encodings of the parts:

```ant
let parts = [String.toBytes("ab"), Bytes.new(2), String.toBytes("z")]
let joined = parts.map(Encoding.toHex) | String.join("")
assert(Encoding.toHex(Bytes.concat(parts)) == joined)
```

## Possible issues

`toHex` belongs to `Encoding`, not `Bytes`, so a `.` call on `Bytes` does not find it. Use the
namespace form or a pipe:

```ant
let b = String.toBytes("hi")
println(b.toHex())
// error: Bytes has no field or function toHex
```

```ant
let b = String.toBytes("hi")
b | Encoding.toHex() // "6869"
```

`toHex` takes `Bytes`, not a string. Convert text with [String.toBytes](../String/toBytes.md):

```ant
println(Encoding.toHex("hi"))
// error: cannot use string where Bytes is expected
```

The output is lower case, while [Encoding.fromHex](fromHex.md) accepts either case. Comparing hex
strings from another source can then fail on case alone. Compare the decoded bytes, or lower-case
the other string first:

```ant
let digest = "abc" | String.toBytes() | Crypto.sha256()
let expected = "BA7816BF8F01CFEA414140DE5DAE2223B00361A396177A9CB410FF61F20015AD"
Encoding.toHex(digest) == expected                  // false
Encoding.toHex(digest) == expected.toLower()        // true
Result.unwrap(Encoding.fromHex(expected)) == digest // true
```

To check a received hex signature, such as a webhook HMAC, do not compare hex strings with `==`:
that takes time that depends on their contents. Decode the signature with
[Encoding.fromHex](fromHex.md) and compare the bytes with [Crypto.equal](../Crypto/equal.md), whose
time depends only on the lengths:

```ant
func verify(key: Bytes, body: Bytes, header: string): Result[bool, Error] {
  let sent = try Encoding.fromHex(header)
  Result.ok(Crypto.equal(sent, Crypto.hmacSha256(key, body)))
}

let key = String.toBytes("secret")
let body = String.toBytes("hello")
let header = "88aab3ede8d3adf94d26ab90d3bafd4a2083070c3bcce9c014ee04a443847c0b"
println(verify(key, body, header))
println(verify(key, String.toBytes("hello!"), header))
```

```text
Ok{ value: true }
Ok{ value: false }
```

`Bytes` are mutable, and the string reflects them at the moment of the call:

```ant
let b = Bytes.fromArray([1, 2])
let before = Encoding.toHex(b)
b.push(byte(3))
before            // "0102"
Encoding.toHex(b) // "010203"
```

## See also

[Encoding.fromHex](fromHex.md) · [Encoding.toBase64](toBase64.md) · [Encoding](index.md) ·
[String.toBytes](../String/toBytes.md) · [String.fromInt](../String/fromInt.md) ·
[Crypto.sha256](../Crypto/sha256.md) · [Crypto.equal](../Crypto/equal.md) ·
[Bytes](../Bytes/index.md)
