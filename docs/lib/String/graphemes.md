<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.graphemes

```ant signature
func String.graphemes(s: string): Array[string]
```

String.graphemes returns the extended grapheme clusters of s (Unicode UAX \#29), the units a reader
sees as characters.

<!-- end generated -->

## Details

- `String.graphemes` splits `s` into extended grapheme clusters, as Unicode's UAX #29 defines them:
  the units a reader sees as one character. Joining the elements gives `s` back
  ([spec: String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)).
- A grapheme can be several code points: a letter with combining accents, a flag made of two
  regional indicators, an emoji with a skin tone, or a family emoji joined with zero width joiners.
  `String.chars` splits all of these apart.
- An empty `s` gives `[]`.
- The split follows the Unicode version that `String.unicodeVersion` names.
- `String.graphemes` does not normalize: `"\u{E9}"` and `"e\u{301}"` are each one grapheme, and the
  two graphemes are still different strings.

## Basic examples

Split a word into its characters:

```ant
String.graphemes("ant") // ["a", "n", "t"]
```

A flag is one grapheme, though it is two code points:

```ant
String.graphemes("🇨🇭!") // ["🇨🇭", "!"]
String.chars("🇨🇭!")     // ["🇨", "🇭", "!"]
```

Count what a reader sees with `len`:

```ant
len(String.graphemes("👍🏽")) // 1
String.length("👍🏽")         // 2
```

## Scope

An empty string gives an empty array:

```ant
String.graphemes("") // []
```

A combining accent stays with the letter before it:

```ant
len(String.graphemes("e\u{301}a"))   // 2
len(String.graphemes("noe\u{308}l")) // 4
```

An accent at the very start has no letter to attach to, so it is a grapheme of its own:

```ant
len(String.graphemes("\u{301}a")) // 2
```

A family emoji is three people and two zero width joiners, five code points in one grapheme:

```ant
var family = "👨‍👩‍👧"
String.length(family)         // 5
len(String.graphemes(family)) // 1
```

Two flags in a row pair up their regional indicators, and a lone indicator left over is a grapheme
of its own:

```ant
String.graphemes("🇨🇭🇩🇪") // ["🇨🇭", "🇩🇪"]
String.graphemes("🇨🇭🇩")  // ["🇨🇭", "🇩"]
```

A Windows line break, `"\r\n"`, is one grapheme:

```ant
String.graphemes("a\r\nb") // ["a", "\r\n", "b"]
```

Three Hangul jamo combine into one syllable, and a Devanagari conjunct holds together:

```ant
len(String.graphemes("\u{1100}\u{1161}\u{11A8}")) // 1
len(String.graphemes("नमस्ते"))                   // 3
```

Call it through `.` on a string, or put the string first in a pipe:

```ant
"x\u{303}y".graphemes().map({ String.length(x) }) // [2, 1]
"Zoë" | String.graphemes()                        // ["Z", "o", "ë"]
```

The parameter is named `s`:

```ant
String.graphemes(s: "ok") // ["o", "k"]
```

As a function value, it splits each element of an array:

```ant
Array.map(["hi", "🐜🐜"], String.graphemes) // [["h", "i"], ["🐜", "🐜"]]
```

## Applications

Reverse text without moving an accent onto another letter or splitting a flag. Reversing the code
points instead breaks both:

```ant
func reverse(s: string): string {
  String.concat(String.graphemes(s).reverse())
}

println(reverse("noe\u{308}l 🇨🇭"))
println(String.concat(String.chars("noe\u{308}l 🇨🇭").reverse()))
```

```text
🇨🇭 lëon
🇭🇨 l̈eon
```

Shorten text to a number of characters for display, never cutting one in half:

```ant
func truncate(s: string, n: int): string {
  var gs = String.graphemes(s)
  if len(gs) <= n {
    return s
  }
  String.concat(gs.slice(0, n)) + "…"
}

println(truncate("Ant 🐜 language", 5))
println(truncate("👨‍👩‍👧👨‍👩‍👧👨‍👩‍👧", 2))
```

```text
Ant 🐜…
👨‍👩‍👧👨‍👩‍👧…
```

Take the initials of a name, keeping each one whole:

```ant
func initials(name: string): string {
  String.fields(name).map({ String.graphemes(x)[0] }) | String.concat()
}

initials("E\u{301}mile Zola") // "ÉZ"
```

## Properties and relations

Joining the graphemes gives the string back, and every grapheme is a non-empty string:

```ant
for s in ["", "ant", "noe\u{308}l", "🇨🇭🇩", "👨‍👩‍👧\r\n"] {
  var gs = String.graphemes(s)
  assert(String.concat(gs) == s)
  assert(gs.every({ !String.isEmpty(x) }))
}
```

A grapheme holds one code point or more, so there are never more graphemes than code points:

```ant
for s in ["", "ant", "e\u{301}", "👍🏽", "नमस्ते"] {
  assert(len(String.graphemes(s)) <= String.length(s))
}
```

For text with no combining marks, joiners, or flags, the graphemes are the code points:

```ant
var s = "plain ASCII text"
assert(String.graphemes(s) == String.chars(s))
```

## Possible issues

`String.length` counts code points, not graphemes. Count graphemes with `len` of the result:

```ant
String.length("e\u{301}")         // 2
len(String.graphemes("e\u{301}")) // 1
```

A `for` loop over a string steps through its code points, not its graphemes. Loop over the graphemes
to visit the characters a reader sees:

```ant
var s = "e\u{301}"
var points = 0
for c in s {
  points += 1
}
var chars = 0
for g in String.graphemes(s) {
  chars += 1
}
assert(points == 2)
assert(chars == 1)
```

`String.slice` takes code point indices, so it can cut a grapheme in half. Here it keeps the first
person and a joiner of the family:

```ant
String.slice("👨‍👩‍👧", 0, 2) | String.length()  // 2
String.graphemes("👨‍👩‍👧")[0] | String.length() // 5
```

`String.graphemes` does not normalize. The two spellings of "é" are each one grapheme, but not the
same string. Normalize first to compare them:

```ant
String.graphemes("\u{E9}") == String.graphemes("e\u{301}") // false
String.graphemes(String.normalize("e\u{301}", "NFC"))      // ["é"]
```

## See also

[String.chars](chars.md) · [String.length](length.md) · [String.slice](slice.md) ·
[String.normalize](normalize.md) · [String.concat](concat.md) · [String.codePoint](codePoint.md) ·
[String.toCodePoints](toCodePoints.md) · [Array.reverse](../Array/reverse.md)
