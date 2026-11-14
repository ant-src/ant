<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http.get

```ant signature
func Http.get(url: string, headers: Map[string, string] = Map{}, timeout: Time.Duration | null = null, ctx: Context | null = null): Result[Http.Response, Error]
```

Http.get sends a GET request to url and reads the whole response. It is [`Http.request`](request.md)
with the method "GET" and an empty body.

<!-- end generated -->

## Details

- Every response is `Ok`, whatever its status: a 404 or a 500 is `Ok` with that `status`. `Err`
  means no response arrived: a malformed URL, no connection, header names that differ only in case,
  the timeout, or the context.
- No redirect is followed. A 3xx comes back as it is, with its `Location` header.
- The whole body is read before `Http.get` returns, and it is held once, as received, in `bytes`.
  [Http.text](text.md) decodes it, and bytes that are not valid UTF-8 become U+FFFD.
- `headers` holds one string per name. A header the server repeats is joined with `, `.
- `timeout` bounds the whole call, from connecting to the end of the body. A timeout that is zero or
  negative, or that has years, months, or weeks, is `Err` before anything is sent. Without `timeout`
  or `ctx`, the call has no deadline and waits as long as the server takes.

## Basic examples

Start a server on a free local port, and get a page from it:

```ant
let listener = try Net.listen("127.0.0.1:0")
go { Http.serve(listener, { Http.Reply{ status: 200, headers: Map{}, body: "Hello, ${x.path}" } }) }

let response = try Http.get("http://${Net.address(listener)}/world")
println(response.status)
println(response.text())
```

```text
200
Hello, /world
```

The examples below run against a local server like this one, at the address `base`. Each says what
the paths it asks for answer. Here `/hello` answers `Hello, Ant`, and a path the server does not
know answers 404:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
println((try Http.get("${base}/hello")).text())
let missing = try Http.get("${base}/nothing")
println("${missing.status}: ${missing.text()}")
```

```text
Hello, Ant
404: not found: /nothing
```

When no server answers, the result is `Err`, and the message names the URL:

```ant
println(Http.get("http://127.0.0.1:1/"))
```

```text
Err{ error: Error{ message: "get \"http://127.0.0.1:1/\": connection refused" } }
```

## Scope

### A query string

The query goes in the URL as written. Escape each value with [Url.escape](../Url/escape.md). Here
`/search` answers with the raw query it received:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let term = "red & blue"
let response = try Http.get("${base}/search?q=${Url.escape(term)}&page=2")
println(response.text())
```

```text
query: q=red+%26+blue&page=2
```

### A redirect

`Http.get` returns the redirect itself. Here `/old` answers 301 with `Location: /hello`:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let response = try Http.get("${base}/old")
println(response.status)
println(response.headers["Location"])
```

```text
301
/hello
```

### An empty body

A response without a body has no bytes in `bytes`, and its `text()` is `""`. Here `/empty` answers
204:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let response = try Http.get("${base}/empty")
println(response.status)
println(response.text() == "")
println(response.bytes.length())
```

```text
204
true
0
```

### A binary body

Bytes that are not UTF-8 become U+FFFD in `text()`, while `bytes` keeps them as received. Here
`/logo` answers the four bytes a PNG file starts with:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let response = try Http.get("${base}/logo")
println(response.bytes)
println(response.text().count("\u{FFFD}"))
```

```text
Bytes(89 50 4e 47)
1
```

### A malformed URL

A URL without a scheme is `Err` before anything is sent:

```ant
println(Http.get("example.com/index.html"))
```

```text
Err{ error: Error{ message: "get \"example.com/index.html\": unsupported protocol scheme \"\"" } }
```

## Named arguments

### headers

`headers` are sent with the request. Here `/accept` answers with the `Accept` header it received:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
println((try Http.get("${base}/accept", headers: Map{ "Accept": "application/json" })).text())
println((try Http.get("${base}/accept")).text())
```

```text
application/json
none
```

### timeout

`timeout` bounds the whole call. When it passes, the result is `Err` with the message `timeout`.
Here `/slow` answers after one second:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
println(Http.get("${base}/slow", timeout: 100ms))
println(Http.get("${base}/slow", timeout: 5s).map({ x.text() }))
```

```text
Err{ error: Error{ message: "timeout" } }
Ok{ value: "done" }
```

### ctx

A context bounds the call as well, and shares the bound with other work. When the context is done,
the result is `Err` with the context's reason:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let ctx = Context.new()
go {
  Time.sleep(100ms)
  Context.cancel(ctx)
}
println(Http.get("${base}/slow", ctx: ctx))
```

```text
Err{ error: Error{ message: "cancelled" } }
```

When both are given, the earlier bound wins:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let ctx = Context.new(10s)
println(Http.get("${base}/slow", timeout: 100ms, ctx: ctx))
```

```text
Err{ error: Error{ message: "timeout" } }
```

## Applications

### Fetch JSON

Check the status, then parse the body with [Json.parse](../Json/parse.md). Here `/users/1` answers
with a user, and `/users/2` is not found:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
func fetchJson(url: string): Result[Json.Value, Error] {
  let response = try Http.get(url)
  if response.status != 200 {
    return Result.fail("get ${url}: status ${response.status}")
  }
  Json.parse(response.text()) | Result.wrapErr("parse ${url}")
}

let user = try fetchJson("${base}/users/1")
if user is Map { println(user["name"]) }
println(fetchJson("${base}/users/2").isOk())
```

```text
Ann
false
```

### Follow redirects

Ask for the path in each redirect's `Location`, up to a limit. Here `/old` redirects to `/hello`,
and `/loop` redirects to itself:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
func getFollowing(path: string, limit: int): Result[Http.Response, Error] {
  var next = path
  for _ in Iter.range(0, limit) {
    let response = try Http.get("${base}${next}")
    let location = response.headers["Location"]
    if response.status < 300 || response.status > 399 || location == null {
      return Result.ok(response)
    }
    next = location
  }
  Result.fail("get ${path}: more than ${limit} redirects")
}

println((try getFollowing("/old", 5)).text())
println(getFollowing("/loop", 5))
```

```text
Hello, Ant
Err{ error: Error{ message: "get /loop: more than 5 redirects" } }
```

### Fetch several pages at once

[Parallel.map](../Parallel/map.md) runs the requests on goroutines and keeps the order of the URLs:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let paths = ["/hello", "/users/1", "/nothing"]
let results = Parallel.map(
  paths,
  func(path: string, work: Context) { Http.get("${base}${path}", ctx: work) },
  limit: 3,
)
for result in results ?? [] {
  println(result.map({ x.status }))
}
```

```text
Ok{ value: 200 }
Ok{ value: 200 }
Ok{ value: 404 }
```

## Properties and relations

`Http.get(url)` is [Http.request](request.md) with the method `"GET"`:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let viaGet = try Http.get("${base}/hello")
let viaRequest = try Http.request("GET", "${base}/hello")
assert(viaGet.status == viaRequest.status && viaGet.text() == viaRequest.text())
```

`text()` is `bytes` decoded, with U+FFFD for bytes that are not UTF-8:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
for path in ["/hello", "/logo"] {
  let response = try Http.get("${base}${path}")
  assert(response.text() == String.fromBytesLossy(response.bytes))
}
```

## Possible issues

### A 404 is not an Err

`try` passes on every response, so a program that skips the status check reads the error page as if
it were the one it asked for. Here `/users/2` answers 404 with the text `not found: /users/2`, and
parsing it as JSON fails on that text, with a message that says nothing of the 404:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let response = try Http.get("${base}/users/2")
println(Json.parse(response.text()))
```

```text
Err{ error: Error{ message: "expected \"null\", found 'o' at line 1, column 2" } }
```

Check `status` before reading the body, as [Fetch JSON](#fetch-json) does.

### A timeout does not name the URL

A failed connection names the URL in its message, but a timeout's message is only `timeout`, and a
context's reason names no request either. Wrap the error with [Result.wrapErr](../Result/wrapErr.md)
to say which request failed. Here `/slow` answers after one second, and the wrap names its path:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/users/1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"id\": 1, \"name\": \"Ann\"}" }
  }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: none, body: "query: ${req.query}" } }
  if req.path == "/accept" { return Http.Reply{ status: 200, headers: none, body: req.headers.get("Accept") ?? "none" } }
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" } }
  if req.path == "/loop" { return Http.Reply{ status: 302, headers: Map{ "Location": "/loop" }, body: "" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  if req.path == "/error" { return Http.Reply{ status: 500, headers: none, body: "database down" } }
  if req.path == "/slow" {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "done" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
println(Http.get("${base}/slow", timeout: 100ms) | Result.wrapErr("get /slow"))
```

```text
Err{ error: Error{ message: "get /slow: timeout" } }
```

### A timeout in weeks

A week, a month, or a year has no fixed length, so a timeout with one is `Err`, as is one that is
not positive:

```ant
println(Http.get("http://127.0.0.1:1/", timeout: Time.duration(weeks: 1)))
println(Http.get("http://127.0.0.1:1/", timeout: 0s))
```

```text
Err{ error: Error{ message: "get \"http://127.0.0.1:1/\": the timeout has years, months or weeks, which have no fixed length" } }
Err{ error: Error{ message: "get \"http://127.0.0.1:1/\": the timeout is zero or negative" } }
```

### Header names that differ only in case

Header names are not case sensitive, so two that differ only in case name one header, and the call
is `Err`:

```ant
println(Http.get("http://127.0.0.1:1/", headers: Map{ "Accept": "text/html", "accept": "text/plain" }))
```

```text
Err{ error: Error{ message: "get \"http://127.0.0.1:1/\": the headers \"Accept\" and \"accept\" differ only in case" } }
```

### The result is a Result

The response is inside the `Result`, so reading a field of it directly is a compile error:

```ant
let status = Http.get("http://127.0.0.1:1/").status
// error: Result[Http.Response, Error] has no field status on every member: narrow the value first
```

## See also

[Http.request](request.md) · [Http.post](post.md) · [Http.serve](serve.md) ·
[Http.Response](types/Response.md) · [Url.escape](../Url/escape.md) · [Json.parse](../Json/parse.md)
· [Parallel.map](../Parallel/map.md) · [Context.new](../Context/new.md)
