<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.concat

```ant signature
func String.concat(parts: Array[string]): string
```

String.concat returns the strings in parts joined with nothing between them.

<!-- end generated -->

## Details

- Returns a new string; `parts` is unchanged.
- Takes one array. To join two strings, use `+`.
- `parts.concat()` on an array is [Array.concat](../Array/concat.md). Use a pipe instead:
  `parts | String.concat()`.

## Basic examples

Join three strings:

```ant
String.concat(["ant", "hill", "s"]) // "anthills"
```

Nothing goes between the parts. To put a separator between them, use [String.join](join.md):

```ant
String.concat(["a", "b", "c"])    // "abc"
String.join(["a", "b", "c"], ",") // "a,b,c"
```

## Scope

An empty array gives the empty string, and empty parts add nothing:

```ant
var none: Array[string] = []
String.concat(none)          // ""
String.concat(["", "x", ""]) // "x"
String.concat(["", ""])      // ""
```

Parts in any script join the same way:

```ant
String.concat(["é", "日本", "🐜"]) // "é日本🐜"
```

A part may be a combining mark, which joins the character before it into one grapheme of two code
points:

```ant
var s = String.concat(["e", "\u{301}"])
String.length(s)         // 2
len(String.graphemes(s)) // 1
```

Pass the parts with a pipe, which puts the array in as the first argument:

```ant
["a", "b"] | String.concat() // "ab"
```

The array can come from any expression, such as a map over numbers:

```ant
String.concat(Array.map([1, 2, 3], { String.from(x) })) // "123"
```

## Named arguments

Name the argument `parts`:

```ant
String.concat(parts: ["x", "y"]) // "xy"
```

## Applications

Remove every separator from a string:

```ant
String.concat(String.split("555-0100", "-")) // "5550100"
```

Reverse a word:

```ant
String.concat(Array.reverse(String.graphemes("ant"))) // "tna"
```

Collect pieces in an array as a loop goes, then join them once:

```ant
var parts: Array[string] = []
for i in Array.range(0, 3) {
  parts.push("${i};")
}
println(String.concat(parts))
```

```text
0;1;2;
```

Build a banner from repeated pieces:

```ant
func banner(title: string): string {
  var rule = String.repeat("=", 3)
  String.concat([rule, " ", title, " ", rule])
}

println(banner("ant"))
```

```text
=== ant ===
```

## Properties and relations

It is [String.join](join.md) with an empty separator:

```ant
var parts = ["a", "bc", "", "d"]
assert(String.concat(parts) == String.join(parts, ""))
```

For two parts, it is `+`:

```ant
var a = "ant"
var b = "hill"
assert(String.concat([a, b]) == a + b)
```

The length of the result is the sum of the lengths of the parts:

```ant
var parts = ["é", "日本", "", "🐜"]
var total = Array.sum(Array.map(parts, { String.length(x) })) ?? 0
assert(String.length(String.concat(parts)) == total)
```

It undoes [String.chars](chars.md) and [String.graphemes](graphemes.md):

```ant
var s = "héllo 日本"
assert(String.concat(String.chars(s)) == s)
assert(String.concat(String.graphemes(s)) == s)
```

It is associative: joining joined parts gives the same string as joining them all at once:

```ant
var a = ["a", "b"]
var b = ["c"]
assert(String.concat([String.concat(a), String.concat(b)]) == String.concat([...a, ...b]))
```

## Possible issues

`parts.concat()` reaches [Array.concat](../Array/concat.md), which joins two arrays, so the call is
a compile error. Call `String.concat(parts)` or use a pipe:

<!-- test: error Array.concat needs an argument for ys -->

```ant
var s = ["a", "b"].concat()
```

The parts go in one array. Strings passed as separate arguments are a compile error:

<!-- test: error String.concat takes 1 argument, got more -->

```ant
var s = String.concat("a", "b")
```

Every part must be a string. Convert other values with [String.from](from.md) first:

<!-- test: error cannot use int where string is expected -->

```ant
var s = String.concat(["a", 1])
```

A part that may be `null` is a compile error too. Supply a fallback with `??`:

<!-- test: error cannot use string | null where string is expected -->

```ant
var x: string | null = null
var s = String.concat(["a", x])
```

The result is not normalized. An `e` joined with a combining accent looks like `é`, but it is not
equal to the one-code-point `é`, U+00E9, until [String.normalize](normalize.md) brings both to one
form:

```ant
var s = String.concat(["e", "\u{301}"])
s == "\u{e9}"                          // false
String.normalize(s, "NFC") == "\u{e9}" // true
```

Reversing by code points moves a combining mark onto another character. Reverse the graphemes
instead:

```ant
var s = "ne\u{301}"
String.toCodePoints(s)                                                 // [110, 101, 769]
String.toCodePoints(String.concat(Array.reverse(String.chars(s))))     // [769, 101, 110]
String.toCodePoints(String.concat(Array.reverse(String.graphemes(s)))) // [101, 769, 110]
```

## See also

[String.join](join.md), [String.split](split.md), [String.repeat](repeat.md),
[String.chars](chars.md), [String.graphemes](graphemes.md), [String.from](from.md),
[Array.concat](../Array/concat.md), [Bytes.concat](../Bytes/concat.md)
