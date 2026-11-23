<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.Entry

```ant signature
type File.Entry = { name: string, isDir: bool, isSymlink: bool }
```

File.Entry is one entry of a directory, as [`File.readDir`](../readDir.md) reports it. name is the
entry's name, without the directory. isDir and isSymlink describe the entry itself and not what a
link points to: a symbolic link has isSymlink true and isDir false, also when it points to a
directory, and [`File.stat`](../stat.md) of its path tells what it points to.

<!-- end generated -->

## Details

- `name` is the entry's name alone, as [File.list](../list.md) gives it. Join it to the directory
  with [`Path.join`](../../Path/join.md) to get a path.
- `isDir` and `isSymlink` are never both true: a symbolic link is a link, whatever it points to.
- A `File.Entry` has no size, time, or mode; [File.stat](../stat.md) has those, as a
  [File.Info](Info.md).

## Basic examples

Get the entries of a directory with [File.readDir](../readDir.md) and read their fields:

<!-- test: file notes/todo.txt
buy milk
-->
<!-- test: file notes/archive/2025.md
# 2025
-->

```ant
for entry in try File.readDir("notes") {
  println("${entry.name} ${entry.isDir} ${entry.isSymlink}")
}
```

```text
archive true false
todo.txt false false
```

## Scope

A `File.Entry` is a record, so it destructures and compares as one:

<!-- test: file inbox/todo.txt
buy milk
-->

```ant
let entry = (try File.readDir("inbox"))[0]
let { name, isDir } = entry
println(name)
println(isDir)
println(entry == File.Entry{ name: "todo.txt", isDir: false, isSymlink: false })
```

```text
todo.txt
false
true
```

A `File.Entry` prints as a record:

<!-- test: file shelf/archive/2025.md
# 2025
-->

```ant
println((try File.readDir("shelf"))[0])
```

```text
File.Entry{ name: "archive", isDir: true, isSymlink: false }
```

## Properties and relations

A `File.Entry` holds the three fields of the same names that a [File.Info](Info.md) has when
[File.stat](../stat.md) is told not to follow a link:

<!-- test: file desk/todo.txt
buy milk
-->
<!-- test: file desk/archive/2025.md
# 2025
-->

```ant
try File.symlink("archive", "desk/old")
for entry in try File.readDir("desk") {
  let info = try File.stat(Path.join(["desk", entry.name]), follow: false)
  assert(entry.name == info.name)
  assert(entry.isDir == info.isDir)
  assert(entry.isSymlink == info.isSymlink)
}
```

## Possible issues

`isDir` is false for a link to a directory. Ask [File.stat](../stat.md), which follows the link,
what it points to:

```ant
try File.mkdir("shortcuts/sub")
try File.symlink("sub", "shortcuts/to-dir")
let entry = (try File.readDir("shortcuts")).find({ x.name == "to-dir" })
println(entry?.isDir)
println((try File.stat("shortcuts/to-dir")).isDir)
```

```text
false
true
```

`name` is not a path. Join it to the directory with [`Path.join`](../../Path/join.md) before you
read or stat the entry, as the same issue on [File.readDir](../readDir.md) shows.

## See also

- [File.readDir](../readDir.md): the entries of a directory
- [File.list](../list.md): the names alone
- [File.Info](Info.md): what [File.stat](../stat.md) reports of one path
- [Path.join](../../Path/join.md): the path of an entry, from the directory and its name
