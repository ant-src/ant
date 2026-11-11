<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.trim

```ant signature
func String.trim(s: string): string
```

String.trim returns s without the leading and trailing code points [`String.isSpace`](isSpace.md)
accepts.

<!-- end generated -->

## Details

- Removes all white space at both ends; white space in the middle stays.
- White space is Unicode white space, as [`String.isSpace`](isSpace.md) accepts, so the no-break
  space U+00A0 and the ideographic space U+3000 are removed too.
- A string of only white space gives `""`.

## Basic examples

Remove the spaces around a word:

```ant
String.trim("  hello  ") // "hello"
```

Tabs and line breaks at the ends go too, such as the line break at the end of a line of input:

```ant
String.trim("\tname: Ann\r\n") // "name: Ann"
```

## Scope

White space inside the string stays:

```ant
"  two  words  ".trim() // "two  words"
```

A string with no white space at its ends comes back unchanged:

```ant
"ant".trim()  // "ant"
"a\tb".trim() // "a\tb"
```

An empty string, and a string of only white space, give an empty string:

```ant
"".trim()         // ""
" \t\n\r ".trim() // ""
```

White space beyond ASCII is trimmed as well, such as the no-break space U+00A0, the ideographic
space U+3000, and the line separator U+2028:

```ant
"\u{A0}\u{3000}日本\u{2028}".trim() // "日本"
```

Text of any script, and emoji, keep every code point that is not white space:

```ant
" 🐜 Ünïcödé ".trim() // "🐜 Ünïcödé"
```

Call it through `.`, in a pipe, with the argument named, or inside a lambda:

```ant
"  x  ".trim()                        // "x"
"  x  " | String.trim()               // "x"
String.trim(s: "  x  ")               // "x"
[" a", "b ", " c "].map({ x.trim() }) // ["a", "b", "c"]
```

## Applications

Read answers from standard input and ignore the spaces around them:

<!-- test: stdin
  yes
NO
  maybe
-->

```ant
let input = Os.stdin().readAll().unwrap()
for line in input.lines() {
  let answer = line.trim().toLower()
  let reply = if answer == "yes" {
    "accepted"
  } else if answer == "no" {
    "declined"
  } else {
    "unknown: ${answer}"
  }
  println(reply)
}
```

```text
accepted
declined
unknown: maybe
```

Skip blank lines and comment lines in a configuration text:

```ant
let text = "  # settings\n\n  port = 8080  \n   \nhost = example.org\n"
let kept = text
  | String.lines()
  | Array.map({ x.trim() })
  | Array.filter({ x != "" && !x.startsWith("#") })
println(kept)
```

```text
["port = 8080", "host = example.org"]
```

Read `key = value` pairs, trimming the sides that `String.cut` leaves:

```ant
type Setting = { key: string, value: string }

func parse(line: string): Setting | null {
  let c = line.cut("=")
  if c == null {
    return null
  }
  Setting{ key: c.before.trim(), value: c.after.trim() }
}

println(parse("  port =  8080 "))
println(parse("no setting here"))
```

```text
Setting{ key: "port", value: "8080" }
null
```

Check that a required field is not blank:

```ant
func isBlank(s: string): bool {
  s.trim().isEmpty()
}

isBlank("   ")   // true
isBlank(" Ann ") // false
```

## Properties and relations

Trimming twice gives the same string as trimming once, and the result neither starts nor ends with
white space:

```ant
for s in ["", "  ", " a ", "\u{3000}b\n", "c d"] {
  let t = s.trim()
  assert(t.trim() == t)
  assert(s.contains(t))
  if !t.isEmpty() {
    assert(!t.slice(0, 1).isSpace())
    assert(!t.slice(t.length() - 1, t.length()).isSpace())
  }
}
```

`String.trim` is [`String.trimStart`](trimStart.md) and [`String.trimEnd`](trimEnd.md) together, in
either order:

```ant
let s = "\t  both ends \n"
assert(s.trim() == s.trimStart().trimEnd())
assert(s.trim() == s.trimEnd().trimStart())
```

A string is blank exactly when it is empty or every code point is white space:

```ant
for s in ["", " ", "\u{A0}\t", " x "] {
  assert(s.trim().isEmpty() == (s.isEmpty() || s.isSpace()))
}
```

## Possible issues

The zero-width space U+200B and the byte order mark U+FEFF are not white space, so `String.trim`
keeps them:

```ant
"\u{FEFF}x\u{200B}".trim() // "\u{feff}x\u{200b}"
```

`String.trim` removes only white space. Remove other characters, such as quotes, with
[`String.trimPrefix`](trimPrefix.md) and [`String.trimSuffix`](trimSuffix.md):

```ant
"\"quoted\"".trim()                            // "\"quoted\""
"\"quoted\"".trimPrefix("\"").trimSuffix("\"") // "quoted"
```

A pipe binds looser than `==`, so compare a piped trim inside parentheses. Without them, the type
checker reports a compile error:

```ant
let s = "  "
println(s | String.trim() == "")
// error: '|' binds looser than '==', so parenthesize the pipe, as in (a | f()) == b
```

```ant
let s = "  "
(s | String.trim()) == "" // true
```

## See also

[String.trimStart](trimStart.md) · [String.trimEnd](trimEnd.md) · [String.trimPrefix](trimPrefix.md)
· [String.trimSuffix](trimSuffix.md) · [String.isSpace](isSpace.md) · [String.fields](fields.md) ·
[String.lines](lines.md)
