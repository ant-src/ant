<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.isSpace

```ant signature
func String.isSpace(s: string): bool
```

String.isSpace reports whether every code point of s has the White\_Space property. It is false on
"".

<!-- end generated -->

## Details

- `true` when every code point of `s` has the Unicode White_Space property; `false` on `""`.
- Tab, line feed, carriage return, the no-break space U+00A0, and the line separator U+2028 all
  count. The zero-width space U+200B and the byte order mark U+FEFF do not.

## Basic examples

```ant
String.isSpace(" ")    // true
String.isSpace("\t\n") // true
String.isSpace(" a ")  // false
```

Call it through `.`:

```ant
"   ".isSpace() // true
```

## Scope

The empty string has no code points to test, and gives `false`:

```ant
String.isSpace("") // false
```

The control characters tab, line feed, vertical tab, form feed, carriage return, and next line
U+0085 are white space:

```ant
String.isSpace("\t")     // true
String.isSpace("\n")     // true
String.isSpace("\u{B}")  // true
String.isSpace("\u{C}")  // true
String.isSpace("\r")     // true
String.isSpace("\u{85}") // true
```

Space separators of other widths count, and so do the line and paragraph separators:

```ant
String.isSpace("\u{A0}")   // true
String.isSpace("\u{2009}") // true
String.isSpace("\u{3000}") // true
String.isSpace("\u{2028}") // true
String.isSpace("\u{2029}") // true
```

Invisible characters that are not white space do not: the zero-width space, the byte order mark, and
the Mongolian vowel separator:

```ant
String.isSpace("\u{200B}") // false
String.isSpace("\u{FEFF}") // false
String.isSpace("\u{180E}") // false
```

Test one code point at a time with `String.chars`, here in a pipe:

```ant
"a b\tc" | String.chars() | Array.map({ x.isSpace() }) // [false, true, false, true, false]
```

Pass it to a higher-order function as a function value:

```ant
["a", " ", "", "\n", "\u{A0}"].filter(String.isSpace) // [" ", "\n", "\u{a0}"]
```

## Applications

Drop the lines of a text that hold only white space:

```ant
let text = "first\n   \n\t\nsecond"
text.lines().filter({ !x.isSpace() }) // ["first", "second"]
```

Collapse every run of white space into one space:

```ant
func squeeze(s: string): string {
  var out = ""
  var inSpace = false
  for c in s.chars() {
    if c.isSpace() {
      if !inSpace {
        out += " "
      }
      inSpace = true
    } else {
      out += c
      inSpace = false
    }
  }
  out
}

squeeze("one  two\t\n three") // "one two three"
```

Report the lines of a text that end in white space:

```ant
let source = "let a = 1 \nlet b = 2\nlet c = 3\t"
var number = 0
for line in source.lines() {
  number += 1
  let last = line.chars().last()
  if last != null && last.isSpace() {
    println("line ${number}: trailing white space")
  }
}
```

```text
line 1: trailing white space
line 3: trailing white space
```

## Properties and relations

`String.isSpace(s)` is `s` not empty and every code point white space:

```ant
for s in ["", " ", "\t\n", "a", " a ", "\u{A0}", "\u{200B}"] {
  let all = !s.isEmpty() && s.chars().every({ x.isSpace() })
  assert(String.isSpace(s) == all)
}
```

On a nonempty string, `String.isSpace` agrees with [`String.trim`](trim.md) and
[`String.fields`](fields.md), which use the same white space:

```ant
for s in [" ", "\t\n", "a", " a ", "\u{A0}", "\u{3000}", "\u{200B}", "\u{85}"] {
  assert(String.isSpace(s) == (String.trim(s) == ""))
  assert(String.isSpace(s) == String.fields(s).isEmpty())
}
```

## Possible issues

`String.isSpace` is `false` on the empty string, so it does not answer whether a string is blank. An
empty line passes a filter that drops white-space lines:

```ant
"a\n\n \nb".lines().filter({ !x.isSpace() }) // ["a", "", "b"]
```

To test for blank, empty included, compare the trimmed string with `""`:

```ant
"a\n\n \nb".lines().filter({ x.trim() != "" }) // ["a", "b"]
```

Every code point must be white space, so `String.isSpace` does not answer whether a string contains
white space. Use `Array.some` for that:

```ant
String.isSpace("a b")               // false
"a b".chars().some({ x.isSpace() }) // true
```

The zero-width space and the byte order mark are not white space. `String.isSpace` rejects them, and
`String.trim` leaves them in place:

```ant
String.isSpace("\u{FEFF}") // false
String.trim("\u{FEFF} x ") // "\u{feff} x"
```

The no-break space is white space, though it exists to keep its neighbors together. `String.fields`
splits at it:

```ant
"10\u{A0}km".fields() // ["10", "km"]
```

## See also

[String.trim](trim.md) · [String.fields](fields.md) · [String.lines](lines.md) ·
[String.isControl](isControl.md) · [String.isGraphic](isGraphic.md) · [String.isPrint](isPrint.md) ·
[String.category](category.md) · [String.chars](chars.md) · [Array.every](../Array/every.md) ·
[Array.some](../Array/some.md)
