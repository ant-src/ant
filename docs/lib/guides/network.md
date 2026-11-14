<!-- status: reviewed -->

# Network

Three namespaces cover the network. [Http](../Http/index.md) sends requests and serves them,
[Net](../Net/index.md) opens TCP listeners and connections, and [Url](../Url/index.md) takes URLs
and query strings apart and puts them back together. They meet at two points: `Http.serve` answers
requests on a `Net.Listener`, and a handler reads its raw `query` with `Url.parseQuery`. `Net` is
TCP only, with no TLS, UDP, or Unix sockets. The client speaks `http` and `https`, and `Http.serve`
speaks plain HTTP.

```ant
func search(req: Http.Request): Http.Reply {
  let query = Url.parseQuery(req.query)
  if query is Err { return Http.Reply{ status: 400, headers: Map{}, body: query.error.message } }
  let q = query.value["q"]?.first() ?? ""
  Http.Reply{ status: 200, headers: Map{}, body: "${req.path}: ${q}" }
}

let listener = try Net.listen("127.0.0.1:0")
go { Http.serve(listener, search) }

let url = try Url.parse("http://${listener.address()}/find")
url.query = Url.stringifyQuery(Map{ "q": ["ant & bee"] })
let res = try Http.get(url.stringify())
if res.status != 200 { panic("status ${res.status}") }
println(res.text())
```

```text
/find: ant & bee
```

A TCP connection is a `File.Handle`, so the `File` functions read and write it. A line protocol is a
loop over `readLine`:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  let conn = Result.unwrap(listener.accept())
  defer { conn.close() }
  let line = Result.unwrap(conn.readLine())
  Result.unwrap(conn.writeLine("hello, ${line ?? "nobody"}"))
}

let conn = try Net.dial(listener.address())
defer { conn.close() }
try conn.writeLine("Ann")
println(try conn.readLine())
```

```text
hello, Ann
```

Every call that connects, listens, accepts, sends, or receives returns a `Result`: a failure to
connect, a timeout, or a done context is `Err`. An HTTP response of any status is `Ok`, so check
`status` before you read the body. The client follows no redirect: a 3xx comes back as it is, with
its `Location` header. Header names arrive in canonical form, so read `res.headers["Content-Type"]`,
whatever case the server sent.

## Functions by job

### Sending HTTP requests

- [Http.get](../Http/get.md): a GET request, and the whole response
- [Http.post](../Http/post.md): a POST request with a body
- [Http.request](../Http/request.md): a request with any method, body, and headers
- [Http.Response](../Http/types/Response.md): what the client gets back: status, headers, and the
  body as `bytes`

### Serving HTTP

- [Http.serve](../Http/serve.md): answer every request on a listener with a handler, each on a
  goroutine of its own
- [Http.Request](../Http/types/Request.md): what a handler gets: method, path, query, headers, body,
  and the client's address
- [Http.Reply](../Http/types/Reply.md): what a handler returns: status, headers, and a body of text
  or `Bytes`

### Reading a body

- [Http.text](../Http/text.md): the body of a response or a request, decoded as text

### Listening for connections

- [Net.listen](../Net/listen.md): a TCP listener on `host:port`, any free port with `:0`
- [Net.address](../Net/address.md): the address a listener listens on
- [Net.accept](../Net/accept.md): the next connection that arrives on a listener
- [Net.close](../Net/close.md): stop a listener, so pending and later accepts are `Err`
- [Net.Listener](../Net/types/Listener.md): a listening TCP socket

### Connecting

- [Net.dial](../Net/dial.md): a TCP connection to `host:port`
- [Net.peer](../Net/peer.md): the address at the other end of a connection

### Talking over a connection

- [File.readLine](../File/readLine.md): the next line, or `null` at the end
- [File.readAll](../File/readAll.md): everything until the other end closes
- [File.writeLine](../File/writeLine.md): a line of text and a line break
- [File.writeText](../File/writeText.md): text as it is
- [File.pipe](../File/pipe.md): copy one handle into another until its end
- [File.close](../File/close.md): close the connection

### Whole URLs

- [Url.parse](../Url/parse.md): a reference divided into its parts, `Err` when it is malformed
- [Url.stringify](../Url/stringify.md): the reference a set of parts makes
- [Url.Parts](../Url/types/Parts.md): the record of parts, from `scheme` to `fragment`

### Query strings

- [Url.parseQuery](../Url/parseQuery.md): `k=v&k=v` as a map from each key to its values
- [Url.stringifyQuery](../Url/stringifyQuery.md): a map of keys and values written as a query
- [Url.escape](../Url/escape.md): one query key or value with `%XX` escapes and a space as `+`
- [Url.unescape](../Url/unescape.md): the reverse, `Err` on a bad escape

### Bounding and stopping

- `timeout:` on [Http.request](../Http/request.md), [Http.get](../Http/get.md), and
  [Http.post](../Http/post.md): a bound on one request, from connecting to the end of the body
- [Context](../Context/index.md): a deadline or a cancel shared by several calls
- [Os.interrupt](../Os/interrupt.md): a context done on Ctrl-C, to stop a server that runs on the
  main path

## See also

[Http](../Http/index.md) · [Net](../Net/index.md) · [Url](../Url/index.md) ·
[Files and paths](files.md) · [Concurrency](concurrency.md) · [Data formats](data-formats.md)
