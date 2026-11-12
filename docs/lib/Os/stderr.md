<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.stderr

```ant signature
func Os.stderr(): File.Handle
```

Os.stderr returns the program's standard error, which [`eprint`](../prelude/eprint.md) and
[`eprintln`](../prelude/eprintln.md) write to as well.

<!-- end generated -->

## Details

- The handle is not buffered, so each write reaches standard error at once, in order with what
  [eprint](../prelude/eprint.md), [eprintln](../prelude/eprintln.md) and standard output write.
- Unlike `eprintln`, the `File` functions on the handle return a `Result`, so a failed write is an
  `Err` the program can see. A write whose reader has gone, a broken pipe, ends the program as
  SIGPIPE does, and a shell reports status 141. Only on Windows, or under a host that gives the
  program its own writer, is it an `Err`.
- [File.close](../File/close.md) on the handle does nothing and is `Ok`; standard error stays open.

## Basic examples

Write a line to standard error. Standard output holds only what [println](../prelude/println.md)
writes:

```ant
println("total: 42")
try File.writeLine(Os.stderr(), "warning: 3 rows skipped")
```

```text
total: 42
```

Call the `File` functions on the handle through `.`:

```ant
let err = Os.stderr()
try err.writeText("checking... ")
try err.writeLine("done")
```

## Scope

### Text, lines, and bytes

[File.writeText](../File/writeText.md) writes text as it is, [File.writeLine](../File/writeLine.md)
adds a newline, and [File.writeChunk](../File/writeChunk.md) writes bytes:

```ant
let err = Os.stderr()
try err.writeText("no newline")
try err.writeChunk("\n".toBytes())
try err.writeLine("one line")
```

Standard error then holds:

```text
no newline
one line
```

### A handle like any other

A function that writes to a `File.Handle` writes to standard error, standard output, or a file
alike:

```ant
func report(out: File.Handle, failed: int): Result[void, Error] {
  out.writeLine("${failed} checks failed")
}

try report(Os.stdout(), 0)
try report(Os.stderr(), 2)
```

```text
0 checks failed
```

### Buffered

[File.buffered](../File/buffered.md) collects many small writes and writes them out at
[File.flush](../File/flush.md):

```ant
let log = File.buffered(Os.stderr())
for i in Iter.range(0, 1000) {
  try log.writeLine("step ${i}")
}
try log.flush()
```

## Applications

### Route messages by level

An `if` expression picks the stream, so errors go to standard error and the rest to standard output:

```ant
func log(level: string, msg: string): Result[void, Error] {
  let out = if level == "info" { Os.stdout() } else { Os.stderr() }
  out.writeLine("[${level}] ${msg}")
}

try log("info", "server started")
try log("error", "port 8080 in use")
```

```text
[info] server started
```

### Color only on a terminal

[File.isTerminal](../File/isTerminal.md) tells whether a person reads standard error, so escape
codes stay out of a log file:

<!-- test: check-only whether standard error is a terminal depends on how the program runs -->

```ant
let label = if Os.stderr().isTerminal() { "\u{1b}[31merror\u{1b}[0m" } else { "error" }
try Os.stderr().writeLine("${label}: config not found")
```

### Merge a child's error output

A command started with [Cmd.start](../Cmd/start.md) writes to the program's standard error by
default. Pass `stderr: Os.stdout()` to merge both outputs, as `2>&1` does in a shell:

```ant
let c = Cmd.new("sh", ["-c", "echo out; echo err >&2"])
let p = try c.start(stderr: Os.stdout())
let status = try p.wait()
println("status ${status}")
```

```text
out
err
status 0
```

## Properties and relations

Closing standard error and flushing it both do nothing and are `Ok`, and the stream stays open:

```ant
assert(File.close(Os.stderr()) is Ok)
assert(File.flush(Os.stderr()) is Ok)
assert(Os.stderr().writeLine("still open") is Ok)
```

## Possible issues

### `Os.stderr` is a function

Call it to get the handle; the function itself is not a `File.Handle`:

```ant
let r = File.writeLine(Os.stderr, "oops")
// error: cannot use func(): File.Handle where File.Handle is expected
```

### The `File` functions take text

[File.writeLine](../File/writeLine.md) takes a `string`, while `eprintln` takes any value.
Interpolate a value to write it:

```ant
let r = File.writeLine(Os.stderr(), 42) // error: cannot use int where string is expected
```

```ant
let count = 42
try Os.stderr().writeLine("${count}")
```

### A dropped `Result` hides a failed write

A write called as a statement compiles, and its `Result` is dropped with any error in it. Use `try`
when the write matters, or [eprintln](../prelude/eprintln.md) when it does not:

```ant
Os.stderr().writeLine("ignored if it fails")
try Os.stderr().writeLine("checked")
```

### An unflushed buffer is lost

A buffered handle writes only when full, at [File.flush](../File/flush.md) and at
[File.close](../File/close.md). What it holds when the run ends is lost, so close it at the end:

```ant
let log = File.buffered(Os.stderr())
try log.writeLine("finished")
try log.close()
```

### `eprintln` goes past a buffered handle

[eprint](../prelude/eprint.md), [eprintln](../prelude/eprintln.md) and writes to `Os.stderr()`
itself go past the buffer of `File.buffered(Os.stderr())`, so they can come out before what the
buffer holds:

```ant
let log = File.buffered(Os.stderr())
try log.writeLine("buffered")
eprintln("direct")
try log.flush()
```

Standard error then holds `direct` before `buffered`:

```text
direct
buffered
```

Flush the buffered handle before writing to the stream directly, or write everything through the
buffered handle.

## See also

[eprintln](../prelude/eprintln.md), [eprint](../prelude/eprint.md), [Os.stdout](stdout.md),
[Os.stdin](stdin.md), [File.writeLine](../File/writeLine.md), [File.buffered](../File/buffered.md),
[File.isTerminal](../File/isTerminal.md), [Cmd.start](../Cmd/start.md)
