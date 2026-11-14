<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Net.dial

```ant signature
func Net.dial(addr: string, ctx: Context | null = null): Result[File.Handle, Error]
```

Net.dial opens a TCP connection to addr, written "host:port". The connection is a handle read and
written with the File functions.

<!-- end generated -->

## Details

- `addr` needs a port. The host is an IP address or a name, such as `localhost`.
- The connection is a `File.Handle`: read it with `readLine`, `readAll`, `readChunk`, or
  `readExactly`, write it with `writeLine`, `writeText`, or `writeChunk`, and close it with
  [File.close](../File/close.md). Close it with `defer` right after the dial.
- `Net.dial` is `Err` when it cannot connect, as when nothing listens at `addr`, and when `ctx` is
  done first, with the context's reason as the message.
- TCP only: there is no TLS, UDP, or Unix socket.

## Basic examples

`Net.dial(addr)` connects to a server. Here the server is a [Net.listen](listen.md) listener in the
same program, which reads a name and answers with a greeting:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  let peer = Result.unwrap(Net.accept(listener))
  defer { peer.close() }
  let name = Result.unwrap(peer.readLine()) ?? "nobody"
  Result.unwrap(peer.writeLine("Hello, ${name}"))
}

let conn = try Net.dial(listener.address())
defer { conn.close() }
try conn.writeLine("Ann")
println(try conn.readLine())
```

```text
Hello, Ann
```

An address without a port is `Err`:

```ant
println(Net.dial("127.0.0.1"))
```

```text
Err{ error: Error{ message: "dial tcp: address 127.0.0.1: missing port in address" } }
```

## Scope

The host can be a name. `localhost` reaches a listener on `127.0.0.1`:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let port = listener.address().cut(":")?.after ?? ""
go {
  let peer = Result.unwrap(Net.accept(listener))
  defer { peer.close() }
  Result.unwrap(peer.writeLine("reached on port ${port}"))
}

let conn = try Net.dial("localhost:${port}")
defer { conn.close() }
println((try conn.readLine()) == "reached on port ${port}")
```

```text
true
```

A connection carries bytes as well as text. This server echoes everything back with
[File.pipe](../File/pipe.md), and the client reads exactly the bytes it sent:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  let peer = Result.unwrap(Net.accept(listener))
  defer { peer.close() }
  Result.unwrap(peer.pipe(peer))
}

let conn = try Net.dial(listener.address())
defer { conn.close() }
try conn.writeChunk(Bytes.fromArray([0, 1, 255]))
println(try conn.readExactly(3))
```

```text
Bytes(00 01 ff)
```

## Named arguments

`ctx` gives up on connecting. A context that is already done makes `Net.dial` `Err` with its reason,
before it connects:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let ctx = Context.new()
ctx.cancel()
println(Net.dial(listener.address(), ctx: ctx))
```

```text
Err{ error: Error{ message: "cancelled" } }
```

A timeout bounds how long the dial waits for the server to answer; a listener on this machine
answers well within it:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let conn = try Net.dial(listener.address(), ctx: Context.new(timeout: 2s))
defer { conn.close() }
println("connected")
```

```text
connected
```

## Applications

Speak a text protocol by hand. This sends an HTTP request to an [Http.serve](../Http/serve.md)
server and reads the status line of the reply:

```ant
let listener = try Net.listen("127.0.0.1:0")
go {
  Http.serve(listener, { Http.Reply{ status: 200, headers: Map{}, body: "hi" } })
}

let conn = try Net.dial(listener.address())
defer { conn.close() }
try conn.writeText("GET / HTTP/1.0\r\nHost: localhost\r\n\r\n")
println(try conn.readLine())
```

```text
HTTP/1.0 200 OK
```

Check whether something listens at an address:

```ant
func reachable(addr: string): bool {
  let dialed = Net.dial(addr, ctx: Context.new(timeout: 1s))
  if dialed is Err { return false }
  dialed.value.close()
  true
}

let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
println(reachable(listener.address()))

let closed = try Net.listen("127.0.0.1:0")
let gone = closed.address()
closed.close()
println(reachable(gone))
```

```text
true
false
```

A client function that sends one command and returns the reply. `Result.wrapErr` says which step
failed:

```ant
func ask(addr: string, command: string): Result[string, Error] {
  let conn = try (Net.dial(addr) | Result.wrapErr("connect"))
  defer { conn.close() }
  try conn.writeLine(command)
  let reply = try conn.readLine() ?? return Result.fail("${addr} closed without a reply")
  Result.ok(reply)
}

let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  let peer = Result.unwrap(Net.accept(listener))
  defer { peer.close() }
  let command = Result.unwrap(peer.readLine()) ?? ""
  Result.unwrap(peer.writeLine(command.toUpper()))
}

println(ask(listener.address(), "ping"))
println(ask("localhost", "ping"))
```

```text
Ok{ value: "PING" }
Err{ error: Error{ message: "connect: dial tcp: address localhost: missing port in address" } }
```

Read a reply of several lines. This server sends a list and closes the connection, and the client
reads lines until `readLine` gives `null` at the end:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  let peer = Result.unwrap(Net.accept(listener))
  defer { peer.close() }
  for item in ["apples", "bread", "milk"] {
    Result.unwrap(peer.writeLine(item))
  }
}

let conn = try Net.dial(listener.address())
defer { conn.close() }
while true {
  let line = try conn.readLine() ?? break
  println("- ${line}")
}
```

```text
- apples
- bread
- milk
```

When the whole reply is wanted at once, `try conn.readAll()` reads everything up to the close.

## Properties and relations

[Net.peer](peer.md) of a dialed connection is the address the listener reports:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let conn = try Net.dial(listener.address())
defer { conn.close() }
assert(Net.peer(conn) == listener.address())
```

What one end writes, the other end reads, in order:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let conn = try Net.dial(listener.address())
defer { conn.close() }
let peer = try Net.accept(listener)
defer { peer.close() }
for line in ["one", "two"] {
  try conn.writeLine(line)
}
assert((try peer.readLine()) == "one")
assert((try peer.readLine()) == "two")
```

## Possible issues

`Net.dial` returns a `Result`, not the connection. Unwrap it with `try` before you use it:

```ant
let conn = Net.dial("127.0.0.1:8080")
conn.writeLine("hello")
// error: Result[File.Handle, Error] has no field or function writeLine
```

Nothing listening at the address is an `Err`, not a panic:

```ant
let closed = try Net.listen("127.0.0.1:0")
let gone = closed.address()
closed.close()
println(Net.dial(gone).isOk())
```

```text
false
```

A read waits until the server sends something. Give the read a context of its own when the server
may stay silent:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let conn = try Net.dial(listener.address())
defer { conn.close() }
println(conn.readLine(ctx: Context.new(timeout: 50ms)))
```

```text
Err{ error: Error{ message: "timeout" } }
```

## See also

[Net.listen](listen.md) · [Net.accept](accept.md) · [Net.peer](peer.md) ·
[File.readLine](../File/readLine.md) · [File.writeLine](../File/writeLine.md) ·
[File.close](../File/close.md) · [Http.request](../Http/request.md) ·
[Context.new](../Context/new.md)
