<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Url.stringifyQuery

```ant signature
func Url.stringifyQuery(params: Map[string, Array[string]]): string
```

Url.stringifyQuery writes params as k=v pairs joined by &, in the map's order, each key and value
escaped as [`Url.escape`](escape.md) does. A key with no values writes nothing, so it round-trips
with [`Url.parseQuery`](parseQuery.md) for maps whose keys each have a value.

<!-- end generated -->

## Details

- Each value of a key becomes its own `k=v` pair, in the order of the key's array; the keys come in
  the map's insertion order.
- An empty map gives `""`, and a key with no values writes nothing, so it does not come back from
  [Url.parseQuery](parseQuery.md).
- The result has no leading `?`.

## Basic examples

Write a query with one value per key:

```ant
Url.stringifyQuery(Map{ "q": ["ant"], "page": ["2"] }) // "q=ant&page=2"
```

A key with several values repeats, and a space becomes `+`:

```ant
Url.stringifyQuery(Map{ "q": ["a b"], "tag": ["x", "y"] }) // "q=a+b&tag=x&tag=y"
```

## Scope

An empty map gives an empty string:

```ant
let none: Map[string, Array[string]] = Map{}
Url.stringifyQuery(none) // ""
```

An empty value still writes its `=`, and so does an empty key:

```ant
Url.stringifyQuery(Map{ "flag": [""], "": ["x"] }) // "flag=&=x"
```

`&`, `=`, `+`, `%`, `/`, and `?` in a key or value are escaped, so they cannot break the query
apart:

```ant
Url.stringifyQuery(Map{ "a&b": ["c=d+e"], "p": ["50%"], "next": ["/x?y"] }) // "a%26b=c%3Dd%2Be&p=50%25&next=%2Fx%3Fy"
```

The unreserved characters `-`, `_`, `.`, and `~` stay as they are:

```ant
Url.stringifyQuery(Map{ "file": ["v1.2_final-draft~"] }) // "file=v1.2_final-draft~"
```

Unicode text is written as the `%XX` escapes of its UTF-8 bytes:

```ant
Url.stringifyQuery(Map{ "city": ["Zürich"], "名": ["日本"] }) // "city=Z%C3%BCrich&%E5%90%8D=%E6%97%A5%E6%9C%AC"
```

The keys come in insertion order, not sorted:

```ant
Url.stringifyQuery(Map{ "z": ["1"], "a": ["2"] }) // "z=1&a=2"
```

## Named arguments

```ant
Url.stringifyQuery(params: Map{ "q": ["ant"] }) // "q=ant"
```

## Applications

### Change one parameter of a URL

Parse the URL and its query, set the key, and write both back. A value that holds `&` comes out
escaped:

```ant
func withParam(url: string, key: string, value: string): Result[string, Error] {
  let parts = try Url.parse(url)
  let params = try Url.parseQuery(parts.query)
  params[key] = [value]
  parts.query = params | Url.stringifyQuery()
  Result.ok(Url.stringify(parts))
}

println(withParam("https://example.com/search?q=ant&page=1", "page", "2").unwrap())
println(withParam("https://example.com/search", "q", "a&b").unwrap())
```

```text
https://example.com/search?q=ant&page=2
https://example.com/search?q=a%26b
```

Replacing the value of `page` keeps its place in the map, so the query keeps its order.

### Build a query from a record

Add optional parameters only when they are set:

```ant
type Search = { text: string, tags: Array[string], limit: int | null }

func searchQuery(s: Search): string {
  let params = Map{ q: [s.text], tag: s.tags }
  if s.limit != null {
    params["limit"] = ["${s.limit}"]
  }
  Url.stringifyQuery(params)
}

println(searchQuery(Search{ text: "maps vs sets", tags: ["lang", "go"], limit: null }))
println(searchQuery(Search{ text: "ant", tags: ["lang"], limit: 10 }))
```

```text
q=maps+vs+sets&tag=lang&tag=go
q=ant&tag=lang&limit=10
```

### A canonical query for signing

A signature over a query needs one fixed order. Sort the entries by key and rebuild the map:

```ant
let params = Map{ "timestamp": ["1700000000"], "action": ["list"], "key": ["abc"] }
params.entries()
  | Array.sortBy({ x.key })
  | Map.fromEntries()
  | Url.stringifyQuery() // "action=list&key=abc&timestamp=1700000000"
```

## Properties and relations

`Url.parseQuery` reads the result back as the same map:

```ant
let params = Map{ "q": ["a b", "ü"], "next": ["/x?y"], "flag": [""] }
assert(Url.parseQuery(Url.stringifyQuery(params)).unwrap() == params)
```

A single pair is the key and the value, each escaped by `Url.escape`, joined by `=`:

```ant
let key = "a&b"
let value = "1 + 1"
assert(Url.stringifyQuery(Map{ (key): [value] }) == "${Url.escape(key)}=${Url.escape(value)}")
```

## Possible issues

A query that repeats a key with other keys between comes back grouped, since the map holds each
key's values together:

```ant
let params = Url.parseQuery("tag=a&q=x&tag=b").unwrap()
Url.stringifyQuery(params) // "tag=a&tag=b&q=x"
```

`Url.stringify` writes `query` as it is, so a query built by joining strings keeps its `&` and
spaces, and a server splits the value at the `&`. Build it with `Url.stringifyQuery`:

```ant
let parts = Url.parse("https://example.com/search").unwrap()
parts.query = "q=salt & pepper"
println(Url.stringify(parts))
parts.query = Url.stringifyQuery(Map{ q: ["salt & pepper"] })
println(Url.stringify(parts))
```

```text
https://example.com/search?q=salt & pepper
https://example.com/search?q=salt+%26+pepper
```

## See also

- [`Url.parseQuery`](parseQuery.md)
- [`Url.escape`](escape.md)
- [`Url.stringify`](stringify.md)
- [`Url.parse`](parse.md)
- [`Url.Parts`](types/Parts.md)
