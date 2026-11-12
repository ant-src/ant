<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.rename

```ant signature
func File.rename(from: string, to: string): Result[void, Error]
```

File.rename renames or moves the file or directory at from to the path to, replacing a file at to.
It is Err when a directory is at to, since there is no "into" form, and across file systems; it is
Ok and nothing changes when from and to are the same file.

<!-- end generated -->

## Details

- `File.rename` replaces a file at `to` without asking. It has no `replace:` parameter.
- On Unix the replacement is atomic: a reader of `to` sees the old file or the new one, never a mix.
  Elsewhere atomicity is not promised.
- At a symbolic link, `File.rename` moves the link itself, not what it points at.

## Basic examples

Rename a file:

```ant
try File.write("draft.txt", "hello")
println(File.rename("draft.txt", "final.txt"))
println(File.exists("draft.txt"))
println(try File.read("final.txt"))
```

```text
Ok{ value: void }
false
hello
```

A source that does not exist is `Err`:

```ant
println(File.rename("missing.txt", "other.txt"))
```

```text
Err{ error: Error{ message: "rename missing.txt other.txt: no such file or directory" } }
```

## Scope

### Directories

A directory moves with everything in it:

```ant
try File.mkdir("logs/2026")
try File.write("logs/2026/app.log", "started")
try File.rename("logs", "archive")
println(File.exists("logs"))
println(try File.walk("archive"))
```

```text
false
["2026/app.log"]
```

### Moving to another directory

A `to` in another directory moves the file there.

```ant
try File.mkdir("bin")
try File.write("deploy.sh", "echo deploying")
try File.rename("deploy.sh", "bin/deploy.sh")
println(try File.read("bin/deploy.sh"))
```

```text
echo deploying
```

### Replacing a file

A file already at `to` is replaced:

```ant
try File.write("config.new", "port = 8080")
try File.write("config", "port = 80")
try File.rename("config.new", "config")
println(try File.read("config"))
println(File.exists("config.new"))
```

```text
port = 8080
false
```

### Symbolic links

The link moves and keeps its target, which stays where it was:

```ant
try File.write("v2.txt", "version 2")
try File.symlink("v2.txt", "next")
try File.rename("next", "current")
println(try File.readLink("current"))
println(try File.read("v2.txt"))
```

```text
v2.txt
version 2
```

## Applications

Save a file so that no reader ever sees it half-written: write a temporary file beside it, then
rename it into place. The deferred `File.remove` cleans up when a step fails, and is an `Err` that
nobody reads after the rename succeeds:

```ant
func save(path: string, text: string): Result[void, Error] {
  let tmp = try File.tempFile(".save-*", dir: Path.dir(path))
  defer { File.remove(tmp) }
  try File.write(tmp, text)
  File.rename(tmp, path)
}

try File.write("settings.json", "{\"theme\": \"light\"}")
try save("settings.json", "{\"theme\": \"dark\"}")
println(try File.read("settings.json"))
println(try File.glob(".save-*", hidden: true))
```

```text
{"theme": "dark"}
[]
```

Change the extension of every `.txt` file in a directory:

<!-- test: file notes/a.txt
a
-->
<!-- test: file notes/b.txt
b
-->
<!-- test: file notes/c.md
c
-->

```ant
let dir = "notes"
for name in try File.list(dir) {
  if Path.ext(name) != ".txt" { continue }
  let from = Path.join([dir, name])
  try File.rename(from, from.trimSuffix(".txt") + ".md")
}
println(try File.list(dir))
```

```text
["a.md", "b.md", "c.md"]
```

Move a file into a directory, joining the directory and the file's name, and add context to a
failure with `Result.wrapErr`:

```ant
func moveInto(path: string, dir: string): Result[void, Error] {
  try File.mkdir(dir)
  File.rename(path, Path.join([dir, Path.base(path)]))
    | Result.wrapErr("archive ${path}")
}

try File.write("report.csv", "id,total")
try moveInto("report.csv", "done/2026")
println(try File.walk("done"))
println(moveInto("report.csv", "done/2026"))
```

```text
["2026/report.csv"]
Err{ error: Error{ message: "archive report.csv: rename report.csv done/2026/report.csv: no such file or directory" } }
```

## Properties and relations

After an `Ok`, nothing is at `from`, and `to` holds what `from` held:

```ant
let text = "contents"
try File.write("old.txt", text)
try File.rename("old.txt", "new.txt")
assert(!File.exists("old.txt"))
assert(try File.read("new.txt") == text)
```

## Possible issues

`File.rename` replaces a file at `to` without warning. Checking with [File.exists](exists.md) first
is not atomic: another program can create `to` between the check and the rename.

```ant
try File.write("notes.txt", "keep me")
try File.write("draft.txt", "scratch")
try File.rename("draft.txt", "notes.txt")
println(try File.read("notes.txt"))
```

```text
scratch
```

`to` is the new path of the file, not a directory to move it into. To move a file into a directory,
join the directory and the name with [Path.join](../Path/join.md):

```ant
try File.mkdir("box")
try File.write("item.txt", "x")
try File.rename("item.txt", Path.join(["box", "item.txt"]))
println(try File.walk("box"))
```

```text
["item.txt"]
```

The directory of `to` must exist. Create it with [File.mkdir](mkdir.md) first:

```ant
try File.write("photo.jpg", "jpeg")
println(File.rename("photo.jpg", "albums/2026/photo.jpg"))
try File.mkdir("albums/2026")
println(File.rename("photo.jpg", "albums/2026/photo.jpg"))
```

```text
Err{ error: Error{ message: "rename photo.jpg albums/2026/photo.jpg: no such file or directory" } }
Ok{ value: void }
```

## See also

[File.copy](copy.md) · [File.copyAll](copyAll.md) · [File.remove](remove.md) ·
[File.tempFile](tempFile.md) · [File.mkdir](mkdir.md) · [File.exists](exists.md) ·
[Path.join](../Path/join.md) · [Path.base](../Path/base.md)
