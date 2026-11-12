<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http.request

```ant signature
func Http.request(method: string, url: string, body: string | Bytes = "", headers: Map[string, string] = Map{}, timeout: Time.Duration | null = null, ctx: Context | null = null): Result[Http.Response, Error]
```

Http.request sends one request with method, body, and headers to url and reads the whole response. A
response of any status is Ok, and no redirect is followed; Err means no response: a malformed url,
an empty method, header names that differ only in case, no connection, the timeout ("timeout"), or
ctx done. timeout bounds the whole call and is no deadline when null; one that is not positive or
has years, months, or weeks is Err; when both timeout and ctx are given, the earlier wins.

<!-- end generated -->

## Details

- A 404 or a 500 is `Ok` too. Check `status` before you use the body.
- The response's headers hold one string per name; a repeated header is joined with `, `.
- When `timeout` passes first, the result is `Err` with the message `"timeout"`. A `timeout` that is
  zero or negative, or that has years, months, or weeks, is `Err` before anything is sent.

## Basic examples

Most examples on this page talk to a small server on this machine, whose address is `base`. It
answers with the method and path it got, and the body after them.

Send a `PUT` with a body:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let res = try Http.request("PUT", "${base}/notes/1", body: "buy milk")
println(res.status)
println(res.body)
```

```text
200
PUT /notes/1 buy milk
```

Send a `DELETE`, which has no body:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let res = try Http.request("DELETE", "${base}/notes/1")
println("${res.status} ${res.body}")
```

```text
200 DELETE /notes/1
```

## Scope

The method is sent as given, so a method `Http` has no function for, such as `PATCH`, `OPTIONS`, or
a custom one, works the same way:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
for method in ["PATCH", "OPTIONS", "PURGE"] {
  let res = try Http.request(method, "${base}/cache")
  println(res.body)
}
```

```text
PATCH /cache
OPTIONS /cache
PURGE /cache
```

A status that reports a failure is still `Ok`. The response carries the status and the body the
server sent:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let res = try Http.request("GET", "${base}/missing")
println("${res.status}: ${res.body}")
```

```text
404: no /missing
```

A redirect is not followed. The 3xx response comes back with its `Location` header:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let res = try Http.request("GET", "${base}/old")
println(res.status)
println(res.headers["Location"])
```

```text
301
/new
```

A body can be `Bytes`, sent as they are. The response has the body twice, as text in `body` and as
received in `bytes`. Here the server sends back what it got:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let sent = Bytes.fromArray([0x41, 0x6e, 0x74, 0xff])
let res = try Http.request("POST", "${base}/echo", body: sent)
println(res.bytes == sent)
println(res.bytes.length())
println(res.body)
```

```text
true
4
Ant�
```

The byte `0xff` is not UTF-8, so `body` has U+FFFD in its place while `bytes` keeps it.

A query value with a space or an `&` must be escaped before it goes into the url.
[Url.stringifyQuery](../Url/stringifyQuery.md) escapes each key and value. Here the server answers
`/search` with the raw query it got:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let query = Url.stringifyQuery(Map{ "q": ["tea & milk"], "page": ["2"] })
let res = try Http.request("GET", "${base}/search?${query}")
println(res.body)
```

```text
q=tea+%26+milk&page=2
```

A request that gets no response is `Err`, with a message that says why:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
println(Http.request("", base))
println(Http.request("GET", base, headers: Map{ "Accept": "a", "accept": "b" }))
```

```text
Err{ error: Error{ message: "the method is empty" } }
Err{ error: Error{ message: "the headers \"Accept\" and \"accept\" differ only in case" } }
```

## Named arguments

### body

`body` is `""` by default, an empty body. A string is sent as its UTF-8:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let res = try Http.request("POST", "${base}/notes", body: "café")
println(res.body)
```

```text
POST /notes café
```

### headers

`headers` maps each header name to its value. Here the server echoes the `Content-Type` it got:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let headers = Map{ "Content-Type": "text/csv" }
let res = try Http.request("POST", "${base}/echo", body: "a,b\n1,2", headers: headers)
println(res.headers["Content-Type"])
```

```text
text/csv
```

### timeout

`timeout` bounds the whole call, from connecting to the end of the body. The server takes a second
to answer `/slow`:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let res = Http.request("GET", "${base}/slow", timeout: Time.duration(milliseconds: 50))
println(res)
```

```text
Err{ error: Error{ message: "timeout" } }
```

A timeout with weeks, months, or years has no fixed length, and one that is zero or negative has
already passed. Both are `Err` before anything is sent:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
println(Http.request("GET", base, timeout: Time.duration(weeks: 1)))
println(Http.request("GET", base, timeout: Time.duration(seconds: 0)))
```

```text
Err{ error: Error{ message: "the timeout has years, months or weeks, which have no fixed length" } }
Err{ error: Error{ message: "the timeout is zero or negative" } }
```

### ctx

`ctx` ends the call when the context is done, and the `Err` has the context's reason. A context that
is already cancelled makes the call `Err` with its reason, `cancelled`:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let ctx = Context.new()
ctx.cancel()
println(Http.request("GET", base, ctx: ctx))
```

```text
Err{ error: Error{ message: "cancelled" } }
```

One context can bound several requests at once, where a `timeout` bounds one. Here a half-second
budget covers every request of the loop, and the slow one runs out of it:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let budget = Context.new(timeout: Time.duration(milliseconds: 500))
for path in ["/a", "/slow", "/b"] {
  let res = Http.request("GET", "${base}${path}", ctx: budget)
  let line = if res is Ok { res.value.body } else { "${path}: ${res.error.message}" }
  println(line)
}
```

```text
GET /a
/slow: timeout
/b: timeout
```

## Applications

### Send and receive JSON

Encode a record with [Json.stringify](../Json/stringify.md), say what the body is in `Content-Type`,
and parse the answer with [Json.parse](../Json/parse.md). Here the server sends the document back:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
type Note = { title: string, done: bool }

let note = Note{ title: "buy milk", done: false }
let headers = Map{ "Content-Type": "application/json" }
let res = try Http.request("PUT", "${base}/echo", body: Json.stringify(note), headers: headers)
if res.status != 200 {
  panic("unexpected status ${res.status}")
}
let doc = try Json.parse(res.body)
println(res.headers["Content-Type"])
println(doc)
```

```text
application/json
Map{ "title": "buy milk", "done": false }
```

### Follow redirects

`Http.request` returns a redirect as it is, so a program that wants the next page asks for it. This
function follows up to five redirects to paths on the same server:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
func fetch(method: string, url: string): Result[Http.Response, Error] {
  var next = url
  for _ in Iter.range(0, 6) {
    let res = try Http.request(method, next)
    let location = res.headers.get("Location")
    if res.status < 300 || res.status > 399 || location == null {
      return Result.ok(res)
    }
    next = "${base}${location}"
  }
  Result.fail("too many redirects from ${url}")
}

let res = try fetch("GET", "${base}/old")
println("${res.status} ${res.body}")
```

```text
200 the new page
```

### Turn a status into an error

Since any status is `Ok`, a function that needs a success says so itself. Here a 404 becomes an
`Err` with the method, the path, and the server's answer:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
func send(method: string, path: string, body: string = ""): Result[string, Error] {
  let res = try (
    Http.request(method, "${base}${path}", body: body) | Result.wrapErr("${method} ${path}")
  )
  if res.status >= 400 {
    return Result.fail("${method} ${path}: status ${res.status}: ${res.body}")
  }
  Result.ok(res.body)
}

println(send("PUT", "/notes/2", body: "call Ann"))
println(send("GET", "/missing"))
```

```text
Ok{ value: "PUT /notes/2 call Ann" }
Err{ error: Error{ message: "GET /missing: status 404: no /missing" } }
```

### Test your own server

[Http.serve](serve.md) takes a listener, so a program can start a server on a free port and send it
requests. Every line of this example is shown:

```ant
func handle(req: Http.Request): Http.Reply {
  if req.method != "POST" {
    return Http.Reply{ status: 405, headers: Map{ "Allow": "POST" }, body: "" }
  }
  Http.Reply{ status: 201, headers: Map{}, body: req.body.toUpper() }
}

let listener = try Net.listen("127.0.0.1:0")
let url = "http://${Net.address(listener)}/shout"
go { Http.serve(listener, handle) }

let created = try Http.request("POST", url, body: "hi")
println("${created.status} ${created.body}")
let refused = try Http.request("PATCH", url)
println("${refused.status} allow ${refused.headers["Allow"]}")
```

```text
201 HI
405 allow POST
```

## Properties and relations

[Http.get](get.md) and [Http.post](post.md) are `Http.request` with the method filled in:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let get = try Http.get("${base}/a")
let getVia = try Http.request("GET", "${base}/a")
assert(get.status == getVia.status)
assert(get.body == getVia.body)

let post = try Http.post("${base}/a", "x")
let postVia = try Http.request("POST", "${base}/a", body: "x")
assert(post.body == postVia.body)
```

When the body is UTF-8, `body` is the text of `bytes`:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let res = try Http.request("POST", "${base}/echo", body: "naïve ☃")
assert(res.body == "naïve ☃")
assert(res.bytes == res.body.toBytes())
```

## Possible issues

`Result.unwrap` does not catch a failed request that got a response. The value is a 404 response,
and the program goes on with its body:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let res = Result.unwrap(Http.request("GET", "${base}/missing"))
println(res.body)
```

```text
no /missing
```

Header names are not case-sensitive in HTTP, so two names that differ only in case name one header,
and the request is `Err`. Keep one spelling per header when you merge maps of headers:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let defaults = Map{ "Content-Type": "text/plain" }
let extra = Map{ "content-type": "text/csv" }
let headers = Map.merge(defaults, extra)
println(Http.request("POST", "${base}/echo", body: "a,b", headers: headers))
```

```text
Err{ error: Error{ message: "the headers \"Content-Type\" and \"content-type\" differ only in case" } }
```

The method is sent as given, not upper-cased, so `"get"` reaches the server as `get`, and a server
that compares it with `"GET"` does not match it:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/old" { return Http.Reply{ status: 301, headers: Map{ "Location": "/new" }, body: "" } }
  if req.path == "/new" { return Http.Reply{ status: 200, headers: Map{}, body: "the new page" } }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no /missing" } }
  if req.path == "/search" { return Http.Reply{ status: 200, headers: Map{}, body: req.query } }
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 1)) }
  if req.path == "/echo" {
    let kind = req.headers.get("Content-Type") ?? "none"
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": kind }, body: req.bytes }
  }
  let body = if req.body.isEmpty() { "" } else { " ${req.body}" }
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}${body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let res = try Http.request("get", "${base}/a")
println(res.body)
```

```text
get /a
```

## See also

[Http.get](get.md) · [Http.post](post.md) · [Http.serve](serve.md) ·
[Http.Response](types/Response.md) · [Net.listen](../Net/listen.md) ·
[Context.new](../Context/new.md) · [Json.stringify](../Json/stringify.md) ·
[Url.escape](../Url/escape.md) · [Url.stringifyQuery](../Url/stringifyQuery.md)
