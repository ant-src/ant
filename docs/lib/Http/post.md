<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http.post

```ant signature
func Http.post(url: string, body: string | Bytes, headers: Map[string, string] = Map{}, timeout: Time.Duration | null = null, ctx: Context | null = null): Result[Http.Response, Error]
```

Http.post sends body to url in a POST request and reads the whole response. It is
[`Http.request`](request.md) with the method "POST".

<!-- end generated -->

## Details

- A response of any status is `Ok`: a 404 or a 500 comes back with its `status`, and the program
  checks it. `Err` means there was no response at all, such as no connection, a malformed URL, the
  timeout, or the context.
- No redirect is followed. A 3xx comes back as it is, with its `Location` header.
- The body is sent as given: a string as its UTF-8, `Bytes` unchanged. Say what it is with a
  `Content-Type` header.
- The whole response is read before `post` returns, and its body is there twice: as received in
  `bytes`, and as text in `body`.

## Basic examples

Start a server on a free local port, then post a string to it with `Http.post`:

```ant
func echo(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path}: ${req.body}" }
}

let listener = try Net.listen("127.0.0.1:0")
go { Http.serve(listener, echo) }
let res = try Http.post("http://${listener.address()}/notes", "buy milk")
println(res.status)
println(res.body)
```

```text
200
POST /notes: buy milk
```

The rest of this page posts to `url`, a local server like the one above. It answers with the method,
the path, the `Content-Type` it got, and the body; `/missing` answers 404, `/moved` redirects,
`/slow` takes two seconds, `/raw` sends the bytes it got back unchanged, and `/json` sends the body
back as JSON.

Name the type of the body with a header:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
let res = try Http.post("${url}/notes", "buy milk", headers: Map{ "Content-Type": "text/plain" })
println(res.body)
```

```text
POST /notes (text/plain): buy milk
```

## Scope

The body is a string or `Bytes`. An empty string sends an empty body, which `/raw` sends back:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
let text = try Http.post(url, "text")
println(text.body)
let bytes = try Http.post(url, Bytes.fromArray([0x68, 0x69]))
println(bytes.body)
let empty = try Http.post("${url}/raw", "")
println(empty.bytes)
```

```text
POST / (no type): text
POST / (no type): hi
Bytes()
```

`Bytes` are sent as they are, even when they are not UTF-8. The response keeps them in `bytes`,
while `body` has U+FFFD where the text is not valid:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
let res = try Http.post("${url}/raw", Bytes.fromArray([0xFF, 0x41]))
println(res.bytes)
println(res.body)
```

```text
Bytes(ff 41)
�A
```

A status that reports a failure is still `Ok`:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
let res = try Http.post("${url}/missing", "x")
println(res.status)
println(res.body)
```

```text
404
no such note
```

A redirect comes back as it is, and the program decides whether to follow it:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
let res = try Http.post("${url}/moved", "x")
println(res.status)
println(res.headers.get("Location"))
```

```text
303
/notes/7
```

## Named arguments

Two names that differ only in case name one header, so the call is `Err` and sends nothing:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
println(Http.post("http://127.0.0.1:1/", "x", headers: Map{ "Accept": "text/plain", "accept": "*/*" }))
```

```text
Err{ error: Error{ message: "post \"http://127.0.0.1:1/\": the headers \"Accept\" and \"accept\" differ only in case" } }
```

`timeout` bounds the whole call, from connecting to the end of the body. When it passes, the result
is `Err` with the message `timeout`:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
println(Http.post("${url}/slow", "x", timeout: Time.duration(milliseconds: 100)))
```

```text
Err{ error: Error{ message: "timeout" } }
```

`ctx` bounds the call too, and the result is `Err` with the context's reason. A context shares one
bound across several calls:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
let ctx = Context.new(timeout: Time.duration(milliseconds: 100))
println(Http.post("${url}/slow", "first", ctx: ctx))
println(Http.post(url, "second", ctx: ctx))
```

```text
Err{ error: Error{ message: "timeout" } }
Err{ error: Error{ message: "timeout" } }
```

When both are given, the earlier bound wins:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
let ctx = Context.new(timeout: Time.duration(seconds: 10))
println(Http.post("${url}/slow", "x", timeout: Time.duration(milliseconds: 100), ctx: ctx))
```

```text
Err{ error: Error{ message: "timeout" } }
```

A context cancelled by the program gives the reason `cancelled`:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
let ctx = Context.new()
ctx.cancel()
println(Http.post(url, "x", ctx: ctx))
```

```text
Err{ error: Error{ message: "cancelled" } }
```

## Applications

Send a record as JSON and read the JSON reply:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
type Note = { title: string, done: bool }

let note = Note{ title: "buy milk", done: false }
let headers = Map{ "Content-Type": "application/json" }
let res = try Http.post("${url}/json", Json.stringify(note), headers: headers)
println(res.status)
println(try Json.parse(res.body))
```

```text
201
Map{ "title": "buy milk", "done": false }
```

Post a form, encoded with [Url.stringifyQuery](../Url/stringifyQuery.md):

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
let form = Url.stringifyQuery(Map{ "name": ["Ann Lee"], "tag": ["a", "b"] })
let headers = Map{ "Content-Type": "application/x-www-form-urlencoded" }
println((try Http.post("${url}/signup", form, headers: headers)).body)
```

```text
POST /signup (application/x-www-form-urlencoded): name=Ann+Lee&tag=a&tag=b
```

Turn a failing status into an `Err`, and name the call in every error with
[Result.wrapErr](../Result/wrapErr.md):

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
func create(endpoint: string, title: string): Result[string, Error] {
  let res = try (Http.post(endpoint, title) | Result.wrapErr("create ${title}"))
  if res.status != 200 {
    return Result.fail("create ${title}: status ${res.status}")
  }
  Result.ok(res.body)
}

println(create("${url}/notes", "buy milk"))
println(create("${url}/missing", "call Bob"))
```

```text
Ok{ value: "POST /notes (no type): buy milk" }
Err{ error: Error{ message: "create call Bob: status 404" } }
```

## Properties and relations

`Http.post` is [Http.request](request.md) with the method `"POST"`:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
let a = try Http.post("${url}/notes", "x")
let b = try Http.request("POST", "${url}/notes", body: "x")
assert(a.status == b.status)
assert(a.body == b.body)
```

When the response is UTF-8, `body` is its `bytes` as text:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
let res = try Http.post(url, "héllo")
assert(res.body.toBytes() == res.bytes)
```

## Possible issues

A zero or negative `timeout` is `Err` before anything is sent, as is one with years, months, or
weeks, which have no fixed length:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
println(Http.post("http://127.0.0.1:1/", "x", timeout: Time.duration(seconds: 0)))
println(Http.post("http://127.0.0.1:1/", "x", timeout: Time.duration(weeks: 1)))
```

```text
Err{ error: Error{ message: "post \"http://127.0.0.1:1/\": the timeout is zero or negative" } }
Err{ error: Error{ message: "post \"http://127.0.0.1:1/\": the timeout has years, months or weeks, which have no fixed length" } }
```

`try` returns early only when there is no response. A 404 is `Ok`, so check `status` before using
the body:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
let res = try Http.post("${url}/missing", "x")
let text = if res.status == 200 { res.body } else { "failed: ${res.status}" }
println(text)
```

```text
failed: 404
```

A malformed URL is `Err`:

<!-- test: before
func echo(req: Http.Request): Http.Reply {
  if req.path == "/slow" { Time.sleep(Time.duration(seconds: 2)) }
  if req.path == "/missing" { return Http.Reply{ status: 404, headers: Map{}, body: "no such note" } }
  if req.path == "/moved" { return Http.Reply{ status: 303, headers: Map{ "Location": "/notes/7" }, body: "" } }
  if req.path == "/raw" { return Http.Reply{ status: 200, headers: Map{}, body: req.bytes } }
  if req.path == "/json" {
    return Http.Reply{ status: 201, headers: Map{ "Content-Type": "application/json" }, body: req.body }
  }
  let kind = req.headers.get("Content-Type") ?? "no type"
  Http.Reply{ status: 200, headers: Map{}, body: "${req.method} ${req.path} (${kind}): ${req.body}" }
}
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
let url = "http://${listener.address()}"
go { Http.serve(listener, echo) }
-->

```ant
println(Http.post("http://example.com/%zz", "x"))
```

```text
Err{ error: Error{ message: "post \"http://example.com/%zz\": invalid URL escape \"%zz\"" } }
```

## See also

[Http.request](request.md) · [Http.get](get.md) · [Http.serve](serve.md) ·
[Http.Response](types/Response.md) · [Json.stringify](../Json/stringify.md) ·
[Url.stringifyQuery](../Url/stringifyQuery.md) · [Context.new](../Context/new.md)
