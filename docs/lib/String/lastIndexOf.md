<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.lastIndexOf

```ant signature
func String.lastIndexOf(s: string, t: string): int | null
```

String.lastIndexOf returns the code point index of the last occurrence of t in s, or null when t
does not occur.

<!-- end generated -->

## Details

- The index counts code points, as [`String.length`](length.md) and [`String.slice`](slice.md) do.
- When `t` does not occur in `s`, the result is `null`, not `-1`.
- `t` matches code point by code point, so `é` written as `e` plus a combining accent does not match
  a precomposed `é`; normalize both with [`String.normalize`](normalize.md) first.

## Basic examples

```ant
String.lastIndexOf("banana", "an") // 3
String.lastIndexOf("banana", "x")  // null
```

Call it through `.` on any string:

```ant
"src/lib/main.ant".lastIndexOf("/") // 7
"src/lib/main.ant".indexOf("/")     // 3
```

## Scope

The last occurrence wins, even when it overlaps an earlier one:

```ant
String.lastIndexOf("aaa", "aa")    // 1
String.lastIndexOf("abcabc", "ab") // 3
```

`t` may be all of `s`, but not longer:

```ant
String.lastIndexOf("ant", "ant")  // 0
String.lastIndexOf("ant", "ants") // null
```

Nothing occurs in `""`:

```ant
String.lastIndexOf("", "a") // null
```

The index counts code points, not bytes. Each of `ï` and `é` is one code point and two bytes:

```ant
String.lastIndexOf("naïve café, café", "café") // 12
String.lastIndexOf("🐜🐝🐜🐞", "🐜")                // 2
```

It counts code points, not what a reader sees as characters. The family emoji is five code points:

```ant
String.lastIndexOf("👨‍👩‍👧x👨‍👩‍👧", "👨‍👩‍👧") // 6
```

In a pipe, the string is the first argument:

```ant
"a/b/c" | String.lastIndexOf("/") // 3
```

A call can name both arguments:

```ant
String.lastIndexOf(s: "a.b.c", t: ".") // 3
```

## Applications

Take the extension of a file name, the text after the last dot, and return early with `?? return`
when there is none:

```ant
func extension(name: string): string {
  var i = String.lastIndexOf(name, ".") ?? return ""
  String.slice(name, i + 1, String.length(name))
}

println(extension("archive.tar.gz"))
println(extension("README"))
```

```text
gz

```

Split a path into its directory and its base name at the last slash:

```ant
type Split = { dir: string, base: string }

func splitPath(path: string): Split {
  var i = String.lastIndexOf(path, "/")
  if i != null {
    Split{ dir: String.slice(path, 0, i), base: String.slice(path, i + 1, String.length(path)) }
  } else {
    Split{ dir: "", base: path }
  }
}

println(splitPath("docs/lib/String/lastIndexOf.md"))
println(splitPath("README.md"))
```

```text
Split{ dir: "docs/lib/String", base: "lastIndexOf.md" }
Split{ dir: "", base: "README.md" }
```

Find every occurrence from the end, searching again in the text in front of each match, and stop
with `?? break`:

```ant
var s = "the cat and the hat"
var to = String.length(s)
var hits: Array[int] = []
while true {
  var i = String.lastIndexOf(String.slice(s, 0, to), "the") ?? break
  hits.push(i)
  to = i
}
println(hits)
```

```text
[12, 0]
```

## Properties and relations

The result is `null` exactly when [`String.contains`](contains.md) is `false`, and it is never
before the index [`String.indexOf`](indexOf.md) gives:

```ant
var s = "banana"
for t in ["an", "nab", "a", "x"] {
  var first = String.indexOf(s, t)
  var last = String.lastIndexOf(s, t)
  assert((last == null) == !String.contains(s, t))
  if first != null && last != null {
    assert(first <= last)
  }
}
```

When `t` occurs, the slice of `s` at the index is `t`, and `t` does not occur in the text after the
index:

```ant
var s = "mississippi"
var t = "ssi"
var i = String.lastIndexOf(s, t)
assert(i != null)
assert(String.slice(s, i, i + String.length(t)) == t)
assert(!String.contains(String.slice(s, i + 1, String.length(s)), t))
```

For a one-code-point `t`, the index is the one [`Array.lastIndexOf`](../Array/lastIndexOf.md) gives
on [`String.chars`](chars.md):

```ant
var s = "hello"
for c in ["h", "l", "o", "z"] {
  assert(String.lastIndexOf(s, c) == Array.lastIndexOf(String.chars(s), c))
}
```

## Possible issues

A missing match is `null`, not `-1`, so a test written as in JavaScript is a compile error:

```ant
if String.lastIndexOf("abc", "z") >= 0 {
  // error: >= needs both operands of one type, got int | null and int
  println("found")
}
```

Compare against `null` instead, or use [`String.contains`](contains.md) when the index does not
matter:

```ant
String.lastIndexOf("abc", "z") != null // false
String.contains("abc", "z")            // false
```

The result is not an `int` until you narrow it, so arithmetic on it is a compile error:

```ant
var i = String.lastIndexOf("a.b", ".")
println(i + 1) // error: + needs both operands of one type, got int | null and int
```

Give a default with `??` when one makes sense:

```ant
(String.lastIndexOf("a.b", ".") ?? -1) + 1 // 2
```

The index counts code points, so it differs from a byte offset once non-ASCII text comes before the
match:

```ant
String.lastIndexOf("café!", "!")     // 4
Bytes.length(String.toBytes("café")) // 5
```

The search is case-sensitive, and a precomposed `é` does not contain the code point `e`:

```ant
String.lastIndexOf("Ant ant", "Ant")           // 0
String.lastIndexOf("Ant ant".toLower(), "ant") // 4
String.lastIndexOf("café", "e")                // null
```

The index counts code points of the lowered copy, which can be longer than `s`, since
[`String.toLower`](toLower.md) uses the full case mappings:

```ant
String.lastIndexOf("İx", "x")           // 1
String.lastIndexOf("İx".toLower(), "x") // 2
```

## See also

[String.indexOf](indexOf.md) · [String.contains](contains.md) · [String.endsWith](endsWith.md) ·
[String.count](count.md) · [String.slice](slice.md) · [String.chars](chars.md) ·
[Array.lastIndexOf](../Array/lastIndexOf.md) · [Bytes.indexOf](../Bytes/indexOf.md) ·
[?? operator](../../language/operators/null-coalescing.md)
