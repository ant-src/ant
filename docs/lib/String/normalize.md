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
String.normalize("\u{E9}", "NFD").toCodePoints() // [101, 769]
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
let s = "ﬁancé"
s.normalize("NFC").length()  // 5
s.normalize("NFD").length()  // 6
s.normalize("NFKC").length() // 6
s.normalize("NFKD").length() // 7
```

Compatibility characters include full-width letters and digits, superscripts, and Roman numerals:

```ant
String.normalize("Ａｎｔ １２３", "NFKC") // "Ant 123"
String.normalize("x²", "NFKC")            // "x2"
String.normalize("Ⅻ", "NFKC")             // "XII"
```

Some characters have a canonical equivalent of their own, so even `"NFC"` replaces them. The
Ångström sign U+212B becomes the letter `Å`, U+00C5:

```ant
String.normalize("\u{212B}", "NFC").codePoint() // 197
```

`"NFD"` splits a Korean Hangul syllable into its three jamo, and `"NFC"` joins them again:

```ant
String.normalize("한", "NFD").length()         // 3
"한".normalize("NFD").normalize("NFC") == "한" // true
```

`"NFC"` puts several combining marks in a canonical order, so the same marks typed in a different
order compare equal after normalizing:

```ant
let a = "q\u{307}\u{323}"
let b = "q\u{323}\u{307}"
a == b                                   // false
a.normalize("NFC") == b.normalize("NFC") // true
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
  s
    | String.normalize("NFD")
    | String.chars()
    | Array.filter({ !x.isMark() })
    | String.join("")
}

stripAccents("Crème Brûlée") // "Creme Brulee"
stripAccents("Ångström")     // "Angstrom"
```

Key a map by normalized names, so a name typed with a combining accent finds the same entry:

```ant
let ages: Map[string, int] = Map{}
func key(name: string): string {
  name.normalize("NFC")
}
ages[key("Zoë")] = 31
ages[key("Zoe\u{308}")] // 31
```

Build a search key that ignores case, accents, and compatibility forms by combining `"NFKD"`,
`String.isMark`, and `String.foldCase`:

```ant
func searchKey(s: string): string {
  let bare = s.normalize("NFKD").chars().filter({ !x.isMark() })
  String.join(bare, "").foldCase()
}

searchKey("Ｃａｆé") == searchKey("CAFE") // true
searchKey("ﬁlm") == searchKey("Film")     // true
```

## Properties and relations

Normalizing twice gives the same result as normalizing once, in every form:

```ant
let s = "ﬁancé Å 한"
for form in ["NFC", "NFD", "NFKC", "NFKD"] {
  let once = s.normalize(form)
  assert(once.normalize(form) == once)
}
```

`"NFC"` of the `"NFD"` form is the `"NFC"` form, and the same holds for the compatibility forms:

```ant
let s = "ﬁancé Å 한"
assert(s.normalize("NFD").normalize("NFC") == s.normalize("NFC"))
assert(s.normalize("NFKD").normalize("NFKC") == s.normalize("NFKC"))
```

## Possible issues

The form is spelled in upper case. Any other spelling panics:

```ant
String.normalize("e\u{301}", "nfc") // panic: String.normalize needs the form NFC, NFD, NFKC or NFKD, got "nfc"
```

`"NFC"` does not always give one code point per character. Where Unicode has no precomposed
character, the marks stay separate; count with `String.graphemes` instead:

```ant
let s = String.normalize("q\u{307}\u{323}", "NFC")
s.length()             // 3
s.graphemes().length() // 1
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
String.equalFold("\u{C9}", "e\u{301}")                                   // false
String.equalFold("\u{C9}".normalize("NFC"), "e\u{301}".normalize("NFC")) // true
```

## See also

[String.foldCase](foldCase.md) · [String.equalFold](equalFold.md) · [String.isMark](isMark.md) ·
[String.graphemes](graphemes.md) · [String.chars](chars.md) · [String.toCodePoints](toCodePoints.md)
· [String.length](length.md)
