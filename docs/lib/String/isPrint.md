<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.isPrint

```ant signature
func String.isPrint(s: string): bool
```

String.isPrint reports whether every code point of s is in L, M, N, P, or S, or is the space U+0020.
It is false on "".

<!-- end generated -->

## Details

- `String.isPrint` is `true` when every code point of `s` is a letter (`L`), a mark (`M`), a number
  (`N`), punctuation (`P`), or a symbol (`S`), or is the space U+0020. A single code point tests one
  character; a longer string tests that all of it is printable.
- It returns `false` on the empty string.
- The space U+0020 is the only space it accepts. Every other space separator (`Zs`), such as the
  no-break space U+00A0, is not printable; `String.isGraphic` accepts them.
- Control characters (`Cc`), such as a tab or a line feed, format characters (`Cf`), such as the
  zero-width joiner, line and paragraph separators, private-use, and unassigned code points are not
  printable.
- `String.from` writes every code point that `String.isPrint` rejects inside a quoted string as an
  escape, such as `\u{a0}`, or as `\n`, `\t`, or `\r`.
- The tables are those of the Unicode version that `String.unicodeVersion` names.
- [spec: String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)

## Basic examples

```ant
String.isPrint("hello")       // true
String.isPrint("hello world") // true
String.isPrint("a\tb")        // false
```

Call it through `.`:

```ant
"Ant 1.0!".isPrint() // true
```

## Scope

The empty string has no code points, so it is not all printable:

```ant
String.isPrint("") // false
```

Letters, digits, punctuation, and symbols of any script are printable:

```ant
String.isPrint("日本語")   // true
String.isPrint("½€!")   // true
String.isPrint("Ⅻ + ٣") // true
```

A combining mark is printable, alone or after its base letter:

```ant
String.isPrint("e\u{301}") // true
String.isPrint("\u{301}")  // true
```

The space U+0020 is printable. The no-break space and the ideographic space are space separators
too, but they are not:

```ant
String.isPrint(" ")        // true
String.isPrint("\u{A0}")   // false
String.isPrint("\u{3000}") // false
```

Invisible characters of other categories are not printable: a control character, a format character
such as the zero-width space or a bidirectional override, a line separator, a private-use code
point, and an unassigned one:

```ant
String.isPrint("\u{0}")    // false
String.isPrint("\u{200B}") // false
String.isPrint("\u{202E}") // false
String.isPrint("\u{2028}") // false
String.isPrint("\u{E000}") // false
String.isPrint("\u{378}")  // false
```

An emoji is a symbol. An emoji joined from several holds the zero-width joiner U+200D, which is a
format character:

```ant
String.isPrint("👍")     // true
String.isPrint("👨‍👩‍👧") // false
```

Test each code point of a string with `String.chars` in a pipe:

```ant
"a b\n" | String.chars() | Array.map({ String.isPrint(x) }) // [true, true, true, false]
```

## Applications

Accept a display name only when every code point of it is printable and it has no space at either
end:

```ant
func validName(name: string): bool {
  name.isPrint() && name.trim() == name
}

validName("Queen Ant") // true
validName("ant\u{0}")  // false
validName(" ant")      // false
```

Remove the code points that are not printable from untrusted text, such as before writing it to a
log:

```ant
func clean(s: string): string {
  String.concat(s.chars().filter({ String.isPrint(x) }))
}

clean("pay\u{A0}100\u{200B}€\n") // "pay100€"
```

Show the code point of every character that `String.isPrint` rejects, as `U+` and its hexadecimal
value:

```ant
func reveal(s: string): string {
  var parts = s.chars().map({
    if String.isPrint(x) {
      x
    } else {
      var hex = String.fromInt(String.codePoint(x), 16).toUpper()
      "<U+${hex.padStart(4, "0")}>"
    }
  })
  String.concat(parts)
}

println(reveal("pay\u{A0}100\u{200B}€\n"))
```

```text
pay<U+00A0>100<U+200B>€<U+000A>
```

`String.from` writes a string inside a container with escapes for the code points that
`String.isPrint` rejects, so it shows what a string holds:

```ant
println(String.from(["a\u{A0}b", "tab\there", "日本"]))
```

```text
["a\u{a0}b", "tab\there", "日本"]
```

## Properties and relations

`String.isPrint(s)` is `s` not empty and every code point in the categories `L`, `M`, `N`, `P`, or
`S`, or the space:

```ant
func printable(c: string): bool {
  c == " " || ["L", "M", "N", "P", "S"].contains(String.category(c).slice(0, 1))
}

for s in ["", "abc", "a b", "a\tb", "\u{A0}", "e\u{301}", "½", "\u{200D}", "\u{E000}"] {
  var chars = String.chars(s)
  assert(String.isPrint(s) == (!chars.isEmpty() && chars.every(printable)))
}
```

A printable string is graphic. The two differ only on the space separators other than U+0020:

```ant
for s in ["abc", "a b", "€", "\u{A0}", "\u{3000}", "\u{2009}"] {
  if String.isPrint(s) {
    assert(String.isGraphic(s))
  }
}
assert(String.isGraphic("\u{A0}") && !String.isPrint("\u{A0}"))
```

No control character is printable:

```ant
for n in [0, 9, 10, 13, 27, 127, 133, 159] {
  var c = String.fromCodePoint(n) ?? ""
  assert(String.isControl(c) && !String.isPrint(c))
}
```

## Possible issues

A tab and a line break are not printable, so a text of several lines is not:

```ant
String.isPrint("one\ntwo") // false
String.isPrint("one two")  // true
```

The no-break space looks like a space, and a text pasted from a web page often holds it. It is not
printable:

```ant
String.isPrint("10\u{A0}km")   // false
String.isGraphic("10\u{A0}km") // true
```

An emoji that looks like one character can be several code points joined by U+200D, and the joiner
is not printable:

```ant
String.isPrint("👨‍👩‍👧")                    // false
"👨‍👩‍👧".chars().map({ String.isPrint(x) }) // [true, false, true, false, true]
```

`String.isPrint` tests that the whole string is printable, not that it holds a printable character.
To find out whether a string has one, test each code point:

```ant
String.isPrint("ok\n")                     // false
"ok\n".chars().some({ String.isPrint(x) }) // true
```

## See also

[String.isGraphic](isGraphic.md) · [String.isControl](isControl.md) · [String.isSpace](isSpace.md) ·
[String.category](category.md) · [String.chars](chars.md) · [String.from](from.md)
