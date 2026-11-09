<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.padEnd

```ant signature
func String.padEnd(s: string, width: int, pad: string = " "): string
```

String.padEnd returns s with pad repeated on the right until it is width code points long. It pads
and panics as [`String.padStart`](padStart.md) does.

<!-- end generated -->

## Details

- `width` is the length of the result in code points, not the amount of padding. A string already at
  or beyond `width` comes back unchanged; it is never cut.
- The padding is `pad` repeated and cut to fit, so a `pad` of several code points may end partway
  through.
- Panics if `pad` is empty.

## Basic examples

```ant
String.padEnd("ant", 6, ".") // "ant..."
String.padEnd("7", 4, "ab")  // "7aba"
```

Without `pad`, the padding is spaces:

```ant
String.padEnd("ant", 6) // "ant   "
```

## Scope

A string at or beyond `width` comes back unchanged, whatever `width` is:

```ant
String.padEnd("ant", 3, ".")     // "ant"
String.padEnd("antenna", 3, ".") // "antenna"
String.padEnd("ant", 0, ".")     // "ant"
String.padEnd("ant", -5, ".")    // "ant"
```

The empty string becomes `width` code points of padding:

```ant
String.padEnd("", 5, "-") // "-----"
```

A pad of several code points repeats and is cut where the width ends:

```ant
String.padEnd("ab", 5, "xyz") // "abxyz"
String.padEnd("ab", 6, "xyz") // "abxyzx"
String.padEnd("ab", 3, "xyz") // "abx"
```

Both `s` and `pad` count in code points, in any script:

```ant
String.padEnd("é", 3, "ü")    // "éüü"
String.padEnd("東京", 4, "・")   // "東京・・"
String.padEnd("Ωμ", 4, "αβγ") // "Ωμαβ"
```

Call it through `.` on any string, in a pipeline, where the string is the first argument, or inside
a lambda:

```ant
"ant".padEnd(5, "!")                  // "ant!!"
"ant" | String.padEnd(5, "?")         // "ant??"
["a", "bb"].map({ x.padEnd(3, ".") }) // ["a..", "bb."]
```

## Named arguments

`pad:` names the padding; leave it out for spaces:

```ant
String.padEnd("total", 8, pad: "_") // "total___"
String.padEnd("total", 8)           // "total   "
```

Naming every argument lets them come in any order:

```ant
String.padEnd(width: 6, pad: "=", s: "id") // "id===="
```

## Applications

Line up a column of labels so the values start at one place:

```ant
type Item = { name: string, price: int }

var prices = [Item{ name: "tea", price: 3 }, Item{ name: "coffee", price: 4 }, Item{ name: "pie", price: 12 }]
for p in prices {
  println("${p.name.padEnd(8, ".")} ${p.price}")
}
```

```text
tea..... 3
coffee.. 4
pie..... 12
```

Pair it with [`String.padStart`](padStart.md) for a table with left-aligned text and right-aligned
numbers:

```ant
type Row = { item: string, count: int }

func render(rows: Array[Row]) {
  for r in rows {
    println(r.item.padEnd(10) + String.fromInt(r.count).padStart(5))
  }
}

render([
  Row{ item: "apples", count: 12 },
  Row{ item: "kiwis", count: 7 },
  Row{ item: "plums", count: 140 },
])
```

```text
apples       12
kiwis         7
plums       140
```

Size the column to the longest label instead of a fixed number:

```ant
var fields = [["name", "ant"], ["version", "0.4.0"], ["license", "MIT"]]
var width = fields.map({ x[0].length() }).max() ?? 0
for f in fields {
  println("${f[0].padEnd(width)} : ${f[1]}")
}
```

```text
name    : ant
version : 0.4.0
license : MIT
```

Fill a fixed-width record field, such as a ten-character name in a data file:

```ant
func field(s: string, n: int): string {
  s.slice(0, Math.min(s.length(), n)).padEnd(n)
}

field("Ann", 10) + "|"         // "Ann       |"
field("Bartholomew", 10) + "|" // "Bartholome|"
```

## Properties and relations

The result is never shorter than `width`, never shorter than `s`, and always starts with `s`:

```ant
for s in ["", "a", "ant", "antenna"] {
  for width in [-1, 0, 3, 5] {
    var r = String.padEnd(s, width, "xy")
    assert(r.length() == Math.max(s.length(), width))
    assert(r.startsWith(s))
  }
}
```

The padding is `pad` repeated and then cut, which [`String.repeat`](repeat.md) and
[`String.slice`](slice.md) can write out:

```ant
var s = "ab"
var pad = "xyz"
var width = 7
var lacking = width - s.length()
var padding = String.repeat(pad, lacking).slice(0, lacking)
assert(String.padEnd(s, width, pad) == s + padding)
```

With a one-code-point pad, `String.padEnd` and [`String.padStart`](padStart.md) add the same padding
on opposite sides:

```ant
var s = "42"
var right = String.padEnd(s, 6, "0")
var left = String.padStart(s, 6, "0")
assert(String.trimSuffix(right, "0000") == s)
assert(String.trimPrefix(left, "0000") == s)
```

## Possible issues

`width` is the length of the result, not how much to add:

```ant
String.padEnd("ant", 2, ".") // "ant"
String.padEnd("ant", 5, ".") // "ant.."
```

An empty pad panics:

```ant
String.padEnd("ant", 5, "") // panic: String.padEnd needs a pad that is not empty
```

A width above 2^31 − 1 panics, even when it is a valid `int`:

```ant
String.padEnd("ant", 3000000000) // panic: String.padEnd needs a width of at most 2147483647, got 3000000000
```

A width inside the limit still panics when the result would pass 2^31 − 1 bytes, because a code
point of the pad can take up to four bytes:

```ant
String.padEnd("", 2147483647, "🐜") // panic: String.padEnd would give a string longer than 2147483647 bytes
```

`String.padEnd` counts code points, not what a reader sees as one character. A family emoji is five
code points, so it takes most of the width:

```ant
String.padEnd("👨‍👩‍👧", 6, ".")     // "👨\u{200d}👩\u{200d}👧."
String.length("👨‍👩‍👧")             // 5
String.graphemes("👨‍👩‍👧").length() // 1
```

The same holds for "e" followed by a combining accent, which is two code points. Normalize with
[`String.normalize`](normalize.md) when text can arrive in either spelling:

```ant
String.padEnd("cafe\u{301}", 6, ".")                          // "café."
String.padEnd(String.normalize("cafe\u{301}", "NFC"), 6, ".") // "café.."
```

A code point is not a column on a terminal either: `"東京"` is two code points, but most terminals
draw each of them two columns wide, so a column padded with `String.padEnd` does not line up there.

`width` is an `int`. A `float64` width is a compile error; convert it first:

<!-- test: error cannot use float64 where int is expected -->

```ant
var w = 6.0
println(String.padEnd("ant", w))
```

## See also

[String.padStart](padStart.md) · [String.repeat](repeat.md) · [String.length](length.md) ·
[String.slice](slice.md) · [String.trimEnd](trimEnd.md) · [String.graphemes](graphemes.md) ·
[String.fromInt](fromInt.md)
