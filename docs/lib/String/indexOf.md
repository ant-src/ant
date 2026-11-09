<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.indexOf

```ant signature
func String.indexOf(s: string, t: string): int | null
```

String.indexOf returns the code point index of the first occurrence of t in s, or null when t does
not occur.

<!-- end generated -->

## Details

- The index counts code points, like [`String.length`](length.md) and [`String.slice`](slice.md).
- When `t` does not occur in `s`, the result is `null`, not `-1`. Narrow it with `!= null` or give a
  default with `??` before you use it as an `int`.

## Basic examples

```ant
String.indexOf("banana", "an") // 1
String.indexOf("banana", "x")  // null
```

Call it through `.` on any string:

```ant
"hello world".indexOf("o") // 4
"hello world".indexOf("w") // 6
```

## Scope

The first occurrence wins, even when a later one overlaps it:

```ant
String.indexOf("aaa", "aa")   // 0
String.indexOf("abcabc", "c") // 2
```

`t` may be all of `s`, but not longer:

```ant
String.indexOf("ant", "ant")  // 0
String.indexOf("ant", "ants") // null
```

Nothing occurs in an empty string:

```ant
String.indexOf("", "a") // null
```

The index counts code points, not bytes. Each of `ï` and `é` is one code point and two bytes:

```ant
String.indexOf("naïve café", "café") // 6
String.indexOf("🐜🐝🐞", "🐞")       // 2
```

It counts code points, not what a reader sees as characters. The family emoji is five code points,
and the accented `e` below is two:

```ant
String.indexOf("👨‍👩‍👧x", "x")       // 5
String.indexOf("e\u{301}x", "x") // 2
```

Call it with named arguments, or in a pipe, where the string is the first argument:

```ant
String.indexOf(s: "a-b", t: "-") // 1
"k=v" | String.indexOf("=")      // 1
```

## Applications

Take the extension of a file name, the text after the first dot, and return early with `?? return`
when there is none:

```ant
func extension(name: string): string {
  var i = String.indexOf(name, ".") ?? return ""
  String.slice(name, i + 1, String.length(name))
}

println(extension("archive.tar.gz"))
println(extension("README"))
```

```text
tar.gz

```

Find every occurrence by searching again after each match, and stop with `?? break`:

```ant
var s = "the cat and the hat"
var from = 0
var hits: Array[int] = []
while true {
  var i = String.indexOf(String.slice(s, from, String.length(s)), "the") ?? break
  hits.push(from + i)
  from = from + i + 1
}
println(hits)
```

```text
[0, 12]
```

Narrow the result with `!= null` to split a header line at its colon:

```ant
type Header = { name: string, value: string }

func parseHeader(line: string): Header | null {
  var i = String.indexOf(line, ": ")
  if i != null {
    Header{ name: String.slice(line, 0, i), value: String.slice(line, i + 2, String.length(line)) }
  } else {
    null
  }
}

println(parseHeader("Content-Type: text/html"))
println(parseHeader("garbage"))
```

```text
Header{ name: "Content-Type", value: "text/html" }
null
```

When you need both sides rather than the index, [`String.cut`](cut.md) splits around the first
occurrence:

```ant
String.cut("Content-Type: text/html", ": ") // String.Cut{ before: "Content-Type", after: "text/html" }
```

## Properties and relations

The result is `null` exactly when [`String.contains`](contains.md) is `false`:

```ant
var s = "banana"
for t in ["an", "nab", "a", "x"] {
  assert((String.indexOf(s, t) == null) == !String.contains(s, t))
}
```

When `t` occurs, the slice of `s` at the index is `t`, and no occurrence of `t` starts before the
index:

```ant
var s = "mississippi"
var t = "ssi"
var i = String.indexOf(s, t)
assert(i != null)
assert(String.slice(s, i, i + String.length(t)) == t)
assert(!String.contains(String.slice(s, 0, i + String.length(t) - 1), t))
```

The first index is never after the last one that [`String.lastIndexOf`](lastIndexOf.md) gives:

```ant
for s in ["banana", "an", "nana"] {
  var first = String.indexOf(s, "an")
  var last = String.lastIndexOf(s, "an")
  assert((first == null) == (last == null))
  if first != null && last != null {
    assert(first <= last)
  }
}
```

On the code points of `s`, a one-code-point `t` has the same index that
[`Array.indexOf`](../Array/indexOf.md) gives on [`String.chars`](chars.md):

```ant
var s = "hello"
for c in ["h", "l", "o", "z"] {
  assert(String.indexOf(s, c) == Array.indexOf(String.chars(s), c))
}
```

## Possible issues

A missing match is `null`, not `-1`, so a test written as in JavaScript is a compile error:

```ant
if String.indexOf("abc", "z") >= 0 {
  // error: >= needs both operands of one type, got int | null and int
  println("found")
}
```

Compare against `null` instead, or use [`String.contains`](contains.md) when the index does not
matter:

```ant
String.indexOf("abc", "z") != null // false
String.contains("abc", "z")        // false
```

The result is not an `int` until you narrow it, so arithmetic on it is a compile error:

```ant
var i = String.indexOf("abc", "b")
println(i + 1) // error: + needs both operands of one type, got int | null and int
```

Give a default with `??` when one makes sense:

```ant
(String.indexOf("abc", "b") ?? 0) + 1 // 2
```

The index counts code points. [`Bytes.indexOf`](../Bytes/indexOf.md) on the UTF-8 bytes counts
bytes, and gives a different number for the same text:

```ant
String.indexOf("café!", "!")                                // 4
Bytes.indexOf(String.toBytes("café!"), String.toBytes("!")) // 5
```

The search is case-sensitive, and a precomposed `é` does not contain the code point `e`:

```ant
String.indexOf("Ant", "ant")           // null
String.indexOf("Ant".toLower(), "ant") // 0
String.indexOf("café", "e")            // null
```

## See also

[String.lastIndexOf](lastIndexOf.md) · [String.contains](contains.md) · [String.count](count.md) ·
[String.cut](cut.md) · [String.slice](slice.md) · [String.chars](chars.md) ·
[Array.indexOf](../Array/indexOf.md) · [Bytes.indexOf](../Bytes/indexOf.md) ·
[?? operator](../../language/operators/null-coalescing.md)
