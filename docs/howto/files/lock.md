<!-- status: reviewed -->

# Lock a file between processes

Let one run of a program at a time do its work, such as a backup script that cron starts again while
its last run is still going.

1. Pick a lock file that every run names the same way, such as a file under
   [Os.cacheDir](../../lib/Os/cacheDir.md). The directory must exist; the file need not.
2. Call [File.tryLock](../../lib/File/tryLock.md) on it. It returns `null` when another run holds
   the lock.
3. Exit when the result is `null`.
4. Close the handle with `defer` right after you get it. Closing it releases the lock.

## Skip the run when another one holds the lock

<!-- test: exit 1 -->
<!-- test: before
let running = File.lock(Path.join([Os.cacheDir().unwrap(), "backup.lock"])).unwrap()
-->

```ant
let path = Path.join([try Os.cacheDir(), "backup.lock"])
let lock = try File.tryLock(path)
if lock == null {
  eprintln("backup: already running")
  Os.exit(1)
}
defer { lock.close() }
println("backing up")
```

Here another handle holds the lock file, so the script writes `backup: already running` to standard
error and exits with status 1. When no one holds the lock, it prints `backing up` and releases the
lock as it ends.

The operating system releases the lock when the process ends, however it ends, a crash and `kill -9`
included. A lock is never left stale, as a pid file is, so you never delete the lock file by hand.
Leave the file in place: the next run locks it again.

## Wait for the lock

To queue behind the run that holds the lock instead of giving up, call
[File.lock](../../lib/File/lock.md). It waits until the lock is free. Give it a context from
[Context.new](../../lib/Context/new.md) to limit the wait:

<!-- test: before
let other = File.lock("deploy.lock").unwrap()
go {
  Time.sleep(100ms)
  _ = other.close()
}
-->

```ant
let lock = try (File.lock("deploy.lock", ctx: Context.new(30s)) | Result.wrapErr("wait for deploy.lock"))
defer { lock.close() }
println("deploying")
```

```text
deploying
```

When the context ends first, `File.lock` returns an `Err` with the context's reason, and `try` ends
the script with status 1:

<!-- test: exit 1 -->
<!-- test: before
let other = File.lock("deploy.lock").unwrap()
-->

```ant
let lock = try (File.lock("deploy.lock", ctx: Context.new(100ms)) | Result.wrapErr("wait for deploy.lock"))
defer { lock.close() }
println("deploying")
```

The script writes `wait for deploy.lock: timeout` to standard error.

## Guard another file

The lock keeps out only programs that lock too, so every program that changes the data takes the
same lock first. Keep the data in its own file and lock a file next to it. Read, change, and write
the data while you hold the lock:

```ant
func addVisit(path: string): Result[void, Error] {
  let lock = try File.lock("${path}.lock")
  defer { lock.close() }
  let count = if File.exists(path) { (try File.read(path)).trim().toInt() ?? 0 } else { 0 }
  File.write(path, "${count + 1}\n")
}

let results = Parallel.map(Array.range(0, 20), func(i: int, ctx: Context): Result[void, Error] {
  addVisit("visits.txt")
}, limit: 8) ?? return Result.fail("cancelled")
for r in results {
  try r
}
print(try File.read("visits.txt"))
```

```text
20
```

The lock belongs to the handle, not to the process, so it orders goroutines of one program as it
orders processes: here [Parallel.map](../../lib/Parallel/map.md) runs `addVisit` eight at a time,
and no visit is lost.

## Things to know

- There is no unlock. Closing the handle is the one release.
- On Unix the lock is advisory: a program that writes the file without locking it is not stopped. On
  Windows other processes cannot read or write the file while the lock is held.
- To lock again right after a close, call `File.lock`, which waits, not `File.tryLock`: a child
  process that another goroutine starts at that moment can hold the lock a moment after the close.
- Whether the lock is held is what `File.tryLock` answers, not what the file holds. Do not keep a
  pid in it: the lock never truncates the file, so a shorter pid leaves the tail of a longer one.

## See also

[File.tryLock](../../lib/File/tryLock.md) · [File.lock](../../lib/File/lock.md) ·
[File.close](../../lib/File/close.md) · [Context.new](../../lib/Context/new.md) ·
[Result.wrapErr](../../lib/Result/wrapErr.md) · [Os.exit](../../lib/Os/exit.md) ·
[Os.cacheDir](../../lib/Os/cacheDir.md) · [Path.join](../../lib/Path/join.md) ·
[Parallel.map](../../lib/Parallel/map.md) · [File.read](../../lib/File/read.md) ·
[File.write](../../lib/File/write.md) · [File.exists](../../lib/File/exists.md)
