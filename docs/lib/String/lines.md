<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.lines

```ant signature
func String.lines(s: string): Array[string]
```

String.lines splits s at every \\n, dropping a \\r just before one. The piece after the last \\n is
omitted when it is empty, so a final newline adds no empty line.

<!-- end generated -->

## Details

- A line ends at `\n`. A `\r` right before the `\n` is dropped with it, so text with Windows line
  endings gives the same lines as text with Unix ones.
- A `\r` anywhere else stays in the line. That includes a `\r` at the end of the text.
- The piece after the last `\n` is a line when it is not empty, so text without a final newline
  loses nothing, and a final newline adds no empty line.
- An empty string has no lines, while `"\n"` has one empty line.
- Only `\n` ends a line: U+0085, U+2028, and U+2029 stay in the line.

See [String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string).

## Basic examples

Split text into its lines:

```ant
String.lines("one\ntwo\nthree") // ["one", "two", "three"]
```

A final newline adds no empty line, and `\r\n` ends a line as `\n` does:

```ant
String.lines("a\r\nb\n") // ["a", "b"]
```

## Scope

Text with no `\n` is one line:

```ant
String.lines("one") // ["one"]
```

An empty string has no lines, and a lone newline is one empty line:

```ant
String.lines("")   // []
String.lines("\n") // [""]
```

Empty lines stay, at the start and in the middle:

```ant
String.lines("a\n\nb") // ["a", "", "b"]
String.lines("\n\n\n") // ["", "", ""]
```

Lines may hold any text, other scripts and emoji included:

```ant
String.lines("日本\r\n🐜 ant\n") // ["日本", "🐜 ant"]
```

Call it through `.` or in a pipe:

```ant
"x\ny".lines()          // ["x", "y"]
"x\ny" | String.lines() // ["x", "y"]
```

## Applications

Read a file line by line, with `String.lines` over what `File.read` returns:

<!-- test: file todo.txt
buy milk
call Ann
water plants
-->

```ant
var text = Result.unwrap(File.read("todo.txt"))
for line in String.lines(text) {
  println("- ${line}")
}
```

```text
- buy milk
- call Ann
- water plants
```

Count the lines of a text, whether or not it ends with a newline:

```ant
len(String.lines("a\nb\nc\n")) // 3
len(String.lines("a\nb\nc"))   // 3
```

Number the lines, starting at 1:

```ant
func numbered(text: string): string {
  var out: Array[string] = []
  for line in String.lines(text) {
    out.push("${len(out) + 1}: ${line}")
  }
  String.join(out, "\n")
}

numbered("alpha\nbeta\n") // "1: alpha\n2: beta"
```

Parse `key=value` settings, skipping blank lines and comments:

```ant
func settings(text: string): Map[string, string] {
  var m: Map[string, string] = Map{}
  for line in String.lines(text) {
    var t = String.trim(line)
    if t == "" || String.startsWith(t, "#") {
      continue
    }
    var kv = String.cut(t, "=")
    if kv != null {
      m[String.trim(kv.before)] = String.trim(kv.after)
    }
  }
  m
}

settings("# server\nhost = localhost\n\nport=8080\r\n") // Map{ "host": "localhost", "port": "8080" }
```

Keep the lines that match, as `grep` does:

```ant
var log = "INFO start\nERROR disk full\nINFO retry\nERROR disk full\n"
String.lines(log).filter({ String.startsWith(x, "ERROR") }) // ["ERROR disk full", "ERROR disk full"]
```

## Properties and relations

For text that ends with `\n` and holds no `\r`, joining the lines with `\n` and adding the final
newline gives the text back:

```ant
var text = "first\n\nthird\n"
assert(String.join(String.lines(text), "\n") + "\n" == text)
```

For such text, there is one line per `\n`:

```ant
var text = "a\nb\n\nc\n"
assert(len(String.lines(text)) == String.count(text, "\n"))
```

Adding a final newline to text that is not empty and does not end with one does not change the
lines:

```ant
var text = "a\nb"
assert(String.lines(text + "\n") == String.lines(text))
```

For text with no `\r`, `String.lines` is [String.split](split.md) at `\n` without the empty piece
after a final newline:

```ant
var text = "a\nb\n"
String.split(text, "\n") // ["a", "b", ""]
String.lines(text)       // ["a", "b"]
```

A `\r` before a `\n` stays in the pieces of `String.split`:

```ant
String.split("a\r\nb", "\n") // ["a\r", "b"]
String.lines("a\r\nb")       // ["a", "b"]
```

## Possible issues

A `\r` ends no line on its own, so text with old Mac line endings is one line:

```ant
String.lines("a\rb\rc") // ["a\rb\rc"]
```

Only one `\r` before a `\n` is dropped, and a `\r` at the very end stays:

```ant
String.lines("a\r\r\nb\r") // ["a\r", "b\r"]
```

Only the empty piece after the last `\n` is omitted. Text that ends in a blank line keeps one empty
line:

```ant
String.lines("a\n\n") // ["a", ""]
```

The Unicode line separators U+2028 and U+2029 and the next-line character U+0085 do not end a line:

```ant
String.lines("a\u{2028}b") // ["a\u{2028}b"]
```

`String.lines` does not trim lines. Trim each one when the white space around it does not matter:

```ant
String.lines("  a \n b").map({ String.trim(x) }) // ["a", "b"]
```

## See also

[String.split](split.md), [String.fields](fields.md), [String.join](join.md),
[String.trim](trim.md), [File.read](../File/read.md), [File.readLine](../File/readLine.md)
