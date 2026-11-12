<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.base

```ant signature
func Path.base(p: string): string
```

Path.base returns the last element of p, or "." when p is empty.

<!-- end generated -->

## Details

- `Path.base` looks only at the text. It does not touch the file system and does not clean `p`
  first, so `Path.base("a/..")` is `".."`.

## Basic examples

The file name at the end of a path:

```ant
Path.base("src/main.ant")   // "main.ant"
Path.base("/usr/local/bin") // "bin"
```

A name with no directory is its own base:

```ant
Path.base("README.md") // "README.md"
```

## Scope

The empty path gives `"."`:

```ant
Path.base("") // "."
```

`.`, `..`, and hidden names are elements like any other:

```ant
Path.base(".")             // "."
Path.base("../..")         // ".."
Path.base("home/ann/.env") // ".env"
```

Unicode names come back whole:

```ant
Path.base("docs/résumé.pdf") // "résumé.pdf"
Path.base("фото/кот.jpg")    // "кот.jpg"
```

`Path.base` works as a function value:

```ant
let paths = ["src/a.ant", "src/lib/b.ant", "test/c_test.ant"]
paths.map(Path.base) // ["a.ant", "b.ant", "c_test.ant"]
```

It is in the `Path` namespace, not `String`, so a string has no `.base()`:

```ant
let p = "src/a.ant"
let name = p.base()
// error: string has no field or function base
```

## Applications

The name without its extension, with [Path.ext](ext.md):

```ant
let path = "reports/2024/summary.csv"
let name = Path.base(path)
name.trimSuffix(Path.ext(name)) // "summary"
```

Move a file into another directory under the same name:

```ant
let from = "downloads/invoice-42.pdf"
Path.join(["archive", Path.base(from)]) // "archive/invoice-42.pdf"
```

Find file names that occur in more than one directory:

```ant
let files = ["a/config.json", "b/data.csv", "c/config.json", "d/notes.txt"]
let byName = files | Array.groupBy(Path.base)
byName
  | Map.entries()
  | Array.filter({ x.value.length() > 1 })
  | Array.map({ x.key }) // ["config.json"]
```

List the files a directory holds by name only:

```ant
try File.mkdir("logs")
try File.write("logs/app.log", "")
try File.write("logs/db.log", "")
let names = (try File.glob("logs/*.log")) | Array.map(Path.base) | String.join(", ")
println(names)
```

```text
app.log, db.log
```

## Properties and relations

For a cleaned path with no trailing separator, [Path.join](join.md) of [Path.dir](dir.md) and
`Path.base` gives the path back:

```ant
let p = "src/lib/util.ant"
assert(Path.join([Path.dir(p), Path.base(p)]) == p)
```

The base of a base is the same name:

```ant
let p = "a/b/c.txt"
assert(Path.base(Path.base(p)) == Path.base(p))
```

## Possible issues

`Path.base` does not resolve `..`. Clean the path first with [Path.clean](clean.md) to get the name
the path leads to:

```ant
Path.base("photos/2024/..")             // ".."
Path.base(Path.clean("photos/2024/..")) // "photos"
```

The empty path gives `"."`, not `""`, so a missing name looks like the current directory. Check for
the empty string before taking the name:

```ant
func fileName(input: string): string | null {
  if input.isEmpty() { return null }
  Path.base(input)
}

fileName("")        // null
fileName("a/b.txt") // "b.txt"
```

The separator is the host's. On Unix a backslash is part of the name, so a Windows-style path comes
back whole. On Windows both `/` and `\` separate:

```ant
Path.base("dir\\file.txt") // "dir\\file.txt"
```

## See also

[Path.dir](dir.md) · [Path.ext](ext.md) · [Path.join](join.md) · [Path.clean](clean.md) ·
[Path.abs](abs.md) · [File.glob](../File/glob.md)
