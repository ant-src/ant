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

- A line ends at `\n`, and a `\r` right before it is dropped too, so Windows and Unix line endings
  give the same lines. A `\r` anywhere else stays.
- A final newline adds no empty line, and text without one loses nothing.
- `""` has no lines, while `"\n"` has one empty line.

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
let text = Result.unwrap(File.read("todo.txt"))
for line in text.lines() {
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
  let out: Array[string] = []
  for line in text.lines() {
    out.push("${len(out) + 1}: ${line}")
  }
  String.join(out, "\n")
}

numbered("alpha\nbeta\n") // "1: alpha\n2: beta"
```

Parse `key=value` settings, skipping blank lines and comments:

```ant
func settings(text: string): Map[string, string] {
  let m: Map[string, string] = Map{}
  for line in text.lines() {
    let t = line.trim()
    if t.isEmpty() || t.startsWith("#") { continue }
    let kv = t.cut("=") ?? continue
    m[kv.before.trim()] = kv.after.trim()
  }
  m
}

settings("# server\nhost = localhost\n\nport=8080\r\n") // Map{ "host": "localhost", "port": "8080" }
```

Keep the lines that match, as `grep` does:

```ant
let log = "INFO start\nERROR disk full\nINFO retry\nERROR disk full\n"
log.lines().filter({ x.startsWith("ERROR") }) // ["ERROR disk full", "ERROR disk full"]
```

## Properties and relations

For text that ends with `\n` and holds no `\r`, joining the lines with `\n` and adding the final
newline gives the text back:

```ant
let text = "first\n\nthird\n"
assert(String.join(text.lines(), "\n") + "\n" == text)
```

For such text, there is one line per `\n`:

```ant
let text = "a\nb\n\nc\n"
assert(text.lines().length() == text.count("\n"))
```

Adding a final newline to text that is not empty and does not end with one does not change the
lines:

```ant
let text = "a\nb"
assert((text + "\n").lines() == text.lines())
```

For text with no `\r`, `String.lines` is [String.split](split.md) at `\n` without the empty piece
after a final newline:

```ant
let text = "a\nb\n"
text.split("\n") // ["a", "b", ""]
text.lines()     // ["a", "b"]
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
String.lines("  a \n b").map({ x.trim() }) // ["a", "b"]
```

## See also

[String.split](split.md), [String.fields](fields.md), [String.join](join.md),
[String.trim](trim.md), [File.read](../File/read.md), [File.readLine](../File/readLine.md)
