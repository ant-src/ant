<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.readAll

```ant signature
func File.readAll(f: File.Handle, ctx: Context | null = null): Result[string, Error]
```

File.readAll reads the rest of f as text. Bytes that are not valid UTF-8 become U+FFFD.

<!-- end generated -->

## Details

- `File.readAll` reads from where the handle is to the end. What earlier reads of the handle took,
  such as a header line from [`File.readLine`](readLine.md), is not in the result.
- At the end of the file it is `Ok` with `""`, never `null`, so a second `File.readAll` gives `""`.
- For binary data, read the handle with [`File.readChunk`](readChunk.md), or the path with
  [`File.readBytes`](readBytes.md).
- On a connection, a pipe, or a terminal, `File.readAll` waits for the end of the input; standard
  input is often one of these. A `ctx` ends the wait, and what was read by then stays for the next
  read of the handle.
- An `Err` from a read of a handle does not name the file. Add the path with `Result.wrapErr`.

## Basic examples

Read a whole file through its handle:

<!-- test: file notes.txt
buy milk
call Ann
-->

```ant
let f = try File.open("notes.txt")
defer { f.close() }
print(try File.readAll(f))
```

```text
buy milk
call Ann
```

Read everything on standard input:

<!-- test: stdin
hello from stdin
-->

```ant
let input = try File.readAll(Os.stdin())
println(input.trim().toUpper())
```

```text
HELLO FROM STDIN
```

## Scope

After [`File.readLine`](readLine.md), `File.readAll` returns the rest. The reads of a handle share
one buffer, so nothing is skipped or read twice:

<!-- test: file people.csv
name,age
Ann,31
Bob,27
-->

```ant
let f = try File.open("people.csv")
defer { f.close() }
let header = try f.readLine()
let rest = try f.readAll()
println("header: ${header}")
print(rest)
```

```text
header: name,age
Ann,31
Bob,27
```

An empty file, and a handle already at its end, give `Ok` with `""`:

<!-- test: file empty.txt
-->

```ant
let f = try File.open("empty.txt")
defer { f.close() }
println(f.readAll())
```

```text
Ok{ value: "" }
```

Bytes that are not valid UTF-8 become U+FFFD, the replacement character. Here `0xe9` is `é` in
Latin-1:

```ant
try File.writeBytes("latin1.txt", Bytes.fromArray([0x63, 0x61, 0x66, 0xe9]))
let f = try File.open("latin1.txt")
defer { f.close() }
println(try f.readAll())
```

```text
caf�
```

A pipe from [`Os.openPipe`](../Os/openPipe.md) is read to its end once every copy of its write end
is closed. Here a child process writes into it:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
let child = try Cmd.start(Cmd.new("echo", ["built", "ok"]), stdout: pipe.write)
try pipe.write.close()
print(try pipe.read.readAll())
println(try child.wait())
```

```text
built ok
0
```

## Named arguments

### ctx

A context ends a wait on a pipe, a connection, or a terminal with `Err` and the context's reason.
The text that arrived before stays in the handle, and the next read returns it:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
try pipe.write.writeText("first part, ")

let ctx = Context.new(Time.duration(milliseconds: 50))
println(pipe.read.readAll(ctx: ctx))

try pipe.write.writeText("second part")
try pipe.write.close()
println(pipe.read.readAll())
```

```text
Err{ error: Error{ message: "timeout" } }
Ok{ value: "first part, second part" }
```

A read of a regular file never waits, so there the context is looked at only before the read begins:

<!-- test: file notes.txt
buy milk
call Ann
-->

```ant
let ctx = Context.new()
Context.cancel(ctx)
let f = try File.open("notes.txt")
defer { f.close() }
println(f.readAll(ctx: ctx))
```

```text
Err{ error: Error{ message: "cancelled" } }
```

## Applications

Count the words a pipeline sends to a script, as `wc -w` does:

<!-- test: stdin
the quick brown fox
jumps over the lazy dog
-->

```ant
let text = try Os.stdin().readAll()
println(text.fields().length())
```

```text
9
```

Read a message whose first line is a subject and whose body is everything after it. The reads of the
handle name no file, so `loadMessage` adds the path:

<!-- test: file message.txt
Subject: Release notes
Version 1.2 is out.
It fixes two bugs.
-->

```ant
type Message = { subject: string, body: string }

func readMessage(f: File.Handle): Result[Message, Error] {
  let first = try f.readLine() ?? ""
  let body = try f.readAll()
  Result.ok(Message{ subject: first.trimPrefix("Subject: "), body: body.trim() })
}

func loadMessage(path: string): Result[Message, Error] {
  let f = try File.open(path)
  defer { f.close() }
  readMessage(f) | Result.wrapErr("read ${path}")
}

let message = try loadMessage("message.txt")
println(message.subject)
println(message.body)
```

```text
Release notes
Version 1.2 is out.
It fixes two bugs.
```

## Properties and relations

Reading a fresh handle to the end gives what [`File.read`](read.md) gives for the path:

<!-- test: file notes.txt
buy milk
call Ann
-->

```ant
let f = try File.open("notes.txt")
defer { f.close() }
assert(try f.readAll() == try File.read("notes.txt"))
assert(try f.readAll() == "")
```

A header line read with `File.readLine`, its line break, and the rest make up the whole file:

<!-- test: file people.csv
name,age
Ann,31
Bob,27
-->

```ant
let f = try File.open("people.csv")
defer { f.close() }
let header = try f.readLine() ?? ""
let rest = try f.readAll()
assert("${header}\n${rest}" == try File.read("people.csv"))
```

## Possible issues

A second `File.readAll` does not start over. It returns `""`, since the first one read to the end:

<!-- test: file notes.txt
buy milk
call Ann
-->

```ant
let f = try File.open("notes.txt")
defer { f.close() }
println(f.readAll())
println(f.readAll())
```

```text
Ok{ value: "buy milk\ncall Ann\n" }
Ok{ value: "" }
```

A pipe is at its end only when every copy of its write end is closed, the program's own included.
Without the close, `File.readAll` waits forever; here a context stops it:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
defer { pipe.write.close() }
try pipe.write.writeText("done")
// The write end is still open, so the end never comes.
println(pipe.read.readAll(ctx: Context.new(Time.duration(milliseconds: 50))))
```

```text
Err{ error: Error{ message: "timeout" } }
```

A handle from [`File.create`](create.md) or [`File.append`](append.md) is for writing, and a closed
handle reads nothing. Both are `Err`, and the message does not name the file:

```ant
let out = try File.create("log.txt")
println(out.readAll())
try out.close()
let f = try File.open("log.txt")
try f.close()
println(f.readAll())
```

```text
Err{ error: Error{ message: "the file is not open for reading" } }
Err{ error: Error{ message: "the file is closed" } }
```

## See also

[File.read](read.md) · [File.open](open.md) · [File.readLine](readLine.md) ·
[File.readChunk](readChunk.md) · [File.readExactly](readExactly.md) · [File.pipe](pipe.md) ·
[File.close](close.md) · [Os.stdin](../Os/stdin.md) · [Os.openPipe](../Os/openPipe.md)
