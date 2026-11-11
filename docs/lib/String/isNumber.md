<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.isNumber

```ant signature
func String.isNumber(s: string): bool
```

String.isNumber reports whether every code point of s is a number (N): Nd, Nl such as Ⅻ, or No such
as ½. It is false on "".

<!-- end generated -->

## Details

- `true` when every code point of `s` has a general category in N (Nd, Nl, No); `false` on `""`.
- Besides decimal digits it accepts letter numbers such as `Ⅻ` and other numbers such as `½`, `²`,
  and `①`.
- It tests characters, not syntax: a sign or a decimal point is not a number, and a string that
  passes need not be one [`String.toInt`](toInt.md) reads.

## Basic examples

```ant
String.isNumber("42") // true
String.isNumber("½")  // true
String.isNumber("4a") // false
```

Call it through `.`:

```ant
"Ⅻ".isNumber() // true
```

## Scope

The empty string has no code points to test, and gives `false`:

```ant
String.isNumber("") // false
```

Decimal digits of any script are numbers:

```ant
String.isNumber("2026") // true
String.isNumber("٣٤")   // true
String.isNumber("１２") // true
```

Letter numbers count: Roman numerals, upper and lower case, and the ideographic zero:

```ant
String.isNumber("Ⅻ")  // true
String.isNumber("ⅻ")  // true
String.isNumber("〇") // true
```

Other numbers count: fractions, superscripts, and circled numbers. The string may mix the three
categories:

```ant
String.isNumber("¾")  // true
String.isNumber("²")  // true
String.isNumber("①")  // true
String.isNumber("3½") // true
String.isNumber("2²") // true
```

A Han numeral such as `三` is a letter, Lo, and `π` is a lower-case letter, so neither is a number:

```ant
String.isNumber("三") // false
String.isNumber("π")  // false
```

Signs, points, separators, exponents, and white space are not number characters:

```ant
String.isNumber("-1")    // false
String.isNumber("1.5")   // false
String.isNumber("1_000") // false
String.isNumber("1e3")   // false
String.isNumber(" 42")   // false
```

Test one code point at a time with `String.chars`, here in a pipe:

```ant
"x½2" | String.chars() | Array.map({ x.isNumber() }) // [false, true, true]
```

Pass it to a higher-order function as a function value:

```ant
["12", "Ⅻ", "", "ab", "¼"].filter(String.isNumber) // ["12", "Ⅻ", "¼"]
```

## Applications

Count the number characters in a recipe line:

```ant
"Add 1½ cups and 2 eggs".chars().countWhere({ x.isNumber() }) // 3
```

Split an amount from its unit, where the amount may hold a fraction:

```ant
type Amount = { qty: string, unit: string }

func amount(s: string): Amount {
  let cs = s.chars()
  let i = cs.findIndex({ !x.isNumber() }) ?? cs.length()
  let qty = String.concat(cs.slice(0, i))
  let unit = String.concat(cs.slice(i, cs.length())).trim()
  Amount{ qty, unit }
}

for s in ["1½ cups", "250 g", "¾ tsp", "pinch"] {
  let a = amount(s)
  println("${a.qty}|${a.unit}")
}
```

```text
1½|cups
250|g
¾|tsp
|pinch
```

Mark the fields of a line that are numbers of any kind:

```ant
for f in "Chapter Ⅻ page 42 note ①".fields() {
  println(if f.isNumber() { "[${f}]" } else { f })
}
```

```text
Chapter
[Ⅻ]
page
[42]
note
[①]
```

## Properties and relations

`String.isNumber` is `true` exactly when the string is not empty and the category of every code
point starts with `N`:

```ant
for s in ["", "0", "42", "٣", "½", "Ⅻ", "²", "1.5", "a1", "三"] {
  let byCategory = s != "" && s.chars().every({ x.category().startsWith("N") })
  assert(String.isNumber(s) == byCategory)
}
```

A decimal digit is a number, so [`String.isDigit`](isDigit.md) implies `String.isNumber`, but not
the other way around:

```ant
for s in ["7", "٣", "½", "Ⅻ", "²", "x"] {
  assert(!String.isDigit(s) || String.isNumber(s))
}
String.isDigit("Ⅻ")  // false
String.isNumber("Ⅻ") // true
```

A number character is not a letter, even a Roman numeral that looks like one:

```ant
for c in ["7", "Ⅻ", "½", "〇"] {
  assert(String.isNumber(c) && !String.isLetter(c))
}
String.isLetter("X") // true
String.isNumber("X") // false
```

## Possible issues

`String.isNumber` does not mean the string parses as a number. `String.toInt` reads only the ASCII
digits, so a Roman numeral, a fraction, or a digit of another script passes the test and still gives
`null`:

```ant
String.isNumber("Ⅻ") // true
String.toInt("Ⅻ")    // null
String.toInt("½")    // null
String.toInt("٣")    // null
```

To parse a number, call `String.toInt` or `String.toFloat` and check for `null` instead of testing
first.

A number that `String.toFloat` reads can fail the test: a sign and a decimal point are not number
characters:

```ant
String.isNumber("-2.5") // false
String.toFloat("-2.5")  // -2.5
```

A Roman numeral written with ordinary letters is letters. Only the Roman numeral code points, such
as U+216B `Ⅻ`, are numbers:

```ant
String.isNumber("XII")      // false
String.isNumber("\u{216B}") // true
```

A combining mark such as the keycap sign is not a number, so a digit with a mark attached fails the
test:

```ant
String.isNumber("1\u{20E3}") // false
String.category("\u{20E3}")  // "Me"
```

Every code point must be a number, so `String.isNumber` does not answer whether a string contains
one. Use `Array.some` for that:

```ant
String.isNumber("No. 5")               // false
"No. 5".chars().some({ x.isNumber() }) // true
```

## See also

[String.isDigit](isDigit.md) · [String.isLetter](isLetter.md) · [String.category](category.md) ·
[String.toInt](toInt.md) · [String.toFloat](toFloat.md) · [String.chars](chars.md) ·
[Array.every](../Array/every.md) · [Array.some](../Array/some.md)
