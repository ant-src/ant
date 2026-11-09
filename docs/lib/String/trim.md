<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.trim

```ant signature
func String.trim(s: string): string
```

String.trim returns s without leading and trailing Unicode white space.

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
String.trim("  two  words  ") // "two  words"
```

A string with no white space at its ends comes back unchanged:

```ant
String.trim("ant")  // "ant"
String.trim("a\tb") // "a\tb"
```

An empty string, and a string of only white space, give an empty string:

```ant
String.trim("")         // ""
String.trim(" \t\n\r ") // ""
```

White space beyond ASCII is trimmed as well, such as the no-break space U+00A0, the ideographic
space U+3000, and the line separator U+2028:

```ant
String.trim("\u{A0}\u{3000}日本\u{2028}") // "日本"
```

Text of any script, and emoji, keep every code point that is not white space:

```ant
String.trim(" 🐜 Ünïcödé ") // "🐜 Ünïcödé"
```

Call it through `.`, in a pipe, with the argument named, or inside a lambda:

```ant
"  x  ".trim()                               // "x"
"  x  " | String.trim()                      // "x"
String.trim(s: "  x  ")                      // "x"
Array.map([" a", "b ", " c "], { x.trim() }) // ["a", "b", "c"]
```

## Applications

Read answers from standard input and ignore the spaces around them:

<!-- test: stdin
  yes
NO
  maybe
-->

```ant
var input = Result.unwrap(File.readAll(Os.stdin()))
for line in String.lines(input) {
  var answer = String.toLower(String.trim(line))
  if answer == "yes" {
    println("accepted")
  } else if answer == "no" {
    println("declined")
  } else {
    println("unknown: ${answer}")
  }
}
```

```text
accepted
declined
unknown: maybe
```

Skip blank lines and comment lines in a configuration text:

```ant
var text = "  # settings\n\n  port = 8080  \n   \nhost = example.org\n"
var kept = String.lines(text)
  | Array.map({ String.trim(x) })
  | Array.filter({ x != "" && !String.startsWith(x, "#") })
println(kept)
```

```text
["port = 8080", "host = example.org"]
```

Read `key = value` pairs, trimming the sides that `String.cut` leaves:

```ant
type Setting = { key: string, value: string }

func parse(line: string): Setting | null {
  var c = String.cut(line, "=")
  if c == null {
    return null
  }
  Setting{ key: String.trim(c.before), value: String.trim(c.after) }
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
  String.isEmpty(String.trim(s))
}

isBlank("   ")   // true
isBlank(" Ann ") // false
```

## Properties and relations

Trimming twice gives the same string as trimming once, and the result neither starts nor ends with
white space:

```ant
for s in ["", "  ", " a ", "\u{3000}b\n", "c d"] {
  var t = String.trim(s)
  assert(String.trim(t) == t)
  assert(String.contains(s, t))
  if !String.isEmpty(t) {
    assert(!String.isSpace(String.slice(t, 0, 1)))
    assert(!String.isSpace(String.slice(t, String.length(t) - 1, String.length(t))))
  }
}
```

`String.trim` is [`String.trimStart`](trimStart.md) and [`String.trimEnd`](trimEnd.md) together, in
either order:

```ant
var s = "\t  both ends \n"
assert(String.trim(s) == String.trimEnd(String.trimStart(s)))
assert(String.trim(s) == String.trimStart(String.trimEnd(s)))
```

A string is blank exactly when it is empty or every code point is white space:

```ant
for s in ["", " ", "\u{A0}\t", " x "] {
  assert(String.isEmpty(String.trim(s)) == (String.isEmpty(s) || String.isSpace(s)))
}
```

## Possible issues

The zero-width space U+200B and the byte order mark U+FEFF are not white space, so `String.trim`
keeps them:

```ant
String.trim("\u{FEFF}x\u{200B}") // "\u{feff}x\u{200b}"
```

`String.trim` removes only white space. Remove other characters, such as quotes, with
[`String.trimPrefix`](trimPrefix.md) and [`String.trimSuffix`](trimSuffix.md):

```ant
String.trim("\"quoted\"")                                      // "\"quoted\""
String.trimSuffix(String.trimPrefix("\"quoted\"", "\""), "\"") // "quoted"
```

A pipe binds looser than `==`, so compare a piped trim inside parentheses. Without them, the type
checker reports a compile error:

<!-- test: error binds looser than -->

```ant
var s = "  "
println(s | String.trim() == "")
```

```ant
var s = "  "
(s | String.trim()) == "" // true
```

## See also

[String.trimStart](trimStart.md) · [String.trimEnd](trimEnd.md) · [String.trimPrefix](trimPrefix.md)
· [String.trimSuffix](trimSuffix.md) · [String.isSpace](isSpace.md) · [String.fields](fields.md) ·
[String.lines](lines.md)
