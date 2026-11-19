<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Net.address

```ant signature
func Net.address(l: Net.Listener): string
```

Net.address returns the address l listens on, such as "127.0.0.1:50321".

<!-- end generated -->

## Details

- After [Net.listen](listen.md) with port 0, `Net.address` reports the port the listener took, so
  the result differs from run to run.

## Basic examples

Listen on any free port and read the port the listener took:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let addr = Net.address(listener)
println(addr.startsWith("127.0.0.1:"))
println(addr == "127.0.0.1:0")
```

```text
true
false
```

Connect to the listener at its address:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let conn = try Net.dial(Net.address(listener))
defer { conn.close() }
println("connected")
```

```text
connected
```

## Scope

An address with no host, such as `":0"`, works too. Either way the reported address ends in the port
the listener took, after the last colon:

```ant
func port(addr: string): int | null {
  let colon = addr.lastIndexOf(":")
  if colon == null { return null }
  addr.slice(colon + 1, addr.length()).toInt()
}

let local = try Net.listen("127.0.0.1:0")
defer { local.close() }
let every = try Net.listen(":0")
defer { every.close() }
for listener in [local, every] {
  let p = port(listener.address())
  println(p != null && p > 0)
}
```

```text
true
true
```

## Applications

### Talk to a server you started

A test or a script starts a server on a free port and hands its address to the client. Here the
server echoes one line back in upper case:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  let conn = Result.unwrap(listener.accept())
  defer { conn.close() }
  let line = Result.unwrap(conn.readLine()) ?? ""
  Result.unwrap(conn.writeLine(line.toUpper()))
}
let client = try Net.dial(listener.address())
defer { client.close() }
try client.writeLine("hello")
println(try client.readLine())
```

```text
HELLO
```

### Report where a server listens

A server that listens on port 0 learns its port only from `Net.address`, so it logs the address and
hands it to its clients. Here an [Http.serve](../Http/serve.md) server runs on a goroutine and a
client fetches its health check at the reported address:

```ant
func handle(req: Http.Request): Http.Reply {
  if req.path == "/health" {
    Http.Reply{ status: 200, headers: Map{}, body: "ok" }
  } else {
    Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
  }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
defer { stop.cancel() }
let addr = listener.address()
go { _ = Http.serve(listener, handle, stop) }
let reply = try Http.get("http://${addr}/health")
println(reply.text())
```

```text
ok
```

The output leaves the address out, since the port changes from run to run; a real server prints
`listening on ${addr}`.

## Properties and relations

The client's [Net.peer](peer.md) is the address the listener reports:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let client = try Net.dial(listener.address())
defer { client.close() }
assert(Net.peer(client) == listener.address())
println("ok")
```

```text
ok
```

Two listeners on port 0 take different ports, so their addresses differ:

```ant
let a = try Net.listen("127.0.0.1:0")
defer { a.close() }
let b = try Net.listen("127.0.0.1:0")
defer { b.close() }
assert(a.address() != b.address())
println("ok")
```

```text
ok
```

## Possible issues

The port is taken while the listener is open, so a second listener on the same address is `Err`:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
let again = Net.listen(listener.address())
println(again.isOk())
```

```text
false
```

Do not write a port 0 address where a client needs the real one. Pass the result of `Net.address`
instead of the string given to [Net.listen](listen.md):

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
println(Net.dial("127.0.0.1:0").isOk())
println(Net.dial(listener.address()).isOk())
```

```text
false
true
```

## See also

[Net.listen](listen.md), [Net.Listener](types/Listener.md), [Net.dial](dial.md),
[Net.accept](accept.md), [Net.peer](peer.md), [Net.close](close.md), [Net](index.md)
