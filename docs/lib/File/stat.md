<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.stat

```ant signature
func File.stat(path: string, follow: bool = true): Result[File.Info, Error]
```

File.stat returns the name, size, kind, modification time and mode of the file at path. It follows a
symbolic link and describes its target, unless follow is false, when it describes the link itself.

<!-- end generated -->

## Details

- A dangling link is `Err` unless `follow` is `false`, and `isSymlink` is true only with
  `follow: false`.
- `modified` is the time of the last write, as precise as the file system keeps it: to the
  nanosecond on ext4 and APFS, to the second on HFS+, to 2 seconds on FAT.
- `mode` holds the permissions, not the kind of file: `isDir` and `isSymlink` say that.
- `File.stat` is `Err` when nothing is at `path`.

## Basic examples

Read the size of a file:

```ant
try File.write("notes.txt", "buy milk")
let info = try File.stat("notes.txt")
println("${info.name}: ${info.size} bytes")
```

```text
notes.txt: 8 bytes
```

Tell a directory from a file:

```ant
try File.mkdir("logs")
try File.write("logs/today.log", "started")
println(File.stat("logs").map({ x.isDir }))
println(File.stat("logs/today.log").map({ x.isDir }))
```

```text
Ok{ value: true }
Ok{ value: false }
```

A missing path is an `Err`:

```ant
println(File.stat("missing.txt"))
```

```text
Err{ error: Error{ message: "stat missing.txt: no such file or directory" } }
```

## Scope

Read the name and size of a nested file:

```ant
try File.mkdir("src/lib")
try File.write("src/lib/geo.ant", "")
let info = try File.stat("src/lib/geo.ant")
println(info.name)
println(info.size)
```

```text
geo.ant
0
```

A Unicode name works:

```ant
try File.write("café.txt", "☕")
println((try File.stat("café.txt")).name)
```

```text
café.txt
```

`modified` is a `Time.Instant`. Set it with [File.touch](touch.md) and read it back:

```ant
let release = try Time.parseInstant("2026-03-01T12:00:00Z")
try File.write("CHANGES.md", "# 1.0")
try File.touch("CHANGES.md", time: release)
println((try File.stat("CHANGES.md")).modified)
```

```text
2026-03-01T12:00:00Z
```

`mode` is a number; print it in octal with `String.fromInt`:

```ant
try File.write("deploy.sh", "echo deploying\n")
try File.chmod("deploy.sh", 0o755)
let info = try File.stat("deploy.sh")
println(String.fromInt(info.mode, base: 8))
```

```text
755
```

A path that runs through a file is an `Err`, as is an empty path:

```ant
try File.write("plain.txt", "text")
println(File.stat("plain.txt/inner"))
println(File.stat(""))
```

```text
Err{ error: Error{ message: "stat plain.txt/inner: not a directory" } }
Err{ error: Error{ message: "empty path" } }
```

## Named arguments

### follow

By default `File.stat` follows a link and describes its target. With `follow: false` it describes
the link itself:

```ant
try File.write("v2.txt", "second")
try File.chmod("v2.txt", 0o600)
try File.symlink("v2.txt", "current")

let target = try File.stat("current")
let link = try File.stat("current", follow: false)
println("${target.name} ${target.isSymlink} ${String.fromInt(target.mode, base: 8)}")
println("${link.name} ${link.isSymlink}")
```

```text
current false 600
current true
```

A link whose target is missing is an `Err` when followed, and `Ok` with `follow: false`:

```ant
try File.symlink("gone.txt", "broken")
println(File.stat("broken"))
println(File.stat("broken", follow: false).map({ x.isSymlink }))
```

```text
Err{ error: Error{ message: "stat broken: no such file or directory" } }
Ok{ value: true }
```

A link to a directory is a directory when followed, and a link, not a directory, with
`follow: false`:

```ant
try File.mkdir("releases/v1.5.0")
try File.symlink("releases/v1.5.0", "live")
let followed = try File.stat("live")
let link = try File.stat("live", follow: false)
println("${followed.isDir} ${followed.isSymlink}")
println("${link.isDir} ${link.isSymlink}")
```

```text
true false
false true
```

## Applications

Add up the sizes of the files below a directory:

```ant
try File.mkdir("site/css")
try File.write("site/index.html", "<h1>Home</h1>")
try File.write("site/css/main.css", "body { margin: 0 }")

var total = 0
for path in try File.walk("site") {
  total += (try File.stat(Path.join(["site", path]))).size
}
println("${total} bytes")
```

```text
31 bytes
```

Rebuild an output only when its source changed after it, as `make` does:

```ant
func stale(source: string, output: string): Result[bool, Error] {
  if !File.exists(output) { return Result.ok(true) }
  let src = try File.stat(source)
  let out = try File.stat(output)
  Result.ok(Time.compare(src.modified, out.modified) > 0)
}

try File.write("app.ant", "println(1)")
try File.write("app.js", "console.log(1)")
try File.touch("app.ant", time: try Time.parseInstant("2026-05-02T09:00:00Z"))
try File.touch("app.js", time: try Time.parseInstant("2026-05-01T09:00:00Z"))
println(stale("app.ant", "app.js"))
println(stale("app.ant", "app.wasm"))
try File.touch("app.js", time: try Time.parseInstant("2026-05-03T09:00:00Z"))
println(stale("app.ant", "app.js"))
```

```text
Ok{ value: true }
Ok{ value: true }
Ok{ value: false }
```

List a directory with the kind of each entry, without following links:

```ant
try File.mkdir("project/src")
try File.write("project/README.md", "# Project")
try File.symlink("README.md", "project/readme")

for name in try File.list("project") {
  let info = try File.stat(Path.join(["project", name]), follow: false)
  let kind = if info.isSymlink { "link" } else if info.isDir { "dir" } else { "file" }
  println("${kind}\t${name}")
}
```

```text
file	README.md
link	readme
dir	src
```

Check that a script is executable by its owner before running it:

```ant
func checkExecutable(path: string): Result[void, Error] {
  let info = try (File.stat(path) | Result.wrapErr("check ${path}"))
  if Bits.and(info.mode, 0o100) == 0 {
    return Result.fail("${path} is not executable; run chmod +x ${path}")
  }
  Result.ok(void)
}

try File.write("build.sh", "echo building\n")
try File.chmod("build.sh", 0o644)
println(checkExecutable("build.sh"))
println(checkExecutable("test.sh"))
```

```text
Err{ error: Error{ message: "build.sh is not executable; run chmod +x build.sh" } }
Err{ error: Error{ message: "check test.sh: stat test.sh: no such file or directory" } }
```

## Properties and relations

`size` matches the length of the file's bytes, not of its text:

```ant
try File.write("menu.txt", "café ☕\n")
let info = try File.stat("menu.txt")
assert(info.size == (try File.readBytes("menu.txt")).length())
assert(info.size != (try File.read("menu.txt")).length())
```

[File.exists](exists.md) follows links as `File.stat` does, so a dangling link is `false` for one
and `Err` for the other:

```ant
try File.write("a.txt", "")
try File.symlink("a.txt", "good")
try File.symlink("gone.txt", "bad")
for path in ["a.txt", "good", "bad", "nothing"] {
  assert(File.exists(path) == File.stat(path).isOk())
}
```

Following a link gives the target's size, mode and time:

```ant
try File.write("data.csv", "a,b\n1,2\n")
try File.symlink("data.csv", "latest.csv")
let file = try File.stat("data.csv")
let viaLink = try File.stat("latest.csv")
assert(viaLink.size == file.size)
assert(viaLink.mode == file.mode)
assert(viaLink.modified == file.modified)
```

## Possible issues

Without `follow: false`, `isSymlink` is false, so a check for links that leaves it out never finds
one:

```ant
try File.write("todo.txt", "")
try File.symlink("todo.txt", "today")
println((try File.stat("today")).isSymlink)
println((try File.stat("today", follow: false)).isSymlink)
```

```text
false
true
```

`size` and the text's length differ for non-ASCII text. For the number of characters, read the text
and take its length:

```ant
try File.write("greeting.txt", "héllo")
println((try File.stat("greeting.txt")).size)
println((try File.read("greeting.txt")).length())
```

```text
6
5
```

`modified` is a `Time.Instant`, which `<` and `>` do not order; `src.modified > out.modified` is a
compile error. Use `Time.compare`:

```ant
try File.write("older.txt", "")
try File.write("newer.txt", "")
try File.touch("older.txt", time: try Time.parseInstant("2026-01-01T00:00:00Z"))
try File.touch("newer.txt", time: try Time.parseInstant("2026-02-01T00:00:00Z"))
let older = try File.stat("older.txt")
let newer = try File.stat("newer.txt")
println(Time.compare(older.modified, newer.modified))
```

```text
-1
```

## See also

[File.exists](exists.md) · [File.Info](types/Info.md) · [File.touch](touch.md) ·
[File.chmod](chmod.md) · [File.readLink](readLink.md) · [File.realPath](realPath.md) ·
[File.list](list.md) · [File.walk](walk.md)
