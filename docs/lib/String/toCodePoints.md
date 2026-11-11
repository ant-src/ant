<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.toCodePoints

```ant signature
func String.toCodePoints(s: string): Array[int]
```

String.toCodePoints returns the scalar value of each code point of s.

<!-- end generated -->

## Details

- Returns one `int` per code point, in order. `""` gives `[]`.
- Each value is from `0` to `0x10FFFF` and never a surrogate.
- It counts code points, as `String.length` does, not bytes and not what a reader sees as one
  character.

## Basic examples

```ant
String.toCodePoints("ant") // [97, 110, 116]
String.toCodePoints("é")   // [233]
```

Call it through `.` on a string:

```ant
"Hi".toCodePoints() // [72, 105]
```

## Scope

The empty string gives an empty array:

```ant
String.toCodePoints("") // []
```

Escapes, control characters, and U+FFFD are code points like any other:

```ant
String.toCodePoints("\t\n")       // [9, 10]
String.toCodePoints("\u{0}")      // [0]
String.toCodePoints("\u{FFFD}")   // [65533]
String.toCodePoints("\u{10FFFF}") // [1114111]
```

A code point outside the Basic Multilingual Plane is one value, though UTF-8 stores it in four
bytes:

```ant
String.toCodePoints("🐜") // [128028]
"🐜".toBytes().length()   // 4
```

What a reader sees as one character can be several code points. A letter with a combining accent is
two, and a family emoji is three people joined by U+200D, the zero-width joiner:

```ant
String.toCodePoints("e\u{301}") // [101, 769]
String.toCodePoints("👨‍👩‍👧")       // [128104, 8205, 128105, 8205, 128103]
```

In a pipe, the string is the first argument:

```ant
"abc" | String.toCodePoints() | Array.map({ x - 96 }) // [1, 2, 3]
```

The argument can be named `s`:

```ant
String.toCodePoints(s: "OK") // [79, 75]
```

A function value works wherever a `func(string): Array[int]` is expected:

```ant
let points: func(string): Array[int] = String.toCodePoints
["a", "bc"].map(points) // [[97], [98, 99]]
```

## Applications

Tell whether a string is plain ASCII:

```ant
func isAscii(s: string): bool {
  s.toCodePoints().every({ x < 128 })
}

isAscii("ant") // true
isAscii("añt") // false
```

Count the UTF-16 code units of a string, which is what JavaScript's `length` reports: a code point
above `0xFFFF` takes two:

```ant
func utf16Length(s: string): int {
  s.toCodePoints().map({ if x > 0xFFFF { 2 } else { 1 } }).sum() ?? 0
}

utf16Length("ant")    // 3
utf16Length("ant 🐜") // 6
```

Print each code point of a string in `U+` notation:

```ant
let codes = "Añ🐜".toCodePoints().map({
  "U+" + String.fromInt(x, 16).toUpper().padStart(4, "0")
})
println(String.join(codes, " "))
```

```text
U+0041 U+00F1 U+1F41C
```

## Properties and relations

[String.fromCodePoints](fromCodePoints.md) undoes it, and the array has one value per code point, as
[String.length](length.md) counts them:

```ant
for s in ["", "ant", "héllo 日本", "e\u{301}", "🐜\u{0}"] {
  let points = s.toCodePoints()
  assert(String.fromCodePoints(points) == s)
  assert(len(points) == s.length())
}
```

It is [String.codePoint](codePoint.md) of each of [String.chars](chars.md), and of each value a
`for` loop yields:

```ant
let s = "Ant 🐜 é"
assert(s.toCodePoints() == s.chars().map(String.codePoint))
let points: Array[int] = []
for c in s {
  points.push(c.codePoint())
}
assert(points == s.toCodePoints())
```

Every value is a scalar value:

```ant
for n in "a\u{0}é日🐜\u{FFFD}\u{10FFFF}".toCodePoints() {
  assert(n >= 0 && n <= 0x10FFFF && (n < 0xD800 || n > 0xDFFF))
}
```

## Possible issues

The values are code points, not bytes. Use [String.toBytes](toBytes.md) for the UTF-8 encoding:

```ant
String.toCodePoints("é") // [233]
"é".toBytes().length()   // 2
```

Two spellings of the same text can give different code points. Bring both to one form with
[String.normalize](normalize.md) before you compare them:

```ant
"\u{E9}".toCodePoints() == "e\u{301}".toCodePoints()                  // false
"e\u{301}".normalize("NFC").toCodePoints() == "\u{E9}".toCodePoints() // true
```

To split text into what a reader sees as characters, use [String.graphemes](graphemes.md) instead:

```ant
len(String.toCodePoints("👍🏽")) // 2
len(String.graphemes("👍🏽"))    // 1
```

The argument is a `string`; passing `Bytes` is a compile error. Decode the bytes with
[String.fromBytes](fromBytes.md) first:

```ant
let points = String.toCodePoints(String.toBytes("a"))
// error: cannot use Bytes where string is expected
```

## See also

[String.fromCodePoints](fromCodePoints.md) · [String.codePoint](codePoint.md) ·
[String.fromCodePoint](fromCodePoint.md) · [String.chars](chars.md) ·
[String.graphemes](graphemes.md) · [String.toBytes](toBytes.md) · [String.length](length.md) ·
[String.normalize](normalize.md)
