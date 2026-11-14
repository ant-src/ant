<!-- status: reviewed -->

# Build and split paths

Put a path together from its parts and take one apart, with the `Path` functions, not by splitting
on the separator. They work on the text in the host's convention, `/` on Unix and `\` on Windows,
and do not look at the file system. The examples show Unix paths.

1. Join the parts with [Path.join](../../lib/Path/join.md).
2. Take a path apart with [Path.dir](../../lib/Path/dir.md), [Path.base](../../lib/Path/base.md),
   and [Path.ext](../../lib/Path/ext.md).
3. Change the extension with `Path.ext` and [String.trimSuffix](../../lib/String/trimSuffix.md).
4. Tidy a path from input with [Path.clean](../../lib/Path/clean.md).
5. Turn a path into one from another directory with [Path.rel](../../lib/Path/rel.md), or into an
   absolute one with [Path.abs](../../lib/Path/abs.md).
6. Walk up the parents with `Path.dir`.

## Join the parts

Pass the parts to `Path.join` as an array. It puts the separator between them and cleans the result,
so extra separators, `.` elements, and empty parts disappear, and a `..` removes the element before
it:

```ant
Path.join(["site", "posts", "hello.md"])      // "site/posts/hello.md"
Path.join(["out/", "./img", "", "logo.png"])  // "out/img/logo.png"
Path.join(["site", "../shared", "style.css"]) // "shared/style.css"
```

An absolute part does not replace the parts before it, and an empty first part does not make the
path absolute. To start at the root, pass `"/"` as the first part:

```ant
Path.join(["out", "/img"]) // "out/img"
Path.join(["", "etc"])     // "etc"
Path.join(["/", "etc"])    // "/etc"
```

## Split a path

`Path.dir` gives everything up to the last element, `Path.base` the last element, and `Path.ext` the
extension of the last element, dot included:

```ant
let page = "site/posts/2026/hello.md"
Path.dir(page)  // "site/posts/2026"
Path.base(page) // "hello.md"
Path.ext(page)  // ".md"
```

Only the last dot counts, and a name that starts with a dot is all extension:

```ant
Path.ext("archive.tar.gz") // ".gz"
Path.ext(".env")           // ".env"
Path.ext("Makefile")       // ""
```

A trailing separator is dropped first, so for a directory path `dir` and `base` overlap:

```ant
Path.base("site/posts/") // "posts"
Path.dir("site/posts/")  // "site/posts"
```

## Change the extension

Take the extension off the name with [String.trimSuffix](../../lib/String/trimSuffix.md), then join
the new name to the directory:

```ant
let page = "site/posts/2026/hello.md"
let name = Path.base(page).trimSuffix(Path.ext(page))
Path.join([Path.dir(page), name + ".html"]) // "site/posts/2026/hello.html"
```

## Clean a path from input

A path a user types may have doubled separators, `.` and `..` elements, or a trailing separator.
`Path.clean` gives the shortest equivalent path:

```ant
Path.clean("posts//2026/./../2025/") // "posts/2025"
Path.clean("./notes.txt")            // "notes.txt"
Path.clean("")                       // "."
Path.clean("posts/../../etc")        // "../etc"
```

A leading `..` stays, so cleaning does not keep a path inside a directory. Check that with
`Path.rel`, as the whole program does.

## Make a path relative or absolute

`Path.rel` gives the path that leads from one directory to a target:

```ant
Path.rel("site", "site/posts/hello.md")     // Ok{ value: "posts/hello.md" }
Path.rel("site/posts", "site/img/logo.png") // Ok{ value: "../img/logo.png" }
```

It is an `Err` when no relative path leads there, as from an absolute directory to a relative
target:

```ant
Path.rel("/srv", "a") // Err{ error: Error{ message: "no relative path leads from /srv to a" } }
```

`Path.abs` joins a relative path to the working directory and cleans it.
[Path.isAbs](../../lib/Path/isAbs.md) tells whether a path is absolute already:

```ant
let notes = try Path.abs("posts/../notes.txt")
let cwd = try Os.cwd()
println(Path.isAbs(notes))
println(notes == Path.join([cwd, "notes.txt"]))
```

```text
true
true
```

## Walk up the parent directories

Call `Path.dir` until it returns its argument. A relative path ends at `.`, an absolute one at `/`:

```ant
var dir = "site/posts/2026"
while dir != Path.dir(dir) {
  println(dir)
  dir = Path.dir(dir)
}
println(dir)
```

```text
site/posts/2026
site/posts
site
.
```

## The whole program

This script maps source files under one directory to HTML files at the same place under another, as
a static site generator does. It takes the two directories and then the source files as arguments. A
source outside the source directory is an error, found from the first element of its relative path:

<!-- test: args content/ ./public content/index.md content//blog/first.post.md ./content/about/team.md -->

```ant
func firstElement(rel: string): string {
  var head = rel
  while Path.dir(head) != "." {
    head = Path.dir(head)
  }
  head
}

func outputPath(source: string, src: string, out: string): Result[string, Error] {
  let rel = try Path.rel(src, source)
  if firstElement(rel) == ".." {
    return Result.fail("${source} is outside ${src}")
  }
  let name = Path.base(rel).trimSuffix(Path.ext(rel))
  Result.ok(Path.join([out, Path.dir(rel), name + ".html"]))
}

let args = Os.args()
if args.length() < 2 {
  return Result.fail("usage: paths <src> <out> <file>...")
}
let src = args[0]
let out = args[1]
for source in args.slice(2, args.length()) {
  println("${source} -> ${try outputPath(source, src, out)}")
}
```

```text
content/index.md -> public/index.html
content//blog/first.post.md -> public/blog/first.post.html
./content/about/team.md -> public/about/team.html
```

Given `../notes/x.md`, the script stops with `../notes/x.md is outside content/` and exits with
status 1, as it does with fewer than two arguments. Besides the `Path` functions above, it uses
[Os.args](../../lib/Os/args.md), [Array.slice](../../lib/Array/slice.md), and
[Result.fail](../../lib/Result/fail.md).

## See also

[Path.join](../../lib/Path/join.md) · [Path.dir](../../lib/Path/dir.md) ·
[Path.base](../../lib/Path/base.md) · [Path.ext](../../lib/Path/ext.md) ·
[Path.clean](../../lib/Path/clean.md) · [Path.rel](../../lib/Path/rel.md) ·
[Path.abs](../../lib/Path/abs.md) · [Path.isAbs](../../lib/Path/isAbs.md) ·
[Path.match](../../lib/Path/match.md) · [Os.cwd](../../lib/Os/cwd.md) ·
[String.trimSuffix](../../lib/String/trimSuffix.md) ·
[Walk a directory and match globs](walk-glob.md)
