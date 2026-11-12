<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.writeLine

```ant signature
func File.writeLine(f: File.Handle, line: string): Result[void, Error]
```

File.writeLine writes line and a newline to f.

<!-- end generated -->

## Details

- A `\n` inside `line` is written as it is, so the text reads back as more than one line.
- The write reaches the OS before `File.writeLine` returns, with all of its bytes or `Err`. Only a
  handle from [File.buffered](buffered.md) holds writes back.
- A handle from [File.open](open.md), or one already closed, is `Err`.

## Basic examples

Write a line to a file, then close it:

```ant
let out = try File.create("todo.txt")
defer { out.close() }
try File.writeLine(out, "buy milk")
try out.close()
println(File.read("todo.txt"))
```

```text
Ok{ value: "buy milk\n" }
```

Write a message to standard error:

```ant
try File.writeLine(Os.stderr(), "warning: disk almost full")
println("done")
```

```text
done
```

## Scope

An empty line is a newline alone, and Unicode text is written as UTF-8:

```ant
let out = try File.create("lines.txt")
defer { out.close() }
for line in ["one", "", "héllo 👋"] {
  try out.writeLine(line)
}
try out.close()
println(File.read("lines.txt"))
```

```text
Ok{ value: "one\n\nhéllo 👋\n" }
```

A handle from [File.append](append.md) adds lines at the end of what the file holds:

<!-- test: file app.log
started
-->

```ant
let log = try File.append("app.log")
defer { log.close() }
try log.writeLine("stopped")
try log.close()
print(try File.read("app.log"))
```

```text
started
stopped
```

A connection is a handle too. Here a client sends a line over [Net](../Net/index.md), and the server
reads it with [File.readLine](readLine.md):

```ant
let listener = try Net.listen("127.0.0.1:0")
let client = try Net.dial(listener.address())
defer { client.close() }
let server = try Net.accept(listener)
defer { server.close() }
try client.writeLine("ping")
println(server.readLine())
```

```text
Ok{ value: "ping" }
```

## Applications

Write records as CSV-like rows, one line each, and report a failed close:

```ant
type Score = { name: string, points: int }

func saveScores(path: string, scores: Array[Score]): Result[void, Error] {
  let out = try (File.create(path) | Result.wrapErr("scores"))
  defer { out.close() }
  try out.writeLine("name,points")
  for s in scores {
    try out.writeLine("${s.name},${s.points}")
  }
  out.close()
}

println(saveScores("scores.csv", [Score{ name: "Ann", points: 9 }, Score{ name: "Bob", points: 7 }]))
print(try File.read("scores.csv"))
```

```text
Ok{ value: void }
name,points
Ann,9
Bob,7
```

Buffer many small writes with [File.buffered](buffered.md), and close the buffered handle to write
them out:

```ant
let out = File.buffered(Os.stdout())
for i in Iter.range(1, 4) {
  try out.writeLine("line ${i}")
}
try out.close()
```

```text
line 1
line 2
line 3
```

## Properties and relations

`File.writeLine(f, line)` writes what `File.writeText(f, line + "\n")` writes:

```ant
let a = try File.create("a.txt")
defer { a.close() }
let b = try File.create("b.txt")
defer { b.close() }
for line in ["x", "", "y z"] {
  try a.writeLine(line)
  try b.writeText(line + "\n")
}
try a.close()
try b.close()
assert(File.read("a.txt") == File.read("b.txt"))
```

Lines with no `\n` and no final `\r` read back as written, through
[String.lines](../String/lines.md):

```ant
let lines = ["alpha", "", "gamma"]
let out = try File.create("round.txt")
defer { out.close() }
for line in lines {
  try out.writeLine(line)
}
try out.close()
assert((try File.read("round.txt")).lines() == lines)
```

## Possible issues

A `\n` inside the line splits it, so it reads back as two lines:

```ant
let out = try File.create("split.txt")
defer { out.close() }
try out.writeLine("first\nsecond")
try out.close()
println((try File.read("split.txt")).lines())
```

```text
["first", "second"]
```

A handle from [File.open](open.md) is for reading, and a closed handle takes no more writes:

<!-- test: file notes.txt
note
-->

```ant
let input = try File.open("notes.txt")
defer { input.close() }
println(input.writeLine("more"))
let out = try File.create("out.txt")
try out.close()
println(out.writeLine("late"))
```

```text
Err{ error: Error{ message: "the file is not open for writing" } }
Err{ error: Error{ message: "the file is closed" } }
```

Lines written to a buffered handle are not in the file until the buffer fills, or until a flush or a
close:

```ant
let out = File.buffered(try File.create("held.txt"))
defer { out.close() }
try out.writeLine("held")
println(File.read("held.txt"))
try out.flush()
println(File.read("held.txt"))
```

```text
Ok{ value: "" }
Ok{ value: "held\n" }
```

A write to standard output or standard error whose reader has gone, as when the output is piped to
`head`, ends the program with status 141, as SIGPIPE does. It does not return `Err`. Where no signal
comes, as on Windows, it returns `Err`.

The line must be a string. Convert other values first, with `String.from` or interpolation:

```ant
try File.writeLine(Os.stdout(), 42) // error: cannot use int where string is expected
```

## See also

[File.writeText](writeText.md) · [File.writeChunk](writeChunk.md) · [File.readLine](readLine.md) ·
[File.create](create.md) · [File.append](append.md) · [File.buffered](buffered.md) ·
[File.close](close.md) · [println](../prelude/println.md) · [Os.stderr](../Os/stderr.md)
