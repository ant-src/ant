<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.writeBytes

```ant signature
func File.writeBytes(path: string, bytes: Bytes, replace: bool = true): Result[void, Error]
```

File.writeBytes writes bytes as they are to the file at path, as [`File.write`](write.md) writes
text.

<!-- end generated -->

## Details

- It creates the file, or replaces what it held. A new file gets mode `0o666` less the process's
  umask, `0o644` under the usual umask `022`; a file that exists keeps its mode.
- With `replace: false` it writes only when nothing is at the path, and is otherwise `Err`, having
  changed nothing.

## Basic examples

Write bytes to a file, then read them back:

```ant
let data = Bytes.fromArray([0xca, 0xfe, 0xba, 0xbe])
println(File.writeBytes("magic.bin", data))
println(File.readBytes("magic.bin"))
```

```text
Ok{ value: void }
Ok{ value: Bytes(ca fe ba be) }
```

A path in a directory that does not exist is an `Err`:

```ant
println(File.writeBytes("missing/out.bin", Bytes.fromArray([1])))
```

```text
Err{ error: Error{ message: "open missing/out.bin: no such file or directory" } }
```

## Scope

Empty `Bytes` make an empty file:

```ant
try File.writeBytes("empty.bin", "".toBytes())
println(File.stat("empty.bin") | Result.map({ x.size }))
```

```text
Ok{ value: 0 }
```

Bytes that are not valid UTF-8 are written unchanged, which [`File.write`](write.md) cannot do,
since a `string` holds only valid UTF-8:

```ant
try File.writeBytes("latin1.txt", Bytes.fromArray([0x63, 0x61, 0x66, 0xe9]))
println(File.readBytes("latin1.txt"))
```

```text
Ok{ value: Bytes(63 61 66 e9) }
```

A file that exists is replaced, not appended to:

```ant
try File.writeBytes("state.bin", Bytes.fromArray([1, 2, 3, 4]))
try File.writeBytes("state.bin", Bytes.fromArray([9]))
println(File.readBytes("state.bin"))
```

```text
Ok{ value: Bytes(09) }
```

A directory is an `Err`:

```ant
try File.mkdir("out")
println(File.writeBytes("out", Bytes.fromArray([1])))
```

```text
Err{ error: Error{ message: "open out: is a directory" } }
```

## Named arguments

### replace

With `replace: false`, `File.writeBytes` never overwrites: it creates the file only when nothing is
at the path, and otherwise is `Err` and leaves the file as it was:

```ant
let first = Bytes.fromArray([1, 2])
let second = Bytes.fromArray([3, 4])
println(File.writeBytes("once.bin", first, replace: false))
println(File.writeBytes("once.bin", second, replace: false))
println(File.readBytes("once.bin"))
```

```text
Ok{ value: void }
Err{ error: Error{ message: "open once.bin: file exists" } }
Ok{ value: Bytes(01 02) }
```

## Applications

Decode a Base64 payload and save it as a file:

```ant
func savePayload(path: string, payload: string): Result[void, Error] {
  let bytes = try (Encoding.fromBase64(payload) | Result.wrapErr("decode ${path}"))
  File.writeBytes(path, bytes)
}

println(savePayload("dot.gif", "R0lGODlh"))
println(File.readBytes("dot.gif"))
println(savePayload("bad.bin", "not base64!"))
```

```text
Ok{ value: void }
Ok{ value: Bytes(47 49 46 38 39 61) }
Err{ error: Error{ message: "decode bad.bin: illegal base64 data at input byte 3" } }
```

Write a small binary header: a magic number, a version byte, and a count:

```ant
func writeHeader(path: string, version: int, count: int): Result[void, Error] {
  let magic = "ANT1".toBytes()
  let fields = Bytes.fromArray([version, count / 256, count % 256])
  File.writeBytes(path, Bytes.concat([magic, fields]))
}

try writeHeader("data.hdr", 2, 300)
println(File.readBytes("data.hdr"))
```

```text
Ok{ value: Bytes(41 4e 54 31 02 01 2c) }
```

Copy only the first bytes of a file into another:

```ant
try File.writeBytes("full.bin", Bytes.fromArray([1, 2, 3, 4, 5, 6]))
let head = try File.readBytes("full.bin")
try File.writeBytes("head.bin", head.slice(0, 3))
println(File.readBytes("head.bin"))
```

```text
Ok{ value: Bytes(01 02 03) }
```

Save a random key that a later run reads back. `replace: false` creates the key only when nothing is
at the path, so a second run keeps the first key. After an `Err`, `File.exists` tells an existing
key from a failed write:

```ant
func ensureKey(path: string): Result[Bytes, Error] {
  let written = File.writeBytes(path, Crypto.randomBytes(32), replace: false)
  if !File.exists(path) {
    try written
  }
  File.readBytes(path)
}

let a = try ensureKey("secret.key")
let b = try ensureKey("secret.key")
println(a.length())
println(a == b)
println(ensureKey("nope/secret.key"))
```

```text
32
true
Err{ error: Error{ message: "open nope/secret.key: no such file or directory" } }
```

## Properties and relations

What `File.writeBytes` writes, [`File.readBytes`](readBytes.md) reads back unchanged:

```ant
let data = Bytes.fromArray([0, 1, 127, 128, 255])
try File.writeBytes("data.bin", data)
assert((try File.readBytes("data.bin")) == data)
```

Writing a string's UTF-8 bytes gives the same file as [`File.write`](write.md) of the string:

```ant
let text = "naïve\n"
try File.write("a.txt", text)
try File.writeBytes("b.txt", text.toBytes())
assert((try File.readBytes("a.txt")) == (try File.readBytes("b.txt")))
```

A file that exists keeps its mode when it is replaced:

```ant
try File.writeBytes("run.sh", "#!/bin/sh\n".toBytes())
try File.chmod("run.sh", 0o755)
try File.writeBytes("run.sh", "#!/bin/sh\necho hi\n".toBytes())
assert((try File.stat("run.sh")).mode == 0o755)
```

## Possible issues

`File.writeBytes` returns a `Result`, and a call that drops it hides a failed write. Here the
program reports success although nothing was written:

```ant
File.writeBytes("nowhere/out.bin", Bytes.fromArray([1]))
println("saved")
```

```text
saved
```

Use `try` to stop and pass the error on:

```ant
func save(path: string, data: Bytes): Result[void, Error] {
  try File.writeBytes(path, data)
  println("saved ${path}")
  Result.ok(void)
}

println(save("nowhere/out.bin", Bytes.fromArray([1])))
```

```text
Err{ error: Error{ message: "open nowhere/out.bin: no such file or directory" } }
```

A reader can see the file half-written while `File.writeBytes` replaces it. To replace a file in one
step, write a temporary file beside it and rename it into place:

```ant
func saveAtomic(path: string, data: Bytes): Result[void, Error] {
  let tmp = try File.tempFile(".save-*", dir: Path.dir(path))
  defer { File.remove(tmp) }
  try File.writeBytes(tmp, data)
  File.rename(tmp, path)
}

println(saveAtomic("model.bin", Bytes.fromArray([7, 7, 7])))
println(File.readBytes("model.bin"))
```

```text
Ok{ value: void }
Ok{ value: Bytes(07 07 07) }
```

The file that `saveAtomic` leaves is `0o600`, the temporary file's mode, so `File.chmod` it after
the rename when others need to read it.

## See also

[File.readBytes](readBytes.md) · [File.write](write.md) · [File.writeChunk](writeChunk.md) ·
[File.create](create.md) · [File.rename](rename.md) · [String.toBytes](../String/toBytes.md) ·
[Bytes.fromArray](../Bytes/fromArray.md)
