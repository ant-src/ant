<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.fields

```ant signature
func String.fields(s: string): Array[string]
```

String.fields returns the runs of s between the code points [`String.isSpace`](isSpace.md) accepts,
or \[\] when there are none.

<!-- end generated -->

## Details

- Splits at runs of Unicode white space, not only ASCII spaces, tabs, and line breaks. No field is
  empty, even with white space at the start or end.
- A string that is empty or holds only white space gives `[]`.

## Basic examples

Split a sentence into words:

```ant
String.fields("the quick brown fox") // ["the", "quick", "brown", "fox"]
```

Runs of spaces, tabs, and line breaks count as one separator, and white space at the ends is
dropped:

```ant
String.fields("  a \t b\n") // ["a", "b"]
```

## Scope

A string with no white space is one field:

```ant
"one".fields() // ["one"]
```

An empty string and a string of only white space have no fields:

```ant
"".fields()          // []
"   \n\t  ".fields() // []
```

White space beyond ASCII separates fields too, such as the no-break space U+00A0 and the ideographic
space U+3000:

```ant
"a\u{A0}b\u{3000}c".fields() // ["a", "b", "c"]
```

Letters of any script, and emoji, stay in their fields:

```ant
"Ünïcödé  日本 語 🐜".fields() // ["Ünïcödé", "日本", "語", "🐜"]
```

Call it through `.` or in a pipe:

```ant
"x y  z".fields()          // ["x", "y", "z"]
"x y  z" | String.fields() // ["x", "y", "z"]
```

## Applications

Count the words of a text:

```ant
let text = "It was the best of times,\nit was the worst of times."
text.fields().length() // 12
```

Collapse every run of white space to a single space:

```ant
func squish(s: string): string {
  s.fields() | String.join(" ")
}

squish("  too   many\tspaces \n") // "too many spaces"
```

Read a line of numbers separated by any amount of white space:

```ant
let line = " 3  14\t15 "
line.fields().map({ x.toInt() ?? 0 }) // [3, 14, 15]
```

Read a command and its arguments from a line of input:

<!-- test: stdin
  move   north 3

look
-->

```ant
type Command = { name: string, args: Array[string] }

func parse(line: string): Command | null {
  let words = line.fields()
  if words.isEmpty() {
    return null
  }
  Command{ name: words[0], args: words.slice(1, words.length()) }
}

let input = Result.unwrap(File.readAll(Os.stdin()))
for line in input.lines() {
  let c = parse(line)
  if c == null {
    println("empty line")
  } else {
    println("${c.name} ${c.args}")
  }
}
```

```text
move ["north", "3"]
empty line
look []
```

## Properties and relations

No field is empty, and no field holds a code point that `String.isSpace` accepts:

```ant
let fields = " a\u{A0}bc \t d\n".fields()
assert(fields.every({ !x.isEmpty() }))
assert(fields.every({ x.chars().every({ !x.isSpace() }) }))
```

Joining the fields with one space and taking the fields again gives the same fields:

```ant
let fields = "  to be\tor  not \n".fields()
assert(String.join(fields, " ").fields() == fields)
```

White space at the ends does not change the fields, so trimming first is not needed:

```ant
let s = "\t one two  "
assert(s.fields() == s.trim().fields())
```

## Possible issues

`String.fields` never gives an empty field, so it loses empty columns. Use [String.split](split.md)
when every separator counts:

```ant
"a  b".fields()   // ["a", "b"]
"a  b".split(" ") // ["a", "", "b"]
```

Only white space separates fields. Punctuation stays in them:

```ant
"red, green,blue".fields() // ["red,", "green,blue"]
```

The zero-width space U+200B is not white space, so it does not separate fields:

```ant
"a\u{200B}b c".fields() // ["a\u{200b}b", "c"]
```

Quotes do not group words, so a quoted argument with a space in it becomes two fields:

```ant
"say \"hello there\"".fields() // ["say", "\"hello", "there\""]
```

## See also

[String.split](split.md), [String.lines](lines.md), [String.trim](trim.md),
[String.isSpace](isSpace.md), [String.join](join.md), [String.cut](cut.md)
