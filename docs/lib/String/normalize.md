<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.normalize

```ant signature
func String.normalize(s: string, form: string): string
```

String.normalize returns s in the Unicode normalization form form, one of "NFC", "NFD", "NFKC", or
"NFKD". It panics on any other form.

<!-- end generated -->

## Details

- `form` is `"NFC"`, `"NFD"`, `"NFKC"`, or `"NFKD"`, in upper case. Any other string panics.
- `"NFD"` splits an accented letter into the letter and its combining marks; `"NFC"` joins them back
  where Unicode has one code point. The `K` forms also replace compatibility characters, such as `ﬁ`
  with `fi`.
- `==`, [`String.foldCase`](foldCase.md), and [`String.equalFold`](equalFold.md) do not normalize.
  Normalize text that can arrive in either spelling before you compare it or use it as a key.

## Basic examples

The letter `é` has two spellings: one code point, or an `e` followed by a combining acute accent.
They look the same but are not `==`. `"NFC"` brings the second to the first:

```ant
"\u{E9}" == "e\u{301}"                          // false
String.normalize("e\u{301}", "NFC") == "\u{E9}" // true
```

`"NFD"` goes the other way and splits the accent from its letter:

```ant
String.toCodePoints(String.normalize("\u{E9}", "NFD")) // [101, 769]
```

Call it through `.` on the string:

```ant
"e\u{301}".normalize("NFC").length() // 1
```

## Scope

The empty string and plain ASCII come back unchanged in every form:

```ant
String.normalize("", "NFC")         // ""
String.normalize("Ant 1.0", "NFKD") // "Ant 1.0"
```

Each form, on a string with an accent and a ligature. The canonical forms keep the ligature `ﬁ`; the
compatibility forms replace it with `fi`:

```ant
var s = "ﬁancé"
String.length(String.normalize(s, "NFC"))  // 5
String.length(String.normalize(s, "NFD"))  // 6
String.length(String.normalize(s, "NFKC")) // 6
String.length(String.normalize(s, "NFKD")) // 7
```

Compatibility characters include full-width letters and digits, superscripts, and Roman numerals:

```ant
String.normalize("Ａｎｔ １２３", "NFKC") // "Ant 123"
String.normalize("x²", "NFKC")      // "x2"
String.normalize("Ⅻ", "NFKC")       // "XII"
```

Some characters have a canonical equivalent of their own, so even `"NFC"` replaces them. The
Ångström sign U+212B becomes the letter `Å`, U+00C5:

```ant
String.codePoint(String.normalize("\u{212B}", "NFC")) // 197
```

`"NFD"` splits a Korean Hangul syllable into its three jamo, and `"NFC"` joins them again:

```ant
String.length(String.normalize("한", "NFD"))                  // 3
String.normalize(String.normalize("한", "NFD"), "NFC") == "한" // true
```

`"NFC"` puts several combining marks in a canonical order, so the same marks typed in a different
order compare equal after normalizing:

```ant
var a = "q\u{307}\u{323}"
var b = "q\u{323}\u{307}"
a == b                                                   // false
String.normalize(a, "NFC") == String.normalize(b, "NFC") // true
```

In a pipe, the left side is `s`:

```ant
"e\u{301}" | String.normalize("NFC") | String.length() // 1
```

Name the arguments:

```ant
String.normalize(s: "e\u{301}", form: "NFC") == "\u{E9}" // true
```

## Applications

Strip the accents from text: split them off with `"NFD"`, then drop the code points that
`String.isMark` accepts:

```ant
func stripAccents(s: string): string {
  String.join(String.normalize(s, "NFD").chars().filter({ !String.isMark(x) }), "")
}

stripAccents("Crème Brûlée") // "Creme Brulee"
stripAccents("Ångström")     // "Angstrom"
```

Key a map by normalized names, so a name typed with a combining accent finds the same entry:

```ant
var ages: Map[string, int] = Map{}
func key(name: string): string {
  String.normalize(name, "NFC")
}
ages[key("Zoë")] = 31
ages[key("Zoe\u{308}")] // 31
```

Build a search key that ignores case, accents, and compatibility forms by combining `"NFKD"`,
`String.isMark`, and `String.foldCase`:

```ant
func searchKey(s: string): string {
  var bare = String.normalize(s, "NFKD").chars().filter({ !String.isMark(x) })
  String.foldCase(String.join(bare, ""))
}

searchKey("Ｃａｆé") == searchKey("CAFE") // true
searchKey("ﬁlm") == searchKey("Film")  // true
```

## Properties and relations

Normalizing twice gives the same result as normalizing once, in every form:

```ant
var s = "ﬁancé Å 한"
for form in ["NFC", "NFD", "NFKC", "NFKD"] {
  var once = String.normalize(s, form)
  assert(String.normalize(once, form) == once)
}
```

`"NFC"` of the `"NFD"` form is the `"NFC"` form, and the same holds for the compatibility forms:

```ant
var s = "ﬁancé Å 한"
assert(String.normalize(String.normalize(s, "NFD"), "NFC") == String.normalize(s, "NFC"))
assert(String.normalize(String.normalize(s, "NFKD"), "NFKC") == String.normalize(s, "NFKC"))
```

## Possible issues

The form is spelled in upper case. Any other spelling panics:

```ant
String.normalize("e\u{301}", "nfc") // panic: String.normalize needs the form NFC, NFD, NFKC or NFKD, got "nfc"
```

`"NFC"` does not always give one code point per character. Where Unicode has no precomposed
character, the marks stay separate; count with `String.graphemes` instead:

```ant
var s = String.normalize("q\u{307}\u{323}", "NFC")
String.length(s)             // 3
String.graphemes(s).length() // 1
```

The compatibility forms make different strings equal: `x²` and `x2` have the same `"NFKC"` form:

```ant
String.normalize("x²", "NFKC") == String.normalize("x2", "NFKC") // true
```

Normalization does not fold case. Combine it with `String.foldCase` to ignore both:

```ant
String.normalize("É", "NFC") == String.normalize("é", "NFC")                              // false
String.normalize("É", "NFC").foldCase() == String.normalize("e\u{301}", "NFC").foldCase() // true
```

`String.equalFold` does not normalize either; normalize both sides first:

```ant
String.equalFold("\u{C9}", "e\u{301}")                                                   // false
String.equalFold(String.normalize("\u{C9}", "NFC"), String.normalize("e\u{301}", "NFC")) // true
```

## See also

[String.foldCase](foldCase.md) · [String.equalFold](equalFold.md) · [String.isMark](isMark.md) ·
[String.graphemes](graphemes.md) · [String.chars](chars.md) · [String.toCodePoints](toCodePoints.md)
· [String.length](length.md)
