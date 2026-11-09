<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.trimStart

```ant signature
func String.trimStart(s: string): string
```

String.trimStart returns s without leading Unicode white space.

<!-- end generated -->

## Details

- Removes all white space at the start; white space in the middle and at the end stays.
- White space is Unicode white space, as [`String.isSpace`](isSpace.md) accepts, so the no-break
  space U+00A0 and the ideographic space U+3000 are removed too.
- A string of only white space gives `""`.

## Basic examples

Remove the spaces before a word, and keep the ones after it:

```ant
String.trimStart("   hello  ") // "hello  "
```

Tabs and line breaks go too:

```ant
String.trimStart("\t\n  indented") // "indented"
```

## Scope

A string without leading white space comes back unchanged, white space inside it included:

```ant
String.trimStart("a  b ") // "a  b "
```

An empty string, and a string of only white space, give `""`:

```ant
String.trimStart("")        // ""
String.trimStart(" \t\r\n") // ""
```

White space beyond ASCII counts: the no-break space U+00A0, the ideographic space U+3000, the next
line U+0085, and the line separator U+2028:

```ant
String.trimStart("\u{A0}\u{3000}日本") // "日本"
String.trimStart("\u{85}\u{2028}x")    // "x"
```

Call it in a pipe, through `.` on a string, with its argument named, or as a function value:

```ant
"  piped" | String.trimStart()                     // "piped"
"  dotted".trimStart()                             // "dotted"
String.trimStart(s: "  named")                     // "named"
Array.map(["  a", " b ", "c  "], String.trimStart) // ["a", "b ", "c  "]
```

## Applications

Measure the indentation of a line, in code points:

```ant
func indent(line: string): int {
  String.length(line) - String.length(String.trimStart(line))
}

for line in ["func f() {", "  var x = 1", "    return x", "}"] {
  println("${indent(line)}: ${String.trimStart(line)}")
}
```

```text
0: func f() {
2: var x = 1
4: return x
0: }
```

Read `key: value` pairs, where the value may follow the colon after any number of spaces:

```ant
type Header = { key: string, value: string }

func parse(line: string): Header | null {
  var c = String.cut(line, ":")
  if c == null {
    return null
  }
  Header{ key: c.before, value: String.trimStart(c.after) }
}

println(parse("Host:   example.com"))
println(parse("Accept:text/html"))
println(parse("no colon"))
```

```text
Header{ key: "Host", value: "example.com" }
Header{ key: "Accept", value: "text/html" }
null
```

Remove the indentation of every line, and skip the lines that are blank:

```ant
var text = "  first\n\n    second\n   \n  third"
var kept = String.lines(text)
  | Array.map(String.trimStart)
  | Array.filter({ !String.isEmpty(x) })
println(kept)
```

```text
["first", "second", "third"]
```

## Properties and relations

`String.trimStart` keeps the end of `s`: its result is a suffix of `s`, and it is idempotent:

```ant
for s in ["", "  a ", "\t\nb", "c", "   "] {
  var t = String.trimStart(s)
  assert(String.endsWith(s, t))
  assert(String.trimStart(t) == t)
}
```

The result is empty or starts with a code point that is not white space, as
[`String.isSpace`](isSpace.md) tests it:

```ant
for s in ["  a", "\u{3000}b", "   "] {
  var t = String.trimStart(s)
  assert(String.isEmpty(t) || !String.isSpace(String.slice(t, 0, 1)))
}
```

[`String.trimEnd`](trimEnd.md) after `String.trimStart` is [`String.trim`](trim.md):

```ant
var s = "  both ends  "
assert(String.trimEnd(String.trimStart(s)) == String.trim(s))
```

## Possible issues

The zero-width space U+200B and the byte order mark U+FEFF are not white space, so
`String.trimStart` keeps them. Remove a byte order mark with [`String.trimPrefix`](trimPrefix.md):

```ant
String.trimStart("\u{FEFF}  x") == "\u{FEFF}  x"               // true
String.trimStart(String.trimPrefix("\u{FEFF}  x", "\u{FEFF}")) // "x"
```

`String.trimStart` removes white space only. To remove another leading text, use
[`String.trimPrefix`](trimPrefix.md), which removes it once:

```ant
String.trimStart("--flag")        // "--flag"
String.trimPrefix("--flag", "--") // "flag"
```

The right side of a pipe is a call, so it needs its parentheses:

```ant
println("  x" | String.trimStart) // error: the right side of a pipe must be a call
```

`s` is a `string`, so a `string | null` needs a check first:

```ant
var name: string | null = null
println(String.trimStart(name)) // error: cannot use string | null where string is expected
```

## See also

[String.trim](trim.md) · [String.trimEnd](trimEnd.md) · [String.trimPrefix](trimPrefix.md) ·
[String.padStart](padStart.md) · [String.isSpace](isSpace.md) · [String.fields](fields.md)
