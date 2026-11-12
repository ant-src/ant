<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.readBytes

```ant signature
func File.readBytes(path: string): Result[Bytes, Error]
```

File.readBytes reads the whole file at path as bytes, unchanged.

<!-- end generated -->

## Details

- The bytes come back exactly as the file holds them. Bytes that are not valid UTF-8 are kept, where
  [`File.read`](read.md) replaces them with U+FFFD.
- An empty file is `Ok` with empty `Bytes`.
- A missing file, a directory, and a file the program may not read are `Err`, and the message names
  the path.
- The whole file is read into memory at once. For a file too large for that, or a stream, open it
  with [`File.open`](open.md) and read it piece by piece with [`File.readChunk`](readChunk.md).

## Basic examples

<!-- test: file notes.txt
buy milk
-->

Read a file as bytes. `Bytes` prints as hexadecimal:

```ant
let bytes = try File.readBytes("notes.txt")
println(bytes)
println(bytes.length())
```

```text
Bytes(62 75 79 20 6d 69 6c 6b 0a)
9
```

A file that is not there is an `Err`:

```ant
println(File.readBytes("missing.bin"))
```

```text
Err{ error: Error{ message: "open missing.bin: no such file or directory" } }
```

## Scope

An empty file is empty `Bytes`, not an error:

```ant
try File.write("empty.txt", "")
println(File.readBytes("empty.txt"))
```

```text
Ok{ value: Bytes() }
```

Bytes that are not valid UTF-8 come back as they are. [`File.read`](read.md) turns them into U+FFFD:

```ant
try File.writeBytes("latin1.txt", Bytes.fromArray([0x63, 0x61, 0x66, 0xe9]))
println(File.readBytes("latin1.txt"))
println(File.read("latin1.txt"))
```

```text
Ok{ value: Bytes(63 61 66 e9) }
Ok{ value: "caf�" }
```

Index the result to read single bytes, or loop over it:

```ant
try File.writeBytes("header.bin", Bytes.fromArray([0x01, 0x00, 0xff]))
let header = try File.readBytes("header.bin")
println(header[0])
for b in header {
  println(b)
}
```

```text
1
1
0
255
```

A directory is an `Err`:

<!-- test: file logs/today.log
started
-->

```ant
println(File.readBytes("logs"))
```

```text
Err{ error: Error{ message: "read logs: is a directory" } }
```

Call it in a pipe, with the path first:

<!-- test: file notes.txt
buy milk
-->

```ant
let size = "notes.txt" | File.readBytes() | Result.map(Bytes.length)
println(size)
```

```text
Ok{ value: 9 }
```

## Applications

Tell a PNG image by the signature its first eight bytes hold:

```ant
let signature = Bytes.fromArray([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a])
try File.writeBytes("logo.png", Bytes.concat([signature, Bytes.fromArray([0, 0, 0, 13])]))

func isPng(path: string): Result[bool, Error] {
  let bytes = try File.readBytes(path)
  Result.ok(bytes.startsWith(signature))
}

println(isPng("logo.png"))
```

```text
Ok{ value: true }
```

Compute the SHA-256 checksum of a file, as `sha256sum` prints it:

<!-- test: file notes.txt
buy milk
-->

```ant
let digest = try File.readBytes("notes.txt")
  | Crypto.sha256()
  | Encoding.toHex()
println(digest)
```

```text
409baa381eaebfc8c71676ecb0eed6659ea7510b4b42f101b152c7f0696150c5
```

Check whether two files hold the same contents. `Bytes` compares by contents with `==`:

<!-- test: file a.txt
same
-->
<!-- test: file b.txt
same
-->
<!-- test: file c.txt
different
-->

```ant
func sameContents(a: string, b: string): Result[bool, Error] {
  Result.ok((try File.readBytes(a)) == (try File.readBytes(b)))
}

println(sameContents("a.txt", "b.txt"))
println(sameContents("a.txt", "c.txt"))
```

```text
Ok{ value: true }
Ok{ value: false }
```

Read a file as text, but refuse it when it is not valid UTF-8, instead of reading it with U+FFFD in
place of the bad bytes:

<!-- test: file good.txt
héllo
-->

```ant
func readStrict(path: string): Result[string, Error] {
  let bytes = try File.readBytes(path)
  let text = String.fromBytes(bytes)
  if text == null { return Result.fail("${path} is not valid UTF-8") }
  Result.ok(text)
}

try File.writeBytes("bad.txt", Bytes.fromArray([0x68, 0x69, 0xff]))
println(readStrict("good.txt"))
println(readStrict("bad.txt"))
```

```text
Ok{ value: "héllo\n" }
Err{ error: Error{ message: "bad.txt is not valid UTF-8" } }
```

Embed a small file in a data URL:

```ant
try File.writeBytes("dot.gif", Bytes.fromArray([0x47, 0x49, 0x46, 0x38, 0x39, 0x61]))
let url = "data:image/gif;base64,${Encoding.toBase64(try File.readBytes("dot.gif"))}"
println(url)
```

```text
data:image/gif;base64,R0lGODlh
```

## Properties and relations

What [`File.writeBytes`](writeBytes.md) writes, `File.readBytes` reads back unchanged:

```ant
let data = Bytes.fromArray([0, 1, 127, 128, 255])
try File.writeBytes("data.bin", data)
assert((try File.readBytes("data.bin")) == data)
```

[`File.read`](read.md) is `File.readBytes` decoded with
[`String.fromBytesLossy`](../String/fromBytesLossy.md):

```ant
try File.writeBytes("mixed.txt", Bytes.fromArray([0x6f, 0x6b, 0xfe, 0x0a]))
let bytes = try File.readBytes("mixed.txt")
assert(String.fromBytesLossy(bytes) == (try File.read("mixed.txt")))
```

For a file of valid UTF-8, the bytes are the UTF-8 encoding of what `File.read` gives:

<!-- test: file greeting.txt
¡hola!
-->

```ant
let text = try File.read("greeting.txt")
assert((try File.readBytes("greeting.txt")) == text.toBytes())
```

## Possible issues

To see the text the bytes hold, decode them with [`String.fromBytes`](../String/fromBytes.md), or
read the file with [`File.read`](read.md):

<!-- test: file notes.txt
buy milk
-->

```ant
let bytes = try File.readBytes("notes.txt")
print(String.fromBytes(bytes) ?? "not UTF-8")
```

```text
buy milk
```

`File.readBytes` holds the whole file in memory. To count the lines of a large file without that,
read it a chunk at a time and count the newline bytes. For only the size, [`File.stat`](stat.md)
gives it without reading the file:

<!-- test: file list.txt
buy milk
buy bread
-->

```ant
func countLines(path: string): Result[int, Error] {
  let f = try File.open(path)
  defer { File.close(f) }
  var lines = 0
  while true {
    let chunk = try File.readChunk(f, 4096)
    if chunk == null { break }
    for b in chunk {
      if b == byte(0x0a) { lines += 1 }
    }
  }
  Result.ok(lines)
}

println(countLines("list.txt"))
```

```text
Ok{ value: 2 }
```

## See also

[File.writeBytes](writeBytes.md) · [File.read](read.md) · [File.readChunk](readChunk.md) ·
[File.open](open.md) · [String.fromBytes](../String/fromBytes.md) ·
[Encoding.toHex](../Encoding/toHex.md) · [Crypto.sha256](../Crypto/sha256.md)
