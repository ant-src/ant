<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.readExactly

```ant signature
func File.readExactly(f: File.Handle, n: int, ctx: Context | null = null): Result[Bytes | null, Error]
```

File.readExactly reads exactly n bytes from f, waiting until all of them have arrived. It returns
null at end of file, is Err with the message "unexpected end of file" when the input ends partway,
discarding the bytes it read, and is Err when n is below 0. An n of 0 gives empty Bytes without
reading, at the end of the file too.

<!-- end generated -->

## Details

- `n` counts bytes, not characters, and the value keeps the bytes as they are.
- `readExactly` is the read for a protocol that announces a length; for a stream whose pieces have
  no fixed size, [File.readChunk](readChunk.md) returns whatever has arrived, 1 to `max` bytes.
- All reads of a handle share one buffer, so `readExactly` can follow [File.readLine](readLine.md)
  and continues after the line.
- On a connection, a pipe or a terminal, `readExactly` waits until all `n` bytes have arrived. A
  context interrupts that wait, and the bytes read so far stay for the next read of the handle.

## Basic examples

Read a file in pieces of 4 bytes:

<!-- test: file data.bin
ABCDEFGH
-->

```ant
let f = try File.open("data.bin")
defer { File.close(f) }
println(File.readExactly(f, 4))
println(File.readExactly(f, 4))
```

```text
Ok{ value: Bytes(41 42 43 44) }
Ok{ value: Bytes(45 46 47 48) }
```

At the end of the file the value is `null`; when the file ends partway, it is `Err`:

<!-- test: file data.bin
ABCDEFGH
-->

```ant
let f = try File.open("data.bin")
defer { File.close(f) }
println(File.readExactly(f, 6))
println(File.readExactly(f, 6))
println(File.readExactly(f, 6))
```

```text
Ok{ value: Bytes(41 42 43 44 45 46) }
Err{ error: Error{ message: "unexpected end of file" } }
Ok{ value: null }
```

## Scope

An `n` of 0 gives empty `Bytes` without reading, at the end of the file too:

<!-- test: file empty.txt
-->

```ant
let f = try File.open("empty.txt")
defer { File.close(f) }
println(f.readExactly(1))
println(f.readExactly(0))
```

```text
Ok{ value: null }
Ok{ value: Bytes() }
```

Read a body whose length a header line announces, from standard input. The read after it continues
where the body stopped:

<!-- test: stdin
5
hello world
-->

```ant
func readBody(input: File.Handle): Result[Bytes, Error] {
  let header = try input.readLine() ?? return Result.fail("no header")
  let size = header.toInt() ?? return Result.fail("bad length: ${header}")
  let body = try input.readExactly(size) ?? return Result.fail("no body")
  Result.ok(body)
}

let input = Os.stdin()
let body = try readBody(input)
println(String.fromBytes(body))
print(try input.readAll())
```

```text
hello
 world
```

On a connection, `readExactly` waits for every byte, though they arrive in two writes:

```ant
let l = try Net.listen("127.0.0.1:0")
defer { Net.close(l) }
go {
  let c = Result.unwrap(Net.dial(l.address()))
  defer { File.close(c) }
  Result.unwrap(c.writeText("hel"))
  Time.sleep(20ms)
  Result.unwrap(c.writeText("lo"))
}
let conn = try Net.accept(l)
defer { File.close(conn) }
println(conn.readExactly(5))
```

```text
Ok{ value: Bytes(68 65 6c 6c 6f) }
```

## Named arguments

`ctx:` gives up waiting when the context is done. The value is `Err` with the context's reason, and
the bytes that had arrived stay for the next read:

```ant
let l = try Net.listen("127.0.0.1:0")
defer { Net.close(l) }
let more: Channel[bool] = Channel.new()
go {
  let c = Result.unwrap(Net.dial(l.address()))
  defer { File.close(c) }
  Result.unwrap(c.writeText("ab"))
  Channel.receive(more)
  Result.unwrap(c.writeText("cd"))
}
let conn = try Net.accept(l)
defer { File.close(conn) }
let ctx = Context.new(timeout: 50ms)
println(conn.readExactly(4, ctx: ctx))
Channel.send(more, true)
println(conn.readExactly(4))
```

```text
Err{ error: Error{ message: "timeout" } }
Ok{ value: Bytes(61 62 63 64) }
```

## Applications

Read messages that each start with a byte giving their length. A message of length 0 is empty, and a
file cut off inside a message is `Err`:

```ant
func readMessages(path: string): Result[Array[string], Error] {
  let f = try File.open(path)
  defer { File.close(f) }
  let messages: Array[string] = []
  while true {
    let header = try f.readExactly(1) ?? break
    let size = int(header[0])
    let body = try f.readExactly(size) ?? return Result.fail("a message has no body")
    messages.push(String.fromBytesLossy(body))
  }
  Result.ok(messages)
}

let hi = Bytes.concat([Bytes.fromArray([2]), "hi".toBytes()])
let ant = Bytes.concat([Bytes.fromArray([3]), "ant".toBytes()])
let empty = Bytes.fromArray([0])
let cut = Bytes.concat([Bytes.fromArray([5]), "cu".toBytes()])
try File.writeBytes("good.bin", Bytes.concat([hi, empty, ant]))
try File.writeBytes("cut.bin", Bytes.concat([hi, cut]))
println(readMessages("good.bin"))
println(readMessages("cut.bin"))
```

```text
Ok{ value: ["hi", "", "ant"] }
Err{ error: Error{ message: "unexpected end of file" } }
```

Read a file of fixed-size records, here 3-byte colors, and print each as hex:

```ant
try File.writeBytes("colors.bin", Bytes.fromArray([255, 0, 0, 0, 128, 255]))
let f = try File.open("colors.bin")
defer { File.close(f) }
while true {
  let rgb = try f.readExactly(3) ?? break
  println("#${Encoding.toHex(rgb)}")
}
```

```text
#ff0000
#0080ff
```

## Properties and relations

Reading a whole file of `n` bytes with `readExactly(n)` gives what [File.readBytes](readBytes.md)
gives, and the next read is `null`:

<!-- test: file data.bin
ABCDEFGH
-->

```ant
let size = (try File.stat("data.bin")).size
let f = try File.open("data.bin")
defer { File.close(f) }
assert(try f.readExactly(size) == try File.readBytes("data.bin"))
assert(try f.readExactly(1) == null)
```

## Possible issues

The bytes read before the input ends are discarded, not returned, and the next read starts after
them. A loop over fixed-size records cannot recover a short last record:

```ant
try File.writeBytes("records.bin", "ABCDEFGHIJ".toBytes())
let f = try File.open("records.bin")
defer { File.close(f) }
println(f.readExactly(4))
println(f.readExactly(4))
println(f.readExactly(4))
println(f.readChunk(4))
```

```text
Ok{ value: Bytes(41 42 43 44) }
Ok{ value: Bytes(45 46 47 48) }
Err{ error: Error{ message: "unexpected end of file" } }
Ok{ value: null }
```

`n` counts bytes, so a piece can end inside a character. Decode only whole messages:

```ant
try File.write("word.txt", "héllo")
let f = try File.open("word.txt")
defer { File.close(f) }
let two = try f.readExactly(2) ?? "".toBytes()
println(two)
println(String.fromBytes(two))
```

```text
Bytes(68 c3)
null
```

A negative `n` is `Err`:

<!-- test: file data.bin
ABCDEFGH
-->

```ant
let f = try File.open("data.bin")
defer { File.close(f) }
println(f.readExactly(-1))
```

```text
Err{ error: Error{ message: "the count to read is -1; it must be at least 0" } }
```

## See also

[File.readChunk](readChunk.md) · [File.readBytes](readBytes.md) · [File.readLine](readLine.md) ·
[File.writeChunk](writeChunk.md) · [File.open](open.md) · [Net.dial](../Net/dial.md) ·
[Bytes](../Bytes/index.md)
