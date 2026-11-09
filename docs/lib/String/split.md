<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.split

```ant signature
func String.split(s: string, sep: string): Array[string]
```

String.split returns the pieces of s between each occurrence of sep. It panics if sep is empty.

<!-- end generated -->

## Details

- `sep` matches code point by code point, as `==` compares, with no case folding or normalization.
- Occurrences of `sep` are taken from the left and never overlap.

## Basic examples

Split a comma-separated list:

```ant
String.split("red,green,blue", ",") // ["red", "green", "blue"]
```

Two commas in a row:

```ant
String.split("a,b,,c", ",") // ["a", "b", "", "c"]
```

## Scope

A string without the separator, and the empty string:

```ant
String.split("abc", ",") // ["abc"]
String.split("", ",")    // [""]
```

A separator at either end:

```ant
String.split(",a,", ",")   // ["", "a", ""]
String.split(",", ",")     // ["", ""]
String.split("abc", "abc") // ["", ""]
```

A separator can be longer than one character:

```ant
String.split("a::b::c", "::")        // ["a", "b", "c"]
String.split("a\r\nb", "\r\n")       // ["a", "b"]
String.split("one and two", " and ") // ["one", "two"]
```

Unicode text and separators work like any other:

```ant
String.split("🐜·🐝·🐞", "·") // ["🐜", "🐝", "🐞"]
String.split("日本,語", ",")  // ["日本", "語"]
```

Occurrences are taken from the left without overlap, so `"aaa"` holds one `"aa"`, not two:

```ant
String.split("aaa", "aa")  // ["", "a"]
String.split("aaaa", "aa") // ["", "", ""]
```

Here `"a"` does not match `"A"`, and the combining accent U+0301 on its own is the separator:

```ant
String.split("A,b", "a")             // ["A,b"]
String.split("e\u{301}x", "\u{301}") // ["e", "x"]
```

Call it through `.`, in a pipe, or with named arguments:

```ant
"a-b".split("-")                 // ["a", "b"]
"a-b" | String.split("-")        // ["a", "b"]
String.split(s: "a-b", sep: "-") // ["a", "b"]
```

## Applications

Read a line of comma-separated numbers, with `?? 0` for a field that is not a number:

```ant
String.split("3,14,x,15", ",").map({ String.toInt(x) ?? 0 }) // [3, 14, 0, 15]
```

Take a path apart and drop the empty pieces that the slashes at the ends leave:

```ant
String.split("/usr/local/bin/", "/").filter({ !String.isEmpty(x) }) // ["usr", "local", "bin"]
```

Parse rows of a simple table into records, checking the number of columns:

```ant
type Person = { name: string, age: int }

func parse(row: string): Person | null {
  var cols = String.split(row, ";")
  if len(cols) != 2 {
    return null
  }
  var age = String.toInt(String.trim(cols[1])) ?? return null
  Person{ name: String.trim(cols[0]), age }
}

for row in ["Ann; 31", "Bob;27", "broken", "Eve;x"] {
  println(parse(row))
}
```

```text
Person{ name: "Ann", age: 31 }
Person{ name: "Bob", age: 27 }
null
null
```

Count how often each tag occurs in a comma-separated list:

```ant
var counts: Map[string, int] = Map{}
for tag in String.split("ant,bee,ant,ant,bee", ",") {
  counts[tag] = (counts[tag] ?? 0) + 1
}
println(counts)
```

```text
Map{ "ant": 3, "bee": 2 }
```

## Properties and relations

For this `s`, joining the pieces with `","` gives `s` back, and no piece contains `","`:

```ant
var s = ",a,,b,"
var pieces = String.split(s, ",")
assert(String.join(pieces, ",") == s)
assert(Array.every(pieces, { !String.contains(x, ",") }))
```

For each of these strings, there is one piece more than there are commas:

```ant
for s in ["", "a", "a,b", ",,"] {
  assert(len(String.split(s, ",")) == String.count(s, ",") + 1)
}
```

The first piece is the `before` of [`String.cut`](cut.md), when the separator occurs:

```ant
var s = "k=v=w"
var c = String.cut(s, "=")
assert(c != null)
assert(String.split(s, "=")[0] == c.before)
```

## Possible issues

An empty separator panics. Use [String.chars](chars.md) to take a string apart into code points:

```ant
String.split("abc", "") // panic: String.split needs a separator that is not empty
String.chars("abc")     // ["a", "b", "c"]
```

The pieces keep any space next to the separator. Trim each piece, or split at `", "` when the input
always has the space:

```ant
String.split("a, b, c", ",")                         // ["a", " b", " c"]
String.split("a, b, c", ",").map({ String.trim(x) }) // ["a", "b", "c"]
String.split("a, b, c", ", ")                        // ["a", "b", "c"]
```

The empty string gives one empty piece, not an empty array:

```ant
String.split("", ",")      // [""]
len(String.split("", ",")) // 1
```

Here the line break at the end, as a file has it, gives an empty last piece.
[String.lines](lines.md) drops it, and drops a `\r` before each `\n` too:

```ant
String.split("a\r\nb\n", "\n") // ["a\r", "b", ""]
String.lines("a\r\nb\n")       // ["a", "b"]
```

Splitting at one space leaves an empty piece at each end and between the two spaces.
[String.fields](fields.md) splits at runs of any white space:

```ant
String.split(" a  b ", " ") // ["", "a", "", "b", ""]
String.fields(" a  b ")     // ["a", "b"]
```

To split only at the first separator, as in `key=value` text where the value can hold the separator,
use [String.cut](cut.md):

```ant
String.split("k=v=w", "=") // ["k", "v", "w"]
String.cut("k=v=w", "=")   // String.Cut{ before: "k", after: "v=w" }
```

Splitting CSV at `","` breaks a quoted field that holds a comma. Use [Csv.parse](../Csv/parse.md)
for real CSV:

```ant
String.split("a,\"b,c\"", ",") // ["a", "\"b", "c\""]
Csv.parse("a,\"b,c\"")         // Ok{ value: [["a", "b,c"]] }
```

To split at a pattern rather than a fixed string, use [Regex.split](../Regex/split.md).

## See also

[String.join](join.md) · [String.cut](cut.md) · [String.fields](fields.md) ·
[String.lines](lines.md) · [String.chars](chars.md) · [String.count](count.md) ·
[Regex.split](../Regex/split.md) · [Csv.parse](../Csv/parse.md)
