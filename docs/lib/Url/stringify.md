<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Url.stringify

```ant signature
func Url.stringify(parts: Url.Parts): string
```

Url.stringify returns the reference parts make, escaping what needs it in every part but query and
opaque, which are written as they are, so a program changes a URL by parsing it, replacing a field,
and stringifying it. It writes // before an empty authority only for http, https, ws, wss, ftp and
file, or when path starts with //, so x:///a comes back as x:/a, the same reference.

<!-- end generated -->

## Details

- `query` and `opaque` are written as they are, so build a query with
  [Url.stringifyQuery](stringifyQuery.md) or [Url.escape](escape.md) first. Every other part is
  escaped where it needs it.
- In `path`, `%2F` and `%25` are written as they are and any other `%` is escaped, so a path from
  [Url.parse](parse.md) comes back with the same meaning.
- An empty authority gets `//` only for `http`, `https`, `ws`, `wss`, `ftp` and `file`, in any case,
  or when `path` starts with `//`, with or without a scheme: a lone `path` of `//a` is written
  `////a`.
- Where the parts do not fit together, the reference stays readable: a `path` after a host gets a
  leading `/`, a relative `path` whose first segment holds a `:` gets `./`, and an `opaque` part
  replaces `user`, `host`, `port` and `path`.

## Basic examples

Turn the parts of a URL back into a string:

```ant
let parts = Url.Parts{
  scheme: "https",
  opaque: "",
  user: "",
  host: "ant-lang.org",
  port: "",
  path: "/docs",
  query: "q=url",
  fragment: "top",
}
Url.stringify(parts) // "https://ant-lang.org/docs?q=url#top"
```

Parse a URL, replace one field with a spread, and write it back:

```ant
let home = Url.parse("https://ant-lang.org/docs?q=url#top") | Result.unwrap()
Url.stringify(Url.Parts{ ...home, path: "/play" }) // "https://ant-lang.org/play?q=url#top"
```

## Scope

Every part that is not empty appears, `user` and `port` included:

```ant
let e = Url.Parts{ scheme: "", opaque: "", user: "", host: "", port: "", path: "", query: "", fragment: "" }
let ftp = Url.Parts{ ...e, scheme: "ftp", user: "ann", host: "files.example.com", port: "2121", path: "/pub" }
ftp.stringify() // "ftp://ann@files.example.com:2121/pub"
```

A relative reference has no scheme or host, and all parts empty make the empty reference:

<!-- test: before
let e = Url.Parts{ scheme: "", opaque: "", user: "", host: "", port: "", path: "", query: "", fragment: "" }
-->

```ant
Url.stringify(Url.Parts{ ...e, path: "../img/logo.png" })         // "../img/logo.png"
Url.stringify(Url.Parts{ ...e, path: "/search", query: "q=ant" }) // "/search?q=ant"
Url.stringify(Url.Parts{ ...e, fragment: "intro" })               // "#intro"
Url.stringify(e)                                                  // ""
```

The path and fragment are escaped, Unicode as its UTF-8 bytes, and so is a `?` or `#` that would
otherwise end the path:

<!-- test: before
let e = Url.Parts{ scheme: "", opaque: "", user: "", host: "", port: "", path: "", query: "", fragment: "" }
-->

```ant
Url.stringify(Url.Parts{ ...e, path: "/a b/ü" })    // "/a%20b/%C3%BC"
Url.stringify(Url.Parts{ ...e, path: "/a?b#c" })    // "/a%3Fb%23c"
Url.stringify(Url.Parts{ ...e, fragment: "sec 1" }) // "#sec%201"
```

The `user` is escaped too, so a `@` in a password does not end the userinfo:

```ant
let site = Url.parse("https://example.com/") | Result.unwrap()
Url.stringify(Url.Parts{ ...site, user: "bob:p@ss" }) // "https://bob:p%40ss@example.com/"
```

A path keeps `%2F` and `%25` as written, and escapes any other `%`:

```ant
let doc = Url.parse("https://example.com/") | Result.unwrap()
Url.stringify(Url.Parts{ ...doc, path: "/a%2Fb" }) // "https://example.com/a%2Fb"
Url.stringify(Url.Parts{ ...doc, path: "/50%" })   // "https://example.com/50%25"
```

An opaque reference, such as a `mailto:` address, is `scheme:` followed by `opaque`:

```ant
let mail = Url.parse("mailto:alice@example.com") | Result.unwrap()
mail.opaque                                                    // "alice@example.com"
Url.stringify(Url.Parts{ ...mail, opaque: "bob@example.com" }) // "mailto:bob@example.com"
```

An empty authority keeps its `//` for the schemes whose references always have one, and loses it for
the others:

```ant
Url.parse("file:///etc/hosts") | Result.unwrap() | Url.stringify() // "file:///etc/hosts"
Url.parse("x:///a") | Result.unwrap() | Url.stringify()            // "x:/a"
Url.parse("x:/a") | Result.unwrap() | Url.stringify()              // "x:/a"
```

A `path` that starts with `//` gets `//` before the empty authority for any scheme, so the path is
not read as an authority:

<!-- test: before
let e = Url.Parts{ scheme: "", opaque: "", user: "", host: "", port: "", path: "", query: "", fragment: "" }
-->

```ant
Url.stringify(Url.Parts{ ...e, scheme: "x", path: "//a/b" }) // "x:////a/b"
```

Call it through `.` on a `Url.Parts` value, or at the end of a pipe:

```ant
let page = Url.parse("https://ant-lang.org/docs") | Result.unwrap()
page.stringify()       // "https://ant-lang.org/docs"
page | Url.stringify() // "https://ant-lang.org/docs"
```

## Applications

Set the query of a URL, escaping each value with `Url.stringifyQuery`; any query it had is replaced:

```ant
func withQuery(url: string, params: Map[string, Array[string]]): Result[string, Error] {
  let parts = try Url.parse(url)
  let withParams = Url.Parts{ ...parts, query: Url.stringifyQuery(params) }
  Result.ok(withParams.stringify())
}

println(withQuery("https://example.com/search", Map{ "q": ["a b&c"], "page": ["2"] }) | Result.unwrap())
```

```text
https://example.com/search?q=a+b%26c&page=2
```

Strip the password and the fragment from a URL before it goes into a log:

```ant
func redact(url: string): string {
  let parts = Url.parse(url)
  if parts is Err {
    return "<bad url>"
  }
  let user = parts.value.user.split(":")[0]
  let safe = Url.Parts{ ...parts.value, user, fragment: "" }
  safe.stringify()
}

println(redact("https://ann:secret@db.example.com:5432/app#notes"))
println(redact("http://exa mple.com"))
```

```text
https://ann@db.example.com:5432/app
<bad url>
```

Move every link to `https` and drop a default port:

```ant
func upgrade(url: string): Result[string, Error] {
  let parts = try (Url.parse(url) | Result.wrapErr("upgrade ${url}"))
  if parts.scheme != "http" {
    return Result.ok(url)
  }
  let port = if parts.port == "80" { "" } else { parts.port }
  let secure = Url.Parts{ ...parts, scheme: "https", port }
  Result.ok(secure.stringify())
}

for url in ["http://example.com:80/a", "http://example.com:8080/b", "mailto:ann@example.com"] {
  println(upgrade(url) | Result.unwrap())
}
```

```text
https://example.com/a
https://example.com:8080/b
mailto:ann@example.com
```

## Properties and relations

A URL that `Url.parse` reads comes back from `Url.stringify` with the same parts:

```ant
let urls = [
  "https://ann@example.com:8443/a%20b/c%2Fd?x=%20&y#f%20g",
  "mailto:alice@example.com",
  "../img/logo.png",
  "file:///etc/hosts",
]
for url in urls {
  let parts = Url.parse(url) | Result.unwrap()
  assert((Url.parse(parts.stringify()) | Result.unwrap()) == parts)
}
```

For these URLs, already written the way `Url.stringify` writes them, the text itself comes back:

```ant
for url in ["https://ann@example.com:8443/a%20b/c%2Fd?x=%20#f%20g", "x:/a", "http://h/p"] {
  assert((Url.parse(url) | Result.unwrap() | Url.stringify()) == url)
}
```

## Possible issues

`query` is written as it is, so a value put in it without escaping can break the URL. Build it with
`Url.stringifyQuery`:

```ant
let base = Url.parse("https://example.com/search") | Result.unwrap()
Url.stringify(Url.Parts{ ...base, query: "q=a b&c" })                                 // "https://example.com/search?q=a b&c"
Url.stringify(Url.Parts{ ...base, query: Url.stringifyQuery(Map{ "q": ["a b&c"] }) }) // "https://example.com/search?q=a+b%26c"
```

The text can change on the way through even when the reference does not: an empty `?` or `#` is
dropped, the scheme is in lower case after parsing, and `x:///a` loses its empty authority:

```ant
Url.parse("https://example.com/p?") | Result.unwrap() | Url.stringify() // "https://example.com/p"
Url.parse("HTTPS://example.com/") | Result.unwrap() | Url.stringify()   // "https://example.com/"
Url.parse("x:///a") | Result.unwrap() | Url.stringify()                 // "x:/a"
```

`Url.Parts` has no field defaults, so a literal must give all eight fields, even the empty ones.
Start from a parsed URL, or from one all-empty value, and spread it:

```ant
let parts = Url.Parts{ scheme: "https", opaque: "", user: "", host: "example.com", port: "", path: "/" }
// error: Url.Parts literal is missing the field query
// error: Url.Parts literal is missing the field fragment
```

## See also

[Url.parse](parse.md), [Url.Parts](types/Parts.md), [Url.stringifyQuery](stringifyQuery.md),
[Url.escape](escape.md), [Url.parseQuery](parseQuery.md), [Url](index.md)
