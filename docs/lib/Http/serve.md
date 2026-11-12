<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http.serve

```ant signature
func Http.serve(l: Net.Listener, handler: func(Http.Request): Http.Reply, ctx: Context | null = null): Result[void, Error]
```

Http.serve answers every request that arrives on l by calling handler on a goroutine of its own. It
returns Ok once ctx is done and the requests in flight have finished, waiting at most five seconds
for them, or serves until the run is cancelled when ctx is null; it closes l on return and is Err
when l is closed under it. A panic in handler sends status 500, and a reply status outside 200 to
599, or reply header names that differ only in case, panic in the handler's goroutine.

<!-- end generated -->

## Details

- `Http.serve` blocks until `ctx` is done, or with no `ctx` until the run is cancelled. Run it on a
  goroutine to talk to the server from the same program. The program exits when the script ends,
  server or not.
- Each request runs `handler` on a goroutine of its own, so handlers run at the same time. Guard
  state they share with a [Mutex](../Mutex/index.md).
- A panic in `handler` ends that handler's goroutine only: the server keeps serving.
- `Http.serve` closes the listener when it returns, so a listener serves once.
- `query` is raw; parse it with [Url.parseQuery](../Url/parseQuery.md).

## Basic examples

`Http.serve(l, handler, ctx)` answers requests on a listener until `ctx` is done. Listen on port 0
for any free port, serve on a goroutine, and send a request with [Http.get](get.md):

```ant
func hello(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: "hello from ${req.path}" }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
let served: Channel[Result[void, Error]] = Channel.new(1)
go { served.send(Http.serve(listener, hello, stop)) }
let reply = try Http.get("http://${Net.address(listener)}/docs")
println("${reply.status} ${reply.body}")
stop.cancel()
println(served.receive())
```

```text
200 hello from /docs
Ok{ value: void }
```

A handler can be a lambda, with the request as `x`:

```ant
let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go {
  Http.serve(listener, { Http.Reply{ status: 200, headers: Map{}, body: x.method } }, stop)
}
println((try Http.get("http://${Net.address(listener)}/")).body)
stop.cancel()
```

```text
GET
```

A server that runs until Ctrl-C passes [Os.interrupt](../Os/interrupt.md) as `ctx`. The handler
routes on the request's path:

<!-- test: before
go {
  Time.sleep(Time.duration(milliseconds: 200))
  println(Http.get("http://127.0.0.1:48213/health").unwrap().body)
  println(Http.get("http://127.0.0.1:48213/missing").unwrap().body)
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
func handle(req: Http.Request): Http.Reply {
  if req.path == "/health" {
    Http.Reply{ status: 200, headers: Map{}, body: "ok" }
  } else {
    Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
  }
}

let listener = try Net.listen("127.0.0.1:48213")
println("listening on ${Net.address(listener)}")
let stop = Os.interrupt()
try Http.serve(listener, handle, stop)
println("stopped: ${stop.reason()}")
```

```text
listening on 127.0.0.1:48213
ok
no /missing
stopped: interrupt
```

## Scope

The request carries the method, the decoded path, the raw query, the headers, and the body:

```ant
func echo(req: Http.Request): Http.Reply {
  let lines = [
    "method: ${req.method}",
    "path: ${req.path}",
    "query: ${req.query}",
    "tag: ${req.headers["X-Tag"]}",
    "body: ${req.body}",
  ]
  Http.Reply{ status: 200, headers: Map{}, body: lines | String.join("\n") }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(listener, echo, stop) }
let url = "http://${Net.address(listener)}/caf%C3%A9/menu?q=a%20b&n=2"
let reply = try Http.post(url, "two teas", headers: Map{ "X-Tag": "order" })
println(reply.body)
stop.cancel()
```

```text
method: POST
path: /café/menu
query: q=a%20b&n=2
tag: order
body: two teas
```

`remote` is the client's address, `ip:port`, and the headers include `Host`:

```ant
func whoami(req: Http.Request): Http.Reply {
  let local = req.remote.startsWith("127.0.0.1:")
  Http.Reply{ status: 200, headers: Map{}, body: "${local} ${req.headers["Host"]}" }
}

let listener = try Net.listen("127.0.0.1:0")
let address = Net.address(listener)
let stop = Context.new()
go { Http.serve(listener, whoami, stop) }
let reply = try Http.get("http://${address}/")
println(reply.body == "true ${address}")
stop.cancel()
```

```text
true
```

A reply body may be `Bytes`, sent as they are. Headers go out as given:

```ant
func logo(req: Http.Request): Http.Reply {
  let png = Bytes.fromArray([137, 80, 78, 71])
  Http.Reply{ status: 200, headers: Map{ "Content-Type": "image/png" }, body: png }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(listener, logo, stop) }
let reply = try Http.get("http://${Net.address(listener)}/logo.png")
println(reply.headers["Content-Type"])
println(reply.bytes)
stop.cancel()
```

```text
image/png
Bytes(89 50 4e 47)
```

A redirect is a reply like any other, with a 3xx status and a `Location` header:

```ant
func moved(req: Http.Request): Http.Reply {
  Http.Reply{ status: 301, headers: Map{ "Location": "/new${req.path}" }, body: "" }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(listener, moved, stop) }
let reply = try Http.get("http://${Net.address(listener)}/page")
println("${reply.status} ${reply.headers["Location"]}")
stop.cancel()
```

```text
301 /new/page
```

Handlers run at the same time. Here a request to `/wait` blocks its handler until a request to `/go`
arrives, which only works because each has a goroutine of its own:

```ant
let gate: Channel[string] = Channel.new()

func handle(req: Http.Request): Http.Reply {
  if req.path == "/go" {
    gate.send("released")
    return Http.Reply{ status: 200, headers: Map{}, body: "sent" }
  }
  Http.Reply{ status: 200, headers: Map{}, body: gate.receive() ?? "closed" }
}

let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
let stop = Context.new()
go { Http.serve(listener, handle, stop) }
let waited: Channel[string] = Channel.new(1)
go { waited.send(Http.get("${base}/wait").unwrap().body) }
println((try Http.get("${base}/go")).body)
println(waited.receive())
stop.cancel()
```

```text
sent
released
```

## Named arguments

### ctx

`ctx` decides when `Http.serve` returns. A context with a timeout serves for that long:

```ant
let listener = try Net.listen("127.0.0.1:0")
let ctx = Context.new(timeout: Time.duration(milliseconds: 100))
let result = Http.serve(listener, { Http.Reply{ status: 204, headers: Map{}, body: "" } }, ctx: ctx)
println(result)
```

```text
Ok{ value: void }
```

A request in flight when `ctx` is done still gets its reply, and `Http.serve` waits for it before
returning. A new request after that finds no server:

```ant
func slow(req: Http.Request): Http.Reply {
  Time.sleep(Time.duration(milliseconds: 200))
  Http.Reply{ status: 200, headers: Map{}, body: "finished" }
}

let listener = try Net.listen("127.0.0.1:0")
let url = "http://${Net.address(listener)}/"
let stop = Context.new()
let served: Channel[Result[void, Error]] = Channel.new(1)
go { served.send(Http.serve(listener, slow, stop)) }
go {
  Time.sleep(Time.duration(milliseconds: 50))
  stop.cancel()
}
let reply = try Http.get(url)
println(reply.body)
println(served.receive())
println(Http.get(url).isOk())
```

```text
finished
Ok{ value: void }
false
```

## Applications

Read query parameters with [Url.parseQuery](../Url/parseQuery.md):

```ant
func greet(req: Http.Request): Http.Reply {
  let params = Url.parseQuery(req.query).unwrapOr(Map{})
  let name = params["name"]?.get(0) ?? "stranger"
  Http.Reply{ status: 200, headers: Map{}, body: "hello, ${name}" }
}

let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
let stop = Context.new()
go { Http.serve(listener, greet, stop) }
println((try Http.get("${base}/?name=Ann")).body)
println((try Http.get("${base}/")).body)
stop.cancel()
```

```text
hello, Ann
hello, stranger
```

A small key-value store. Handlers share the map, so a [Mutex](../Mutex/index.md) guards it:

```ant
let store: Map[string, string] = Map{}
let lock = Mutex.new()

func handle(req: Http.Request): Http.Reply {
  if req.method == "PUT" {
    lock.withLock({ store[req.path] = req.body })
    return Http.Reply{ status: 204, headers: Map{}, body: "" }
  }
  let value = lock.withLock({ store[req.path] })
  if value == null {
    return Http.Reply{ status: 404, headers: Map{}, body: "not found" }
  }
  Http.Reply{ status: 200, headers: Map{}, body: value }
}

let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
let stop = Context.new()
go { Http.serve(listener, handle, stop) }
try Http.request("PUT", "${base}/color", "green")
let found = try Http.get("${base}/color")
let missing = try Http.get("${base}/size")
println("${found.status} ${found.body}")
println("${missing.status} ${missing.body}")
stop.cancel()
```

```text
200 green
404 not found
```

A JSON endpoint parses the body with [Json.parse](../Json/parse.md) and answers with
[Json.stringify](../Json/stringify.md):

```ant
func sum(req: Http.Request): Http.Reply {
  let items = Json.parse(req.body).unwrapOr(null)
  if !(items is Array) {
    return Http.Reply{ status: 400, headers: Map{}, body: "want a JSON array" }
  }
  let total = items.map({ if x is float64 { x } else { 0.0 } }).sum() ?? 0.0
  let headers = Map{ "Content-Type": "application/json" }
  Http.Reply{ status: 200, headers, body: Json.stringify(Map{ "total": total }) }
}

let listener = try Net.listen("127.0.0.1:0")
let url = "http://${Net.address(listener)}/sum"
let stop = Context.new()
go { Http.serve(listener, sum, stop) }
println((try Http.post(url, "[1, 2, 3.5]")).body)
let bad = try Http.post(url, "nope")
println("${bad.status} ${bad.body}")
stop.cancel()
```

```text
{"total":6.5}
400 want a JSON array
```

A proxy passes the upstream body on as `bytes`, unchanged:

```ant
let back = try Net.listen("127.0.0.1:0")
let front = try Net.listen("127.0.0.1:0")
let upstream = "http://${Net.address(back)}"
let stop = Context.new()

func origin(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: "origin saw ${req.method} ${req.path}" }
}

func proxy(req: Http.Request): Http.Reply {
  let reply = Http.request(req.method, "${upstream}${req.path}", req.bytes)
  if reply is Err {
    return Http.Reply{ status: 502, headers: Map{}, body: reply.error.message }
  }
  Http.Reply{ status: reply.value.status, headers: Map{}, body: reply.value.bytes }
}

go { Http.serve(back, origin, stop) }
go { Http.serve(front, proxy, stop) }
println((try Http.get("http://${Net.address(front)}/a/b")).body)
stop.cancel()
```

```text
origin saw GET /a/b
```

## Properties and relations

A request body arrives twice, as text in `body` and as received in `bytes`. `Content-Length` is the
reply body's length in bytes:

```ant
let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go {
  Http.serve(listener, { Http.Reply{ status: 200, headers: Map{}, body: x.body } }, stop)
}
let reply = try Http.post("http://${Net.address(listener)}/", "naïve")
assert(reply.body == "naïve")
assert(reply.bytes == "naïve".toBytes())
assert(reply.headers["Content-Length"] == "${reply.bytes.length()}")
stop.cancel()
```

`Http.serve` closes its listener when it returns, so a listener cannot be served twice:

```ant
let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
stop.cancel()
assert(Http.serve(listener, { Http.Reply{ status: 204, headers: Map{}, body: "" } }, stop).isOk())
println(Http.serve(listener, { Http.Reply{ status: 204, headers: Map{}, body: "" } }))
```

```text
Err{ error: Error{ message: "the listener is closed" } }
```

## Possible issues

A panic in a handler does not stop the server. The panic is reported on standard error, the client
gets status 500 with an empty body, and the next request is served as usual:

```ant
func handle(req: Http.Request): Http.Reply {
  if req.path == "/crash" { panic("handler bug") }
  Http.Reply{ status: 200, headers: Map{}, body: "fine" }
}

let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
let stop = Context.new()
go { Http.serve(listener, handle, stop) }
let crashed = try Http.get("${base}/crash")
println("${crashed.status} '${crashed.body}'")
println((try Http.get("${base}/")).body)
stop.cancel()
```

```text
500 ''
fine
```

A reply status outside 200 to 599 panics in the handler's goroutine, so the client gets 500 as well.
Two header names that differ only in case name one header and panic the same way:

```ant
func handle(req: Http.Request): Http.Reply {
  if req.path == "/continue" {
    return Http.Reply{ status: 100, headers: Map{}, body: "" }
  }
  Http.Reply{ status: 200, headers: Map{ "X-Id": "1", "x-id": "2" }, body: "" }
}

let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
let stop = Context.new()
go { Http.serve(listener, handle, stop) }
println((try Http.get("${base}/continue")).status)
println((try Http.get("${base}/twice")).status)
stop.cancel()
```

```text
500
500
```

Closing the listener while `Http.serve` uses it is `Err`, not `Ok`. Cancel a context to stop a
server cleanly:

```ant
let listener = try Net.listen("127.0.0.1:0")
let served: Channel[Result[void, Error]] = Channel.new(1)
go { served.send(Http.serve(listener, { Http.Reply{ status: 204, headers: Map{}, body: "" } })) }
Time.sleep(Time.duration(milliseconds: 50))
listener.close()
println(served.receive())
```

```text
Err{ error: Error{ message: "the listener is closed" } }
```

## See also

[Http.get](get.md), [Http.post](post.md), [Http.request](request.md),
[Http.Request](types/Request.md), [Http.Reply](types/Reply.md), [Net.listen](../Net/listen.md),
[Net.address](../Net/address.md), [Os.interrupt](../Os/interrupt.md),
[Context.new](../Context/new.md), [Url.parseQuery](../Url/parseQuery.md)
