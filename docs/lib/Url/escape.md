<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Url.escape

```ant signature
func Url.escape(s: string): string
```

Url.escape escapes s for use as one query key or value: every character but the unreserved ones
becomes %XX, in upper-case hex digits, and a space becomes +.

<!-- end generated -->

## Details

- The unreserved characters are the ASCII letters and digits and `-`, `_`, `.`, `~`. Every other
  character becomes one `%XX` per byte of its UTF-8 encoding.
- It escapes `&`, `=`, `+`, `%`, `?`, and `#`, so the result is always one key or value. Escape each
  key and value on its own, never a whole query.
- The `+` for a space belongs to query strings. Do not use the result as a path segment, where a `+`
  stays a `+`.
- [Url.unescape](unescape.md) reverses it, and [Url.stringifyQuery](stringifyQuery.md) escapes every
  key and value of a map the same way.

## Basic examples

Escape a value for a query. A space becomes `+` and `&` becomes `%26`:

```ant
Url.escape("a b&c")        // "a+b%26c"
Url.escape("fish & chips") // "fish+%26+chips"
```

Put the escaped value into a URL:

```ant
let term = "rock & roll"
"https://example.com/search?q=${Url.escape(term)}" // "https://example.com/search?q=rock+%26+roll"
```

## Scope

The unreserved characters come back as they are, and so does the empty string:

```ant
Url.escape("Ant-lang_1.0~beta") // "Ant-lang_1.0~beta"
Url.escape("")                  // ""
```

Every other ASCII character becomes `%XX`, the reserved delimiters and `%` included:

```ant
Url.escape("!*'();:@&=+$,/?#[]%") // "%21%2A%27%28%29%3B%3A%40%26%3D%2B%24%2C%2F%3F%23%5B%5D%25"
Url.escape("100%")                // "100%25"
Url.escape("line\nbreak\t")       // "line%0Abreak%09"
```

A character outside ASCII becomes one `%XX` for each byte of its UTF-8 encoding:

```ant
Url.escape("café") // "caf%C3%A9"
Url.escape("🐜")   // "%F0%9F%90%9C"
```

It escapes the bytes as they are and does not normalize the text first. The same `é`, composed and
decomposed, escapes two ways:

```ant
let word = "café"
Url.escape(word.normalize("NFC")) // "caf%C3%A9"
Url.escape(word.normalize("NFD")) // "cafe%CC%81"
```

It reads well at the end of a pipe:

```ant
"Ann Lee" | Url.escape() // "Ann+Lee"
```

## Applications

Build a search link from text a user typed. Escape the value, and keep the `&` that separates the
pairs:

```ant
func searchUrl(query: string, page: int): string {
  "https://example.com/search?q=${Url.escape(query)}&page=${page}"
}

let url = searchUrl("C++ & Go?", 2)
println(url)
let parts = Result.unwrap(Url.parse(url))
println(Url.parseQuery(parts.query))
```

```text
https://example.com/search?q=C%2B%2B+%26+Go%3F&page=2
Ok{ value: Map{ "q": ["C++ & Go?"], "page": ["2"] } }
```

Write a form body, `application/x-www-form-urlencoded`, from a list of fields:

```ant
type Field = { name: string, value: string }

let fields = [
  Field{ name: "user", value: "ann@example.com" },
  Field{ name: "note", value: "hi = hello" },
]
let body = fields | Array.map({ "${Url.escape(x.name)}=${Url.escape(x.value)}" }) | String.join("&")
println(body)
```

```text
user=ann%40example.com&note=hi+%3D+hello
```

## Properties and relations

[Url.unescape](unescape.md) gives back the original string:

```ant
for s in ["a b&c", "100%", "a+b", "café 🐜", "", "x=1&y=2"] {
  assert(Result.unwrap(Url.unescape(Url.escape(s))) == s)
}
```

The result holds only unreserved characters, `%`, and `+`:

```ant
let allowed = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.~%+"
for s in ["a b&c", "!*'();:@&=+$,/?#[]", "日本語", "tab\there"] {
  assert(s | Url.escape() | String.chars() | Array.every({ allowed.contains(x) }))
}
```

Joining escaped keys and values with `=` and `&` gives what [Url.stringifyQuery](stringifyQuery.md)
writes:

```ant
let params = Map{ "q": ["rock & roll"], "page": ["2"] }
let query = params.entries()
  | Array.map({ "${Url.escape(x.key)}=${Url.escape(x.value[0])}" })
  | String.join("&")
assert(query == Url.stringifyQuery(params))
```

## Possible issues

A `+` in the input is escaped as `%2B`, since a bare `+` reads back as a space:

```ant
Url.escape("a+b")               // "a%2Bb"
Url.unescape("a+b")             // Ok{ value: "a b" }
Url.unescape(Url.escape("a+b")) // Ok{ value: "a+b" }
```

Escaping a whole query escapes its `=` and `&` too, and the result is one value, not two pairs:

```ant
Url.escape("a=1&b=2")                 // "a%3D1%26b%3D2"
Url.parseQuery(Url.escape("a=1&b=2")) // Ok{ value: Map{ "a=1&b=2": [""] } }
```

Escaping twice escapes the `%` and `+` of the first pass, so one unescape does not give the text
back:

```ant
let twice = Url.escape(Url.escape("a b"))
twice               // "a%2Bb"
Url.unescape(twice) // Ok{ value: "a+b" }
```

A path does not read `+` as a space, so a file name escaped with `Url.escape` comes back with a `+`
in it:

```ant
let parts = Result.unwrap(Url.parse("https://example.com/files/${Url.escape("my notes.txt")}"))
parts.path // "/files/my+notes.txt"
```

To put text in a path, set the `path` field and let [Url.stringify](stringify.md) escape it:

```ant
let parts = Result.unwrap(Url.parse("https://example.com/"))
parts.path = "/files/my notes.txt"
Url.stringify(parts) // "https://example.com/files/my%20notes.txt"
```

## See also

[Url.unescape](unescape.md) · [Url.stringifyQuery](stringifyQuery.md) ·
[Url.parseQuery](parseQuery.md) · [Url.stringify](stringify.md) · [Url.parse](parse.md) ·
[Url.Parts](types/Parts.md) · [Regex.escape](../Regex/escape.md)
