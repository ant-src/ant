<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Url.parseQuery

```ant signature
func Url.parseQuery(query: string): Result[Map[string, Array[string]], Error]
```

Url.parseQuery parses query, k=v pairs joined by &, decoding each key and value as
[`Url.unescape`](unescape.md) does. Keys are in order of first appearance, each with its values in
order, and a key without = has the value "". It is Err where [`Url.unescape`](unescape.md) is.

<!-- end generated -->

## Details

- The query is the text after `?`, without it. A leading `?` or a whole URL is read as part of the
  first key; take the query from [`Url.parse`](parse.md) first.
- Only `&` separates pairs. A `+` decodes to a space, so a literal plus arrives as `%2B`.
- The result is `Err` on a malformed escape such as `%zz`, and on escapes that do not decode to
  UTF-8 text, such as `%FF`.

## Basic examples

Parse a query into a map from each key to its values:

```ant
Url.parseQuery("q=ant&page=2") // Ok{ value: Map{ "q": ["ant"], "page": ["2"] } }
```

A repeated key keeps every value, in order, and a key without `=` has the value `""`:

```ant
Url.parseQuery("tag=a&tag=b&q") // Ok{ value: Map{ "tag": ["a", "b"], "q": [""] } }
```

## Scope

An empty query gives an empty map:

```ant
Url.parseQuery("") // Ok{ value: Map{} }
```

Keys and values are decoded: `+` is a space and `%XX` escapes are UTF-8 bytes:

```ant
Url.parseQuery("q=hello+world&name=J%C3%BCrgen") // Ok{ value: Map{ "q": ["hello world"], "name": ["Jürgen"] } }
Url.parseQuery("first%20name=Ann")               // Ok{ value: Map{ "first name": ["Ann"] } }
```

Text that needs no escape passes through, emoji included:

```ant
Url.parseQuery("q=🐜") // Ok{ value: Map{ "q": ["🐜"] } }
```

A key with `=` and nothing after it also has the value `""`:

```ant
Url.parseQuery("a=&b") // Ok{ value: Map{ "a": [""], "b": [""] } }
```

Keys come in order of first appearance, and a key that repeats later adds to its first entry:

```ant
let params = Url.parseQuery("b=1&a=2&b=3").unwrap()
params.keys() // ["b", "a"]
params["b"]   // ["1", "3"]
```

A bad escape is `Err`:

```ant
Url.parseQuery("a=%zz") // Err{ error: Error{ message: "invalid URL escape \"%zz\"" } }
Url.parseQuery("a=%FF") // Err{ error: Error{ message: "the %XX escapes do not decode to UTF-8 text" } }
```

The query of a parsed URL is raw, ready for `parseQuery`:

```ant
let url = Url.parse("https://example.com/search?q=ant+lang&lang=en").unwrap()
url.query                 // "q=ant+lang&lang=en"
Url.parseQuery(url.query) // Ok{ value: Map{ "q": ["ant lang"], "lang": ["en"] } }
```

## Applications

Read a parameter in a handler; the request's `query` is raw, ready for `parseQuery`:

```ant
func greet(req: Http.Request): Http.Reply {
  let params = Url.parseQuery(req.query).unwrapOr(Map{})
  let name = params["name"]?.first() ?? "stranger"
  Http.Reply{ status: 200, headers: Map{}, body: "hello, ${name}" }
}

let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
let stop = Context.new()
go { Http.serve(listener, greet, stop) }
println((try Http.get("${base}/?name=Ann")).text())
println((try Http.get("${base}/")).text())
stop.cancel()
```

```text
hello, Ann
hello, stranger
```

Read one parameter with a default. Indexing the map gives `null` for a missing key:

```ant
func pageNumber(query: string): Result[int, Error] {
  let params = try Url.parseQuery(query).wrapErr("query")
  let page = params["page"]?.first() ?? "1"
  let n = page.toInt() ?? return Result.fail("page ${page} is not a number")
  Result.ok(n)
}

println(pageNumber("q=ant&page=3"))
println(pageNumber("q=ant"))
println(pageNumber("page=two"))
println(pageNumber("page=%zz"))
```

```text
Ok{ value: 3 }
Ok{ value: 1 }
Err{ error: Error{ message: "page two is not a number" } }
Err{ error: Error{ message: "query: invalid URL escape \"%zz\"" } }
```

Collect a repeated filter, such as checkboxes in a search form:

```ant
let params = Url.parseQuery("color=red&size=m&color=blue").unwrap()
let colors = params["color"] ?? []
println(colors | String.join(", "))
```

```text
red, blue
```

Strip tracking parameters from a link and write it back with
[`Url.stringifyQuery`](stringifyQuery.md):

```ant
func clean(link: string): Result[string, Error] {
  let url = try Url.parse(link)
  let params = try Url.parseQuery(url.query)
  url.query = params
    | Map.filter({ !x.key.startsWith("utm_") })
    | Url.stringifyQuery()
  Result.ok(Url.stringify(url))
}

println(clean("https://example.com/post?id=7&utm_source=mail&utm_medium=email").unwrap())
```

```text
https://example.com/post?id=7
```

## Properties and relations

[`Url.stringifyQuery`](stringifyQuery.md) writes the map back so that `parseQuery` reads the same
map, keys and values in the same order:

```ant
let params = Map{ "q": ["a b & c"], "tag": ["x", "y"], "empty": [""] }
let query = Url.stringifyQuery(params)
assert(Url.parseQuery(query).unwrap() == params)
```

Each key and value decodes as [`Url.unescape`](unescape.md) decodes it:

```ant
let raw = "caf%C3%A9+au+lait"
assert(Url.parseQuery("q=${raw}").unwrap()["q"] == [Url.unescape(raw).unwrap()])
```

## Possible issues

A leading `?` or a whole URL stays in the first key. Pass only the query:

```ant
Url.parseQuery("?q=ant")                     // Ok{ value: Map{ "?q": ["ant"] } }
Url.parseQuery("https://example.com/?q=ant") // Ok{ value: Map{ "https://example.com/?q": ["ant"] } }
```

A `;` is not a separator; it stays in the value:

```ant
Url.parseQuery("a=1;b=2") // Ok{ value: Map{ "a": ["1;b=2"] } }
```

A literal `+` reads as a space. A sender that means a plus writes `%2B`, as
[`Url.escape`](escape.md) does:

```ant
Url.parseQuery("expr=1+2")   // Ok{ value: Map{ "expr": ["1 2"] } }
Url.parseQuery("expr=1%2B2") // Ok{ value: Map{ "expr": ["1+2"] } }
```

A value is an array, even for a key that appears once. Comparing it with a string does not compile:

```ant
let params = Url.parseQuery("q=ant").unwrap()
let isAnt = params["q"] == "ant"
// error: Array[string] | null and string cannot be compared with ==
```

## See also

[Url.stringifyQuery](stringifyQuery.md) · [Url.parse](parse.md) · [Url.unescape](unescape.md) ·
[Url.escape](escape.md) · [Url.Parts](types/Parts.md) · [Http.Request](../Http/types/Request.md) ·
[Map.get](../Map/get.md)
