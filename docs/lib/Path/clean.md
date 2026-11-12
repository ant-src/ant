<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.clean

```ant signature
func Path.clean(p: string): string
```

Path.clean returns the shortest path equivalent to p, working on the text alone: one separator for
each run, no . elements and no trailing separator, and each .. removed with the element before it. A
.. right after the root is dropped; one at the start of a relative path is kept.

<!-- end generated -->

## Details

- `Path.clean` never touches the file system: it does not check that the path exists and does not
  resolve symbolic links.
- The empty path cleans to `"."`.
- A `..` right after the root is dropped, so `Path.clean("/../a")` is `"/a"`; a `..` at the start
  of a relative path is kept, so `Path.clean("a/../..")` is `".."`.
- Paths follow the host's convention: `/` on Unix, `\` and drive letters on Windows. The examples on
  this page use `/`.

## Basic examples

Remove repeated separators, `.` elements, and `..` elements that cancel a name:

```ant
Path.clean("a//b/./c/..") // "a/b"
```

A trailing separator goes away:

```ant
Path.clean("src/lib/") // "src/lib"
```

## Scope

The empty path and `.` both clean to `.`:

```ant
Path.clean("")     // "."
Path.clean(".")    // "."
Path.clean("a/..") // "."
```

A leading `./` goes away:

```ant
Path.clean("./a/b") // "a/b"
```

A relative path keeps the `..` elements that climb above its start:

```ant
Path.clean("..")        // ".."
Path.clean("a/../..")   // ".."
Path.clean("../a/../b") // "../b"
```

An absolute path cannot climb above the root, so `..` at the root goes away:

```ant
Path.clean("/")             // "/"
Path.clean("//")            // "/"
Path.clean("/../a")         // "/a"
Path.clean("/a/b/../../..") // "/"
```

Only an element that is exactly `.` or `..` is special; a name that begins with dots is an ordinary
name:

```ant
Path.clean("..a/b/.c") // "..a/b/.c"
```

Elements are compared as text, so any name works, Unicode and spaces included:

```ant
Path.clean("日本/./語/")         // "日本/語"
Path.clean("my docs//notes.txt") // "my docs/notes.txt"
```

## Applications

Keep a user-supplied name inside a directory. Clean the name first, then reject one that is absolute
or climbs out with `..`:

```ant
func inside(root: string, name: string): string | null {
  let p = Path.clean(name)
  if Path.isAbs(p) || p == ".." || p.startsWith("../") { return null }
  Path.join([root, p])
}

println(inside("/srv/files", "docs/./a.txt"))
println(inside("/srv/files", "docs/../../etc/passwd"))
println(inside("/srv/files", "/etc/passwd"))
println(inside("/srv/files", "..hidden/a.txt"))
```

```text
/srv/files/docs/a.txt
null
null
/srv/files/..hidden/a.txt
```

Find the distinct files in a list of paths written in different ways:

```ant
["src/main.ant", "./src/main.ant", "src//lib/../main.ant", "src/util.ant"]
  | Array.map(Path.clean)
  | Array.unique() // ["src/main.ant", "src/util.ant"]
```

## Properties and relations

Cleaning a clean path changes nothing:

```ant
let paths = ["a//b/./c/..", "", "/../x", "../a/../b", "src/lib/"]
assert(paths.every({ Path.clean(Path.clean(x)) == Path.clean(x) }))
```

For non-empty parts, `Path.join` is `Path.clean` of the parts joined with the separator:

```ant
let parts = ["a", "b/../c", "d.ant"]
assert(Path.join(parts) == Path.clean(parts | String.join("/")))
```

## Possible issues

`Path.clean` works on the text, so it removes `name/..` even when `name` is a symbolic link. The
operating system follows the link first, so the cleaned path can name a different file. Use
[File.realPath](../File/realPath.md) to resolve links:

```ant
Path.clean("link/../config.toml") // "config.toml"
```

`Path.clean` does not make a relative path absolute; use [Path.abs](abs.md) for that:

```ant
Path.clean("./a/../b") // "b"
```

Checking for `..` with `startsWith("..")` rejects ordinary names such as `..hidden`. Compare the
element instead, as in the application above:

```ant
let p = Path.clean("..hidden/a.txt")
p.startsWith("..")               // true
p == ".." || p.startsWith("../") // false
```

## See also

[Path.join](join.md) · [Path.abs](abs.md) · [Path.rel](rel.md) · [Path.dir](dir.md) ·
[Path.base](base.md) · [File.realPath](../File/realPath.md)
