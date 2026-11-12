<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.Info

```ant signature
type File.Info = { name: string, size: int, isDir: bool, isSymlink: bool, modified: Time.Instant, mode: int }
```

File.Info describes the file at a path, as [`File.stat`](../stat.md) reports it. modified is the
time of the last write. mode holds the permission bits and the setuid, setgid and sticky bits, at
most 0o7777; on Windows a file is 0o444 when read-only and 0o666 when not, and a directory adds
0o111. isSymlink is true only when stat was told not to follow a link. name is the last element of
the path, also when a link is followed. size is a regular file's length in bytes; for a directory,
and for a link described with follow false, it is what the host reports, as is a link's own mode.

<!-- end generated -->

## Details

- `mode` is a plain `int` and prints in decimal: `0o644` prints as `420`. Print it with
  `String.fromInt(mode, base: 8)`.
- `mode` does not say what kind of file it is; `isDir` and `isSymlink` do.
- `modified` is rounded as the file system keeps times: to 2 seconds on FAT, to the second on HFS+,
  to the nanosecond on ext4 and APFS. The access time is not in `File.Info`.
- A `File.Info` is a snapshot: it does not change when the file does. Call [File.stat](../stat.md)
  again for the current values.

## Basic examples

Get the `File.Info` of a file with [File.stat](../stat.md) and read its fields:

```ant
try File.write("notes.txt", "buy milk\n")
let info = try File.stat("notes.txt")
println(info.name)
println(info.size)
println(info.isDir)
```

```text
notes.txt
9
false
```

A `File.Info` prints as a record. [File.touch](../touch.md) fixes `modified` here so the output is
the same on every run:

```ant
try File.write("notes.txt", "buy milk\n")
try File.touch("notes.txt", time: try Time.parseInstant("2026-03-01T09:30:00Z"))
try File.chmod("notes.txt", 0o644)
println(try File.stat("notes.txt"))
```

```text
File.Info{ name: "notes.txt", size: 9, isDir: false, isSymlink: false, modified: 2026-03-01T09:30:00Z, mode: 420 }
```

## Scope

A directory has `isDir` set:

```ant
try File.mkdir("photos")
let info = try File.stat("photos")
println("${info.name} ${info.isDir}")
```

```text
photos true
```

By default `File.stat` follows a symbolic link and describes its target, so `isSymlink` is false.
With `follow: false` it describes the link itself:

```ant
try File.write("v2.txt", "second")
try File.symlink("v2.txt", "latest")
let target = try File.stat("latest")
let link = try File.stat("latest", follow: false)
println("${target.name} ${target.isSymlink} ${target.size}")
println("${link.name} ${link.isSymlink}")
```

```text
latest false 6
latest true
```

`mode` holds the setuid, setgid and sticky bits too, so it is a number in `0 ..= 0o7777`:

```ant
try File.mkdir("shared")
try File.chmod("shared", 0o1777)
let info = try File.stat("shared")
println(String.fromInt(info.mode, base: 8))
```

```text
1777
```

Destructure the fields you need:

```ant
try File.write("report.csv", "id,total\n1,40\n")
let { name, size } = try File.stat("report.csv")
println("${name}: ${size} bytes")
```

```text
report.csv: 14 bytes
```

`File.Info` is a record type, so a literal builds one, for example to test a function that takes a
`File.Info` without touching the disk:

```ant
func describe(info: File.Info): string {
  if info.isDir { "${info.name}/" } else { "${info.name} (${info.size})" }
}

let when = try Time.parseInstant("2026-01-01T00:00:00Z")
let file = File.Info{ name: "a.txt", size: 3, isDir: false, isSymlink: false, modified: when, mode: 0o644 }
let dir = File.Info{ name: "src", size: 0, isDir: true, isSymlink: false, modified: when, mode: 0o755 }
println(describe(file))
println(describe(dir))
```

```text
a.txt (3)
src/
```

## Applications

Add up the size of every file in a tree:

```ant
try File.mkdir("site/css")
try File.write("site/index.html", "<h1>Hi</h1>")
try File.write("site/css/main.css", "body{}")

var total = 0
for path in try File.walk("site") {
  total += (try File.stat(Path.join(["site", path]))).size
}
println("${total} bytes")
```

```text
17 bytes
```

Rebuild an output only when its source changed after it, as `make` does:

```ant
func stale(source: string, output: string): Result[bool, Error] {
  if !File.exists(output) { return Result.ok(true) }
  let src = try File.stat(source)
  let out = try File.stat(output)
  Result.ok(Time.compare(src.modified, out.modified) > 0)
}

try File.write("page.md", "# Title")
try File.write("page.html", "<h1>Title</h1>")
try File.touch("page.md", time: try Time.parseInstant("2026-05-02T10:00:00Z"))
try File.touch("page.html", time: try Time.parseInstant("2026-05-01T10:00:00Z"))
println(stale("page.md", "page.html"))
println(stale("page.md", "page.pdf"))
```

```text
Ok{ value: true }
Ok{ value: true }
```

Find the most recently changed file in a directory:

```ant
try File.mkdir("backups")
let days = ["2026-04-01", "2026-04-03", "2026-04-02"]
for day in days {
  let path = "backups/${day}.tar"
  try File.write(path, "")
  try File.touch(path, time: try Time.parseInstant("${day}T03:00:00Z"))
}

let infos: Array[File.Info] = []
for name in try File.list("backups") {
  infos.push(try File.stat("backups/${name}"))
}
let newest = infos.maxBy({ x.modified.epochNanos() })
println(newest?.name)
```

```text
2026-04-03.tar
```

Show a mode as `ls -l` does, one letter per permission bit:

```ant
func permissions(info: File.Info): string {
  var out = if info.isDir { "d" } else { "-" }
  for shift in [6, 3, 0] {
    let rwx = Bits.and(Bits.shiftRight(info.mode, shift), 0o7)
    out += if Bits.and(rwx, 4) != 0 { "r" } else { "-" }
    out += if Bits.and(rwx, 2) != 0 { "w" } else { "-" }
    out += if Bits.and(rwx, 1) != 0 { "x" } else { "-" }
  }
  out
}

try File.write("deploy.sh", "make\n")
try File.chmod("deploy.sh", 0o750)
try File.mkdir("public")
try File.chmod("public", 0o755)
println(permissions(try File.stat("deploy.sh")))
println(permissions(try File.stat("public")))
```

```text
-rwxr-x---
drwxr-xr-x
```

Split the entries of a directory into directories and files:

```ant
try File.mkdir("project")
try File.mkdir("project/src")
try File.write("project/README.md", "")
try File.write("project/go.mod", "")

let dirs: Array[string] = []
let files: Array[string] = []
for name in try File.list("project") {
  let info = try File.stat("project/${name}")
  if info.isDir { dirs.push(name) } else { files.push(name) }
}
println(dirs)
println(files)
```

```text
["src"]
["README.md", "go.mod"]
```

## Properties and relations

`size` is the number of bytes [File.readBytes](../readBytes.md) reads:

```ant
try File.mkdir("data")
for [path, text] in [["data/a.txt", ""], ["data/café ☕.txt", "café ☕"], ["data/b.txt", "x\ny\n"]] {
  try File.write(path, text)
  let info = try File.stat(path)
  assert(info.size == (try File.readBytes(path)).length())
}
```

`modified` reads back the time [File.touch](../touch.md) set, and `mode` the mode
[File.chmod](../chmod.md) set:

```ant
let when = try Time.parseInstant("2026-07-04T12:00:00Z")
try File.touch("stamp", time: when)
try File.chmod("stamp", 0o640)
let info = try File.stat("stamp")
assert(info.modified == when)
assert(info.mode == 0o640)
```

With the default `follow`, `isSymlink` is false, even for a link:

```ant
try File.write("target.txt", "")
try File.symlink("target.txt", "alias")
for path in ["target.txt", "alias"] {
  assert(!(try File.stat(path)).isSymlink)
}
assert((try File.stat("alias", follow: false)).isSymlink)
```

## Possible issues

`mode` prints in decimal. Format it in base 8 to see the familiar digits:

```ant
try File.write("notes.txt", "")
try File.chmod("notes.txt", 0o644)
let info = try File.stat("notes.txt")
println(info.mode)
println(String.fromInt(info.mode, base: 8))
```

```text
420
644
```

Checking `isSymlink` after a plain `File.stat` never finds a link, since `File.stat` followed it.
Pass `follow: false`:

```ant
try File.write("real.txt", "")
try File.symlink("real.txt", "shortcut")
println((try File.stat("shortcut")).isSymlink)
println((try File.stat("shortcut", follow: false)).isSymlink)
```

```text
false
true
```

With `follow: false`, a link to a directory describes the link, so `isDir` is false:

```ant
try File.mkdir("photos")
try File.symlink("photos", "pics")
let target = try File.stat("pics")
let link = try File.stat("pics", follow: false)
println("${target.isDir} ${target.isSymlink}")
println("${link.isDir} ${link.isSymlink}")
```

```text
true false
false true
```

A link whose target is missing is `Err` from a plain `File.stat`, and `follow: false` still
describes it:

```ant
try File.symlink("missing.txt", "broken")
println(File.stat("broken") is Err)
println((try File.stat("broken", follow: false)).isSymlink)
```

```text
true
true
```

A `File.Info` does not follow later changes to the file:

```ant
try File.write("log.txt", "one\n")
let before = try File.stat("log.txt")
try File.write("log.txt", "one\ntwo\n")
println(before.size)
println((try File.stat("log.txt")).size)
```

```text
4
8
```

## See also

[File.stat](../stat.md) · [File.chmod](../chmod.md) · [File.touch](../touch.md) ·
[File.list](../list.md) · [File.walk](../walk.md) · [File.exists](../exists.md) ·
[Time.Instant](../../Time/types/Instant.md) · [String.fromInt](../../String/fromInt.md)
