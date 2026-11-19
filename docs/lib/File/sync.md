<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.sync

```ant signature
func File.sync(f: File.Handle): Result[void, Error]
```

File.sync flushes f as [`File.flush`](flush.md) does, then has the OS write the file to storage, so
that what was written survives a crash or a power loss. Sync a temporary file before renaming it
into place. It is Err where the OS refuses, as for a pipe, and for a connection or a closed handle.

<!-- end generated -->

## Details

- [File.close](close.md) hands what was written to the OS but does not wait for the disk.
- [File.write](write.md) returns no handle, so there is nothing to sync. To write a file that must
  survive a crash, write through [File.create](create.md) and sync the handle.

## Basic examples

Write a file, sync it to storage, and close it:

```ant
let out = try File.create("orders.txt")
defer { File.close(out) }
try out.writeLine("order 1: 3 apples")
println(File.sync(out))
```

```text
Ok{ value: void }
```

On a buffered handle, `File.sync` writes out the buffer first. Before the sync the file is still
empty:

```ant
let out = File.buffered(try File.create("log.txt"))
defer { File.close(out) }
try out.writeLine("started")
let before = try File.stat("log.txt")
println("before: ${before.size} bytes")
try File.sync(out)
print(try File.read("log.txt"))
```

```text
before: 0 bytes
started
```

## Scope

A pipe cannot be synced; the message depends on the OS. A buffered handle over a pipe is `Err` too:

```ant
let pipe = try Os.openPipe()
defer { File.close(pipe.read) }
defer { File.close(pipe.write) }
println(pipe.write.sync() is Err)
println(pipe.write.buffered().sync() is Err)
```

```text
true
true
```

A connection is not a file either:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { Net.close(listener) }
let conn = try Net.dial(listener.address())
defer { File.close(conn) }
println(conn.sync())
```

```text
Err{ error: Error{ message: "the handle is not a file, which sync writes to storage" } }
```

A closed handle is `Err`, buffered or not:

```ant
let plain = try File.create("plain.txt")
try plain.close()
println(plain.sync())
let out = File.buffered(try File.create("buffered.txt"))
try out.close()
println(out.sync())
```

```text
Err{ error: Error{ message: "the file is closed" } }
Err{ error: Error{ message: "the file is closed" } }
```

## Applications

### Replace a file safely

Write the new contents to a temporary file next to the target, sync it, close it, and rename it over
the target. On Unix a reader sees the old file or the new one, and after a crash the new name never
points at contents the disk has not kept:

```ant
func save(path: string, text: string): Result[void, Error] {
  let tmp = try File.tempFile(".${Path.base(path)}-*", dir: Path.dir(path))
  defer { File.remove(tmp) }
  let out = try File.create(tmp)
  defer { File.close(out) }
  try out.writeText(text)
  try out.sync()
  try out.close()
  File.rename(tmp, path)
}

try save("config.json", "{\"debug\": false}")
try save("config.json", "{\"debug\": true}")
println(try File.read("config.json"))
```

```text
{"debug": true}
```

The new file keeps the temporary file's mode, 0o600; call [File.chmod](chmod.md) before the rename
when others read it. After the rename the deferred `File.remove` finds nothing and its `Err` is
discarded; it cleans up only when `save` returns early.

### A journal

Sync each entry before reporting it saved, so that an entry the program has confirmed survives a
crash:

```ant
func record(journal: File.Handle, entry: string): Result[void, Error] {
  try journal.writeLine(entry)
  journal.sync() | Result.wrapErr("record ${entry}")
}

let journal = try File.append("journal.log")
defer { File.close(journal) }
for entry in ["deposit 100", "withdraw 30"] {
  try record(journal, entry)
  println("saved: ${entry}")
}
```

```text
saved: deposit 100
saved: withdraw 30
```

### Many lines, one sync

One sync covers everything written before it. Write the lines through a buffer and sync once at the
end:

```ant
let out = File.buffered(try File.create("rows.csv"))
defer { File.close(out) }
for row in ["id,qty", "1,3", "2,12"] {
  try out.writeLine(row)
}
try out.sync()
print(try File.read("rows.csv"))
```

```text
id,qty
1,3
2,12
```

## Properties and relations

After `File.sync`, the file holds everything written to the handle, as after `File.flush`. The
handle stays open, so writing continues, and a second sync is `Ok`:

```ant
let out = File.buffered(try File.create("twice.txt"))
defer { File.close(out) }
try out.writeText("abc")
try out.sync()
assert(try File.read("twice.txt") == "abc")
try out.writeText("def")
assert(out.sync() is Ok)
assert(try File.read("twice.txt") == "abcdef")
```

## Possible issues

A deferred sync cannot report its error, since a deferred block's value is discarded. Deferred
blocks also run in reverse order, so a `defer` of the sync written above the `defer` of the close
runs after the close and is `Err`. Sync on the main path with `try`, before the close, as the
examples above do.

A function that syncs whatever handle it gets fails for a pipe or a connection, where a flush
succeeds. Flush in a function that takes any handle, and sync only where you open the file:

```ant
func finish(f: File.Handle): Result[void, Error] {
  f.sync()
}

let pipe = try Os.openPipe()
defer { File.close(pipe.read) }
defer { File.close(pipe.write) }
println(finish(pipe.write) is Err)
println(pipe.write.flush())
```

```text
true
Ok{ value: void }
```

## See also

[File.flush](flush.md) · [File.close](close.md) · [File.buffered](buffered.md) ·
[File.create](create.md) · [File.tempFile](tempFile.md) · [File.rename](rename.md)
