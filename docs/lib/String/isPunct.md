<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.isPunct

```ant signature
func String.isPunct(s: string): bool
```

String.isPunct reports whether every code point of s is punctuation (P). It is false on "".

<!-- end generated -->

## Details

- `true` when every code point of `s` has a general category in P; `false` on `""`.
- Some ASCII signs are symbols, not punctuation: `$`, `+`, `<`, `=`, `>`, `^`, `` ` ``, `|`, `~`.
  [`String.isSymbol`](isSymbol.md) tests those.

## Basic examples

```ant
String.isPunct("!")   // true
String.isPunct("?!")  // true
String.isPunct("Hi!") // false
```

Call it through `.`:

```ant
"...".isPunct() // true
```

## Scope

The empty string has no code points to test, and gives `false`:

```ant
String.isPunct("") // false
```

ASCII punctuation: brackets, quotes, the hyphen-minus, the underscore, `#`, `%`, `&`, `*`, and `@`:

```ant
String.isPunct("()[]{}")    // true
String.isPunct("\"'")       // true
String.isPunct("-_")        // true
String.isPunct("#%&*@")     // true
String.isPunct(",.:;!?/\\") // true
```

The ASCII math and currency signs are symbols, not punctuation:

```ant
String.isPunct("+")  // false
String.isPunct("$")  // false
String.isPunct("<=") // false
String.isPunct("|~") // false
```

Punctuation of every script counts: dashes, curly quotes, guillemets, the inverted question mark,
the ellipsis, and the CJK comma and full stop:

```ant
String.isPunct("—–")   // true
String.isPunct("“”‘’") // true
String.isPunct("«»")   // true
String.isPunct("¿¡")   // true
String.isPunct("…")    // true
String.isPunct("、。") // true
```

The section and pilcrow signs are punctuation; the euro, copyright, degree, and emoji are symbols:

```ant
String.isPunct("§¶") // true
String.isPunct("€")  // false
String.isPunct("©")  // false
String.isPunct("°")  // false
String.isPunct("😀") // false
```

White space is not punctuation:

```ant
String.isPunct(" ")   // false
String.isPunct("! ?") // false
```

[`String.category`](category.md) gives the subcategory of one code point:

```ant
"(—)!_«".chars().map(String.category) // ["Ps", "Pd", "Pe", "Po", "Pc", "Pi"]
```

Test one code point at a time with `String.chars`, here in a pipe:

```ant
"a,b." | String.chars() | Array.map({ String.isPunct(x) }) // [false, true, false, true]
```

Pass it to a higher-order function as a function value:

```ant
["?", "ok", "", "--", "¿"].filter(String.isPunct) // ["?", "--", "¿"]
```

## Applications

Remove the punctuation from a sentence:

```ant
let line = "Well, well: who's there?"
String.concat(line.chars().filter({ !String.isPunct(x) })) // "Well well whos there"
```

Count words, ignoring tokens that are only punctuation, such as a dash between spaces:

```ant
String.fields("Wait — what? ...").countWhere({ !String.isPunct(x) }) // 2
```

Strip the punctuation from both ends of each word, keeping the apostrophe inside:

```ant
func strip(word: string): string {
  let cs = word.chars()
  var start = 0
  var end = cs.length()
  while start < end && String.isPunct(cs[start]) {
    start += 1
  }
  while end > start && String.isPunct(cs[end - 1]) {
    end -= 1
  }
  String.concat(cs.slice(start, end))
}

println(String.fields("\"Don't,\" she said (quietly).").map(strip))
```

```text
["Don't", "she", "said", "quietly"]
```

## Properties and relations

`String.isPunct` is `true` exactly when the string is not empty and every code point has a category
that starts with `P`:

```ant
for s in ["", "!", "?!", "«»", "$", "a.", "—", "😀"] {
  let byCategory = s != "" && s.chars().every({ String.category(x).startsWith("P") })
  assert(String.isPunct(s) == byCategory)
}
```

Punctuation and symbols are separate categories, so no code point is both:

```ant
for c in "!$+-.<@^_`|~«€©§".chars() {
  assert(!(String.isPunct(c) && String.isSymbol(c)))
}
```

Every ASCII printable character that is not a letter, a digit, or the space is either punctuation or
a symbol:

```ant
for n in Iter.range(0x21, 0x7f) {
  let c = String.fromCodePoint(n)
  if c != null && !String.isLetter(c) && !String.isDigit(c) {
    assert(String.isPunct(c) != String.isSymbol(c))
  }
}
```

For nonempty strings, `a + b` is punctuation exactly when both `a` and `b` are:

```ant
for a in ["!", "«", "x"] {
  for b in ["?", "»", "y"] {
    assert(String.isPunct(a + b) == (String.isPunct(a) && String.isPunct(b)))
  }
}
```

## Possible issues

To accept any ASCII character that is not a letter, a digit, or white space, `String.isPunct` alone
is not enough, because some of those characters are symbols. Combine it with
[`String.isSymbol`](isSymbol.md):

```ant
String.isPunct("+=")                                            // false
"+=".chars().every({ String.isPunct(x) || String.isSymbol(x) }) // true
```

The test runs per code point, not per grapheme. A punctuation mark followed by a combining accent
looks like one punctuation character, but the accent is a mark, M, so the string fails.
[`String.isMark`](isMark.md) tests the accent:

```ant
String.isPunct("!\u{301}")            // false
String.isMark("\u{301}")              // true
String.graphemes("!\u{301}").length() // 1
```

Every code point must be punctuation, so `String.isPunct` does not answer whether a string contains
punctuation. Use `Array.some` for that:

```ant
String.isPunct("Hi!")                     // false
"Hi!".chars().some({ String.isPunct(x) }) // true
```

Removing punctuation also removes the apostrophe and the hyphen inside words:

```ant
String.concat("don't re-read".chars().filter({ !String.isPunct(x) })) // "dont reread"
```

## See also

[String.isSymbol](isSymbol.md) · [String.isLetter](isLetter.md) · [String.isSpace](isSpace.md) ·
[String.isGraphic](isGraphic.md) · [String.isPrint](isPrint.md) · [String.category](category.md) ·
[String.chars](chars.md) · [String.fields](fields.md) · [Array.every](../Array/every.md) ·
[Array.some](../Array/some.md)
