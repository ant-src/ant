<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.openPipe

```ant signature
func Os.openPipe(): Result[Os.Pipe, Error]
```

Os.openPipe returns a pipe, two handles the program closes: what is written to write is read from
read. A read waits for data. The reading end is at end of file once every writing end is closed, the
copies given to children and the program's own. A write waits while the pipe is full, so the program
reads it on another goroutine, or gives it to a child, before writing much. A write once every copy
of read is closed is Err with the message "broken pipe", and the program goes on.

<!-- end generated -->

## Details

- A `ctx` on a read of `pipe.read` ends the wait. After [Cmd.start](../Cmd/start.md) gives
  `pipe.write` to a child, close the program's own copy, or `pipe.read` never reaches end of file.
- A child that writes into a pipe nobody reads blocks once the pipe is full. Read the pipe before
  [Cmd.wait](../Cmd/wait.md), or on another goroutine. The program's own write waits the same way.
- A write once every copy of `pipe.read` is closed is `Err` with the message `broken pipe`, and the
  program goes on. Only a broken pipe on standard output or standard error ends the program.

## Basic examples

What is written to `pipe.write` is read from `pipe.read`. Closing the write end makes the next read
see the end:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
try pipe.write.writeLine("hello")
try pipe.write.close()
println(try pipe.read.readLine())
println(try pipe.read.readLine())
```

```text
hello
null
```

Read a child's output line by line, as it writes it:

```ant
let out = try Os.openPipe()
defer { out.read.close() }
let child = try Cmd.start(Cmd.new("seq", ["3"]), stdout: out.write)
try out.write.close()
while true {
  let line = try out.read.readLine() ?? break
  println("got ${line}")
}
println(try child.wait())
```

```text
got 1
got 2
got 3
0
```

## Scope

Write a child's standard input a piece at a time. Here `tr` writes to the program's standard output,
the default, and stops once the write end is closed:

```ant
let input = try Os.openPipe()
defer { input.read.close() }
let child = try Cmd.start(Cmd.new("tr", ["a-z", "A-Z"]), stdin: input.read)
for word in ["ant", "bee", "cat"] {
  try input.write.writeLine(word)
}
try input.write.close()
println(try child.wait())
```

```text
ANT
BEE
CAT
0
```

Give the same write end as `stdout` and `stderr` to merge both outputs into one pipe, as `2>&1`
does:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
let script = "echo out; echo err >&2; echo out again"
let child = try Cmd.start(Cmd.new("sh", ["-c", script]), stdout: pipe.write, stderr: pipe.write)
try pipe.write.close()
print(try pipe.read.readAll())
println(try child.wait())
```

```text
out
err
out again
0
```

Two pipes keep the outputs apart. One goroutine reads standard error while the program reads
standard output, so neither pipe fills up:

```ant
let out = try Os.openPipe()
let err = try Os.openPipe()
defer { out.read.close() }
defer { err.read.close() }
let script = "echo built main; echo 'warning: unused x' >&2; echo built test"
let child = try Cmd.start(Cmd.new("sh", ["-c", script]), stdout: out.write, stderr: err.write)
try out.write.close()
try err.write.close()

let warnings: Channel[string] = Channel.new(1)
go {
  warnings.send(err.read.readAll().unwrap())
}
print(try out.read.readAll())
print("stderr: ${warnings.receive() ?? ""}")
println(try child.wait())
```

```text
built main
built test
stderr: warning: unused x
0
```

A pipeline from [Cmd.pipe](../Cmd/pipe.md) writes into a pipe as one command does. Its last stage
writes to `pipe.write`:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
let words = Cmd.pipe([Cmd.new("printf", ["b\\na\\nb\\nc\\n"]), Cmd.new("sort", ["-u"])])
let child = try Cmd.start(words, stdout: pipe.write)
try pipe.write.close()
print(try pipe.read.readAll())
println(try child.wait())
```

```text
a
b
c
0
```

A pipe joins two goroutines of the program too. The writer closes its end when it is done, which
ends the reader's loop:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
go {
  defer { pipe.write.close() }
  for n in Array.range(1, 4) {
    pipe.write.writeLine("tick ${n}").unwrap()
  }
}
while true {
  let line = try pipe.read.readLine() ?? break
  println(line)
}
```

```text
tick 1
tick 2
tick 3
```

## Applications

Run a command and tag each line of its output, both streams merged, as
`cmd 2>&1 | sed 's/^/[build] /'` does. The write end is closed before `try` looks at the result of
`Cmd.start`, so it is closed when the start fails too:

```ant
func runTagged(tag: string, c: Cmd): Result[int, Error] {
  let pipe = try Os.openPipe()
  defer { pipe.read.close() }
  let started = Cmd.start(c, stdout: pipe.write, stderr: pipe.write)
  try pipe.write.close()
  let child = try started
  while true {
    let line = try pipe.read.readLine() ?? break
    println("[${tag}] ${line}")
  }
  child.wait()
}

let script = "echo compiling; echo 'warning: slow' >&2; echo done"
let status = try runTagged("build", Cmd.new("sh", ["-c", script]))
println("exit ${status}")
```

```text
[build] compiling
[build] warning: slow
[build] done
exit 0
```

## Properties and relations

A pipe passes bytes through unchanged, and after the write end is closed every read is at the end:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
let data = Bytes.fromArray([0, 1, 254, 255])
try pipe.write.writeChunk(data)
try pipe.write.close()
assert(try pipe.read.readExactly(4) == data)
assert(try pipe.read.readChunk(10) == null)
assert(try pipe.read.readLine() == null)
```

Each call makes a new pipe. What is written to one is read only from that one:

```ant
let a = try Os.openPipe()
let b = try Os.openPipe()
defer { a.read.close() }
defer { b.read.close() }
defer { a.write.close() }
defer { b.write.close() }
try a.write.writeLine("to a")
try b.write.writeLine("to b")
assert(try b.read.readLine() == "to b")
assert(try a.read.readLine() == "to a")
```

## Possible issues

The read end is not at its end while the program's own copy of the write end is open, even after the
child has exited. Without the close, `File.readAll` waits forever; here a context stops it:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
let child = try Cmd.start(Cmd.new("seq", ["3"]), stdout: pipe.write)
// The program's copy of the write end is still open.
println(pipe.read.readAll(ctx: Context.new(100ms)))
try pipe.write.close()
println(try child.wait())
```

```text
Err{ error: Error{ message: "timeout" } }
0
```

Waiting for a child before reading its pipe can wait forever: once the pipe is full the child blocks
on its next write and never exits. Here a context stops the wait, and reading the pipe lets the
child finish:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
let child = try Cmd.start(Cmd.new("seq", ["100000"]), stdout: pipe.write)
try pipe.write.close()
println(child.wait(ctx: Context.new(100ms)))
let lines = try pipe.read.readAll() | String.lines()
println(lines.length())
println(try child.wait())
```

```text
Err{ error: Error{ message: "timeout" } }
100000
0
```

A write after the read end is closed has nowhere to go. It is `Err`, and the program goes on:

```ant
let pipe = try Os.openPipe()
defer { pipe.write.close() }
try pipe.read.close()
println(pipe.write.writeLine("x"))
println("still running")
```

```text
Err{ error: Error{ message: "broken pipe" } }
still running
```

## See also

[Cmd.start](../Cmd/start.md) · [Cmd.wait](../Cmd/wait.md) · [Cmd.pipe](../Cmd/pipe.md) ·
[File.readLine](../File/readLine.md) · [File.readAll](../File/readAll.md) ·
[File.close](../File/close.md) · [Os.Pipe](types/Pipe.md) · [Os.stdout](stdout.md)
