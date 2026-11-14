<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http.text

```ant signature
func Http.text(r: Http.Response | Http.Request): string
```

Http.text returns the body of r decoded as UTF-8, with bytes that are not valid UTF-8 as U+FFFD, as
all text from outside is. It decodes on every call, so a program that needs the text twice binds it
once.

<!-- end generated -->

## Details

- `Http.text` decodes the body as UTF-8 whatever the `Content-Type` header says. Each malformed
  sequence becomes U+FFFD, so `Http.text` never fails, and the bytes it replaced cannot be recovered
  from the text. `bytes` keeps the body as received.
- A byte order mark at the start of the body is kept, as U+FEFF.

## Basic examples

Get a page from a local server and decode its body:

```ant
let listener = try Net.listen("127.0.0.1:0")
go { Http.serve(listener, { Http.Reply{ status: 200, headers: Map{}, body: "Hello, ${x.path}" } }) }

let response = try Http.get("http://${Net.address(listener)}/world")
println(Http.text(response))
```

```text
Hello, /world
```

In a handler, `Http.text` decodes the body of the request:

```ant
func handle(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: Http.text(req).toUpper() }
}

let listener = try Net.listen("127.0.0.1:0")
go { Http.serve(listener, handle) }

let response = try Http.post("http://${Net.address(listener)}/shout", "hello")
println(response.text())
```

```text
HELLO
```

## Scope

### A response or a request

`Http.text` takes an `Http.Response` or an `Http.Request`. A record built by hand decodes like one
that arrived, which shows what the decoding does without a server:

```ant
let none: Map[string, string] = Map{}
let response = Http.Response{ status: 200, headers: none, bytes: "héllo".toBytes() }
let request = Http.Request{
  method: "POST",
  path: "/",
  query: "",
  headers: none,
  bytes: "a=1".toBytes(),
  remote: "127.0.0.1:5000",
}
println(Http.text(response))
println(Http.text(request))
```

```text
héllo
a=1
```

### Bytes that are not UTF-8

Each malformed sequence becomes U+FFFD. The bytes `FF FE` are two such sequences, and `E2 82`, a
euro sign cut short, is one:

```ant
let none: Map[string, string] = Map{}
let broken = Http.Response{ status: 200, headers: none, bytes: Bytes.fromArray([0xFF, 0xFE, 0x68, 0x69]) }
let cut = Http.Response{ status: 200, headers: none, bytes: Bytes.fromArray([0x31, 0xE2, 0x82]) }
println(broken.text() == "\u{FFFD}\u{FFFD}hi")
println(cut.text() == "1\u{FFFD}")
```

```text
true
true
```

### A byte order mark

A byte order mark at the start stays in the text. Remove it with
[String.trimPrefix](../String/trimPrefix.md):

```ant
let none: Map[string, string] = Map{}
let marked = Http.Response{ status: 200, headers: none, bytes: Bytes.fromArray([0xEF, 0xBB, 0xBF, 0x68, 0x69]) }
let text = marked.text()
println(text.length())
println(text.trimPrefix("\u{FEFF}"))
```

```text
3
hi
```

### An empty body

A response without a body decodes to `""`. Here `/empty` answers 204:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/user" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/todo" { return Http.Reply{ status: 200, headers: none, body: "milk\neggs\nbread\n" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/latin1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain; charset=iso-8859-1" }, body: Bytes.fromArray([0x63, 0x61, 0x66, 0xE9]) }
  }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
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
```

```text
204
true
```

### As a function value

`Http.text` is a function value like any other, so it can be passed to
[Result.map](../Result/map.md). Here `/hello` answers `Hello, Ant`:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/user" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/todo" { return Http.Reply{ status: 200, headers: none, body: "milk\neggs\nbread\n" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/latin1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain; charset=iso-8859-1" }, body: Bytes.fromArray([0x63, 0x61, 0x66, 0xE9]) }
  }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
println(Http.get("${base}/hello").map(Http.text))
```

```text
Ok{ value: "Hello, Ant" }
```

## Applications

### Parse a JSON response

Decode the body, then parse it with [Json.parse](../Json/parse.md). Here `/user` answers
`{"name": "Ann", "age": 31}`:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/user" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/todo" { return Http.Reply{ status: 200, headers: none, body: "milk\neggs\nbread\n" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/latin1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain; charset=iso-8859-1" }, body: Bytes.fromArray([0x63, 0x61, 0x66, 0xE9]) }
  }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let response = try Http.get("${base}/user")
let user = try Json.parse(response.text())
println(user)
```

```text
Map{ "name": "Ann", "age": 31.0 }
```

### Read a body line by line

Here `/todo` answers three lines:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/user" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/todo" { return Http.Reply{ status: 200, headers: none, body: "milk\neggs\nbread\n" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/latin1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain; charset=iso-8859-1" }, body: Bytes.fromArray([0x63, 0x61, 0x66, 0xE9]) }
  }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let response = try Http.get("${base}/todo")
for item in response.text().lines() {
  println("- ${item}")
}
```

```text
- milk
- eggs
- bread
```

### Read a form in a handler

A form posted as `application/x-www-form-urlencoded` has the same shape as a query string, so
[Url.parseQuery](../Url/parseQuery.md) reads the decoded body:

```ant
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  let form = Url.parseQuery(req.text())
  if form is Err { return Http.Reply{ status: 400, headers: none, body: form.error.message } }
  let name = form.value.get("name")?.first() ?? "stranger"
  Http.Reply{ status: 200, headers: none, body: "Hello, ${name}" }
}

let listener = try Net.listen("127.0.0.1:0")
go { Http.serve(listener, handle) }

let base = "http://${Net.address(listener)}"
let form = Map{ "Content-Type": "application/x-www-form-urlencoded" }
println((try Http.post("${base}/greet", "name=Ann+Lee", headers: form)).text())
println((try Http.post("${base}/greet", "", headers: form)).text())
```

```text
Hello, Ann Lee
Hello, stranger
```

### Tell text from binary

[String.fromBytes](../String/fromBytes.md) is `null` when the bytes are not UTF-8, which `Http.text`
cannot tell you. Here `/hello` answers text and `/logo` the four bytes a PNG file starts with:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/user" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/todo" { return Http.Reply{ status: 200, headers: none, body: "milk\neggs\nbread\n" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/latin1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain; charset=iso-8859-1" }, body: Bytes.fromArray([0x63, 0x61, 0x66, 0xE9]) }
  }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
for path in ["/hello", "/logo"] {
  let response = try Http.get("${base}${path}")
  let text = String.fromBytes(response.bytes)
  println(if text == null { "${path}: binary" } else { "${path}: text" })
}
```

```text
/hello: text
/logo: binary
```

## Properties and relations

`Http.text` decodes `bytes` as [String.fromBytesLossy](../String/fromBytesLossy.md) does. For a body
that is UTF-8, encoding the text gives the bytes back:

```ant
let none: Map[string, string] = Map{}
let response = Http.Response{ status: 200, headers: none, bytes: "naïve café".toBytes() }
assert(Http.text(response) == String.fromBytesLossy(response.bytes))
assert(response.text().toBytes() == response.bytes)
assert(response.text() == String.fromBytes(response.bytes))
```

## Possible issues

### The text loses bytes that are not UTF-8

Each malformed sequence becomes the three bytes of U+FFFD, so the text of a binary body does not
encode back to the body. Use `bytes` for an image, or for a signature computed over what was sent.
Here `/logo` answers the four bytes a PNG file starts with:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/user" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/todo" { return Http.Reply{ status: 200, headers: none, body: "milk\neggs\nbread\n" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/latin1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain; charset=iso-8859-1" }, body: Bytes.fromArray([0x63, 0x61, 0x66, 0xE9]) }
  }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let response = try Http.get("${base}/logo")
println(response.bytes)
println(response.text().toBytes())
```

```text
Bytes(89 50 4e 47)
Bytes(ef bf bd 50 4e 47)
```

### The charset in Content-Type is not used

A body in another encoding decodes as UTF-8 all the same. Here `/latin1` answers `café` in
ISO-8859-1, where `é` is the single byte `E9`, which is not UTF-8:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/user" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/todo" { return Http.Reply{ status: 200, headers: none, body: "milk\neggs\nbread\n" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/latin1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain; charset=iso-8859-1" }, body: Bytes.fromArray([0x63, 0x61, 0x66, 0xE9]) }
  }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let response = try Http.get("${base}/latin1")
println(response.headers["Content-Type"])
println(response.text() == "caf\u{FFFD}")
```

```text
text/plain; charset=iso-8859-1
true
```

### A reply has no text

`Http.text` takes what arrives, not what a handler sends. An `Http.Reply` holds its body in `body`,
already a string or `Bytes`:

```ant
let reply = Http.Reply{ status: 200, headers: Map{}, body: "hi" }
println(Http.text(reply)) // error: cannot use Http.Reply where Http.Response | Http.Request is expected
```

### Decoding twice

Each call decodes the whole body again. Bind the text once and use the name:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path == "/hello" { return Http.Reply{ status: 200, headers: none, body: "Hello, Ant" } }
  if req.path == "/user" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "application/json" }, body: "{\"name\": \"Ann\", \"age\": 31}" }
  }
  if req.path == "/todo" { return Http.Reply{ status: 200, headers: none, body: "milk\neggs\nbread\n" } }
  if req.path == "/logo" { return Http.Reply{ status: 200, headers: none, body: Bytes.fromArray([0x89, 0x50, 0x4E, 0x47]) } }
  if req.path == "/latin1" {
    return Http.Reply{ status: 200, headers: Map{ "Content-Type": "text/plain; charset=iso-8859-1" }, body: Bytes.fromArray([0x63, 0x61, 0x66, 0xE9]) }
  }
  if req.path == "/empty" { return Http.Reply{ status: 204, headers: none, body: "" } }
  Http.Reply{ status: 404, headers: none, body: "not found: ${req.path}" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
let response = try Http.get("${base}/todo")
let text = response.text()
println(text.lines().length())
println(text.count("e"))
```

```text
3
2
```

## See also

[Http.Response](types/Response.md) · [Http.Request](types/Request.md) · [Http.get](get.md) ·
[Http.serve](serve.md) · [String.fromBytesLossy](../String/fromBytesLossy.md) ·
[String.fromBytes](../String/fromBytes.md) · [Bytes](../Bytes/index.md)
