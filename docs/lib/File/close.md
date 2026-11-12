<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.close

```ant signature
func File.close(f: File.Handle): Result[void, Error]
```

File.close closes f, and is Err when the close fails; f is closed either way. A buffered handle
writes its buffer out first and then closes the handle under it, reporting the first error. When a
write of it is in progress on another goroutine, it closes the handle under first, which ends that
write, and is Err when what the buffer held is lost. It does nothing and is Ok on a standard stream
and on a handle already closed, so a deferred close can follow an explicit one. Called from another
goroutine, it ends a read or write in progress on f with an Err.

<!-- end generated -->

## Details

- A closed handle stays closed: a later write of it is `Err`, and a later close does nothing and is
  `Ok`. This holds even when the close itself is `Err`.
- On a [`File.buffered`](buffered.md) handle, the close is the last chance to write out what the
  buffer holds. A buffered handle the program never closes or flushes loses it when the run ends.
- A deferred close cannot report an error, since a deferred block's value is discarded. When a close
  can fail, as a buffered close can, end with an explicit close and keep the deferred one for the
  paths that leave early.

## Basic examples

Close a file after writing it:

```ant
let out = Result.unwrap(File.create("notes.txt"))
Result.unwrap(out.writeLine("buy milk"))
println(File.close(out))
println(File.read("notes.txt"))
```

```text
Ok{ value: void }
Ok{ value: "buy milk\n" }
```

Close a file you read with `defer`, right after opening it, so that every way out of the function
closes it:

```ant
func firstLine(path: string): Result[string | null, Error] {
  let file = try File.open(path)
  defer { File.close(file) }
  file.readLine()
}

Result.unwrap(File.write("names.txt", "Ann\nBob\n"))
println(firstLine("names.txt"))
```

```text
Ok{ value: "Ann" }
```

## Scope

### Buffered handles

Closing a buffered handle writes out its buffer first. Before the close, the file is still empty:

```ant
let out = Result.unwrap(File.create("log.txt")).buffered()
Result.unwrap(out.writeLine("started"))
println(File.read("log.txt"))
println(out.close())
println(File.read("log.txt"))
```

```text
Ok{ value: "" }
Ok{ value: void }
Ok{ value: "started\n" }
```

When writing the buffer out fails, the close is `Err`. Here nobody reads the pipe any more, so the
write succeeds into the buffer and the failure shows at the close. The handle is closed even though
the close is `Err`, so the second close does nothing and is `Ok`:

```ant
let pipe = Result.unwrap(Os.openPipe())
Result.unwrap(pipe.read.close())
let out = pipe.write.buffered()
println(out.writeLine("hello"))
println(out.close())
println(out.close())
```

```text
Ok{ value: void }
Err{ error: Error{ message: "write |1: broken pipe" } }
Ok{ value: void }
```

A buffered handle over a standard stream is the usual way to buffer output. Its close writes the
buffer out, and the stream stays open:

```ant
let out = Os.stdout().buffered()
Result.unwrap(out.writeLine("one"))
Result.unwrap(out.close())
println("still printing")
```

```text
one
still printing
```

### A closed handle

A second close is `Ok`. A write to a closed handle is `Err`:

```ant
let file = Result.unwrap(File.create("data.txt"))
println(file.close())
println(file.close())
println(file.writeLine("late"))
```

```text
Ok{ value: void }
Ok{ value: void }
Err{ error: Error{ message: "the file is not open for writing" } }
```

### Standard streams

Closing a standard stream is `Ok` and leaves it open:

```ant
println(Os.stdout().close())
println("still printing")
```

```text
Ok{ value: void }
still printing
```

### From another goroutine

A close from another goroutine ends a read that waits on the handle. Here a goroutine waits for a
line from a pipe that never gets one, and the close ends the read with an `Err`:

```ant
let pipe = Result.unwrap(Os.openPipe())
let done: Channel[Result[string | null, Error]] = Channel.new()
go {
  Channel.send(done, pipe.read.readLine())
}
Time.sleep(Time.duration(milliseconds: 50))
Result.unwrap(pipe.read.close())
println(Channel.receive(done) is Err)
Result.unwrap(pipe.write.close())
```

```text
true
```

## Applications

Write a report and report a failed close. The deferred close covers the paths that leave early; on
the main path the explicit close runs first, and the deferred one then does nothing:

```ant
func writeReport(path: string, rows: Array[string]): Result[void, Error] {
  let out = (try File.create(path)).buffered()
  defer { out.close() }
  for row in rows { try out.writeLine(row) }
  out.close()
}

println(writeReport("report.txt", ["north 12", "south 7"]))
println(File.read("report.txt"))
```

```text
Ok{ value: void }
Ok{ value: "north 12\nsouth 7\n" }
```

Release a lock. A handle from [`File.lock`](lock.md) holds the file's lock until it is closed:

```ant
let held = Result.unwrap(File.lock("app.lock"))
println(File.tryLock("app.lock"))
Result.unwrap(held.close())
println(Result.unwrap(File.tryLock("app.lock")) != null)
```

```text
Ok{ value: null }
true
```

## Properties and relations

Closing is safe to repeat, on a file and on a standard stream:

```ant
let file = Result.unwrap(File.create("a.txt"))
assert(file.close() is Ok)
assert(file.close() is Ok)
assert(Os.stdout().close() is Ok)
assert(file.writeText("x") is Err)
```

## Possible issues

A buffered handle that is never closed loses what its buffer holds when the run ends:

```ant
let out = Result.unwrap(File.create("lost.txt")).buffered()
Result.unwrap(out.writeLine("hello"))
println(File.read("lost.txt"))
```

```text
Ok{ value: "" }
```

With only a deferred close, a failed close goes unreported. This function returns `Ok` although its
line never arrives, since nobody reads the pipe:

```ant
func send(conn: File.Handle, line: string): Result[void, Error] {
  let out = conn.buffered()
  defer { out.close() }
  try out.writeLine(line)
  Result.ok(void)
}

let pipe = Result.unwrap(Os.openPipe())
Result.unwrap(pipe.read.close())
println(send(pipe.write, "hello"))
```

```text
Ok{ value: void }
```

End the function with `out.close()` instead of `Result.ok(void)`, and it returns the `Err`.

## See also

[File.open](open.md), [File.create](create.md), [File.buffered](buffered.md),
[File.flush](flush.md), [File.lock](lock.md), [Os.openPipe](../Os/openPipe.md)
