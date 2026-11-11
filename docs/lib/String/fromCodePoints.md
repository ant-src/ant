<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.fromCodePoints

```ant signature
func String.fromCodePoints(points: Array[int]): string | null
```

String.fromCodePoints returns the string of the code points in points, or null when any is not a
Unicode scalar value.

<!-- end generated -->

## Details

- An empty array gives `""`.
- Returns `null` when any value is not a Unicode scalar value: a negative number, a number above
  `0x10FFFF`, or a surrogate. There is no partial string and no replacement character.

## Basic examples

```ant
String.fromCodePoints([97, 110, 116]) // "ant"
String.fromCodePoints([0x1F41C])      // "🐜"
String.fromCodePoints([0xD800])       // null
```

## Scope

An empty array gives the empty string:

```ant
String.fromCodePoints([]) // ""
```

Every scalar value counts, from `0` up to `0x10FFFF`, control characters and U+FFFD among them:

```ant
String.fromCodePoints([0]) == "\u{0}"             // true
String.fromCodePoints([9, 10]) == "\t\n"          // true
String.fromCodePoints([0x10FFFF]) == "\u{10FFFF}" // true
String.fromCodePoints([0xFFFD]) == "\u{FFFD}"     // true
```

The values just around the surrogate range are scalar values; the surrogates are not:

```ant
String.fromCodePoints([0xD7FF]) == "\u{D7FF}" // true
String.fromCodePoints([0xDFFF])               // null
String.fromCodePoints([0xE000]) == "\u{E000}" // true
```

Values out of range give `null`, wherever they sit in the array:

```ant
String.fromCodePoints([-1])         // null
String.fromCodePoints([0x110000])   // null
String.fromCodePoints([97, 98, -1]) // null
```

A combining mark is a code point of its own, so `e` and U+0301 give a string of two code points that
reads as one character:

```ant
let s = String.fromCodePoints([101, 0x301]) ?? ""
s.length()             // 2
s.graphemes().length() // 1
```

In a pipe, the array is the first argument:

```ant
[72, 105] | String.fromCodePoints() // "Hi"
```

A function value works wherever a `func(Array[int]): string | null` is expected:

```ant
let decode: func(Array[int]): string | null = String.fromCodePoints
decode([79, 75]) // "OK"
```

## Named arguments

Name the argument `points`:

```ant
String.fromCodePoints(points: [65, 66]) // "AB"
```

## Applications

Build the lower-case Latin alphabet from a range of code points:

```ant
String.fromCodePoints(Array.range(97, 123)) // "abcdefghijklmnopqrstuvwxyz"
```

Rotate lower-case Latin letters by 13 with ROT13, working on the code points and leaving the rest as
it is:

```ant
func rot13(s: string): string {
  let points = s.toCodePoints().map({
    if x >= 97 && x <= 122 { (x - 97 + 13) % 26 + 97 } else { x }
  })
  String.fromCodePoints(points) ?? s
}

println(rot13("hello, ant"))
println(rot13(rot13("hello, ant")))
```

```text
uryyb, nag
hello, ant
```

Decode hexadecimal code points from untrusted input, and reject the input when any is not a scalar
value:

```ant
func decode(hex: string): string | null {
  let points = hex.fields().map({ x.toInt(16) ?? -1 })
  String.fromCodePoints(points)
}

println(decode("41 F1 1F41C") ?? "invalid")
println(decode("41 D800") ?? "invalid")
println(decode("41 zz") ?? "invalid")
```

```text
Añ🐜
invalid
invalid
```

## Properties and relations

It undoes [String.toCodePoints](toCodePoints.md):

```ant
for s in ["", "ant", "héllo 日本", "e\u{301}", "🐜\u{0}"] {
  assert(String.fromCodePoints(s.toCodePoints()) == s)
}
```

For one value, it agrees with [String.fromCodePoint](fromCodePoint.md):

```ant
for n in [-1, 0, 65, 0xD800, 0xE000, 0x1F41C, 0x10FFFF, 0x110000] {
  assert(String.fromCodePoints([n]) == String.fromCodePoint(n))
}
```

A result that is not `null` has one code point per value:

```ant
let points = [0x41, 0xE9, 0x65E5, 0x1F41C]
let s = String.fromCodePoints(points)
assert(s != null && s.length() == len(points))
```

It is the [String.concat](concat.md) of each value's one-code-point string:

```ant
let points = [97, 0x301, 0x1F41C]
let parts = points.map({ String.fromCodePoint(x) ?? "" })
assert(String.fromCodePoints(points) == String.concat(parts))
```

## Possible issues

The result may be `null`, so using it where a `string` is expected is a compile error. Supply a
fallback with `??`, or narrow it with `if`:

```ant
let s: string = String.fromCodePoints([97])
// error: cannot use string | null where string is expected
```

One bad value makes the whole result `null`. To keep the valid ones, filter the array first:

```ant
let points = [72, 0xD800, 105, -5]
String.fromCodePoints(points)                                             // null
String.fromCodePoints(points.filter({ String.fromCodePoint(x) != null })) // "Hi"
```

A UTF-16 surrogate pair, as JavaScript's `String.fromCharCode` takes it, is not a code point. Each
half is a surrogate, so the pair gives `null`. Pass the scalar value itself:

```ant
String.fromCodePoints([0xD83D, 0xDC1C]) // null
String.fromCodePoints([0x1F41C])        // "🐜"
```

The argument is an array, not a list of values:

```ant
let s = String.fromCodePoints(97, 98) // error: String.fromCodePoints takes 1 argument, got 2
```

The array is the subject of a call through `.`, and a value reaches its own type's namespace, so
`points.fromCodePoints()` looks in `Array` and is a compile error. Use a pipe instead:

```ant
let s = [97, 98].fromCodePoints() // error: Array[int] has no field or function fromCodePoints
```

The result is not normalized. A letter and a combining accent differ from the one-code-point letter
until [String.normalize](normalize.md) brings both to one form:

```ant
let s = String.fromCodePoints([101, 0x301]) ?? ""
s == "\u{E9}"                  // false
s.normalize("NFC") == "\u{E9}" // true
```

## See also

[String.toCodePoints](toCodePoints.md) · [String.fromCodePoint](fromCodePoint.md) ·
[String.codePoint](codePoint.md) · [String.chars](chars.md) · [String.concat](concat.md) ·
[String.fromBytes](fromBytes.md) · [String.normalize](normalize.md)
