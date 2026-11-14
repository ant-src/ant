<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Net.listen

```ant signature
func Net.listen(addr: string): Result[Net.Listener, Error]
```

Net.listen listens for TCP connections on addr, written "host:port". A port of 0, as in ":0", takes any free port, which [`Net.address`](address.md) reports, and an empty host listens on every interface. An Err names the address: listen "127.0.0.1:x": unknown port.

<!-- end generated -->

## Details

- `Net.listen` returns at once: it opens the socket, and [Net.accept](accept.md) waits for each
  connection.
- It is `Err` when the address is malformed or the port is taken.
- `addr` is `host:port`. An empty host listens on every interface, and a port may be a number or a
  service name such as `http`.
- The listener stays open until [Net.close](close.md) closes it, or until
  [Http.serve](../Http/serve.md), which takes it, returns.
- An `Err` names the address, as in `listen "127.0.0.1:8080": address already in use`, so a
  wrap adds what the program was doing, not the address.

## Basic examples

`Net.listen(addr)` opens a listener on `"host:port"`. Port 0 takes any free port, and
[Net.address](address.md) tells which:

```ant
let listener = try Net.listen("127.0.0.1:0")
let address = Net.address(listener)
println(address.startsWith("127.0.0.1:"))
println(address != "127.0.0.1:0")
```

```text
true
true
```

A connection to the listener arrives through [Net.accept](accept.md) as a `File.Handle`:

```ant
let listener = try Net.listen("127.0.0.1:0")
let client = try Net.dial(Net.address(listener))
defer { client.close() }
let server = try Net.accept(listener)
defer { server.close() }
try client.writeLine("hello")
println(try server.readLine())
```

```text
hello
```

## Scope

A fixed port is reported as given:

```ant
let listener = try Net.listen("127.0.0.1:48217")
println(Net.address(listener))
```

```text
127.0.0.1:48217
```

The host may be empty, as in `":8080"`:

```ant
let listener = try Net.listen(":0")
println(Net.address(listener).endsWith(":0"))
```

```text
false
```

An IPv6 host goes in brackets:

```ant
let listener = try Net.listen("[::1]:0")
println(Net.address(listener).startsWith("[::1]:"))
```

```text
true
```

An address without a port is `Err`:

```ant
println(Net.listen("127.0.0.1"))
```

```text
Err{ error: Error{ message: "listen \"127.0.0.1\": missing port in address" } }
```

## Applications

### A line echo server

Accept connections on a goroutine and answer each line, while the same program talks to it as a
client:

```ant
func echo(conn: File.Handle): Result[void, Error] {
  defer { conn.close() }
  while true {
    let line = try conn.readLine()
    if line == null { return Result.ok(void) }
    try conn.writeLine(line.toUpper())
  }
}

func serve(listener: Net.Listener): Result[void, Error] {
  while true {
    let conn = try Net.accept(listener)
    go { echo(conn) }
  }
}

let listener = try Net.listen("127.0.0.1:0")
defer { Net.close(listener) }
go { serve(listener) }

let client = try Net.dial(Net.address(listener))
defer { client.close() }
for word in ["ping", "ant"] {
  try client.writeLine(word)
  println(try client.readLine())
}
```

```text
PING
ANT
```

### Serve HTTP

[Http.serve](../Http/serve.md) takes a listener, so the program knows the port before the first
request:

```ant
func hello(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: "hello ${req.path}" }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(listener, hello, stop) }
let reply = try Http.get("http://${Net.address(listener)}/ants")
println(reply.text())
stop.cancel()
```

```text
hello /ants
```

### Take the port from the environment

Read the port from `PORT`, fall back to 8080, and say what failed. The error already names the
address:

<!-- test: env PORT=48218 -->

```ant
let port = Os.env("PORT") ?? "8080"
let addr = "127.0.0.1:${port}"
let listener = try (Net.listen(addr) | Result.wrapErr("serve"))
println("listening on ${Net.address(listener)}")
```

```text
listening on 127.0.0.1:48218
```

### Fall back to a free port

Try a preferred address first and take any free port when it is busy:

```ant
func listenPreferred(addr: string): Result[Net.Listener, Error] {
  let preferred = Net.listen(addr)
  if preferred.isOk() { return preferred }
  Net.listen("127.0.0.1:0")
}

let busy = try Net.listen("127.0.0.1:48219")
let listener = try listenPreferred("127.0.0.1:48219")
println(Net.address(listener) == Net.address(busy))
```

```text
false
```

## Properties and relations

Each listen on port 0 takes a port of its own:

```ant
let a = try Net.listen("127.0.0.1:0")
let b = try Net.listen("127.0.0.1:0")
assert(Net.address(a) != Net.address(b))
```

After [Net.close](close.md), the port is free to listen on again:

```ant
let first = try Net.listen("127.0.0.1:48220")
Net.close(first)
let second = try Net.listen("127.0.0.1:48220")
assert(Net.address(second) == "127.0.0.1:48220")
```

## Possible issues

A port that another listener holds is `Err`, here unwrapped into a panic:

```ant
let first = Result.unwrap(Net.listen("127.0.0.1:48221"))
let second = Result.unwrap(Net.listen("127.0.0.1:48221"))
// panic: Result.unwrap on an Err: Error{ message: "listen \"127.0.0.1:48221\": address already in use" }
```

`Http.serve` closes its listener when it returns, so the listener cannot serve again:

```ant
let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
stop.cancel()
try Http.serve(listener, { Http.Reply{ status: 200, headers: Map{}, body: "" } }, stop)
println(Net.accept(listener))
```

```text
Err{ error: Error{ message: "the listener is closed" } }
```

## See also

[Net.address](address.md) · [Net.accept](accept.md) · [Net.close](close.md) · [Net.dial](dial.md) ·
[Net.Listener](types/Listener.md) · [Http.serve](../Http/serve.md)
