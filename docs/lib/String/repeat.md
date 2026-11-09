<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.repeat

```ant signature
func String.repeat(s: string, n: int): string
```

String.repeat returns s repeated n times. It panics if n \< 0 or n is above 2^31 - 1.

<!-- end generated -->

## Details

- An `n` of `0`, or an empty `s`, gives `""`.
- Panics if `n` is negative.
- `n` counts copies, not code points. To fill up to a width, use [`String.padStart`](padStart.md) or
  [`String.padEnd`](padEnd.md).

## Basic examples

Repeat a string three times:

```ant
String.repeat("ab", 3) // "ababab"
```

Call it through `.` on any string:

```ant
"-".repeat(10) // "----------"
```

## Scope

Zero copies give the empty string, and one copy gives `s` itself:

```ant
String.repeat("ab", 0) // ""
String.repeat("ab", 1) // "ab"
```

An empty `s` gives the empty string for any count, even the largest:

```ant
String.repeat("", 5)          // ""
String.repeat("", 2147483647) // ""
```

`s` may be text in any script, an emoji, or an escape such as `\n`:

```ant
String.repeat("日本", 2) // "日本日本"
String.repeat("🐜", 3)   // "🐜🐜🐜"
String.repeat("ab\n", 2) // "ab\nab\n"
```

The count may be any `int` expression:

```ant
var word = "ant"
String.repeat("^", String.length(word)) // "^^^"
```

On a string that may be `null`, call it through `?.`; the result is `null` when the string is:

```ant
var name: string | null = null
name?.repeat(2) // null

var tag: string | null = "ab"
tag?.repeat(2) // "abab"
```

In a pipeline, the string is the first argument:

```ant
"ha" | String.repeat(3) // "hahaha"
```

## Named arguments

Name the arguments `s` and `n`, in either order:

```ant
String.repeat(s: "x", n: 4) // "xxxx"
String.repeat(n: 2, s: "o") // "oo"
```

## Applications

Underline a title with a rule as long as the title:

```ant
func heading(title: string): string {
  "${title}\n${String.repeat("=", String.length(title))}"
}

println(heading("Ant docs"))
```

```text
Ant docs
========
```

Indent a line by its depth in a tree:

```ant
type Item = { name: string, depth: int }

var items = [Item{ name: "src", depth: 0 }, Item{ name: "lib", depth: 1 }, Item{ name: "a.ant", depth: 2 }]
for item in items {
  println("${String.repeat("  ", item.depth)}${item.name}")
}
```

```text
src
  lib
    a.ant
```

Draw a progress bar from two repeated pieces:

```ant
func bar(done: int, total: int): string {
  "[${String.repeat("#", done)}${String.repeat(".", total - done)}]"
}

bar(3, 10)  // "[###.......]"
bar(10, 10) // "[##########]"
```

Print a small bar chart:

```ant
type Count = { name: string, count: int }

var counts = [Count{ name: "a", count: 3 }, Count{ name: "b", count: 5 }, Count{ name: "c", count: 1 }]
for c in counts {
  println("${c.name} ${String.repeat("*", c.count)}")
}
```

```text
a ***
b *****
c *
```

## Properties and relations

The length of the result is `n` times the length of `s`:

```ant
for s in ["", "a", "日本", "🐜!"] {
  for n in [0, 1, 4] {
    assert(String.length(String.repeat(s, n)) == n * String.length(s))
  }
}
```

Repeating `a + b` times is repeating `a` times and then `b` times:

```ant
var s = "xy"
var a = 2
var b = 3
assert(String.repeat(s, a + b) == String.repeat(s, a) + String.repeat(s, b))
```

It is [String.concat](concat.md) of an array of `n` copies:

```ant
var s = "ab"
assert(String.repeat(s, 4) == String.concat(Array.filled(4, s)))
```

Repeating a repeat multiplies the counts:

```ant
assert(String.repeat(String.repeat("ab", 2), 3) == String.repeat("ab", 6))
```

For a single code point, it is [String.padStart](padStart.md) of the empty string:

```ant
assert(String.repeat("*", 5) == String.padStart("", 5, "*"))
```

## Possible issues

A negative count panics; it does not give the empty string:

```ant
String.repeat("ab", -1) // panic: String.repeat needs a count of at least 0, got -1
```

Clamp a count that may go below zero with [Math.max](../Math/max.md):

```ant
var width = 4
var title = "library"
String.repeat(".", Math.max(width - String.length(title), 0)) // ""
```

A count above 2^31 − 1 panics:

```ant
String.repeat("a", 2147483648) // panic: String.repeat needs a count of at most 2147483647, got 2147483648
```

A count that is in range still panics when the result would be longer than 2^31 − 1 bytes:

```ant
String.repeat("ab", 1073741824) // panic: String.repeat would give a string longer than 2147483647 bytes
```

`n` counts copies, not code points, so a string of several code points overshoots a width. Use
[String.padEnd](padEnd.md) to fill an exact width:

```ant
String.repeat("ab", 5)     // "ababababab"
String.padEnd("", 5, "ab") // "ababa"
```

The count is an `int`. A `float64`, or an `int` that may be `null`, is a compile error:

```ant
var w = 2.5
var s = String.repeat("a", w) // error: cannot use float64 where int is expected
```

```ant
var i = String.indexOf("a.b", ".")
var s = String.repeat("-", i) // error: cannot use int | null where int is expected
```

Give a default with `??` to turn the `int | null` into an `int`:

```ant
String.repeat("-", String.indexOf("a.b", ".") ?? 0) // "-"
```

Repeating a combining mark does not give separate characters: every copy joins the grapheme before
it:

```ant
var s = "e" + String.repeat("\u{301}", 3)
String.length(s)         // 4
len(String.graphemes(s)) // 1
```

## See also

[String.padStart](padStart.md) · [String.padEnd](padEnd.md) · [String.concat](concat.md) ·
[String.join](join.md) · [Array.filled](../Array/filled.md) · [String.length](length.md)
