<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.category

```ant signature
func String.category(c: string): string
```

String.category returns the two-letter Unicode general category of c, such as "Lu", or "Cn" when c
is unassigned. It panics unless c is one code point.

<!-- end generated -->

## Details

- Returns the two-letter Unicode general category, such as `"Lu"`, `"Nd"`, or `"Zs"`. The first
  letter is the major class: L letter, M mark, N number, P punctuation, S symbol, Z separator, C
  other.
- A code point that Unicode does not assign gives `"Cn"`.
- Panics unless `c` is exactly one code point. The empty string panics, and so does a character
  written with several code points, such as a flag or a letter with a combining accent.
- `for c in s` and [String.chars](chars.md) give one code point at a time.

## Basic examples

```ant
String.category("A") // "Lu"
String.category("a") // "Ll"
String.category("7") // "Nd"
String.category("!") // "Po"
```

Call it through `.` on a one-code-point string:

```ant
" ".category() // "Zs"
```

## Scope

Letters: upper case, lower case, title case, modifier, and other, such as a Han ideograph that has
no case:

```ant
String.category("Ä")  // "Lu"
String.category("ß")  // "Ll"
String.category("ǅ")  // "Lt"
String.category("ʰ")  // "Lm"
String.category("中") // "Lo"
```

Marks: a combining accent is nonspacing, a Devanagari vowel sign is spacing, and U+20DD, a combining
circle drawn around the character before it, is an enclosing mark:

```ant
String.category("\u{301}")  // "Mn"
String.category("\u{903}")  // "Mc"
String.category("\u{20DD}") // "Me"
```

Numbers: decimal digits in any script, letter numbers such as Roman numerals, and other numbers such
as fractions:

```ant
String.category("٣") // "Nd"
String.category("Ⅻ") // "Nl"
String.category("½") // "No"
```

Punctuation splits into connectors, dashes, opening and closing brackets, initial and final quotes,
and the rest:

```ant
String.category("_") // "Pc"
String.category("-") // "Pd"
String.category("(") // "Ps"
String.category(")") // "Pe"
String.category("«") // "Pi"
String.category("»") // "Pf"
String.category("?") // "Po"
```

Symbols: math, currency, modifier, and other. An emoji of one code point is an other symbol:

```ant
String.category("+")  // "Sm"
String.category("€")  // "Sc"
String.category("^")  // "Sk"
String.category("©")  // "So"
String.category("🐜") // "So"
```

Separators: space, line, and paragraph. A line feed is not a separator but a control character:

```ant
String.category(" ")        // "Zs"
String.category("\u{2028}") // "Zl"
String.category("\u{2029}") // "Zp"
String.category("\n")       // "Cc"
```

The other C categories: control, format such as the soft hyphen, private use, and unassigned:

```ant
String.category("\u{AD}")   // "Cf"
String.category("\u{E000}") // "Co"
String.category("\u{378}")  // "Cn"
```

A code point that `String.fromCodePoint` builds works as well as a literal one. U+FFFF is a
noncharacter, which Unicode never assigns:

```ant
var c = String.fromCodePoint(0xFFFF) ?? "?"
c.category() // "Cn"
```

Map it over the code points of a string, here in a pipe:

```ant
"Hi, 2!" | String.chars() | Array.map({ String.category(x) }) // ["Lu", "Ll", "Po", "Zs", "Nd", "Po"]
```

## Applications

Count the categories in a text:

```ant
var counts: Map[string, int] = Map{}
for c in "Hi, 2 ants!" {
  var k = String.category(c)
  counts[k] = (counts[k] ?? 0) + 1
}
println(counts)
```

```text
Map{ "Lu": 1, "Ll": 5, "Po": 2, "Zs": 2, "Nd": 1 }
```

Remove accents: decompose the text with `String.normalize`, drop the nonspacing marks, and compose
it again:

```ant
func stripAccents(s: string): string {
  var kept = String.normalize(s, "NFD") | String.chars() | Array.filter({ String.category(x) != "Mn" })
  String.normalize(String.concat(kept), "NFC")
}

println(stripAccents("Crème brûlée à Zürich"))
```

```text
Creme brulee a Zurich
```

Remove invisible format characters, such as a soft hyphen or a zero-width space, that a copy from a
web page leaves behind:

```ant
var pasted = "co\u{AD}operate\u{200B}"
String.concat(pasted.chars().filter({ String.category(x) != "Cf" })) // "cooperate"
```

Cf also holds the zero-width joiner that binds an emoji sequence, so this filter splits a family
emoji into three emoji:

```ant
var family = "👨\u{200D}👩\u{200D}👧"
String.concat(family.chars().filter({ String.category(x) != "Cf" })) // "👨👩👧"
```

Split a string into runs of code points that share a major class, the first letter of the category:

```ant
type Run = { major: string, text: string }

func major(c: string): string {
  String.slice(String.category(c), 0, 1)
}

func runs(s: string): Array[Run] {
  var out: Array[Run] = []
  for c in s {
    var m = major(c)
    var last = out.last()
    if last != null && last.major == m {
      last.text = last.text + c
    } else {
      out.push(Run{ major: m, text: c })
    }
  }
  out
}

for r in runs("abc123!?") {
  println("${r.major} ${r.text}")
}
```

```text
L abc
N 123
P !?
```

## Properties and relations

Each class test of `String` except `String.isSpace` is a test on the category of every code point.
`String.isSpace` follows the White_Space property instead:

```ant
for c in ["A", "a", "ǅ", "7", "Ⅻ", "\u{301}", "-", "€", " ", "\n", "\u{AD}", "\u{378}"] {
  var k = String.category(c)
  var m = String.slice(k, 0, 1)
  assert(String.isUpper(c) == (k == "Lu"))
  assert(String.isLower(c) == (k == "Ll"))
  assert(String.isDigit(c) == (k == "Nd"))
  assert(String.isControl(c) == (k == "Cc"))
  assert(String.isLetter(c) == (m == "L"))
  assert(String.isMark(c) == (m == "M"))
  assert(String.isNumber(c) == (m == "N"))
  assert(String.isPunct(c) == (m == "P"))
  assert(String.isSymbol(c) == (m == "S"))
  assert(String.isGraphic(c) == (["L", "M", "N", "P", "S"].contains(m) || k == "Zs"))
  assert(String.isPrint(c) == (["L", "M", "N", "P", "S"].contains(m) || c == " "))
}
```

The result is always two letters, and a code point keeps its category whichever way it was written:

```ant
for c in ["A", "\u{41}", String.fromCodePoint(65) ?? ""] {
  assert(String.category(c) == "Lu")
  assert(String.length(String.category(c)) == 2)
}
```

## Possible issues

`c` must be one code point. A word panics:

```ant
String.category("ab") // panic: String.category needs a string of one code point, got "ab"
```

So does the empty string:

```ant
String.category("") // panic: String.category needs a string of one code point, got ""
```

What a reader sees as one character can be several code points. A flag is two, and an accent written
as a combining mark is a second code point:

```ant
String.category("🇨🇭")       // panic: String.category needs a string of one code point, got "🇨🇭"
String.category("e\u{301}") // panic: String.category needs a string of one code point, got "é"
```

Split such text with `String.chars` and ask about each code point, or normalize it to NFC first when
a precomposed form exists:

```ant
String.chars("e\u{301}").map({ String.category(x) }) // ["Ll", "Mn"]
String.category(String.normalize("e\u{301}", "NFC")) // "Ll"
```

White space is a property, not a category. A tab and a line feed are white space with the category
`"Cc"`, so test white space with [`String.isSpace`](isSpace.md):

```ant
String.category("\t") // "Cc"
String.isSpace("\t")  // true
```

`"Cn"` means unassigned in the tables Ant uses, the Unicode version that `String.unicodeVersion`
names. A code point that a later Unicode version assigns is `"Cn"` until Ant moves to that version:

```ant
String.unicodeVersion // "17.0.0"
```

## See also

[String.inScript](inScript.md) · [String.isLetter](isLetter.md) · [String.isDigit](isDigit.md) ·
[String.isNumber](isNumber.md) · [String.isPunct](isPunct.md) · [String.isSymbol](isSymbol.md) ·
[String.isMark](isMark.md) · [String.isControl](isControl.md) · [String.isSpace](isSpace.md) ·
[String.chars](chars.md) · [String.codePoint](codePoint.md) · [String.normalize](normalize.md)
