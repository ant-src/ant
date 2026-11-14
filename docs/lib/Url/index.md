<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Url

Url divides URLs into parts and escapes and parses query strings.

<!-- end generated -->

## Details

- A URL goes through Url in two layers: [Url.parse](parse.md) and [Url.stringify](stringify.md) turn
  a whole reference into a [Url.Parts](types/Parts.md) record and back, and
  [Url.parseQuery](parseQuery.md) and [Url.stringifyQuery](stringifyQuery.md) turn its `query` into
  a `Map[string, Array[string]]` and back.
- `Url.parse` decodes `path`, except `%2F` and `%25`, which stay as written, and decodes `fragment`.
  It leaves `query` raw, so read the query with `Url.parseQuery`. `Url.stringify` writes `query` as
  it is, so build it with `Url.stringifyQuery`.
- A query key can repeat, so each key maps to an array of values, in order. A one-value query is
  `Map{ "q": ["ant"] }`.
- [Url.escape](escape.md) and [Url.unescape](unescape.md) are for one query key or value: a space is
  `+`, and every character but the unreserved ones, `/` included, becomes `%XX`.
- The functions that read text return a `Result`: `Err` on a bad `%XX`, on escapes that do not
  decode to UTF-8, and on a malformed reference.

## Basic examples

Divide a URL into its parts with `Url.parse`:

```ant
let parts = Url.parse("https://example.com/search?q=ant+lang#results").unwrap()
parts.host     // "example.com"
parts.path     // "/search"
parts.query    // "q=ant+lang"
parts.fragment // "results"
```

Read and write a query with `Url.parseQuery` and `Url.stringifyQuery`:

```ant
Url.parseQuery("q=ant+lang&tag=a&tag=b")                    // Ok{ value: Map{ "q": ["ant lang"], "tag": ["a", "b"] } }
Url.stringifyQuery(Map{ "q": ["ant lang"], "page": ["2"] }) // "q=ant+lang&page=2"
```

Escape one value with `Url.escape`, and decode it with `Url.unescape`:

```ant
Url.escape("tea & cakes")     // "tea+%26+cakes"
Url.unescape("tea+%26+cakes") // Ok{ value: "tea & cakes" }
```

Write parts back with `Url.stringify`:

```ant
let parts = Url.parse("https://example.com/search?q=ant").unwrap()
Url.stringify(Url.Parts{ ...parts, host: "ant-lang.org" }) // "https://ant-lang.org/search?q=ant"
```

## Scope

Absolute and relative references parse alike; the parts a reference does not have are `""`:

```ant
Url.parse("/docs?page=2").unwrap().host               // ""
Url.parse("/docs?page=2").unwrap().query              // "page=2"
Url.parse("mailto:alice@example.com").unwrap().opaque // "alice@example.com"
```

`Url.escape` escapes Unicode as its UTF-8 bytes, and `/`, `?`, and `=` too:

```ant
Url.escape("café")    // "caf%C3%A9"
Url.escape("a/b?c=d") // "a%2Fb%3Fc%3Dd"
Url.escape("~-_.")    // "~-_."
```

An empty query is an empty map, and an empty map is an empty query:

```ant
Url.parseQuery("")        // Ok{ value: Map{} }
Url.stringifyQuery(Map{}) // ""
```

A query key without `=` has the value `""`:

```ant
Url.parseQuery("debug&page=2") // Ok{ value: Map{ "debug": [""], "page": ["2"] } }
```

Text that does not decode is an `Err`:

```ant
Url.unescape("%zz")     // Err{ error: Error{ message: "invalid URL escape \"%zz\"" } }
Url.parseQuery("q=%FF") // Err{ error: Error{ message: "the %XX escapes do not decode to UTF-8 text" } }
```

## Functions by job

### Whole references

- [Url.parse](parse.md): a reference divided into its parts, `Err` when it is malformed
- [Url.stringify](stringify.md): the reference a set of parts makes
- [Url.Parts](types/Parts.md): the record of parts, from `scheme` to `fragment`

### Query strings

- [Url.parseQuery](parseQuery.md): `k=v&k=v` as a map from each key to its values
- [Url.stringifyQuery](stringifyQuery.md): a map of keys and values written as a query

### Escaping

- [Url.escape](escape.md): one query key or value with `%XX` escapes and a space as `+`
- [Url.unescape](unescape.md): the reverse, `Err` on a bad escape

## Applications

Build a search link from user input, escaping the query with `Url.stringifyQuery`:

```ant
func searchLink(term: string, page: int): string {
  let query = Url.stringifyQuery(Map{ "q": [term], "page": ["${page}"] })
  "https://example.com/search?${query}"
}

searchLink("ant & bee", 2) // "https://example.com/search?q=ant+%26+bee&page=2"
```

Add a parameter to a link and keep the rest of it:

```ant
func withParam(link: string, key: string, value: string): Result[string, Error] {
  let parts = try (Url.parse(link) | Result.wrapErr("link"))
  let params = try (Url.parseQuery(parts.query) | Result.wrapErr("query"))
  params[key] = [value]
  Result.ok(Url.stringify(Url.Parts{ ...parts, query: Url.stringifyQuery(params) }))
}

println(withParam("https://example.com/a?b=1#top", "utm_source", "docs"))
println(withParam("https://example.com/%zz", "utm_source", "docs"))
```

```text
Ok{ value: "https://example.com/a?b=1&utm_source=docs#top" }
Err{ error: Error{ message: "link: parse \"https://example.com/%zz\": invalid URL escape \"%zz\"" } }
```

Read a form body, which uses the same `k=v&k=v` encoding as a query, with `get` and `??` for a key
that may be missing:

```ant
let body = "name=Ann+Lee&topics=go&topics=ant"
let form = try Url.parseQuery(body)
println(form.get("name") ?? [])
println(form.get("topics") ?? [])
```

```text
["Ann Lee"]
["go", "ant"]
```

## Properties and relations

`Url.unescape` reverses `Url.escape`:

```ant
for s in ["", "a b", "tea & cakes", "café/ü", "100%"] {
  assert(Url.unescape(Url.escape(s)).unwrap() == s)
}
```

`Url.parseQuery` reverses `Url.stringifyQuery`, keeping the order of the keys and of each key's
values:

```ant
let params = Map{ "q": ["a b"], "tag": ["x", "y"], "empty": [""] }
assert(Url.parseQuery(Url.stringifyQuery(params)).unwrap() == params)
```

## Possible issues

`Url.stringify` writes `query` as it is, so a raw value with a space or `&` breaks the URL. Build
the query with `Url.stringifyQuery`:

```ant
let parts = Url.parse("https://example.com/search").unwrap()
Url.stringify(Url.Parts{ ...parts, query: "q=a b&c" })                                 // "https://example.com/search?q=a b&c"
Url.stringify(Url.Parts{ ...parts, query: Url.stringifyQuery(Map{ "q": ["a b&c"] }) }) // "https://example.com/search?q=a+b%26c"
```

`Url.escape` is for query keys and values, so it escapes every `/` of a path:

```ant
Url.escape("/docs/intro") // "%2Fdocs%2Fintro"
```

To put text in a path, set `path` and let `Url.stringify` escape it:

```ant
let parts = Url.parse("https://example.com/").unwrap()
Url.stringify(Url.Parts{ ...parts, path: "/docs/a b" }) // "https://example.com/docs/a%20b"
```

`Url.parse` keeps `%2F` and `%25` in `path` as written, while it decodes other escapes such as
`%20`. `/a%2Fb` and `/a/b` are different paths, and `Url.stringify` writes the path back unchanged:

```ant
let parts = Url.parse("https://example.com/a%2Fb%20c").unwrap()
parts.path           // "/a%2Fb c"
Url.stringify(parts) // "https://example.com/a%2Fb%20c"
```

`Url.unescape` reads `+` as a space, as a query does, so it changes a path that holds a `+`:

```ant
Url.unescape("/c++") // Ok{ value: "/c  " }
```

## See also

- [Http](../Http/index.md): requests whose `query` field `Url.parseQuery` reads
- [Encoding](../Encoding/index.md): hex and base64, other ways to write bytes as text
- [String](../String/index.md): splitting and joining text by hand
