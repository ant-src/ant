<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.join

```ant signature
func Path.join(parts: Array[string]): string
```

Path.join joins parts with the path separator and cleans the result. It skips empty parts, so no
parts, or empty ones only, give ".".

<!-- end generated -->

## Details

- Empty parts are skipped, so `Path.join(["", "a"])` is `"a"`, not `"/a"`. No parts, or only empty
  ones, give `"."`.
- A later absolute part does not replace the parts before it: `Path.join(["a", "/etc"])` is
  `"a/etc"`.
- The result is cleaned, so `..` removes the element before it and can climb above the first part.
- Paths use the host's separator: `/` on Unix, `\` on Windows. The examples on this page show Unix
  paths.

## Basic examples

Join directory names and a file name:

```ant
Path.join(["src", "lib", "main.ant"]) // "src/lib/main.ant"
```

An absolute first part gives an absolute path:

```ant
Path.join(["/usr", "local", "bin"]) // "/usr/local/bin"
```

The result is cleaned, so `..` and `.` disappear:

```ant
Path.join(["a", "b/../c", "d.ant"]) // "a/c/d.ant"
```

## Scope

No parts, or only empty ones, give the current directory:

```ant
Path.join([])       // "."
Path.join(["", ""]) // "."
```

An empty part between two others adds nothing:

```ant
Path.join(["a", "", "b"]) // "a/b"
```

One part is cleaned like [Path.clean](clean.md):

```ant
Path.join(["./docs//guide/"]) // "docs/guide"
```

Separators at the edges of parts collapse into one:

```ant
Path.join(["a/", "/b/"]) // "a/b"
```

`..` climbs past the parts before it, and stops at the root:

```ant
Path.join(["..", "x"])    // "../x"
Path.join(["a", "../.."]) // ".."
Path.join(["/", ".."])    // "/"
```

Parts are text, so any Unicode name works:

```ant
Path.join(["café", "日本.txt"]) // "café/日本.txt"
```

Pass the parts through a pipe. An array has no `join` of its own to call with `.`:

```ant
["src", "main.ant"] | Path.join() // "src/main.ant"
```

Name the argument:

```ant
Path.join(parts: ["x", "y"]) // "x/y"
```

## Applications

Put each output file in a build directory, with a new extension:

```ant
let sources = ["src/main.ant", "src/util/strings.ant"]
for source in sources {
  let name = Path.base(source).trimSuffix(Path.ext(source))
  println(Path.join(["build", Path.dir(source), "${name}.js"]))
}
```

```text
build/src/main.js
build/src/util/strings.js
```

Write a file inside a temporary directory:

```ant
let dir = Result.unwrap(File.tempDir("notes-*"))
defer { File.removeAll(dir) }
let path = Path.join([dir, "today.txt"])
Result.unwrap(File.write(path, "buy milk\n"))
println(Result.unwrap(File.list(dir)))
```

```text
["today.txt"]
```

Build a path under the home directory. `Os.home` can fail, so the function returns a `Result`:

```ant
func notesPath(): Result[string, Error] {
  Result.ok(Path.join([try Os.home(), "notes.txt"]))
}

let path = Result.unwrap(notesPath())
println(Path.base(path))
println(Path.isAbs(path))
```

```text
notes.txt
true
```

Build the parts from data, then join them:

```ant
type Post = { year: int, month: int, slug: string }
let post = Post{ year: 2026, month: 11, slug: "paths" }
let parts = ["posts", "${post.year}", "${post.month}".padStart(2, "0"), "${post.slug}.md"]
println(Path.join(parts))
```

```text
posts/2026/11/paths.md
```

## Properties and relations

The result is already clean:

```ant
let parts = ["a/./b", "..", "c//d"]
assert(Path.clean(Path.join(parts)) == Path.join(parts))
```

Joining the parts is cleaning them glued with `/`, when no part is empty:

```ant
let parts = ["src", "../lib", "x.ant"]
assert(Path.join(parts) == Path.clean(parts | String.join("/")))
```

[Path.dir](dir.md) and [Path.base](base.md) split a path into parts that join back to it, cleaned:

```ant
for p in ["a/b/c.txt", "c.txt", "/etc/hosts", "./a//b"] {
  assert(Path.join([Path.dir(p), Path.base(p)]) == Path.clean(p))
}
```

## Possible issues

An absolute part in the middle does not start over, as it does in some other languages. It is joined
like any other part:

```ant
Path.join(["/home/ann", "/etc/hosts"]) // "/home/ann/etc/hosts"
```

Since `..` climbs out of the first part, joining a name a user gives can leave the directory you
meant:

```ant
Path.join(["/srv/files", "../../etc/passwd"]) // "/etc/passwd"
```

Check with [Path.rel](rel.md) that the joined path stays inside the directory:

```ant
func inside(root: string, name: string): Result[string, Error] {
  let path = Path.join([root, name])
  let rel = try Path.rel(root, path)
  if rel == ".." || rel.startsWith("../") {
    return Result.fail("${name}: outside ${root}")
  }
  Result.ok(path)
}

println(inside("/srv/files", "docs/a.txt"))
println(inside("/srv/files", "../../etc/passwd"))
```

```text
Ok{ value: "/srv/files/docs/a.txt" }
Err{ error: Error{ message: "../../etc/passwd: outside /srv/files" } }
```

A `~` in a part is not expanded. The result is a relative path that names a directory called `~`:

```ant
Path.join(["~", "notes.txt"]) // "~/notes.txt"
```

Use [Os.home](../Os/home.md) for the home directory, as in Applications above.

`Path.join` takes one array, not the parts as separate arguments:

```ant
let path = Path.join("a", "b") // error: Path.join takes 1 argument, got 2
```

Gluing parts with string interpolation does not clean them, and doubles separators:

```ant
let dir = "out/"
"${dir}/report.txt"            // "out//report.txt"
Path.join([dir, "report.txt"]) // "out/report.txt"
```

## See also

[Path.clean](clean.md), [Path.dir](dir.md), [Path.base](base.md), [Path.ext](ext.md),
[Path.rel](rel.md), [Path.abs](abs.md), [Path.isAbs](isAbs.md), [String.join](../String/join.md)
