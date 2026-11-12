<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.glob

```ant signature
func File.glob(pattern: string, dir: string = ".", hidden: bool = false): Result[Array[string], Error]
```

File.glob returns the files and directories below dir whose paths relative to dir match pattern, as
paths relative to dir, cleaned, each once, in lexical order. The syntax is
[`Path.match`](../Path/match.md)'s: wildcards skip names that begin with a dot unless hidden is
true, \*\* does not descend through symbolic links, and a pattern that ends in a separator matches
directories only. An absolute pattern ignores dir. Nothing matching is Ok with an empty array; a
malformed pattern, and a directory the pattern leads into that cannot be read, is Err.

<!-- end generated -->

## Details

- A directory missing below `dir`, as `missing/*.txt` asks for, is no match, not `Err`.
- `File.glob` checks the pattern before it reads anything, and a `dir` that is not a directory is
  `Err`.
- `File.glob` reads only the directories the pattern leads into: `src/*/main.ant` reads `src` and
  its subdirectories, not the rest of `dir`.
- An element without wildcards is looked up as written, so on a file system that ignores case
  `README.MD` finds `README.md` and comes back as `README.MD`; a wildcard compares code points
  exactly.

The examples on this page run in a directory that holds these files:

```text
notes.txt
todo.txt
.notes.txt
app/README.md
app/.env
app/.git/config
app/docs/guide.md
app/src/main.ant
app/src/util/strings.ant
app/src/util/strings_test.ant
app/tests/main_test.ant
```

## Basic examples

<!-- test: file notes.txt
buy milk
-->
<!-- test: file todo.txt
write docs
-->
<!-- test: file .notes.txt
private
-->
<!-- test: file app/README.md
# App
-->
<!-- test: file app/.env
PORT=8080
-->
<!-- test: file app/.git/config
[core]
-->
<!-- test: file app/docs/guide.md
# Guide
-->
<!-- test: file app/src/main.ant
import "util"

println(util.shout("hi"))
-->
<!-- test: file app/src/util/strings.ant
func shout(s: string): string {
  s.toUpper()
}
-->
<!-- test: file app/src/util/strings_test.ant
test "shout" {
  assert(shout("a") == "A")
}
-->
<!-- test: file app/tests/main_test.ant
test "nothing" {
  assert(true)
}
-->

Find the text files in the working directory. `*` does not match the hidden `.notes.txt`:

```ant
let paths = try File.glob("*.txt")
println(paths)
```

```text
["notes.txt", "todo.txt"]
```

Search below another directory with `dir:`. `**` matches any number of directories, none included:

```ant
let sources = try File.glob("**/*.ant", dir: "app")
for path in sources {
  println(path)
}
```

```text
src/main.ant
src/util/strings.ant
src/util/strings_test.ant
tests/main_test.ant
```

Nothing matching is an empty array, not an error:

```ant
let found = try File.glob("*.go", dir: "app")
println(found.isEmpty())
```

```text
true
```

## Scope

`*` matches any run of code points within one name, `?` one code point, and `[...]` one code point
of a set or a range; `[^...]` and `[!...]` match one code point not in the set:

```ant
println(try File.glob("?odo.txt"))
println(try File.glob("[nt]*.txt"))
println(try File.glob("[^n]*.txt"))
println(try File.glob("[!n]*.txt"))
```

```text
["todo.txt"]
["notes.txt", "todo.txt"]
["todo.txt"]
["todo.txt"]
```

A `\` escapes a wildcard on Unix; on Windows write `[*]`, `[?]`, `[[]`.

A pattern with several elements matches one name per element. A wildcard in a directory element
reads each directory it matches:

```ant
println(try File.glob("*/*.ant", dir: "app"))
println(try File.glob("src/*/*_test.ant", dir: "app"))
```

```text
["src/main.ant", "tests/main_test.ant"]
["src/util/strings_test.ant"]
```

`**` as the last element matches everything below a directory, its directories included, but not the
directory itself:

```ant
let below = try File.glob("src/**", dir: "app")
for path in below {
  println(path)
}
```

```text
src/main.ant
src/util
src/util/strings.ant
src/util/strings_test.ant
```

A pattern that ends in `/` matches directories only, and each path keeps the `/`:

```ant
println(try File.glob("*/", dir: "app"))
println(try File.glob("src/*/", dir: "app"))
```

```text
["docs/", "src/", "tests/"]
["src/util/"]
```

A pattern without wildcards checks whether a path is there:

```ant
println(try File.glob("README.md", dir: "app"))
println(try File.glob("CHANGES.md", dir: "app"))
```

```text
["README.md"]
[]
```

The paths come back cleaned. An element `..` steps back on the text:

```ant
println(try File.glob("./src/../*.md", dir: "app"))
```

```text
["README.md"]
```

An absolute pattern ignores `dir` and gives absolute paths:

```ant
let src = try Path.abs("app/src")
let paths = try File.glob(Path.join([src, "*.ant"]), dir: "app/docs")
println(paths.map({ x.startsWith(src) }))
println(paths.map(Path.base))
```

```text
[true]
["main.ant"]
```

Chain on the result to turn the paths into a list of names:

```ant
let names = (try File.glob("*.txt")) | Array.map(Path.base) | String.join(", ")
println(names)
```

```text
notes.txt, todo.txt
```

## Named arguments

`dir` is the directory to search, `"."` when left out. Its name is not part of the pattern, so a
directory whose name holds `*` or `[` needs no escaping:

<!-- test: file drafts [old]/plan.txt
first plan
-->

```ant
println(try File.glob("*.txt", dir: "drafts [old]"))
```

```text
["plan.txt"]
```

`hidden: true` lets wildcards match names that begin with a dot, and lets `**` descend into hidden
directories:

```ant
println(try File.glob("*", dir: "app"))
println(try File.glob("*", dir: "app", hidden: true))
println(try File.glob("**/config", dir: "app", hidden: true))
```

```text
["README.md", "docs", "src", "tests"]
[".env", ".git", "README.md", "docs", "src", "tests"]
[".git/config"]
```

Without `hidden`, a pattern element that begins with a literal `.` still matches hidden names:

```ant
println(try File.glob(".*", dir: "app"))
```

```text
[".env", ".git"]
```

## Applications

Count the lines of every source file below a directory. Join each path with `dir` to read it:

```ant
let dir = "app"
var total = 0
for path in try File.glob("**/*.ant", dir: dir) {
  let text = try File.read(Path.join([dir, path]))
  total += text.lines().length()
}
println("${total} lines")
```

```text
12 lines
```

Group test files by the directory they are in:

```ant
let tests = try File.glob("**/*_test.ant", dir: "app")
for entry in tests.groupBy(Path.dir).entries() {
  println("${entry.key}: ${entry.value.length()}")
}
```

```text
src/util: 1
tests: 1
```

Find the top-level directories that lack a `README.md`:

```ant
let dir = "app"
let undocumented = try File.glob("*/", dir: dir)
  | Array.filter({ !File.exists(Path.join([dir, x, "README.md"])) })
println(undocumented)
```

```text
["docs/", "src/", "tests/"]
```

Remove leftover temporary files, and report the first failure with its context:

<!-- test: file build/a.tmp
-->
<!-- test: file build/out/b.tmp
-->
<!-- test: file build/out/app.js
-->

```ant
func clean(dir: string): Result[int, Error] {
  let leftovers = try (File.glob("**/*.tmp", dir: dir) | Result.wrapErr("clean ${dir}"))
  for path in leftovers {
    try (File.remove(Path.join([dir, path])) | Result.wrapErr("clean ${dir}"))
  }
  Result.ok(leftovers.length())
}

println("removed ${try clean("build")}")
println(try File.glob("**", dir: "build"))
```

```text
removed 2
["out", "out/app.js"]
```

## Properties and relations

Every path `File.glob` returns matches the pattern under [Path.match](../Path/match.md):

```ant
let pattern = "**/*.ant"
let paths = try File.glob(pattern, dir: "app")
assert(paths.every({ Path.match(pattern, x).unwrapOr(false) }))
```

For a pattern that matches only files, the result is [File.walk](walk.md) filtered with
`Path.match`, in the same order:

```ant
let pattern = "**/*_test.ant"
let walked = (try File.walk("app")).filter({ Path.match(pattern, x).unwrapOr(false) })
assert(walked == (try File.glob(pattern, dir: "app")))
```

## Possible issues

The paths are relative to `dir`. Reading one as it comes back looks for it in the working directory
instead:

```ant
let paths = try File.glob("*.md", dir: "app")
println(File.read(paths[0]))
println(File.read(Path.join(["app", paths[0]])))
```

```text
Err{ error: Error{ message: "open README.md: no such file or directory" } }
Ok{ value: "# App\n" }
```

Putting the directory in the pattern instead keeps it in every path:

```ant
println(try File.glob("app/*.md"))
```

```text
["app/README.md"]
```

`**` does not descend into a hidden directory, such as `.git`, without `hidden: true`, so the
`config` in `app/.git` is not found:

```ant
println(try File.glob("**/config", dir: "app"))
```

```text
[]
```

Wildcards compare code points exactly, on every system, so `*.TXT` does not match `notes.txt`:

```ant
println(try File.glob("*.TXT"))
```

```text
[]
```

A malformed pattern is an `Err`, not an empty array:

```ant
println(File.glob("[a-"))
```

```text
Err{ error: Error{ message: "malformed pattern [a-: [ without ]" } }
```

The order is lexical element by element, so `a/x` comes before `a-b/x`. Sorting the paths as strings
gives another order, because `-` sorts before `/`:

<!-- test: file order/a/x
-->
<!-- test: file order/a-b/x
-->

```ant
let paths = try File.glob("*/x", dir: "order")
println(paths)
println(paths.sort())
```

```text
["a/x", "a-b/x"]
["a-b/x", "a/x"]
```

`**` does not descend through a symbolic link, but a wildcard element goes through a link to a
directory:

```ant
try File.symlink("src", "app/linked")
println(try File.glob("**/main.ant", dir: "app"))
println(try File.glob("*/main.ant", dir: "app"))
```

```text
["src/main.ant"]
["linked/main.ant", "src/main.ant"]
```

## See also

[Path.match](../Path/match.md) · [File.walk](walk.md) · [File.list](list.md) ·
[File.exists](exists.md) · [Path.join](../Path/join.md) · [File.stat](stat.md)
