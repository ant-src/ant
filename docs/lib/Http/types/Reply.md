<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http.Reply

```ant signature
type Http.Reply = { status: int, headers: Map[string, string], body: string | Bytes }
```

Http.Reply is what an [`Http.serve`](../serve.md) handler returns to answer a request. body is sent
as given, a string as its UTF-8 and Bytes as they are, with the headers given plus Date and
Content-Length from the body, which replaces one the handler gives; a 204 or 304 reply has no body
and no Content-Length. A status outside 200 to 599, or two header names that differ only in case,
panics.

<!-- end generated -->

## Details

- The server sends the headers in `headers`, plus `Date` and a `Content-Length` from the body,
  which replaces one the handler gives. It fills in none, such as `Content-Type`, from the body.
- A 204 or 304 reply goes out with no body and no `Content-Length`.
- The panic ends only that request: the client gets status 500 with an empty body, and the server
  keeps serving.
- A reply with a 3xx `status` and a `Location` header is a redirect. Whether it is followed is up to
  the client; [Http.request](../request.md) does not follow it.

## Basic examples

A handler for [Http.serve](../serve.md) returns an `Http.Reply`. The client gets its status, its
headers, and its body:

```ant
func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain" }, body: "hello ${req.path}" }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
defer { Context.cancel(stop) }
go { Http.serve(listener, handle, stop) }

let res = try Http.get("http://${Net.address(listener)}/ant")
println(res.status)
println(res.headers["Content-Type"])
println(res.body)
```

```text
200
text/plain
hello /ant
```

`Http.Reply` is a record type, so a literal builds one, and it prints as a record:

```ant
let reply = Http.Reply{ status: 404, headers: Map{}, body: "not found" }
println(reply)
println(reply.status)
```

```text
Http.Reply{ status: 404, headers: Map{}, body: "not found" }
404
```

## Scope

`body` is a string or `Bytes`. `Bytes` are sent as they are, so a reply can carry binary data that
is not UTF-8:

```ant
func handle(req: Http.Request): Http.Reply {
  let png = Bytes.fromArray([0x89, 0x50, 0x4e, 0x47])
  Http.Reply{ status: 200, headers: Map{ "Content-Type": "image/png" }, body: png }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
defer { Context.cancel(stop) }
go { Http.serve(listener, handle, stop) }

let res = try Http.get("http://${Net.address(listener)}/logo.png")
println(res.bytes)
println(res.headers["Content-Length"])
```

```text
Bytes(89 50 4e 47)
4
```

`Content-Length` counts bytes, not characters: `"café ☕"` is six characters and nine bytes:

```ant
func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: "café ☕" }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
defer { Context.cancel(stop) }
go { Http.serve(listener, handle, stop) }

let res = try Http.get("http://${Net.address(listener)}/")
println(res.body)
println(res.headers["Content-Length"])
```

```text
café ☕
9
```

A status other than 200 goes out as given, with whatever body the handler chose. A redirect is a 3xx
status with a `Location` header:

```ant
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" {
    return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" }
  }
  if req.path != "/new" {
    return Http.Reply{ status: 404, headers: Map{}, body: "no page at ${req.path}" }
  }
  Http.Reply{ status: 200, headers: Map{}, body: "the new page" }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
defer { Context.cancel(stop) }
go { Http.serve(listener, handle, stop) }

let base = "http://${Net.address(listener)}"
for path in ["/new", "/old", "/gone"] {
  let res = try Http.get("${base}${path}")
  println("${res.status} ${res.headers["Location"] ?? res.body}")
}
```

```text
200 the new page
301 /new
404 no page at /gone
```

When the names match the fields, the record shorthand builds a reply from local values:

```ant
let status = 201
let headers = Map{ "Location": "/items/7" }
let body = "created"
println(Http.Reply{ status, headers, body })
```

```text
Http.Reply{ status: 201, headers: Map{ "Location": "/items/7" }, body: "created" }
```

## Applications

Small helpers keep the handlers short. Here `text` and `json` build the two kinds of reply a JSON
API sends:

```ant
type Item = { id: int, name: string }

func text(status: int, message: string): Http.Reply {
  Http.Reply{ status, headers: Map{ "Content-Type": "text/plain; charset=utf-8" }, body: message }
}

func json[T](value: T): Http.Reply {
  let headers = Map{ "Content-Type": "application/json" }
  Http.Reply{ status: 200, headers, body: Json.stringify(value) }
}

let items = [Item{ id: 1, name: "tea" }, Item{ id: 2, name: "milk" }]

func handle(req: Http.Request): Http.Reply {
  if req.method != "GET" { return text(405, "use GET") }
  if req.path != "/items" { return text(404, "no ${req.path}") }
  json(items)
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
defer { Context.cancel(stop) }
go { Http.serve(listener, handle, stop) }

let base = "http://${Net.address(listener)}"
let list = try Http.get("${base}/items")
println("${list.status} ${list.headers["Content-Type"]} ${list.body}")
let wrong = try Http.post("${base}/items", "{}")
println("${wrong.status} ${wrong.body}")
```

```text
200 application/json [{"id":1,"name":"tea"},{"id":2,"name":"milk"}]
405 use GET
```

Serve a file as it is on disk, reading it as `Bytes` so that any content passes unchanged. The
request path is decoded, so `/%2e%2e/secret.txt` arrives as `/../secret.txt`; cleaning it before the
join keeps the read inside `public`:

```ant
func handle(req: Http.Request): Http.Reply {
  let data = File.readBytes(Path.join(["public", Path.clean(req.path)]))
  if data is Err {
    return Http.Reply{ status: 404, headers: Map{}, body: "not found" }
  }
  Http.Reply{ status: 200, headers: Map{}, body: data.value }
}

try File.mkdir("public")
try File.write("public/hello.txt", "hi from disk\n")
try File.write("secret.txt", "not for clients\n")

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
defer { Context.cancel(stop) }
go { Http.serve(listener, handle, stop) }

let base = "http://${Net.address(listener)}"
for path in ["/hello.txt", "/missing.txt", "/%2e%2e/secret.txt"] {
  let res = try Http.get("${base}${path}")
  println("${res.status} ${res.body.trim()}")
}
```

```text
200 hi from disk
404 not found
404 not found
```

A proxy passes a `Response`'s `bytes` on unchanged, since `body` takes `Bytes`. Here the front
server forwards each request to a backend that answers with binary content, and answers 502 when the
backend cannot be reached:

```ant
func backend(req: Http.Request): Http.Reply {
  let data = Bytes.fromArray([0x00, 0xff, 0x10, 0x80])
  Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/x-demo" }, body: data }
}

let backendListener = try Net.listen("127.0.0.1:0")
let origin = "http://${Net.address(backendListener)}"
let stop = Context.new()
defer { Context.cancel(stop) }
go { Http.serve(backendListener, backend, stop) }

func front(req: Http.Request): Http.Reply {
  let res = Http.get("${origin}${req.path}")
  if res is Err {
    return Http.Reply{ status: 502, headers: Map{}, body: "backend unreachable" }
  }
  let contentType = res.value.headers["Content-Type"] ?? "application/octet-stream"
  Http.Reply{
    status: res.value.status,
    headers: Map{ "Content-Type": contentType },
    body: res.value.bytes,
  }
}

let frontListener = try Net.listen("127.0.0.1:0")
go { Http.serve(frontListener, front, stop) }

let direct = try Http.get("${origin}/blob")
let proxied = try Http.get("http://${Net.address(frontListener)}/blob")
println(proxied.bytes)
println(proxied.headers["Content-Type"])
println(proxied.bytes == direct.bytes)
```

```text
Bytes(00 ff 10 80)
application/x-demo
true
```

A handler is an ordinary function, so a test calls it with an `Http.Request` literal and checks the
reply, with no server and no network:

```ant
func greet(req: Http.Request): Http.Reply {
  let params = Url.parseQuery(req.query)
  if params is Err {
    return Http.Reply{ status: 400, headers: Map{}, body: "bad query" }
  }
  let name = params.value["name"]?.[0] ?? ""
  if name.isEmpty() {
    return Http.Reply{ status: 400, headers: Map{}, body: "name is missing" }
  }
  Http.Reply{ status: 200, headers: Map{}, body: "hello, ${name}" }
}

func get(path: string, query: string): Http.Request {
  Http.Request{
    method: "GET",
    path,
    query,
    headers: Map{},
    body: "",
    bytes: "".toBytes(),
    remote: "127.0.0.1:50000",
  }
}

let ok = greet(get("/greet", "name=Ann%20Lee"))
assert(ok.status == 200)
assert(ok.body == "hello, Ann Lee")
assert(greet(get("/greet", "")).status == 400)
```

## Properties and relations

The client reads exactly the bytes of `body`, and `Content-Length` is their number:

```ant
let bodies = Map{ "/empty": "", "/plain": "plain", "/unicode": "café ☕", "/line": "line\n" }

func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: bodies[req.path] ?? "" }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
defer { Context.cancel(stop) }
go { Http.serve(listener, handle, stop) }

let base = "http://${Net.address(listener)}"
for { key, value } in bodies {
  let res = try Http.get("${base}${key}")
  assert(res.bytes == value.toBytes())
  assert(res.headers["Content-Length"] == "${value.toBytes().length()}")
}
```

A string body and its UTF-8 as `Bytes` reach the client alike:

```ant
let text = "naïve"

func handle(req: Http.Request): Http.Reply {
  let body: string | Bytes = if req.path == "/bytes" { text.toBytes() } else { text }
  Http.Reply{ status: 200, headers: Map{}, body }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
defer { Context.cancel(stop) }
go { Http.serve(listener, handle, stop) }

let base = "http://${Net.address(listener)}"
let a = try Http.get("${base}/string")
let b = try Http.get("${base}/bytes")
assert(a.bytes == b.bytes)
assert(a.body == text && b.body == text)
```

## Possible issues

A status outside 200 to 599 panics in the handler's goroutine, so the client gets 500 instead. A 1xx
status is an interim response and cannot be the reply. The panic is reported on standard error, and
the server keeps serving:

```ant
func handle(req: Http.Request): Http.Reply {
  let status = if req.path == "/early" { 103 } else { 200 }
  Http.Reply{ status, headers: Map{}, body: "done" }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
defer { Context.cancel(stop) }
go { Http.serve(listener, handle, stop) }

let base = "http://${Net.address(listener)}"
println((try Http.get("${base}/early")).status)
println((try Http.get("${base}/late")).status)
```

```text
500
200
```

Header names are not case-sensitive, so two names that differ only in case name one header twice.
That panics too, and the client gets 500:

```ant
func handle(req: Http.Request): Http.Reply {
  let headers = Map{ "Cache-Control": "no-store", "cache-control": "max-age=60" }
  Http.Reply{ status: 200, headers, body: "fresh" }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
defer { Context.cancel(stop) }
go { Http.serve(listener, handle, stop) }

let res = try Http.get("http://${Net.address(listener)}/")
println("${res.status} ${res.body.isEmpty()}")
```

```text
500 true
```

The server does not guess a `Content-Type` from the body. Without one in `headers`, the client gets
none:

```ant
func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: "<h1>Hi</h1>" }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
defer { Context.cancel(stop) }
go { Http.serve(listener, handle, stop) }

let res = try Http.get("http://${Net.address(listener)}/")
println(res.headers["Content-Type"] ?? "no Content-Type")
```

```text
no Content-Type
```

`body` takes a string or `Bytes`, not `Html`. Interpolate the markup into a string:

```ant
func page(): Html {
  <h1>Hi</h1>
}

let wrong = Http.Reply{ status: 200, headers: Map{}, body: page() }
// error: cannot use Html where string | Bytes is expected
let right = Http.Reply{ status: 200, headers: Map{}, body: "<!doctype html>${page()}" }
```

## See also

[Http.serve](../serve.md) · [Http.Request](Request.md) · [Http.Response](Response.md) ·
[Http.request](../request.md) · [Net.listen](../../Net/listen.md) ·
[Json.stringify](../../Json/stringify.md) · [Bytes](../../Bytes/index.md)
