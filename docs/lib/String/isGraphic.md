<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.isGraphic

```ant signature
func String.isGraphic(s: string): bool
```

String.isGraphic reports whether every code point of s is in L, M, N, P, S, or Zs. It is false on
"".

<!-- end generated -->

## Details

- `String.isGraphic` is `true` when every code point of `s` has one of the Unicode general
  categories L (letter), M (mark), N (number), P (punctuation), S (symbol), or Zs (space separator).
- It returns `false` on the empty string.
- Every space separator counts, not only the space U+0020: the no-break space U+00A0 and the
  ideographic space U+3000 are graphic. [`String.isPrint`](isPrint.md) admits the space U+0020 and
  no other separator.
- Control characters (Cc), format characters (Cf) such as the zero-width joiner, private use (Co),
  unassigned code points (Cn), and the line and paragraph separators (Zl, Zp) are not graphic.
- The tables are those of the Unicode version that `String.unicodeVersion` names.
- [spec: String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)

## Basic examples

```ant
String.isGraphic("Hello, world!") // true
String.isGraphic("a\tb")          // false
String.isGraphic("🐜")             // true
```

Call it through `.`:

```ant
"x = 1 + 2".isGraphic() // true
```

## Scope

The empty string has no code points, so it is not all graphic:

```ant
String.isGraphic("") // false
```

Every code point in the classes L, M, N, P, and S is graphic, whatever its script: letters, marks,
numbers, punctuation, and symbols:

```ant
String.isGraphic("中文")      // true
String.isGraphic("\u{301}") // true
String.isGraphic("½Ⅻ٣")     // true
String.isGraphic("«»¿")     // true
String.isGraphic("€©+")     // true
```

Of the separators, only the space separators are graphic. The line separator U+2028 is not:

```ant
String.isGraphic(" ")        // true
String.isGraphic("\u{A0}")   // true
String.isGraphic("\u{3000}") // true
String.isGraphic("\u{2028}") // false
```

Control, format, private-use, and unassigned code points are not graphic:

```ant
String.isGraphic("\n")       // false
String.isGraphic("\u{200B}") // false
String.isGraphic("\u{E000}") // false
String.isGraphic("\u{378}")  // false
```

A letter with a combining accent is two code points, a letter and a mark, and both are graphic:

```ant
String.isGraphic("e\u{301}") // true
```

A flag is two regional indicator symbols, category So, so it is graphic too:

```ant
String.isGraphic("🇨🇭") // true
```

A string built with `String.fromCodePoint` works as a literal one does:

```ant
var nbsp = String.fromCodePoint(0xA0) ?? ""
nbsp.isGraphic() // true
```

Test each code point of a string with `String.chars` in a pipe:

```ant
"a\u{200B}b\n" | String.chars() | Array.map({ String.isGraphic(x) }) // [true, false, true, false]
```

## Applications

Accept a label only when it is one line of text with no control or format characters and not only
spaces:

```ant
func validLabel(s: string): bool {
  s.isGraphic() && String.trim(s) != ""
}

for s in ["Inbox", "In\nbox", "In\u{200B}box", "   "] {
  println(validLabel(s))
}
```

```text
true
false
false
false
```

Report each code point that is not graphic, with its position and its category:

```ant
type Bad = { index: int, category: string }

func badCodePoints(s: string): Array[Bad] {
  var out: Array[Bad] = []
  var i = 0
  for c in s {
    if !c.isGraphic() {
      out.push(Bad{ index: i, category: c.category() })
    }
    i += 1
  }
  out
}

for b in badCodePoints("ok\t\u{AD}fine\u{2028}") {
  println("${b.index}: ${b.category}")
}
```

```text
2: Cc
3: Cf
8: Zl
```

Replace every code point that is not graphic with U+FFFD, the replacement character:

```ant
func visible(s: string): string {
  String.concat(s.chars().map({ if x.isGraphic() { x } else { "\u{FFFD}" } }))
}

visible("bell\u{7} tab\t") // "bell� tab�"
```

## Properties and relations

`String.isGraphic(s)` is `s` not empty and every code point in L, M, N, P, S, or Zs:

```ant
for s in ["", "a", "a b", "\u{A0}", "\n", "x\u{AD}", "\u{2028}", "🐜!", "\u{E000}"] {
  var chars = String.chars(s)
  var all = !chars.isEmpty() &&
    chars.every({
      var k = String.category(x)
      ["L", "M", "N", "P", "S"].contains(String.slice(k, 0, 1)) || k == "Zs"
    })
  assert(String.isGraphic(s) == all)
}
```

Every printable string is graphic. The converse fails only on the space separators other than
U+0020:

```ant
for s in ["a", " ", "\u{A0}", "\u{3000}", "\t", "€", "\u{301}"] {
  if String.isPrint(s) {
    assert(String.isGraphic(s))
  }
  var differ = String.isGraphic(s) != String.isPrint(s)
  assert(differ == (String.category(s) == "Zs" && s != " "))
}
```

For two non-empty strings, the concatenation is graphic exactly when both parts are:

```ant
for a in ["a", "\n", "\u{A0}"] {
  for b in ["1", "\u{200D}", "!"] {
    assert(String.isGraphic(a + b) == (String.isGraphic(a) && String.isGraphic(b)))
  }
}
```

## Possible issues

`String.isGraphic` is not a test for visible ink. A string of spaces, no-break spaces, or a lone
combining mark is graphic:

```ant
String.isGraphic("   ")     // true
String.isGraphic("\u{A0}")  // true
String.isGraphic("\u{301}") // true
```

An emoji joined from several, such as a family, holds the zero-width joiner U+200D. The joiner is a
format character, so the whole emoji is not graphic:

```ant
String.isGraphic("👨‍👩‍👧")   // false
String.category("\u{200D}") // "Cf"
```

A graphic string can still render with escapes. A string inside a container shows every code point
that `String.isPrint` rejects as `\u{...}`, and the no-break space is one of them:

```ant
["a\u{A0}b"]                 // ["a\u{a0}b"]
String.isGraphic("a\u{A0}b") // true
```

`String.isGraphic` tests that the whole string is graphic, not that it holds a graphic code point:

```ant
String.isGraphic("a\n")                     // false
"a\n".chars().some({ String.isGraphic(x) }) // true
```

A code point that a Unicode version newer than `String.unicodeVersion` assigns, such as a newly
added emoji, is unassigned (Cn) here, so `String.isGraphic` returns false for it. U+0378 is
unassigned in the tables in use:

```ant
String.unicodeVersion       // "17.0.0"
String.category("\u{378}")  // "Cn"
String.isGraphic("\u{378}") // false
```

See the [spec: String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string) section.

## See also

[String.isPrint](isPrint.md) · [String.isControl](isControl.md) · [String.category](category.md) ·
[String.isSpace](isSpace.md) · [String.isLetter](isLetter.md) · [String.isMark](isMark.md) ·
[String.chars](chars.md) · [String.graphemes](graphemes.md)
