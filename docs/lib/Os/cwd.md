<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.cwd

```ant signature
func Os.cwd(): Result[string, Error]
```

Os.cwd returns the working directory as an absolute path: the directory the ant command was started
in, not the script's. Nothing in Ant changes it. It is Err when the host cannot name it, as when it
has been removed.

<!-- end generated -->

## Details

- The working directory is the one `ant` was started in, not the directory of the script. Every
  `File` function, and [Path.abs](../Path/abs.md), takes a relative path against it.
- The result is absolute. It is `Err` when the host cannot name the directory, as when it has been
  removed.
- No function changes the working directory. Join paths to the directory you need, or pass `dir:` to
  [Cmd.new](../Cmd/new.md) for a child process.

## Basic examples

Print the working directory:

<!-- test: check-only the working directory depends on where the program runs -->

```ant
println(try Os.cwd())
```

Started from `/home/ann/site`, it prints `/home/ann/site`.

`Os.cwd` returns a `Result`, so a program can report a failure:

```ant
let cwd = Os.cwd()
if cwd is Err {
  eprintln("no working directory: ${cwd.error.message}")
  Os.exit(1)
}
println(cwd.value == (try Path.abs(".")))
```

```text
true
```

## Applications

Name a file by its absolute path, so a log or message that mentions it still finds it from another
directory:

```ant
let cwd = try Os.cwd()
try File.write("report.txt", "all green")
let saved = Path.join([cwd, "report.txt"])
println(Path.isAbs(saved))
println(try File.read(saved))
```

```text
true
all green
```

Show a path relative to the working directory, as a tool shows the files it touched.
[Path.rel](../Path/rel.md) turns an absolute path back into the short one:

```ant
let cwd = try Os.cwd()
let found = [Path.join([cwd, "src", "main.ant"]), Path.join([cwd, "README.md"])]
for path in found {
  println(try Path.rel(cwd, path))
}
```

```text
src/main.ant
README.md
```

Find the nearest directory that holds a marker such as `.git`, from the working directory up to the
root. [Path.dir](../Path/dir.md) of the root is the root, which ends the search:

```ant
func findUp(start: string, marker: string): string | null {
  var dir = start
  while !File.exists(Path.join([dir, marker])) {
    let parent = Path.dir(dir)
    if parent == dir { return null }
    dir = parent
  }
  dir
}

let cwd = try Os.cwd()
try File.mkdir("repo/.git")
try File.mkdir("repo/src/app")
let root = findUp(Path.join([cwd, "repo/src/app"]), ".git")
if root != null {
  println(try Path.rel(cwd, root))
}
println(findUp(cwd, "no-such-marker"))
```

```text
repo
null
```

## Properties and relations

[Path.abs](../Path/abs.md) resolves a relative path against the working directory:

```ant
let cwd = try Os.cwd()
assert(Path.abs(".") == Os.cwd())
assert((try Path.abs("notes.txt")) == Path.join([cwd, "notes.txt"]))
```

A relative path and the same path joined to `Os.cwd()` name the same file:

```ant
let cwd = try Os.cwd()
try File.write("notes.txt", "hello")
assert(File.exists(Path.join([cwd, "notes.txt"])))
assert((try File.read(Path.join([cwd, "notes.txt"]))) == (try File.read("notes.txt")))
```

A child process starts in the same working directory unless `dir:` gives another. Here `pwd` prints
it, and [File.realPath](../File/realPath.md) settles links such as `/tmp` on macOS:

```ant
let cwd = try Os.cwd()
try File.mkdir("sub")
let here = try Cmd.read(Cmd.new("pwd"))
let there = try Cmd.read(Cmd.new("pwd", dir: "sub"))
assert((try File.realPath(here.trim())) == (try File.realPath(cwd)))
assert((try File.realPath(there.trim())) == (try File.realPath(Path.join([cwd, "sub"]))))
```

## Possible issues

`Os.cwd` returns a `Result`, not a `string`. Unwrap it with `try`:

```ant
let cwd: string = Os.cwd()
// error: cannot use Result[string, Error] where string is expected
```

In a string, the `Result` shows as it is:

<!-- test: check-only the working directory depends on where the program runs -->

```ant
println("working in ${Os.cwd()}")
println("working in ${try Os.cwd()}")
```

The first line prints `working in Ok{ value: "/home/ann/site" }`, the second
`working in /home/ann/site`.

The working directory is not the script's directory. Run from its parent, `tools/where.ant` does not
find itself by its bare name; [Os.program](program.md) leads to it:

<!-- test: cmd ant run tools/where.ant -->

```ant [tools/where.ant]
println(Os.program())
println(File.exists("where.ant"))
println(File.exists(Path.join([Path.dir(Os.program()), "where.ant"])))
```

```text
tools/where.ant
false
true
```

Climbing a relative path with [Path.dir](../Path/dir.md) stops at `"."`, the working directory, and
never looks above it. Start from `Os.cwd()` to reach the parents:

```ant
println(Path.dir("repo/src"))
println(Path.dir("repo"))
println(Path.dir("."))
let cwd = try Os.cwd()
println(Path.dir(cwd) == cwd)
```

```text
repo
.
.
false
```

## See also

[Path.abs](../Path/abs.md) · [Path.rel](../Path/rel.md) · [Path.join](../Path/join.md) ·
[Os.program](program.md) · [Os.home](home.md) · [File.realPath](../File/realPath.md) ·
[Cmd.new](../Cmd/new.md)
