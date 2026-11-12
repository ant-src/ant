<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.dir

```ant signature
func Path.dir(p: string): string
```

Path.dir returns p up to its last separator, cleaned, or "." when there is none.

<!-- end generated -->

## Details

- Works on the text alone: the path does not have to exist, and nothing on disk is read.
- Separators are the host's: `/` on Unix, `\` and drive letters on Windows.
- `Path.dir` is the path up to its last separator, cleaned, so `Path.dir("a/b/")` is `"a/b"` and
  `Path.dir("/")` is `"/"`. It rebuilds `p` with [Path.base](base.md) only when `p` does not end in
  a separator.

## Basic examples

The directory part of a file path:

```ant
Path.dir("a/b.ant") // "a"
Path.dir("b.ant")   // "."
```

An absolute path keeps its root:

```ant
Path.dir("/home/ann/notes.txt") // "/home/ann"
```

## Scope

A path with one element, and the empty path, have no directory part, so the result is `"."`:

```ant
Path.dir("notes.txt") // "."
Path.dir("")          // "."
```

An element right under the root gives the root:

```ant
Path.dir("/etc") // "/"
```

Paths that climb with `..` keep it:

```ant
Path.dir("../shared/lib.ant") // "../shared"
```

Elements of any script come back as they are:

```ant
Path.dir("文書/日本語.txt") // "文書"
```

Call it in a pipe, with the argument named, or inside a lambda:

```ant
"src/main.ant" | Path.dir()                 // "src"
Path.dir(p: "src/main.ant")                 // "src"
["a/x.ant", "b/y.ant"].map({ Path.dir(x) }) // ["a", "b"]
```

## Applications

Name a file next to another one, in the same directory:

```ant
func sibling(path: string, name: string): string {
  Path.join([Path.dir(path), name])
}

sibling("config/app.toml", "app.lock") // "config/app.lock"
sibling("app.toml", "app.lock")        // "app.lock"
```

Group files by the directory that holds them:

```ant
let files = ["src/main.ant", "src/util.ant", "tests/main_test.ant", "README.md"]
let byDir = files | Array.groupBy({ Path.dir(x) })
println(byDir)
```

```text
Map{ "src": ["src/main.ant", "src/util.ant"], "tests": ["tests/main_test.ant"], ".": ["README.md"] }
```

List every directory above a relative path, nearest first:

```ant
func parents(path: string): Array[string] {
  let dirs: Array[string] = []
  var dir = Path.dir(path)
  while dir != "." {
    dirs.push(dir)
    dir = Path.dir(dir)
  }
  dirs
}

println(parents("a/b/c/d.ant"))
```

```text
["a/b/c", "a/b", "a"]
```

Create the directory a file goes in before writing it:

```ant
func save(path: string, text: string): Result[void, Error] {
  try File.mkdir(Path.dir(path)).wrapErr("save ${path}")
  File.write(path, text).wrapErr("save ${path}")
}

save("out/reports/today.txt", "done").unwrap()
println(File.read("out/reports/today.txt").unwrap())
```

```text
done
```

## Properties and relations

For a path that does not end in a separator, `Path.dir` and [`Path.base`](base.md) split it into two
parts that [`Path.join`](join.md) joins back into the cleaned path:

```ant
for p in ["a/b.ant", "b.ant", "/home/ann/notes.txt", "../shared/lib.ant"] {
  assert(Path.join([Path.dir(p), Path.base(p)]) == Path.clean(p))
}
```

## Possible issues

A trailing separator makes the last element empty, so `Path.dir` returns the directory itself, not
its parent:

```ant
Path.dir("logs/2026/") // "logs/2026"
Path.dir("logs/2026")  // "logs"
```

Remove the separator first with [`String.trimSuffix`](../String/trimSuffix.md) when the path may end
in one:

```ant
"logs/2026/".trimSuffix("/") | Path.dir() // "logs"
```

`Path.dir` takes a `string`, which belongs to the `String` namespace, so a dot call does not find
it. Call it by its namespace or in a pipe:

```ant
let path = "src/main.ant"
println(path.dir())
// error: string has no field or function dir
```

## See also

[Path.base](base.md) · [Path.ext](ext.md) · [Path.join](join.md) · [Path.clean](clean.md) ·
[Path.abs](abs.md) · [File.mkdir](../File/mkdir.md)
