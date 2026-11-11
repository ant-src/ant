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
- When `t` does not occur in `s`, the result is `null`, not `-1`.
- An empty `t` occurs at every index, so the result is `0`.
- Matching is exact, code point by code point: no case folding and no normalization.

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
"aaa".indexOf("aa")      // 0
"abababa".indexOf("aba") // 0
```

`t` may be all of `s`, but not longer:

```ant
"ant".indexOf("ant")  // 0
"ant".indexOf("ants") // null
```

Nothing occurs in an empty string, but the empty string occurs at the start of any string:

```ant
"".indexOf("a")   // null
"ant".indexOf("") // 0
"".indexOf("")    // 0
```

The index counts code points, not bytes. Each of `ï` and `é` is one code point and two bytes:

```ant
"naïve café".indexOf("café") // 6
"🐜🐝🐞".indexOf("🐞")       // 2
```

It counts code points, not what a reader sees as characters. The family emoji is five code points,
and the accented `e` below is two:

```ant
"👨‍👩‍👧x".indexOf("x")       // 5
"e\u{301}x".indexOf("x") // 2
```

Call it with named arguments, or in a pipe, where the string is the first argument:

```ant
String.indexOf(s: "a-b", t: "-") // 1
"k=v" | String.indexOf("=")      // 1
```

## Applications

Take the text after the first dot, and return early with `?? return` when there is none:

```ant
func afterFirstDot(name: string): string {
  let i = name.indexOf(".") ?? return ""
  name.slice(i + 1, name.length())
}

println(afterFirstDot("archive.tar.gz"))
println(afterFirstDot("README"))
```

```text
tar.gz

```

Find every occurrence by searching again after each match, and stop with `?? break`:

```ant
let s = "the cat and the hat"
var from = 0
let hits: Array[int] = []
while true {
  let i = s.slice(from, s.length()).indexOf("the") ?? break
  hits.push(from + i)
  from = from + i + 1
}
println(hits)
```

```text
[0, 12]
```

Advancing by one also finds overlapping matches; advance by `String.length(t)` to match
[`String.count`](count.md).

Narrow the result with `!= null` to split a header line at its colon:

```ant
type Header = { name: string, value: string }

func parseHeader(line: string): Header | null {
  let i = line.indexOf(": ")
  if i != null {
    Header{ name: line.slice(0, i), value: line.slice(i + 2, line.length()) }
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
"Content-Type: text/html".cut(": ") // String.Cut{ before: "Content-Type", after: "text/html" }
```

## Properties and relations

The result is `null` exactly when [`String.contains`](contains.md) is `false`:

```ant
let s = "banana"
for t in ["an", "nab", "a", "x"] {
  assert((s.indexOf(t) == null) == !s.contains(t))
}
```

When `t` occurs, the slice of `s` at the index is `t`, and no occurrence of `t` starts before the
index:

```ant
let s = "mississippi"
let t = "ssi"
let i = s.indexOf(t)
assert(i != null)
assert(s.slice(i, i + t.length()) == t)
assert(!s.slice(0, i + t.length() - 1).contains(t))
```

The first index is never after the last one that [`String.lastIndexOf`](lastIndexOf.md) gives:

```ant
for s in ["banana", "an", "nana"] {
  let first = s.indexOf("an")
  let last = s.lastIndexOf("an")
  assert((first == null) == (last == null))
  if first != null && last != null {
    assert(first <= last)
  }
}
```

On the code points of `s`, a one-code-point `t` has the same index that
[`Array.indexOf`](../Array/indexOf.md) gives on [`String.chars`](chars.md):

```ant
let s = "hello"
for c in ["h", "l", "o", "z"] {
  assert(s.indexOf(c) == Array.indexOf(s.chars(), c))
}
```

## Possible issues

A missing match is `null`, not `-1`, so a test written as in JavaScript is a compile error:

```ant
if "abc".indexOf("z") >= 0 {
  // error: >= needs both operands of one type, got int | null and int
  println("found")
}
```

Compare against `null` instead, or use [`String.contains`](contains.md) when the index does not
matter:

```ant
"abc".indexOf("z") != null // false
"abc".contains("z")        // false
```

The result is not an `int` until you narrow it, so arithmetic on it is a compile error:

```ant
let i = "abc".indexOf("b")
println(i + 1) // error: + needs both operands of one type, got int | null and int
```

Give a default with `??` when one makes sense:

```ant
("abc".indexOf("b") ?? 0) + 1 // 2
```

The index counts code points. [`Bytes.indexOf`](../Bytes/indexOf.md) on the UTF-8 bytes counts
bytes, and gives a different number for the same text:

```ant
String.indexOf("café!", "!")                    // 4
Bytes.indexOf("café!".toBytes(), "!".toBytes()) // 5
```

The search is case-sensitive, and a precomposed `é` does not contain the code point `e`:

```ant
"Ant".indexOf("ant")           // null
"Ant".toLower().indexOf("ant") // 0
"café".indexOf("e")            // null
```

The search does not normalize either, so the two spellings of `é` do not find each other. Bring both
strings to one form with [`String.normalize`](normalize.md) first:

```ant
"cafe\u{301}".indexOf("\u{E9}")                  // null
"cafe\u{301}".normalize("NFC").indexOf("\u{E9}") // 3
```

## See also

[String.lastIndexOf](lastIndexOf.md) · [String.contains](contains.md) · [String.count](count.md) ·
[String.cut](cut.md) · [String.slice](slice.md) · [String.chars](chars.md) ·
[Array.indexOf](../Array/indexOf.md) · [Bytes.indexOf](../Bytes/indexOf.md) ·
[?? operator](../../language/operators/null-coalescing.md)
