<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File

File holds the functions of the file system and of handles: open files, the standard streams, and
network connections. Every operation that can fail returns a Result with an Error; a read that waits
takes a context. An empty path is Err in every call that takes one; File.exists("") is false. Only
[`File.mkdir`](mkdir.md) creates missing directories: every other call that creates a file, a
directory or a link is Err when the directory that would hold it is missing, and those that write a
file are Err when a directory is at the path.

<!-- end generated -->

## Details

- Every function that can fail returns a `Result`, and none panics on bad input. A call that takes a
  path names it in the `Err`'s message; a call on a handle names no target, so add one with
  `Result.wrapErr`.
- [File.read](read.md), [File.readLine](readLine.md), and [File.readAll](readAll.md) decode what
  they read as text, so bytes that are not UTF-8 become U+FFFD. [File.readBytes](readBytes.md),
  [File.readChunk](readChunk.md), and [File.readExactly](readExactly.md) keep the bytes.
- A [File.Handle](types/Handle.md) is an open file, a standard stream, a connection from `Net`, or a
  buffer over one, and the same reads and writes serve each. The reads of a handle share one buffer,
  so lines and bytes can follow each other.
- A write is handed to the OS before it returns, so nothing needs a flush, except on a handle from
  [File.buffered](buffered.md), which holds its writes until it is full, flushed, or closed.
- A deferred close cannot report an error. A function that writes ends with an explicit
  `try File.close(out)`; the deferred close then does nothing.
- `write`, `writeBytes`, `create`, and `copy` replace what is at the path. With `replace: false`
  they are `Err` when anything is there.

## Basic examples

Write a whole file with `File.write` and read it back with `File.read`:

```ant
try File.write("hello.txt", "Hello, Ant\n")
print(try File.read("hello.txt"))
```

```text
Hello, Ant
```

Open a handle with `File.open`, close it with `defer`, and read it line by line with
`File.readLine`:

<!-- test: file todo.txt
buy milk
water plants
-->

```ant
let f = try File.open("todo.txt")
defer { File.close(f) }
while true {
  let line = try File.readLine(f) ?? break
  println("- ${line}")
}
```

```text
- buy milk
- water plants
```

A failure is an `Err` that names the path:

```ant
println(File.read("missing.txt"))
```

```text
Err{ error: Error{ message: "open missing.txt: no such file or directory" } }
```

## Scope

A function of the namespace is called through `.` on a handle, so `f.writeLine(line)` is
`File.writeLine(f, line)`. A function that takes a handle serves a file and a standard stream alike:

```ant
func greet(out: File.Handle, names: Array[string]): Result[void, Error] {
  for name in names {
    try out.writeLine("Hello, ${name}")
  }
  Result.ok(void)
}

try greet(Os.stdout(), ["Ann"])
let f = try File.create("greetings.txt")
defer { f.close() }
try greet(f, ["Bob", "Cy"])
try f.close()
print(try File.read("greetings.txt"))
```

```text
Hello, Ann
Hello, Bob
Hello, Cy
```

Text reads decode, and byte reads keep the bytes as they are:

```ant
try File.writeBytes("cafe.txt", Bytes.fromArray([0x63, 0x61, 0x66, 0xE9]))
println(try File.read("cafe.txt"))
println(try File.readBytes("cafe.txt"))
```

```text
caf�
Bytes(63 61 66 e9)
```

The reads of a handle share one buffer, so a header line can come first and the bytes after it:

```ant
try File.write("packet.txt", "HEAD 3\nabc")
let f = try File.open("packet.txt")
defer { f.close() }
println(try f.readLine())
println(try f.readExactly(3))
```

```text
HEAD 3
Bytes(61 62 63)
```

A connection is a handle too. This server echoes one line back:

```ant
let listener = try Net.listen("127.0.0.1:0")
go {
  let conn = Result.unwrap(Net.accept(listener))
  defer { conn.close() }
  let line = Result.unwrap(conn.readLine()) ?? ""
  Result.unwrap(conn.writeLine("echo: ${line}"))
}
let conn = try Net.dial(listener.address())
defer { conn.close() }
try conn.writeLine("ping")
println(try conn.readLine())
```

```text
echo: ping
```

A read of a connection waits for input, and a context gives up the wait. Here the server never
writes:

```ant
let listener = try Net.listen("127.0.0.1:0")
let conn = try Net.dial(listener.address())
defer { conn.close() }
let ctx = Context.new(timeout: Time.duration(milliseconds: 50))
let line = conn.readLine(ctx: ctx)
println(line)
println(Context.reason(ctx))
```

```text
Err{ error: Error{ message: "timeout" } }
timeout
```

The handle stays open after such a read, and a later read gets what arrives. A read of a regular
file does not wait, so there a context is looked at only before the read begins.

Standard input is a handle from `Os.stdin()`, read line by line like a file:

<!-- test: stdin
hello
ant
-->

```ant
let input = Os.stdin()
while true {
  let line = try input.readLine() ?? break
  println(line.toUpper())
}
```

```text
HELLO
ANT
```

[File.list](list.md) gives the names in one directory, [File.walk](walk.md) the files below it, and
[File.glob](glob.md) the paths that match a pattern:

```ant
try File.mkdir("notes/2026")
try File.write("notes/todo.txt", "")
try File.write("notes/2026/march.txt", "")
try File.write("notes/2026/march.md", "")
println(try File.list("notes"))
println(try File.walk("notes"))
println(try File.glob("**/*.txt", dir: "notes"))
```

```text
["2026", "todo.txt"]
["2026/march.md", "2026/march.txt", "todo.txt"]
["2026/march.txt", "todo.txt"]
```

## Functions by job

### Types

- [File.Handle](types/Handle.md): an open file, a standard stream, a connection, or a buffer over
  one
- [File.Info](types/Info.md): the name, size, kind, modified time, and mode of a file

### Whole files

- [File.read](read.md), [File.readBytes](readBytes.md): the contents of a file, as text or as bytes
- [File.write](write.md), [File.writeBytes](writeBytes.md): create a file, or replace what it held

### Opening a handle

- [File.open](open.md): open a file for reading
- [File.create](create.md): create or truncate a file for writing
- [File.append](append.md): open a file for writing at its end
- [File.buffered](buffered.md): a handle that holds its writes until full, flush, or close
- [File.isTerminal](isTerminal.md): whether a handle is a terminal
- [File.close](close.md): close a handle, writing out a buffer first

### Reading a handle

- [File.readLine](readLine.md): the next line, or `null` at the end
- [File.readAll](readAll.md): the rest of the input as text
- [File.readChunk](readChunk.md): what is there, up to a number of bytes
- [File.readExactly](readExactly.md): exactly a number of bytes

### Writing a handle

- [File.writeLine](writeLine.md), [File.writeText](writeText.md): text, with or without a newline
- [File.writeChunk](writeChunk.md): bytes as they are
- [File.pipe](pipe.md): copy the rest of one handle to another
- [File.flush](flush.md): hand what a buffered handle holds to the OS

### Directories and listings

- [File.mkdir](mkdir.md): make a directory with its parents
- [File.list](list.md): the names in a directory
- [File.walk](walk.md): every file below a directory
- [File.glob](glob.md): the paths that match a pattern

### Metadata

- [File.exists](exists.md): whether anything is at a path
- [File.stat](stat.md): the [File.Info](types/Info.md) of a path
- [File.chmod](chmod.md): set the mode
- [File.touch](touch.md): set the modified time, creating an empty file when missing

### Moving, copying, and removing

- [File.rename](rename.md): move a file or directory
- [File.copy](copy.md): copy one file's contents and mode
- [File.copyAll](copyAll.md): copy a file or a whole tree
- [File.remove](remove.md): remove a file or an empty directory
- [File.removeAll](removeAll.md): remove a whole tree

### Links

- [File.symlink](symlink.md): make a symbolic link
- [File.readLink](readLink.md): the target of a link, as stored
- [File.realPath](realPath.md): a path made absolute with every link resolved

### Temporary files and locks

- [File.tempFile](tempFile.md), [File.tempDir](tempDir.md): a new file or directory under a free
  name
- [File.lock](lock.md): open a file holding its lock, waiting for it
- [File.tryLock](tryLock.md): the same without waiting, or `null` when another handle holds it

## Applications

Write many lines through a buffer. The deferred close covers an early return, and the final close
reports whether the output was written:

```ant
func writeReport(path: string, rows: Array[string]): Result[void, Error] {
  let out = File.buffered(try File.create(path))
  defer { out.close() }
  for row in rows {
    try out.writeLine(row)
  }
  out.close()
}

try writeReport("report.txt", ["Ann 9", "Bob 7"])
print(try File.read("report.txt"))
```

```text
Ann 9
Bob 7
```

Replace a file so that a reader never sees it half-written: write a temporary file beside it, then
rename it into place:

```ant
func save(path: string, text: string): Result[void, Error] {
  let tmp = try File.tempFile(".save-*", dir: Path.dir(path))
  defer { File.remove(tmp) }
  try File.write(tmp, text)
  File.rename(tmp, path)
}

try save("config.txt", "port=8080")
println(try File.read("config.txt"))
println(try File.glob(".save-*", hidden: true))
```

```text
port=8080
[]
```

Count the lines of every text file in a directory:

```ant
try File.mkdir("logs")
try File.write("logs/a.txt", "one\ntwo\n")
try File.write("logs/b.txt", "three\n")
let dir = "logs"
for path in try File.glob("*.txt", dir: dir) {
  let text = try File.read(Path.join([dir, path]))
  println("${path}: ${text.lines().length()}")
}
```

```text
a.txt: 2
b.txt: 1
```

Let one run at a time do the work. A second lock on the file, even in the same process, does not get
it while the first is held:

```ant
let first = try File.tryLock("job.lock")
let second = try File.tryLock("job.lock")
println(first != null)
println(second == null)
```

```text
true
true
```

## Properties and relations

What `File.write` writes, `File.read` reads back, and its size in bytes is what `File.stat` reports:

```ant
let text = "héllo"
try File.write("round.txt", text)
assert((try File.read("round.txt")) == text)
assert((try File.stat("round.txt")).size == text.toBytes().length())
```

`File.exists` is true after a write and false after a remove:

```ant
try File.write("gone.txt", "")
assert(File.exists("gone.txt"))
try File.remove("gone.txt")
assert(!File.exists("gone.txt"))
```

## Possible issues

A `Result` is not the value it holds. Use `try` or narrow it first:

```ant
let text = File.read("notes.txt")
println(text.length())
// error: Result[string, Error] has no field or function length
```

`File.remove` refuses a directory with entries. [File.removeAll](removeAll.md) removes the tree:

```ant
try File.mkdir("build/out")
println(File.remove("build"))
println(File.removeAll("build"))
```

```text
Err{ error: Error{ message: "remove build: directory not empty" } }
Ok{ value: void }
```

`replace: false` is `Err` when the file exists, and changes nothing:

```ant
try File.write("first.txt", "one")
println(File.write("first.txt", "two", replace: false))
println(try File.read("first.txt"))
```

```text
Err{ error: Error{ message: "open first.txt: file exists" } }
one
```

A buffered handle neither flushed nor closed loses what it holds when the run ends:

```ant
let out = File.buffered(Os.stdout())
try out.writeLine("lost")
println("printed")
```

```text
printed
```

A `print` writes to the stream under the buffer, so it can come out before text the buffer still
holds:

```ant
let out = File.buffered(Os.stdout())
try out.writeLine("first")
println("second")
try out.close()
```

```text
second
first
```

An error from a handle names no file. Add the target with `Result.wrapErr`. A handle from
`File.open` is for reading, so a write to it is `Err`:

```ant
try File.write("log.txt", "")
let f = try File.open("log.txt")
defer { f.close() }
println(f.writeLine("x"))
println(f.writeLine("x") | Result.wrapErr("write log.txt"))
```

```text
Err{ error: Error{ message: "the file is not open for writing" } }
Err{ error: Error{ message: "write log.txt: the file is not open for writing" } }
```

## See also

[Path](../Path/index.md) · [Os](../Os/index.md) · [Net](../Net/index.md) ·
[Read and write a whole file](../../howto/files/read-write.md) ·
[Replace a file without a partial write](../../howto/files/replace-safely.md) ·
[Walk a directory and match globs](../../howto/files/walk-glob.md)
