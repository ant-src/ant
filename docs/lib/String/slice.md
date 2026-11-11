<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.slice

```ant signature
func String.slice(s: string, from: int, to: int): string
```

String.slice returns the code points of s from index from up to but excluding index to. It panics if
from \> to, or if either is outside 0 ..= String.length(s). A code point range can cut through a
grapheme; see [`String.graphemes`](graphemes.md).

<!-- end generated -->

## Details

- Counts code points, not bytes and not graphemes, so a slice can end inside what a reader sees as
  one character, such as an emoji or a letter with a combining accent.
- `from` is inclusive and `to` is exclusive. `from == to` gives `""`.
- Panics when `from > to`, or when either index is outside `0 ..= String.length(s)`. It does not
  clamp, and a negative index does not count from the end.
- A string has no `s[i]`; use [`String.chars`](chars.md) to get an array of characters to index.

## Basic examples

Take the code points from index 1 up to but excluding index 3:

```ant
String.slice("hello", 1, 3) // "el"
```

Take a prefix from index 0, or a suffix up to the length:

```ant
let s = "anthill"
String.slice(s, 0, 3)                // "ant"
String.slice(s, 3, String.length(s)) // "hill"
```

## Scope

Equal indices give an empty string, at any position up to the length, the empty string included:

```ant
String.slice("hello", 2, 2) // ""
String.slice("hello", 5, 5) // ""
String.slice("", 0, 0)      // ""
```

Indices count code points, so a letter outside ASCII and an emoji are one index each, though they
take several bytes:

```ant
String.slice("héllo 🐜", 1, 2) // "é"
String.slice("héllo 🐜", 6, 7) // "🐜"
```

A grapheme of several code points can be cut. The accent of `e\u{301}`, the second half of a flag,
and the rest of a family emoji fall outside these slices:

```ant
String.slice("e\u{301}x", 0, 1)         // "e"
String.slice("🇨🇭", 0, 1)                // "🇨"
String.length(String.slice("👨‍👩‍👧", 0, 1)) // 1
```

Call it through `.`, in a pipe, or with named arguments:

```ant
"hello".slice(1, 4)                      // "ell"
"hello" | String.slice(1, 4)             // "ell"
String.slice(s: "hello", from: 1, to: 4) // "ell"
```

## Applications

Read the extension of a file name, from the last `.` to the end:

```ant
func extension(name: string): string {
  let i = String.lastIndexOf(name, ".") ?? return ""
  String.slice(name, i + 1, String.length(name))
}

extension("archive.tar.gz") // "gz"
extension("README")         // ""
```

Shorten text to at most `n` code points, ending in `…` when it was cut:

```ant
func shorten(s: string, n: int): string {
  if String.length(s) <= n { s } else if n <= 0 { "" } else { String.slice(s, 0, n - 1) + "…" }
}

shorten("Ant is a small language", 10) // "Ant is a …"
shorten("Ant", 10)                     // "Ant"
shorten("Ant", 0)                      // ""
```

Take the text between two markers with `String.indexOf`:

```ant
func between(s: string, open: string, close: string): string | null {
  let start = String.indexOf(s, open) ?? return null
  let from = start + String.length(open)
  let rest = String.slice(s, from, String.length(s))
  let end = String.indexOf(rest, close) ?? return null
  String.slice(rest, 0, end)
}

between("see [the docs] here", "[", "]") // "the docs"
between("no brackets", "[", "]")         // null
```

Cut the fields of a fixed-width line into a record:

```ant
type Row = { date: string, code: string, amount: string }

func parse(line: string): Row {
  Row{
    date: String.slice(line, 0, 10),
    code: String.slice(line, 10, 13),
    amount: String.trim(String.slice(line, 13, String.length(line))),
  }
}

println(parse("2026-11-08EUR   12.50"))
```

```text
Row{ date: "2026-11-08", code: "EUR", amount: "12.50" }
```

Take the text around a regular expression match, whose `start` and `end` are code point indices:

```ant
let s = "total 42 EUR"
let m = Regex.find(/\d+/, s)
if m != null {
  println(String.slice(s, 0, m.start) + "<" + m.text + ">" + String.slice(s, m.end, String.length(s)))
}
```

```text
total <42> EUR
```

## Properties and relations

The result holds `to - from` code points, and the whole range gives `s` back:

```ant
let s = "héllo 🐜"
assert(String.length(String.slice(s, 2, 5)) == 3)
assert(String.slice(s, 0, String.length(s)) == s)
```

Two adjacent slices join to the slice that spans both:

```ant
let s = "e\u{301}🇨🇭 ant"
let n = String.length(s)
for i in Iter.range(0, n + 1) {
  assert(String.slice(s, 0, i) + String.slice(s, i, n) == s)
}
```

A slice is the same code points that `Array.slice` takes from `String.chars`:

```ant
let s = "añt🐜hill"
assert(String.slice(s, 2, 6) == String.join(Array.slice(String.chars(s), 2, 6), ""))
```

For a match of `Regex.find`, the slice from `start` to `end` is the matched text:

```ant
let s = "añt 🐜 42"
let m = Regex.find(/\d+/, s)
assert(m != null)
assert(String.slice(s, m.start, m.end) == m.text)
```

## Possible issues

An index past the end panics; `String.slice` does not clamp it:

```ant
String.slice("abc", 0, 4) // panic: String.slice: the range 0..4 is not within 0 ..= 3
```

Keep `to` within the string with `Math.min`:

```ant
String.slice("abc", 0, Math.min(4, String.length("abc"))) // "abc"
```

A negative index panics rather than counting from the end. Count from `String.length(s)` instead:

```ant
String.slice("abc", -2, 3)                       // panic: String.slice: the range -2..3 is not within 0 ..= 3
String.slice("abc", String.length("abc") - 2, 3) // "bc"
```

`from` greater than `to` panics, even when both are within the string:

```ant
String.slice("abc", 2, 1) // panic: String.slice: the range 2..1 is not within 0 ..= 3
```

All three arguments are required; there is no form that runs to the end of the string:

```ant
println(String.slice("abc", 1)) // error: String.slice needs an argument for to
```

Code point indices can split a letter from its accent, so a shortened string can end in a bare
letter or a broken emoji. Slice the graphemes when the text is for a reader:

```ant
let s = "cafe\u{301}!"
String.slice(s, 0, 4) == "cafe"                                          // true
String.join(Array.slice(String.graphemes(s), 0, 4), "") == "cafe\u{301}" // true
```

## See also

[String.length](length.md) · [String.chars](chars.md) · [String.graphemes](graphemes.md) ·
[String.indexOf](indexOf.md) · [String.lastIndexOf](lastIndexOf.md) · [String.cut](cut.md) ·
[String.trimPrefix](trimPrefix.md) · [Array.slice](../Array/slice.md) ·
[Bytes.slice](../Bytes/slice.md) · [Regex.find](../Regex/find.md)
