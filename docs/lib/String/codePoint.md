<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.codePoint

```ant signature
func String.codePoint(c: string): int
```

String.codePoint returns the scalar value of c. It panics unless c is one code point.

<!-- end generated -->

## Details

- Panics unless `c` holds exactly one code point: `""`, a longer string, and a character written
  with a combining accent all panic.
- The result is at most `0x10FFFF` and never a surrogate.

## Basic examples

```ant
String.codePoint("A")  // 65
String.codePoint("é")  // 233
String.codePoint("🐜") // 128028
```

Call it through `.` on a string:

```ant
"a".codePoint() // 97
```

## Scope

Control characters and escapes are code points like any other:

```ant
String.codePoint("\n")         // 10
String.codePoint("\t")         // 9
String.codePoint("\u{0}")      // 0
String.codePoint("\u{10FFFF}") // 1114111
```

A code point outside the Basic Multilingual Plane is still one code point. `String.fromInt` with
base 16 shows it in hexadecimal:

```ant
String.codePoint("🐜") | String.fromInt(16) // "1f41c"
```

The replacement character U+FFFD, which stands in for ill-formed UTF-8 that reached the program from
outside, is an ordinary code point:

```ant
String.codePoint("\u{FFFD}") // 65533
```

The parameter is named `c`:

```ant
String.codePoint(c: "Z") // 90
```

In a pipe, the string is the first argument:

```ant
"A" | String.codePoint() // 65
```

A function value works wherever a `func(string): int` is expected:

```ant
var code: func(string): int = String.codePoint
code("0")                                            // 48
"ant" | String.chars() | Array.map(String.codePoint) // [97, 110, 116]
```

## Applications

Turn a decimal digit into its value by distance from `"0"`:

```ant
func digit(c: string): int {
  String.codePoint(c) - String.codePoint("0")
}

"2026" | String.chars() | Array.map(digit) // [2, 0, 2, 6]
```

Shift lower-case Latin letters for a Caesar cipher, and leave everything else as it is:

```ant
func shift(c: string, k: int): string {
  var n = String.codePoint(c)
  var a = String.codePoint("a")
  if n < a || n > String.codePoint("z") {
    return c
  }
  String.fromCodePoint((n - a + k) % 26 + a) ?? c
}

func caesar(s: string, k: int): string {
  String.join(s.chars().map({ shift(x, k) }), "")
}

println(caesar("hello, ant", 3))
```

```text
khoor, dqw
```

Print each character of a string with its `U+` code:

```ant
for c in "Añ🐜" {
  var hex = String.padStart(String.toUpper(String.fromInt(String.codePoint(c), 16)), 4, "0")
  println("${c} U+${hex}")
}
```

```text
A U+0041
ñ U+00F1
🐜 U+1F41C
```

## Properties and relations

`String.codePoint` of each character agrees with `String.toCodePoints` of the whole string, and
`String.fromCodePoint` turns the value back into the character:

```ant
var s = "Ant 🐜 é"
assert(s.chars().map(String.codePoint) == s.toCodePoints())
for c in s {
  assert(String.fromCodePoint(String.codePoint(c)) == c)
}
```

For a scalar value `n`, `String.fromCodePoint` gives a character whose code point is `n` again:

```ant
for n in [0, 65, 233, 0xD7FF, 0xE000, 0x1F41C, 0x10FFFF] {
  var c = String.fromCodePoint(n)
  assert(c != null && String.codePoint(c) == n)
}
```

## Possible issues

The string must hold exactly one code point. The empty string and a longer string panic:

```ant
String.codePoint("")   // panic: String.codePoint needs a string of one code point, got ""
String.codePoint("ab") // panic: String.codePoint needs a string of one code point, got "ab"
```

What a reader sees as one character can be several code points. An emoji with a skin tone is two:

```ant
String.codePoint("👍🏽") // panic: String.codePoint needs a string of one code point, got "👍🏽"
```

A letter with a combining accent is two code points as well. `String.normalize` with `"NFC"`
composes it into one when Unicode has a single code point for it:

```ant
var e = "e\u{301}"
String.length(e)                       // 2
String.normalize(e, "NFC").codePoint() // 233
```

The code point is not the UTF-8 encoding. Outside ASCII the two differ: `"é"` is code point 233 but
two bytes. [`String.toBytes`](toBytes.md) gives the bytes:

```ant
String.codePoint("é") // 233
String.toBytes("é")   // Bytes(c3 a9)
```

To get the code points of a longer string, use [`String.toCodePoints`](toCodePoints.md), or split it
with [`String.chars`](chars.md) first:

```ant
String.toCodePoints("ab") // [97, 98]
```

## See also

[String.fromCodePoint](fromCodePoint.md) · [String.toCodePoints](toCodePoints.md) ·
[String.fromCodePoints](fromCodePoints.md) · [String.chars](chars.md) ·
[String.graphemes](graphemes.md) · [String.category](category.md) · [String.normalize](normalize.md)
· [String.toBytes](toBytes.md)
