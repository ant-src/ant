<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.padStart

```ant signature
func String.padStart(s: string, width: int, pad: string = " "): string
```

String.padStart returns s with pad repeated on the left until it is width code points long. The
padding is cut to the code points s lacks, and s comes back unchanged when it is already at or
beyond width. pad defaults to a space. It panics if pad is empty or width is above 2^31 - 1.

<!-- end generated -->

## Details

- `width` is the length of the result in code points, not the amount of padding. A string already at
  or beyond `width` comes back unchanged; it is never cut.
- The padding is `pad` repeated and cut to fit, so a `pad` of several code points may end partway
  through.
- Panics if `pad` is empty.

## Basic examples

Pad a number with zeros to a fixed width:

```ant
String.padStart("42", 5, "0") // "00042"
```

Without `pad`, the padding is spaces, which right-aligns the text:

```ant
String.padStart("42", 5) // "   42"
```

A `pad` of several code points repeats and is cut where the width is reached:

```ant
String.padStart("7", 4, "ab") // "aba7"
```

## Scope

A string at or beyond `width` comes back unchanged, never cut:

```ant
String.padStart("hello", 5, "*") // "hello"
String.padStart("hello", 3, "*") // "hello"
String.padStart("hello", 0, "*") // "hello"
String.padStart("hello", -1)     // "hello"
```

An empty string becomes all padding:

```ant
String.padStart("", 3, "0") // "000"
String.padStart("", 0, "0") // ""
```

Width counts code points, so a non-ASCII `pad` or `s` takes one place per code point, whatever its
size in bytes:

```ant
String.padStart("x", 4, "🐜")  // "🐜🐜🐜x"
String.padStart("é", 3, "·")  // "··é"
String.padStart("日本", 4, "＿") // "＿＿日本"
```

Call it in a pipe, through `.` on a string, with every argument named, or inside a lambda:

```ant
"42" | String.padStart(6, pad: "0")                      // "000042"
"7".padStart(3, "0")                                     // "007"
String.padStart(s: "a", width: 3, pad: "-")              // "--a"
Array.map(["1", "22", "333"], { String.padStart(x, 3) }) // ["  1", " 22", "333"]
```

## Named arguments

### pad

`pad` is the text repeated on the left and defaults to `" "`. Name it to make a call read clearly:

```ant
String.padStart("9", 3)            // "  9"
String.padStart("9", 3, pad: "0")  // "009"
String.padStart("9", 3, pad: "-=") // "-=9"
```

## Applications

Right-align a column of numbers:

```ant
var prices = [5, 120, 42, 1999]
for p in prices {
  println(String.padStart(String.from(p), 6))
}
```

```text
     5
   120
    42
  1999
```

Format a time of day with two digits per field:

```ant
type Clock = { hour: int, minute: int, second: int }

func two(n: int): string {
  String.padStart(String.from(n), 2, "0")
}

func show(c: Clock): string {
  "${two(c.hour)}:${two(c.minute)}:${two(c.second)}"
}

println(show(Clock{ hour: 9, minute: 5, second: 30 }))
println(show(Clock{ hour: 23, minute: 59, second: 0 }))
```

```text
09:05:30
23:59:00
```

Write a byte as eight binary digits, the zero-padded form a format string would give:

```ant
for b in [byte(5), byte(160)] {
  println(String.padStart(String.fromInt(b, 2), 8, "0"))
}
```

```text
00000101
10100000
```

Draw a dotted leader in a table of contents:

```ant
type Chapter = { title: string, page: int }

var toc = [Chapter{ title: "Intro", page: 1 }, Chapter{ title: "Strings", page: 14 }]
for e in toc {
  println(e.title + String.padStart(String.from(e.page), 16 - String.length(e.title), "."))
}
```

```text
Intro..........1
Strings.......14
```

## Properties and relations

The result is `width` code points long, or `s` itself when `s` is already that long:

```ant
for s in ["", "a", "abc", "日本語"] {
  for w in [0, 2, 5] {
    var r = String.padStart(s, w, "xy")
    assert(String.length(r) == Math.max(w, String.length(s)))
    assert(String.endsWith(r, s))
  }
}
```

With a `pad` of one code point, the padding is [`String.repeat`](repeat.md) of it:

```ant
var s = "42"
assert(String.padStart(s, 6, "0") == String.repeat("0", 4) + s)
```

[`String.padEnd`](padEnd.md) puts the same padding on the right instead:

```ant
assert(String.padStart("7", 4, "ab") == "aba7")
assert(String.padEnd("7", 4, "ab") == "7aba")
```

## Possible issues

An empty `pad` panics, even when no padding is needed:

```ant
String.padStart("a", 3, "") // panic: String.padStart needs a pad that is not empty
```

A `width` above 2^31 − 1 panics, and so does a result longer than 2^31 − 1 bytes:

```ant
String.padStart("a", 2147483648) // panic: String.padStart needs a width of at most 2147483647, got 2147483648
```

Width counts code points, not what a reader sees as characters. An `e` followed by a combining
accent is two code points, so it gets one place less of padding than a precomposed `é`:

```ant
String.padStart("e\u{301}", 3, "*") == "*e\u{301}" // true
String.padStart("\u{E9}", 3, "*") == "**\u{E9}"    // true
```

Normalize the text with [`String.normalize`](normalize.md) first, or count
[`String.graphemes`](graphemes.md) when you need what a reader sees.

The padding is cut by code points too, so a `pad` that is one character made of several code points
can end inside that character. A family emoji is three people joined by U+200D, five code points,
and two places of padding keep only the first person and a joiner:

```ant
String.padStart("x", 3, "👨\u{200D}👩\u{200D}👧") == "👨\u{200D}x" // true
```

Use a `pad` of one code point when the padding must stay whole, or split the `pad` with
[`String.graphemes`](graphemes.md) and repeat whole graphemes yourself.

A negative number gets its zeros in front of the sign, because `String.from` puts the `-` first and
the padding goes on the left of everything:

```ant
String.padStart(String.from(-42), 5, "0") // "00-42"
```

Pad the absolute value and put the sign in front:

```ant
func zeros(n: int, width: int): string {
  if n < 0 {
    "-" + String.padStart(String.from(-n), width - 1, "0")
  } else {
    String.padStart(String.from(n), width, "0")
  }
}

println(zeros(-42, 5))
println(zeros(42, 5))
```

```text
-0042
00042
```

`padStart` never shortens a string, so a column of values wider than `width` stays ragged. Cut first
with [`String.slice`](slice.md) when the width is a hard limit:

```ant
String.padStart("overflow", 4)                     // "overflow"
String.padStart(String.slice("overflow", 0, 4), 4) // "over"
```

`width` is an `int`, so a float is a compile error:

```ant
println(String.padStart("7", 2.0)) // error: cannot use float64 where int is expected
```

## See also

[String.padEnd](padEnd.md) · [String.repeat](repeat.md) · [String.fromInt](fromInt.md) ·
[String.fromFloat](fromFloat.md) · [String.length](length.md) · [String.slice](slice.md) ·
[String.trimStart](trimStart.md)
