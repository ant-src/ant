<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Net

Net holds TCP listeners and connections. A connection is a File.Handle, read and written with the
File functions.

<!-- end generated -->

`Net` is TCP in six functions. A server opens a [Net.Listener](types/Listener.md) with
[Net.listen](listen.md) and takes each client with [Net.accept](accept.md); a client connects with
[Net.dial](dial.md). Both ends get a `File.Handle`, so a connection is read and written with the
same `File` functions as a file: `readLine`, `readAll`, `readChunk`, `writeLine`, `writeText`,
`pipe`, and `close`. A server serves each connection on a goroutine of its own, so one slow client
does not hold up the others.

## Details

- A connection is a `File.Handle` like any other: a function that takes a handle reads a file, a
  pipe, or a connection alike. Close it with `defer` right after `accept` or `dial`.
- `Net.listen` on port 0 takes any free port, and [Net.address](address.md) reports which. The host
  may be left out, as in `":0"`.
- A read of a connection waits until data arrives, or until its `ctx` is done.
- Several goroutines may `accept` on one listener, each with a context of its own; each connection
  goes to one of them.
- A refused connection, a closed listener, and a done context are `Err`, not a panic. A second
  [Net.close](close.md) does nothing.
- TCP only: there is no TLS, UDP, or Unix socket. [Http](../Http/index.md) serves HTTP on a
  `Net.Listener`.

## Basic examples

Listen, accept a connection on a goroutine, and dial it from the main path:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { Net.close(listener) }
go {
  let conn = Result.unwrap(Net.accept(listener))
  defer { conn.close() }
  let name = Result.unwrap(conn.readLine()) ?? "nobody"
  Result.unwrap(conn.writeLine("Hello, ${name}"))
}

let conn = try Net.dial(Net.address(listener))
defer { conn.close() }
try conn.writeLine("Ann")
println(try conn.readLine())
```

```text
Hello, Ann
```

`Net.peer` gives the other end of a connection, and `null` for a handle that is not one:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let client = try Net.dial(listener.address())
defer { client.close() }
println(Net.peer(client) == listener.address())
println(Net.peer(Os.stdout()))
```

```text
true
null
```

## Scope

A function over `File.Handle` serves a file and a connection alike. This one counts lines until the
end of the input:

<!-- test: file notes.txt
one
two
three
-->

```ant
func lineCount(h: File.Handle): Result[int, Error] {
  var n = 0
  while (try h.readLine()) != null {
    n += 1
  }
  Result.ok(n)
}

let file = try File.open("notes.txt")
defer { file.close() }
println(try lineCount(file))

let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  let conn = Result.unwrap(Net.accept(listener))
  defer { conn.close() }
  Result.unwrap(conn.writeText("a\nb\n"))
}
let conn = try Net.dial(listener.address())
defer { conn.close() }
println(try lineCount(conn))
```

```text
3
2
```

A connection carries bytes as well as text. [File.pipe](../File/pipe.md) from a connection to itself
echoes everything back:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  let conn = Result.unwrap(Net.accept(listener))
  defer { conn.close() }
  Result.unwrap(conn.pipe(conn))
}

let conn = try Net.dial(listener.address())
defer { conn.close() }
try conn.writeChunk(Bytes.fromArray([0, 1, 255]))
println(try conn.readExactly(3))
```

```text
Bytes(00 01 ff)
```

Two goroutines accept on one listener, and each of the two connections goes to one of them:

```ant
func serveOne(listener: Net.Listener): Result[void, Error] {
  let conn = try listener.accept()
  defer { conn.close() }
  let word = try conn.readLine() ?? return Result.ok(void)
  conn.writeLine(word.toUpper())
}

func ask(addr: string, word: string): Result[string, Error] {
  let conn = try Net.dial(addr)
  defer { conn.close() }
  try conn.writeLine(word)
  let reply = try conn.readLine() ?? return Result.fail("no reply")
  Result.ok(reply)
}

let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go { serveOne(listener) }
go { serveOne(listener) }
let replies = [try ask(listener.address(), "ant"), try ask(listener.address(), "bee")]
println(replies)
```

```text
["ANT", "BEE"]
```

## Named arguments

`ctx` bounds a wait. One context can bound a whole exchange, from the dial to the last read. This
server accepts and then stays silent, so the read gives up with the context's reason:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  let conn = Result.unwrap(Net.accept(listener))
  defer { conn.close() }
  Time.sleep(1s)
}

let ctx = Context.new(50ms)
let conn = try Net.dial(listener.address(), ctx: ctx)
defer { conn.close() }
try conn.writeLine("hello")
println(conn.readLine(ctx: ctx))
```

```text
Err{ error: Error{ message: "timeout" } }
```

A context that is already cancelled makes `Net.accept` give up at once:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let ctx = Context.new()
ctx.cancel()
println(listener.accept(ctx: ctx))
```

```text
Err{ error: Error{ message: "cancelled" } }
```

## Functions by job

### Servers

- [Net.listen](listen.md): listen for TCP connections on `"host:port"`
- [Net.address](address.md): the address a listener listens on, with the port it took
- [Net.accept](accept.md): wait for the next connection
- [Net.close](close.md): stop listening, so pending and later accepts are `Err`

### Clients

- [Net.dial](dial.md): open a connection to `"host:port"`

### Connections

- [Net.peer](peer.md): the address of the other end of a connection
- [File.readLine](../File/readLine.md), [File.writeLine](../File/writeLine.md),
  [File.close](../File/close.md), and the other `File` functions read, write, and close it

### Types

- [Net.Listener](types/Listener.md): a listening TCP socket
- [File.Handle](../File/types/Handle.md): an open file, stream, or connection

## Applications

### A server and its clients

Accept in a loop and serve each connection on a goroutine of its own. The clients run at once with
[Parallel.map](../Parallel/map.md), two at a time, and each passes its scope context to its waits:

```ant
func shout(conn: File.Handle): Result[void, Error] {
  defer { conn.close() }
  let line = try conn.readLine() ?? return Result.ok(void)
  conn.writeLine(line.toUpper())
}

func ask(addr: string, word: string, ctx: Context): Result[string, Error] {
  let conn = try Net.dial(addr, ctx: ctx)
  defer { conn.close() }
  try conn.writeLine(word)
  let reply = try conn.readLine(ctx: ctx) ?? return Result.fail("${addr} closed without a reply")
  Result.ok(reply)
}

let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  while true {
    let conn = listener.accept()
    if conn is Err { break }
    go { shout(conn.value) }
  }
}

let words = ["ant", "colony", "queen", "worker"]
let replies = Parallel.map(words, func(word: string, work: Context): Result[string, Error] {
  ask(listener.address(), word, work)
}, limit: 2) ?? return
println(try (replies | Result.all()))
```

```text
["ANT", "COLONY", "QUEEN", "WORKER"]
```

### The first server that answers

[Parallel.any](../Parallel/any.md) dials several addresses at once and keeps the first connection.
Here the primary is down and the mirror answers:

```ant
let down = try Net.listen("127.0.0.1:0")
let primary = down.address()
down.close()
let mirror = try Net.listen("127.0.0.1:0")
defer { mirror.close() }

let dialed = Parallel.any([
  { Net.dial(primary, ctx: x) },
  { Net.dial(mirror.address(), ctx: x) },
]) ?? return
if dialed is Err {
  println("no server answered")
  return
}
let conn = dialed.value
defer { conn.close() }
println(Net.peer(conn) == mirror.address())
```

```text
true
```

## Properties and relations

The two ends agree: the client's peer is the listener's address, and what one end writes the other
reads, in order:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let client = try Net.dial(listener.address())
defer { client.close() }
let server = try listener.accept()
defer { server.close() }
assert(Net.peer(client) == listener.address())
for line in ["one", "two"] {
  try client.writeLine(line)
}
assert((try server.readLine()) == "one")
assert((try server.readLine()) == "two")
```

## Possible issues

A `go` block has no result, so `try` inside it is a compile error. Unwrap inside the goroutine, or
call a function that returns a `Result`:

```ant
let listener = try Net.listen("127.0.0.1:0")
go {
  let conn = try listener.accept() // error: try needs an enclosing function that returns a Result
}
```

A connection wrapped in [File.buffered](../File/buffered.md) holds a request until a flush, so the
server never sees it and the reply never comes. Flush before you read the reply:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  let conn = Result.unwrap(Net.accept(listener))
  defer { conn.close() }
  let line = Result.unwrap(conn.readLine()) ?? ""
  Result.unwrap(conn.writeLine("got ${line}"))
}

let conn = File.buffered(try Net.dial(listener.address()))
defer { conn.close() }
try conn.writeLine("hello")
println(conn.readLine(ctx: Context.new(50ms)))
try conn.flush()
println(try conn.readLine())
```

```text
Err{ error: Error{ message: "timeout" } }
got hello
```

An address needs a port, for a listener and for a dial alike:

```ant
println(Net.listen("8080").isOk())
println(Net.dial("127.0.0.1").isOk())
```

```text
false
false
```

## See also

[Net.listen](listen.md) · [Net.accept](accept.md) · [Net.dial](dial.md) · [Net.peer](peer.md) ·
[Net.address](address.md) · [Net.close](close.md) · [Net.Listener](types/Listener.md) ·
[File](../File/index.md) · [Http](../Http/index.md) · [Context](../Context/index.md) ·
[Parallel](../Parallel/index.md)
