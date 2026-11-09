<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.isControl

```ant signature
func String.isControl(s: string): bool
```

String.isControl reports whether every code point of s is a control character (Cc). It is false on
"".

<!-- end generated -->

## Details

- `true` when every code point of `s` has the general category `Cc`; `false` on `""`.
- Only `Cc` counts. Format characters (`Cf`), such as the zero-width space U+200B, are not control
  characters.

## Basic examples

```ant
String.isControl("\n")   // true
String.isControl("\t")   // true
String.isControl("a")    // false
String.isControl("\r\n") // true
```

Call it through `.`:

```ant
"\u{7}".isControl() // true
```

## Scope

The empty string has no code points, so it is not all control characters:

```ant
String.isControl("") // false
```

One code point that is not a control character makes the result `false`, wherever it sits in the
string:

```ant
String.isControl("a\n")      // false
String.isControl("\n\n\n")   // true
String.isControl("\u{1B}[m") // false
```

The C0 controls, the delete character, and the C1 controls are all `Cc`:

```ant
String.isControl("\u{0}")  // true
String.isControl("\u{1F}") // true
String.isControl("\u{7F}") // true
String.isControl("\u{80}") // true
String.isControl("\u{9F}") // true
```

The code points next to them are not. The space is a space separator, and U+00A0 is the no-break
space:

```ant
String.isControl(" ")      // false
String.isControl("\u{A0}") // false
```

Format characters, line separators, and private-use characters have their own categories and are not
`Cc`:

```ant
String.isControl("\u{AD}")   // false
String.isControl("\u{200B}") // false
String.isControl("\u{2028}") // false
String.isControl("\u{E000}") // false
```

A string built with `String.fromCodePoint` works as a literal one does:

```ant
let bell = String.fromCodePoint(7) ?? ""
bell.isControl() // true
```

Test each code point of a string with `String.chars` in a pipe:

```ant
"a\tb\n" | String.chars() | Array.map({ String.isControl(x) }) // [false, true, false, true]
```

## Applications

Remove the control characters from text that came from outside, such as a terminal escape sequence
in a log line. The rest of the sequence is printable and stays:

```ant
func clean(s: string): string {
  String.concat(s.chars().filter({ !String.isControl(x) }))
}

clean("\u{1B}[31mred\u{1B}[0m\u{7}") // "[31mred[0m"
```

Keep the line breaks and tabs of a text, and drop every other control character:

```ant
func keepLayout(c: string): bool {
  !String.isControl(c) || c == "\n" || c == "\t"
}

String.concat("a\u{0}b\tc\n".chars().filter(keepLayout)) // "ab\tc\n"
```

Find the first control character of a name before you store it. `Array.findIndex` gives its index,
or `null` when there is none:

```ant
func badIndex(name: string): int | null {
  name.chars().findIndex({ String.isControl(x) })
}

for name in ["ant", "an\u{0}t", "colony\n"] {
  let i = badIndex(name)
  if i == null {
    println("${name} ok")
  } else {
    println("control character at ${i}")
  }
}
```

```text
ant ok
control character at 2
control character at 6
```

Count the control characters of a text:

```ant
"one\ttwo\r\nthree\n".chars().countWhere({ String.isControl(x) }) // 4
```

## Properties and relations

`String.isControl(s)` is `s` not empty and every code point in the category `"Cc"`:

```ant
for s in ["", "\n", "\t\r", "a\n", "\u{85}", "\u{AD}", " "] {
  let chars = String.chars(s)
  let all = !chars.isEmpty() && chars.every({ String.category(x) == "Cc" })
  assert(String.isControl(s) == all)
}
```

A control character is neither graphic nor printable:

```ant
for n in [0, 9, 10, 13, 27, 127, 128, 133, 159] {
  let c = String.fromCodePoint(n) ?? ""
  assert(String.isControl(c))
  assert(!String.isGraphic(c))
  assert(!String.isPrint(c))
}
```

## Possible issues

A tab, a line feed, and a carriage return are control characters, and white space too. A test that
drops control characters drops line breaks with them:

```ant
String.isControl("\n") // true
String.isSpace("\n")   // true
```

Not every control character is white space. The bell and NUL are not:

```ant
String.isSpace("\u{7}") // false
String.isSpace("\u{0}") // false
```

`String.isControl` tests that the whole string is control characters, not that it holds one. To find
out whether a string contains a control character, test each code point:

```ant
String.isControl("ok\n")                     // false
"ok\n".chars().some({ String.isControl(x) }) // true
```

Invisible characters that a copy from a web page leaves behind, such as the zero-width space and the
soft hyphen, are format characters, `"Cf"`. `String.isControl` does not catch them:

```ant
String.isControl("\u{200B}") // false
String.category("\u{200B}")  // "Cf"
```

## See also

[String.category](category.md) · [String.isSpace](isSpace.md) · [String.isGraphic](isGraphic.md) ·
[String.isPrint](isPrint.md) · [String.chars](chars.md) · [String.fromCodePoint](fromCodePoint.md)
