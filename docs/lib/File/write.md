<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.write

```ant signature
func File.write(path: string, text: string, replace: bool = true): Result[void, Error]
```

File.write writes text as UTF-8 to the file at path, creating the file or replacing what it held; a
file that exists keeps its mode. With replace false it creates the file only when nothing is at
path, and is otherwise Err, having changed nothing.

<!-- end generated -->

## Details

- `File.write` replaces the whole file: what it held before is gone, also when the new text is
  shorter. [File.append](append.md) adds to the end instead.
- It writes `text` exactly, with no line break added at the end.
- A new file gets mode `0o666` less the process's umask, `0o644` under the usual umask `022`.
- A symbolic link at `path` is followed, and its target gets the text. With `replace: false` a link
  counts as something at `path`, even one that leads nowhere, and is not followed.

## Basic examples

Write a file and read it back:

```ant
println(File.write("greeting.txt", "Hello, Ant\n"))
println(File.read("greeting.txt"))
```

```text
Ok{ value: void }
Ok{ value: "Hello, Ant\n" }
```

Inside a function that returns a `Result`, `try` returns the `Err` early:

```ant
func backup(path: string): Result[string, Error] {
  let text = try File.read(path)
  let dest = "${path}.bak"
  try File.write(dest, text)
  Result.ok(dest)
}

try File.write("notes.txt", "buy milk\n")
println(backup("notes.txt"))
println(File.read("notes.txt.bak"))
```

```text
Ok{ value: "notes.txt.bak" }
Ok{ value: "buy milk\n" }
```

## Scope

### An existing file

The new text replaces all of the old, even when it is shorter:

```ant
try File.write("version.txt", "a long first version\n")
try File.write("version.txt", "v2\n")
print(try File.read("version.txt"))
```

```text
v2
```

### Empty text

Empty text gives an empty file:

```ant
try File.write("empty.txt", "")
println(File.read("empty.txt"))
```

```text
Ok{ value: "" }
```

### Unicode

The text is written as UTF-8, so a non-ASCII character takes more than one byte:

```ant
try File.write("hello.txt", "héllo 😀")
let bytes = try File.readBytes("hello.txt")
println(bytes)
println(bytes.length())
```

```text
Bytes(68 c3 a9 6c 6c 6f 20 f0 9f 98 80)
11
```

### Modes

A file that exists keeps its mode, so a private file stays private when it is written again:

```ant
try File.write("token.txt", "old")
try File.chmod("token.txt", 0o600)
try File.write("token.txt", "new")
println(String.fromInt((try File.stat("token.txt")).mode, base: 8))
```

```text
600
```

### A symbolic link

Writing through a link replaces the contents of its target, and the link stays a link:

```ant
try File.symlink("target.txt", "current")
try File.write("current", "through the link\n")
print(try File.read("target.txt"))
println((try File.stat("current", follow: false)).isSymlink)
```

```text
through the link
true
```

### Paths that cannot be written

A missing directory and a directory itself are `Err`, with the OS's message:

```ant
println(File.write("nowhere/out.txt", "x"))
try File.mkdir("reports")
println(File.write("reports", "x"))
```

```text
Err{ error: Error{ message: "open nowhere/out.txt: no such file or directory" } }
Err{ error: Error{ message: "open reports: is a directory" } }
```

## Named arguments

### replace

With `replace: false`, `File.write` creates the file only when nothing is at `path`, and is
otherwise an `Err` that changes nothing:

```ant
println(File.write("first.txt", "first\n", replace: false))
println(File.write("first.txt", "second\n", replace: false))
print(try File.read("first.txt"))
```

```text
Ok{ value: void }
Err{ error: Error{ message: "open first.txt: file exists" } }
first
```

A link that leads nowhere counts too, and is not followed, so a link planted at the name cannot
redirect the write:

```ant
try File.symlink("elsewhere.txt", "planted")
println(File.write("planted", "data", replace: false))
println(File.exists("elsewhere.txt"))
```

```text
Err{ error: Error{ message: "open planted: file exists" } }
false
```

## Applications

### Save a record as JSON

Create the directory first, then write the text with a final line break:

```ant
type Config = { name: string, port: int, debug: bool }

func saveConfig(path: string, config: Config): Result[void, Error] {
  try File.mkdir(Path.dir(path))
  File.write(path, Json.stringify(config, indent: "  ") + "\n")
}

try saveConfig("app/config.json", Config{ name: "api", port: 8080, debug: false })
print(try File.read("app/config.json"))
```

```text
{
  "name": "api",
  "port": 8080,
  "debug": false
}
```

### Write a CSV file

Build the whole text, then write it in one call:

```ant
let rows = [["Ann", "31"], ["Bob", "27"]]
let body = rows | Array.map({ x | String.join(",") }) | String.join("\n")
try File.write("people.csv", "name,age\n${body}\n")
print(try File.read("people.csv"))
```

```text
name,age
Ann,31
Bob,27
```

### Add context to the error

Wrap the `Err` so that the caller learns what the program was doing:

```ant
func saveNote(path: string, text: string): Result[void, Error] {
  File.write(path, text) | Result.wrapErr("save note")
}

println(saveNote("drafts/note.txt", "remember\n"))
```

```text
Err{ error: Error{ message: "save note: open drafts/note.txt: no such file or directory" } }
```

### Replace a file atomically

A reader can see a file half-written while `File.write` replaces it. Write a temporary file beside
it and rename that into place. On Unix the rename is atomic, so a reader sees the old file or the
new one:

```ant
func saveAtomic(path: string, text: string): Result[void, Error] {
  let tmp = try File.tempFile(".save-*", dir: Path.dir(path))
  defer { File.remove(tmp) }
  try File.write(tmp, text)
  File.rename(tmp, path)
}

try File.mkdir("state")
try saveAtomic("state/counter.txt", "41\n")
try saveAtomic("state/counter.txt", "42\n")
print(try File.read("state/counter.txt"))
println(try File.list("state"))
```

```text
42
["counter.txt"]
```

The temporary file is `0o600`, and keeps that mode when `File.write` writes it, so this is also how
a private file such as a token is saved.

### Pick a free name

`replace: false` claims a name in one step, even against another program writing at the same time.
After an `Err`, `File.exists` tells a taken name from another failure:

```ant
func writeNew(base: string, text: string): Result[string, Error] {
  for n in Iter.range(1, 100) {
    let path = "${base}-${n}.txt"
    let r = File.write(path, text, replace: false)
    if r.isOk() {
      return Result.ok(path)
    }
    if !File.exists(path) {
      try r
    }
  }
  Result.fail("no free name for ${base}")
}

println(writeNew("report", "first run\n"))
println(writeNew("report", "second run\n"))
```

```text
Ok{ value: "report-1.txt" }
Ok{ value: "report-2.txt" }
```

## Properties and relations

Reading the file back gives the text that was written:

```ant
let text = "line one\nline two\n"
try File.write("round.txt", text)
assert((try File.read("round.txt")) == text)
```

`File.write` writes the bytes of the text as UTF-8, as [File.writeBytes](writeBytes.md) writes them:

```ant
let text = "naïve café"
try File.write("as-text.txt", text)
try File.writeBytes("as-bytes.txt", text.toBytes())
assert((try File.readBytes("as-text.txt")) == (try File.readBytes("as-bytes.txt")))
```

## Possible issues

### Writing in a loop keeps only the last text

Each call replaces the file, so a loop of `File.write` leaves only the last line:

```ant
for line in ["a", "b", "c"] {
  try File.write("loop.txt", line + "\n")
}
print(try File.read("loop.txt"))
```

```text
c
```

Join the lines and write once, as in [Write a CSV file](#write-a-csv-file), or open the file with
[File.create](create.md) and write each line with [File.writeLine](writeLine.md), or text without a
line break with [File.writeText](writeText.md).

### No final line break

`String.join` puts the separator only between parts, so the last line has no `\n`:

```ant
try File.write("names.txt", ["Ann", "Bob"] | String.join("\n"))
println(File.read("names.txt"))
```

```text
Ok{ value: "Ann\nBob" }
```

Add `+ "\n"` when the file should end with a line break.

### An ignored `Err`

A call whose `Result` is not used loses its error, so the program goes on as if the file were
written:

```ant
File.write("lost/out.txt", "data")
println(File.exists("lost/out.txt"))
```

```text
false
```

Use `try`, or check the `Result`.

### Bytes are not text

`File.write` takes a `string`. Write `Bytes` with [File.writeBytes](writeBytes.md):

```ant
let data = "hi".toBytes()
let r = File.write("data.bin", data) // error: cannot use Bytes where string is expected
```

## See also

[File.writeBytes](writeBytes.md) · [File.read](read.md) · [File.append](append.md) ·
[File.create](create.md) · [File.writeLine](writeLine.md) · [File.writeText](writeText.md) ·
[File.tempFile](tempFile.md) · [File.rename](rename.md) · [File.mkdir](mkdir.md) ·
[Path.dir](../Path/dir.md)
