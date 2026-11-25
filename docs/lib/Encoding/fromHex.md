<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Encoding.fromHex

```ant signature
func Encoding.fromHex(s: string): Result[Bytes, Error]
```

Encoding.fromHex decodes s, hex digits of either case, into bytes. It is Err on an odd length or a
character that is not a hex digit.

<!-- end generated -->

## Details

- Upper-case and lower-case digits decode alike, and may be mixed.
- Every character must be a hex digit: a `0x` prefix, spaces, line breaks, and separators such as
  `:` make the result `Err`. Remove them first.

## Basic examples

```ant
Encoding.fromHex("6869")     // Ok{ value: Bytes(68 69) }
Encoding.fromHex("DEADbeef") // Ok{ value: Bytes(de ad be ef) }
Encoding.fromHex("abc")      // Err{ error: Error{ message: "odd length hex string", kind: null } }
```

## Scope

The empty string decodes to empty bytes:

```ant
Encoding.fromHex("") // Ok{ value: Bytes() }
```

A character that is not a hex digit is an error:

```ant
Encoding.fromHex("1g") // Err{ error: Error{ message: "invalid byte: U+0067 'g'", kind: null } }
```

A `0x` prefix, a space, and a letter other than `a` to `f` are all errors:

```ant
Encoding.fromHex("0x12").isOk()  // false
Encoding.fromHex("12 34").isOk() // false
Encoding.fromHex("é0").isOk()    // false
```

Two digits make one byte, so the result has half as many bytes as the string has digits:

```ant
let b = Encoding.fromHex("00ff7f") | Result.unwrap()
b.length() // 3
b[1]       // 255
b[2]       // 127
```

The parameter is named `s`:

```ant
Encoding.fromHex(s: "ff") // Ok{ value: Bytes(ff) }
```

In a pipe, the string is the first argument:

```ant
"cafe" | Encoding.fromHex() | Result.map({ x.length() }) // Ok{ value: 2 }
```

## Applications

### Check a digest

Compare data's SHA-256 with a hex digest published next to it, such as a download's checksum. Decode
the published digest and compare bytes, not strings, so that the case of its digits does not matter:

```ant
func verify(data: Bytes, digest: string): Result[bool, Error] {
  let want = try (Encoding.fromHex(digest) | Result.wrapErr("digest"))
  Result.ok(Crypto.equal(Crypto.sha256(data), want))
}

let data = "hello".toBytes()
let published = "2CF24DBA5FB0A30E26E83B2AC5B9E29E1B161E5C1FA7425E73043362938B9824"
println(verify(data, published))
println(verify(data, "2cf2"))
println(verify(data, "not hex"))
```

```text
Ok{ value: true }
Ok{ value: false }
Err{ error: Error{ message: "digest: invalid byte: U+006E 'n'", kind: null } }
```

### Parse a color

Read a CSS color such as `#ff8800` into its three channels:

```ant
type Rgb = { r: byte, g: byte, b: byte }

func parseColor(s: string): Result[Rgb, Error] {
  if !s.startsWith("#") || s.length() != 7 {
    return Result.fail("color ${s}: want #rrggbb")
  }
  let bytes = try (Encoding.fromHex(s.trimPrefix("#")) | Result.wrapErr("color ${s}"))
  Result.ok(Rgb{ r: bytes[0], g: bytes[1], b: bytes[2] })
}

println(parseColor("#ff8800"))
println(parseColor("#FFF"))
println(parseColor("#12345z"))
```

```text
Ok{ value: Rgb{ r: 255, g: 136, b: 0 } }
Err{ error: Error{ message: "color #FFF: want #rrggbb", kind: null } }
Err{ error: Error{ message: "color #12345z: invalid byte: U+007A 'z'", kind: null } }
```

### Hex with separators

A MAC address writes its bytes with `:` between them. Remove the separators, then decode:

```ant
let mac = "00:1A:2b:3C:4d:5E"
let b = mac.replace(":", "") | Encoding.fromHex() | Result.unwrap()
println(b.length())
println(Encoding.toHex(b))
```

```text
6
001a2b3c4d5e
```

### Hex-encoded text

Decode hex to bytes, then the bytes to a string with [String.fromBytes](../String/fromBytes.md):

```ant
Encoding.fromHex("416e74") | Result.unwrap() | String.fromBytes() // "Ant"
```

## Properties and relations

[Encoding.toHex](toHex.md) writes lower-case digits, so a round trip gives back the string in lower
case:

```ant
for s in ["", "00", "deadbeef", "DEADBEEF", "0aFf"] {
  assert((Encoding.fromHex(s) | Result.unwrap() | Encoding.toHex()) == s.toLower())
}
```

Decoding what `Encoding.toHex` wrote gives back the same bytes:

```ant
for b in ["", "ant", "🐜"].map(String.toBytes) {
  assert((Encoding.toHex(b) | Encoding.fromHex() | Result.unwrap()) == b)
}
```

A string of an odd length is always `Err`, even when every character is a digit:

```ant
for s in ["0", "abc", "12345"] {
  assert(!Encoding.fromHex(s).isOk())
}
```

## Possible issues

The result is a `Result`, not `Bytes`. Unwrap it, or use `try` in a function that returns a
`Result`:

```ant
let b: Bytes = Encoding.fromHex("ff")
// error: cannot use Result[Bytes, Error] where Bytes is expected
```

A leading zero matters. A number written in hex without it, such as `f`, has an odd length; pad it
to an even length first:

```ant
Encoding.fromHex("f").isOk()                // false
Encoding.fromHex("f".padStart(2, pad: "0")) // Ok{ value: Bytes(0f) }
```

A `0x` prefix and surrounding whitespace are not hex digits. Trim them before decoding:

```ant
Encoding.fromHex(" 0x1f\n").isOk()                  // false
Encoding.fromHex(" 0x1f\n".trim().trimPrefix("0x")) // Ok{ value: Bytes(1f) }
```

## See also

[Encoding.toHex](toHex.md), [Encoding.fromBase64](fromBase64.md), [Encoding](index.md),
[String.fromBytes](../String/fromBytes.md), [Bytes.fromArray](../Bytes/fromArray.md),
[Crypto.sha256](../Crypto/sha256.md), [Crypto.equal](../Crypto/equal.md)
