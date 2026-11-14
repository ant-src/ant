<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http.Response

```ant signature
type Http.Response = { status: int, headers: Map[string, string], bytes: Bytes }
```

Http.Response is the response to a client request. Its body is held once, as received, in bytes;
[`Http.text`](../text.md) decodes it. The headers hold one string per name, a repeated header joined
with ", ", in order of name, each name in canonical form, as Go's CanonicalHeaderKey gives it:
"content-type" is read as headers\["Content-Type"\].

<!-- end generated -->

## Details

- Every status is a response: a 404 or a 500 comes back as `Ok` with that `status`. The call is
  `Err` only when no response arrives.
- No redirect is followed. A 3xx comes back as it is, with its `Location` header.
- The whole body is read before the call returns.
- `headers` is in order of name, and each name is in canonical form, as Go's `CanonicalHeaderKey`
  gives it: a header sent as `content-type` is read as `headers["Content-Type"]`.
- The body is held once, in `bytes`. [Http.text](../text.md) decodes it on every call, so a program
  that needs the text twice binds it once.

## Basic examples

Get a page from a local server with [Http.get](../get.md) and read its `status` and its body as
text:

```ant
func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain" }, body: "hello" }
}

let server = try Net.listen("127.0.0.1:0")
go { Http.serve(server, handle) }

let res = try Http.get("http://${Net.address(server)}/hello")
println(res.status)
println(res.text())
```

```text
200
hello
```

The examples below ask a server like this one, at `url`, which answers `/hello` with text, `/user`
with JSON, `/blob` with three bytes that are not UTF-8, `/old` with a redirect to `/hello`, and any
other path with a 404.

Look up a header by name:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain" }, body: "hello" }
  }
  if req.path == "/user" {
    return Http.Reply{
      status: 200,
      headers: Map{ "Content-Type": "application/json" },
      body: "{\"name\":\"Ann\",\"admin\":true}",
    }
  }
  if req.path == "/blob" {
    return Http.Reply{ status: 200, headers: Map{}, body: Bytes.fromArray([255, 0, 104]) }
  }
  if req.path == "/old" {
    return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let server = Result.unwrap(Net.listen("127.0.0.1:0"))
go { Http.serve(server, handle) }
let url = "http://${Net.address(server)}"
-->

```ant
let res = try Http.get("${url}/hello")
println(res.headers["Content-Type"])
println(res.headers["X-Missing"] ?? "none")
```

```text
text/plain
none
```

`bytes` holds the body as it arrived:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain" }, body: "hello" }
  }
  if req.path == "/user" {
    return Http.Reply{
      status: 200,
      headers: Map{ "Content-Type": "application/json" },
      body: "{\"name\":\"Ann\",\"admin\":true}",
    }
  }
  if req.path == "/blob" {
    return Http.Reply{ status: 200, headers: Map{}, body: Bytes.fromArray([255, 0, 104]) }
  }
  if req.path == "/old" {
    return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let server = Result.unwrap(Net.listen("127.0.0.1:0"))
go { Http.serve(server, handle) }
let url = "http://${Net.address(server)}"
-->

```ant
let res = try Http.get("${url}/hello")
println(res.bytes)
println(res.bytes.length())
```

```text
Bytes(68 65 6c 6c 6f)
5
```

## Scope

A 404 is a response like any other, with its status and its body:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain" }, body: "hello" }
  }
  if req.path == "/user" {
    return Http.Reply{
      status: 200,
      headers: Map{ "Content-Type": "application/json" },
      body: "{\"name\":\"Ann\",\"admin\":true}",
    }
  }
  if req.path == "/blob" {
    return Http.Reply{ status: 200, headers: Map{}, body: Bytes.fromArray([255, 0, 104]) }
  }
  if req.path == "/old" {
    return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let server = Result.unwrap(Net.listen("127.0.0.1:0"))
go { Http.serve(server, handle) }
let url = "http://${Net.address(server)}"
-->

```ant
let res = try Http.get("${url}/nope")
println(res.status)
println(res.text())
```

```text
404
no /nope
```

A redirect is not followed. The response is the redirect itself, and `Location` names the next page:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain" }, body: "hello" }
  }
  if req.path == "/user" {
    return Http.Reply{
      status: 200,
      headers: Map{ "Content-Type": "application/json" },
      body: "{\"name\":\"Ann\",\"admin\":true}",
    }
  }
  if req.path == "/blob" {
    return Http.Reply{ status: 200, headers: Map{}, body: Bytes.fromArray([255, 0, 104]) }
  }
  if req.path == "/old" {
    return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let server = Result.unwrap(Net.listen("127.0.0.1:0"))
go { Http.serve(server, handle) }
let url = "http://${Net.address(server)}"
-->

```ant
let res = try Http.get("${url}/old")
println(res.status)
println(res.headers["Location"])
println(res.text().isEmpty())
```

```text
301
/hello
true
```

A `HEAD` request gets the headers and an empty body:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain" }, body: "hello" }
  }
  if req.path == "/user" {
    return Http.Reply{
      status: 200,
      headers: Map{ "Content-Type": "application/json" },
      body: "{\"name\":\"Ann\",\"admin\":true}",
    }
  }
  if req.path == "/blob" {
    return Http.Reply{ status: 200, headers: Map{}, body: Bytes.fromArray([255, 0, 104]) }
  }
  if req.path == "/old" {
    return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let server = Result.unwrap(Net.listen("127.0.0.1:0"))
go { Http.serve(server, handle) }
let url = "http://${Net.address(server)}"
-->

```ant
let res = try Http.request("HEAD", "${url}/hello")
println(res.headers["Content-Length"])
println(res.text().isEmpty())
println(res.bytes.length())
```

```text
5
true
0
```

A body that is not UTF-8 is exact in `bytes`. In `text()`, the bad byte becomes U+FFFD:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain" }, body: "hello" }
  }
  if req.path == "/user" {
    return Http.Reply{
      status: 200,
      headers: Map{ "Content-Type": "application/json" },
      body: "{\"name\":\"Ann\",\"admin\":true}",
    }
  }
  if req.path == "/blob" {
    return Http.Reply{ status: 200, headers: Map{}, body: Bytes.fromArray([255, 0, 104]) }
  }
  if req.path == "/old" {
    return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let server = Result.unwrap(Net.listen("127.0.0.1:0"))
go { Http.serve(server, handle) }
let url = "http://${Net.address(server)}"
-->

```ant
let res = try Http.get("${url}/blob")
println(res.bytes)
println(res.text() == "\u{FFFD}\u{0}h")
```

```text
Bytes(ff 00 68)
true
```

A header sent twice is one entry, its values joined with `", "`. This server writes its response by
hand over [Net](../../Net/index.md) to send `Vary` twice:

```ant
let server = try Net.listen("127.0.0.1:0")
go {
  let conn = Result.unwrap(Net.accept(server))
  defer { conn.close() }
  while true {
    let line = Result.unwrap(conn.readLine())
    if line == null || line.trim().isEmpty() { break }
  }
  let reply = "HTTP/1.1 200 OK\r\nVary: Accept\r\nVary: Cookie\r\nContent-Length: 2\r\n\r\nok"
  Result.unwrap(conn.writeText(reply))
}

let res = try Http.get("http://${Net.address(server)}/")
println(res.headers["Vary"])
```

```text
Accept, Cookie
```

`Http.Response` is a record type, so a literal builds one, for example to test a function that takes
a response without a server:

```ant
func describe(res: Http.Response): string {
  if res.status >= 400 { return "error ${res.status}: ${res.text()}" }
  "${res.bytes.length()} bytes of ${res.headers["Content-Type"] ?? "unknown type"}"
}

let ok = Http.Response{
  status: 200,
  headers: Map{ "Content-Type": "text/html" },
  bytes: "<p>hi</p>".toBytes(),
}
let missing = Http.Response{ status: 404, headers: Map{}, bytes: "gone".toBytes() }
println(describe(ok))
println(describe(missing))
```

```text
9 bytes of text/html
error 404: gone
```

## Applications

Turn a status that is not 2xx into an `Err`, so `try` stops on it:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain" }, body: "hello" }
  }
  if req.path == "/user" {
    return Http.Reply{
      status: 200,
      headers: Map{ "Content-Type": "application/json" },
      body: "{\"name\":\"Ann\",\"admin\":true}",
    }
  }
  if req.path == "/blob" {
    return Http.Reply{ status: 200, headers: Map{}, body: Bytes.fromArray([255, 0, 104]) }
  }
  if req.path == "/old" {
    return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let server = Result.unwrap(Net.listen("127.0.0.1:0"))
go { Http.serve(server, handle) }
let url = "http://${Net.address(server)}"
-->

```ant
func fetchText(address: string): Result[string, Error] {
  let res = try Http.get(address)
  if res.status < 200 || res.status > 299 {
    return Result.fail("GET ${address}: status ${res.status}")
  }
  Result.ok(res.text())
}

println(fetchText("${url}/hello"))
println(fetchText("${url}/nope") is Err)
```

```text
Ok{ value: "hello" }
true
```

Parse a JSON body with [Json.parse](../../Json/parse.md):

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain" }, body: "hello" }
  }
  if req.path == "/user" {
    return Http.Reply{
      status: 200,
      headers: Map{ "Content-Type": "application/json" },
      body: "{\"name\":\"Ann\",\"admin\":true}",
    }
  }
  if req.path == "/blob" {
    return Http.Reply{ status: 200, headers: Map{}, body: Bytes.fromArray([255, 0, 104]) }
  }
  if req.path == "/old" {
    return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let server = Result.unwrap(Net.listen("127.0.0.1:0"))
go { Http.serve(server, handle) }
let url = "http://${Net.address(server)}"
-->

```ant
let res = try Http.get("${url}/user")
let user = try Json.parse(res.text())
if user is Map {
  println(user["name"])
}
```

```text
Ann
```

Follow a redirect by asking for the page that `Location` names:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain" }, body: "hello" }
  }
  if req.path == "/user" {
    return Http.Reply{
      status: 200,
      headers: Map{ "Content-Type": "application/json" },
      body: "{\"name\":\"Ann\",\"admin\":true}",
    }
  }
  if req.path == "/blob" {
    return Http.Reply{ status: 200, headers: Map{}, body: Bytes.fromArray([255, 0, 104]) }
  }
  if req.path == "/old" {
    return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let server = Result.unwrap(Net.listen("127.0.0.1:0"))
go { Http.serve(server, handle) }
let url = "http://${Net.address(server)}"
-->

```ant
func getFollowing(base: string, path: string): Result[Http.Response, Error] {
  let res = try Http.get("${base}${path}")
  let location = res.headers["Location"]
  if res.status < 300 || res.status > 399 || location == null {
    return Result.ok(res)
  }
  Http.get("${base}${location}")
}

let res = try getFollowing(url, "/old")
println(res.status)
println(res.text())
```

```text
200
hello
```

Check a download against a known digest, computed over `bytes`, the body as it was sent:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain" }, body: "hello" }
  }
  if req.path == "/user" {
    return Http.Reply{
      status: 200,
      headers: Map{ "Content-Type": "application/json" },
      body: "{\"name\":\"Ann\",\"admin\":true}",
    }
  }
  if req.path == "/blob" {
    return Http.Reply{ status: 200, headers: Map{}, body: Bytes.fromArray([255, 0, 104]) }
  }
  if req.path == "/old" {
    return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let server = Result.unwrap(Net.listen("127.0.0.1:0"))
go { Http.serve(server, handle) }
let url = "http://${Net.address(server)}"
-->

```ant
let res = try Http.get("${url}/blob")
let digest = Crypto.sha256(res.bytes) | Encoding.toHex()
println(digest == "12368291592ee4f015324a36b16c50a003ebe99204d222046f617b271515d290")
```

```text
true
```

## Properties and relations

`text()` is `bytes` decoded as text:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain" }, body: "hello" }
  }
  if req.path == "/user" {
    return Http.Reply{
      status: 200,
      headers: Map{ "Content-Type": "application/json" },
      body: "{\"name\":\"Ann\",\"admin\":true}",
    }
  }
  if req.path == "/blob" {
    return Http.Reply{ status: 200, headers: Map{}, body: Bytes.fromArray([255, 0, 104]) }
  }
  if req.path == "/old" {
    return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let server = Result.unwrap(Net.listen("127.0.0.1:0"))
go { Http.serve(server, handle) }
let url = "http://${Net.address(server)}"
-->

```ant
for path in ["/hello", "/user", "/blob", "/nope"] {
  let res = try Http.get("${url}${path}")
  assert(res.text() == String.fromBytesLossy(res.bytes))
}
```

## Possible issues

`try` does not stop on an error status, since every response is `Ok`. Check `status`:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain" }, body: "hello" }
  }
  if req.path == "/user" {
    return Http.Reply{
      status: 200,
      headers: Map{ "Content-Type": "application/json" },
      body: "{\"name\":\"Ann\",\"admin\":true}",
    }
  }
  if req.path == "/blob" {
    return Http.Reply{ status: 200, headers: Map{}, body: Bytes.fromArray([255, 0, 104]) }
  }
  if req.path == "/old" {
    return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let server = Result.unwrap(Net.listen("127.0.0.1:0"))
go { Http.serve(server, handle) }
let url = "http://${Net.address(server)}"
-->

```ant
let res = try Http.get("${url}/nope")
println("still here, status ${res.status}")
```

```text
still here, status 404
```

`text()` loses bytes that are not UTF-8, so turning it back into `Bytes` does not give the body that
was sent. Use `bytes` for binary data:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  if req.path == "/hello" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain" }, body: "hello" }
  }
  if req.path == "/user" {
    return Http.Reply{
      status: 200,
      headers: Map{ "Content-Type": "application/json" },
      body: "{\"name\":\"Ann\",\"admin\":true}",
    }
  }
  if req.path == "/blob" {
    return Http.Reply{ status: 200, headers: Map{}, body: Bytes.fromArray([255, 0, 104]) }
  }
  if req.path == "/old" {
    return Http.Reply{ status: 301, headers: Map{ "Location": "/hello" }, body: "" }
  }
  Http.Reply{ status: 404, headers: Map{}, body: "no ${req.path}" }
}
let server = Result.unwrap(Net.listen("127.0.0.1:0"))
go { Http.serve(server, handle) }
let url = "http://${Net.address(server)}"
-->

```ant
let res = try Http.get("${url}/blob")
println(res.text().toBytes() == res.bytes)
println(res.text().toBytes())
```

```text
false
Bytes(ef bf bd 00 68)
```

## See also

[Http.get](../get.md) · [Http.post](../post.md) · [Http.request](../request.md) ·
[Http.Request](Request.md) · [Http.Reply](Reply.md) · [Http.serve](../serve.md) ·
[String.fromBytesLossy](../../String/fromBytesLossy.md) · [Json.parse](../../Json/parse.md)
