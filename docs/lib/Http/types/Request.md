<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http.Request

```ant signature
type Http.Request = { method: string, path: string, query: string, headers: Map[string, string], bytes: Bytes, remote: string }
```

Http.Request is a request that [`Http.serve`](../serve.md) passes to its handler. path is decoded, query is raw for [`Url.parseQuery`](../../Url/parseQuery.md), and remote is the client's "ip:port". The body is held once, in bytes, as in [`Http.Response`](Response.md), and the headers hold one string per name, Host first and the rest in order of name, each name in canonical form.

<!-- end generated -->

## Details

- `path` is decoded, so `%20` arrives as a space and `%2F` as `/`. `query` is raw: pass it to
  [Url.parseQuery](../../Url/parseQuery.md) to decode it.
- A header the client sent twice arrives as one string, its values joined with `, `.
- `headers` has `Host` first and the rest in order of name, each name in canonical form: a header
  sent as `x-request-id` is read as `headers["X-Request-Id"]`.
- The body is read whole and held once, as received, in `bytes`. [Http.text](../text.md) decodes
  it, with U+FFFD for bytes that are not UTF-8.

## Basic examples

A handler of [Http.serve](../serve.md) takes an `Http.Request`. This one answers with its method and
path:

```ant
func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}" }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let base = "http://${l.address()}"
println((try Http.get("${base}/notes")).text())
println((try Http.post("${base}/notes", "buy milk")).text())
```

```text
GET /notes
POST /notes
```

Read the query and the body:

```ant
func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: "query=${req.query} body=${req.text()}" }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let res = try Http.post("http://${l.address()}/search?q=ant&page=2", "hello")
println(res.text())
```

```text
query=q=ant&page=2 body=hello
```

## Scope

`path` is decoded. A request for the root has the path `/` and an empty query:

```ant
func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: "[${req.path}] [${req.query}]" }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let base = "http://${l.address()}"
println((try Http.get("${base}/files/my%20notes.txt")).text())
println((try Http.get(base)).text())
```

```text
[/files/my notes.txt] []
[/] []
```

`query` stays raw. [Url.parseQuery](../../Url/parseQuery.md) decodes it and collects a repeated key:

```ant
func handle(req: Http.Request): Http.Reply {
  let params = Url.parseQuery(req.query)
  if params is Err {
    return Http.Reply{ status: 400, headers: Map{}, body: "bad query" }
  }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.query}\n${params.value}" }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

println((try Http.get("http://${l.address()}/?name=Ann%20Lee&tag=a&tag=b")).text())
```

```text
name=Ann%20Lee&tag=a&tag=b
Map{ "name": ["Ann Lee"], "tag": ["a", "b"] }
```

`method` is the method as the client sent it, any token:

```ant
func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: req.method }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let url = "http://${l.address()}/cache"
for method in ["GET", "DELETE", "PURGE"] {
  println((try Http.request(method, url)).text())
}
```

```text
GET
DELETE
PURGE
```

`headers` holds the request's headers, `Host` among them. Look one up by name; a missing one is
`null`:

```ant
func handle(req: Http.Request): Http.Reply {
  let id = req.headers["X-Request-Id"] ?? "none"
  Http.Reply{ status: 200, headers: Map{}, body: "host=${req.headers.has("Host")} id=${id}" }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let url = "http://${l.address()}/"
println((try Http.get(url, headers: Map{ "X-Request-Id": "42" })).text())
println((try Http.get(url)).text())
```

```text
host=true id=42
host=true id=none
```

A header sent twice arrives as one string, its values joined with `, `. Here the request is written
by hand over [Net.dial](../../Net/dial.md), since a `Map` cannot hold one name twice:

```ant
func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: req.headers["Accept"] ?? "none" }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let conn = try Net.dial(l.address())
defer { conn.close() }
try conn.writeText(
  "GET / HTTP/1.1\r\nHost: x\r\nAccept: text/html\r\nAccept: application/json\r\nConnection: close\r\n\r\n",
)
let response = try conn.readAll()
println(response.split("\r\n\r\n")[1])
```

```text
text/html, application/json
```

The body is held once, as received, in `bytes`, and `text()` decodes it. A byte that is not UTF-8
becomes U+FFFD in the text:

```ant
func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: "${req.bytes} ${req.text().length()}" }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let url = "http://${l.address()}/upload"
println((try Http.post(url, "café")).text())
println((try Http.post(url, Bytes.fromArray([0xff, 0x41]))).text())
println((try Http.get(url)).text())
```

```text
Bytes(63 61 66 c3 a9) 4
Bytes(ff 41) 2
Bytes() 0
```

`remote` is the client's `"ip:port"`. The port changes from run to run, so this client on the same
machine checks only the address:

```ant
func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: "${req.remote.startsWith("127.0.0.1:")}" }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

println((try Http.get("http://${l.address()}/")).text())
```

```text
true
```

`Http.Request` is a record type, so a literal builds one, for example to test a handler without a
server:

```ant
func handle(req: Http.Request): Http.Reply {
  if req.path != "/health" {
    return Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
  }
  Http.Reply{ status: 200, headers: Map{}, body: "ok" }
}

let req = Http.Request{
  method: "GET",
  path: "/health",
  query: "",
  headers: Map{ "Host": "example.com" },
  bytes: "".toBytes(),
  remote: "127.0.0.1:50000",
}
println(handle(req))
println(handle(Http.Request{ ...req, path: "/admin" }).status)
```

```text
Http.Reply{ status: 200, headers: Map{}, body: "ok" }
404
```

## Applications

Route by method and path, and answer anything else with 404 or 405:

```ant
var notes: Array[string] = []

func handle(req: Http.Request): Http.Reply {
  if req.path != "/notes" {
    return Http.Reply{ status: 404, headers: Map{}, body: "not found" }
  }
  if req.method == "GET" {
    Http.Reply{ status: 200, headers: Map{}, body: notes | String.join(", ") }
  } else if req.method == "POST" {
    notes.push(req.text())
    Http.Reply{ status: 201, headers: Map{}, body: "added" }
  } else {
    Http.Reply{ status: 405, headers: Map{ "Allow": "GET, POST" }, body: "" }
  }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let url = "http://${l.address()}/notes"
let _ = try Http.post(url, "milk")
let _ = try Http.post(url, "bread")
println((try Http.get(url)).text())
println((try Http.request("DELETE", url)).status)
println((try Http.get("http://${l.address()}/other")).status)
```

```text
milk, bread
405
404
```

Read a query parameter with a default. A malformed query is an error, answered with 400; only a
missing parameter takes the default:

```ant
func handle(req: Http.Request): Http.Reply {
  let params = Url.parseQuery(req.query)
  if params is Err {
    return Http.Reply{ status: 400, headers: Map{}, body: "bad query" }
  }
  let name = params.value.get("name")?.first() ?? "stranger"
  Http.Reply{ status: 200, headers: Map{}, body: "hello, ${name}" }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let base = "http://${l.address()}/greet"
println((try Http.get("${base}?name=Ann")).text())
println((try Http.get(base)).text())
println((try Http.get("${base}?name=%zz")).status)
```

```text
hello, Ann
hello, stranger
400
```

Turn away a request without the right token before doing any work:

```ant
func handle(req: Http.Request): Http.Reply {
  if req.headers["Authorization"] != "Bearer s3cret" {
    return Http.Reply{ status: 401, headers: Map{}, body: "unauthorized" }
  }
  Http.Reply{ status: 200, headers: Map{}, body: "welcome" }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let url = "http://${l.address()}/admin"
println((try Http.get(url)).status)
println((try Http.get(url, headers: Map{ "Authorization": "Bearer s3cret" })).text())
```

```text
401
welcome
```

Parse a JSON body, and answer 400 when it is not JSON:

```ant
func handle(req: Http.Request): Http.Reply {
  let value = Json.parse(req.text())
  if value is Err {
    return Http.Reply{ status: 400, headers: Map{}, body: "invalid JSON" }
  }
  let doc = value.value
  let name = if doc is Map { doc["name"] } else { null }
  if name is string {
    Http.Reply{ status: 201, headers: Map{}, body: "created ${name}" }
  } else {
    Http.Reply{ status: 422, headers: Map{}, body: "name must be a string" }
  }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let url = "http://${l.address()}/users"
for body in ["{\"name\": \"Ann\"}", "{\"name\": 7}", "{oops"] {
  let res = try Http.post(url, body, headers: Map{ "Content-Type": "application/json" })
  println("${res.status} ${res.text()}")
}
```

```text
201 created Ann
422 name must be a string
400 invalid JSON
```

Echo an upload back unchanged. `bytes` is the body as received, and a reply takes `Bytes`:

```ant
func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: req.bytes }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let data = Bytes.fromArray([0x89, 0x50, 0x4e, 0x47, 0xff])
let res = try Http.post("http://${l.address()}/echo", data)
println(res.bytes == data)
```

```text
true
```

## Properties and relations

`body` is `bytes` decoded as text, as [String.fromBytesLossy](../../String/fromBytesLossy.md)
decodes it:

```ant
func handle(req: Http.Request): Http.Reply {
  let same = req.text() == String.fromBytesLossy(req.bytes)
  Http.Reply{ status: if same { 200 } else { 500 }, headers: Map{}, body: "" }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let url = "http://${l.address()}/"
for data in ["", "plain", "日本語"] {
  assert((try Http.post(url, data)).status == 200)
}
assert((try Http.post(url, Bytes.fromArray([0xc3, 0x28, 0xff]))).status == 200)
```

## Possible issues

`path` is decoded, so an encoded slash looks like a real one. `/a%2Fb` and `/a/b` reach the handler
as the same path:

```ant
func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: req.path }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let base = "http://${l.address()}"
println((try Http.get("${base}/a%2Fb")).text())
println((try Http.get("${base}/a/b")).text())
```

```text
/a/b
/a/b
```

`query` is not decoded. Comparing it, or splitting it by hand, sees `%20` as written:

```ant
func handle(req: Http.Request): Http.Reply {
  let raw = req.query.split("=")[1]
  let decoded = Url.parseQuery(req.query).unwrapOr(Map{}).get("q")?.first() ?? ""
  Http.Reply{ status: 200, headers: Map{}, body: "${raw} | ${decoded}" }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

println((try Http.get("http://${l.address()}/?q=red%20ant")).text())
```

```text
red%20ant | red ant
```

`body` replaces bytes that are not UTF-8, so a handler that passes `body` on changes the data. Pass
`bytes` on instead:

```ant
func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: req.text() }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let res = try Http.post("http://${l.address()}/", Bytes.fromArray([0xff, 0x41]))
println(res.bytes)
```

```text
Bytes(ef bf bd 41)
```

A handler that panics, here on a body that is not JSON, gives the client status 500 and an empty
body. The panic is reported on standard error and the server goes on:

```ant
func handle(req: Http.Request): Http.Reply {
  let doc = Json.parse(req.text()).unwrap()
  Http.Reply{ status: 200, headers: Map{}, body: Json.stringify(doc) }
}

let l = try Net.listen("127.0.0.1:0")
let stop = Context.new()
go { Http.serve(l, handle, ctx: stop) }
defer { stop.cancel() }

let url = "http://${l.address()}/"
let bad = try Http.post(url, "{oops")
println("${bad.status} [${bad.text()}]")
println((try Http.post(url, "[\"ok\"]")).text())
```

```text
500 []
["ok"]
```

## See also

[Http.serve](../serve.md) · [Http.Reply](Reply.md) · [Http.Response](Response.md) ·
[Http.request](../request.md) · [Url.parseQuery](../../Url/parseQuery.md) ·
[Net.listen](../../Net/listen.md) · [String.fromBytesLossy](../../String/fromBytesLossy.md)
