<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.stdin

```ant signature
func Os.stdin(): File.Handle
```

Os.stdin returns the program's standard input.

<!-- end generated -->

## Details

- `Os.stdin` returns a [File.Handle](../File/types/Handle.md) open for reading, so every `File` read
  works on it: [File.readLine](../File/readLine.md), [File.readAll](../File/readAll.md),
  [File.readChunk](../File/readChunk.md), [File.pipe](../File/pipe.md).
- At the end of the input, `readLine` and `readChunk` give `null` and `readAll` gives `""`. In the
  REPL, standard input belongs to the REPL: `Os.stdin()` is at end of file, and so is the standard
  input of a child the REPL starts with [Cmd.run](../Cmd/run.md).
- [File.close](../File/close.md) on standard input does nothing and is `Ok`, so a function may close
  any handle it reads without asking which one it is.
- Standard input is read only while a read waits, so between reads a child started with
  [Cmd.run](../Cmd/run.md) gets the input. A read whose `ctx` ends returns `Err`, but the stream
  itself cannot be interrupted: what arrives afterwards is kept for the program's next read, not
  given to a child.

## Basic examples

Read all of standard input with [File.readAll](../File/readAll.md):

<!-- test: stdin
one
two
three
-->

```ant
let text = try File.readAll(Os.stdin())
println("${text.lines().length()} lines")
```

```text
3 lines
```

Read it a line at a time with [File.readLine](../File/readLine.md), which gives `null` at the end:

<!-- test: stdin
Ann
Bob
-->

```ant
let input = Os.stdin()
while true {
  let name = try input.readLine() ?? break
  println("Hello, ${name}")
}
```

```text
Hello, Ann
Hello, Bob
```

## Scope

Empty input is at its end from the first read:

```ant
let input = Os.stdin()
println(input.readLine())
println(input.readAll())
```

```text
Ok{ value: null }
Ok{ value: "" }
```

The reads of standard input share one buffer, so a program can read a header line and then the rest
in one piece:

<!-- test: stdin
name,score
Ann,9
Bob,7
-->

```ant
let input = Os.stdin()
let header = try input.readLine() ?? ""
println(header.split(","))
print(try input.readAll())
```

```text
["name", "score"]
Ann,9
Bob,7
```

[File.readChunk](../File/readChunk.md) reads the input as bytes, at most `max` at a time:

<!-- test: stdin
hi!
-->

```ant
let input = Os.stdin()
while true {
  let chunk = try input.readChunk(2) ?? break
  println(chunk)
}
```

```text
Bytes(68 69)
Bytes(21 0a)
```

## Applications

Copy standard input to standard output, as `cat` does, with [File.pipe](../File/pipe.md):

<!-- test: stdin
first
second
-->

```ant
try File.pipe(Os.stdin(), Os.stdout())
```

```text
first
second
```

Ask a question and read the answer. Standard output is not buffered, so the prompt shows before the
read waits:

<!-- test: stdin
Ann
-->

```ant
print("Name: ")
let name = try Os.stdin().readLine() ?? "stranger"
println("Hello, ${name.trim()}")
```

```text
Name: Hello, Ann
```

Read the file named on the command line, or standard input when there is none. The deferred close is
`Ok` for both:

<!-- test: stdin
alpha
beta
-->

```ant
let args = Os.args()
let input = if args.isEmpty() { Os.stdin() } else { try File.open(args[0]) }
defer { input.close() }
var n = 0
while true {
  let line = try input.readLine() ?? break
  n += 1
  println("${n}: ${line}")
}
```

```text
1: alpha
2: beta
```

Count the words of the input:

<!-- test: stdin
the cat saw The dog
the end
-->

```ant
let counts: Map[string, int] = Map{}
for word in (try File.readAll(Os.stdin())).fields() {
  let key = word.toLower()
  counts[key] = (counts[key] ?? 0) + 1
}
println(counts)
```

```text
Map{ "the": 3, "cat": 1, "saw": 1, "dog": 1, "end": 1 }
```

Check whether the input comes from a terminal before reading it, so that a run without piped input
prints a usage line instead of waiting at the terminal. The test's input is not a terminal:

<!-- test: stdin
one
two
-->

```ant
if Os.stdin().isTerminal() {
  eprintln("usage: tool < input")
  Os.exit(2)
}
let text = try File.readAll(Os.stdin())
println("${text.lines().length()} lines")
```

```text
2 lines
```

Keep the lines that contain a word given as an argument, as `grep` does:

<!-- test: args error -->
<!-- test: stdin
ok: started
error: disk full
ok: retried
error: gave up
-->

```ant
let word = Os.args().first() ?? ""
let input = Os.stdin()
while true {
  let line = try input.readLine() ?? break
  if line.contains(word) {
    println(line)
  }
}
```

```text
error: disk full
error: gave up
```

## Properties and relations

Closing standard input leaves it open:

<!-- test: stdin
still here
-->

```ant
let input = Os.stdin()
assert(input.close().isOk())
assert((try input.readLine()) == "still here")
```

## Possible issues

A read gives `string | null`, so the line cannot be used before the end of the input is handled:

```ant
let line = try Os.stdin().readLine()
println(line.toUpper())
// error: string | null has no field or function toUpper on every member: narrow the value first
```

## See also

[Os.stdout](stdout.md) · [Os.stderr](stderr.md) · [Os.args](args.md) ·
[File.readLine](../File/readLine.md) · [File.readAll](../File/readAll.md) ·
[File.readChunk](../File/readChunk.md) · [File.pipe](../File/pipe.md) ·
[File.isTerminal](../File/isTerminal.md) · [Cmd.run](../Cmd/run.md)
