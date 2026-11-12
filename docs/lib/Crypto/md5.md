<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Crypto.md5

```ant signature
func Crypto.md5(bytes: Bytes): Bytes
```

Crypto.md5 returns the 16-byte MD5 digest of bytes. MD5 is broken against collisions and is here for
the formats that still use it.

<!-- end generated -->

## Details

- The digest is always 16 bytes, whatever the length of the input, the empty input included.
- MD5 is broken against collisions: anyone can make two inputs with the same digest. Use it to match
  a format that asks for MD5, such as a checksum file or a `Content-MD5` header, not to sign or
  protect data.
- The digest is `Bytes`. Turn it into text with `Encoding.toHex` or `Encoding.toBase64`.

## Basic examples

Hash the UTF-8 bytes of a string:

```ant
Crypto.md5(String.toBytes("ant")) // Bytes(63 b0 7e 82 8b f0 16 e9 76 ff 95 d6 ee 07 a1 05)
```

Write the digest as 32 hexadecimal digits, the form checksum tools print:

```ant
Crypto.md5(String.toBytes("ant")) | Encoding.toHex() // "63b07e828bf016e976ff95d6ee07a105"
```

## Scope

Empty input has a digest too, of the same length:

```ant
Crypto.md5(String.toBytes("")) | Encoding.toHex() // "d41d8cd98f00b204e9800998ecf8427e"
```

Text outside ASCII is hashed as its UTF-8 bytes:

```ant
Crypto.md5(String.toBytes("日本語")) | Encoding.toHex() // "00110af8b4393ef3f72c50be5b332bec"
```

Any bytes, not only text, such as bytes built from integers:

```ant
Crypto.md5(Bytes.fromArray([0, 255])) | Encoding.toHex() // "d07d34efac6328007ad67c7e0a985e00"
```

## Applications

Compute a `Content-MD5` header, which holds the digest of a body in Base64:

```ant
func contentMd5(body: string): string {
  String.toBytes(body) | Crypto.md5() | Encoding.toBase64()
}

contentMd5("ant") // "Y7B+govwFul2/5XW7gehBQ=="
```

Check a downloaded file against the digest in its `.md5` file, which holds the hex digest and then
the file name:

<!-- test: file colony.txt
ants: 12
-->
<!-- test: file colony.txt.md5
ed9dd10deb4c4cfaa5f00c7b729833a1  colony.txt
-->

```ant
func verify(path: string): Result[bool, Error] {
  let data = try (File.readBytes(path) | Result.wrapErr("read ${path}"))
  let line = try (File.read("${path}.md5") | Result.wrapErr("read ${path}.md5"))
  let expected = line.split(" ")[0]
  Result.ok((Crypto.md5(data) | Encoding.toHex()) == expected)
}

println(verify("colony.txt"))
```

```text
Ok{ value: true }
```

Spread keys over a fixed number of buckets, the same bucket on every run:

```ant
func bucket(key: string, n: int): int {
  let digest = Crypto.md5(String.toBytes(key))
  int(digest[0]) % n
}

bucket("ant", 4) // 3
bucket("ant", 4) // 3
bucket("bee", 4) // 1
```

Find files with the same content by grouping them on their digest:

```ant
type Upload = { name: string, data: Bytes }

let uploads = [
  Upload{ name: "a.txt", data: String.toBytes("hill") },
  Upload{ name: "b.txt", data: String.toBytes("nest") },
  Upload{ name: "c.txt", data: String.toBytes("hill") },
]
uploads
  | Array.groupBy({ Crypto.md5(x.data) | Encoding.toHex() })
  | Map.values()
  | Array.filter({ x.length() > 1 })
  | Array.map({ x.map({ x.name }) }) // [["a.txt", "c.txt"]]
```

## Properties and relations

The digest is 16 bytes, and its hex form 32 digits, for any input:

```ant
for s in ["", "ant", "日本語", "a".repeat(1000)] {
  let digest = Crypto.md5(String.toBytes(s))
  assert(digest.length() == 16)
  assert(Encoding.toHex(digest).length() == 32)
}
```

The same input gives the same digest, and a one-byte change gives another:

```ant
let a = Crypto.md5(String.toBytes("ant"))
assert(a == Crypto.md5(String.toBytes("ant")))
assert(a != Crypto.md5(String.toBytes("Ant")))
```

`Encoding.fromHex` turns a hex digest back into the digest:

```ant
let digest = Crypto.md5(String.toBytes("ant"))
assert(Encoding.fromHex(Encoding.toHex(digest)).unwrap() == digest)
```

## Possible issues

A string is not `Bytes`, so passing one is a compile error. Convert it with `String.toBytes`:

```ant
println(Crypto.md5("ant")) // error: cannot use string where Bytes is expected
```

Every byte counts: a trailing line break or space gives a different digest. Text read from a file
often ends in `"\n"`, so trim it when the digest is of the text without it:

```ant
Crypto.md5(String.toBytes("ant")) | Encoding.toHex()   // "63b07e828bf016e976ff95d6ee07a105"
Crypto.md5(String.toBytes("ant\n")) | Encoding.toHex() // "5f0e8e4883ce6fefa3ebeb7375c16da8"
```

`Encoding.toHex` writes lower case, so comparing it with an upper-case digest as strings fails.
Compare the bytes instead, since `Encoding.fromHex` reads either case:

```ant
let digest = Crypto.md5(String.toBytes("ant"))
let given = "63B07E828BF016E976FF95D6EE07A105"
(digest | Encoding.toHex()) == given       // false
Encoding.fromHex(given).unwrap() == digest // true
```

A digest, from MD5 or any other hash, proves nothing about who sent a message, since anyone can
compute the digest of a changed message. Sign with a key using `Crypto.hmacSha256`, and compare
signatures with `Crypto.equal`:

```ant
let key = String.toBytes("secret")
let message = String.toBytes("pay 10")
let signature = Crypto.hmacSha256(key, message)
Crypto.equal(signature, Crypto.hmacSha256(key, message)) // true
```

## See also

[Crypto.sha256](sha256.md) · [Crypto.sha1](sha1.md) · [Crypto.sha512](sha512.md) ·
[Crypto.hmacSha256](hmacSha256.md) · [Crypto.equal](equal.md) ·
[Encoding.toHex](../Encoding/toHex.md) · [Encoding.toBase64](../Encoding/toBase64.md) ·
[String.toBytes](../String/toBytes.md) · [Crypto](index.md)
