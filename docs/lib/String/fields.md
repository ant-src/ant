<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.fields

```ant signature
func String.fields(s: string): Array[string]
```

String.fields returns the runs of s between Unicode white space, or \[\] when there are none.

<!-- end generated -->

## Details

- A field is a run of code points with no white space in it. A run of white space between two fields
  separates them once, however long it is, so no field is empty.
- White space at the start and at the end of `s` gives no empty field either.
- A string that is empty, or holds only white space, gives `[]`.
- White space is Unicode white space, not only the ASCII space, tab, and line breaks
  ([String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)).

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
String.fields("one") // ["one"]
```

An empty string and a string of only white space have no fields:

```ant
String.fields("")          // []
String.fields("   \n\t  ") // []
```

White space beyond ASCII separates fields too, such as the no-break space U+00A0 and the ideographic
space U+3000:

```ant
String.fields("a\u{A0}b\u{3000}c") // ["a", "b", "c"]
```

Letters of any script, and emoji, stay in their fields:

```ant
String.fields("Ünïcödé  日本 語 🐜") // ["Ünïcödé", "日本", "語", "🐜"]
```

Call it through `.` or in a pipe:

```ant
"x y  z".fields()          // ["x", "y", "z"]
"x y  z" | String.fields() // ["x", "y", "z"]
```

## Applications

Count the words of a text:

```ant
var text = "It was the best of times,\nit was the worst of times."
len(String.fields(text)) // 12
```

Collapse every run of white space to a single space:

```ant
func squish(s: string): string {
  String.join(String.fields(s), " ")
}

squish("  too   many\tspaces \n") // "too many spaces"
```

Read a line of numbers separated by any amount of white space:

```ant
var line = " 3  14\t15 "
String.fields(line).map({ String.toInt(x) ?? 0 }) // [3, 14, 15]
```

Read a command and its arguments from a line of input:

<!-- test: stdin
  move   north 3

look
-->

```ant
type Command = { name: string, args: Array[string] }

func parse(line: string): Command | null {
  var words = String.fields(line)
  if len(words) == 0 {
    return null
  }
  Command{ name: words[0], args: Array.slice(words, 1, len(words)) }
}

var input = Result.unwrap(File.readAll(Os.stdin()))
for line in String.lines(input) {
  var c = parse(line)
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
var fields = String.fields(" a\u{A0}bc \t d\n")
assert(Array.every(fields, { !String.isEmpty(x) }))
assert(Array.every(fields, { Array.every(String.chars(x), { !String.isSpace(x) }) }))
```

Joining the fields with one space and taking the fields again gives the same fields:

```ant
var fields = String.fields("  to be\tor  not \n")
assert(String.fields(String.join(fields, " ")) == fields)
```

White space at the ends does not change the fields, so trimming first is not needed:

```ant
var s = "\t one two  "
assert(String.fields(s) == String.fields(String.trim(s)))
```

## Possible issues

`String.fields` never gives an empty field, so it loses empty columns. Use [String.split](split.md)
when every separator counts:

```ant
String.fields("a  b")     // ["a", "b"]
String.split("a  b", " ") // ["a", "", "b"]
```

Only white space separates fields. Punctuation stays in them:

```ant
String.fields("red, green,blue") // ["red,", "green,blue"]
```

The zero-width space U+200B is not white space, so it does not separate fields:

```ant
String.fields("a\u{200B}b c") // ["a\u{200b}b", "c"]
```

Quotes do not group words, so a quoted argument with a space in it becomes two fields:

```ant
String.fields("say \"hello there\"") // ["say", "\"hello", "there\""]
```

## See also

[String.split](split.md), [String.lines](lines.md), [String.trim](trim.md),
[String.isSpace](isSpace.md), [String.join](join.md), [String.cut](cut.md)
