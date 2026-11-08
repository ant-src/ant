<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.fromCodePoint

```ant signature
func String.fromCodePoint(n: int): string | null
```

String.fromCodePoint returns the one-code-point string for n, or null when n is not a Unicode scalar
value.

<!-- end generated -->

## Details

- `String.fromCodePoint` returns a string that holds exactly the code point `n`
  ([spec: String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)).
- It returns `null` when `n` is not a Unicode scalar value: a negative number, a number above
  `0x10FFFF`, or a surrogate, `0xD800` to `0xDFFF`. Every string holds only scalar values, so no
  string could hold such a code point
  ([spec: String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)).
- The result is `string | null`, so narrow it with `??`, `!= null`, or `match` before you use it as
  a `string` ([spec: Narrowing](https://github.com/ant-src/spec/blob/master/ant-lang.md#narrowing)).

## Basic examples

```ant
String.fromCodePoint(65)      // "A"
String.fromCodePoint(233)     // "é"
String.fromCodePoint(0x1F41C) // "🐜"
```

A number that is not a scalar value gives `null`:

```ant
String.fromCodePoint(0xD800) // null
```

## Scope

The valid range runs from `0` to `0x10FFFF`, with a gap for the surrogates. Each side of each edge:

```ant
String.fromCodePoint(-1)       // null
String.fromCodePoint(0)        // "\u{0}"
String.fromCodePoint(0xD7FF)   // "\u{d7ff}"
String.fromCodePoint(0xDFFF)   // null
String.fromCodePoint(0xE000)   // "\u{e000}"
String.fromCodePoint(0x10FFFF) // "\u{10ffff}"
String.fromCodePoint(0x110000) // null
```

The result for a control character is that character, shown here by its escape:

```ant
String.fromCodePoint(10) == "\n" // true
String.fromCodePoint(9)          // "\t"
```

A combining mark is a code point on its own, even though it draws nothing by itself:

```ant
String.fromCodePoint(0x301)?.length() // 1
```

The parameter is named `n`:

```ant
String.fromCodePoint(n: 90) // "Z"
```

In a pipe, the number is the first argument:

```ant
0x41 | String.fromCodePoint() // "A"
```

As a function value, it maps an array of numbers to an array of nullable strings:

```ant
[72, 105, -1].map(String.fromCodePoint) // ["H", "i", null]
```

## Applications

Build the lower-case Latin alphabet from the code point of `"a"`:

```ant
var a = String.codePoint("a")
var letters = Iter.range(0, 26) | Iter.map({ String.fromCodePoint(a + x) ?? "" }) | Iter.toArray()
String.join(letters, "") // "abcdefghijklmnopqrstuvwxyz"
```

Read a code point written as `U+` and hexadecimal digits, and reject one that names no character:

```ant
func parseUPlus(s: string): string | null {
  if !String.startsWith(s, "U+") {
    return null
  }
  var n = String.toInt(String.slice(s, 2, String.length(s)), 16)
  if n == null {
    return null
  }
  String.fromCodePoint(n)
}

parseUPlus("U+1F41C") // "🐜"
parseUPlus("U+00E9")  // "é"
parseUPlus("U+D800")  // null
parseUPlus("U+zz")    // null
```

Replace a value that is not a scalar value with U+FFFD, the replacement character:

```ant
func toText(points: Array[int]): string {
  String.join(points.map({ String.fromCodePoint(x) ?? "\u{FFFD}" }), "")
}

toText([65, 0xD800, 110, 116]) // "A�nt"
```

## Properties and relations

The result of a scalar value is one code point long, and `String.codePoint` turns it back into `n`:

```ant
for n in [0, 65, 233, 0xD7FF, 0xE000, 0x1F41C, 0x10FFFF] {
  var c = String.fromCodePoint(n)
  assert(c != null && String.length(c) == 1 && String.codePoint(c) == n)
}
```

It agrees with the `\u{...}` escape, and with `String.fromCodePoints` on a one-element array:

```ant
for n in [0x41, 0xE9, 0x1F41C] {
  assert(String.fromCodePoint(n) == String.fromCodePoints([n]))
}
assert(String.fromCodePoint(0x1F41C) == "\u{1F41C}")
assert(String.fromCodePoint(0xDC00) == String.fromCodePoints([0xDC00]))
```

## Possible issues

The result may be `null`, so passing it where a `string` is expected is a compile error:

<!-- test: error cannot use string | null where string is expected -->

```ant
var s: string = String.fromCodePoint(65)
```

Give a default with `??`, or test against `null` first.

`String.fromCodePoint` is a function of `String`, not of `int`, so it cannot be called through `.`
on a number:

<!-- test: error int has no field or function fromCodePoint -->

```ant
var n = 97
println(n.fromCodePoint())
```

Call it as `String.fromCodePoint(n)`, or in a pipe as `n | String.fromCodePoint()`.

A code point held in another integer type, such as a `byte` read from `Bytes`, is not an `int`:

<!-- test: error cannot use byte where int is expected -->

```ant
var b = byte(65)
String.fromCodePoint(b)
```

Convert it first: `String.fromCodePoint(int(b))`.

Code point `0` is a string of one character, not the empty string:

```ant
String.fromCodePoint(0) == ""     // false
String.fromCodePoint(0)?.length() // 1
```

One code point is not always one character as a reader sees it. A combining accent joins the letter
before it into one grapheme:

```ant
var e = "e" + (String.fromCodePoint(0x301) ?? "")
String.length(e)             // 2
String.graphemes(e).length() // 1
```

## See also

[String.codePoint](codePoint.md) · [String.fromCodePoints](fromCodePoints.md) ·
[String.toCodePoints](toCodePoints.md) · [String.chars](chars.md) · [String.graphemes](graphemes.md)
· [String.fromBytes](fromBytes.md)
