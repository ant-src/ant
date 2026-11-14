<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.tryLock

```ant signature
func File.tryLock(path: string): Result[File.Handle | null, Error]
```

File.tryLock is [`File.lock`](lock.md) without the wait: it is Ok(null) when another handle holds
the lock.

<!-- end generated -->

## Details

- `Err` is left for a real failure, such as a directory that cannot be written; a held lock is
  `Ok(null)`, never `Err`.
- When the result is `null`, there is no handle to close.
- The lock belongs to the handle, so a second handle in the same process, or the same goroutine,
  conflicts with it too.
- Closing the handle releases the lock. A child process that another goroutine starts at that moment
  can delay the release until its program has started, so a program that locks again at once uses
  [File.lock](lock.md), which waits.
- The OS also releases the lock when the process ends, however it ends, so a crashed run never
  leaves it stale. On Unix the lock keeps out only processes that lock too; on Windows other
  processes cannot read or write the file while it is held.

## Basic examples

Take the lock when it is free, and close the handle to release it:

```ant
let lock = try File.tryLock("backup.lock")
if lock == null {
  eprintln("backup: already running")
  Os.exit(1)
}
defer { lock.close() }
println("backing up")
```

```text
backing up
```

While another handle holds the lock, `File.tryLock` is `Ok(null)` at once:

```ant
let held = try File.lock("job.lock")
defer { held.close() }
println(File.tryLock("job.lock"))
```

```text
Ok{ value: null }
```

## Scope

A missing file is created empty, and the handle reads and writes it:

```ant
println(File.exists("new.lock"))
let lock = try File.tryLock("new.lock")
if lock == null { return }
defer { lock.close() }
try lock.writeText("taken")
println(try File.read("new.lock"))
```

```text
false
taken
```

An existing file is never truncated:

<!-- test: file state.lock
last run: Monday
-->

```ant
let lock = try File.tryLock("state.lock")
if lock == null { return }
defer { lock.close() }
print(try lock.readAll())
```

```text
last run: Monday
```

A path that cannot be opened for reading and writing is `Err`, not `null`:

```ant
println(File.tryLock("missing/dir/app.lock"))
println(File.tryLock("."))
```

```text
Err{ error: Error{ message: "open missing/dir/app.lock: no such file or directory" } }
Err{ error: Error{ message: "open .: is a directory" } }
```

## Applications

Skip a periodic task while its last run still holds the lock, instead of queueing behind it:

```ant
func tick(): Result[string, Error] {
  let lock = try File.tryLock("sync.lock")
  if lock == null { return Result.ok("skipped: last run still going") }
  defer { lock.close() }
  Result.ok("synced")
}

let running = try File.lock("sync.lock")
println(try tick())
try running.close()
println(try tick())
```

```text
skipped: last run still going
synced
```

Take the first free slot of a small pool, so at most three copies of a job run at once:

```ant
func takeSlot(slots: int): Result[File.Handle | null, Error] {
  for i in Array.range(0, slots) {
    let lock = try File.tryLock("slot-${i}.lock")
    if lock != null { return Result.ok(lock) }
  }
  Result.ok(null)
}

let a = try takeSlot(3)
let b = try takeSlot(3)
let c = try takeSlot(3)
println(takeSlot(3))
if b != null { try b.close() }
let d = try takeSlot(3)
println(d != null)
```

```text
Ok{ value: null }
true
```

## Properties and relations

`File.tryLock` sees a lock taken by [File.lock](lock.md), and the other way round; after the close
the lock is free again:

```ant
let first = try File.tryLock("pair.lock")
assert(first != null)
let ctx = Context.new(timeout: 20ms)
assert(!File.lock("pair.lock", ctx: ctx).isOk())
try first.close()
let second = try File.tryLock("pair.lock")
assert(second != null)
```

The file stays after the close, so the next lock reuses it:

```ant
let lock = try File.tryLock("keep.lock")
if lock != null { try lock.close() }
assert(File.exists("keep.lock"))
```

## Possible issues

The result may be `null`, so a handle function on it does not compile until the program checks for
`null`:

```ant
let lock = try File.tryLock("app.lock")
defer { lock.close() }
// error: File.Handle | null has no field or function close on every member: narrow the value first
```

The lock is not reentrant. A second `File.tryLock` in the same process, while the first handle is
open, is `Ok(null)`:

```ant
let outer = try File.tryLock("twice.lock")
println(outer != null)
println(File.tryLock("twice.lock"))
```

```text
true
Ok{ value: null }
```

What the lock file holds says nothing about the lock. A pid written there by an earlier run stays
after its handle closes; ask `File.tryLock` instead:

```ant
let first = try File.tryLock("server.lock")
if first == null { return }
try first.writeLine("12345")
try first.close()
println(try File.read("server.lock"))
println(try File.tryLock("server.lock") != null)
```

```text
12345

true
```

## See also

[File.lock](lock.md) · [File.close](close.md) · [File.open](open.md) · [Mutex](../Mutex/index.md)
