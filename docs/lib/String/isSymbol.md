<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.isSymbol

```ant signature
func String.isSymbol(s: string): bool
```

String.isSymbol reports whether every code point of s is a symbol (S). It is false on "".

<!-- end generated -->

## Details

- `true` when every code point of `s` has a general category in S (`Sm`, `Sc`, `Sk`, `So`); `false`
  on `""`.
- Many ASCII signs, such as `%`, `#`, `&`, `*`, `@`, and `-`, are punctuation, not symbols.

## Basic examples

```ant
String.isSymbol("+") // true
String.isSymbol("$") // true
String.isSymbol("©") // true
String.isSymbol("a") // false
```

Call it through `.`:

```ant
"<=>".isSymbol() // true
```

## Scope

The empty string has no code points, so it is not all symbols:

```ant
String.isSymbol("") // false
```

One code point that is not a symbol makes the result `false`, wherever it is in the string:

```ant
String.isSymbol("$5")  // false
String.isSymbol("+++") // true
String.isSymbol("1+1") // false
```

Each of the four kinds of symbol counts. `String.category` names the kind:

```ant
String.isSymbol("√") // true
String.category("√") // "Sm"
String.isSymbol("€") // true
String.category("€") // "Sc"
String.isSymbol("^") // true
String.category("^") // "Sk"
String.isSymbol("°") // true
String.category("°") // "So"
```

The ASCII symbols are `$`, `+`, `<`, `=`, `>`, `^`, `` ` ``, `|`, and `~`:

```ant
"$+<=>^`|~".chars().every({ String.isSymbol(x) }) // true
```

An emoji of one code point is a symbol, and so is a flag, which is two regional indicator symbols. A
skin tone modifier is a modifier symbol, so a thumbs-up with a skin tone is all symbols too:

```ant
String.isSymbol("😀") // true
String.isSymbol("🇫🇷") // true
String.isSymbol("👍🏽") // true
```

Letters, digits, other numbers, and white space are not symbols:

```ant
String.isSymbol("x") // false
String.isSymbol("7") // false
String.isSymbol("½") // false
String.isSymbol(" ") // false
```

Test each code point of a string with `String.chars` in a pipe:

```ant
"2+2=4" | String.chars() | Array.map({ String.isSymbol(x) }) // [false, true, false, true, false]
```

## Applications

Split a price into its currency sign and its amount. The sign is a symbol; the digits are not:

```ant
type Price = { currency: string, amount: string }

func parsePrice(s: string): Price | null {
  var chars = s.chars()
  if chars.isEmpty() || !String.isSymbol(chars[0]) {
    return null
  }
  Price{ currency: chars[0], amount: s.trimPrefix(chars[0]) }
}

parsePrice("€25") // Price{ currency: "€", amount: "25" }
parsePrice("25")  // null
```

Tell an operator token from a name in a small expression scanner. Operators mix symbols such as `+`
and `<` with punctuation such as `-` and `&`, so test each code point for either:

```ant
func isOperator(token: string): bool {
  !token.isEmpty() && token.chars().every({ String.isSymbol(x) || String.isPunct(x) })
}

for token in ["x", "+", "-", "<=", "&&", "count", "|>"] {
  if isOperator(token) {
    println("operator ${token}")
  } else {
    println("name ${token}")
  }
}
```

```text
name x
operator +
operator -
operator <=
operator &&
name count
operator |>
```

Count the symbols of a text:

```ant
"1 + 2 = 3 ≠ 4 ✓".chars().countWhere({ String.isSymbol(x) }) // 4
```

## Properties and relations

`String.isSymbol(s)` is `s` not empty and every code point in a category that starts with `S`:

```ant
for s in ["", "+", "$5", "😀", "©®", "%", "a", "👍🏽"] {
  var chars = String.chars(s)
  var all = !chars.isEmpty() && chars.every({ String.category(x).startsWith("S") })
  assert(String.isSymbol(s) == all)
}
```

A code point is never both a symbol and punctuation, a letter, or a number. Every symbol is graphic
and printable:

```ant
for c in ["+", "$", "^", "©", "😀", "%", "a", "1"] {
  if String.isSymbol(c) {
    assert(!String.isPunct(c) && !String.isLetter(c) && !String.isNumber(c))
    assert(String.isGraphic(c) && String.isPrint(c))
  }
}
```

## Possible issues

Many ASCII signs are punctuation, not symbols. To accept both, test both on each code point. Testing
the whole string with each function fails on a string that mixes the two, such as `!=`:

```ant
String.isSymbol("%")                                            // false
String.isPunct("%")                                             // true
String.isSymbol("-")                                            // false
String.isSymbol("!=") || String.isPunct("!=")                   // false
"!=".chars().every({ String.isSymbol(x) || String.isPunct(x) }) // true
```

An emoji that looks like one character can hold code points that are not symbols. The red heart `❤️`
ends with the variation selector U+FE0F, a mark, and a family emoji joins its people with the
zero-width joiner U+200D, a format character:

```ant
String.isSymbol("❤️")                    // false
"❤️".chars().map({ String.category(x) }) // ["So", "Mn"]
String.isSymbol("👨‍👩‍👧")                    // false
```

An emoji that a Unicode version newer than `String.unicodeVersion` adds is unassigned, category
`"Cn"`, not `"So"`, so `String.isSymbol` is `false` on it until the implementation moves to that
version.

`String.isSymbol` tests that the whole string is symbols, not that it holds one. To find out whether
a string contains a symbol, test each code point:

```ant
String.isSymbol("a+b")                     // false
"a+b".chars().some({ String.isSymbol(x) }) // true
```

## See also

[String.category](category.md) · [String.isPunct](isPunct.md) · [String.isMark](isMark.md) ·
[String.isGraphic](isGraphic.md) · [String.isPrint](isPrint.md) · [String.chars](chars.md) ·
[String.graphemes](graphemes.md)
