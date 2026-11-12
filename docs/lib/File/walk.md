<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.walk

```ant signature
func File.walk(dir: string): Result[Array[string], Error]
```

File.walk returns every file below dir, at any depth, as paths relative to dir. It lists files and
not directories, does not follow symbolic links, and returns the paths in lexical order.

<!-- end generated -->

## Details

- The paths are relative to `dir`, not to the working directory. Join each with `dir` to reach the
  file: `Path.join([dir, path])`.
- Hidden files, whose names begin with a dot, are listed, and so are the files inside hidden
  directories.
- An empty directory leaves no trace in the result.
- A symbolic link below `dir` is listed as a file and never descended into, whatever it points at.
- A `dir` that is missing or not a directory is `Err`, and so is a directory below it that cannot be
  read.

The examples on this page run in a directory that holds these files:

```text
proj/.env
proj/README.md
proj/docs/guide.md
proj/src/main.ant
proj/src/util/strings.ant
proj/src/util/strings_test.ant
proj/.git/config
```

## Basic examples

<!-- test: file proj/.env
PORT=8080
-->
<!-- test: file proj/README.md
# App
-->
<!-- test: file proj/docs/guide.md
# Guide

Start here.
-->
<!-- test: file proj/src/main.ant
import "util"

println(util.shout("hi"))
-->
<!-- test: file proj/src/util/strings.ant
func shout(s: string): string {
  s.toUpper()
}
-->
<!-- test: file proj/src/util/strings_test.ant
test "shout" {
  assert(shout("a") == "A")
}
-->
<!-- test: file proj/.git/config
[core]
-->

List every file below a directory, at any depth:

```ant
println(File.walk("proj/src"))
```

```text
Ok{ value: ["main.ant", "util/strings.ant", "util/strings_test.ant"] }
```

Unwrap the paths with `try` and loop over them. In a script, `try` ends the script on an `Err`:

```ant
for path in try File.walk("proj") {
  println(path)
}
```

```text
.env
.git/config
README.md
docs/guide.md
src/main.ant
src/util/strings.ant
src/util/strings_test.ant
```

## Scope

A directory with no files below it gives an empty array, even when it holds empty directories:

```ant
try File.mkdir("blank/a/b")
println(File.walk("blank"))
```

```text
Ok{ value: [] }
```

The order is lexical element by element, so the files of `a` come before those of `a-b`, although
`-` sorts before `/` as a code point:

<!-- test: file order/a/x
-->
<!-- test: file order/a-b/x
-->

```ant
let paths = try File.walk("order")
println(paths)
println(paths.sort())
```

```text
["a/x", "a-b/x"]
["a-b/x", "a/x"]
```

Names are compared code point by code point: uppercase before lowercase, `10` before `9`:

<!-- test: file mixed/apple.txt
-->
<!-- test: file mixed/Zeta.txt
-->
<!-- test: file mixed/9.txt
-->
<!-- test: file mixed/10.txt
-->

```ant
println(try File.walk("mixed"))
```

```text
["10.txt", "9.txt", "Zeta.txt", "apple.txt"]
```

`dir` may be spelled with `./`, `..`, or a final `/`; the paths below it come out the same:

```ant
println(try File.walk("./proj/../proj/docs/"))
```

```text
["guide.md"]
```

Call it in a pipe, with the directory first:

```ant
let count = "proj" | File.walk() | Result.map({ x.length() })
println(count)
```

```text
Ok{ value: 7 }
```

## Applications

Find the test files of a project:

```ant
let tests = try File.walk("proj")
  | Array.filter({ x.endsWith("_test.ant") })
println(tests)
```

```text
["src/util/strings_test.ant"]
```

Count the files of each extension:

```ant
let byExt = (try File.walk("proj")).groupBy(Path.ext)
for entry in byExt.entries() {
  let ext = if entry.key.isEmpty() { "(none)" } else { entry.key }
  println("${ext}: ${entry.value.length()}")
}
```

```text
.env: 1
(none): 1
.md: 2
.ant: 3
```

Add up the size of a tree. Join each path with `dir` to stat it, and wrap a failure with the path:

```ant
func treeSize(dir: string): Result[int, Error] {
  var total = 0
  for path in try File.walk(dir) {
    let full = Path.join([dir, path])
    let info = try (File.stat(full) | Result.wrapErr("size of ${full}"))
    total += info.size
  }
  Result.ok(total)
}

println(treeSize("proj/docs"))
println(treeSize("proj/src"))
```

```text
Ok{ value: 21 }
Ok{ value: 134 }
```

Find the Markdown pages that do not mention a word. `Path.match` skips hidden names at any depth,
and a page that cannot be read ends the script:

```ant
let dir = "proj"
let missing: Array[string] = []
for path in try File.walk(dir) {
  if !Path.match("**/*.md", path).unwrapOr(false) {
    continue
  }
  let text = try File.read(Path.join([dir, path]))
  if !text.contains("Start") {
    missing.push(path)
  }
}
println(missing)
```

```text
["README.md"]
```

## Properties and relations

No path is absolute, and each one joined with `dir` names an entry that exists. `File.stat` with
`follow: false` checks the entry itself, so a dangling link counts too:

```ant
let dir = "proj"
let paths = try File.walk(dir)
assert(paths.every({ !Path.isAbs(x) }))
assert(paths.every({ File.stat(Path.join([dir, x]), follow: false).isOk() }))
```

For a pattern that matches only files, [File.glob](glob.md) gives `File.walk` filtered with
[Path.match](../Path/match.md), in the same order:

```ant
let pattern = "**/*.ant"
let walked = (try File.walk("proj")).filter({ Path.match(pattern, x).unwrapOr(false) })
assert(walked == (try File.glob(pattern, dir: "proj")))
```

[File.list](list.md) names the entries of one directory; `File.walk` replaces each subdirectory with
the files inside it:

```ant
println(try File.list("proj/src"))
println(try File.walk("proj/src"))
```

```text
["main.ant", "util"]
["main.ant", "util/strings.ant", "util/strings_test.ant"]
```

## Possible issues

The paths are relative to `dir`. Reading one as it comes back looks for it in the working directory
instead:

```ant
let paths = try File.walk("proj/docs")
println(File.read(paths[0]))
println(File.read(Path.join(["proj/docs", paths[0]])))
```

```text
Err{ error: Error{ message: "open guide.md: no such file or directory" } }
Ok{ value: "# Guide\n\nStart here.\n" }
```

A missing `dir` and a file in place of a directory are errors, not an empty array:

```ant
println(File.walk("nothing"))
println(File.walk("proj/README.md"))
```

```text
Err{ error: Error{ message: "stat nothing: no such file or directory" } }
Err{ error: Error{ message: "proj/README.md is not a directory" } }
```

Copying a tree file by file with `File.walk` loses its empty directories, since they are not listed:

```ant
try File.mkdir("site/assets")
try File.write("site/index.html", "<h1>Hi</h1>")
for path in try File.walk("site") {
  let to = Path.join(["copy", path])
  try File.mkdir(Path.dir(to))
  try File.copy(Path.join(["site", path]), to)
}
println(File.exists("copy/assets"))
```

```text
false
```

[File.copyAll](copyAll.md) copies the whole tree, empty directories included.

A symbolic link is listed as a file, even when it points at a directory, and the files behind it are
not listed. A dangling link is listed too:

```ant
try File.symlink("src", "proj/linked")
try File.symlink("nowhere", "proj/dangling")
let paths = try File.walk("proj")
println(paths.filter({ !x.contains("/") }))
println(File.stat("proj/linked").map({ x.isDir }))
println(File.exists("proj/dangling"))
```

```text
[".env", "README.md", "dangling", "linked"]
Ok{ value: true }
false
```

## See also

[File.list](list.md) · [File.glob](glob.md) · [File.stat](stat.md) · [File.copyAll](copyAll.md) ·
[Path.join](../Path/join.md) · [Path.match](../Path/match.md)
