<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Url.unescape

```ant signature
func Url.unescape(s: string): Result[string, Error]
```

Url.unescape reverses [`Url.escape`](escape.md). It is Err on a bad %XX or on escapes that decode to
bytes that are not UTF-8.

<!-- end generated -->

## Details

- `+` decodes to a space, so a `+` that is data arrives as `%2B`.

## Basic examples

Decode one escaped query value:

```ant
Url.unescape("a+b%26c")   // Ok{ value: "a b&c" }
Url.unescape("caf%C3%A9") // Ok{ value: "café" }
Url.unescape("hello")     // Ok{ value: "hello" }
```

A bad escape is `Err`:

```ant
Url.unescape("100%") // Err{ error: Error{ message: "invalid URL escape \"%\"" } }
```

## Scope

The empty string decodes to itself:

```ant
Url.unescape("") // Ok{ value: "" }
```

`+` and `%20` both decode to a space, and `%2B` to a plus sign:

```ant
Url.unescape("a+b")   // Ok{ value: "a b" }
Url.unescape("a%20b") // Ok{ value: "a b" }
Url.unescape("a%2Bb") // Ok{ value: "a+b" }
```

`%25` is a literal percent sign, so it decodes only once:

```ant
Url.unescape("%2525") // Ok{ value: "%25" }
```

A `%` that is not followed by two hex digits is `Err`:

```ant
Url.unescape("%")   // Err{ error: Error{ message: "invalid URL escape \"%\"" } }
Url.unescape("%4")  // Err{ error: Error{ message: "invalid URL escape \"%4\"" } }
Url.unescape("%zz") // Err{ error: Error{ message: "invalid URL escape \"%zz\"" } }
```

Escapes whose bytes are not UTF-8 are `Err`, such as `%FF`, or `%C3` without the byte that completes
it:

```ant
Url.unescape("%FF")    // Err{ error: Error{ message: "the %XX escapes do not decode to UTF-8 text" } }
Url.unescape("caf%C3") // Err{ error: Error{ message: "the %XX escapes do not decode to UTF-8 text" } }
```

Through a pipe, with the error given context:

```ant
"q%3Dant" | Url.unescape() | Result.wrapErr("search") // Ok{ value: "q=ant" }
"q%3" | Url.unescape() | Result.wrapErr("search")     // Err{ error: Error{ message: "search: invalid URL escape \"%3\"" } }
```

## Applications

Read back one value that was stored with `Url.escape`, such as a name kept in a cookie, and give a
bad value context:

```ant
func greeting(stored: string): Result[string, Error] {
  let name = try (Url.unescape(stored) | Result.wrapErr("saved name"))
  Result.ok("Hello, ${name}!")
}

let stored = Url.escape("Ann Lee & co")
println(stored)
println(greeting(stored))
println(greeting("Ann%2"))
```

```text
Ann+Lee+%26+co
Ok{ value: "Hello, Ann Lee & co!" }
Err{ error: Error{ message: "saved name: invalid URL escape \"%2\"" } }
```

A whole `k=v&k=v` string is a query; read it with [`Url.parseQuery`](parseQuery.md), which splits it
and decodes each key and value, rather than splitting it and unescaping by hand.

Decode every value of a list, failing on the first bad one, with `Result.all`:

```ant
let tags = ["red+wine", "caf%C3%A9", "a%26b"]
tags.map(Url.unescape) | Result.all() // Ok{ value: ["red wine", "café", "a&b"] }
```

Fall back to the raw text when a value a user typed does not decode:

```ant
let raw = "50%off"
Url.unescape(raw) | Result.unwrapOr(raw) // "50%off"
```

## Properties and relations

`Url.unescape` reverses `Url.escape` for every string:

```ant
for s in ["", "a b&c", "a+b", "100%", "café €5", "line\nbreak"] {
  assert(Url.unescape(Url.escape(s)) == Ok{ value: s })
}
```

`Url.parseQuery` decodes each key and value as `Url.unescape` does:

```ant
let query = try Url.parseQuery("q=a+b%26c")
assert(query["q"] == [try Url.unescape("a+b%26c")])
```

## Possible issues

`Url.unescape` is for query keys and values, where `+` is a space. On other text it turns every `+`
into a space too:

```ant
Url.unescape("/c++/notes") // Ok{ value: "/c  /notes" }
```

Text that `Url.parse` or `Url.parseQuery` gives back is already decoded: `path`, `fragment`, and
each key and value of a query. Unescaping it again decodes twice and changes the value. Unescape
only raw text, such as `query` from `Url.parse`, which stays raw:

```ant
let p = try Url.parse("http://h/x#d=50%2525")
println(p.fragment)
println(Url.unescape(p.fragment))
```

```text
d=50%25
Ok{ value: "d=50%" }
```

Unescape a value once. A second call decodes what was an escaped `%`, and fails where the first did
not:

```ant
let once = try Url.unescape("50%25off")
println(once)
println(Url.unescape(once))
```

```text
50%off
Err{ error: Error{ message: "invalid URL escape \"%of\"" } }
```

## See also

[Url.escape](escape.md), [Url.parseQuery](parseQuery.md), [Url.parse](parse.md),
[Url.stringifyQuery](stringifyQuery.md), [Result.all](../Result/all.md),
[Result.wrapErr](../Result/wrapErr.md)
