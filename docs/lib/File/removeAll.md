<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.removeAll

```ant signature
func File.removeAll(path: string): Result[void, Error]
```

File.removeAll removes path and everything below it. It is Ok when nothing is there and the path
could hold something, and Err for . and ..; [`File.remove`](remove.md) is the form that refuses a
directory with entries.

<!-- end generated -->

## Details

- At a symbolic link it removes the link and never what the link points at.

## Basic examples

Remove a directory with everything in it:

<!-- test: file build/assets/icons/logo.svg
<svg/>
-->

```ant
println(File.removeAll("build"))
println(File.exists("build"))
```

```text
Ok{ value: void }
false
```

Call it again, or on a path that was never there, and it is still `Ok`:

```ant
println(File.removeAll("build"))
println(File.removeAll("never-made"))
```

```text
Ok{ value: void }
Ok{ value: void }
```

## Scope

A single file is a tree of one, and goes like any other:

<!-- test: file notes.txt
buy milk
-->

```ant
println(File.removeAll("notes.txt"))
println(File.exists("notes.txt"))
```

```text
Ok{ value: void }
false
```

A symbolic link to a directory is removed as a link. The directory it points at keeps its files:

<!-- test: file data/sub/a.txt
a
-->

```ant
try File.symlink("data", "current")
println(File.removeAll("current"))
println(File.exists("current"))
println(try File.walk("data"))
```

```text
Ok{ value: void }
false
["sub/a.txt"]
```

A dangling link, one whose target is missing, is still something to remove:

```ant
try File.symlink("missing", "dangling")
println(File.removeAll("dangling"))
println(File.stat("dangling", follow: false).isOk())
```

```text
Ok{ value: void }
false
```

## Applications

Work in a [temporary directory](tempDir.md) and remove it, with all it holds, when the function
ends. Nothing else removes it:

```ant
func countWords(text: string): Result[int, Error] {
  let dir = try File.tempDir("words-*")
  defer { File.removeAll(dir) }
  let path = Path.join([dir, "input.txt"])
  try File.write(path, text)
  let words = (try File.read(path)).fields()
  Result.ok(words.length())
}

println(countWords("the quick brown fox"))
```

```text
Ok{ value: 4 }
```

Start a build from a clean output directory, whether or not an earlier run left one:

<!-- test: file out/stale.html
old
-->

```ant
func build(out: string, pages: Array[string]): Result[void, Error] {
  try File.removeAll(out)
  try File.mkdir(out)
  for page in pages {
    try File.write(Path.join([out, "${page}.html"]), "<h1>${page}</h1>")
  }
  Result.ok(void)
}

try build("out", ["index", "about"])
println(try File.list("out"))
```

```text
["about.html", "index.html"]
```

## Properties and relations

After an `Ok`, nothing is at the path, and a second call changes nothing:

<!-- test: file cache/a/b/c.bin
data
-->

```ant
assert(File.removeAll("cache").isOk())
assert(!File.exists("cache"))
assert(!File.stat("cache", follow: false).isOk())
assert(File.removeAll("cache").isOk())
```

On a file or an empty directory, `File.removeAll` does what [File.remove](remove.md) does:

```ant
try File.mkdir("empty")
try File.write("one.txt", "1")
assert(File.removeAll("empty").isOk())
assert(File.removeAll("one.txt").isOk())
assert(!File.exists("empty") && !File.exists("one.txt"))
```

## Possible issues

[File.remove](remove.md) refuses a directory that has entries. Use `File.removeAll` when you mean
the whole tree:

<!-- test: file logs/2026/today.log
started
-->

```ant
println(File.remove("logs"))
println(File.removeAll("logs"))
```

```text
Err{ error: Error{ message: "remove logs: directory not empty" } }
Ok{ value: void }
```

Checking [File.exists](exists.md) first adds nothing: a missing path is already `Ok`. Worse, a
dangling link is `false` to `File.exists`, so the check skips a link that `File.removeAll` would
remove:

```ant
try File.symlink("gone", "latest")
if File.exists("latest") {
  try File.removeAll("latest")
}
println(File.stat("latest", follow: false).isOk())
try File.removeAll("latest")
println(File.stat("latest", follow: false).isOk())
```

```text
true
false
```

[`Os.exit`](../Os/exit.md) runs no deferred block, so a function that defers `File.removeAll(dir)`
and then calls `Os.exit` leaves the directory behind. Remove it yourself before exiting:

```ant
let dir = try File.tempDir("job-*")
try File.write(Path.join([dir, "out.txt"]), "done")
try File.removeAll(dir)
println(File.exists(dir))
Os.exit(0)
```

```text
false
```

## See also

[File.remove](remove.md) · [File.tempDir](tempDir.md) · [File.mkdir](mkdir.md) ·
[File.exists](exists.md) · [File.walk](walk.md) · [File.symlink](symlink.md) ·
[File.copyAll](copyAll.md)
