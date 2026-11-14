<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Net.Listener

```ant signature
type Net.Listener
```

Net.Listener is an opaque listening TCP socket, made by [`Net.listen`](../listen.md).

<!-- end generated -->

## Details

- Several goroutines may accept on one listener, and each connection goes to one of them.
- `Net.close` makes pending and later accepts `Err`. A second close does nothing.
  [`Http.serve`](../../Http/serve.md) serves HTTP on a listener and closes it when it returns.
- Two listeners are equal only when they are the same listener. `String.from` renders every listener
  as `Net.Listener`.

## Basic examples

Listen on a free port of the loopback address, accept one connection, and greet it:

```ant
let server = try Net.listen("127.0.0.1:0")
defer { Net.close(server) }
go {
  let conn = Result.unwrap(Net.accept(server))
  defer { conn.close() }
  Result.unwrap(conn.writeLine("hello"))
}

let conn = try Net.dial(Net.address(server))
defer { conn.close() }
println(try conn.readLine())
```

```text
hello
```

A listener renders as its type's name:

```ant
let server = try Net.listen("127.0.0.1:0")
defer { server.close() }
println(server)
println(server.address().startsWith("127.0.0.1:"))
```

```text
Net.Listener
true
```

## Scope

### As a parameter

Take a `Net.Listener` to write a server that does not choose its own address. This one answers each
line in upper case, for as many clients as `count` says:

```ant
func answer(conn: File.Handle): Result[void, Error] {
  defer { conn.close() }
  while true {
    let line = try conn.readLine()
    if line == null { break }
    try conn.writeLine(line.toUpper())
  }
  Result.ok(void)
}

func shout(l: Net.Listener, count: int): Result[void, Error] {
  for _ in Iter.range(0, count) {
    try answer(try l.accept())
  }
  Result.ok(void)
}

let server = try Net.listen("127.0.0.1:0")
defer { server.close() }
go { Result.unwrap(shout(server, 2)) }

for word in ["ant", "hill"] {
  let conn = try Net.dial(server.address())
  defer { conn.close() }
  try conn.writeLine(word)
  println(try conn.readLine())
}
```

```text
ANT
HILL
```

### In a record

A record can hold a listener next to what the server needs:

```ant
type Server = { name: string, listener: Net.Listener }

func open(name: string): Result[Server, Error] {
  let listener = try (Net.listen("127.0.0.1:0") | Result.wrapErr("open ${name}"))
  Result.ok(Server{ name, listener })
}

let api = try open("api")
defer { api.listener.close() }
println(api)
```

```text
Server{ name: "api", listener: Net.Listener }
```

### A listener or null

`Net.Listener | null` is an optional listener. Narrow it by testing `null`:

```ant
func stop(l: Net.Listener | null): string {
  if l == null { return "nothing to stop" }
  l.close()
  "stopped"
}

println(stop(null))
println(stop(try Net.listen("127.0.0.1:0")))
```

```text
nothing to stop
stopped
```

### Waiting with a context

`Net.accept` waits for a connection. Pass `ctx:` to give up after a time:

```ant
let server = try Net.listen("127.0.0.1:0")
defer { server.close() }
println(server.accept(ctx: Context.new(50ms)))
```

```text
Err{ error: Error{ message: "timeout" } }
```

### HTTP

[`Http.serve`](../../Http/serve.md) takes a listener, so the program knows the port before the first
request. The server stops when its context is done, and closes the listener:

```ant
func hello(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: "hello ${req.path}" }
}

let server = try Net.listen("127.0.0.1:0")
let stop = Context.new()
let served: Channel[Result[void, Error]] = Channel.new(1)
go { served.send(Http.serve(server, hello, stop)) }

let res = try Http.get("http://${server.address()}/ant")
println(res.text())
stop.cancel()
println(served.receive())
println(server.accept())
```

```text
hello /ant
Ok{ value: void }
Err{ error: Error{ message: "the listener is closed" } }
```

## Applications

Serve with a pool of workers. Every worker accepts on the same listener, and each connection goes to
one of them. Closing the listener ends the workers' waiting accepts:

```ant
func count(conn: File.Handle): Result[void, Error] {
  defer { conn.close() }
  let line = try conn.readLine() ?? ""
  conn.writeLine("${line}: ${line.length()}")
}

let server = try Net.listen("127.0.0.1:0")
let wg = WaitGroup.new()
wg.add(3)
for _ in Iter.range(0, 3) {
  go {
    defer { wg.done() }
    while true {
      let conn = server.accept()
      if conn is Err { break }
      Result.unwrap(count(conn.value))
    }
  }
}

for word in ["ant", "hill", "colony"] {
  let conn = try Net.dial(server.address())
  defer { conn.close() }
  try conn.writeLine(word)
  println(try conn.readLine())
}
server.close()
wg.wait()
```

```text
ant: 3
hill: 4
colony: 6
```

## Properties and relations

Listeners compare by identity. Two calls to `Net.listen` give two listeners on two ports:

```ant
let a = try Net.listen("127.0.0.1:0")
let b = try Net.listen("127.0.0.1:0")
defer { a.close() }
defer { b.close() }
assert(a == a)
assert(a != b)
assert(a.address() != b.address())
assert(String.from(a) == "Net.Listener")
```

Closing a listener twice is the same as closing it once:

```ant
let server = try Net.listen("127.0.0.1:0")
server.close()
server.close()
assert(server.accept() is Err)
```

## Possible issues

A listener holds its port until it is closed, so a second listener on the same address is `Err`:

```ant
let server = try Net.listen("127.0.0.1:0")
defer { server.close() }
let again = Net.listen(server.address())
println(again is Err)
```

```text
true
```

`Net.Listener` is not a record, so it has no literal:

```ant
let l = Net.Listener{} // error: Net.Listener is not a record type, so it has no literal
```

`is` cannot test a listener. In a union, test the other members instead:

```ant
func describe(x: Net.Listener | string): string {
  if x is Net.Listener { "listener" } else { x }
  // error: is cannot test Net.Listener; narrow by testing the other members
}
```

A union may hold at most one type with no discriminator: one of the sync types, `Random`, the
resource types, and the `Reflect` types. So a function cannot take `Net.Listener | File.Handle`:

```ant
func describe(x: Net.Listener | File.Handle): string { "x" } // error: a union may hold at most one type with no discriminator, and Net.Listener and File.Handle both have none
```

Take two parameters, or a record with a field for each, to carry both.

A listener has no JSON form, so `Json.stringify` panics on one:

```ant
let server = try Net.listen("127.0.0.1:0")
defer { server.close() }
let json = Json.stringify(server) // panic: Json.stringify: Net.Listener has no JSON form
```

## See also

[Net.listen](../listen.md) · [Net.address](../address.md) · [Net.accept](../accept.md) ·
[Net.close](../close.md) · [Net.dial](../dial.md) · [Http.serve](../../Http/serve.md) ·
[File.Handle](../../File/types/Handle.md)
