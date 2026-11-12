<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.Handle

```ant signature
type File.Handle
```

File.Handle is an opaque handle to an open file, a standard stream, a network connection, or a
buffer over one ([`File.buffered`](../buffered.md)). The reads of a handle share one buffer, so a
program can read lines and then bytes from it. A write is handed to the OS before it returns, all of
it or Err, except a write to a buffered handle. A read of a handle opened only for writing, and a
write to one opened only for reading, is Err. One goroutine may read a handle while another writes
it.

<!-- end generated -->

## Details

- A handle comes only from these functions: [`File.open`](../open.md),
  [`File.create`](../create.md), [`File.append`](../append.md), [`File.lock`](../lock.md),
  [`File.tryLock`](../tryLock.md), [`File.buffered`](../buffered.md),
  [`Os.stdin()`](../../Os/stdin.md), [`Os.stdout()`](../../Os/stdout.md),
  [`Os.stderr()`](../../Os/stderr.md), [`Os.openPipe`](../../Os/openPipe.md),
  [`Net.accept`](../../Net/accept.md), and [`Net.dial`](../../Net/dial.md). It has no literal.
- Every read and write of the `File` namespace works on every kind of handle, so a function that
  takes a `File.Handle` serves a file, a standard stream, a pipe, and a connection alike.
- One goroutine may read a handle while another writes it, and `File.close` from any goroutine ends
  a read or write in progress with `Err`.
- A handle opened for reading is `Err` on a write, and one opened for writing is `Err` on a read.
  `File.open` opens for reading, `File.create` and `File.append` for writing, and `File.lock` and
  `File.tryLock` for both.
- [`File.close`](../close.md) ends a handle. A second close, and a close of a standard stream, does
  nothing and is `Ok`. The handle is closed even when the close is `Err`. A read or write after the
  close is `Err`.
- Two handles are equal only when they are the same handle. `String.from` renders every handle as
  `File.Handle`.

## Basic examples

Write a function over `File.Handle`, and pass it a file or standard input:

<!-- test: stdin
red
green
-->

```ant
func countLines(f: File.Handle): Result[int, Error] {
  var n = 0
  while try f.readLine() != null {
    n += 1
  }
  Result.ok(n)
}

Result.unwrap(File.write("names.txt", "Ann\nBob\nCid\n"))
let file = Result.unwrap(File.open("names.txt"))
defer { file.close() }
println(countLines(file))
println(countLines(Os.stdin()))
```

```text
Ok{ value: 3 }
Ok{ value: 2 }
```

Create a file, write to its handle, and close it:

```ant
let out = Result.unwrap(File.create("todo.txt"))
Result.unwrap(out.writeLine("buy milk"))
Result.unwrap(out.close())
println(File.read("todo.txt"))
```

```text
Ok{ value: "buy milk\n" }
```

## Scope

### Files

The reads of a handle share one buffer, so a program reads a header line and then the rest as bytes:

```ant
Result.unwrap(File.write("msg.txt", "GET /\nhello"))
let f = Result.unwrap(File.open("msg.txt"))
defer { f.close() }
println(f.readLine())
println(f.readChunk(100))
println(f.readLine())
```

```text
Ok{ value: "GET /" }
Ok{ value: Bytes(68 65 6c 6c 6f) }
Ok{ value: null }
```

### Standard streams

`Os.stdout()` and `Os.stderr()` are handles, so a parameter can default to one. Here `report` writes
to standard output unless the caller passes a file:

```ant
type Row = { name: string, score: int }

func report(rows: Array[Row], out: File.Handle = Os.stdout()): Result[void, Error] {
  for row in rows {
    try out.writeLine("${row.name.padEnd(6)}${row.score}")
  }
  Result.ok(void)
}

let rows = [Row{ name: "Ann", score: 12 }, Row{ name: "Bob", score: 7 }]
Result.unwrap(report(rows))
let file = Result.unwrap(File.create("scores.txt"))
Result.unwrap(report(rows, out: file))
Result.unwrap(file.close())
println(File.read("scores.txt"))
```

```text
Ann   12
Bob   7
Ok{ value: "Ann   12\nBob   7\n" }
```

### Pipes

[`Os.openPipe`](../../Os/openPipe.md) gives two handles: what is written to `write` is read from
`read`. Closing `write` ends the input of `read`:

```ant
let pipe = Result.unwrap(Os.openPipe())
go {
  Result.unwrap(pipe.write.writeLine("ping"))
  Result.unwrap(pipe.write.close())
}
println(pipe.read.readAll())
Result.unwrap(pipe.read.close())
```

```text
Ok{ value: "ping\n" }
```

### Connections

A TCP connection from `Net.dial` or `Net.accept` is a handle, read and written with the same
functions. This server answers each line in upper case:

```ant
let server = Result.unwrap(Net.listen("127.0.0.1:0"))
go {
  let conn = Result.unwrap(server.accept())
  defer { conn.close() }
  while true {
    let line = Result.unwrap(conn.readLine())
    if line == null { break }
    Result.unwrap(conn.writeLine(line.toUpper()))
  }
}

let conn = Result.unwrap(Net.dial(server.address()))
defer { conn.close() }
Result.unwrap(conn.writeLine("hello"))
println(conn.readLine())
Result.unwrap(conn.writeLine("bye"))
println(conn.readLine())
```

```text
Ok{ value: "HELLO" }
Ok{ value: "BYE" }
```

### Commands

[`Cmd.start`](../../Cmd/start.md) takes handles for the `stdin`, `stdout`, and `stderr` of the child
process. Here the output of `echo` goes to a file:

```ant
let log = Result.unwrap(File.create("ls.txt"))
defer { log.close() }
let p = Result.unwrap(Cmd.start(Cmd.new("echo", ["hi"]), stdout: log))
println(p.wait())
println(File.read("ls.txt"))
```

```text
Ok{ value: 0 }
Ok{ value: "hi\n" }
```

### Buffered handles

[`File.buffered`](../buffered.md) returns a new handle over another one. Its writes wait in a buffer
until a flush or a close:

```ant
let out = Result.unwrap(File.create("log.txt")).buffered()
Result.unwrap(out.writeLine("started"))
println(File.read("log.txt"))
Result.unwrap(out.close())
println(File.read("log.txt"))
```

```text
Ok{ value: "" }
Ok{ value: "started\n" }
```

### A handle or null

[`File.tryLock`](../tryLock.md) returns `File.Handle | null`, `null` when another handle holds the
lock:

```ant
let held = Result.unwrap(File.lock("app.lock"))
let second = Result.unwrap(File.tryLock("app.lock"))
println(second == null)
Result.unwrap(held.close())
```

```text
true
```

## Applications

Copy standard input to a file with [`File.pipe`](../pipe.md), which takes two handles:

<!-- test: stdin
one
two
-->

The deferred close covers early returns; the final `out.close()` reports a failed close, which a
deferred close cannot.

```ant
func save(path: string): Result[void, Error] {
  let out = try File.create(path)
  defer { out.close() }
  try File.pipe(Os.stdin(), out)
  out.close()
}

println(save("input.txt"))
println(File.read("input.txt"))
```

```text
Ok{ value: void }
Ok{ value: "one\ntwo\n" }
```

Read a length-prefixed message: a header line gives the size, and the body follows as bytes:

```ant
func readMessage(f: File.Handle): Result[Bytes | null, Error] {
  let header = try f.readLine()
  if header == null { return Result.ok(null) }
  let size = header.toInt()
  if size == null { return Result.fail("bad header: ${header}") }
  f.readExactly(size)
}

Result.unwrap(File.write("inbox.txt", "5\nhello3\nhey"))
let inbox = Result.unwrap(File.open("inbox.txt"))
defer { inbox.close() }
println(readMessage(inbox))
println(readMessage(inbox))
println(readMessage(inbox))
```

```text
Ok{ value: Bytes(68 65 6c 6c 6f) }
Ok{ value: Bytes(68 65 79) }
Ok{ value: null }
```

## Properties and relations

Handles compare by identity. Opening the same file twice gives two handles:

```ant
Result.unwrap(File.write("a.txt", "x"))
let a = Result.unwrap(File.open("a.txt"))
let b = Result.unwrap(File.open("a.txt"))
assert(a == a)
assert(a != b)
assert(String.from(a) == "File.Handle")
```

Buffering a handle that is already buffered with at least as many bytes returns that same handle:

```ant
let out = Result.unwrap(File.create("b.txt")).buffered(size: 8192)
assert(out.buffered() == out)
assert(out.buffered(size: 16384) != out)
Result.unwrap(out.close())
```

## Possible issues

A handle opened for reading cannot write, and one opened for writing cannot read:

```ant
Result.unwrap(File.write("r.txt", "data"))
let r = Result.unwrap(File.open("r.txt"))
println(r.writeLine("more"))
let w = Result.unwrap(File.create("w.txt"))
println(w.readLine())
```

```text
Err{ error: Error{ message: "the file is not open for writing" } }
Err{ error: Error{ message: "the file is not open for reading" } }
```

A closed handle stays closed. A write after the close is `Err`, and a second close is `Ok`:

```ant
let out = Result.unwrap(File.create("x.txt"))
Result.unwrap(out.close())
println(out.writeLine("late"))
println(out.close())
```

```text
Err{ error: Error{ message: "the file is closed" } }
Ok{ value: void }
```

`File.Handle` is not a record, so it has no literal:

```ant
let h = File.Handle{} // error: File.Handle is not a record type, so it has no literal
```

`is` cannot test a handle. In a union, test the other members instead:

```ant
func name(x: File.Handle | string): string {
  if x is File.Handle { "handle" } else { x }
  // error: is cannot test File.Handle; narrow by testing the other members
}
```

A handle has no JSON form, so `Json.stringify` panics on one:

```ant
let json = Json.stringify(Os.stdout()) // panic: Json.stringify: File.Handle has no JSON form
```

## See also

[File.open](../open.md), [File.create](../create.md), [File.close](../close.md),
[File.buffered](../buffered.md), [File.readLine](../readLine.md), [File.pipe](../pipe.md),
[Os.openPipe](../../Os/openPipe.md), [Os.stdin](../../Os/stdin.md), [Os.stdout](../../Os/stdout.md),
[Os.stderr](../../Os/stderr.md), [Net.accept](../../Net/accept.md), [Net.dial](../../Net/dial.md),
[Cmd.start](../../Cmd/start.md), [File.Info](Info.md)
