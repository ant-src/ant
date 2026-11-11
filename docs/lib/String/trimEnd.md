<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.trimEnd

```ant signature
func String.trimEnd(s: string): string
```

String.trimEnd returns s without the trailing code points [`String.isSpace`](isSpace.md) accepts.

<!-- end generated -->

## Details

- Removes all white space at the end; white space at the start and in the middle stays.
- White space is Unicode white space, as [`String.isSpace`](isSpace.md) accepts, so the no-break
  space U+00A0 and the ideographic space U+3000 are removed too.
- A string of only white space gives `""`.

## Basic examples

```ant
String.trimEnd("hello   ") // "hello"
String.trimEnd("  hi  ")   // "  hi"
```

Call it through `.` on any string:

```ant
let line = "total: 42\r\n"
line.trimEnd() // "total: 42"
```

## Scope

The empty string, a string with nothing to remove, and a string of only white space:

```ant
String.trimEnd("")        // ""
String.trimEnd("ant")     // "ant"
String.trimEnd(" \t\n\r") // ""
```

Space between words stays; only the end changes:

```ant
String.trimEnd("a  b  ") // "a  b"
```

Every kind of white space at the end goes, the vertical tab, the form feed, the no-break space, and
the ideographic space among them:

```ant
String.trimEnd("x\u{B}\u{C}\u{A0}\u{3000}") // "x"
```

Text in any script stays as it is, and so does a combining accent right before the white space: the
five code points of `"cafe\u{301}"` remain:

```ant
String.trimEnd("東京 ")                 // "東京"
String.trimEnd("cafe\u{301} ").length() // 5
```

In a pipeline, the string is the first argument:

```ant
"done\n" | String.trimEnd() // "done"
```

The parameter is named `s`, so the call can name it:

```ant
String.trimEnd(s: "ant  ") // "ant"
```

## Applications

Strip the trailing spaces from each line of a text, as an editor does on save:

```ant
let text = "func f() {  \n  return 1 \n}\t"
String.join(String.lines(text).map({ x.trimEnd() }), "\n") // "func f() {\n  return 1\n}"
```

Compare two versions of a text and ignore trailing white space, as `diff -b` does at line ends:

```ant
func clean(s: string): Array[string] {
  String.lines(s).map({ x.trimEnd() })
}

func sameText(a: string, b: string): bool {
  clean(a) == clean(b)
}

sameText("one  \ntwo\n", "one\ntwo \n") // true
sameText("one\ntwo", "one\n two")       // false
```

Right-align a column of values that came with padding on the right:

```ant
let cells = ["12   ", "3    ", "456  "]
cells.map({ x.trimEnd().padStart(5) }) // ["   12", "    3", "  456"]
```

## Properties and relations

`String.trimEnd` never adds anything: `s` starts with its result, and the part it removed is all
white space:

```ant
for s in ["ant  ", "  ant", " ", "", "a b\t\n"] {
  let t = s.trimEnd()
  assert(s.startsWith(t))
  let rest = s.slice(t.length(), s.length())
  assert(rest == "" || rest.isSpace())
}
```

Trimming twice is the same as trimming once:

```ant
let s = "x \t "
assert(s.trimEnd().trimEnd() == s.trimEnd())
```

[`String.trimStart`](trimStart.md) after `String.trimEnd` is [`String.trim`](trim.md), in either
order:

```ant
for s in ["  a b  ", "\ta", "b\n", "", "   "] {
  assert(s.trimEnd().trimStart() == s.trim())
  assert(s.trimStart().trimEnd() == s.trim())
}
```

## Possible issues

`String.trimEnd` leaves white space at the start. Use [`String.trim`](trim.md) to remove both ends:

```ant
String.trimEnd("  ant  ") // "  ant"
String.trim("  ant  ")    // "ant"
```

It removes white space only. To remove a given ending, such as a period or a `/`, use
[`String.trimSuffix`](trimSuffix.md):

```ant
String.trimEnd("end.")         // "end."
String.trimSuffix("end.", ".") // "end"
```

It removes every white space code point at the end, not one line break. A blank last line goes too,
along with trailing spaces you may have meant to keep:

```ant
String.trimEnd("a  \n\n") // "a"
```

The zero-width space U+200B and the byte order mark U+FEFF are not white space, so they stay:

```ant
String.trimEnd("x\u{200B}").length() // 2
String.trimEnd("x\u{FEFF}").length() // 2
```

The function takes no set of characters to remove; a second argument is a compile error:

```ant
println(String.trimEnd("x--", "-")) // error: String.trimEnd takes 1 argument, got 2
```

## See also

[String.trim](trim.md) · [String.trimStart](trimStart.md) · [String.trimSuffix](trimSuffix.md) ·
[String.isSpace](isSpace.md) · [String.fields](fields.md) · [String.lines](lines.md) ·
[String.padEnd](padEnd.md)
