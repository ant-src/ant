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
- `t` matches code point by code point, with no case folding or normalization, so `é` written as `e`
  plus a combining accent does not match a precomposed `é`; normalize both with
  [`String.normalize`](normalize.md) first.
- The empty string occurs at every index, so an empty `t` gives the length of `s`.

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

The last occurrence wins:

```ant
String.lastIndexOf("abcabc", "ab") // 3
String.lastIndexOf("a-b-c", "-")   // 3
```

The last occurrence may overlap an earlier one, unlike the occurrences [`String.count`](count.md)
counts:

```ant
String.lastIndexOf("aaa", "aa") // 1
String.count("aaa", "aa")       // 1
```

`t` may be all of `s`, but not longer:

```ant
String.lastIndexOf("ant", "ant")  // 0
String.lastIndexOf("ant", "ants") // null
```

Nothing but `""` occurs in `""`, and `""` occurs at the end of every string:

```ant
String.lastIndexOf("", "a")    // null
String.lastIndexOf("", "")     // 0
String.lastIndexOf("café", "") // 4
String.indexOf("café", "")     // 0
```

The index counts code points, not bytes. Each of `ï` and `é` is one code point and two bytes:

```ant
String.lastIndexOf("naïve café, café", "café") // 12
String.lastIndexOf("🐜🐝🐜🐞", "🐜")           // 2
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
  let i = name.lastIndexOf(".") ?? return ""
  name.slice(i + 1, name.length())
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
  let i = path.lastIndexOf("/")
  if i == null { return Split{ dir: "", base: path } }
  Split{ dir: path.slice(0, i), base: path.slice(i + 1, path.length()) }
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
let s = "the cat and the hat"
var to = s.length()
let hits: Array[int] = []
while true {
  let i = s.slice(0, to).lastIndexOf("the") ?? break
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
let s = "banana"
for t in ["an", "nab", "a", "x"] {
  let first = s.indexOf(t)
  let last = s.lastIndexOf(t)
  assert((last == null) == !s.contains(t))
  if first != null && last != null {
    assert(first <= last)
  }
}
```

When `t` occurs, the slice of `s` at the index is `t`:

```ant
let s = "mississippi"
let t = "ssi"
let i = s.lastIndexOf(t)
assert(i != null)
assert(s.slice(i, i + t.length()) == t)
```

For a one-code-point `t`, the index is the one [`Array.lastIndexOf`](../Array/lastIndexOf.md) gives
on [`String.chars`](chars.md):

```ant
let s = "hello"
for c in ["h", "l", "o", "z"] {
  assert(String.lastIndexOf(s, c) == Array.lastIndexOf(s.chars(), c))
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
let i = String.lastIndexOf("a.b", ".")
println(i + 1) // error: + needs both operands of one type, got int | null and int
```

Give a default with `??` when one makes sense:

```ant
(String.lastIndexOf("a.b", ".") ?? -1) + 1 // 2
```

The index counts code points, so it differs from a byte offset once non-ASCII text comes before the
match:

```ant
String.lastIndexOf("café!", "!") // 4
"café".toBytes().length()        // 5
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
