<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.realPath

```ant signature
func File.realPath(path: string): Result[string, Error]
```

File.realPath returns path made absolute and clean, with every symbolic link along it resolved. It
is Err when the path leads nowhere.

<!-- end generated -->

## Details

- The path must exist: a missing file, or a link whose target is missing, is an `Err`. A link that
  leads back to itself is an `Err` too.
- A relative path is resolved against the working directory, and the working directory's own links
  are resolved as well.
- [`Path.abs`](../Path/abs.md) also makes a path absolute, but it reads only the text and leaves
  links in place.

## Basic examples

<!-- test: file notes.txt
buy milk
-->

```ant
try File.symlink("notes.txt", "current")
defer { File.remove("current") }

let here = try File.realPath(".")
let target = try File.realPath("current")
println(Path.isAbs(target))
println(try Path.rel(here, target))
```

```text
true
notes.txt
```

A path that leads nowhere is an `Err`:

```ant
println(File.realPath("missing.txt").isOk())
```

```text
false
```

## Scope

A link anywhere along the path is resolved, not only the last element, and `.` and `..` are cleaned
away:

<!-- test: file logs/2026/today.log
started
-->

```ant
try File.symlink("logs/2026", "latest")
defer { File.remove("latest") }

let here = try File.realPath(".")
println(try Path.rel(here, try File.realPath("latest/today.log")))
println(try Path.rel(here, try File.realPath("./logs/../logs/2026")))
```

```text
logs/2026/today.log
logs/2026
```

A link whose target is missing, a link that points at itself, and a path that goes through a file
are all `Err`:

<!-- test: file notes.txt
buy milk
-->

```ant
try File.symlink("gone.txt", "broken")
defer { File.remove("broken") }
try File.symlink("loop", "loop")
defer { File.remove("loop") }

println(File.realPath("broken").isOk())
println(File.realPath("loop").isOk())
println(File.realPath("notes.txt/more").isOk())
```

```text
false
false
false
```

## Applications

Keep a requested file inside a served directory. A link inside the directory can point anywhere, so
compare the resolved paths, not the text:

<!-- test: file public/index.html
<h1>Home</h1>
-->

<!-- test: file secret.txt
password
-->

```ant
func inside(root: string, path: string): Result[bool, Error] {
  let realRoot = try File.realPath(root)
  let real = try File.realPath(Path.join([root, path]))
  let rel = try Path.rel(realRoot, real)
  Result.ok(rel != ".." && !rel.startsWith("../"))
}

try File.symlink("../secret.txt", "public/notes.txt")
defer { File.remove("public/notes.txt") }

println(inside("public", "index.html"))
println(inside("public", "notes.txt"))
println(inside("public", "../secret.txt"))
```

```text
Ok{ value: true }
Ok{ value: false }
Ok{ value: false }
```

Tell whether two paths name the same file:

<!-- test: file notes.txt
buy milk
-->

```ant
func same(a: string, b: string): Result[bool, Error] {
  Result.ok(try File.realPath(a) == try File.realPath(b))
}

try File.symlink("notes.txt", "current")
defer { File.remove("current") }

println(same("current", "./notes.txt"))
println(same("current", "."))
```

```text
Ok{ value: true }
Ok{ value: false }
```

## Properties and relations

The result is absolute, clean, and already resolved, so resolving it again changes nothing:

<!-- test: file notes.txt
buy milk
-->

```ant
try File.symlink("notes.txt", "current")
defer { File.remove("current") }

let real = try File.realPath("current")
assert(Path.isAbs(real))
assert(Path.clean(real) == real)
assert(try File.realPath(real) == real)
```

`Path.abs` keeps the name of the link; `File.realPath` gives the name of the file it leads to:

<!-- test: file notes.txt
buy milk
-->

```ant
try File.symlink("notes.txt", "current")
defer { File.remove("current") }

println(Path.base(try Path.abs("current")))
println(Path.base(try File.realPath("current")))
```

```text
current
notes.txt
```

## Possible issues

The working directory itself can sit behind a link, as the temporary directories of macOS do, so
`File.realPath(".")` and `Path.abs(".")` can differ. Compare resolved paths only with resolved
paths, and turn them into relative ones with `Path.rel` against `File.realPath(".")`, not against
`Path.abs(".")`.

`File.realPath` fails for a file that does not exist yet. To place a new file, resolve its directory
and join the name on:

<!-- test: file out/.keep
-->

```ant
let name = "report.txt"
println(File.realPath(Path.join(["out", name])).isOk())

let dir = try File.realPath("out")
let path = Path.join([dir, name])
try File.write(path, "done")
println(File.read("out/report.txt"))
```

```text
false
Ok{ value: "done" }
```

## See also

[File.readLink](readLink.md) · [File.symlink](symlink.md) · [File.stat](stat.md) ·
[File.exists](exists.md) · [Path.abs](../Path/abs.md) · [Path.rel](../Path/rel.md) ·
[Path.clean](../Path/clean.md)
