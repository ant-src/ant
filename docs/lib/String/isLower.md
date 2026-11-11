<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.isLower

```ant signature
func String.isLower(s: string): bool
```

String.isLower reports whether every code point of s is a lower-case letter (Ll). It is false on "".

<!-- end generated -->

## Details

- `true` when every code point of `s` has the general category Ll; `false` on `""`.
- Only Ll counts: a modifier letter such as `ʰ`, a title-case letter such as `ǅ`, and a combining
  accent are not lower-case letters.

## Basic examples

```ant
String.isLower("ant") // true
String.isLower("Ant") // false
String.isLower("x")   // true
```

Call it through `.`:

```ant
"colony".isLower() // true
```

## Scope

The empty string has no code points to test, and gives `false`:

```ant
"".isLower() // false
```

Every code point must be a lower-case letter, so a space, a digit, a hyphen, or an underscore makes
the result `false`:

```ant
"hello world".isLower() // false
"ab1".isLower()         // false
"well-known".isLower()  // false
"snake_case".isLower()  // false
```

Lower-case letters of every script count, with their accents when the accent is part of the code
point:

```ant
"été".isLower()    // true
"straße".isLower() // true
"π".isLower()      // true
"жук".isLower()    // true
"ﬁ".isLower()      // true
```

Han characters such as `日` are letters of category Lo, not lower-case letters:

```ant
"日本".isLower() // false
"日".category()  // "Lo"
```

The title-case letter `ǅ` is neither lower nor upper case, and the modifier letter `ʰ` is not a
lower-case letter either:

```ant
"ǅ".isLower()  // false
"ǅ".isUpper()  // false
"ǅ".category() // "Lt"
"ʰ".isLower()  // false
"ʰ".category() // "Lm"
```

Test one code point at a time with `String.chars`, here in a pipe:

```ant
"aBc" | String.chars() | Array.map({ x.isLower() }) // [true, false, true]
```

Pass it to a higher-order function as a function value:

```ant
["id", "Name", "", "x2", "età"].filter(String.isLower) // ["id", "età"]
```

## Applications

Split a camelCase name into words before each code point that is not a lower-case letter:

```ant
func words(name: string): Array[string] {
  let out: Array[string] = []
  var word = ""
  for c in name {
    if word != "" && !c.isLower() {
      out.push(word)
      word = ""
    }
    word += c.toLower()
  }
  if word != "" {
    out.push(word)
  }
  out
}

words("parseHttpRequest") // ["parse", "http", "request"]
```

Count the lower-case letters of a password:

```ant
"Tr0ub4dor&3".chars().countWhere({ x.isLower() }) // 6
```

Accept a package name only when it is all lower-case letters:

```ant
for name in ["json", "Json", "json5", "время"] {
  if name.isLower() {
    println("${name}: ok")
  } else {
    println("${name}: use lower-case letters only")
  }
}
```

```text
json: ok
Json: use lower-case letters only
json5: use lower-case letters only
время: ok
```

## Properties and relations

`String.isLower` is `true` exactly when the string is not empty and every code point has the
category `"Ll"`:

```ant
for s in ["", "a", "abc", "aB", "ǅ", "ʰ", "π", "a1", "日"] {
  let byCategory = s != "" && s.chars().every({ x.category() == "Ll" })
  assert(s.isLower() == byCategory)
}
```

A lower-case letter is a letter, so `String.isLower` implies [`String.isLetter`](isLetter.md), and
no string is both lower and upper case:

```ant
for s in ["a", "A", "ǅ", "日", "aB", "π", "Ω"] {
  assert(!s.isLower() || s.isLetter())
  assert(!(s.isLower() && s.isUpper()))
}
```

## Possible issues

`String.isLower` does not ask whether `String.toLower` would leave the string unchanged. A string
with a digit or a space is already in lower case, but is not all lower-case letters:

```ant
"ab 1".toLower() == "ab 1" // true
"ab 1".isLower()           // false
```

To check that `String.toLower` leaves a string unchanged, compare `s == s.toLower()`. To check that
no code point is an upper-case letter, use `!s.chars().some({ x.isUpper() })`:

```ant
!"ab 1".chars().some({ x.isUpper() }) // true
```

The same accented letter can be one code point or a letter followed by a combining accent. The
combining accent is a mark, Mn, so the second spelling is not all lower-case letters. Normalize to
NFC first:

```ant
let decomposed = "été".normalize("NFD")
decomposed.isLower()                  // false
decomposed.normalize("NFC").isLower() // true
```

A lower-case letter can have an upper-case form longer than itself. `ß` is lower case, and
`String.toUpper` turns it into two letters:

```ant
"ß".isLower() // true
"ß".toUpper() // "SS"
```

`String.isLower` tests that the whole string is lower-case letters, not that it holds one. Use
`Array.some` for that:

```ant
"ABc".isLower()                     // false
"ABc".chars().some({ x.isLower() }) // true
```

## See also

[String.isUpper](isUpper.md) · [String.isLetter](isLetter.md) · [String.toLower](toLower.md) ·
[String.toUpper](toUpper.md) · [String.category](category.md) · [String.normalize](normalize.md) ·
[String.chars](chars.md) · [Array.every](../Array/every.md) · [Array.some](../Array/some.md)
