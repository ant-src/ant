<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.lock

```ant signature
func File.lock(path: string, ctx: Context | null = null): Result[File.Handle, Error]
```

File.lock opens the file at path for reading and writing, creating it when it is missing and never
truncating it, and waits for an exclusive lock on it. It is Err with ctx's reason when ctx is done
first. The lock is released when the handle is closed or the process ends, however it ends; a child
process starting at that moment can delay the release until its program has started.

<!-- end generated -->

## Details

- The lock belongs to the handle. A second handle on the same file conflicts with it, even in the
  same process, so a lock orders goroutines as well as processes.
- There is no unlock. Closing the handle is the release.
- The file stays after the close and keeps what it holds. The next lock reuses it.
- On Unix the lock is advisory: it keeps out only processes that lock the file too. On Windows other
  processes cannot read or write the file while the lock is held.
- The release can come a moment after the close when another goroutine starts a child process at
  that moment, so to lock again at once use `File.lock`, which waits, not `File.tryLock`.

## Basic examples

Take the lock, do the work, and release it by closing the handle:

```ant
let lock = try File.lock("backup.lock")
defer { lock.close() }
println("backing up")
```

```text
backing up
```

While one handle holds the lock, [File.tryLock](tryLock.md) on the same file is `Ok(null)`. After
the close it gets the lock:

```ant
let first = try File.lock("job.lock")
println(File.tryLock("job.lock"))
try first.close()
let second = try File.tryLock("job.lock")
println(second != null)
```

```text
Ok{ value: null }
true
```

## Scope

`File.lock` creates the file when it is missing, empty:

```ant
println(File.exists("new.lock"))
let lock = try File.lock("new.lock")
defer { lock.close() }
println(File.read("new.lock"))
```

```text
false
Ok{ value: "" }
```

An existing file is never truncated. The handle reads and writes it:

<!-- test: file state.lock
last run: Monday
-->

```ant
let lock = try File.lock("state.lock")
defer { lock.close() }
print(try lock.readAll())
```

```text
last run: Monday
```

A path that cannot be opened for reading and writing is `Err`:

```ant
println(File.lock("missing/dir/app.lock"))
println(File.lock("."))
```

```text
Err{ error: Error{ message: "open missing/dir/app.lock: no such file or directory" } }
Err{ error: Error{ message: "open .: is a directory" } }
```

## Named arguments

`ctx` bounds the wait. When the context is done first, `File.lock` is `Err` with its reason:

```ant
let held = try File.lock("slow.lock")
defer { held.close() }
let ctx = Context.new(timeout: Time.duration(milliseconds: 50))
println(File.lock("slow.lock", ctx: ctx))
```

```text
Err{ error: Error{ message: "timeout" } }
```

A cancelled context gives up at once:

```ant
let held = try File.lock("busy.lock")
defer { held.close() }
let ctx = Context.new()
ctx.cancel()
println(File.lock("busy.lock", ctx: ctx))
```

```text
Err{ error: Error{ message: "cancelled" } }
```

## Applications

Let one run of a job do its work at a time, and give up after a while instead of queueing forever:

```ant
func runExclusive[T](name: string, work: func(): T): Result[T, Error] {
  let ctx = Context.new(timeout: Time.duration(seconds: 5))
  let lock = try (File.lock("${name}.lock", ctx: ctx) | Result.wrapErr("lock ${name}"))
  defer { lock.close() }
  Result.ok(work())
}

println(runExclusive("report", { "report written" }))
```

```text
Ok{ value: "report written" }
```

Order two goroutines: the second waits in `File.lock` until the first closes its handle:

```ant
let first = try File.lock("order.lock")
let done: Channel[string] = Channel.new()
go {
  let second = File.lock("order.lock").unwrap()
  defer { second.close() }
  Channel.send(done, "second has the lock")
}
Time.sleep(Time.duration(milliseconds: 50))
println("first has the lock")
try first.close()
println(Channel.receive(done))
```

```text
first has the lock
second has the lock
```

Write the process id into the lock file for people to read. Whether the lock is held is what
[File.tryLock](tryLock.md) answers, not what the file says:

```ant
let lock = try File.lock("server.lock")
defer { lock.close() }
try lock.writeLine("${Os.pid()}")
println(try File.read("server.lock") == "${Os.pid()}\n")
```

```text
true
```

## Properties and relations

The file stays after the close, with what was written to it, and the next lock reuses it:

```ant
let first = try File.lock("keep.lock")
try first.writeText("kept")
try first.close()
assert(File.exists("keep.lock"))
let second = try File.lock("keep.lock")
defer { second.close() }
assert(try second.readAll() == "kept")
```

## Possible issues

A handle that is never closed holds the lock until the program ends. Without a `defer`, an early
return leaves the lock held:

```ant
func step(fail: bool): Result[void, Error] {
  let lock = try File.lock("step.lock")
  if fail { return Result.fail("step failed") }
  lock.close()
}

println(step(true))
println(File.tryLock("step.lock"))
```

```text
Err{ error: Error{ message: "step failed" } }
Ok{ value: null }
```

A second `File.lock` on the same file in the same goroutine, while the first handle is open, waits
for itself. Give it a context, or use [File.tryLock](tryLock.md):

```ant
let outer = try File.lock("twice.lock")
defer { outer.close() }
let ctx = Context.new(timeout: Time.duration(milliseconds: 50))
println(File.lock("twice.lock", ctx: ctx))
```

```text
Err{ error: Error{ message: "timeout" } }
```

## See also

[File.tryLock](tryLock.md) · [File.close](close.md) · [File.open](open.md) ·
[Context.new](../Context/new.md) · [Mutex](../Mutex/index.md)
