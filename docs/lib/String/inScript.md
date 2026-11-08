<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.inScript

```ant signature
func String.inScript(c: string, script: string): bool
```

String.inScript reports whether c has the Unicode Script property script, named in full as in
Scripts.txt, such as "Latin" or "Common". It reads Script, not Script\_Extensions. It panics unless
c is one code point, or if script names no script.

<!-- end generated -->

## Details

- `String.inScript` reports whether the code point `c` has the Unicode Script property `script`.
- `script` is a name from Unicode's `Scripts.txt`, written in full: `"Latin"`, `"Han"`,
  `"Cyrillic"`, `"Old_Italic"`. A short code such as `"Latn"`, a name in another case such as
  `"latin"`, and a script that a later Unicode version adds all panic, the last until Ant's tables
  move to that version; `String.unicodeVersion` names the version they follow.
- `"Common"` holds the characters that many scripts use: digits, punctuation, symbols, emoji.
  `"Inherited"` holds marks that many scripts use, such as the acute accent U+0301, which take the
  script of the letter before them. A mark that belongs to one script, such as the Devanagari vowel
  sign `ि`, is in that script. `"Unknown"` holds every code point that no other script holds, the
  unassigned among them.
- The test reads Script, not Script_Extensions, so a character counts in one script only, even when
  several scripts use it: `ー`, which Katakana and Hiragana both use, is in `Common`.
- `c` must be exactly one code point. The empty string, two code points, and a single grapheme
  written with several code points all panic.
- [spec: String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)

## Basic examples

```ant
String.inScript("a", "Latin")    // true
String.inScript("я", "Cyrillic") // true
String.inScript("я", "Latin")    // false
```

Call it through `.` on a one-code-point string:

```ant
"中".inScript("Han") // true
```

## Scope

Letters of different writing systems:

```ant
String.inScript("α", "Greek")    // true
String.inScript("ا", "Arabic")   // true
String.inScript("あ", "Hiragana") // true
String.inScript("ｱ", "Katakana") // true
String.inScript("ǅ", "Latin")    // true
```

A script whose name has two words takes an underscore, as in `Scripts.txt`:

```ant
String.inScript("𐌀", "Old_Italic") // true
```

Digits, punctuation, white space, and emoji are in `Common`, whatever text they appear in:

```ant
String.inScript("7", "Common") // true
String.inScript("１", "Common") // true
String.inScript("!", "Common") // true
String.inScript(" ", "Common") // true
String.inScript("🐜", "Common") // true
String.inScript("7", "Latin")  // false
```

A combining mark that many scripts use, such as an accent, is in `Inherited`, not in the script of
the letter it follows. A mark of one script is in that script:

```ant
String.inScript("\u{301}", "Inherited")  // true
String.inScript("\u{301}", "Latin")      // false
String.inScript("\u{3099}", "Inherited") // true
String.inScript("\u{93F}", "Devanagari") // true
```

An unassigned code point and a private-use one are in `Unknown`:

```ant
String.inScript("\u{378}", "Unknown")  // true
String.inScript("\u{E000}", "Unknown") // true
String.inScript("\u{378}", "Latin")    // false
```

Pass the arguments by name:

```ant
String.inScript(c: "b", script: "Latin") // true
```

Test each code point of a string, here in a pipe:

```ant
"Ab1" | String.chars() | Array.map({ String.inScript(x, "Latin") }) // [true, true, false]
```

## Applications

Find a host name that mixes Latin and Cyrillic letters, a common trick for a look-alike domain:

```ant
func mixesScripts(host: string): bool {
  var latin = false
  var cyrillic = false
  for c in host {
    if String.inScript(c, "Latin") {
      latin = true
    }
    if String.inScript(c, "Cyrillic") {
      cyrillic = true
    }
  }
  latin && cyrillic
}

println(mixesScripts("paypal.com"))
println(mixesScripts("pаypal.com"))
```

```text
false
true
```

Name the scripts of a text, skipping the characters that every script shares:

```ant
let scripts = ["Latin", "Greek", "Cyrillic", "Han", "Hiragana", "Katakana"]

func scriptsOf(s: string): Array[string] {
  var found: Array[string] = []
  for c in s {
    for name in scripts {
      if String.inScript(c, name) && !found.contains(name) {
        found.push(name)
      }
    }
  }
  found
}

println(scriptsOf("Tokyo 東京 とうきょう"))
println(scriptsOf("α + β = 3"))
```

```text
["Latin", "Han", "Hiragana"]
["Greek"]
```

Keep only the Han characters of a mixed text:

```ant
String.concat("漢字 kanji かんじ".chars().filter({ String.inScript(x, "Han") })) // "漢字"
```

## Properties and relations

Script has one value per code point, so each code point below is in exactly one of the listed
scripts:

```ant
let names = ["Latin", "Cyrillic", "Han", "Common", "Inherited", "Unknown"]
for c in ["a", "я", "中", "7", "\u{301}", "\u{378}"] {
  assert(names.filter({ String.inScript(c, x) }).length() == 1)
}
```

An unassigned code point, with category `"Cn"`, is in `Unknown`:

```ant
for c in ["\u{378}", String.fromCodePoint(0xFFFF) ?? ""] {
  assert(String.category(c) == "Cn")
  assert(String.inScript(c, "Unknown"))
}
```

## Possible issues

The script name is exact. Another case or a four-letter code panics:

```ant
String.inScript("a", "latin") // panic: String.inScript needs the name of a script, got "latin"
String.inScript("a", "Latn")  // panic: String.inScript needs the name of a script, got "Latn"
```

A space in place of the underscore panics too:

```ant
String.inScript("a", "Old Italic") // panic: String.inScript needs the name of a script, got "Old Italic"
```

`c` must be one code point. A word, the empty string, and a flag, which is two code points, panic:

```ant
String.inScript("ab", "Latin")  // panic: String.inScript needs a string of one code point, got "ab"
String.inScript("", "Latin")    // panic: String.inScript needs a string of one code point, got ""
String.inScript("🇨🇭", "Common") // panic: String.inScript needs a string of one code point, got "🇨🇭"
```

A letter with an accent written as two code points, `e` and U+0301, looks like one character and
panics too:

```ant
String.inScript("e\u{301}", "Latin") // panic: String.inScript needs a string of one code point, got "é"
```

Compose it first with `String.normalize(s, "NFC")`, or test each code point with `for c in s`:

```ant
String.inScript(String.normalize("e\u{301}", "NFC"), "Latin") // true
```

A script that a later Unicode version adds panics until Ant's tables move to that version.
`String.unicodeVersion` names the version they follow:

```ant
String.unicodeVersion // "17.0.0"
```

Test a whole string with `for c in s` or `String.chars`, and decide what the characters in `Common`
and `Inherited` count as:

```ant
"Привет!".chars().every({ String.inScript(x, "Cyrillic") })                                 // false
"Привет!".chars().every({ String.inScript(x, "Cyrillic") || String.inScript(x, "Common") }) // true
```

The test reads Script, not Script_Extensions. The Japanese long-vowel mark `ー` appears in Katakana
and Hiragana words, yet its script is `Common`:

```ant
String.inScript("ー", "Katakana") // false
String.inScript("ー", "Common")   // true
```

## See also

[String.category](category.md) · [String.isLetter](isLetter.md) · [String.chars](chars.md) ·
[String.codePoint](codePoint.md) · [String.normalize](normalize.md)
