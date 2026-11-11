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

- `script` is a full name from Unicode's `Scripts.txt`, such as `"Latin"`, `"Han"`, or
  `"Old_Italic"`. A short code such as `"Latn"`, another case such as `"latin"`, or an unknown name
  panics.
- `"Common"` holds characters many scripts share: digits, punctuation, symbols, emoji. `"Inherited"`
  holds shared marks, such as U+0301.
- It reads Script, not Script_Extensions, so `ー`, used by both Katakana and Hiragana, is in
  `"Common"`.
- `c` must be exactly one code point; anything else, including `""`, panics.

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
"α".inScript("Greek")     // true
"ا".inScript("Arabic")    // true
"あ".inScript("Hiragana") // true
"ｱ".inScript("Katakana")  // true
"ǅ".inScript("Latin")     // true
```

A script whose name has two words takes an underscore, as in `Scripts.txt`:

```ant
"𐌀".inScript("Old_Italic") // true
```

Digits, punctuation, white space, and emoji are in `Common`, whatever text they appear in:

```ant
"7".inScript("Common")  // true
"１".inScript("Common") // true
"!".inScript("Common")  // true
" ".inScript("Common")  // true
"🐜".inScript("Common") // true
"7".inScript("Latin")   // false
```

A combining mark that many scripts use, such as an accent, is in `Inherited`, not in the script of
the letter it follows. A mark of one script is in that script:

```ant
"\u{301}".inScript("Inherited")  // true
"\u{301}".inScript("Latin")      // false
"\u{3099}".inScript("Inherited") // true
"\u{93F}".inScript("Devanagari") // true
```

An unassigned code point and a private-use one are in `Unknown`:

```ant
"\u{378}".inScript("Unknown")  // true
"\u{E000}".inScript("Unknown") // true
"\u{378}".inScript("Latin")    // false
```

Pass the arguments by name:

```ant
String.inScript(c: "b", script: "Latin") // true
```

Test each code point of a string, here in a pipe:

```ant
"Ab1" | String.chars() | Array.map({ x.inScript("Latin") }) // [true, true, false]
```

## Applications

Find a host name that mixes Latin and Cyrillic letters, a common trick for a look-alike domain:

```ant
func mixesScripts(host: string): bool {
  var latin = false
  var cyrillic = false
  for c in host {
    if c.inScript("Latin") {
      latin = true
    }
    if c.inScript("Cyrillic") {
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
  let found: Array[string] = []
  for c in s {
    for name in scripts {
      if c.inScript(name) && !found.contains(name) {
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
String.concat("漢字 kanji かんじ".chars().filter({ x.inScript("Han") })) // "漢字"
```

## Properties and relations

Script has one value per code point, so each code point below is in exactly one of the listed
scripts:

```ant
let names = ["Latin", "Cyrillic", "Han", "Common", "Inherited", "Unknown"]
for c in ["a", "я", "中", "7", "\u{301}", "\u{378}"] {
  assert(names.filter({ c.inScript(x) }).length() == 1)
}
```

An unassigned code point, with category `"Cn"`, is in `Unknown`:

```ant
for c in ["\u{378}", String.fromCodePoint(0xFFFF) ?? ""] {
  assert(c.category() == "Cn")
  assert(c.inScript("Unknown"))
}
```

## Possible issues

The script name is exact. Another case or a four-letter code panics:

```ant
"a".inScript("latin") // panic: String.inScript needs the name of a script, got "latin"
"a".inScript("Latn")  // panic: String.inScript needs the name of a script, got "Latn"
```

A space in place of the underscore panics too:

```ant
"a".inScript("Old Italic") // panic: String.inScript needs the name of a script, got "Old Italic"
```

`c` must be one code point. A word, the empty string, and a flag, which is two code points, panic:

```ant
"ab".inScript("Latin")  // panic: String.inScript needs a string of one code point, got "ab"
"".inScript("Latin")    // panic: String.inScript needs a string of one code point, got ""
"🇨🇭".inScript("Common") // panic: String.inScript needs a string of one code point, got "🇨🇭"
```

A letter with an accent written as two code points, `e` and U+0301, looks like one character and
panics too:

```ant
"e\u{301}".inScript("Latin") // panic: String.inScript needs a string of one code point, got "é"
```

Compose it first with `String.normalize(s, "NFC")`, or test each code point with `for c in s`:

```ant
"e\u{301}".normalize("NFC").inScript("Latin") // true
```

A script that a later Unicode version adds panics until Ant's tables move to that version.
`String.unicodeVersion` names the version they follow:

```ant
String.unicodeVersion // "17.0.0"
```

Test a whole string with `for c in s` or `String.chars`, and decide what the characters in `Common`
and `Inherited` count as:

```ant
"Привет!".chars().every({ x.inScript("Cyrillic") })                         // false
"Привет!".chars().every({ x.inScript("Cyrillic") || x.inScript("Common") }) // true
```

The test reads Script, not Script_Extensions. The Japanese long-vowel mark `ー` appears in Katakana
and Hiragana words, yet its script is `Common`:

```ant
"ー".inScript("Katakana") // false
"ー".inScript("Common")   // true
```

## See also

[String.category](category.md) · [String.isLetter](isLetter.md) · [String.chars](chars.md) ·
[String.codePoint](codePoint.md) · [String.normalize](normalize.md)
