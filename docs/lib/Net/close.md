<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Net.close

```ant signature
func Net.close(l: Net.Listener): void
```

Net.close closes l, so pending and later [`Net.accept`](accept.md) calls are Err. Connections it
accepted stay open, and the address is free again. Closing it again does nothing.

<!-- end generated -->

## Details

- An accept that is waiting on another goroutine when the listener closes returns `Err` right away,
  so closing the listener is one way for one goroutine to stop another's accept loop; to stop a
  single accept without closing the listener, pass `Net.accept` a context and cancel it.
- `Net.close` returns `void`, not a `Result`, so a deferred close discards no error, unlike
  `File.close`.
- Closing a listener twice does nothing, so a deferred close can follow an explicit one.
- Connections the listener accepted stay open until they are closed. The address is free again for a
  later [Net.listen](listen.md), and [Net.address](address.md) still reports it.
- `Http.serve` closes the listener when it returns, so a listener passed to it needs no close of its
  own. Closing the listener while `Http.serve` runs makes `Http.serve` return `Err`.

## Basic examples

After `Net.close`, an accept is `Err`:

```ant
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
Net.close(listener)
println(Net.accept(listener))
```

```text
Err{ error: Error{ message: "the listener is closed", kind: null } }
```

Close a listener with `defer`, right after opening it, so that every way out of the block closes it:

```ant
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
defer { Net.close(listener) }
go {
  let conn = Result.unwrap(listener.accept())
  defer { conn.close() }
  Result.unwrap(conn.writeLine("welcome"))
}
let conn = Result.unwrap(Net.dial(listener.address()))
defer { conn.close() }
println(Result.unwrap(conn.readLine()))
```

```text
welcome
```

## Scope

### A pending accept

An accept that waits on another goroutine returns `Err` as soon as the listener closes:

```ant
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let accepted: Channel[Result[File.Handle, Error]] = Channel.new()
go { accepted.send(listener.accept()) }
Time.sleep(10ms)
listener.close()
println(Channel.receive(accepted))
```

```text
Err{ error: Error{ message: "the listener is closed", kind: null } }
```

### A closed listener

A second close does nothing, and every later accept is `Err`:

```ant
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
listener.close()
listener.close()
println(listener.accept())
println(listener.accept())
```

```text
Err{ error: Error{ message: "the listener is closed", kind: null } }
Err{ error: Error{ message: "the listener is closed", kind: null } }
```

## Applications

### Stop a server

A server loop accepts until the listener closes. Here `serve` greets each client, and the script
closes the listener after two clients; the pending accept is then `Err`, and `try` ends the loop:

```ant
func serve(listener: Net.Listener): Result[void, Error] {
  while true {
    let conn = try listener.accept()
    defer { conn.close() }
    let name = try conn.readLine()
    try conn.writeLine("hello, ${name ?? "stranger"}")
  }
}

let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let stopped: Channel[Result[void, Error]] = Channel.new()
go { stopped.send(serve(listener)) }
for name in ["Ann", "Bob"] {
  let conn = Result.unwrap(Net.dial(listener.address()))
  defer { conn.close() }
  Result.unwrap(conn.writeLine(name))
  println(Result.unwrap(conn.readLine()))
}
listener.close()
println(Channel.receive(stopped))
```

```text
hello, Ann
hello, Bob
Err{ error: Error{ message: "the listener is closed", kind: null } }
```

### Stop an HTTP server

Closing the listener under [`Http.serve`](../Http/serve.md) stops it, and `Http.serve` returns
`Err`:

```ant
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let served: Channel[Result[void, Error]] = Channel.new()
go { served.send(Http.serve(listener, { Http.Reply{ status: 204, headers: Map{}, body: "" } })) }
Time.sleep(50ms)
listener.close()
println(Channel.receive(served))
```

```text
Err{ error: Error{ message: "the listener is closed", kind: null } }
```

## Properties and relations

After `Net.close`, an accept is `Err` even with a context that is not done:

```ant
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
listener.close()
let ctx = Context.new()
assert(listener.accept(ctx) is Err)
assert(!ctx.isDone())
```

## Possible issues

### Unwrapping in an accept loop

An accept loop that unwraps every accept panics when the listener closes, since closing is what
makes the accept `Err`. Return the error with `try`, as in [Stop a server](#stop-a-server), or check
for `Err` and leave the loop:

```ant
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
go {
  Time.sleep(10ms)
  listener.close()
}
while true {
  let conn = Result.unwrap(listener.accept())
  // panic: Result.unwrap on an Err: Error{ message: "the listener is closed", kind: null }
  _ = conn.close()
}
```

### There is nothing to try

`Net.close` returns `void`, so `try` on it is a compile error:

```ant
func stop(listener: Net.Listener): Result[void, Error] {
  try Net.close(listener) // error: try needs a Result, got void
  Result.ok(void)
}
```

## See also

- [Net.listen](listen.md): open a listener.
- [Net.accept](accept.md): wait for the next connection.
- [Net.address](address.md): the address a listener listens on.
- [File.close](../File/close.md): close a connection.
- [Context.cancel](../Context/cancel.md): give up on one accept instead of closing the listener.
