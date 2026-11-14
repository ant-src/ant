<!-- status: reviewed -->

# Parse and build URLs and query strings

Read the parts of a URL, look up its query parameters, and write a changed or new URL back.

1. Divide the URL with [Url.parse](../../lib/Url/parse.md). It returns a `Result`, so take the parts
   with [`try`](../../language/keywords/try.md).
2. Read the query with [Url.parseQuery](../../lib/Url/parseQuery.md), which gives a map from each
   key to its values.
3. Change the parts with a spread into a new [Url.Parts](../../lib/Url/types/Parts.md) and write it
   with [Url.stringify](../../lib/Url/stringify.md).
4. Write a query from a map with [Url.stringifyQuery](../../lib/Url/stringifyQuery.md), which
   escapes every key and value.

## Read the parts of a URL

`Url.parse` divides the URL into the fields of a `Url.Parts` record. The path and fragment come back
decoded; the query is the raw text after `?`:

```ant
let url = try Url.parse("https://ann@shop.example.com:8443/search/red%20shoes?q=boots#results")
println(url.scheme)
println(url.user)
println(url.host)
println(url.port)
println(url.path)
println(url.query)
println(url.fragment)
```

```text
https
ann
shop.example.com
8443
/search/red shoes
q=boots
results
```

`port` is text too, so convert it with [String.toInt](../../lib/String/toInt.md) and give a default
with `??` for a URL without one:

```ant
let url = try Url.parse("https://shop.example.com/search")
let port = url.port.toInt() ?? 443
println(port)
```

```text
443
```

A part the URL does not have is `""`. A relative reference, such as a link inside a page, parses the
same way, with an empty scheme and host:

```ant
let link = try Url.parse("/search?q=ant#top")
println("scheme=${link.scheme} host=${link.host} path=${link.path}")
```

```text
scheme= host= path=/search
```

## Handle a malformed URL

`Url.parse` is an `Err` for a URL it cannot read, such as one with a space in the host or a bad
`%XX` escape. Add what the program was doing with [Result.wrapErr](../../lib/Result/wrapErr.md):

```ant
let parsed = Url.parse("http://exa mple.com") | Result.wrapErr("read link")
println(parsed)
```

```text
Err{ error: Error{ message: "read link: parse \"http://exa mple.com\": invalid character \" \" in host name" } }
```

Outside the host, `Url.parse` is lenient, as browsers are: a space in the path is kept as written.

## Read query parameters

Pass the `query` field to `Url.parseQuery`. Each key maps to an array of its values, in order, since
a key may repeat. Indexing the map gives `null` for a missing key, so take one value with
[Array.first](../../lib/Array/first.md) through `?.` and give a default with `??`. The `query` field
of an [Http.Request](../../lib/Http/types/Request.md) that a server receives is the same raw text,
so pass it to `Url.parseQuery` the same way:

```ant
let url = try Url.parse("https://shop.example.com/search?q=red+shoes&size=42&size=43")
let params = try Url.parseQuery(url.query)
let q = params["q"]?.first() ?? ""
let sizes = params["size"] ?? []
let page = params["page"]?.first()?.toInt() ?? 1
println(q)
println(sizes)
println(page)
```

```text
red shoes
["42", "43"]
1
```

`Url.parseQuery` decodes `+` as a space and `%XX` escapes. It is an `Err` for an escape that is not
UTF-8 text, such as `%FF`.

## Build a query string

Make a `Map[string, Array[string]]` and write it with `Url.stringifyQuery`. It keeps the map's
order, writes one pair per value, and escapes every key and value, so a `&` or `=` inside a value
cannot end it:

```ant
let query = Url.stringifyQuery(Map{ "q": ["salt & pepper"], "size": ["42", "43"] })
println(query)
```

```text
q=salt+%26+pepper&size=42&size=43
```

A key with no values writes nothing.

## Change part of a URL

Parse the URL, change the query map with [Map.set](../../lib/Map/set.md), and spread the old parts
into a new `Url.Parts` with the fields you replace. `Url.stringify` writes the result:

```ant
let current = try Url.parse("https://shop.example.com/search?q=boots&page=1#results")
let params = try Url.parseQuery(current.query)
params.set("page", ["2"])
let next = Url.Parts{ ...current, query: Url.stringifyQuery(params), fragment: "" }
println(next.stringify())
```

```text
https://shop.example.com/search?q=boots&page=2
```

`Url.stringify` escapes the path, the fragment, and the user, so set them as plain text:

```ant
let site = try Url.parse("https://shop.example.com/")
let report = Url.Parts{ ...site, path: "/files/my report.pdf" }
println(report.stringify())
```

```text
https://shop.example.com/files/my%20report.pdf
```

It writes `query` as it is, so build that field with `Url.stringifyQuery`.

## Build a URL from nothing

A `Url.Parts` literal names all eight fields. Start from one with every field empty and spread it:

```ant
let empty = Url.Parts{
  scheme: "",
  opaque: "",
  user: "",
  host: "",
  port: "",
  path: "",
  query: "",
  fragment: "",
}
let api = Url.Parts{
  ...empty,
  scheme: "https",
  host: "api.example.com",
  path: "/v1/items",
  query: Url.stringifyQuery(Map{ "q": ["green tea"], "limit": ["10"] }),
}
println(api.stringify())
```

```text
https://api.example.com/v1/items?q=green+tea&limit=10
```

## Escape one value

When you write a query by hand, escape each key and value with
[Url.escape](../../lib/Url/escape.md). [Url.unescape](../../lib/Url/unescape.md) reverses it:

```ant
let term = "salt & pepper"
println("https://shop.example.com/search?q=${Url.escape(term)}")
println(try Url.unescape("salt+%26+pepper"))
```

```text
https://shop.example.com/search?q=salt+%26+pepper
salt & pepper
```

`Url.escape` is for the query only: it writes a space as `+`, which a path reads as a plus sign.
This example escapes a path segment with it on purpose, to show the mistake:

```ant
println("https://shop.example.com/files/${Url.escape("my report.pdf")}")
```

```text
https://shop.example.com/files/my+report.pdf
```

Set the `path` field and call `Url.stringify` instead, as in
[Change part of a URL](#change-part-of-a-url).

## The whole program

This script cleans the links given as arguments: it drops the `utm_` tracking parameters and the
fragment, and reports a link it cannot read and goes on to the next.
[Map.filter](../../lib/Map/filter.md) keeps the parameters, and
[String.startsWith](../../lib/String/startsWith.md) tests each key:

<!-- test: args https://shop.example.com/item?id=7&utm_source=mail&utm_medium=email#reviews http://exa%20mple.com /about?utm_campaign=fall -->

```ant
func clean(link: string): Result[string, Error] {
  let url = try (Url.parse(link) | Result.wrapErr("clean"))
  let params = try (Url.parseQuery(url.query) | Result.wrapErr("clean ${link}"))
  let kept = params.filter({ !x.key.startsWith("utm_") })
  let cleaned = Url.Parts{ ...url, query: Url.stringifyQuery(kept), fragment: "" }
  Result.ok(cleaned.stringify())
}

for link in Os.args() {
  let cleaned = clean(link)
  if cleaned is Ok { println(cleaned.value) } else { println("skip: ${cleaned.error.message}") }
}
```

```text
https://shop.example.com/item?id=7
skip: clean: parse "http://exa%20mple.com": invalid URL escape "%20"
/about
```

It also uses [Os.args](../../lib/Os/args.md).

## See also

[Url.parse](../../lib/Url/parse.md) · [Url.stringify](../../lib/Url/stringify.md) ·
[Url.Parts](../../lib/Url/types/Parts.md) · [Url.parseQuery](../../lib/Url/parseQuery.md) ·
[Url.stringifyQuery](../../lib/Url/stringifyQuery.md) · [Url.escape](../../lib/Url/escape.md) ·
[Url.unescape](../../lib/Url/unescape.md) · [Url](../../lib/Url/index.md) ·
[Result.wrapErr](../../lib/Result/wrapErr.md) · [Map.set](../../lib/Map/set.md) ·
[Map.filter](../../lib/Map/filter.md)
