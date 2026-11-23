<!-- status: draft -->
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
- An entry that is neither is a regular file, or something rarer such as a socket or a device.
- A `File.Entry` is a snapshot of the listing. It has no size, time, or mode;
  [File.stat](../stat.md) has those, as a [File.Info](Info.md).

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
let { name, isDir } = (try File.readDir("inbox"))[0]
println(name)
println(isDir)
println((try File.readDir("inbox"))[0] == File.Entry{ name: "todo.txt", isDir: false, isSymlink: false })
```

```text
todo.txt
false
true
```

A link has `isSymlink` true and `isDir` false:

<!-- test: file links/sub/inner.txt
-->

```ant
try File.symlink("sub", "links/to-dir")
println((try File.readDir("links"))[1])
```

```text
File.Entry{ name: "to-dir", isDir: false, isSymlink: true }
```

## See also

- [File.readDir](../readDir.md): the entries of a directory
- [File.Info](Info.md): what [File.stat](../stat.md) reports of one path
