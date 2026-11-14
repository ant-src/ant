<!-- status: reviewed -->

# Walk a directory and match globs

The examples on this page run in a directory that holds a project:

```text
app/.env
app/.git/config
app/README.md
app/docs/guide.md
app/src/main.ant
app/src/util/strings.ant
app/src/util/strings_test.ant
app/tests/main_test.ant
```

## Find files by pattern

This script lists every Ant file below `app` with its size:

<!-- test: file app/.env
PORT=8080
-->
<!-- test: file app/.git/config
[core]
-->
<!-- test: file app/README.md
# App
-->
<!-- test: file app/docs/guide.md
# Guide
-->
<!-- test: file app/src/main.ant
println("hi")
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
test "main" {
  assert(true)
}
-->

```ant
let root = "app"
let sources = try File.glob("**/*.ant", dir: root)
for path in sources {
  let info = try File.stat(Path.join([root, path]))
  println("${path}  ${info.size} bytes")
}
```

```text
src/main.ant  14 bytes
src/util/strings.ant  48 bytes
src/util/strings_test.ant  45 bytes
tests/main_test.ant  31 bytes
```

1. Pass the pattern and the directory to search to [File.glob](../../lib/File/glob.md). `**` matches
   zero or more directories, and `*` any run of characters within one name.
2. Join the directory and each path with [Path.join](../../lib/Path/join.md). `File.glob` returns
   paths relative to `dir:`, in lexical order.
3. Use the joined path, here with [File.stat](../../lib/File/stat.md).

The pattern syntax is that of [Path.match](../../lib/Path/match.md). Wildcards skip names that begin
with a dot, so the pattern never descends into `.git`. Nothing matching gives an empty array, not an
error. A pattern has no `{a,b}` alternation: `{` matches itself, so `*.{md,txt}` matches only a name
that ends in `.{md,txt}`. Matching is case-sensitive on every system, so `*.TXT` does not match
`a.txt`.

## Walk every file

[File.walk](../../lib/File/walk.md) returns every file below a directory, hidden ones included, as
paths relative to it. It lists files, not directories:

```ant
for path in try File.walk("app") {
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
tests/main_test.ant
```

## Filter walked paths with a pattern

To match among paths you already have, or to include hidden files, filter `File.walk` with
[Path.match](../../lib/Path/match.md). Match each path against a pattern with `Path.match`. It
returns `Err` only for a malformed pattern, so a fixed pattern can unwrap it:

```ant
let tests = try File.walk("app")
  | Array.filter({ Path.match("**/*_test.ant", x).unwrap() })
println(tests)
```

```text
["src/util/strings_test.ant", "tests/main_test.ant"]
```

To match the file name wherever the file is, match against [Path.base](../../lib/Path/base.md) of
the path:

```ant
let docs = try File.walk("app")
  | Array.filter({ Path.match("*.md", Path.base(x)).unwrap() })
println(docs)
```

```text
["README.md", "docs/guide.md"]
```

To match several extensions, keep the patterns in an array and ask whether `some` of them matches.
The outer callback is a function literal, so the inner lambda's `x` is the pattern:

```ant
let patterns = ["*.md", "*.ant"]
let found = try File.walk("app")
  | Array.filter(func(path: string): bool {
    patterns.some({ Path.match(x, Path.base(path)).unwrap() })
  })
println(found)
```

```text
["README.md", "docs/guide.md", "src/main.ant", "src/util/strings.ant", "src/util/strings_test.ant", "tests/main_test.ant"]
```

Walked paths include hidden ones, and wildcards skip them; pass `hidden: true` to match them too:

```ant
let configs = try File.walk("app")
  | Array.filter({ Path.match("**/config", x, hidden: true).unwrap() })
println(configs)
```

```text
[".git/config"]
```

## Include hidden files

Pass `hidden: true` to let wildcards match names that begin with a dot. `**` alone matches
everything below the directory, directories included:

```ant
for path in try File.glob("**", dir: "app", hidden: true) {
  println(path)
}
```

```text
.env
.git
.git/config
README.md
docs
docs/guide.md
src
src/main.ant
src/util
src/util/strings.ant
src/util/strings_test.ant
tests
tests/main_test.ant
```

## Find directories

End the pattern with `/` to match directories only. Each path keeps the `/`:

```ant
println(try File.glob("**/", dir: "app"))
```

```text
["docs/", "src/", "src/util/", "tests/"]
```

## Report a bad pattern

`File.glob` returns `Err` for a malformed pattern and for a `dir:` that is not a directory. Add the
directory to the error with [Result.wrapErr](../../lib/Result/wrapErr.md):

<!-- test: exit 1 -->

```ant
let root = "app"
let found = try (File.glob("[a-", dir: root) | Result.wrapErr("search ${root}"))
println(found)
```

```sh
$ ant run find.ant; echo $?
search app: malformed pattern [a-: [ without ]
1
```

A directory missing below `dir:`, as `missing/*.txt` asks for, is no match, not an error.

## See also

[File.glob](../../lib/File/glob.md) · [File.walk](../../lib/File/walk.md) ·
[File.list](../../lib/File/list.md) · [File.stat](../../lib/File/stat.md) ·
[Path.match](../../lib/Path/match.md) · [Path.join](../../lib/Path/join.md) ·
[Path.base](../../lib/Path/base.md) · [Result.wrapErr](../../lib/Result/wrapErr.md) ·
[Build and split paths](paths.md)
