<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String

String holds the functions of string, an immutable sequence of Unicode code points stored as UTF-8.
Lengths, indices, and slices count code points, and a character is a string of one code point;
[`String.graphemes`](graphemes.md) gives what a reader sees as characters.

<!-- end generated -->

The `String` namespace holds the functions on strings. Each function that works on a string takes
that string first, so `String.trim(s)`, `s.trim()`, and `s | String.trim()` are the same call. A few
take something else first and are called through the namespace or a pipe: `String.join` and
`String.concat` take an array, and `String.fromInt`, `String.fromCodePoint`, and the other `from`
functions build a string from another value.

## Details

- A `string` is immutable and stored as UTF-8. No function changes a string; the ones that edit text
  return a new one ([String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)).
- Lengths, indices, and slices count code points. There is no `s[i]`: `String.chars` gives the code
  points as an array to index, `String.toCodePoints` gives them as integers, and `String.toBytes`
  gives the bytes.
- There is no character type. A character is a string of one code point, so the class tests such as
  `String.isDigit` take a string. What a reader sees as one character can be several code points;
  `String.graphemes` splits text into those units.
- Every string holds only Unicode scalar values. Text that reaches a program as a `string` from
  outside it (a file, standard input, the network, an argument, the environment, or a child process)
  is decoded as UTF-8, with each ill-formed sequence replaced by U+FFFD. To keep the bytes as they
  are, read them as `Bytes` and test them with `String.fromBytes`.
- `String.unicodeVersion` names the version of the Unicode tables that grapheme splitting,
  normalization, case mapping, and the class tests follow, such as `"17.0.0"`.
- There is no format string: fixed-point and zero-padded numbers are `String.fromFloat(x, 2)` and
  `String.padStart(String.fromInt(n, 16), 8, "0")`, put in with interpolation
  ([String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)).
- `"..."` is a string that takes escapes and `${...}` interpolation; each interpolated value is
  rendered with `String.from`. A raw string between backticks has neither
  ([Literals](https://github.com/ant-src/spec/blob/master/ant-lang.md#literals)).
- `+` concatenates two strings, `==` compares their code points, and `<`, `<=`, `>`, `>=` compare by
  code point ([Expressions](https://github.com/ant-src/spec/blob/master/ant-lang.md#expressions)).
- `for c in s` visits each code point as a one-code-point string
  ([Loops](https://github.com/ant-src/spec/blob/master/ant-lang.md#loops)).
- `len(s)` and `String.length(s)` are the same count
  ([Prelude](https://github.com/ant-src/spec/blob/master/ant-lang.md#prelude)).
- A call through `.` on a string goes to the `String` namespace and passes the string first
  ([Calls through `.`](https://github.com/ant-src/spec/blob/master/ant-lang.md#calls-through-)).

## Basic examples

Clean up a value and test it:

```ant
var name = "  Ada Lovelace "
String.trim(name)                   // "Ada Lovelace"
String.startsWith(name.trim(), "A") // true
String.length("ant")                // 3
```

Split a line into fields and join them back:

```ant
var parts = String.split("red,green,blue", ",")
parts                     // ["red", "green", "blue"]
String.join(parts, " / ") // "red / green / blue"
```

Calls chain through `.`, or the string goes down a pipe:

```ant
var raw = "  Hello, World  "
raw.trim().toLower()                                 // "hello, world"
raw | String.trim() | String.replace("World", "Ant") // "Hello, Ant"
```

## Scope

Interpolation renders any value with `String.from`, and `+` joins two strings:

```ant
var n = 3
"${n} ants"         // "3 ants"
"ant" + "hill"      // "anthill"
String.from([1, 2]) // "[1, 2]"
```

A raw string keeps backslashes and `${` as written:

```ant
`C:\temp\${x}` // "C:\\temp\\\${x}"
```

A `for` loop visits each code point as a one-code-point string:

```ant
for c in "née" {
  println(c)
}
```

```text
n
é
e
```

Text outside ASCII counts by code point. An accent written as its own code point counts on its own,
and `String.graphemes` keeps it with its letter:

```ant
String.length("日本語")          // 3
String.length("e\u{301}")     // 2
String.graphemes("e\u{301}x") // ["é", "x"]
String.toCodePoints("é")      // [233]
```

Numbers go to and from text through dedicated functions. A string that is not a number reads as
`null`:

```ant
String.toInt("42")           // 42
String.toInt("4 2")          // null
String.fromInt(255, 16)      // "ff"
String.fromFloat(3.14159, 2) // "3.14"
```

There is no format string; a zero-padded number is `String.fromInt` inside `String.padStart`:

```ant
String.padStart(String.fromInt(255, 16), 8, "0") // "000000ff"
```

`String.cut` returns a [String.Cut](types/Cut.md) record, or `null` when the separator is absent:

```ant
String.cut("key=value", "=") // String.Cut{ before: "key", after: "value" }
String.cut("novalue", "=")   // null
```

## Functions by job

### Size and pieces

- [String.length](length.md): the number of code points
- [String.isEmpty](isEmpty.md): whether there are none
- [String.slice](slice.md): the code points from `from` up to `to`
- [String.chars](chars.md): one string per code point
- [String.graphemes](graphemes.md): one string per character as a reader sees it

### Searching

- [String.contains](contains.md): whether `t` occurs in `s`
- [String.startsWith](startsWith.md), [String.endsWith](endsWith.md): whether `s` begins or ends
  with `t`
- [String.indexOf](indexOf.md), [String.lastIndexOf](lastIndexOf.md): the first or last code point
  index of `t`, or `null`
- [String.count](count.md): how many times `t` occurs, without overlap

### Splitting and joining

- [String.split](split.md): the pieces between each separator
- [String.cut](cut.md): the two sides of the first separator, as [String.Cut](types/Cut.md)
- [String.fields](fields.md): the runs between white space
- [String.lines](lines.md): the lines of a text
- [String.join](join.md): pieces with a separator between them
- [String.concat](concat.md): pieces one after the other
- [String.repeat](repeat.md): `n` copies of a string

### Trimming and padding

- [String.trim](trim.md), [String.trimStart](trimStart.md), [String.trimEnd](trimEnd.md): white
  space removed at both ends, the start, or the end
- [String.trimPrefix](trimPrefix.md), [String.trimSuffix](trimSuffix.md): a prefix or suffix removed
  once
- [String.padStart](padStart.md), [String.padEnd](padEnd.md): padding up to a width

### Replacing

- [String.replace](replace.md): every occurrence replaced
- [String.replaceFirst](replaceFirst.md): the first occurrence replaced

### Case and normalization

- [String.toUpper](toUpper.md), [String.toLower](toLower.md): upper or lower case, with Unicode's
  full mappings
- [String.foldCase](foldCase.md): a key in which case does not matter
- [String.equalFold](equalFold.md): whether two strings are equal without regard to case
- [String.normalize](normalize.md): one of the four Unicode normalization forms

### Character classes

- [String.isDigit](isDigit.md), [String.isNumber](isNumber.md): decimal digits, or any numbers
- [String.isLetter](isLetter.md), [String.isUpper](isUpper.md), [String.isLower](isLower.md):
  letters, upper-case, or lower-case letters
- [String.isSpace](isSpace.md): white space
- [String.isPunct](isPunct.md), [String.isSymbol](isSymbol.md), [String.isMark](isMark.md):
  punctuation, symbols, or marks such as combining accents
- [String.isControl](isControl.md), [String.isGraphic](isGraphic.md), [String.isPrint](isPrint.md):
  control characters (Cc), graphic code points (L, M, N, P, S, Zs), or printable ones (L, M, N, P,
  S, and the space)
- [String.category](category.md): the general category of one code point
- [String.inScript](inScript.md): whether one code point belongs to a script

### Code points and bytes

- [String.codePoint](codePoint.md), [String.fromCodePoint](fromCodePoint.md): one code point to an
  `int` and back
- [String.toCodePoints](toCodePoints.md), [String.fromCodePoints](fromCodePoints.md): every code
  point to an `int` and back
- [String.toBytes](toBytes.md): the UTF-8 bytes
- [String.fromBytes](fromBytes.md): bytes as text, or `null` when they are not UTF-8
- [String.fromBytesLossy](fromBytesLossy.md): bytes as text, with U+FFFD for what is not UTF-8

### Numbers and rendering

- [String.toInt](toInt.md), [String.toUint](toUint.md), [String.toFloat](toFloat.md): a number read
  from text, or `null`
- [String.fromInt](fromInt.md): an integer in a base
- [String.fromFloat](fromFloat.md): a float with a fixed number of decimals
- [String.from](from.md): any value, as `println` and interpolation render it

## Applications

Read `key=value` settings, skipping blank lines and lines without `=`:

```ant
var text = "host=localhost\n\nport=8080\nverbose\n"
var settings: Map[string, string] = Map{}
for line in String.lines(text) {
  var kv = String.cut(line, "=")
  if kv != null {
    settings[String.trim(kv.before)] = String.trim(kv.after)
  }
}
settings // Map{ "host": "localhost", "port": "8080" }
```

Format a table row with fixed-width columns:

```ant
func row(name: string, qty: int): string {
  String.padEnd(name, 8, ".") + String.padStart(String.fromInt(qty), 4)
}
row("apples", 12) // "apples..  12"
row("figs", 300)  // "figs.... 300"
```

Build a slug from a title:

```ant
var title = "  Hello, Ant World! "
var slug = String.fields(String.toLower(title))
  | Array.map({ String.chars(x).filter({ String.isLetter(x) }) | String.concat() })
  | String.join("-")
println(slug)
```

```text
hello-ant-world
```

## Properties and relations

`len` and `String.length` agree, and a string is empty when its length is `0`:

```ant
var s = "anthill"
assert(len(s) == String.length(s))
assert(String.isEmpty(s) == (len(s) == 0))
```

Splitting and joining with the same separator gives the string back:

```ant
var s = "a,b,,c"
assert(String.join(String.split(s, ","), ",") == s)
```

The code points of a string, its characters, and its UTF-8 bytes all lead back to it:

```ant
var s = "naïve 🐜"
assert(String.concat(String.chars(s)) == s)
assert(String.fromCodePoints(String.toCodePoints(s)) == s)
assert(String.fromBytes(String.toBytes(s)) == s)
assert(len(String.chars(s)) == len(s))
```

A call through `.`, a pipe, and a direct call are the same call:

```ant
var s = " Ant "
assert(s.trim() == String.trim(s))
assert((s | String.trim()) == String.trim(s))
```

## Possible issues

A string cannot be indexed. Take one code point with `String.slice`, or index the array of
`String.chars`:

<!-- test: error string cannot be indexed -->

```ant
var s = "ant"
println(s[0])
```

```ant
var s = "ant"
String.slice(s, 0, 1) // "a"
String.chars(s)[0]    // "a"
```

`+` joins only two strings; a number goes in through interpolation or `String.fromInt`:

<!-- test: error + needs both operands of one type, got string and int -->

```ant
var n = 3
println("ants: " + n)
```

```ant
var n = 3
"ants: ${n}"                 // "ants: 3"
"ants: " + String.fromInt(n) // "ants: 3"
```

`==` compares code points and does not normalize, so the two spellings of `é` differ until both go
through `String.normalize`:

```ant
"\u{E9}" == "e\u{301}"                                                   // false
String.normalize("\u{E9}", "NFC") == String.normalize("e\u{301}", "NFC") // true
```

Comparison is by code point, not by dictionary order, so every upper-case ASCII letter comes before
every lower-case one:

```ant
"Zebra" < "apple"                                   // true
String.foldCase("Zebra") < String.foldCase("apple") // false
```

`join` is a `String` function, not an `Array` one. A call through `.` on an array goes to the
`Array` namespace, so `parts.join(",")` does not compile
([Calls through `.`](https://github.com/ant-src/spec/blob/master/ant-lang.md#calls-through-)):

<!-- test: error has no field or function join -->

```ant
var parts = ["a", "b"]
println(parts.join(","))
```

```ant
var parts = ["a", "b"]
String.join(parts, ",")  // "a,b"
parts | String.join(",") // "a,b"
```

`s.length` without parentheses reads a field, and a string has no fields. Call it, or use `len`:

<!-- test: error has no field length -->

```ant
var s = "ab"
println(s.length)
```

```ant
var s = "ab"
s.length() // 2
len(s)     // 2
```

`String.toLower` applies the full case mappings and is not a key for comparing text without regard
to case. Compare or key such text through `String.foldCase` or `String.equalFold`:

```ant
String.toLower("STRASSE") == String.toLower("Straße") // false
String.equalFold("STRASSE", "Straße")                 // true
```

A code point is not always a character a reader sees. `String.length` counts the code points of a
flag, and `String.graphemes` counts the flag once:

```ant
String.length("🇨🇭")         // 2
len(String.graphemes("🇨🇭")) // 1
```

## See also

[String type](../../language/types/string.md),
[String literals](../../language/literals/strings.md),
[Raw strings](../../language/literals/raw-strings.md), [Bytes](../Bytes/index.md),
[Regex](../Regex/index.md), [String.Cut](types/Cut.md), [len](../prelude/len.md),
[Text](../guides/text.md)
