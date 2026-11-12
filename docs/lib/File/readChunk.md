<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.readChunk

```ant signature
func File.readChunk(f: File.Handle, max: int, ctx: Context | null = null): Result[Bytes | null, Error]
```

File.readChunk reads at least 1 and at most max bytes from f, returning as soon as there are any. It
returns null at end of file and is Err when max is below 1. It is the read for a stream whose pieces
have no fixed size.

<!-- end generated -->

## Details

- `File.readChunk` returns as soon as there are any bytes, so a chunk can be shorter than `max` even
  before the end. Use [File.readExactly](readExactly.md) when you need a given number of bytes.
- At the end of the input it is `Ok` with `null`.
- It shares the handle's read buffer with [File.readLine](readLine.md) and the other reads, so it
  continues where the last read stopped.
- With `ctx`, a read of a connection, a pipe, or a terminal gives up when the context is done, and
  loses nothing. A read of a regular file looks at the context only before it begins.

## Basic examples

Read a file at most 4 bytes at a time, until `File.readChunk` returns `null`:

<!-- test: file hello.txt
Hello, Ant
-->

```ant
let f = try File.open("hello.txt")
defer { File.close(f) }
while true {
  let chunk = try File.readChunk(f, 4) ?? break
  println(chunk)
}
```

```text
Bytes(48 65 6c 6c)
Bytes(6f 2c 20 41)
Bytes(6e 74 0a)
```

At the end of the file, the read is `Ok` with `null`:

<!-- test: file one.txt
x
-->

```ant
let f = try File.open("one.txt")
defer { File.close(f) }
println(File.readChunk(f, 10))
println(File.readChunk(f, 10))
```

```text
Ok{ value: Bytes(78 0a) }
Ok{ value: null }
```

## Scope

An empty file gives `null` on the first read:

<!-- test: file empty.txt
-->

```ant
let f = try File.open("empty.txt")
defer { File.close(f) }
println(f.readChunk(1024))
```

```text
Ok{ value: null }
```

The bytes come back as they are, including bytes that are not valid UTF-8:

```ant
try File.writeBytes("raw.bin", Bytes.fromArray([0xff, 0x00, 0x41]))
let f = try File.open("raw.bin")
defer { File.close(f) }
println(try f.readChunk(16))
```

```text
Bytes(ff 00 41)
```

A chunk read continues after a line read on the same handle. Read a header line as text, then the
body as bytes:

<!-- test: file message.txt
SIZE 3
abc
-->

```ant
let f = try File.open("message.txt")
defer { File.close(f) }
println(try f.readLine())
println(try f.readChunk(100))
```

```text
SIZE 3
Bytes(61 62 63 0a)
```

On a connection, `File.readChunk` returns what has arrived and does not wait for `max` bytes. Here
the peer sends `ab`, and sends `cd` only after the first chunk is read:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { Net.close(listener) }
let next: Channel[bool] = Channel.new()
go {
  let peer = Net.accept(listener).unwrap()
  defer { File.close(peer) }
  peer.writeText("ab").unwrap()
  Channel.receive(next)
  peer.writeText("cd").unwrap()
}
let conn = try Net.dial(listener.address())
defer { File.close(conn) }
println(try conn.readChunk(4))
Channel.send(next, true)
println(try conn.readChunk(4))
println(conn.readChunk(4))
```

```text
Bytes(61 62)
Bytes(63 64)
Ok{ value: null }
```

Standard input is a handle too:

<!-- test: stdin
piped in
-->

```ant
println(try Os.stdin().readChunk(1024))
```

```text
Bytes(70 69 70 65 64 20 69 6e 0a)
```

## Named arguments

`ctx` bounds the wait for a connection that sends nothing. The read is `Err` with the context's
reason, and a later read still gets what arrives:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { Net.close(listener) }
go {
  let peer = Net.accept(listener).unwrap()
  defer { File.close(peer) }
  Time.sleep(Time.duration(milliseconds: 200))
  peer.writeLine("late").unwrap()
}
let conn = try Net.dial(listener.address())
defer { File.close(conn) }
let ctx = Context.new(timeout: Time.duration(milliseconds: 50))
println(conn.readChunk(100, ctx: ctx))
println(try conn.readChunk(100))
```

```text
Err{ error: Error{ message: "timeout" } }
Bytes(6c 61 74 65 0a)
```

A regular file never waits, but a context that is already done stops the read before it begins:

<!-- test: file data.txt
data
-->

```ant
let f = try File.open("data.txt")
defer { File.close(f) }
let ctx = Context.new()
ctx.cancel()
println(f.readChunk(4, ctx: ctx))
```

```text
Err{ error: Error{ message: "cancelled" } }
```

## Applications

Copy a file of any size through a small buffer with [File.writeChunk](writeChunk.md):

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

Count the bytes and the lines of standard input without holding it in memory:

<!-- test: stdin
one
two
three
-->

```ant
var bytes = 0
var lines = 0
let input = Os.stdin()
while true {
  let chunk = try input.readChunk(4096) ?? break
  bytes += chunk.length()
  lines += chunk.toArray().countWhere({ x == byte(10) })
}
println("${lines} lines, ${bytes} bytes")
```

```text
3 lines, 14 bytes
```

## Properties and relations

The chunks of a file, joined with [Bytes.concat](../Bytes/concat.md), are the bytes
[File.readBytes](readBytes.md) gives, and each chunk holds 1 to `max` bytes:

<!-- test: file text.txt
the quick brown fox
-->

```ant
let f = try File.open("text.txt")
defer { File.close(f) }
let chunks: Array[Bytes] = []
while true {
  let chunk = try f.readChunk(5) ?? break
  assert(chunk.length() >= 1 && chunk.length() <= 5)
  chunks.push(chunk)
}
assert(Bytes.concat(chunks) == try File.readBytes("text.txt"))
```

## Possible issues

A `max` below 1 is `Err`, not an empty chunk:

<!-- test: file small.txt
abc
-->

```ant
let f = try File.open("small.txt")
defer { File.close(f) }
println(f.readChunk(0))
```

```text
Err{ error: Error{ message: "the most to read is 0; it must be at least 1" } }
```

A chunk ends at a byte count, not at a character, so it can split a multi-byte character. Decoding
each chunk on its own then fails; collect the chunks and decode them together:

<!-- test: file accent.txt
café
-->

```ant
let f = try File.open("accent.txt")
defer { File.close(f) }
let chunks: Array[Bytes] = []
while true {
  let chunk = try f.readChunk(4) ?? break
  println("${chunk}: ${String.fromBytes(chunk)}")
  chunks.push(chunk)
}
println(String.fromBytes(Bytes.concat(chunks)))
```

```text
Bytes(63 61 66 c3): null
Bytes(a9 0a): null
café

```

## See also

[File.readExactly](readExactly.md) · [File.readLine](readLine.md) · [File.readAll](readAll.md) ·
[File.readBytes](readBytes.md) · [File.writeChunk](writeChunk.md) · [File.pipe](pipe.md) ·
[Net.dial](../Net/dial.md)
