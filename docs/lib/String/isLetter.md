<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.isLetter

```ant signature
func String.isLetter(s: string): bool
```

String.isLetter reports whether every code point of s is a letter (L). It is false on "".

<!-- end generated -->

## Details

- `String.isLetter` is `true` when `s` is not empty and every code point of `s` has a Unicode
  general category in L, letter: `Lu`, `Ll`, `Lt`, `Lm`, or `Lo`.
- A letter can come from any script: Latin, Cyrillic, Greek, Arabic, the Han ideographs, and the
  rest of L.
- Ant has no character type, so `s` is a `string`. On one code point, such as `c` in `for c in s`,
  it tests a character; on a longer string, it tests that the whole string is letters.
- The test is per code point, not per grapheme. A combining accent is a mark, `M`, not a letter.
- Categories follow the Unicode version `String.unicodeVersion` names; a code point that a later
  version assigns is unassigned, `Cn`, and is not a letter until then.
- [spec: String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)

## Basic examples

```ant
String.isLetter("ant")  // true
String.isLetter("A")    // true
String.isLetter("ant1") // false
```

Call it through `.`:

```ant
"Colony".isLetter() // true
```

## Scope

The empty string has no code points to test, and gives `false`:

```ant
String.isLetter("") // false
```

Every code point must be a letter, so a space, a digit, a hyphen, or an underscore anywhere makes it
`false`:

```ant
String.isLetter("red ant") // false
String.isLetter("ant2")    // false
String.isLetter("re-do")   // false
String.isLetter("my_var")  // false
```

Letters of other scripts count:

```ant
String.isLetter("Привет") // true
String.isLetter("αβγ")    // true
String.isLetter("مرحبا")  // true
String.isLetter("日本語")    // true
String.isLetter("ß")      // true
```

Each kind of letter counts: upper case `Lu`, lower case `Ll`, title case `Lt`, modifier `Lm`, and
other `Lo`:

```ant
String.isLetter("Ǆ") // true
String.isLetter("ǆ") // true
String.isLetter("ǅ") // true
String.isLetter("ʰ") // true
String.isLetter("ア") // true
```

A Roman numeral looks like a letter but is a number, `Nl`. An emoji is a symbol:

```ant
String.isLetter("Ⅻ") // false
String.isLetter("🐜") // false
```

Test one code point at a time with `String.chars`, here in a pipe:

```ant
"a1 b" | String.chars() | Array.map({ String.isLetter(x) }) // [true, false, false, true]
```

Pass it to a higher-order function as a function value:

```ant
["ant", "42", "", "bee"].filter(String.isLetter) // ["ant", "bee"]
```

## Applications

Keep only the letters of a text:

```ant
var code = "A1-B2_c3"
String.concat(code.chars().filter({ String.isLetter(x) })) // "ABc"
```

Count the letters of a sentence:

```ant
"Hi, ants! 3 of you?".chars().countWhere({ String.isLetter(x) }) // 11
```

Split a text into words, where a word is a run of letters:

```ant
func words(text: string): Array[string] {
  var out: Array[string] = []
  var word = ""
  for c in text {
    if String.isLetter(c) {
      word = word + c
    } else if word != "" {
      out.push(word)
      word = ""
    }
  }
  if word != "" {
    out.push(word)
  }
  out
}

words("Ants, bees & wasps: 3 orders?") // ["Ants", "bees", "wasps", "orders"]
```

Check that a name starts with a letter before you accept it:

```ant
type Check = { name: string, ok: bool }

func check(name: string): Check {
  var chars = name.chars()
  Check{ name: name, ok: !chars.isEmpty() && String.isLetter(chars[0]) }
}

for c in ["ant", "_tmp", "9lives", ""].map(check) {
  println("${c.name}: ${c.ok}")
}
```

```text
ant: true
_tmp: false
9lives: false
: false
```

## Properties and relations

`String.isLetter` is `true` exactly when the string is not empty and every code point has a category
that starts with `L`:

```ant
for s in ["", "a", "Ant", "日本", "ǅ", "ʰ", "a1", "Ⅻ", "e\u{301}", "_"] {
  var byCategory = s != "" && s.chars().every({ String.category(x).startsWith("L") })
  assert(String.isLetter(s) == byCategory)
}
```

An upper-case or lower-case letter is a letter, so [`String.isUpper`](isUpper.md) and
[`String.isLower`](isLower.md) imply `String.isLetter`, but not the other way around:

```ant
for s in ["A", "a", "ǅ", "日", "ʰ", "1"] {
  assert(!String.isUpper(s) || String.isLetter(s))
  assert(!String.isLower(s) || String.isLetter(s))
}
String.isLetter("日") // true
String.isUpper("日")  // false
String.isLower("日")  // false
```

For nonempty `a` and `b`, `a + b` is all letters exactly when both `a` and `b` are:

```ant
for a in ["ab", "日", "1"] {
  for b in ["c", "ß", " "] {
    assert(String.isLetter(a + b) == (String.isLetter(a) && String.isLetter(b)))
  }
}
```

## Possible issues

An accented letter written as a base letter and a combining accent is two code points, and the
accent is a mark, not a letter. Normalize to NFC first to compose the pair when Unicode has one code
point for it:

```ant
var e = "e\u{301}"
String.isLetter(e)                          // false
String.isMark("\u{301}")                    // true
String.isLetter(String.normalize(e, "NFC")) // true
```

Real names hold more than letters: an apostrophe, a hyphen, a space. `String.isLetter` rejects them:

```ant
String.isLetter("O'Brien")  // false
String.isLetter("Jean-Luc") // false
String.isLetter("Mary Ann") // false
```

`String.isLetter` tests that the whole string is letters, not that it holds one. Use `Array.some`
for that:

```ant
String.isLetter("a1")                     // false
"a1".chars().some({ String.isLetter(x) }) // true
```

A letter is not only `a` to `z`. To accept ASCII letters only, test the code point:

```ant
func isAsciiLetter(c: string): bool {
  var n = String.codePoint(c)
  (n >= 65 && n <= 90) || (n >= 97 && n <= 122)
}

String.isLetter("é") // true
isAsciiLetter("é")   // false
isAsciiLetter("Z")   // true
```

## See also

[String.isUpper](isUpper.md) · [String.isLower](isLower.md) · [String.isDigit](isDigit.md) ·
[String.isNumber](isNumber.md) · [String.isMark](isMark.md) · [String.category](category.md) ·
[String.inScript](inScript.md) · [String.normalize](normalize.md) · [String.chars](chars.md) ·
[Array.every](../Array/every.md) · [Array.some](../Array/some.md)
