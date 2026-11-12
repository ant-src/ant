<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.pid

```ant signature
func Os.pid(): int
```

Os.pid returns this process's id.

<!-- end generated -->

## Details

- [Cmd.pid](../Cmd/pid.md) gives a child's id, as `Os.pid` gives this process's.

## Basic examples

Print the id of the running program:

<!-- test: check-only the id changes from run to run -->

```ant
println(Os.pid())
```

Every call in a run returns the same id, since it is this process's:

```ant
let pid = Os.pid()
assert(Os.pid() == pid)
```

## Scope

Interpolate it into a string:

<!-- test: check-only the id changes from run to run -->

```ant
println("worker ${Os.pid()} started")
```

## Applications

Tag log lines with the id, so the lines of two copies of a program running at once can be told
apart:

<!-- test: check-only the id changes from run to run -->

```ant
func log(message: string) {
  eprintln("[${Os.pid()}] ${message}")
}

log("starting")
log("done")
```

Run one copy at a time, and write the id into the lock file for people to read:

<!-- test: check-only it locks a file in the user's cache directory -->

```ant
let path = Path.join([try Os.cacheDir(), "app.lock"])
let lock = try File.tryLock(path)
if lock == null {
  eprintln("already running")
  Os.exit(1)
}
defer { lock.close() }
try lock.writeText("${Os.pid()}\n")
```

## Properties and relations

A child process sees this process's id as its parent's, and has an id of its own:

```ant
let parent = try Cmd.new("sh", ["-c", "echo $PPID"]).read()
assert(parent.trim() == "${Os.pid()}")

let child = try Cmd.new("sleep", ["5"]).start()
defer { let _ = child.stop() }
assert(child.pid() != Os.pid())
```

## Possible issues

A pid file says which process wrote it, not whether that process still runs. Decide whether another
copy runs with [File.tryLock](../File/tryLock.md), which the OS releases when the holder ends,
however it ends:

```ant
let path = Path.join([try File.tempDir("app"), "app.lock"])
let first = try File.tryLock(path)
assert(first != null)
defer { first.close() }
assert(try File.tryLock(path) == null)
```

## See also

[Cmd.pid](../Cmd/pid.md) · [Os.uid](uid.md) · [File.tryLock](../File/tryLock.md) · [Os](index.md)
