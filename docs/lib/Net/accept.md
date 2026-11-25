<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Net.accept

```ant signature
func Net.accept(l: Net.Listener, ctx: Context | null = null): Result[File.Handle, Error]
```

Net.accept waits for the next connection on l and returns it as a handle. It is Err when l is
closed, or with ctx's reason when ctx is done. Several goroutines may accept on one listener, and
each connection goes to one of them.

<!-- end generated -->

## Details

- After [Net.close](close.md), a pending `Net.accept` and every later one are `Err`.
- The connection is a `File.Handle`: read and write it with the `File` functions, and close it when
  you are done.

## Basic examples

`Net.accept` returns the next connection as a handle:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  let client = Net.dial(listener.address()).unwrap()
  defer { client.close() }
  client.writeLine("hello").unwrap()
}
let conn = try Net.accept(listener)
defer { conn.close() }
println(try conn.readLine())
```

```text
hello
```

On a closed listener, `Net.accept` is `Err`:

```ant
let listener = try Net.listen("127.0.0.1:0")
listener.close()
println(Net.accept(listener))
```

```text
Err{ error: Error{ message: "the listener is closed", kind: null } }
```

## Scope

A connection that is already waiting is accepted at once. [Net.peer](peer.md) gives its other end,
the client's address:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let client = try Net.dial(listener.address())
defer { client.close() }
let conn = try listener.accept()
defer { conn.close() }
println(Net.peer(conn)?.startsWith("127.0.0.1:"))
```

```text
true
```

A `Net.accept` that is waiting when another goroutine closes the listener returns `Err`:

```ant
let listener = try Net.listen("127.0.0.1:0")
let result: Channel[Result[File.Handle, Error]] = Channel.new()
go { Channel.send(result, listener.accept()) }
Time.sleep(20ms)
listener.close()
println(Channel.receive(result))
```

```text
Err{ error: Error{ message: "the listener is closed", kind: null } }
```

Several goroutines may accept on one listener. Each connection goes to exactly one of them, so three
goroutines that accept once each serve three clients:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
for name in ["w1", "w2", "w3"] {
  go {
    let conn = listener.accept().unwrap()
    defer { conn.close() }
    conn.writeLine(name).unwrap()
  }
}
let replies: Array[string] = []
for _ in Iter.range(0, 3) {
  let client = try Net.dial(listener.address())
  defer { client.close() }
  replies.push(try client.readLine() ?? "")
}
println(replies.sort())
```

```text
["w1", "w2", "w3"]
```

## Named arguments

`ctx` bounds the wait. When the context is done before a client connects, `Net.accept` is `Err` with
its reason:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let ctx = Context.new(timeout: 50ms)
println(listener.accept(ctx: ctx))
```

```text
Err{ error: Error{ message: "timeout", kind: null } }
```

A cancelled context gives up at once, and the listener stays open for the next `accept`:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let ctx = Context.new()
ctx.cancel()
println(listener.accept(ctx: ctx))
let client = try Net.dial(listener.address())
defer { client.close() }
let conn = try listener.accept()
defer { conn.close() }
println(Net.peer(client) == listener.address())
```

```text
Err{ error: Error{ message: "cancelled", kind: null } }
true
```

## Applications

### A line server

Accept in a loop and serve each connection on a goroutine of its own, so a slow client does not hold
up the next one. The loop ends with the `Err` of the first `accept` that fails, here when the
listener is closed:

```ant
func shout(conn: File.Handle): Result[void, Error] {
  defer { conn.close() }
  while true {
    let line = try conn.readLine() ?? break
    try conn.writeLine(line.toUpper())
  }
  Result.ok(void)
}

func serve(listener: Net.Listener): Result[void, Error] {
  while true {
    let conn = try listener.accept()
    go { _ = shout(conn) }
  }
}

let listener = try Net.listen("127.0.0.1:0")
let stopped: Channel[Result[void, Error]] = Channel.new()
go { Channel.send(stopped, serve(listener)) }
for word in ["ant", "bee"] {
  let client = try Net.dial(listener.address())
  defer { client.close() }
  try client.writeLine(word)
  println(try client.readLine())
}
listener.close()
println(Channel.receive(stopped))
```

```text
ANT
BEE
Err{ error: Error{ message: "the listener is closed", kind: null } }
```

### Stop accepting after a deadline

Pass one context to every `accept` of a loop. The loop serves clients until the context is done, and
then returns its reason:

```ant
func greet(listener: Net.Listener, ctx: Context): Result[void, Error] {
  while true {
    let conn = try (listener.accept(ctx: ctx) | Result.wrapErr("accept"))
    defer { conn.close() }
    try conn.writeLine("welcome")
  }
}

let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  let client = Net.dial(listener.address()).unwrap()
  defer { client.close() }
  println(client.readLine().unwrap())
}
println(greet(listener, Context.new(timeout: 100ms)))
```

```text
welcome
Err{ error: Error{ message: "accept: timeout", kind: null } }
```

## Properties and relations

The two ends agree on the addresses: the client's peer is the listener's address, and the accepted
connection's peer is a port of the client's machine:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let client = try Net.dial(listener.address())
defer { client.close() }
let conn = try listener.accept()
defer { conn.close() }
assert(Net.peer(client) == listener.address())
assert(Net.peer(conn) != listener.address())
```

## Possible issues

The client sees the end of the input only when the server closes the connection. A server that
forgets to close it leaves the client's `readAll` waiting, here until its context is done:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  let conn = listener.accept().unwrap()
  conn.writeLine("bye").unwrap()
}
let client = try Net.dial(listener.address())
defer { client.close() }
println(client.readAll(ctx: Context.new(timeout: 50ms)))
```

```text
Err{ error: Error{ message: "timeout", kind: null } }
```

Add `defer { conn.close() }` right after `accept` to end the input on every path.

## See also

[Net.listen](listen.md) · [Net.dial](dial.md) · [Net.close](close.md) · [Net.peer](peer.md) ·
[Net.address](address.md) · [File.readLine](../File/readLine.md) · [Context.new](../Context/new.md)
