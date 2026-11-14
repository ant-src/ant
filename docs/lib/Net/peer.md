<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Net.peer

```ant signature
func Net.peer(f: File.Handle): string | null
```

Net.peer returns the address of the other end of the connection f, as "ip:port", also once f is closed or through [`File.buffered`](../File/buffered.md). It is null for any other handle.

<!-- end generated -->

## Details

- On a connection from [Net.accept](accept.md), the peer is the client; on one from
  [Net.dial](dial.md), it is the address the connection reached.
- A closed connection keeps its peer, and a [File.buffered](../File/buffered.md) handle answers for
  the connection under it. Any other handle gives `null`.

## Basic examples

Ask an accepted connection where the client is:

```ant
let listener = try Net.listen("127.0.0.1:0")
let client = try Net.dial(listener.address())
defer { client.close() }
let conn = try Net.accept(listener)
defer { conn.close() }
let peer = Net.peer(conn) ?? "unknown"
println(peer.startsWith("127.0.0.1:"))
```

```text
true
```

A standard stream has no other end:

```ant
println(Net.peer(Os.stdout()))
```

```text
null
```

## Scope

Both ends of a connection have a peer:

```ant
let listener = try Net.listen("127.0.0.1:0")
let client = try Net.dial(listener.address())
defer { client.close() }
let conn = try Net.accept(listener)
defer { conn.close() }
println(Net.peer(client) != null)
println(Net.peer(conn) != null)
```

```text
true
true
```

A file opened with [File.open](../File/open.md) or [File.create](../File/create.md) is not a
connection, and neither is standard input:

<!-- test: file notes.txt
note
-->

```ant
let file = try File.open("notes.txt")
defer { file.close() }
println(Net.peer(file))
println(Net.peer(Os.stdin()))
```

```text
null
null
```

## Applications

Cut the port off to log the host a client connects from. `lastIndexOf` finds the last `:`, so an
IPv6 address keeps its own colons:

```ant
func host(addr: string): string {
  addr.slice(0, addr.lastIndexOf(":") ?? addr.length())
}

let listener = try Net.listen("127.0.0.1:0")
let client = try Net.dial(listener.address())
defer { client.close() }
let conn = try Net.accept(listener)
defer { conn.close() }
println("connection from ${host(Net.peer(conn) ?? "?")}")
```

```text
connection from 127.0.0.1
```

Serve only clients on the loopback address, and turn the rest away:

```ant
func isLocal(conn: File.Handle): bool {
  let peer = Net.peer(conn)
  if peer == null { return false }
  peer.startsWith("127.0.0.1:")
}

func greet(conn: File.Handle): Result[void, Error] {
  defer { conn.close() }
  if !isLocal(conn) { return conn.writeLine("denied") }
  conn.writeLine("welcome")
}

let listener = try Net.listen("127.0.0.1:0")
go { Net.accept(listener) | Result.unwrap() | greet() | Result.unwrap() }
let client = try Net.dial(listener.address())
defer { client.close() }
println(try client.readLine())
```

```text
welcome
```

Describe any handle, connection or not, with an `if` expression on the narrowed peer:

<!-- test: file data.txt
data
-->

```ant
func describe(f: File.Handle): string {
  let peer = Net.peer(f)
  if peer == null { "a local file or stream" } else { "a connection to ${peer}" }
}

let listener = try Net.listen("127.0.0.1:0")
let client = try Net.dial(listener.address())
defer { client.close() }
let file = try File.open("data.txt")
defer { file.close() }
println(describe(client) == "a connection to ${listener.address()}")
println(describe(file))
```

```text
true
a local file or stream
```

## Properties and relations

Both ends of a connection have a peer, and a file does not:

<!-- test: file notes.txt
note
-->

```ant
let listener = try Net.listen("127.0.0.1:0")
let client = try Net.dial(listener.address())
defer { client.close() }
let conn = try Net.accept(listener)
defer { conn.close() }
let file = try File.open("notes.txt")
defer { file.close() }
assert(Net.peer(client) != null && Net.peer(conn) != null)
assert(Net.peer(file) == null)
```

## Possible issues

`Net.peer` is in the `Net` namespace and a handle is a `File.Handle`, so a dot call does not find
it. Write it in namespace form:

<!-- test: before
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let conn = Result.unwrap(Net.dial(listener.address()))
-->

```ant
println(conn.peer()) // error: File.Handle has no field or function peer
```

The result may be `null`, so string functions do not apply to it until you narrow it:

<!-- test: before
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let conn = Result.unwrap(Net.dial(listener.address()))
-->

```ant
println(Net.peer(conn).startsWith("127.0.0.1:"))
// error: string | null has no field or function startsWith on every member: narrow the value first
```

## See also

[Net.accept](accept.md) · [Net.dial](dial.md) · [Net.address](address.md) · [Net.listen](listen.md)
· [Net](index.md) · [File.Handle](../File/types/Handle.md) ·
[Http.Request](../Http/types/Request.md)
