<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.readDir

```ant signature
func File.readDir(dir: string): Result[Array[File.Entry], Error]
```

File.readDir returns the entries directly in dir, in the order of [`File.list`](list.md), each with
its name and its kind. The kind comes with the listing, so telling directories from files costs no
[`File.stat`](stat.md) per entry. It follows a link at dir. An entry that is a symbolic link is
reported as the link, whatever it points to.

<!-- end generated -->

## Details

- Each entry is a [File.Entry](types/Entry.md) with the fields `name`, `isDir`, and `isSymlink`.
- The entries come in lexical order, the order of [File.list](list.md).
- Names that begin with a dot and symbolic links are entries like any other, as in
  [File.list](list.md).
- The names are bare, as those of [File.list](list.md) are: to read or stat an entry, join its name
  to `dir` with [`Path.join`](../Path/join.md).
- The kind is the one the directory holds. A symbolic link has `isSymlink` true and `isDir` false,
  also when it points to a directory; [File.stat](stat.md) of the path tells what it points to.
- An entry has no size, time, or mode. Call [File.stat](stat.md) for those.
- An empty directory gives `Ok` with an empty array. A missing `dir`, or a `dir` that is a file, is
  `Err`.

## Basic examples

List the entries of a directory with their kinds:

<!-- test: file notes/todo.txt
buy milk
-->
<!-- test: file notes/archive/2025.md
# 2025
-->

```ant
for entry in try File.readDir("notes") {
  println(entry)
}
```

```text
File.Entry{ name: "archive", isDir: true, isSymlink: false }
File.Entry{ name: "todo.txt", isDir: false, isSymlink: false }
```

Mark the directories with a slash, as `ls -p` does:

<!-- test: file notes/todo.txt
buy milk
-->
<!-- test: file notes/archive/2025.md
# 2025
-->

```ant
for { name, isDir } in try File.readDir("notes") {
  println(if isDir { name + "/" } else { name })
}
```

```text
archive/
todo.txt
```

## Scope

The entries come in the order of [File.list](list.md), and with the same names:

<!-- test: file mixed/apple
-->
<!-- test: file mixed/Zeta
-->
<!-- test: file mixed/10.txt
-->
<!-- test: file mixed/9.txt
-->

```ant
let entries = try File.readDir("mixed")
println(entries.map({ x.name }))
println(entries.map({ x.name }) == try File.list("mixed"))
```

```text
["10.txt", "9.txt", "Zeta", "apple"]
true
```

A symbolic link is reported as a link, whether it points to a file, to a directory, or to nothing:

<!-- test: file links/target.txt
hello
-->
<!-- test: file links/sub/inner.txt
-->

```ant
try File.symlink("target.txt", "links/to-file")
try File.symlink("sub", "links/to-dir")
try File.symlink("nowhere", "links/dangling")
for entry in try File.readDir("links") {
  println(entry)
}
```

```text
File.Entry{ name: "dangling", isDir: false, isSymlink: true }
File.Entry{ name: "sub", isDir: true, isSymlink: false }
File.Entry{ name: "target.txt", isDir: false, isSymlink: false }
File.Entry{ name: "to-dir", isDir: false, isSymlink: true }
File.Entry{ name: "to-file", isDir: false, isSymlink: true }
```

A link at `dir` itself is followed:

<!-- test: file shortcuts/sub/inner.txt
-->

```ant
try File.symlink("sub", "shortcuts/to-dir")
println(File.readDir("shortcuts/to-dir"))
```

```text
Ok{ value: [File.Entry{ name: "inner.txt", isDir: false, isSymlink: false }] }
```

An empty directory gives an empty array, and a path that is not a directory is `Err`:

```ant
try File.mkdir("empty")
try File.write("notes.txt", "")
println(File.readDir("empty"))
println(File.readDir("notes.txt") is Err)
println(File.readDir("missing") is Err)
```

```text
Ok{ value: [] }
true
true
```

## Applications

Split a listing into directories and files:

<!-- test: file site/index.html
-->
<!-- test: file site/css/main.css
-->
<!-- test: file site/img/logo.svg
-->

```ant
let entries = try File.readDir("site")
println(entries.filter({ x.isDir }).map({ x.name }))
println(entries.filter({ !x.isDir }).map({ x.name }))
```

```text
["css", "img"]
["index.html"]
```

Count the files below a directory, entering directories and not links:

<!-- test: file web/index.html
-->
<!-- test: file web/css/main.css
-->
<!-- test: file web/css/print.css
-->

```ant
func countFiles(dir: string): Result[int, Error] {
  var n = 0
  for entry in try File.readDir(dir) {
    if entry.isDir {
      n += try countFiles(Path.join([dir, entry.name]))
    } else {
      n += 1
    }
  }
  Result.ok(n)
}

println(countFiles("web"))
```

```text
Ok{ value: 3 }
```

Treat a link to a directory as a directory, asking [File.stat](stat.md) only about the links:

<!-- test: file docs/guide/intro.md
-->
<!-- test: file docs/readme.md
-->

```ant
func isDirectory(dir: string, entry: File.Entry): bool {
  if !entry.isSymlink { return entry.isDir }
  let info = File.stat(Path.join([dir, entry.name]))
  info is Ok && info.value.isDir
}

try File.symlink("guide", "docs/latest")
for entry in try File.readDir("docs") {
  println("${entry.name}: ${isDirectory("docs", entry)}")
}
```

```text
guide: true
latest: true
readme.md: false
```

## Properties and relations

For an entry that is not a link, `isDir` is what [File.stat](stat.md) reports for its path:

<!-- test: file notes/todo.txt
buy milk
-->
<!-- test: file notes/archive/2025.md
# 2025
-->

```ant
for entry in try File.readDir("notes") {
  let info = try File.stat(Path.join(["notes", entry.name]))
  println(entry.isDir == info.isDir)
}
```

```text
true
true
```

For a link, the entry agrees with [File.stat](stat.md) when it is told not to follow:

<!-- test: file refs/sub/inner.txt
-->

```ant
try File.symlink("sub", "refs/to-dir")
let entry = (try File.readDir("refs"))[1]
let info = try File.stat("refs/to-dir", follow: false)
println(entry.name)
println(entry.isSymlink == info.isSymlink)
println(entry.isDir == info.isDir)
```

```text
to-dir
true
true
```

`File.readDir` reports the directories in `dir`, empty ones too, and does not descend.
[File.walk](walk.md) lists only the files below:

<!-- test: file project/readme.md
-->
<!-- test: file project/src/main.ant
-->

```ant
try File.mkdir("project/build")
println((try File.readDir("project")).map({ x.name }))
println(try File.walk("project"))
```

```text
["build", "readme.md", "src"]
["readme.md", "src/main.ant"]
```

## Possible issues

`isDir` is false for a link to a directory, so code that descends on `isDir` alone does not enter
it. That is usually what a walk wants, since a link can point back up the tree:

<!-- test: file tree/data/a.txt
-->

```ant
try File.symlink("data", "tree/alias")
for entry in try File.readDir("tree") {
  println("${entry.name}: isDir ${entry.isDir}")
}
```

```text
alias: isDir false
data: isDir true
```

The names are not paths. Join them to the directory before using them:

<!-- test: file inbox/todo.txt
buy milk
-->

```ant
let entry = (try File.readDir("inbox"))[0]
println(File.read(entry.name) is Err)
println(File.read(Path.join(["inbox", entry.name])))
```

```text
true
Ok{ value: "buy milk\n" }
```

## See also

- [File.Entry](types/Entry.md): what each entry holds
- [File.list](list.md): the names alone
- [File.stat](stat.md): the size, times, and mode of one path, and what a link points to
- [File.walk](walk.md): every file below a directory
- [File.glob](glob.md): the paths that match a pattern
