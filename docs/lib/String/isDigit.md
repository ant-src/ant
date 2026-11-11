<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.isDigit

```ant signature
func String.isDigit(s: string): bool
```

String.isDigit reports whether every code point of s is a decimal digit (Nd) in any script. It is
false on "".

<!-- end generated -->

## Details

- `true` when every code point of `s` has the general category Nd; `false` on `""`.
- Digits of any script count: Arabic-Indic `٣`, Devanagari `४`, full-width `１`.
  [`String.toInt`](toInt.md) reads only ASCII digits and letters, so it can still return `null` on a
  string that passes.

## Basic examples

```ant
String.isDigit("2026") // true
String.isDigit("7")    // true
String.isDigit("20x6") // false
```

Call it through `.`:

```ant
"404".isDigit() // true
```

## Scope

The empty string has no code points to test, and gives `false`:

```ant
"".isDigit() // false
```

A sign, a decimal point, a digit separator, and white space are not digits:

```ant
"-1".isDigit()    // false
"1.5".isDigit()   // false
"1_000".isDigit() // false
" 42".isDigit()   // false
```

Decimal digits of other scripts count: Arabic-Indic, Devanagari, full-width, and mathematical bold:

```ant
"٣٤".isDigit()   // true
"४२".isDigit()   // true
"１２".isDigit() // true
"𝟗".isDigit()    // true
```

Numbers that are not decimal digits do not: a fraction, a superscript, and a Roman numeral. Hex
letters are letters:

```ant
"½".isDigit()  // false
"²".isDigit()  // false
"Ⅻ".isDigit()  // false
"ff".isDigit() // false
```

Test one code point at a time with `String.chars`, here in a pipe:

```ant
"a1b2" | String.chars() | Array.map({ x.isDigit() }) // [false, true, false, true]
```

Pass it to a higher-order function as a function value:

```ant
["12", "x", "", "٣"].filter(String.isDigit) // ["12", "٣"]
```

## Applications

Keep only the digits of a phone number:

```ant
let phone = "+1 (555) 010-9999"
String.concat(phone.chars().filter({ x.isDigit() })) // "15550109999"
```

Count the digits in a text:

```ant
"Room 12, floor 3".chars().countWhere({ x.isDigit() }) // 3
```

Split a line into words and numbers:

```ant
type Word = { text: string }
type Num = { value: int }
type Token = Word | Num

func token(field: string): Token {
  let n = if field.isDigit() { field.toInt() } else { null }
  if n != null { Num{ value: n } } else { Word{ text: field } }
}

for t in "buy 3 apples and 12 pears".fields().map(token) {
  if t is Num {
    println("number ${t.value}")
  } else {
    println("word ${t.text}")
  }
}
```

```text
word buy
number 3
word apples
word and
number 12
word pears
```

## Properties and relations

`String.isDigit` is `true` exactly when the string is not empty and every code point has the
category `"Nd"`:

```ant
for s in ["", "0", "42", "٣", "½", "1.5", "a1", "１２"] {
  let byCategory = s != "" && s.chars().every({ x.category() == "Nd" })
  assert(s.isDigit() == byCategory)
}
```

A decimal digit is a number, so `String.isDigit` implies [`String.isNumber`](isNumber.md), but not
the other way around:

```ant
for s in ["7", "٣", "½", "Ⅻ", "²"] {
  assert(!s.isDigit() || s.isNumber())
}
"½".isNumber() // true
"½".isDigit()  // false
```

Two nonempty strings of digits join into a string of digits:

```ant
for a in ["1", "٣", "x"] {
  for b in ["2", "४", ""] {
    if a != "" && b != "" {
      assert(String.isDigit(a + b) == (a.isDigit() && b.isDigit()))
    }
  }
}
```

## Possible issues

`String.isDigit` does not mean `String.toInt` succeeds. Digits of other scripts pass the test, but
`String.toInt` reads only the ASCII digits `0` to `9`, and in a base above ten the letters `a` to
`z`:

```ant
"٣".isDigit() // true
"٣".toInt()   // null
```

To accept only the ASCII digits `0` to `9`, as for a PIN or a ZIP code, compare each code point:

```ant
"12٣".chars().every({ x >= "0" && x <= "9" })  // false
"1234".chars().every({ x >= "0" && x <= "9" }) // true
```

A string of ASCII digits too large for an `int` passes the test too, and `String.toInt` gives
`null`:

```ant
"99999999999999999999".isDigit() // true
"99999999999999999999".toInt()   // null
```

To parse a number, call `String.toInt` and check for `null` instead of testing first.

A negative number is not all digits:

```ant
"-5".isDigit() // false
"-5".toInt()   // -5
```

Every code point must be a digit, so `String.isDigit` does not answer whether a string contains a
digit. Use `Array.some` for that:

```ant
"a1".isDigit()                     // false
"a1".chars().some({ x.isDigit() }) // true
```

## See also

[String.isNumber](isNumber.md) · [String.isLetter](isLetter.md) · [String.isSpace](isSpace.md) ·
[String.category](category.md) · [String.toInt](toInt.md) · [String.toUint](toUint.md) ·
[String.chars](chars.md) · [Array.every](../Array/every.md) · [Array.some](../Array/some.md)
