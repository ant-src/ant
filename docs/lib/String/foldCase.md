<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.foldCase

```ant signature
func String.foldCase(s: string): string
```

String.foldCase returns s with Unicode's full case folding applied, a key in which case does not
matter. It does not normalize; see [`String.normalize`](normalize.md).

<!-- end generated -->

## Details

- Maps every case variant of a string to one form, so it is the key for grouping or deduplicating
  text without regard to case.
- Full case folding can change the length: `"ß"` folds to `"ss"`.
- It does not normalize: `"\u{E9}"` and `"e\u{301}"` fold to different strings. Normalize with
  `String.normalize(s, "NFC")` first.
- To compare two strings, use [String.equalFold](equalFold.md).

## Basic examples

```ant
String.foldCase("Straße")       // "strasse"
String.foldCase("STRASSE")      // "strasse"
String.foldCase("Hello, World") // "hello, world"
```

## Scope

An empty string folds to an empty string, and characters without case pass through unchanged:

```ant
String.foldCase("")           // ""
String.foldCase("123 + 4 🐜") // "123 + 4 🐜"
```

Folding can lengthen a string. The ligature `ﬁ` folds to the two letters `fi`:

```ant
String.foldCase("ß")                  // "ss"
String.foldCase("ﬁle")                // "file"
String.length(String.foldCase("ﬁle")) // 4
```

Letters of other scripts fold too. Every Greek sigma, final or not, folds to `σ`, and the Kelvin
sign `K` folds to the letter `k`:

```ant
String.foldCase("ΣΟΦΟΣ")    // "σοφοσ"
String.foldCase("σοφος")    // "σοφοσ"
String.foldCase("\u{212A}") // "k"
```

Call it through `.` or in a pipe:

```ant
"Ant".foldCase()          // "ant"
"ANT" | String.foldCase() // "ant"
```

Pass it as a function value, where a function from `string` to `string` is expected:

```ant
Array.map(["Ant", "BEE", "wasp"], String.foldCase) // ["ant", "bee", "wasp"]
```

## Applications

Remove duplicates without regard to case, keeping the first spelling:

```ant
let tags = ["Go", "rust", "GO", "Rust", "ant"]
Array.uniqueBy(tags, String.foldCase) // ["Go", "rust", "ant"]
```

Group words by their folded form:

```ant
let words = ["Ant", "ant", "ANT", "Bee"]
Array.groupBy(words, String.foldCase) // Map{ "ant": ["Ant", "ant", "ANT"], "bee": ["Bee"] }
```

Look up a map whose keys are folded, so that any spelling finds the entry:

```ant
let colors = Map{ "red": "#f00", "green": "#0f0" }
Map.get(colors, String.foldCase("RED"))   // "#f00"
Map.get(colors, String.foldCase("Green")) // "#0f0"
```

Count the words of a text without regard to case:

```ant
let text = "The ant and THE bee and the Wasp"
let counts: Map[string, int] = Map{}
for w in String.fields(text) {
  let key = String.foldCase(w)
  Map.set(counts, key, (Map.get(counts, key) ?? 0) + 1)
}
println(counts)
```

```text
Map{ "the": 3, "ant": 1, "and": 2, "bee": 1, "wasp": 1 }
```

Build a key that ignores both case and spelling by normalizing first:

```ant
func key(s: string): string { String.foldCase(String.normalize(s, "NFC")) }
key("Caf\u{E9}") == key("CAFE\u{301}") // true
```

## Properties and relations

`String.equalFold` compares the folded forms:

```ant
for pair in [["Straße", "STRASSE"], ["Ant", "ant"], ["ant", "bee"]] {
  assert(String.equalFold(pair[0], pair[1]) == (String.foldCase(pair[0]) == String.foldCase(pair[1])))
}
```

Folding a folded string changes nothing:

```ant
for s in ["Straße", "ΣΟΦΟΣ", "ﬁle", ""] {
  assert(String.foldCase(String.foldCase(s)) == String.foldCase(s))
}
```

## Possible issues

A folded string is a key for comparison, not text to show. It is not always the lower case:
`String.toLower` keeps `ß`, while `foldCase` turns it into `ss`:

```ant
String.toLower("Straße")  // "straße"
String.foldCase("Straße") // "strasse"
```

Neither `foldCase` nor `==` normalizes. An `é` written as one code point and an `e` with a combining
accent stay different after folding:

```ant
let a = "\u{E9}"
let b = "E\u{301}"
String.foldCase(a) == String.foldCase(b)                                                   // false
String.foldCase(String.normalize(a, "NFC")) == String.foldCase(String.normalize(b, "NFC")) // true
```

The `i` flag of a regular expression folds one code point at a time, so it does not match what full
folding makes equal. Compare through `foldCase` instead:

```ant
Regex.test(/^straße$/i, "STRASSE")                      // false
String.foldCase("straße") == String.foldCase("STRASSE") // true
```

Full case folding leaves out the rules for Turkish and Lithuanian, which depend on the language. A
dotted capital `İ` does not fold to a plain `i` but to an `i` followed by a combining dot:

```ant
String.foldCase("İ") == "i\u{307}" // true
String.foldCase("İ") == "i"        // false
```

For the same reason, the dotless `ı` and the plain `i` fold apart, even though both have the upper
case `I`. Changing the case of a string before folding it can change the key:

```ant
String.toUpper("ı")                                          // "I"
String.foldCase("ı")                                         // "ı"
String.foldCase(String.toUpper("ı")) == String.foldCase("ı") // false
```

## See also

[String.equalFold](equalFold.md) · [String.toLower](toLower.md) · [String.toUpper](toUpper.md) ·
[String.normalize](normalize.md) · [Array.uniqueBy](../Array/uniqueBy.md) ·
[Array.groupBy](../Array/groupBy.md) · [Regex.test](../Regex/test.md) ·
[Strings guide](../../guide/strings.md)
