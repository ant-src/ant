<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.copy

```ant signature
func File.copy(from: string, to: string, replace: bool = true): Result[void, Error]
```

File.copy copies the contents of the file at from to to, and then sets to's mode to from's. It
replaces a file at to unless replace is false, when it is Err if anything is at to; it is Err when
to is a directory, or from is not a regular file, such as a FIFO. It follows a symbolic link at
either end, does not copy times, and is Err before changing anything when from and to are the same
file.

<!-- end generated -->

## Details

- `to` is the path of the copy, never a directory to copy into. Name the file with
  [Path.join](../Path/join.md).
- The copy gets the mode of `from`, also when it replaces a file that had another mode. Times are
  not copied, so a build tool sees the copy as new.
- A symbolic link at `to` is followed: the file it points to is replaced, and the link stays.
- The copy is not atomic: a reader can see `to` half-written. Copy to a temporary file and
  [File.rename](rename.md) it to avoid that.
- [File.copyAll](copyAll.md) copies a directory.

## Basic examples

Copy a file:

```ant
try File.write("notes.txt", "buy milk\n")
try File.copy("notes.txt", "notes.bak")
print(try File.read("notes.bak"))
```

```text
buy milk
```

A missing source is an `Err` that names it:

```ant
println(File.copy("missing.txt", "copy.txt"))
```

```text
Err{ error: Error{ message: "stat missing.txt: no such file or directory" } }
```

## Scope

A file already at `to` is replaced, contents and mode:

```ant
try File.write("deploy.sh", "make deploy\n")
try File.chmod("deploy.sh", 0o755)
try File.write("old.sh", "make\n")
try File.chmod("old.sh", 0o600)

try File.copy("deploy.sh", "old.sh")
print(try File.read("old.sh"))
println(String.fromInt((try File.stat("old.sh")).mode, base: 8))
```

```text
make deploy
755
```

A link at `from` is followed, so the copy is a regular file with the target's contents:

```ant
try File.write("v1.txt", "version 1\n")
try File.symlink("v1.txt", "latest")
try File.copy("latest", "snapshot.txt")
print(try File.read("snapshot.txt"))
println((try File.stat("snapshot.txt", follow: false)).isSymlink)
```

```text
version 1
false
```

A link at `to` is followed too: the file it points to gets the new contents, and the link stays a
link. A link that leads nowhere gets its target created:

```ant
try File.write("config.v1", "port=80\n")
try File.write("config.v2", "port=8080\n")
try File.symlink("config.v1", "config")
try File.copy("config.v2", "config")
print(try File.read("config.v1"))
println((try File.stat("config", follow: false)).isSymlink)

try File.symlink("created.txt", "dangling")
try File.copy("config.v2", "dangling")
print(try File.read("created.txt"))
```

```text
port=8080
true
port=8080
```

A source that is not a regular file, such as a device, is an `Err`. `File.copy` does not open it, so
an endless source such as `/dev/zero` returns at once:

```ant
println(File.copy("/dev/zero", "zeros"))
```

```text
Err{ error: Error{ message: "copy /dev/zero: not a regular file" } }
```

## Named arguments

With `replace: false`, `File.copy` creates `to` only when nothing is there, and is otherwise an
`Err` that changes nothing:

```ant
try File.write("draft.txt", "first\n")
try File.copy("draft.txt", "keep.txt", replace: false)

try File.write("draft.txt", "second\n")
println(File.copy("draft.txt", "keep.txt", replace: false))
print(try File.read("keep.txt"))
```

```text
Err{ error: Error{ message: "open keep.txt: file exists" } }
first
```

Anything at `to` counts, a directory or a link that leads nowhere included. The link is not
followed:

```ant
try File.write("page.html", "<p>hi</p>\n")
try File.mkdir("public")
try File.symlink("elsewhere.html", "planted")
println(File.copy("page.html", "public", replace: false))
println(File.copy("page.html", "planted", replace: false))
println(File.exists("elsewhere.html"))
```

```text
Err{ error: Error{ message: "open public: file exists" } }
Err{ error: Error{ message: "open planted: file exists" } }
false
```

The `Err` does not say why the copy failed, so check `File.exists` after it to tell a destination
that is already there from other failures:

```ant
func copyNew(from: string, to: string): Result[bool, Error] {
  let r = File.copy(from, to, replace: false)
  if !r.isOk() && File.exists(to) {
    return Result.ok(false)
  }
  try r
  Result.ok(true)
}

try File.write("todo.txt", "write docs\n")
println(try copyNew("todo.txt", "todo.bak"))
println(try copyNew("todo.txt", "todo.bak"))
println(copyNew("missing.txt", "other.txt"))
```

```text
true
false
Err{ error: Error{ message: "stat missing.txt: no such file or directory" } }
```

## Applications

Copy a file into a directory, keeping its name:

```ant
func copyInto(src: string, dir: string): Result[string, Error] {
  let dest = Path.join([dir, Path.base(src)])
  try File.copy(src, dest)
  Result.ok(dest)
}

try File.write("report.csv", "id,total\n1,42\n")
try File.mkdir("archive")
println(try copyInto("report.csv", "archive"))
```

```text
archive/report.csv
```

Keep a backup before changing a file, without overwriting an older backup:

```ant
func editWithBackup(path: string, text: string): Result[void, Error] {
  try (File.copy(path, "${path}.orig", replace: false) | Result.wrapErr("back up ${path}"))
  File.write(path, text)
}

try File.write("hosts", "127.0.0.1 localhost\n")
try editWithBackup("hosts", "127.0.0.1 localhost\n10.0.0.2 db\n")
println(editWithBackup("hosts", ""))
print(try File.read("hosts.orig"))
```

```text
Err{ error: Error{ message: "back up hosts: open hosts.orig: file exists" } }
127.0.0.1 localhost
```

Publish a file so that no reader sees it half-written: copy it to a temporary file next to the
destination, then rename that over it:

```ant
func publish(src: string, dest: string): Result[void, Error] {
  let tmp = try File.tempFile(".publish-*", dir: Path.dir(dest))
  defer { File.remove(tmp) }
  try File.copy(src, tmp)
  File.rename(tmp, dest)
}

try File.mkdir("site")
try File.write("index.html", "<h1>Home</h1>\n")
try publish("index.html", "site/index.html")
println(try File.list("site"))
```

```text
["index.html"]
```

Copy every file that matches a pattern into a directory:

```ant
try File.mkdir("logs")
for name in ["a.log", "b.log", "c.txt"] {
  try File.write("logs/${name}", name)
}
try File.mkdir("saved")

for path in try File.glob("logs/*.log") {
  try File.copy(path, Path.join(["saved", Path.base(path)]))
}
println(try File.list("saved"))
```

```text
["a.log", "b.log"]
```

## Properties and relations

The copy has the same bytes and the same mode as the source:

```ant
try File.writeBytes("blob.bin", Bytes.fromArray([0, 255, 10, 13]))
try File.chmod("blob.bin", 0o640)
try File.copy("blob.bin", "blob2.bin")
assert((try File.readBytes("blob2.bin")) == (try File.readBytes("blob.bin")))
assert((try File.stat("blob2.bin")).mode == (try File.stat("blob.bin")).mode)
```

The modified time is not copied:

```ant
let past = Time.fromEpochMillis(0)
try File.write("old.txt", "")
try File.touch("old.txt", time: past)
try File.copy("old.txt", "new.txt")
assert((try File.stat("new.txt")).modified != past)
```

## Possible issues

A directory at `to` is an `Err`, not a place to copy into. Join the file name to the directory
instead, as in Applications:

```ant
try File.write("photo.jpg", "jpeg")
try File.mkdir("album")
println(File.copy("photo.jpg", "album"))
```

```text
Err{ error: Error{ message: "copy photo.jpg to album: album is a directory" } }
```

A directory at `from` is an `Err`; [File.copyAll](copyAll.md) copies a tree:

```ant
try File.mkdir("src")
println(File.copy("src", "src-copy"))
```

```text
Err{ error: Error{ message: "copy src: is a directory; File.copyAll copies a tree" } }
```

`File.copy` does not create missing parent directories. Make them with [File.mkdir](mkdir.md) first:

```ant
try File.write("a.txt", "a")
println(File.copy("a.txt", "out/a.txt"))
```

```text
Err{ error: Error{ message: "open out/a.txt: no such file or directory" } }
```

Copying a file onto itself is an `Err` and leaves the file as it was, whatever path names it:

```ant
try File.write("data.txt", "keep me\n")
try File.symlink("data.txt", "alias")
println(File.copy("data.txt", "./data.txt"))
println(File.copy("data.txt", "alias"))
print(try File.read("data.txt"))
```

```text
Err{ error: Error{ message: "copy data.txt to ./data.txt: the same file" } }
Err{ error: Error{ message: "copy data.txt to alias: the same file" } }
keep me
```

To replace a link at `to` with a copy, rather than its target, remove the link first:

```ant
try File.write("shared.txt", "shared\n")
try File.write("mine.txt", "mine\n")
try File.symlink("shared.txt", "local.txt")
try File.remove("local.txt")
try File.copy("mine.txt", "local.txt")
print(try File.read("shared.txt"))
println((try File.stat("local.txt", follow: false)).isSymlink)
```

```text
shared
false
```

## See also

[File.copyAll](copyAll.md) · [File.rename](rename.md) · [File.write](write.md) ·
[File.tempFile](tempFile.md) · [File.remove](remove.md) · [File.stat](stat.md) ·
[Path.join](../Path/join.md) · [Path.base](../Path/base.md)
