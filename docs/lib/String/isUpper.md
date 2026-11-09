<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.isUpper

```ant signature
func String.isUpper(s: string): bool
```

String.isUpper reports whether every code point of s is an upper-case letter (Lu). It is false on
"".

<!-- end generated -->

## Details

- `true` when every code point of `s` has the general category Lu; `false` on `""`.
- Only Lu counts: a title-case letter such as `ǅ`, a Roman numeral such as `Ⅻ`, and a circled letter
  such as `Ⓐ` are not upper-case letters.

## Basic examples

```ant
String.isUpper("ANT") // true
String.isUpper("Ant") // false
String.isUpper("Q")   // true
```

Call it through `.`:

```ant
"NASA".isUpper() // true
```

## Scope

The empty string has no code points to test, and gives `false`:

```ant
String.isUpper("") // false
```

Every code point must be an upper-case letter, so a digit, an underscore, or a space makes the
result `false`:

```ant
String.isUpper("HTTP2")    // false
String.isUpper("MAX_SIZE") // false
String.isUpper("NEW YORK") // false
```

Upper-case letters of every script count, including the capital sharp s `ẞ` and the dotted `İ`:

```ant
String.isUpper("ÉTÉ") // true
String.isUpper("ΣΩ")  // true
String.isUpper("ЖУК") // true
String.isUpper("ẞ")   // true
String.isUpper("İ")   // true
```

The full-width `Ａ` and the mathematical bold `𝐀` are Lu too:

```ant
String.isUpper("Ａ") // true
String.isUpper("𝐀") // true
```

Some code points look like capitals but are not in Lu: the Roman numeral `Ⅻ` is a number, the
circled `Ⓐ` is a symbol, and the title-case `ǅ` is Lt:

```ant
String.isUpper("Ⅻ")  // false
String.category("Ⅻ") // "Nl"
String.isUpper("Ⓐ")  // false
String.category("Ⓐ") // "So"
String.isUpper("ǅ")  // false
String.category("ǅ") // "Lt"
```

Test one code point at a time with `String.chars`, here in a pipe:

```ant
"aBcD" | String.chars() | Array.map({ String.isUpper(x) }) // [false, true, false, true]
```

Pass it to a higher-order function as a function value:

```ant
["OK", "Ok", "", "A1", "ΔΣ"].filter(String.isUpper) // ["OK", "ΔΣ"]
```

## Applications

Build an acronym from the capitals of a phrase:

```ant
var phrase = "Portable Network Graphics"
String.concat(phrase.chars().filter({ String.isUpper(x) })) // "PNG"
```

Split a PascalCase name into words at each upper-case letter:

```ant
func words(name: string): Array[string] {
  var out: Array[string] = []
  var word = ""
  for c in name {
    if word != "" && String.isUpper(c) {
      out.push(word)
      word = ""
    }
    word += c
  }
  if word != "" {
    out.push(word)
  }
  out
}

words("HttpRequestParser") // ["Http", "Request", "Parser"]
words("HTTPServer")        // ["H", "T", "T", "P", "Server"]
```

Each upper-case letter starts a new word, so a run of capitals such as an acronym splits letter by
letter.

Flag a message whose words are all in capitals:

```ant
func shouting(text: string): bool {
  var ws = String.fields(text)
  ws.length() > 0 && ws.every(String.isUpper)
}

for msg in ["STOP RIGHT THERE", "Stop right there", "OK THEN", "NO 2"] {
  println("${msg}: ${shouting(msg)}")
}
```

```text
STOP RIGHT THERE: true
Stop right there: false
OK THEN: true
NO 2: false
```

## Properties and relations

`String.isUpper` is `true` exactly when the string is not empty and every code point has the
category `"Lu"`:

```ant
for s in ["", "A", "ABC", "Ab", "ǅ", "Ⅻ", "Ω", "A1", "日"] {
  var byCategory = s != "" && s.chars().every({ String.category(x) == "Lu" })
  assert(String.isUpper(s) == byCategory)
}
```

An upper-case letter is a letter, so `String.isUpper` implies [`String.isLetter`](isLetter.md), and
no string is both upper and lower case:

```ant
for s in ["A", "a", "ǅ", "日", "Ab", "Σ", "ẞ"] {
  assert(!String.isUpper(s) || String.isLetter(s))
  assert(!(String.isUpper(s) && String.isLower(s)))
}
```

## Possible issues

`String.isUpper` does not ask whether `String.toUpper` would leave the string unchanged. A string
with a digit, a space, or an underscore is already in upper case, but is not all upper-case letters:

```ant
String.toUpper("MAX_SIZE") == "MAX_SIZE" // true
String.isUpper("MAX_SIZE")               // false
```

To check that upper-casing leaves a string unchanged, the test a constant name such as `"MAX_SIZE"`
passes, compare it with `String.toUpper(s)` as above.

The same accented capital can be one code point or a letter followed by a combining accent. The
combining accent is a mark, Mn, so the second spelling is not all upper-case letters. Normalize to
NFC first:

```ant
var decomposed = String.normalize("ÉTÉ", "NFD")
String.isUpper(decomposed)                          // false
String.isUpper(String.normalize(decomposed, "NFC")) // true
```

A title-case letter such as `ǅ` is not upper case, but `String.toUpper` turns it into the upper-case
`Ǆ`. Upper-casing can also change the length: `String.toUpper` turns the lower-case `ß` into two
letters:

```ant
String.isUpper("ǅ")                 // false
String.toUpper("ǅ")                 // "Ǆ"
String.isUpper(String.toUpper("ǅ")) // true
String.toUpper("straße")            // "STRASSE"
```

`String.isUpper` tests that the whole string is upper-case letters, not that it holds one. Use
`Array.some` for that:

```ant
String.isUpper("abC")                     // false
"abC".chars().some({ String.isUpper(x) }) // true
```

## See also

[String.isLower](isLower.md) · [String.isLetter](isLetter.md) · [String.toUpper](toUpper.md) ·
[String.toLower](toLower.md) · [String.category](category.md) · [String.normalize](normalize.md) ·
[String.chars](chars.md) · [Array.every](../Array/every.md) · [Array.some](../Array/some.md)
