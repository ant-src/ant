<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.Pipe

```ant signature
type Os.Pipe = { read: File.Handle, write: File.Handle }
```

Os.Pipe is a pair of connected handles, made by [`Os.openPipe`](../openPipe.md): what is written to
write is read from read.

<!-- end generated -->

## Details

- `Os.Pipe` is a record of two [File.Handle](../../File/types/Handle.md) values. The `File`
  functions read `read` and write `write`; the pipe itself has no functions.
- The program closes both ends. `read` is at end of file once every copy of `write` is closed,
  including the copies given to children by [`Cmd.start`](../../Cmd/start.md).
- A read of `read` waits until data comes or every copy of `write` is closed; pass `ctx:` to give
  up, as `pipe.read.readLine(ctx: c)`.
- Copying the record, by assignment, a spread, or a call, copies the handles, not the pipe: every
  copy names the same two ends, so closing one copy's end closes them all.
- [`Cmd.pipe`](../../Cmd/pipe.md) joins commands with pipes of its own. Use an `Os.Pipe` for what it
  does not join: a child's output read as it comes, its input written a piece at a time, or the two
  outputs kept apart.

## Basic examples

`Os.openPipe` returns an `Os.Pipe`. Write to its `write` field and read from its `read` field:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
try pipe.write.writeLine("hello")
try pipe.write.close()
println(try pipe.read.readLine())
```

```text
hello
```

Destructure the two ends into names of their own:

```ant
let { read, write } = try Os.openPipe()
defer { read.close() }
try write.writeLine("through the pipe")
try write.close()
print(try read.readAll())
```

```text
through the pipe
```

A pipe renders as a record of two handles:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
defer { pipe.write.close() }
println(pipe)
```

```text
Os.Pipe{ read: File.Handle, write: File.Handle }
```

## Scope

Rename the ends as you destructure them, to say which way the data flows:

```ant
let { read: fromChild, write: childOut } = try Os.openPipe()
defer { fromChild.close() }
let child = try Cmd.start(Cmd.new("echo", ["ready"]), stdout: childOut)
try childOut.close()
println(try fromChild.readLine())
println(try child.wait())
```

```text
ready
0
```

Take an `Os.Pipe` as a parameter. This function writes lines to the pipe and closes its write end,
on an error too, so the caller reads to the end:

```ant
func send(pipe: Os.Pipe, lines: Array[string]): Result[void, Error] {
  defer { pipe.write.close() }
  for line in lines {
    try pipe.write.writeLine(line)
  }
  pipe.write.close()
}

let pipe = try Os.openPipe()
defer { pipe.read.close() }
try send(pipe, ["one", "two"])
print(try pipe.read.readAll())
```

```text
one
two
```

Accept a pipe or a plain handle, and tell them apart with `is`:

```ant
func sink(to: File.Handle | Os.Pipe): File.Handle {
  if to is Os.Pipe { to.write } else { to }
}

let pipe = try Os.openPipe()
defer { pipe.read.close() }
defer { pipe.write.close() }
println(sink(pipe) == pipe.write)
println(sink(Os.stdout()) == Os.stdout())
```

```text
true
true
```

## Applications

Read a child's output line by line as it comes. `readLine` returns `null` once the child has exited
and the program has closed its copy of the write end:

```ant
let out = try Os.openPipe()
defer { out.read.close() }
let child = try Cmd.start(Cmd.new("seq", ["3"]), stdout: out.write)
try out.write.close()
while true {
  let line = try out.read.readLine()
  if line == null {
    break
  }
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

Write a child's input a piece at a time. The child reads to the end of its input, so the program
closes the write end when it is done:

```ant
let input = try Os.openPipe()
let child = try Cmd.start(Cmd.new("sort"), stdin: input.read)
try input.read.close()
for word in ["pear", "apple", "fig"] {
  try input.write.writeLine(word)
}
try input.write.close()
println(try child.wait())
```

```text
apple
fig
pear
0
```

Start several commands, each writing into a pipe of its own, and read each output whole. The write
end is closed before `Cmd.start`'s result is looked at, so it is closed when the start fails too:

```ant
type Job = { name: string, out: Os.Pipe, process: Cmd.Process }

func startJob(name: string, c: Cmd): Result[Job, Error] {
  let out = try Os.openPipe()
  let started = Cmd.start(c, stdout: out.write)
  try out.write.close()
  if started is Err {
    out.read.close()
    return Result.err(started.error)
  }
  Result.ok(Job{ name, out, process: started.value })
}

let jobs = [
  try startJob("greet", Cmd.new("printf", [`hello\nworld\n`])),
  try startJob("count", Cmd.new("seq", ["3"])),
]
for job in jobs {
  defer { job.out.read.close() }
  let lines = try job.out.read.readAll() | String.lines()
  let status = try job.process.wait()
  println("${job.name}: ${lines.length()} lines, exit ${status}")
}
```

```text
greet: 2 lines, exit 0
count: 3 lines, exit 0
```

## Properties and relations

The two ends of a pipe are different handles, and a new call makes a new pipe. A destructured end
and a spread copy hold the same handles as the original:

```ant
let a = try Os.openPipe()
let b = try Os.openPipe()
defer { a.read.close() }
defer { a.write.close() }
defer { b.read.close() }
defer { b.write.close() }
assert(a.read != a.write)
assert(a != b)
let { read } = a
assert(read == a.read)
assert(Os.Pipe{ ...a } == a)
```

Two pipes between two children do what `Cmd.pipe` does with one call. Each child gets its own copy
of the ends it is given, so the program closes its copies of `link` once both have started.
`Cmd.read` drops the output's final newline, so the comparison puts it back:

```ant
let words = Cmd.new("printf", [`pear\napple\npear\n`])
let sort = Cmd.new("sort", ["-u"])

let link = try Os.openPipe()
let out = try Os.openPipe()
defer { out.read.close() }
let first = try Cmd.start(words, stdout: link.write)
let second = try Cmd.start(sort, stdin: link.read, stdout: out.write)
try link.write.close()
try link.read.close()
try out.write.close()
let byHand = try out.read.readAll()
assert(try first.wait() == 0)
assert(try second.wait() == 0)
assert(byHand == "${try Cmd.read(Cmd.pipe([words, sort]))}\n")
print(byHand)
```

```text
apple
pear
```

## Possible issues

A copy of the record is not a new pipe. Closing the write end through the copy closes the original's
too:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
let copy = Os.Pipe{ ...pipe }
try copy.write.close()
println(pipe.write.writeLine("late"))
```

```text
Err{ error: Error{ message: "the file is closed" } }
```

A literal that takes its ends from two pipes is not connected: what is written to its `write` goes
into the other pipe. Here a context stops the read that would otherwise wait forever:

```ant
let a = try Os.openPipe()
let b = try Os.openPipe()
defer { a.read.close() }
defer { a.write.close() }
defer { b.read.close() }
defer { b.write.close() }
let mixed = Os.Pipe{ read: a.read, write: b.write }
try mixed.write.writeLine("hello")
println(mixed.read.readLine(ctx: Context.new(100ms)))
println(try b.read.readLine())
```

```text
Err{ error: Error{ message: "timeout" } }
hello
```

`Os.Pipe` has no `close`. Close each end:

```ant
let pipe = try Os.openPipe()
pipe.close() // error: Os.Pipe has no field or function close
```

## See also

[Os.openPipe](../openPipe.md) · [File.Handle](../../File/types/Handle.md) ·
[File.close](../../File/close.md) · [Cmd.start](../../Cmd/start.md) · [Cmd.pipe](../../Cmd/pipe.md)
· [Cmd.Pipeline](../../Cmd/types/Pipeline.md)
