<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.create

```ant signature
func File.create(path: string, replace: bool = true): Result[File.Handle, Error]
```

File.create opens the file at path for writing, creating it or truncating what it held. With replace
false it creates the file only when nothing is at path, and is otherwise Err.

<!-- end generated -->

## Details

- The file is truncated as soon as `File.create` returns, before anything is written. A file that
  exists keeps its mode; a new one gets `0o666` less the umask, `0o644` under the usual umask `022`.
- The handle is for writing only. A read from it is `Err`.
- Close the handle with [`File.close`](close.md), which can fail too.
- `File.create` follows a symbolic link at `path` and writes its target.
- With `replace: false`, anything at `path` makes it `Err`, a directory or a link included, even a
  link that leads nowhere. The check and the creation are one atomic step.

## Basic examples

Create a file, write to it, and close it:

```ant
let out = try File.create("notes.txt")
defer { out.close() }
try out.writeLine("buy milk")
try out.writeLine("call Ann")
try out.close()

println(try File.read("notes.txt"))
```

```text
buy milk
call Ann

```

`File.create` replaces what the file held:

```ant
try File.write("todo.txt", "old list")
let out = try File.create("todo.txt")
defer { out.close() }
println(File.read("todo.txt"))
try out.writeText("new list")
try out.close()
println(File.read("todo.txt"))
```

```text
Ok{ value: "" }
Ok{ value: "new list" }
```

Refuse to replace a file that exists:

```ant
try File.write("draft.txt", "keep me")
println(File.create("draft.txt", replace: false))
```

```text
Err{ error: Error{ message: "open draft.txt: file exists" } }
```

## Scope

A file that exists keeps its mode when `File.create` truncates it:

```ant
try File.write("secret.txt", "old token")
try File.chmod("secret.txt", 0o600)
let out = try File.create("secret.txt")
defer { out.close() }
try out.writeText("new token")
try out.close()
let info = try File.stat("secret.txt")
println(String.fromInt(info.mode, base: 8))
```

```text
600
```

A path under a directory that is not there is `Err`. Make the directory with
[`File.mkdir`](mkdir.md) first:

```ant
println(File.create("logs/today.log"))

try File.mkdir("logs")
let out = try File.create("logs/today.log")
defer { out.close() }
try out.close()
println(File.exists("logs/today.log"))
```

```text
Err{ error: Error{ message: "open logs/today.log: no such file or directory" } }
true
```

A directory at `path` is `Err`:

```ant
try File.mkdir("cache")
println(File.create("cache"))
```

```text
Err{ error: Error{ message: "open cache: is a directory" } }
```

A symbolic link at `path` is followed, so the write goes to its target:

```ant
try File.write("target.txt", "old")
try File.symlink("target.txt", "current.txt")
let out = try File.create("current.txt")
defer { out.close() }
try out.writeText("new")
try out.close()
println(File.read("target.txt"))
```

```text
Ok{ value: "new" }
```

## Named arguments

### replace

With `replace: false`, `File.create` makes the file only when nothing is at `path`:

```ant
let out = try File.create("fresh.txt", replace: false)
defer { out.close() }
try out.writeText("first")
try out.close()
println(File.create("fresh.txt", replace: false))
println(File.read("fresh.txt"))
```

```text
Err{ error: Error{ message: "open fresh.txt: file exists" } }
Ok{ value: "first" }
```

A link at `path` counts as something there, even one that leads nowhere, and it is not followed:

```ant
try File.symlink("missing.txt", "planted.txt")
println(File.create("planted.txt", replace: false))
println(File.exists("missing.txt"))
```

```text
Err{ error: Error{ message: "open planted.txt: file exists" } }
false
```

## Applications

Write many lines through [`File.buffered`](buffered.md), and close it explicitly so that a failed
write-out is reported:

```ant
func writeReport(path: string, rows: Array[string]): Result[void, Error] {
  let out = File.buffered(try File.create(path))
  defer { out.close() }
  for row in rows {
    try out.writeLine(row)
  }
  out.close()
}

try writeReport("report.csv", ["name,score", "Ann,31", "Bob,27"])
println(try File.read("report.csv"))
```

```text
name,score
Ann,31
Bob,27

```

Pick the first free name, without overwriting the reports of earlier runs. `replace: false` makes
the check and the creation one step, so two programs never get the same name:

```ant
func claim(base: string): Result[string, Error] {
  for n in Iter.range(1, 100) {
    let name = "${base}-${n}.txt"
    let out = File.create(name, replace: false)
    if out is Err {
      if File.exists(name) { continue }
      return Result.err(out.error)
    }
    try out.value.close()
    return Result.ok(name)
  }
  Result.fail("no free name for ${base}")
}

println(claim("run"))
println(claim("run"))
```

```text
Ok{ value: "run-1.txt" }
Ok{ value: "run-2.txt" }
```

Add context to the error with [`Result.wrapErr`](../Result/wrapErr.md):

```ant
func save(path: string, text: string): Result[void, Error] {
  let out = try (File.create(path) | Result.wrapErr("save"))
  defer { out.close() }
  try out.writeText(text)
  out.close()
}

println(save("nowhere/out.txt", "hello"))
```

```text
Err{ error: Error{ message: "save: open nowhere/out.txt: no such file or directory" } }
```

Replace a file only once its new contents are complete. Write them to a temporary file beside it,
then rename that over the old one, so a reader sees the old file or the new one. The new file has
mode `0o600` from [`File.tempFile`](tempFile.md), not the mode the old one had:

```ant
func replaceFile(path: string, text: string): Result[void, Error] {
  let tmp = try File.tempFile(".page-*", dir: Path.dir(path))
  defer { File.remove(tmp) }
  try File.write(tmp, text)
  File.rename(tmp, path)
}

try File.write("page.html", "<p>old</p>")
try replaceFile("page.html", "<p>new</p>")
println(File.read("page.html"))
let info = try File.stat("page.html")
println(String.fromInt(info.mode, base: 8))
```

```text
Ok{ value: "<p>new</p>" }
600
```

## Properties and relations

A file created and closed without a write is empty:

```ant
let out = try File.create("empty.txt")
defer { out.close() }
try out.close()
let text = try File.read("empty.txt")
assert(text.isEmpty())
```

`File.create`, `File.writeText`, and `File.close` write what [`File.write`](write.md) writes in one
call:

```ant
let out = try File.create("a.txt")
defer { out.close() }
try out.writeText("same")
try out.close()
try File.write("b.txt", "same")
assert(File.read("a.txt") == File.read("b.txt"))
```

## Possible issues

The handle is for writing. Open the file again with [`File.open`](open.md) to read it:

```ant
let out = try File.create("log.txt")
defer { out.close() }
try out.writeLine("started")
println(out.readAll())
try out.close()
```

```text
Err{ error: Error{ message: "the file is not open for reading" } }
```

`File.create` truncates the file before the first write. When the program fails before it writes,
the old contents are gone. Write to a temporary file beside it with [`File.tempFile`](tempFile.md)
and [`File.rename`](rename.md) it into place to keep the old file until the new one is complete, as
`replaceFile` does in Applications:

```ant
func render(): Result[string, Error] {
  Result.fail("template missing")
}

try File.write("page.html", "<p>old</p>")
let out = try File.create("page.html")
defer { out.close() }
println(render())
println(File.read("page.html"))
```

```text
Err{ error: Error{ message: "template missing" } }
Ok{ value: "" }
```

The `Err` of `replace: false` carries only the OS's message, with nothing a program can test. Check
[`File.exists`](exists.md) after it to tell an existing file from other failures, as `claim` does in
Applications.

## See also

[File.open](open.md), [File.append](append.md), [File.write](write.md), [File.close](close.md),
[File.buffered](buffered.md), [File.writeLine](writeLine.md), [File.writeText](writeText.md),
[File.tempFile](tempFile.md), [File.rename](rename.md), [File.mkdir](mkdir.md)
