<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.trimStart

```ant signature
func String.trimStart(s: string): string
```

String.trimStart returns s without the leading code points [`String.isSpace`](isSpace.md) accepts.

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
"a  b ".trimStart() // "a  b "
```

An empty string, and a string of only white space, give `""`:

```ant
"".trimStart()        // ""
" \t\r\n".trimStart() // ""
```

White space beyond ASCII counts: the no-break space U+00A0, the ideographic space U+3000, the next
line U+0085, and the line separator U+2028:

```ant
"\u{A0}\u{3000}日本".trimStart() // "日本"
"\u{85}\u{2028}x".trimStart()    // "x"
```

Call it in a pipe, through `.` on a string, with its argument named, or as a function value:

```ant
"  piped" | String.trimStart()              // "piped"
"  dotted".trimStart()                      // "dotted"
String.trimStart(s: "  named")              // "named"
["  a", " b ", "c  "].map(String.trimStart) // ["a", "b ", "c  "]
```

## Applications

Measure the indentation of a line, in code points:

```ant
func indent(line: string): int {
  line.length() - line.trimStart().length()
}

for line in ["func f() {", "  var x = 1", "    return x", "}"] {
  println("${indent(line)}: ${line.trimStart()}")
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
  let c = line.cut(":")
  if c == null {
    return null
  }
  Header{ key: c.before, value: c.after.trimStart() }
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
let text = "  first\n\n    second\n   \n  third"
let kept = text
  | String.lines()
  | Array.map(String.trimStart)
  | Array.filter({ !x.isEmpty() })
println(kept)
```

```text
["first", "second", "third"]
```

## Properties and relations

`String.trimStart` keeps the end of `s`: its result is a suffix of `s`, and it is idempotent:

```ant
for s in ["", "  a ", "\t\nb", "c", "   "] {
  let t = s.trimStart()
  assert(s.endsWith(t))
  assert(t.trimStart() == t)
}
```

The result is empty or starts with a code point that is not white space, as
[`String.isSpace`](isSpace.md) tests it:

```ant
for s in ["  a", "\u{3000}b", "   "] {
  let t = s.trimStart()
  assert(t.isEmpty() || !t.slice(0, 1).isSpace())
}
```

[`String.trimEnd`](trimEnd.md) after `String.trimStart` is [`String.trim`](trim.md):

```ant
let s = "  both ends  "
assert(s.trimStart().trimEnd() == s.trim())
```

## Possible issues

The zero-width space U+200B and the byte order mark U+FEFF are not white space, so
`String.trimStart` keeps them. Remove a byte order mark with [`String.trimPrefix`](trimPrefix.md):

```ant
"\u{FEFF}  x".trimStart() == "\u{FEFF}  x"       // true
"\u{FEFF}  x".trimPrefix("\u{FEFF}").trimStart() // "x"
```

`String.trimStart` removes white space only. To remove another leading text, use
[`String.trimPrefix`](trimPrefix.md), which removes it once:

```ant
"--flag".trimStart()      // "--flag"
"--flag".trimPrefix("--") // "flag"
```

The right side of a pipe is a call, so it needs its parentheses:

```ant
println("  x" | String.trimStart) // error: the right side of a pipe must be a call
```

`s` is a `string`, so a `string | null` needs a check first:

```ant
let name: string | null = null
println(String.trimStart(name)) // error: cannot use string | null where string is expected
```

## See also

[String.trim](trim.md) · [String.trimEnd](trimEnd.md) · [String.trimPrefix](trimPrefix.md) ·
[String.padStart](padStart.md) · [String.isSpace](isSpace.md) · [String.fields](fields.md)
