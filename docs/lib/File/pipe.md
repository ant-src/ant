<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.pipe

```ant signature
func File.pipe(src: File.Handle, dst: File.Handle, ctx: Context | null = null): Result[void, Error]
```

File.pipe copies the rest of src to dst, writing each chunk as it is read, so a source that waits
streams. What it read has been written even when ctx ends the copy, which is then Err with ctx's
reason. src and dst may be the same connection, which echoes; for a file they may not, which is Err.
It closes neither.

<!-- end generated -->

## Details

- `File.pipe` starts where the reads of `src` stopped, first copying what an earlier read took in
  ahead, and copies to the end of `src`.
- The bytes are copied as they are, so `File.pipe` copies binary data as well as text.
- `ctx` ends a wait for a connection, a pipe, or a terminal. A regular file never waits, so `ctx`
  stops a copy from one only when it is done before a read begins.
- Add the file's name to an `Err` with `Result.wrapErr`.

## Basic examples

Copy a file to standard output:

<!-- test: file notes.txt
first
second
-->

```ant
let src = try File.open("notes.txt")
defer { src.close() }
try File.pipe(src, Os.stdout())
```

```text
first
second
```

Copy standard input to standard output, as `cat` does:

<!-- test: stdin
hello
from stdin
-->

```ant
try File.pipe(Os.stdin(), Os.stdout())
```

```text
hello
from stdin
```

## Scope

`File.pipe` copies the rest of `src`. After a `readLine`, it starts at the next line, and a second
pipe of the same handle copies nothing and is `Ok`:

<!-- test: file names.txt
Ann
Bob
Cid
-->

```ant
let src = try File.open("names.txt")
defer { src.close() }
println(try src.readLine())
try src.pipe(Os.stdout())
println(src.pipe(Os.stdout()))
```

```text
Ann
Bob
Cid
Ok{ value: void }
```

Bytes that are not text are copied unchanged:

```ant
try File.writeBytes("in.bin", Bytes.fromArray([0, 255, 10, 128]))
let src = try File.open("in.bin")
defer { src.close() }
let dst = try File.create("out.bin")
defer { dst.close() }
try src.pipe(dst)
println(try File.readBytes("out.bin"))
```

```text
Bytes(00 ff 0a 80)
```

The source can be a connection. The copy ends when the peer closes it:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  let peer = listener.accept().unwrap()
  defer { peer.close() }
  peer.writeLine("one").unwrap()
  peer.writeLine("two").unwrap()
}
let conn = try Net.dial(listener.address())
defer { conn.close() }
try conn.pipe(Os.stdout())
```

```text
one
two
```

## Named arguments

`ctx` ends a copy from a source that waits. What arrived before has been written:

```ant
let p = try Os.openPipe()
defer { p.read.close() }
defer { p.write.close() }
try p.write.writeLine("arrived in time")
let ctx = Context.new(Time.duration(milliseconds: 50))
println(p.read.pipe(Os.stdout(), ctx: ctx))
```

```text
arrived in time
Err{ error: Error{ message: "timeout" } }
```

A regular file never waits. A context that is already done stops the copy before it reads anything,
and the handle reads on from where it was:

<!-- test: file queue.txt
Ann
Bob
-->

```ant
let src = try File.open("queue.txt")
defer { src.close() }
let ctx = Context.new()
ctx.cancel()
println(src.pipe(Os.stdout(), ctx: ctx))
println(src.readLine())
```

```text
Err{ error: Error{ message: "cancelled" } }
Ok{ value: "Ann" }
```

## Applications

Serve a file over a connection, and save what arrives to a file on the other end:

<!-- test: file report.txt
total: 42
-->

```ant
func serve(listener: Net.Listener, path: string): Result[void, Error] {
  let peer = try listener.accept()
  defer { peer.close() }
  let src = try File.open(path)
  defer { src.close() }
  src.pipe(peer) | Result.wrapErr("send ${path}")
}

let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go { serve(listener, "report.txt").unwrap() }

let conn = try Net.dial(listener.address())
defer { conn.close() }
let dst = try File.create("received.txt")
defer { dst.close() }
try conn.pipe(dst)
print(try File.read("received.txt"))
```

```text
total: 42
```

Show a child's output as it comes, then wait for its status:

```ant
let out = try Os.openPipe()
defer { out.read.close() }
let child = try Cmd.start(Cmd.new("echo", ["from the child"]), stdout: out.write)
try out.write.close()
try out.read.pipe(Os.stdout())
println("status ${try child.wait()}")
```

```text
from the child
status 0
```

Append standard input to a log file:

<!-- test: stdin
login ann
logout ann
-->

```ant
func record(path: string): Result[void, Error] {
  let log = try File.append(path)
  defer { log.close() }
  try (Os.stdin().pipe(log) | Result.wrapErr("append to ${path}"))
  log.close()
}

try record("events.log")
print(try File.read("events.log"))
```

```text
login ann
logout ann
```

## Properties and relations

A pipe of a whole file gives what [File.read](read.md) reads, and leaves the source at its end:

<!-- test: file words.txt
alpha
beta
-->

```ant
let src = try File.open("words.txt")
defer { src.close() }
let dst = try File.create("copy.txt")
defer { dst.close() }
try src.pipe(dst)
assert(try File.read("copy.txt") == try File.read("words.txt"))
assert(try src.readLine() == null)
```

## Possible issues

A buffered `dst` holds what the pipe wrote until it is flushed or closed. This program prints
nothing:

<!-- test: file held.txt
lost
-->

```ant
let src = try File.open("held.txt")
defer { src.close() }
try src.pipe(File.buffered(Os.stdout()))
```

Flush or close the buffered handle after the pipe:

<!-- test: file held.txt
lost
-->

```ant
let src = try File.open("held.txt")
defer { src.close() }
let out = File.buffered(Os.stdout())
try src.pipe(out)
try out.flush()
```

```text
lost
```

The order of the handles matters: `src` is read and `dst` is written. A handle opened the other way,
or already closed, gives `Err`:

<!-- test: file text.txt
text
-->

```ant
let src = try File.open("text.txt")
defer { src.close() }
let dst = try File.create("out.txt")
println(File.pipe(dst, src))
try dst.close()
println(File.pipe(src, dst))
```

```text
Err{ error: Error{ message: "the source is not open for reading" } }
Err{ error: Error{ message: "the destination is not open for writing" } }
```

## See also

[File.open](open.md) · [File.readChunk](readChunk.md) · [File.writeChunk](writeChunk.md) ·
[File.buffered](buffered.md) · [File.copy](copy.md) · [Os.stdin](../Os/stdin.md) ·
[Os.openPipe](../Os/openPipe.md) · [Net.dial](../Net/dial.md) · [Cmd.start](../Cmd/start.md)
