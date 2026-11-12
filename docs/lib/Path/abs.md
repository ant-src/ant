<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.abs

```ant signature
func Path.abs(p: string): Result[string, Error]
```

Path.abs returns p as an absolute path, resolved against the working directory.

<!-- end generated -->

## Details

- The working directory is the one `ant` was started in, as [`Os.cwd`](../Os/cwd.md) reports it, not
  the directory of the script.
- `Path.abs` works on the text: the path does not have to exist, and symbolic links are not
  resolved; use [`File.realPath`](../File/realPath.md) for the path with every link resolved.

## Basic examples

A relative path becomes the working directory joined with it:

```ant
let cwd = try Os.cwd()
let notes = try Path.abs("notes.txt")
println(try Path.rel(cwd, notes))
println(Path.isAbs(notes))
```

```text
notes.txt
true
```

A clean absolute path comes back unchanged:

```ant
Path.abs("/etc/hosts") // Ok{ value: "/etc/hosts" }
```

## Scope

Nothing needs to be at the path. A file that does not exist yet has an absolute path too:

```ant
let report = try Path.abs("out/report.html")
println(File.exists(report))
println(report.endsWith("/out/report.html"))
```

```text
false
true
```

## Applications

Show where a run writes its files, relative to the working directory when it is below it:

```ant
func describe(p: string): Result[string, Error] {
  let cwd = try Os.cwd()
  let full = try Path.abs(p)
  let rel = try Path.rel(cwd, full)
  let outside = rel == ".." || rel.startsWith("../")
  Result.ok(if outside { full } else { "./${rel}" })
}

println(try describe("build/app.js"))
println(try describe("/var/log/app.log"))
```

```text
./build/app.js
/var/log/app.log
```

Pass a child process an absolute path, so that it means the same file in the child's own `dir`:

<!-- test: file data/input.txt
hello
-->

```ant
try File.mkdir("work")
let input = try Path.abs("data/input.txt")
let text = try Cmd.read(Cmd.new("cat", args: [input], dir: "work"))
println(text)
```

```text
hello
```

## Properties and relations

The result is absolute, and `Path.abs` of a clean absolute path is that path:

```ant
let full = try Path.abs("src/main.ant")
assert(Path.isAbs(full))
assert(try Path.abs("/srv/app/main.ant") == "/srv/app/main.ant")
```

`Path.rel` from the working directory leads back to the relative path:

```ant
let cwd = try Os.cwd()
let full = try Path.abs("src/main.ant")
assert(try Path.rel(cwd, full) == "src/main.ant")
```

## Possible issues

`~` is not the home directory. `Path.abs` takes it as a directory named `~` inside the working
directory. `$NAME` is not expanded either; both are ordinary names:

```ant
let wrong = try Path.abs("~/notes.txt")
println(wrong.endsWith("/~/notes.txt"))
let right = Path.join([try Os.home(), "notes.txt"])
println(Path.isAbs(right))
```

```text
true
true
```

## See also

[Path.isAbs](isAbs.md) · [Path.rel](rel.md) · [Path.join](join.md) · [Path.clean](clean.md) ·
[Os.cwd](../Os/cwd.md) · [File.realPath](../File/realPath.md)
