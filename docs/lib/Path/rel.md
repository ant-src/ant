<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.rel

```ant signature
func Path.rel(base: string, target: string): Result[string, Error]
```

Path.rel returns target as a path relative to base, working on the cleaned text. It is Err when no
relative path leads there: when one path is absolute and the other is not, when they are on
different Windows volumes, or when base climbs out with .. further than target does.

<!-- end generated -->

## Details

- `Path.rel` looks only at the text. It does not touch the file system, so it neither needs the
  paths to exist nor follows symbolic links.
- Both paths are cleaned first, so `.`, `..`, doubled separators, and a trailing separator make no
  difference to the result. The result is `.` when the two paths name the same place.

## Basic examples

```ant
Path.rel("/a/b", "/a/c/d")                  // Ok{ value: "../c/d" }
Path.rel("/srv/app", "/srv/app/lib/db.ant") // Ok{ value: "lib/db.ant" }
```

It is `Err` when no relative path leads from `base` to `target`:

```ant
Path.rel("/a", "b") // Err{ error: Error{ message: "no relative path leads from /a to b" } }
```

## Scope

### Down, up, and the same place

A target below `base` gives a path down, one above it a path of `..` elements, and the same path
`.`:

```ant
Path.rel("/home/ann", "/home/ann/notes/todo.txt") // Ok{ value: "notes/todo.txt" }
Path.rel("/home/ann/notes", "/home")              // Ok{ value: "../.." }
Path.rel("/home/ann", "/home/ann")                // Ok{ value: "." }
Path.rel("/", "/usr/lib")                         // Ok{ value: "usr/lib" }
```

### Relative paths

Two relative paths work the same way, as if both started in one directory. An empty path is `.`:

```ant
Path.rel("src", "src/lib/db.ant") // Ok{ value: "lib/db.ant" }
Path.rel("src/lib", "test")       // Ok{ value: "../../test" }
Path.rel("", "docs/index.md")     // Ok{ value: "docs/index.md" }
Path.rel(".", "docs")             // Ok{ value: "docs" }
```

### Paths that are not clean

`Path.rel` cleans both paths before it compares them:

```ant
Path.rel("a/./b/", "a//b/../c")   // Ok{ value: "../c" }
Path.rel("/a/b", "/a/b/../../..") // Ok{ value: "../.." }
Path.rel("docs", "docs/")         // Ok{ value: "." }
```

### `..` at the start

`base` may start with `..` when `target` climbs at least as far:

```ant
Path.rel("../a", "../b") // Ok{ value: "../b" }
Path.rel("../a", "b")    // Err{ error: Error{ message: "no relative path leads from ../a to b" } }
```

From `../a`, the way to `b` goes through the name of the current directory, which the text does not
hold.

### Unicode

Elements are compared as text, whatever characters they hold:

```ant
Path.rel("/música/jazz", "/música/fado/amália.mp3") // Ok{ value: "../fado/amália.mp3" }
```

## Named arguments

The two parameters can be named, which makes the direction plain at the call:

```ant
Path.rel(base: "/srv/app", target: "/srv/app/main.ant") // Ok{ value: "main.ant" }
```

## Applications

### Show paths relative to a root

Print a list of absolute paths relative to the project root. `try` stops at the first path that lies
outside the reach of a relative path:

```ant
func relativeTo(root: string, paths: Array[string]): Result[Array[string], Error] {
  let out: Array[string] = []
  for p in paths {
    out.push(try Path.rel(root, p))
  }
  Result.ok(out)
}

let files = ["/srv/app/main.ant", "/srv/app/lib/db.ant", "/srv/data/users.csv"]
println(relativeTo("/srv/app", files))
```

```text
Ok{ value: ["main.ant", "lib/db.ant", "../data/users.csv"] }
```

### A link from one page to another

A link in a page is relative to the directory that holds the page, so take the path from
[Path.dir](dir.md) of the page:

```ant
func link(from: string, to: string): Result[string, Error] {
  Path.rel(Path.dir(from), to)
}

println(link("docs/guide/intro.md", "docs/lib/Path/rel.md"))
println(link("docs/index.md", "docs/guide/intro.md"))
```

```text
Ok{ value: "../lib/Path/rel.md" }
Ok{ value: "guide/intro.md" }
```

### A relative and an absolute path

Make both paths absolute with [Path.abs](abs.md) before comparing a relative path with an absolute
one. Here the root `reports` is relative to the working directory, and the file is an absolute path
under that same directory:

```ant
func under(root: string, p: string): Result[string, Error] {
  let r = try Path.abs(root)
  let t = try Path.abs(p)
  Path.rel(r, t)
}

let here = Result.unwrap(Path.abs("."))
println(under("reports", Path.join([here, "reports", "2026", "q1.csv"])))
```

```text
Ok{ value: "2026/q1.csv" }
```

## Properties and relations

Joining `base` with the result gives `target`, cleaned:

```ant
let base = "/srv/app/lib"
let target = "/srv/data/./users.csv"
let rel = Result.unwrap(Path.rel(base, target))
assert(Path.join([base, rel]) == Path.clean(target))
```

A path relative to itself is `.`, and the way back up has one `..` for each element of the way down:

```ant
assert(Result.unwrap(Path.rel("/a/b", "/a/b")) == ".")
let down = Result.unwrap(Path.rel("/a/b", "/a/b/c/d"))
let up = Result.unwrap(Path.rel("/a/b/c/d", "/a/b"))
assert(down == "c/d")
assert(up == "../..")
```

## Possible issues

### The base comes first

`Path.rel(base, target)` gives the way from `base` to `target`. Swapping them gives the way back:

```ant
Path.rel("/srv/app", "/srv/app/lib") // Ok{ value: "lib" }
Path.rel("/srv/app/lib", "/srv/app") // Ok{ value: ".." }
```

A pipe passes its value as the first argument, which is `base`:

```ant
"/srv/app/lib" | Path.rel("/srv/app") // Ok{ value: ".." }
```

### An absolute and a relative path

`Path.rel` does not resolve a relative path against the working directory, so mixing the two kinds
is `Err` even when both name the same place:

```ant
Path.rel("/srv/app", "lib") // Err{ error: Error{ message: "no relative path leads from /srv/app to lib" } }
```

Make both absolute with [Path.abs](abs.md) first, as in
[A relative and an absolute path](#a-relative-and-an-absolute-path).

## See also

[Path.join](join.md), [Path.abs](abs.md), [Path.clean](clean.md), [Path.dir](dir.md),
[Path.isAbs](isAbs.md), [File.walk](../File/walk.md), [File.glob](../File/glob.md)
