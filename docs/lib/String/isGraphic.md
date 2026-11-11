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

- `true` when every code point of `s` is a letter, mark, number, punctuation, symbol, or space
  separator (`Zs`); `false` on `""`.
- Every space separator counts, such as the no-break space U+00A0. [`String.isPrint`](isPrint.md)
  admits only the space U+0020.
- Control and format characters, private use, unassigned code points, and the line and paragraph
  separators are not graphic.

## Basic examples

```ant
String.isGraphic("Hello, world!") // true
String.isGraphic("a\tb")          // false
String.isGraphic("🐜")            // true
```

Call it through `.`:

```ant
"x = 1 + 2".isGraphic() // true
```

## Scope

The empty string has no code points, so it is not all graphic:

```ant
"".isGraphic() // false
```

Every code point in the classes L, M, N, P, and S is graphic, whatever its script: letters, marks,
numbers, punctuation, and symbols:

```ant
"中文".isGraphic()    // true
"\u{301}".isGraphic() // true
"½Ⅻ٣".isGraphic()     // true
"«»¿".isGraphic()     // true
"€©+".isGraphic()     // true
```

Of the separators, only the space separators are graphic. The line separator U+2028 is not:

```ant
" ".isGraphic()        // true
"\u{A0}".isGraphic()   // true
"\u{3000}".isGraphic() // true
"\u{2028}".isGraphic() // false
```

Control, format, private-use, and unassigned code points are not graphic:

```ant
"\n".isGraphic()       // false
"\u{200B}".isGraphic() // false
"\u{E000}".isGraphic() // false
"\u{378}".isGraphic()  // false
```

A letter with a combining accent is two code points, a letter and a mark, and both are graphic:

```ant
"e\u{301}".isGraphic() // true
```

A flag is two regional indicator symbols, category So, so it is graphic too:

```ant
"🇨🇭".isGraphic() // true
```

A string built with `String.fromCodePoint` works as a literal one does:

```ant
let nbsp = String.fromCodePoint(0xA0) ?? ""
nbsp.isGraphic() // true
```

Test each code point of a string with `String.chars` in a pipe:

```ant
"a\u{200B}b\n" | String.chars() | Array.map({ x.isGraphic() }) // [true, false, true, false]
```

## Applications

Accept a label only when it is one line of text with no control or format characters and not only
spaces:

```ant
func validLabel(s: string): bool {
  s.isGraphic() && s.trim() != ""
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
  let out: Array[Bad] = []
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
  let chars = s.chars()
  let all = !chars.isEmpty() &&
    chars.every({
      let k = x.category()
      ["L", "M", "N", "P", "S"].contains(k.slice(0, 1)) || k == "Zs"
    })
  assert(s.isGraphic() == all)
}
```

Every printable string is graphic. The converse fails only on the space separators other than
U+0020:

```ant
for s in ["a", " ", "\u{A0}", "\u{3000}", "\t", "€", "\u{301}"] {
  if s.isPrint() {
    assert(s.isGraphic())
  }
  let differ = s.isGraphic() != s.isPrint()
  assert(differ == (s.category() == "Zs" && s != " "))
}
```

For two non-empty strings, the concatenation is graphic exactly when both parts are:

```ant
for a in ["a", "\n", "\u{A0}"] {
  for b in ["1", "\u{200D}", "!"] {
    assert(String.isGraphic(a + b) == (a.isGraphic() && b.isGraphic()))
  }
}
```

## Possible issues

`String.isGraphic` is not a test for visible ink. A string of spaces, no-break spaces, or a lone
combining mark is graphic:

```ant
"   ".isGraphic()     // true
"\u{A0}".isGraphic()  // true
"\u{301}".isGraphic() // true
```

An emoji joined from several, such as a family, holds the zero-width joiner U+200D. The joiner is a
format character, so the whole emoji is not graphic:

```ant
"👨‍👩‍👧".isGraphic()      // false
"\u{200D}".category() // "Cf"
```

A graphic string can still render with escapes. A string inside a container shows every code point
that `String.isPrint` rejects as `\u{...}`, and the no-break space is one of them:

```ant
["a\u{A0}b"]           // ["a\u{a0}b"]
"a\u{A0}b".isGraphic() // true
```

`String.isGraphic` tests that the whole string is graphic, not that it holds a graphic code point:

```ant
"a\n".isGraphic()                     // false
"a\n".chars().some({ x.isGraphic() }) // true
```

A code point that a Unicode version newer than `String.unicodeVersion` assigns, such as a newly
added emoji, is unassigned (Cn) here, so `String.isGraphic` returns false for it. U+0378 is
unassigned in the tables in use:

```ant
String.unicodeVersion // "17.0.0"
"\u{378}".category()  // "Cn"
"\u{378}".isGraphic() // false
```

## See also

[String.isPrint](isPrint.md) · [String.isControl](isControl.md) · [String.category](category.md) ·
[String.isSpace](isSpace.md) · [String.isLetter](isLetter.md) · [String.isMark](isMark.md) ·
[String.chars](chars.md) · [String.graphemes](graphemes.md)
