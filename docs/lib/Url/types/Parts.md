<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Url.Parts

```ant signature
type Url.Parts = { scheme: string, opaque: string, user: string, host: string, port: string, path: string, query: string, fragment: string }
```

Url.Parts is a URL reference divided as RFC 3986 does, with each absent part "". scheme is in lower
case without its colon, opaque is what follows "scheme:" when that does not start with /, as in a
mailto: address, and user is the userinfo as written. path and fragment are decoded, but path keeps
%2F and %25 as written, while query is the raw text after the ?, for
[`Url.parseQuery`](../parseQuery.md).

<!-- end generated -->

## Details

- A part the reference does not have is `""`, never `null`.
- `path` and `fragment` are decoded, but `path` keeps `%2F` and `%25` as written. `query` is raw:
  read its keys and values with [Url.parseQuery](../parseQuery.md).
- `user` holds the whole userinfo as written, a password included.
- The fields are mutable, and a literal must give all eight. To change one part, parse the
  reference, set the field or spread it into a new literal, and write it back with
  [Url.stringify](../stringify.md).
- `Url.stringify` escapes every part but `query` and `opaque`, which it writes as they are; build
  `query` with [Url.stringifyQuery](../stringifyQuery.md).

## Basic examples

[Url.parse](../parse.md) divides a reference into a `Url.Parts`:

```ant
let parts = try Url.parse("https://ann@example.com:8080/docs/intro?lang=en#setup")
println(parts)
```

```text
Url.Parts{ scheme: "https", opaque: "", user: "ann", host: "example.com", port: "8080", path: "/docs/intro", query: "lang=en", fragment: "setup" }
```

Read the parts you need as fields:

```ant
let parts = try Url.parse("https://example.com/search?q=ant")
println(parts.host)
println(parts.path)
println(parts.query)
```

```text
example.com
/search
q=ant
```

[Url.stringify](../stringify.md) turns the parts back into a reference:

```ant
let parts = try Url.parse("https://example.com/search?q=ant")
println(Url.stringify(parts))
```

```text
https://example.com/search?q=ant
```

## Scope

A relative reference has an empty `scheme` and `host`, and only the parts it writes:

```ant
let link = try Url.parse("../img/logo.png")
let search = try Url.parse("/search?page=2")
let anchor = try Url.parse("#intro")
println("${link.path} ${search.path} ${search.query} ${anchor.fragment}")
println(link.scheme.isEmpty() && link.host.isEmpty())
```

```text
../img/logo.png /search page=2 intro
true
```

A reference that starts with `//` has a host but no scheme:

```ant
let parts = try Url.parse("//cdn.example.com/lib.js")
println("[${parts.scheme}] [${parts.host}] [${parts.path}]")
```

```text
[] [cdn.example.com] [/lib.js]
```

`scheme` is in lower case, without its colon:

```ant
println((try Url.parse("HTTPS://example.com")).scheme)
```

```text
https
```

A reference whose content after `scheme:` does not start with `/`, as a `mailto:` address, keeps
that content in `opaque`, and its `path` is empty:

```ant
let parts = try Url.parse("mailto:alice@example.com")
println(parts.scheme)
println(parts.opaque)
println(parts.path.isEmpty())
```

```text
mailto
alice@example.com
true
```

A single `/` after the colon starts the path, so the authority is empty, and so is `host`. Three
slashes in a `file:` URL give the same parts:

```ant
let short = try Url.parse("file:/etc/hosts")
let long = try Url.parse("file:///etc/hosts")
println("[${short.host}] ${short.path}")
println(short == long)
```

```text
[] /etc/hosts
true
```

`user` holds the userinfo as written, a password included, and `port` is the text after the colon:

```ant
let parts = try Url.parse("ftp://ann:s3cret@files.example.com:2121/pub")
println(parts.user)
println(parts.port)
println(parts.port.toInt())
```

```text
ann:s3cret
2121
2121
```

`path` and `fragment` are decoded. `query` is the raw text after `?`:

```ant
let parts = try Url.parse("https://example.com/my%20files/caf%C3%A9?q=a%20b#part%202")
println(parts.path)
println(parts.query)
println(parts.fragment)
```

```text
/my files/café
q=a%20b
part 2
```

`%2F` and `%25` stay escaped in `path`, so a slash inside a segment does not become a separator:

```ant
let parts = try Url.parse("/files/a%2Fb/100%25")
println(parts.path)
println(parts.path.split("/"))
```

```text
/files/a%2Fb/100%25
["", "files", "a%2Fb", "100%25"]
```

`Url.Parts` is a record type, so a literal builds one, with every field given:

```ant
let parts = Url.Parts{
  scheme: "https",
  opaque: "",
  user: "",
  host: "example.com",
  port: "",
  path: "/a b",
  query: "q=1",
  fragment: "",
}
println(Url.stringify(parts))
```

```text
https://example.com/a%20b?q=1
```

## Applications

Change one part of a URL and keep the rest, with a spread:

```ant
func withHost(url: string, host: string): Result[string, Error] {
  let parts = try Url.parse(url)
  Result.ok(Url.stringify(Url.Parts{ ...parts, host }))
}

println(withHost("https://staging.example.com/api/v1?page=2", "example.com"))
```

```text
Ok{ value: "https://example.com/api/v1?page=2" }
```

Set a field in place, here to drop the password before a URL goes into a log:

```ant
func redact(url: string): Result[string, Error] {
  let parts = try Url.parse(url)
  parts.user = parts.user.cut(":")?.before ?? parts.user
  Result.ok(Url.stringify(parts))
}

println(redact("postgres://app:hunter2@db.internal:5432/orders"))
println(redact("https://example.com/"))
```

```text
Ok{ value: "postgres://app@db.internal:5432/orders" }
Ok{ value: "https://example.com/" }
```

Read the query's keys and values with [Url.parseQuery](../parseQuery.md):

```ant
let parts = try Url.parse("https://shop.example.com/list?tag=new&tag=sale&q=red%20shoes")
let params = try Url.parseQuery(parts.query)
println(params["tag"])
println(params["q"])
```

```text
["new", "sale"]
["red shoes"]
```

Add query parameters to a base URL, with [Url.stringifyQuery](../stringifyQuery.md) building the
`query`:

```ant
func withQuery(url: string, params: Map[string, Array[string]]): Result[string, Error] {
  let parts = try Url.parse(url)
  Result.ok(Url.stringify(Url.Parts{ ...parts, query: Url.stringifyQuery(params) }))
}

println(withQuery("https://example.com/search", Map{ "q": ["red shoes"], "page": ["2"] }))
```

```text
Ok{ value: "https://example.com/search?q=red+shoes&page=2" }
```

Tell links that leave a site from those that stay:

```ant
func external(link: string, site: string): bool {
  let parts = Url.parse(link)
  if parts is Err { return false }
  let { host } = parts.value
  !host.isEmpty() && host != site
}

let links = ["/about", "https://example.com/blog", "https://other.org/", "#top", "mailto:hi@example.com"]
println(links | Array.filter({ external(x, "example.com") }))
```

```text
["https://other.org/"]
```

## Properties and relations

[Url.stringify](../stringify.md) writes the parts as a reference that parses back to the same parts:

```ant
let urls = [
  "https://ann@example.com:8080/a%2Fb/my%20docs?q=a%20b#top",
  "mailto:alice@example.com",
  "file:///etc/hosts",
  "../img/logo.png",
  "//cdn.example.com/lib.js",
]
for url in urls {
  let parts = try Url.parse(url)
  assert(try Url.parse(Url.stringify(parts)) == parts)
}
```

Two `Url.Parts` are equal when every field is equal, so references that differ only in the case of
the scheme parse to equal parts:

```ant
assert(try Url.parse("HTTP://example.com/") == try Url.parse("http://example.com/"))
```

## Possible issues

`query` is not decoded, unlike `path` and `fragment`. Decode its keys and values with
[Url.parseQuery](../parseQuery.md):

```ant
let parts = try Url.parse("https://example.com/search?q=red%20shoes")
println(parts.query)
println((try Url.parseQuery(parts.query))["q"])
```

```text
q=red%20shoes
["red shoes"]
```

[Url.stringify](../stringify.md) writes `query` as it is, without escaping, so a space set there
stays a bare space. Build `query` with [Url.stringifyQuery](../stringifyQuery.md):

```ant
let base = try Url.parse("https://example.com/search")
println(Url.stringify(Url.Parts{ ...base, query: "q=red shoes" }))
let query = Url.stringifyQuery(Map{ "q": ["red shoes"], "page": ["2"] })
println(Url.stringify(Url.Parts{ ...base, query }))
```

```text
https://example.com/search?q=red shoes
https://example.com/search?q=red+shoes&page=2
```

A literal must give every field, even the empty ones:

```ant
let parts = Url.Parts{ scheme: "https", opaque: "", user: "", host: "example.com", path: "/" }
// error: Url.Parts literal is missing the field port
// error: Url.Parts literal is missing the field query
// error: Url.Parts literal is missing the field fragment
```

`port` is a string, and `""` when the reference has none, not the scheme's default port:

```ant
let parts = try Url.parse("https://example.com/")
println(parts.port.isEmpty())
println(parts.port.toInt())
```

```text
true
null
```

A reference that is not well formed gives no `Url.Parts`, but an `Err`:

```ant
println(Url.parse("http://my host/"))
println(Url.parse("/100%"))
```

```text
Err{ error: Error{ message: "parse \"http://my host/\": invalid character \" \" in host name" } }
Err{ error: Error{ message: "parse \"/100%\": invalid URL escape \"%\"" } }
```

## See also

[Url.parse](../parse.md) · [Url.stringify](../stringify.md) · [Url.parseQuery](../parseQuery.md) ·
[Url.stringifyQuery](../stringifyQuery.md) · [Url.escape](../escape.md) ·
[Url.unescape](../unescape.md)
