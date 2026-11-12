<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.writeChunk

```ant signature
func File.writeChunk(f: File.Handle, bytes: Bytes): Result[void, Error]
```

File.writeChunk writes bytes to f as they are.

<!-- end generated -->

## Details

- The bytes go out exactly as they are: no encoding, no line end added. For text, use
  [File.writeText](writeText.md) or [File.writeLine](writeLine.md).
- Before it returns, `File.writeChunk` hands all of the bytes to the OS, or is `Err`; nothing needs
  a flush. On a handle from [File.buffered](buffered.md) the bytes go into its buffer instead.
- A write after [File.close](close.md) is `Err`.

## Basic examples

Write bytes to a new file, then read them back:

```ant
let f = try File.create("out.bin")
defer { File.close(f) }
try File.writeChunk(f, Bytes.fromArray([0x89, 0x50, 0x4e, 0x47]))
try f.close()
println(try File.readBytes("out.bin"))
```

```text
Bytes(89 50 4e 47)
```

Each call writes after the last one, so a file can be built a piece at a time:

```ant
let f = try File.create("parts.txt")
defer { File.close(f) }
for part in ["ant", "bee", "cat"] {
  try File.writeChunk(f, "${part}\n".toBytes())
}
try f.close()
print(try File.read("parts.txt"))
```

```text
ant
bee
cat
```

## Scope

Bytes that are not valid UTF-8 are written as they are:

```ant
let f = try File.create("raw.bin")
defer { File.close(f) }
try f.writeChunk(Bytes.fromArray([0xff, 0x00, 0x41]))
try f.close()
let bytes = try File.readBytes("raw.bin")
println(bytes)
println(String.fromBytes(bytes))
```

```text
Bytes(ff 00 41)
null
```

On a handle from [File.append](append.md), the bytes go after what the file holds:

<!-- test: file log.txt
first
-->

```ant
let f = try File.append("log.txt")
defer { File.close(f) }
try f.writeChunk("second\n".toBytes())
try f.close()
print(try File.read("log.txt"))
```

```text
first
second
```

Standard output is a handle too:

```ant
try Os.stdout().writeChunk("héllo\n".toBytes())
```

```text
héllo
```

A connection from [Net.dial](../Net/dial.md) is a handle, and the peer reads the bytes with
[File.readExactly](readExactly.md):

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { Net.close(listener) }
let received: Channel[Bytes | null] = Channel.new()
go {
  let peer = Net.accept(listener).unwrap()
  defer { File.close(peer) }
  Channel.send(received, peer.readExactly(3).unwrap())
}
let conn = try Net.dial(listener.address())
defer { File.close(conn) }
try conn.writeChunk(Bytes.fromArray([1, 2, 3]))
println(Channel.receive(received))
```

```text
Bytes(01 02 03)
```

On a handle from [File.buffered](buffered.md), the bytes wait in the buffer until a flush or a
close:

```ant
let out = File.buffered(try File.create("held.bin"))
defer { File.close(out) }
try out.writeChunk("held".toBytes())
println(try File.readBytes("held.bin") | Bytes.length())
try out.flush()
println(try File.readBytes("held.bin") | Bytes.length())
```

```text
0
4
```

## Applications

Copy a file of any size through a small buffer, with [File.readChunk](readChunk.md):

<!-- test: file source.txt
first line
second line
-->

```ant
func copyFile(from: string, to: string): Result[void, Error] {
  let src = try File.open(from)
  defer { File.close(src) }
  let dst = try File.create(to)
  defer { File.close(dst) }
  while true {
    let chunk = try src.readChunk(8) ?? break
    try dst.writeChunk(chunk)
  }
  dst.close()
}

try copyFile("source.txt", "copy.txt")
print(try File.read("copy.txt"))
```

```text
first line
second line
```

Write length-prefixed messages: one byte of length, then the message, so a message is at most 255
bytes. The reader takes the length with [File.readExactly](readExactly.md) and then exactly that
many bytes:

```ant
func writeMessage(f: File.Handle, text: string): Result[void, Error] {
  let body = text.toBytes()
  if body.length() > 255 {
    return Result.fail("message too long")
  }
  try f.writeChunk(Bytes.fromArray([body.length()]))
  f.writeChunk(body)
}

let out = try File.create("messages.bin")
defer { File.close(out) }
try writeMessage(out, "hi")
try writeMessage(out, "ant")
try out.close()

let input = try File.open("messages.bin")
defer { File.close(input) }
while true {
  let size = try input.readExactly(1) ?? break
  let body = try input.readExactly(int(size.toArray()[0])) ?? break
  println(String.fromBytes(body))
}
```

```text
hi
ant
```

## Properties and relations

`File.writeChunk` of a string's UTF-8 bytes writes what [File.writeText](writeText.md) writes for
the string:

```ant
let text = "naïve café"
let a = try File.create("a.txt")
defer { File.close(a) }
try a.writeText(text)
let b = try File.create("b.txt")
defer { File.close(b) }
try b.writeChunk(text.toBytes())
assert(try File.readBytes("a.txt") == try File.readBytes("b.txt"))
```

Writing chunks one after another gives the same file as writing their concatenation at once with
[File.writeBytes](writeBytes.md):

```ant
let chunks = ["one,".toBytes(), Bytes.fromArray([0x00]), "two".toBytes()]
let f = try File.create("chunks.bin")
defer { File.close(f) }
for chunk in chunks {
  try f.writeChunk(chunk)
}
try f.close()
try File.writeBytes("whole.bin", Bytes.concat(chunks))
assert(try File.readBytes("chunks.bin") == try File.readBytes("whole.bin"))
```

## Possible issues

A write after [File.close](close.md) is `Err`:

```ant
let f = try File.create("closed.bin")
try f.close()
println(f.writeChunk("late".toBytes()))
```

```text
Err{ error: Error{ message: "the file is not open for writing" } }
```

## See also

[File.writeText](writeText.md) · [File.writeLine](writeLine.md) · [File.writeBytes](writeBytes.md) ·
[File.readChunk](readChunk.md) · [File.readExactly](readExactly.md) · [File.buffered](buffered.md) ·
[File.pipe](pipe.md)
