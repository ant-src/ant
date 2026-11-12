<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.stdout

```ant signature
func Os.stdout(): File.Handle
```

Os.stdout returns the program's standard output, which [`print`](../prelude/print.md) and
[`println`](../prelude/println.md) write to as well. It is not buffered; File.buffered(Os.stdout())
buffers a program that writes much.

<!-- end generated -->

## Details

- Every write is handed to the OS before it returns, so `print`, `println`, writes to the handle,
  and the output of child programs come out in the order they were written.
- A write whose reader has gone, a broken pipe, ends the program as SIGPIPE does, and a shell
  reports status 141, so `ant run gen.ant | head` stops once `head` has read enough. Where no signal
  comes, as on Windows, `print` and `println` still end the program with status 141, and the `File`
  functions return `Err`.
- [File.close](../File/close.md) on it does nothing and is `Ok`; standard output stays open.

## Basic examples

Write a line to standard output through its handle:

```ant
try Os.stdout().writeLine("Hello, Ant")
```

```text
Hello, Ant
```

`println` writes to the same stream, and the two come out in program order:

```ant
println("first")
try Os.stdout().writeLine("second")
println("third")
```

```text
first
second
third
```

Buffer it for a loop of many small writes, and flush it at the end:

```ant
let out = File.buffered(Os.stdout())
for i in Iter.range(1, 4) {
  try out.writeLine("row ${i}")
}
try out.flush()
```

```text
row 1
row 2
row 3
```

## Scope

[File.writeText](../File/writeText.md) writes without a newline, and
[File.writeChunk](../File/writeChunk.md) writes bytes as they are:

```ant
let out = Os.stdout()
try out.writeText("no newline, ")
try out.writeChunk("then bytes\n".toBytes())
```

```text
no newline, then bytes
```

A function that takes a `File.Handle` writes to standard output or to a file alike. Its final
[File.flush](../File/flush.md) does nothing on standard output and writes out a buffered handle:

```ant
func report(out: File.Handle, counts: Map[string, int]): Result[void, Error] {
  for { key, value } in counts {
    try out.writeLine("${key}: ${value}")
  }
  out.flush()
}

let counts = Map{ "ants": 3, "bees": 1 }
try report(Os.stdout(), counts)
let file = try File.create("counts.txt")
defer { file.close() }
try report(file, counts)
```

```text
ants: 3
bees: 1
```

A child program started with [Cmd.run](../Cmd/run.md) writes to the same standard output, in order
with the program's own writes:

```ant
println("before")
try Cmd.run(Cmd.new("echo", ["from the child"]))
println("after")
```

```text
before
from the child
after
```

## Applications

Copy standard input to standard output, as `cat` does, with [File.pipe](../File/pipe.md):

<!-- test: stdin
one
two
-->

```ant
try File.pipe(Os.stdin(), Os.stdout())
```

```text
one
two
```

Prompt for a reply. Standard output is not buffered, so the prompt shows before the read waits, with
no flush:

<!-- test: stdin
Ada
-->

```ant
try Os.stdout().writeText("Name: ")
let name = try Os.stdin().readLine() ?? "stranger"
println("Hello, ${name}")
```

```text
Name: Hello, Ada
```

Here the reply comes from a pipe, so it does not echo; at a terminal it shows after the prompt.

Color the output only when it goes to a terminal. Here it goes to a pipe, so the text is plain:

```ant
let out = Os.stdout()
let ok = if out.isTerminal() { "\u{1b}[32mok\u{1b}[0m" } else { "ok" }
try out.writeLine("build: ${ok}")
```

```text
build: ok
```

Report a failed write as an error, which `println` does not do:

```ant
func emit(lines: Array[string]): Result[void, Error] {
  let text = lines | String.join("\n")
  Os.stdout().writeLine(text) | Result.wrapErr("write output")
}

try emit(["alpha", "beta"])
```

```text
alpha
beta
```

## Properties and relations

Standard output is not buffered, so a flush does nothing and is `Ok`; a close is `Ok` and leaves it
open:

```ant
let out = Os.stdout()
assert(out.flush() is Ok)
assert(out.close() is Ok)
assert(out.writeLine("still open") is Ok)
```

```text
still open
```

`println(s)` writes what `File.writeLine` on `Os.stdout()` writes:

```ant
let s = "same line"
println(s)
try Os.stdout().writeLine(s)
```

```text
same line
same line
```

## Possible issues

A `println` goes past the buffer of a buffered standard output, so it comes out before what the
buffer holds:

```ant
let out = File.buffered(Os.stdout())
try out.writeLine("written first")
println("printed second")
try out.flush()
```

```text
printed second
written first
```

A buffered standard output that is neither flushed nor closed loses what it holds when the program
ends. This program prints nothing:

```ant
let out = File.buffered(Os.stdout())
try out.writeLine("lost")
```

Close the buffered handle right after making it. The close writes out what the buffer holds and
leaves standard output open:

```ant
let out = File.buffered(Os.stdout())
defer { out.close() }
try out.writeLine("kept")
```

```text
kept
```

## See also

[Os.stderr](stderr.md) · [Os.stdin](stdin.md) · [print](../prelude/print.md) ·
[println](../prelude/println.md) · [File.buffered](../File/buffered.md) ·
[File.writeLine](../File/writeLine.md) · [File.isTerminal](../File/isTerminal.md)
