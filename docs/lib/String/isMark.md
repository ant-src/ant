<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.isMark

```ant signature
func String.isMark(s: string): bool
```

String.isMark reports whether every code point of s is a mark (M), such as a combining accent. It is
false on "".

<!-- end generated -->

## Details

- `true` when every code point of `s` has a general category in M (`Mn`, `Mc`, `Me`); `false` on
  `""`.
- A precomposed letter such as `é` (U+00E9) is a letter, not a mark. `String.normalize(s, "NFD")`
  splits it into a letter and a mark, so dropping the marks strips accents.

## Basic examples

```ant
String.isMark("\u{301}")  // true
String.isMark("a")        // false
String.isMark("e\u{301}") // false
```

Call it through `.`:

```ant
"\u{308}".isMark() // true
```

## Scope

The empty string has no code points to test, and gives `false`:

```ant
String.isMark("") // false
```

A string of several marks is all marks. Here a combining acute and a combining diaeresis:

```ant
String.isMark("\u{301}\u{308}") // true
```

All three kinds of mark count: nonspacing, spacing combining, and enclosing:

```ant
String.isMark("\u{301}")  // true
String.isMark("\u{903}")  // true
String.isMark("\u{20DD}") // true
```

Marks of other scripts count too: a Devanagari vowel sign, a Hebrew point, an Arabic vowel mark, and
a Thai vowel sign:

```ant
String.isMark("\u{93F}") // true
String.isMark("\u{5B4}") // true
String.isMark("\u{64E}") // true
String.isMark("\u{E31}") // true
```

A precomposed letter is a letter. An accent that stands alone, without a letter to combine with, is
a symbol:

```ant
String.isMark("é") // false
String.isMark("´") // false
String.isMark("^") // false
```

Test each code point of a decomposed string with `String.chars`, here in a pipe:

```ant
"ṩ" | String.normalize("NFD") | String.chars() | Array.map({ x.isMark() }) // [false, true, true]
```

Pass it to a higher-order function as a function value:

```ant
["\u{301}", "a", "", "\u{20DD}"].map(String.isMark) // [true, false, false, true]
```

## Applications

Remove accents: decompose the text with NFD, drop the marks, and compose the rest with NFC:

```ant
func stripAccents(s: string): string {
  s
    | String.normalize("NFD")
    | String.chars()
    | Array.filter({ !x.isMark() })
    | String.concat()
    | String.normalize("NFC")
}

println(stripAccents("Crème brûlée à Zürich"))
println(stripAccents("Ångström"))
```

```text
Creme brulee a Zurich
Angstrom
```

Find text that stacks too many marks on one character, such as the "zalgo" text that spills over the
lines around it. Count the marks in each grapheme with `String.graphemes`:

```ant
func maxMarks(s: string): int {
  var most = 0
  for g in s.graphemes() {
    most = Math.max(most, g.chars().countWhere({ x.isMark() }))
  }
  most
}

for s in ["café", "cafe\u{301}", "z\u{336}\u{301}\u{308}\u{30A}\u{352}a"] {
  println(if maxMarks(s) > 2 { "suspicious" } else { "ok" })
}
```

```text
ok
ok
suspicious
```

Reject a name that starts with a mark, which has no letter to attach to:

```ant
func startsWithMark(s: string): bool {
  let first = s.chars().first()
  first != null && first.isMark()
}

startsWithMark("\u{301}ant") // true
startsWithMark("ant")        // false
startsWithMark("")           // false
```

## Properties and relations

`String.isMark` is `true` exactly when the string is not empty and every code point has a category
that starts with `M`:

```ant
for s in ["", "\u{301}", "\u{903}", "\u{20DD}", "é", "e\u{301}", "´", "\u{200D}"] {
  let chars = s.chars()
  let all = !chars.isEmpty() && chars.every({ x.category().startsWith("M") })
  assert(String.isMark(s) == all)
}
```

A mark is graphic and printable, but not a letter:

```ant
for c in ["\u{301}", "\u{903}", "\u{20DD}", "\u{64E}"] {
  assert(String.isMark(c))
  assert(String.isGraphic(c))
  assert(String.isPrint(c))
  assert(!String.isLetter(c))
}
```

NFD leaves the letter in front and puts its marks after it:

```ant
for s in ["é", "ṩ", "Å", "ü"] {
  let chars = s.normalize("NFD").chars()
  assert(!String.isMark(chars[0]))
  assert(chars.slice(1, chars.length()).every(String.isMark))
}
```

## Possible issues

The same text can be written two ways. A precomposed `é` holds no mark, and `e` followed by U+0301
holds one, though both look the same:

```ant
String.isMark("\u{E9}")                               // false
"\u{E9}".chars().some({ x.isMark() })                 // false
"e\u{301}".chars().some({ x.isMark() })               // true
"\u{E9}".normalize("NFD").chars().some(String.isMark) // true
```

Normalize to NFD first when you look for marks in text that can arrive in either form.

`String.isMark` tests that the whole string is marks, not that it holds one. A letter with its
accent is `false`; use `Array.some` to find out whether a string contains a mark:

```ant
String.isMark("e\u{301}")               // false
"e\u{301}".chars().some({ x.isMark() }) // true
```

Some invisible code points that go with emoji are marks and some are not. The variation selector
U+FE0F is a nonspacing mark, while the zero-width joiner U+200D is a format character and a skin
tone modifier is a symbol:

```ant
String.isMark("\u{FE0F}")  // true
String.isMark("\u{200D}")  // false
String.isMark("\u{1F3FB}") // false
```

So dropping marks to strip accents also drops the variation selectors of emoji. The `stripAccents`
function from Applications turns the emoji heart `❤` U+FE0F into the plain `❤`, which can render as
text instead of emoji:

```ant
func stripAccents(s: string): string {
  s
    | String.normalize("NFD")
    | String.chars()
    | Array.filter({ !x.isMark() })
    | String.concat()
    | String.normalize("NFC")
}

stripAccents("\u{2764}\u{FE0F}") == "\u{2764}" // true
```

Work on graphemes instead, and drop the marks of a grapheme only when it starts with a letter:

```ant
func stripAccents(s: string): string {
  var out = ""
  for g in s.normalize("NFD").graphemes() {
    let chars = g.chars()
    out += if chars[0].isLetter() { String.concat(chars.filter({ !x.isMark() })) } else { g }
  }
  out.normalize("NFC")
}

assert(stripAccents("Crème") == "Creme")
assert(stripAccents("\u{2764}\u{FE0F}") == "\u{2764}\u{FE0F}")
```

## See also

[String.category](category.md) · [String.normalize](normalize.md) · [String.graphemes](graphemes.md)
· [String.chars](chars.md) · [String.isLetter](isLetter.md) · [String.isSymbol](isSymbol.md) ·
[String.isGraphic](isGraphic.md) · [Array.some](../Array/some.md)
