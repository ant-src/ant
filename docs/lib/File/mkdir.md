<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.mkdir

```ant signature
func File.mkdir(path: string): Result[void, Error]
```

File.mkdir creates the directory at path, with any parents it lacks, each made as the path names it,
so x/../y makes x too. It is Ok when the directory already exists, a link to one included, and Err
at a dangling link.

<!-- end generated -->

## Details

- A new directory gets mode `0o777` less the process's umask, so `0o755` under the usual umask
  `022`. Use [File.chmod](chmod.md) afterwards for an exact mode.
- A directory that already exists is left as it is, mode included.
- It is `Err` when a file that is not a directory is at `path` or at one of its parents.

## Basic examples

Create a directory and the parents it needs in one call:

```ant
println(File.mkdir("build/assets/icons"))
let info = try File.stat("build/assets")
println(info.isDir)
```

```text
Ok{ value: void }
true
```

Call it again and it is still `Ok`:

```ant
try File.mkdir("cache")
println(File.mkdir("cache"))
```

```text
Ok{ value: void }
```

## Scope

The current directory, `"."`, always exists, so `File.mkdir(".")` is `Ok`:

```ant
println(File.mkdir("."))
```

```text
Ok{ value: void }
```

## Applications

Write a file at a nested path, making its directory first. [Path.dir](../Path/dir.md) of a bare file
name is `"."`, so the helper works for a file in the current directory too:

```ant
func save(path: string, text: string): Result[void, Error] {
  try File.mkdir(Path.dir(path))
  File.write(path, text)
}

try save("reports/2026/11/summary.txt", "all green")
try save("notes.txt", "top level")
println(try File.walk("reports"))
println(try File.read("notes.txt"))
```

```text
["2026/11/summary.txt"]
top level
```

Keep a script's files under a directory of its own name in the user's cache directory:

<!-- test: check-only the cache directory depends on the host -->

```ant
let dir = Path.join([try Os.cacheDir(), "thumbnailer"])
try File.mkdir(dir)
try File.write(Path.join([dir, "index.json"]), "{}")
println("cache in ${dir}")
```

Tighten an ordinary directory so that only its owner can enter it. Between the two calls it has the
umask mode, so for a directory that must be private from the moment it exists, use
[File.tempDir](tempDir.md), which is `0o700` from the start:

```ant
let drafts = "drafts"
try File.mkdir(drafts)
try File.chmod(drafts, 0o700)
let info = try File.stat(drafts)
println(info.mode | String.fromInt(base: 8))
```

```text
700
```

## Properties and relations

After an `Ok`, the path is a directory, and [File.exists](exists.md) agrees:

```ant
let dir = "a/b/c"
try File.mkdir(dir)
let info = try File.stat(dir)
assert(info.isDir)
assert(File.exists(dir))
```

## Possible issues

`File.mkdir` does not reset the mode of a directory that is already there. A directory left at
`0o755` stays so after another call:

```ant
try File.mkdir("shared")
try File.chmod("shared", 0o755)
try File.mkdir("shared")
let info = try File.stat("shared")
println(info.mode | String.fromInt(base: 8))
```

```text
755
```

For an exact mode, call [File.chmod](chmod.md) after every `File.mkdir`:

```ant
try File.mkdir("shared")
try File.chmod("shared", 0o755)
try File.mkdir("shared")
try File.chmod("shared", 0o700)
let info = try File.stat("shared")
println(info.mode | String.fromInt(base: 8))
```

```text
700
```

A file in the way, at the path or at one of its parents, makes it `Err`:

<!-- test: file notes
not a directory
-->

```ant
println(File.mkdir("notes"))
println(File.mkdir("notes/2026"))
```

```text
Err{ error: Error{ message: "mkdir notes: not a directory" } }
Err{ error: Error{ message: "mkdir notes: not a directory" } }
```

[File.write](write.md) does not create missing directories. Call `File.mkdir` first:

```ant
println(File.write("logs/today.log", "started"))
try File.mkdir("logs")
println(File.write("logs/today.log", "started"))
```

```text
Err{ error: Error{ message: "open logs/today.log: no such file or directory" } }
Ok{ value: void }
```

## See also

[File.tempDir](tempDir.md) · [File.chmod](chmod.md) · [File.remove](remove.md) ·
[File.removeAll](removeAll.md) · [File.list](list.md) · [File.stat](stat.md) ·
[File.exists](exists.md) · [Path.dir](../Path/dir.md)
