<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Url.parse

```ant signature
func Url.parse(s: string): Result[Url.Parts, Error]
```

Url.parse divides s, an absolute or a relative reference, into its parts. It is Err on a malformed
reference, such as a bad %XX, escapes in the path or fragment that are not UTF-8, or a space in the
host.

<!-- end generated -->

## Details

- Every part a reference does not have is `""`, so a relative reference such as `/search?q=ant` or
  `#top` parses too, with an empty `scheme` and `host`.
- `scheme` comes back in lower case.
- `path` and `fragment` are decoded, `%20` read as a space, but `path` keeps `%2F` and `%25` as
  written. `query` is the raw text after `?`; read it with [Url.parseQuery](parseQuery.md).
- `user` is the userinfo as written, a password included.

## Basic examples

Divide a URL into its parts:

```ant
let parts = Url.parse("https://example.com:8080/docs/intro?lang=en#setup").unwrap()
parts.scheme   // "https"
parts.host     // "example.com"
parts.port     // "8080"
parts.path     // "/docs/intro"
parts.query    // "lang=en"
parts.fragment // "setup"
```

The whole result, with the parts a URL does not have as `""`:

```ant
println(Url.parse("https://example.com/search?q=ant"))
```

```text
Ok{ value: Url.Parts{ scheme: "https", opaque: "", user: "", host: "example.com", port: "", path: "/search", query: "q=ant", fragment: "" } }
```

A malformed reference is an `Err`:

```ant
Url.parse("https://example.com/%zz") // Err{ error: Error{ message: "parse \"https://example.com/%zz\": invalid URL escape \"%zz\"" } }
```

## Scope

Relative references: a path, a query alone, a fragment alone, and the empty string:

```ant
Url.parse("../img/logo.png").unwrap().path // "../img/logo.png"
Url.parse("?page=2").unwrap().query        // "page=2"
Url.parse("#top").unwrap().fragment        // "top"
Url.parse("").unwrap().path                // ""
```

A reference that starts with `//` has an authority but no scheme:

```ant
let cdn = Url.parse("//cdn.example.com/lib.js").unwrap()
cdn.scheme // ""
cdn.host   // "cdn.example.com"
cdn.path   // "/lib.js"
```

A user, with a password, stays as written:

```ant
Url.parse("https://alice:s3cret@example.com/").unwrap().user // "alice:s3cret"
Url.parse("ftp://a%20b@files.example.com/").unwrap().user    // "a%20b"
```

When what follows `scheme:` does not start with `/`, it is `opaque`, as in `mailto:` and `urn:`:

```ant
let mail = Url.parse("mailto:alice@example.com").unwrap()
mail.scheme                                      // "mailto"
mail.opaque                                      // "alice@example.com"
mail.path                                        // ""
Url.parse("urn:isbn:0451450523").unwrap().opaque // "isbn:0451450523"
```

A single `/` after the colon starts the path, and the authority is empty:

```ant
let hosts = Url.parse("file:///etc/hosts").unwrap()
hosts.host                      // ""
hosts.path                      // "/etc/hosts"
Url.parse("x:/a").unwrap().path // "/a"
```

The scheme comes back in lower case:

```ant
let loud = Url.parse("HTTPS://Example.COM/").unwrap()
loud.scheme // "https"
```

The path and fragment are decoded, including Unicode, and the query is not:

```ant
let page = Url.parse("https://example.com/caf%C3%A9/a%20b?q=a%20b#see%20also").unwrap()
page.path     // "/café/a b"
page.query    // "q=a%20b"
page.fragment // "see also"
```

`%2F` and `%25` stay escaped in the path, so a `/` inside a segment stays apart from the separators:

```ant
Url.parse("/files/a%2Fb/100%25").unwrap().path // "/files/a%2Fb/100%25"
Url.parse("/files/a/b/100").unwrap().path      // "/files/a/b/100"
```

Through a pipe:

```ant
"https://example.com/a" | Url.parse() | Result.map({ x.host }) // Ok{ value: "example.com" }
```

## Applications

Read one query parameter from a link, with context on the error:

```ant
func param(link: string, key: string): Result[string, Error] {
  let parts = try (Url.parse(link) | Result.wrapErr("link"))
  let params = try (Url.parseQuery(parts.query) | Result.wrapErr("query of ${link}"))
  let values = params.get(key) ?? []
  Result.ok(values.first() ?? "")
}

println(param("https://example.com/search?q=ant+lang&page=2", "q"))
println(param("https://example.com/search?page=2", "q"))
println(param("https://example.com/%zz", "q"))
```

```text
Ok{ value: "ant lang" }
Ok{ value: "" }
Err{ error: Error{ message: "link: parse \"https://example.com/%zz\": invalid URL escape \"%zz\"" } }
```

Change a URL by parsing it, replacing a field with a record spread, and writing it back with
[Url.stringify](stringify.md):

```ant
func nextPage(link: string, page: int): Result[string, Error] {
  let parts = try Url.parse(link)
  let params = try Url.parseQuery(parts.query)
  params["page"] = ["${page}"]
  Result.ok(Url.stringify(Url.Parts{ ...parts, query: Url.stringifyQuery(params) }))
}

println(nextPage("https://example.com/search?q=ant&page=1", 2))
```

```text
Ok{ value: "https://example.com/search?q=ant&page=2" }
```

Accept only absolute `http` and `https` links, rejecting the rest before the main path:

```ant
func webLink(link: string): Result[Url.Parts, Error] {
  let parts = try Url.parse(link)
  if parts.scheme != "http" && parts.scheme != "https" {
    return Result.fail("not a web link: ${link}")
  }
  if parts.host.isEmpty() {
    return Result.fail("no host: ${link}")
  }
  Result.ok(parts)
}

for link in ["https://example.com/", "mailto:alice@example.com", "/about", "http:///x"] {
  println(webLink(link) | Result.map({ x.host }))
}
```

```text
Ok{ value: "example.com" }
Err{ error: Error{ message: "not a web link: mailto:alice@example.com" } }
Err{ error: Error{ message: "not a web link: /about" } }
Err{ error: Error{ message: "no host: http:///x" } }
```

Fill in the port a URL leaves out with an `if` expression:

```ant
func port(parts: Url.Parts): string {
  if !parts.port.isEmpty() {
    return parts.port
  }
  if parts.scheme == "https" { "443" } else { "80" }
}

port(Url.parse("https://example.com/").unwrap())   // "443"
port(Url.parse("http://localhost:8080/").unwrap()) // "8080"
```

Count links by host, skipping those that do not parse:

```ant
let links = [
  "https://example.com/a",
  "https://docs.example.com/b",
  "https://example.com/c",
  "https://example.com/%zz",
]
let hosts = links
  | Array.map(Url.parse)
  | Array.filter(Result.isOk)
  | Array.map({ x.unwrap().host })
  | Array.groupBy({ x })
  | Map.map({ x.value.length() })
println(hosts)
```

```text
Map{ "example.com": 2, "docs.example.com": 1 }
```

## Properties and relations

[Url.stringify](stringify.md) writes the parts back; for this URL, parsing the result gives the same
parts:

```ant
let parts = Url.parse("https://alice@example.com:8080/a%20b/c%2Fd?q=1#top").unwrap()
assert(Url.parse(Url.stringify(parts)).unwrap() == parts)
Url.stringify(parts) // "https://alice@example.com:8080/a%20b/c%2Fd?q=1#top"
```

The query is raw, so [Url.parseQuery](parseQuery.md) reads it into keys and values:

```ant
let parts = Url.parse("https://example.com/?tag=a&tag=b&q=ant+lang").unwrap()
Url.parseQuery(parts.query) // Ok{ value: Map{ "tag": ["a", "b"], "q": ["ant lang"] } }
```

## Possible issues

Without `//`, a host name is a relative path, and `host:port` is read as a scheme and an opaque
part:

```ant
Url.parse("example.com/docs").unwrap().path       // "example.com/docs"
Url.parse("example.com/docs").unwrap().host       // ""
Url.parse("localhost:8080").unwrap().scheme       // "localhost"
Url.parse("localhost:8080").unwrap().opaque       // "8080"
Url.parse("https://localhost:8080").unwrap().port // "8080"
```

The query is not decoded or checked, so a bad escape there parses, and the `Err` comes from
[Url.parseQuery](parseQuery.md):

```ant
let parts = Url.parse("https://example.com/?q=%zz").unwrap()
parts.query                        // "q=%zz"
Url.parseQuery(parts.query).isOk() // false
```

The path is decoded, so a `%20` and a space give the same path, while `%2F` and `/` do not:

```ant
Url.parse("/a%20b").unwrap().path == Url.parse("/a b").unwrap().path // true
Url.parse("/a%2Fb").unwrap().path == Url.parse("/a/b").unwrap().path // false
```

Escapes in the path or fragment that do not decode to UTF-8 are an `Err`, and so is a space in the
host:

```ant
Url.parse("https://example.com/%FF").isOk()  // false
Url.parse("https://example.com/#%FF").isOk() // false
Url.parse("https://exa mple.com/").isOk()    // false
```

## See also

[Url.stringify](stringify.md), [Url.Parts](types/Parts.md), [Url.parseQuery](parseQuery.md),
[Url.unescape](unescape.md), [Url.escape](escape.md), [Result.wrapErr](../Result/wrapErr.md),
[Http.Request](../Http/types/Request.md)
