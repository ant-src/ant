<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path

Path holds pure functions over paths in the host's convention: / on Unix, \\ and drive letters on
Windows. They work on the text and never touch the file system, except [`Path.abs`](abs.md).

<!-- end generated -->

## Details

- Every function works on the text alone: the path does not have to exist, and nothing on disk is
  read or changed. [Path.abs](abs.md) also reads the working directory. [File](../File/index.md)
  touches the file system.
- The examples on this page use `/`.
- [Path.join](join.md) and [Path.clean](clean.md) give a cleaned path, and the empty path is `"."`
  for both. [Path.match](match.md) does not clean `path`.

## Basic examples

Build a path from parts with `Path.join`:

```ant
Path.join(["src", "lib", "main.ant"]) // "src/lib/main.ant"
Path.join(["a", "b/../c", "d.ant"])   // "a/c/d.ant"
```

Take a path apart with `Path.dir`, `Path.base`, and `Path.ext`:

```ant
let p = "src/lib/main.ant"
Path.dir(p)  // "src/lib"
Path.base(p) // "main.ant"
Path.ext(p)  // ".ant"
```

Ask how one path leads to another with `Path.rel`, and whether a path matches a glob pattern with
`Path.match`:

```ant
Path.rel("/srv/app", "/srv/app/lib/db.ant") // Ok{ value: "lib/db.ant" }
Path.match("**/*.ant", "src/main.ant")      // Ok{ value: true }
```

## Scope

`Path.clean` gives the shortest equivalent path, and `Path.isAbs` tells an absolute path from a
relative one:

```ant
Path.clean("a//b/./c/..") // "a/b"
Path.clean("")            // "."
Path.isAbs("/etc/hosts")  // true
Path.isAbs("etc/hosts")   // false
```

No parts, or only empty ones, join to `"."`:

```ant
Path.join([])       // "."
Path.join(["", ""]) // "."
```

A wildcard does not match a name that begins with `.`, unless `hidden: true` is given:

```ant
Path.match("*", ".env")               // Ok{ value: false }
Path.match("*", ".env", hidden: true) // Ok{ value: true }
```

`**/*.ant` also matches a file at the top, and `src/**` matches everything below `src` but not `src`
itself. Wildcards compare code points exactly, so case counts:

```ant
Path.match("**/*.ant", "a.ant") // Ok{ value: true }
Path.match("src/**", "src")     // Ok{ value: false }
Path.match("*.TXT", "a.txt")    // Ok{ value: false }
```

A malformed pattern is `Err`, not `false`:

```ant
Path.match("[a-", "x") // Err{ error: Error{ message: "malformed pattern [a-: [ without ]" } }
```

`Path.rel` climbs out of `base` with `..` when it has to. It is `Err` when no relative path leads
from `base` to `target`: when one is absolute and the other is not, or when `base` climbs out with
`..` further than `target` does:

```ant
Path.rel("src/lib", "src/test/a.ant") // Ok{ value: "../test/a.ant" }
Path.rel("/srv", "app")               // Err{ error: Error{ message: "no relative path leads from /srv to app" } }
Path.rel("../a", "b")                 // Err{ error: Error{ message: "no relative path leads from ../a to b" } }
```

`Path.abs` resolves a relative path against the working directory. It still looks only at the text:
the path does not have to exist, and links stay as written ([File.realPath](../File/realPath.md)
resolves them):

```ant
let cwd = try Os.cwd()
let notes = try Path.abs("notes.txt")
println(Path.isAbs(notes))
println(notes == Path.join([cwd, "notes.txt"]))
```

```text
true
true
```

## Functions by job

### Building and cleaning

- [Path.join](join.md): join parts with the separator, cleaned
- [Path.clean](clean.md): the shortest equivalent path

### Taking a path apart

- [Path.dir](dir.md): all but the last element
- [Path.base](base.md): the last element
- [Path.ext](ext.md): the extension of the last element

### Absolute and relative

- [Path.isAbs](isAbs.md): whether a path is absolute
- [Path.abs](abs.md): a path made absolute against the working directory
- [Path.rel](rel.md): a target as a path from a base

### Matching

- [Path.match](match.md): whether a path matches a glob pattern, the syntax
  [File.glob](../File/glob.md) takes

## Applications

Put each source file's output beside the others in one directory, with a new extension:

```ant
func outputPath(src: string, outDir: string): string {
  let name = Path.base(src)
  Path.join([outDir, "${name.trimSuffix(Path.ext(name))}.html"])
}

for src in ["docs/index.md", "docs/guide/intro.md"] {
  println(outputPath(src, "site"))
}
```

```text
site/index.html
site/intro.html
```

Pick the test files out of a tree. [File.walk](../File/walk.md) gives paths relative to the
directory, so `Path.join` puts the directory back in front:

```ant
try File.mkdir("project/src/util")
try File.write("project/src/main.ant", "")
try File.write("project/src/main_test.ant", "")
try File.write("project/src/util/strings_test.ant", "")
for path in try File.walk("project") {
  if try Path.match("**/*_test.ant", path) {
    println(Path.join(["project", path]))
  }
}
```

```text
project/src/main_test.ant
project/src/util/strings_test.ant
```

Show a path relative to a project root:

```ant
let root = try Path.abs("project")
let file = Path.join([root, "src", "main.ant"])
println(try Path.rel(root, file))
```

```text
src/main.ant
```

## Properties and relations

A clean path is its directory joined with its base, joining cleans, and joining a base with a path
relative to it gives the path back:

```ant
let p = "src/lib/main.ant"
assert(Path.join([Path.dir(p), Path.base(p)]) == p)
assert(Path.join(["a//b", "./c"]) == Path.clean("a//b/./c"))
assert(Path.join([Path.dir(p), try Path.rel(Path.dir(p), p)]) == p)
```

## Possible issues

A part that starts with a separator does not start over at the root. `Path.join` keeps every part:

```ant
Path.join(["/srv", "/etc/passwd"]) // "/srv/etc/passwd"
```

`Path.abs` does not check that the path exists or resolve links; `File.realPath` does, so it is
`Err` for a path that leads nowhere:

```ant
try File.write("notes.txt", "")
println(File.realPath("notes.txt").isOk())
println(Path.abs("missing.txt").isOk())
println(File.realPath("missing.txt").isOk())
```

```text
true
true
false
```

`Path.match` does not clean `path`, so a leading `./` keeps a pattern from matching. Clean the path
first:

```ant
Path.match("*.txt", "./a.txt")             // Ok{ value: false }
Path.match("*.txt", Path.clean("./a.txt")) // Ok{ value: true }
```

## See also

[File](../File/index.md) · [File.glob](../File/glob.md) · [File.walk](../File/walk.md) ·
[File.realPath](../File/realPath.md) · [Os.cwd](../Os/cwd.md) · [Os.home](../Os/home.md) ·
[Build and split paths](../../howto/files/paths.md) ·
[Walk a directory and match globs](../../howto/files/walk-glob.md)
