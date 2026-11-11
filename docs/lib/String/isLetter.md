<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.isLetter

```ant signature
func String.isLetter(s: string): bool
```

String.isLetter reports whether every code point of s is a letter (L). It is false on "".

<!-- end generated -->

## Details

- `true` when every code point of `s` has a general category in L; `false` on `""`.
- Letters of any script count. The test is per code point: a combining accent is a mark, not a
  letter.

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
"".isLetter() // false
```

Every code point must be a letter, so a space, a digit, a hyphen, or an underscore anywhere makes it
`false`:

```ant
"red ant".isLetter() // false
"ant2".isLetter()    // false
"re-do".isLetter()   // false
"my_var".isLetter()  // false
```

Letters of other scripts count:

```ant
"Привет".isLetter() // true
"αβγ".isLetter()    // true
"مرحبا".isLetter()  // true
"日本語".isLetter() // true
"ß".isLetter()      // true
```

Each kind of letter counts: upper case `Lu`, lower case `Ll`, title case `Lt`, modifier `Lm`, and
other `Lo`:

```ant
"Ǆ".isLetter()  // true
"ǆ".isLetter()  // true
"ǅ".isLetter()  // true
"ʰ".isLetter()  // true
"ア".isLetter() // true
```

A Roman numeral looks like a letter but is a number, `Nl`. An emoji is a symbol:

```ant
"Ⅻ".isLetter()  // false
"🐜".isLetter() // false
```

Test one code point at a time with `String.chars`, here in a pipe:

```ant
"a1 b" | String.chars() | Array.map({ x.isLetter() }) // [true, false, false, true]
```

Pass it to a higher-order function as a function value:

```ant
["ant", "42", "", "bee"].filter(String.isLetter) // ["ant", "bee"]
```

## Applications

Keep only the letters of a text:

```ant
let code = "A1-B2_c3"
String.concat(code.chars().filter({ x.isLetter() })) // "ABc"
```

Count the letters of a sentence:

```ant
"Hi, ants! 3 of you?".chars().countWhere({ x.isLetter() }) // 11
```

Split a text into words, where a word is a run of letters:

```ant
func words(text: string): Array[string] {
  let out: Array[string] = []
  var word = ""
  for c in text {
    if c.isLetter() {
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
  let chars = name.chars()
  Check{ name: name, ok: !chars.isEmpty() && chars[0].isLetter() }
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
  let byCategory = s != "" && s.chars().every({ x.category().startsWith("L") })
  assert(s.isLetter() == byCategory)
}
```

An upper-case or lower-case letter is a letter, so [`String.isUpper`](isUpper.md) and
[`String.isLower`](isLower.md) imply `String.isLetter`, but not the other way around:

```ant
for s in ["A", "a", "ǅ", "日", "ʰ", "1"] {
  assert(!s.isUpper() || s.isLetter())
  assert(!s.isLower() || s.isLetter())
}
"日".isLetter() // true
"日".isUpper()  // false
"日".isLower()  // false
```

For nonempty `a` and `b`, `a + b` is all letters exactly when both `a` and `b` are:

```ant
for a in ["ab", "日", "1"] {
  for b in ["c", "ß", " "] {
    assert(String.isLetter(a + b) == (a.isLetter() && b.isLetter()))
  }
}
```

## Possible issues

An accented letter written as a base letter and a combining accent is two code points, and the
accent is a mark, not a letter. Normalize to NFC first to compose the pair when Unicode has one code
point for it:

```ant
let e = "e\u{301}"
e.isLetter()                  // false
"\u{301}".isMark()            // true
e.normalize("NFC").isLetter() // true
```

Real names hold more than letters: an apostrophe, a hyphen, a space. `String.isLetter` rejects them:

```ant
"O'Brien".isLetter()  // false
"Jean-Luc".isLetter() // false
"Mary Ann".isLetter() // false
```

`String.isLetter` tests that the whole string is letters, not that it holds one. Use `Array.some`
for that:

```ant
"a1".isLetter()                     // false
"a1".chars().some({ x.isLetter() }) // true
```

A letter is not only `a` to `z`. To accept ASCII letters only, test the code point:

```ant
func isAsciiLetter(c: string): bool {
  let n = c.codePoint()
  (n >= 65 && n <= 90) || (n >= 97 && n <= 122)
}

"é".isLetter()     // true
isAsciiLetter("é") // false
isAsciiLetter("Z") // true
```

## See also

[String.isUpper](isUpper.md) · [String.isLower](isLower.md) · [String.isDigit](isDigit.md) ·
[String.isNumber](isNumber.md) · [String.isMark](isMark.md) · [String.category](category.md) ·
[String.inScript](inScript.md) · [String.normalize](normalize.md) · [String.chars](chars.md) ·
[Array.every](../Array/every.md) · [Array.some](../Array/some.md)
