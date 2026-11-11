<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.cut

```ant signature
func String.cut(s: string, sep: string): String.Cut | null
```

String.cut splits s around the first occurrence of sep, or returns null when sep does not occur. It
panics if sep is empty.

<!-- end generated -->

## Details

- `sep` is in neither field; any later occurrence of `sep` stays in `after`.

## Basic examples

Split a key from its value at the first `=`. The rest of the string, later `=` included, goes to
`after`:

```ant
String.cut("k=v=w", "=") // String.Cut{ before: "k", after: "v=w" }
```

When the separator does not occur, the result is `null`:

```ant
String.cut("key", "=") // null
```

Read one side with `?.`:

```ant
String.cut("host:8080", ":")?.after // "8080"
String.cut("localhost", ":")?.after // null
```

## Scope

A separator at either end leaves an empty side:

```ant
"=v".cut("=") // String.Cut{ before: "", after: "v" }
"k=".cut("=") // String.Cut{ before: "k", after: "" }
"=".cut("=")  // String.Cut{ before: "", after: "" }
```

The empty string contains no separator:

```ant
"".cut("=") // null
```

A separator can be longer than one character:

```ant
"a::b::c".cut("::")                 // String.Cut{ before: "a", after: "b::c" }
"Content-Type: text/html".cut(": ") // String.Cut{ before: "Content-Type", after: "text/html" }
```

The whole string as the separator leaves both sides empty:

```ant
"abc".cut("abc") // String.Cut{ before: "", after: "" }
```

Unicode text and separators work like any other:

```ant
"🐜🐝🐞".cut("🐝")    // String.Cut{ before: "🐜", after: "🐞" }
"naïve café".cut(" ") // String.Cut{ before: "naïve", after: "café" }
```

The match is exact: case matters, and a precomposed `é` does not contain the code point `e`:

```ant
"Key=V".cut("k") // null
"é".cut("e")     // null
```

`String.cut` matches code points, not graphemes, so a separator can split a letter from its
combining accent:

```ant
"e\u{301}x".cut("\u{301}") // String.Cut{ before: "e", after: "x" }
```

Call it through `.`, in a pipe, or with named arguments:

```ant
"k=v".cut("=")                 // String.Cut{ before: "k", after: "v" }
"k=v" | String.cut("=")        // String.Cut{ before: "k", after: "v" }
String.cut(s: "a-b", sep: "-") // String.Cut{ before: "a", after: "b" }
```

## Applications

Parse `key=value` lines into a map, skipping lines without `=` with `?? continue`:

```ant
let text = "name=ant\n# a comment\nversion=0.4\nurl=https://x.org/?a=1\n"
let config: Map[string, string] = Map{}
for line in text.lines() {
  let kv = line.cut("=") ?? continue
  config[kv.before] = kv.after
}
println(config)
```

```text
Map{ "name": "ant", "version": "0.4", "url": "https://x.org/?a=1" }
```

Narrow the result with `!= null` and fall back when the separator is missing:

```ant
func port(address: string): string {
  let c = address.cut(":")
  if c != null { c.after } else { "80" }
}

println(port("example.org:8080"))
println(port("example.org"))
```

```text
8080
80
```

Return early from a function with `?? return`, then destructure the record:

```ant
type Email = { user: string, domain: string }

func parseEmail(s: string): Email | null {
  let { before, after } = s.cut("@") ?? return null
  Email{ user: before, domain: after }
}

println(parseEmail("ann@example.org"))
println(parseEmail("ann"))
```

```text
Email{ user: "ann", domain: "example.org" }
null
```

Cut in a loop to take a string apart one field at a time:

```ant
var rest = "a/b/c"
let parts: Array[string] = []
while true {
  let c = rest.cut("/")
  if c == null {
    parts.push(rest)
    break
  }
  parts.push(c.before)
  rest = c.after
}
println(parts)
```

```text
["a", "b", "c"]
```

## Properties and relations

When the result is not `null`, joining `before`, `sep`, and `after` gives `s` back, and `before`
does not contain `sep`:

```ant
let s = "user=ann=admin"
let c = s.cut("=")
assert(c != null)
assert(c.before + "=" + c.after == s)
assert(!c.before.contains("="))
```

The result is `null` exactly when `String.contains` is `false`:

```ant
for s in ["k=v", "kv", ""] {
  assert((s.cut("=") == null) == !s.contains("="))
}
```

The length of `before` is the index `String.indexOf` gives, and `before` is the first piece
`String.split` gives:

```ant
let s = "a,b,c"
let c = s.cut(",")
assert(c != null)
assert(c.before.length() == s.indexOf(","))
assert(c.before == s.split(",")[0])
```

## Possible issues

An empty separator panics:

```ant
"abc".cut("") // panic: String.cut needs a separator that is not empty
```

The result may be `null`, so reading a field of it without narrowing is a compile error. The same
holds for a destructuring:

```ant
let c = "a=b".cut("=")
println(c.before) // error: String.Cut | null has no field before: narrow the value first
```

```ant
let { before, after } = "a=b".cut("=")
// error: String.Cut | null has no field before: narrow the value first
// error: String.Cut | null has no field after: narrow the value first
println(before)
```

Use `?.`, `??`, or a test against `null` first, as in the examples above.

To treat a missing separator as an empty `after`, fall back to a `String.Cut` of your own:

```ant
let s = "key"
s.cut("=") ?? String.Cut{ before: s, after: "" } // String.Cut{ before: "key", after: "" }
```

`String.cut` leaves the white space around the separator in place, so `key = value` text keeps the
spaces on both sides:

```ant
"key = value".cut("=") // String.Cut{ before: "key ", after: " value" }
```

Trim each side with [`String.trim`](trim.md):

```ant
("key = value".cut("=")?.after ?? "").trim() // "value"
```

`String.cut` splits at the first occurrence. To split at the last one, such as the extension of a
file name, use [`String.lastIndexOf`](lastIndexOf.md) and [`String.slice`](slice.md):

```ant
let name = "archive.tar.gz"
name.cut(".")?.after // "tar.gz"
let i = name.lastIndexOf(".")
if i != null { name.slice(i + 1, name.length()) } else { "" } // "gz"
```

## See also

[String.Cut](types/Cut.md) · [String.split](split.md) · [String.indexOf](indexOf.md) ·
[String.lastIndexOf](lastIndexOf.md) · [String.contains](contains.md) · [String.trim](trim.md) ·
[String.trimPrefix](trimPrefix.md) · [String.lines](lines.md) ·
[?. operator](../../language/operators/optional-chaining.md) ·
[?? operator](../../language/operators/null-coalescing.md)
