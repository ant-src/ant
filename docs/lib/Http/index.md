<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http

Http holds an HTTP client and server. A request reads the whole response, and a server calls a
handler for each request on a goroutine of its own.

<!-- end generated -->

`Http` is a client and a server in four functions. [Http.request](request.md) sends one request and
reads the whole response, and [Http.get](get.md) and [Http.post](post.md) are `Http.request` with
the method filled in. [Http.serve](serve.md) answers requests on a [Net.Listener](../Net/index.md)
by calling a handler, a function from an [Http.Request](types/Request.md) to an
[Http.Reply](types/Reply.md). Everything else is plain data: the client gets an
[Http.Response](types/Response.md) record back, and a handler returns a record.

## Details

- A response of any status is `Ok`: a 404 or a 500 comes back with its `status`, and the program
  checks it. `Err` means no response arrived: a malformed URL, no connection, the timeout, or the
  context.
- No redirect is followed. A 3xx comes back as it is, with its `Location` header.
- A body is read whole, by the client and by the server, and is held once, as received, in `bytes`.
  [Http.text](text.md) decodes it.
- Headers hold one string per name, and a repeated header is joined with `, `. Two names that differ
  only in case name one header: the client is `Err` on them, and a reply with them panics in the
  handler's goroutine.
- A client call without `timeout` or `ctx` has no deadline and waits as long as the server takes.
  [Http.serve](serve.md) runs until its `ctx` is done, so run it on a goroutine to talk to it from
  the same program.

## Basic examples

Start a server on a free local port with `Http.serve`, and ask it for a page with `Http.get`:

```ant
func hello(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: "hello from ${req.path}" }
}

let listener = try Net.listen("127.0.0.1:0")
let stop = Context.new()
defer { stop.cancel() }
go { Http.serve(listener, hello, stop) }

let res = try Http.get("http://${listener.address()}/docs")
println(res.status)
println(res.text())
```

```text
200
hello from /docs
```

The other examples on this page ask a server like that one, at `base`. It answers `/hello` with
text, `/user` with JSON, `/old` with a redirect to `/hello`, `/slow` after a second, `/lang` with
the `Accept-Language` header it got, `/echo` with the method and body it got, and any other path
with a 404.

Send a body with `Http.post`, or any method with `Http.request`:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: Map{}, body: "hello" } }
  if req.path == "/user" {
    let headers = Map{ "Content-Type": "application/json" }
    return Http.Reply{ status: 200, headers, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/slow" { Time.sleep(1s) }
  if req.path == "/lang" {
    return Http.Reply{ status: 200, headers: Map{}, body: req.headers["Accept-Language"] ?? "any" }
  }
  if req.path == "/echo" {
    return Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.text()}" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
println((try Http.post("${base}/echo", "buy milk")).text())
println((try Http.request("PUT", "${base}/echo", body: "call Ann")).text())
```

```text
POST buy milk
PUT call Ann
```

## Scope

A handler gets `path` decoded and `query` raw, as sent. Read the query with
[Url.parseQuery](../Url/parseQuery.md):

```ant
func greet(req: Http.Request): Http.Reply {
  let query = Url.parseQuery(req.query)
  if query is Err { return Http.Reply{ status: 400, headers: Map{}, body: query.error.message } }
  let name = query.value["name"]?.first() ?? "stranger"
  let body = "path ${req.path}\nquery ${req.query}\nname ${name}"
  Http.Reply{ status: 200, headers: Map{}, body }
}

let listener = try Net.listen("127.0.0.1:0")
go { Http.serve(listener, greet) }

println((try Http.get("http://${listener.address()}/my%20docs?name=Ann%20Lee")).text())
```

```text
path /my docs
query name=Ann%20Lee
name Ann Lee
```

A redirect comes back as it is. Ask for its `Location` to follow it:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: Map{}, body: "hello" } }
  if req.path == "/user" {
    let headers = Map{ "Content-Type": "application/json" }
    return Http.Reply{ status: 200, headers, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/slow" { Time.sleep(1s) }
  if req.path == "/lang" {
    return Http.Reply{ status: 200, headers: Map{}, body: req.headers["Accept-Language"] ?? "any" }
  }
  if req.path == "/echo" {
    return Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.text()}" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let moved = try Http.get("${base}/old")
println("${moved.status} ${moved.headers["Location"]}")
let page = try Http.get("${base}${moved.headers["Location"]}")
println(page.text())
```

```text
301 /hello
hello
```

No server at the address is `Err`, since no response arrived:

```ant
let listener = try Net.listen("127.0.0.1:0")
let url = "http://${listener.address()}/"
listener.close()
println(Http.get(url) is Err)
```

```text
true
```

## Named arguments

Every client function takes `headers`, which go out with the request:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: Map{}, body: "hello" } }
  if req.path == "/user" {
    let headers = Map{ "Content-Type": "application/json" }
    return Http.Reply{ status: 200, headers, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/slow" { Time.sleep(1s) }
  if req.path == "/lang" {
    return Http.Reply{ status: 200, headers: Map{}, body: req.headers["Accept-Language"] ?? "any" }
  }
  if req.path == "/echo" {
    return Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.text()}" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
println((try Http.get("${base}/lang")).text())
println((try Http.get("${base}/lang", headers: Map{ "Accept-Language": "fr" })).text())
```

```text
any
fr
```

`timeout` bounds one call, from connecting to the end of the body:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: Map{}, body: "hello" } }
  if req.path == "/user" {
    let headers = Map{ "Content-Type": "application/json" }
    return Http.Reply{ status: 200, headers, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/slow" { Time.sleep(1s) }
  if req.path == "/lang" {
    return Http.Reply{ status: 200, headers: Map{}, body: req.headers["Accept-Language"] ?? "any" }
  }
  if req.path == "/echo" {
    return Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.text()}" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
println(Http.get("${base}/slow", timeout: 50ms))
```

```text
Err{ error: Error{ message: "timeout" } }
```

`ctx` bounds a call too, and one context can bound several calls, or stop a server:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: Map{}, body: "hello" } }
  if req.path == "/user" {
    let headers = Map{ "Content-Type": "application/json" }
    return Http.Reply{ status: 200, headers, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/slow" { Time.sleep(1s) }
  if req.path == "/lang" {
    return Http.Reply{ status: 200, headers: Map{}, body: req.headers["Accept-Language"] ?? "any" }
  }
  if req.path == "/echo" {
    return Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.text()}" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let ctx = Context.new(timeout: 50ms)
println(Http.get("${base}/slow", ctx: ctx))
println(Http.get("${base}/hello", ctx: ctx))
```

```text
Err{ error: Error{ message: "timeout" } }
Err{ error: Error{ message: "timeout" } }
```

## Functions by job

### Client

- [Http.get](get.md): a GET request, and the whole response
- [Http.post](post.md): a POST request with a body
- [Http.request](request.md): a request with any method, body, and headers

### Server

- [Http.serve](serve.md): answer every request on a listener with a handler, each on a goroutine of
  its own

### Types

- [Http.Response](types/Response.md): what the client gets back: status, headers, and the body
- [Http.Request](types/Request.md): what a handler gets: method, path, query, headers, body, and the
  client's address
- [Http.Reply](types/Reply.md): what a handler returns: status, headers, and a body of text or
  `Bytes`

## Applications

Fetch JSON, turn a bad status into an `Err`, and add context to every failure:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: Map{}, body: "hello" } }
  if req.path == "/user" {
    let headers = Map{ "Content-Type": "application/json" }
    return Http.Reply{ status: 200, headers, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/slow" { Time.sleep(1s) }
  if req.path == "/lang" {
    return Http.Reply{ status: 200, headers: Map{}, body: req.headers["Accept-Language"] ?? "any" }
  }
  if req.path == "/echo" {
    return Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.text()}" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
func fetchName(base: string, path: string): Result[string, Error] {
  let res = try (Http.get("${base}${path}") | Result.wrapErr("get ${path}"))
  if res.status != 200 {
    return Result.fail("get ${path}: status ${res.status}")
  }
  let user = try (Json.parse(res.text()) | Result.wrapErr("parse ${path}"))
  let name = if user is Map { user["name"] } else { null }
  Result.ok(if name is string { name } else { "unknown" })
}

println(fetchName(base, "/user"))
println(fetchName(base, "/nobody"))
```

```text
Ok{ value: "Ann" }
Err{ error: Error{ message: "get /nobody: status 404" } }
```

Check several pages, with a short timeout for each:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: Map{}, body: "hello" } }
  if req.path == "/user" {
    let headers = Map{ "Content-Type": "application/json" }
    return Http.Reply{ status: 200, headers, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/slow" { Time.sleep(1s) }
  if req.path == "/lang" {
    return Http.Reply{ status: 200, headers: Map{}, body: req.headers["Accept-Language"] ?? "any" }
  }
  if req.path == "/echo" {
    return Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.text()}" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
for path in ["/hello", "/slow", "/gone"] {
  let res = Http.get("${base}${path}", timeout: 100ms)
  let state = if res is Err { res.error.message } else { "${res.value.status}" }
  println("${path}: ${state}")
}
```

```text
/hello: 200
/slow: timeout
/gone: 404
```

A server and a client in one program, stopped by a context once the work is done:

```ant
func shout(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: req.text().toUpper() }
}

let listener = try Net.listen("127.0.0.1:0")
let url = "http://${listener.address()}/"
let stop = Context.new()
let served: Channel[Result[void, Error]] = Channel.new(1)
go { served.send(Http.serve(listener, shout, stop)) }

for word in ["ant", "http"] {
  println((try Http.post(url, word)).text())
}
stop.cancel()
println(served.receive())
```

```text
ANT
HTTP
Ok{ value: void }
```

## Properties and relations

`Http.get` and `Http.post` are `Http.request` with the method filled in:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: Map{}, body: "hello" } }
  if req.path == "/user" {
    let headers = Map{ "Content-Type": "application/json" }
    return Http.Reply{ status: 200, headers, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/slow" { Time.sleep(1s) }
  if req.path == "/lang" {
    return Http.Reply{ status: 200, headers: Map{}, body: req.headers["Accept-Language"] ?? "any" }
  }
  if req.path == "/echo" {
    return Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.text()}" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
assert(try Http.get("${base}/echo") == try Http.request("GET", "${base}/echo"))
let posted = try Http.post("${base}/echo", "hi")
assert(posted == try Http.request("POST", "${base}/echo", body: "hi"))
```

The body arrives as received in `bytes`, and `text()` decodes it:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: Map{}, body: "hello" } }
  if req.path == "/user" {
    let headers = Map{ "Content-Type": "application/json" }
    return Http.Reply{ status: 200, headers, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/slow" { Time.sleep(1s) }
  if req.path == "/lang" {
    return Http.Reply{ status: 200, headers: Map{}, body: req.headers["Accept-Language"] ?? "any" }
  }
  if req.path == "/echo" {
    return Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.text()}" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let res = try Http.post("${base}/echo", "naïve")
assert(res.text() == "POST naïve")
assert(res.bytes == res.text().toBytes())
```

## Possible issues

`try` stops only on `Err`, and a 404 or a 500 is `Ok`, so a program that reads the body without
checking `status` reads the error page:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: Map{}, body: "hello" } }
  if req.path == "/user" {
    let headers = Map{ "Content-Type": "application/json" }
    return Http.Reply{ status: 200, headers, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/slow" { Time.sleep(1s) }
  if req.path == "/lang" {
    return Http.Reply{ status: 200, headers: Map{}, body: req.headers["Accept-Language"] ?? "any" }
  }
  if req.path == "/echo" {
    return Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.text()}" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let res = try Http.get("${base}/user/42")
println(res.text())
```

```text
no /user/42
```

A reply status outside 200 to 599, such as an interim 100, panics in the handler's goroutine. Like
any panic in a handler, it is reported as a goroutine's panic is, and the client gets status 500
with an empty body:

```ant
func early(req: Http.Request): Http.Reply {
  Http.Reply{ status: 100, headers: Map{}, body: "wait" }
}

let listener = try Net.listen("127.0.0.1:0")
go { Http.serve(listener, early) }

let res = try Http.get("http://${listener.address()}/")
println(res.status)
println(res.text().isEmpty())
```

```text
500
true
```

A timeout must be positive and of a fixed length. Weeks, months, and years are `Err` before anything
is sent:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: Map{}, body: "hello" } }
  if req.path == "/user" {
    let headers = Map{ "Content-Type": "application/json" }
    return Http.Reply{ status: 200, headers, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/slow" { Time.sleep(1s) }
  if req.path == "/lang" {
    return Http.Reply{ status: 200, headers: Map{}, body: req.headers["Accept-Language"] ?? "any" }
  }
  if req.path == "/echo" {
    return Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.text()}" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
println(Http.get("http://127.0.0.1:1/", timeout: 0s))
println(Http.get("http://127.0.0.1:1/", timeout: Time.duration(weeks: 1)))
```

```text
Err{ error: Error{ message: "get \"http://127.0.0.1:1/\": the timeout is zero or negative" } }
Err{ error: Error{ message: "get \"http://127.0.0.1:1/\": the timeout has years, months or weeks, which have no fixed length" } }
```

A script ends when its last line runs, and a server on a goroutine ends with it. A script that only
serves calls `Http.serve` on its main path, with [Os.interrupt](../Os/interrupt.md) as `ctx` to stop
on Ctrl-C.

## See also

[Http.get](get.md) · [Http.post](post.md) · [Http.request](request.md) · [Http.serve](serve.md) ·
[Http.Response](types/Response.md) · [Http.Request](types/Request.md) · [Http.Reply](types/Reply.md)
· [Net.listen](../Net/listen.md) · [Net.address](../Net/address.md) · [Context](../Context/index.md)
· [Url](../Url/index.md) · [Json.parse](../Json/parse.md)
