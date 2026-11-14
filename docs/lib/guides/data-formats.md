<!-- status: reviewed -->

# Data formats

Five namespaces turn values into text and back. [Json](../Json/index.md) and [Csv](../Csv/index.md)
read and write structured data. [Encoding](../Encoding/index.md) writes `Bytes` as hex or base64
text and reads them back. [Crypto](../Crypto/index.md) hashes and signs `Bytes`.
[Html](../Html/index.md) is the markup that HTML expressions build and `String.from` renders.

```ant
Json.stringify(Map{ "ok": true })                   // "{\"ok\":true}"
Csv.stringify([["a", "b, c"]])                      // "a,\"b, c\"\n"
Csv.parse("a\tb\n1\t2", sep: "\t")                  // Ok{ value: [["a", "b"], ["1", "2"]] }
"hi".toBytes() | Encoding.toBase64()                // "aGk="
"hi".toBytes() | Crypto.sha256() | Encoding.toHex() // "8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4"
String.from(<b>{"Tom & Jerry"}</b>)                 // "<b>Tom &amp; Jerry</b>"
```

Each one works on values in memory. Read the text with [File.read](../File/read.md) or
[Http](../Http/index.md) first, and write the result the same way.

- [Json.parse](../Json/parse.md) gives a [Json.Value](../Json/types/Value.md) to narrow with `is`,
  with every number a `float64`. [Json.stringify](../Json/stringify.md) writes any value that has a
  JSON form, records included.
- Every CSV field is a `string`, both ways. Convert numbers yourself.
- Encoding and Crypto take `Bytes`. Turn text into bytes with
  [String.toBytes](../String/toBytes.md), and back with [String.fromBytes](../String/fromBytes.md).
- `Html` comes from HTML expressions, never from parsing a string. Text in it is escaped when it
  renders.
- The parsers and decoders return a `Result`: malformed input is an `Err`, not a panic.

## Functions by job

### JSON

- [Json.parse](../Json/parse.md): one JSON text as a `Json.Value`, or an `Err` that says where it
  went wrong
- [Json.stringify](../Json/stringify.md): the JSON text of any value, compact or indented with
  `indent:`
- [Json.Value](../Json/types/Value.md): a JSON value of unknown shape, narrowed with `is`

### CSV

- [Csv.parse](../Csv/parse.md): the rows of CSV text, each an array of string fields; another
  separator with `sep:`, such as `"\t"`
- [Csv.stringify](../Csv/stringify.md): CSV text for rows, quoting the fields that need it; another
  separator with `sep:`

### Bytes as text

- [Encoding.toHex](../Encoding/toHex.md): bytes as two lower-case hex digits each
- [Encoding.fromHex](../Encoding/fromHex.md): hex digits of either case back into bytes
- [Encoding.toBase64](../Encoding/toBase64.md): bytes as base64, standard or URL-safe with `url:`
- [Encoding.fromBase64](../Encoding/fromBase64.md): base64 back into bytes, with or without padding;
  URL-safe with `url:`
- [String.toBytes](../String/toBytes.md): the UTF-8 bytes of a string
- [String.fromBytes](../String/fromBytes.md): the string that UTF-8 bytes spell, or `null`

### Hashing

- [Crypto.sha256](../Crypto/sha256.md): the 32-byte SHA-256 digest
- [Crypto.sha512](../Crypto/sha512.md): the 64-byte SHA-512 digest
- [Crypto.sha1](../Crypto/sha1.md): the 20-byte SHA-1 digest, for formats that still use it
- [Crypto.md5](../Crypto/md5.md): the 16-byte MD5 digest, for formats that still use it

### Signing messages

- [Crypto.hmacSha256](../Crypto/hmacSha256.md): the HMAC-SHA-256 of a message under a key
- [Crypto.hmacSha512](../Crypto/hmacSha512.md): the HMAC-SHA-512
- [Crypto.hmacSha1](../Crypto/hmacSha1.md): the HMAC-SHA-1
- [Crypto.equal](../Crypto/equal.md): whether two byte sequences match, in time that does not depend
  on their contents

### Tokens and ids

- [Crypto.randomBytes](../Crypto/randomBytes.md): `n` bytes from the operating system's secure
  source
- [Crypto.uuid](../Crypto/uuid.md): a random version 4 UUID

### Markup

- [HTML expressions](../../language/literals/html.md): the syntax that builds `Html`, such as
  `<p>{name}</p>`
- [Html.unsafe](../Html/unsafe.md): a string of trusted markup as `Html`, rendered unescaped
- [Html.nodes](../Html/nodes.md): the top-level nodes of an `Html`, in order
- [Html.Node](../Html/types/Node.md): one node, an element, a text node, or raw markup
- [Html.Attribute](../Html/types/Attribute.md): one attribute of an element
- [Html.Attributes](../Html/types/Attributes.md): the map an attribute spread `{...attrs}` takes
- [String.from](../String/from.md): the markup of an `Html` as a string

## Applications

### Read a field from JSON

Parse the text, then narrow the `Json.Value` with `is` before using it. `is` takes `Map` without
type arguments. To decode into records, see
[decode JSON into records](../../howto/data/json-decode.md):

```ant
let doc = try Json.parse("{\"name\": \"Ann\"}")
if doc is Map && doc["name"] is string {
  println(doc["name"])
}
```

```text
Ann
```

### Convert CSV to JSON

Parse the rows, turn each one after the header into a record, and write the records as JSON.
[Result.all](../Result/all.md) gives the error of the first bad row:

```ant
type Person = { name: string, age: int }

func toPerson(row: Array[string]): Result[Person, Error] {
  let age = row.get(1)?.toInt() ?? return Result.fail("bad row ${row}")
  Result.ok(Person{ name: row[0], age })
}

let rows = try Csv.parse("name,age\nAnn,31\nBob,27")
let people = try (rows.slice(1, rows.length()).map(toPerson) | Result.all())
println(Json.stringify(people, indent: "  "))
```

```text
[
  {
    "name": "Ann",
    "age": 31
  },
  {
    "name": "Bob",
    "age": 27
  }
]
```

### Sign a JSON payload

Send the body with the hex HMAC of its bytes, and check a received signature with `Crypto.equal`:

```ant
type Event = { action: string, id: int }

func sign(secret: string, body: string): string {
  Crypto.hmacSha256(secret.toBytes(), body.toBytes()) | Encoding.toHex()
}

func verify(secret: string, body: string, signature: string): bool {
  let received = Encoding.fromHex(signature)
  if received is Err { return false }
  Crypto.equal(received.value, Crypto.hmacSha256(secret.toBytes(), body.toBytes()))
}

let body = Json.stringify(Event{ action: "opened", id: 7 })
let signature = sign("hook-secret", body)
println(body)
println(verify("hook-secret", body, signature))
println(verify("hook-secret", body.replace("7", "8"), signature))
```

```text
{"action":"opened","id":7}
true
false
```

### Carry a file in JSON

JSON has no form for `Bytes`. Write the bytes as base64 next to their digest, and check the digest
after decoding:

```ant
type Blob = { sha256: string, data: string }

func pack(data: Bytes): Blob {
  Blob{ sha256: data | Crypto.sha256() | Encoding.toHex(), data: data | Encoding.toBase64() }
}

func unpack(blob: Blob): Result[Bytes, Error] {
  let data = try (Encoding.fromBase64(blob.data) | Result.wrapErr("data"))
  if (data | Crypto.sha256() | Encoding.toHex()) != blob.sha256 {
    return Result.fail("sha256 does not match")
  }
  Result.ok(data)
}

let blob = pack(Bytes.fromArray([0x89, 0x50, 0x4e, 0x47]))
println(Json.stringify(blob))
println(unpack(blob))
println(unpack(Blob{ sha256: blob.sha256, data: "AAAA" }))
```

```text
{"sha256":"0f4636c78f65d3639ece5a064b5ae753e3408614a14fb18ab4d7540d2c248543","data":"iVBORw=="}
Ok{ value: Bytes(89 50 4e 47) }
Err{ error: Error{ message: "sha256 does not match" } }
```

### Render CSV as an HTML table

Build the table with HTML expressions. A field that holds markup renders as text:

```ant
func table(rows: Array[Array[string]]): Html {
  let head = rows[0].map({ <th>{x}</th> })
  let body = rows.slice(1, rows.length()).map({ <tr>{x.map({ <td>{x}</td> })}</tr> })
  <table><tr>{head}</tr>{body}</table>
}

let rows = try Csv.parse("item,note\ntea,\"<b>hot</b>\"")
println(table(rows))
```

```text
<table><tr><th>item</th><th>note</th></tr><tr><td>tea</td><td>&lt;b&gt;hot&lt;/b&gt;</td></tr></table>
```

### Give a page a version tag

A rendered page is a string, so its digest names its content, as an HTTP `ETag` does:

```ant
let page = <html><body><h1>Ants</h1></body></html>
let tag = String.from(page).toBytes() | Crypto.sha256() | Encoding.toHex()
tag.slice(0, 12) // "9fba62fe440b"
```

### Make a token and an id

```ant
let token = Crypto.randomBytes(24) | Encoding.toBase64(url: true)
token.length()         // 32
Crypto.uuid().length() // 36
```

## Possible issues

### JSON numbers come back as floats

```ant
Json.parse(Json.stringify(Map{ "n": 1 })) // Ok{ value: Map{ "n": 1.0 } }
```

Convert with [Math.toInt](../Math/toInt.md) where you need an `int`.

### `Bytes` and `Html` have no JSON form

`Json.stringify` panics on them. Write the base64 of the bytes and the rendered markup instead:

```ant
Json.stringify("hi".toBytes()) // panic: Json.stringify: Bytes has no JSON form; write Encoding.toBase64 of it
Json.stringify(<p>hi</p>)      // panic: Json.stringify: Html has no JSON form; write String.from of it
```

### CSV fields stay strings

Numbers in a CSV file compare as text until you convert them:

```ant
let row = Csv.parse("9,10").unwrap()[0]
row[0] < row[1]        // false
row.map({ x.toInt() }) // [9, 10]
```

### A string is not `Bytes`, and not markup

Convert text with `String.toBytes` before hashing or encoding it. A string never becomes `Html`;
interpolate it as text, or pass trusted markup through `Html.unsafe`:

```ant
let digest = Crypto.sha256("abc") // error: cannot use string where Bytes is expected
let h: Html = "<b>x</b>"          // error: cannot use string where Html is expected
```

### Base64 has two alphabets

`fromBase64` reads only the alphabet `url:` names, so a URL-safe token needs `url: true`:

```ant
Encoding.fromBase64("-_8")            // Err{ error: Error{ message: "illegal base64 data at input byte 0" } }
Encoding.fromBase64("-_8", url: true) // Ok{ value: Bytes(fb ff) }
```

### Encoding is not encryption

Anyone can decode hex and base64, and anyone can compute a digest of public data. Sign with an HMAC
under a secret key, and make secrets with `Crypto.randomBytes`, not with a seeded
[Random](../Random/index.md) generator.

## See also

- [HTML expressions guide](../../guide/html.md): how markup, components, and rendering fit together
- [Working with text](text.md): strings, bytes, and regular expressions
- [Files and paths](files.md): reading and writing the files that hold these formats
- [Network](network.md): sending and receiving them over HTTP
- How-to: [decode JSON into records](../../howto/data/json-decode.md),
  [encode JSON](../../howto/data/json-encode.md), [read and write CSV](../../howto/data/csv.md),
  [encode as base64 or hex](../../howto/data/base64-hex.md),
  [hash data and compare hashes](../../howto/data/hash.md),
  [generate random ids](../../howto/data/uuid.md)
