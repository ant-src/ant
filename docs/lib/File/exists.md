<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.exists

```ant signature
func File.exists(path: string): bool
```

File.exists reports whether anything exists at path that [`File.stat`](stat.md) can reach, a
directory included. It follows a symbolic link, so a link that leads nowhere is false, and it is
false when stat is Err for any reason, such as a directory on the way that cannot be searched.

<!-- end generated -->

## Details

- `true` for a file and a directory alike; [`File.stat`](stat.md) tells them apart.
- The answer can be stale by the time you use it. To create a file only when nothing is there, write
  it with `replace: false` instead of checking first.

## Basic examples

<!-- test: file notes.txt
buy milk
-->

```ant
println(File.exists("notes.txt"))
println(File.exists("missing.txt"))
```

```text
true
false
```

## Scope

A directory exists too:

<!-- test: file logs/today.log
started
-->

```ant
println(File.exists("logs"))
println(File.exists("logs/today.log"))
```

```text
true
true
```

`File.exists` follows a symbolic link. A link to a file is `true`; a link whose target is missing is
`false`, even though the link itself is there, as [`File.stat`](stat.md) with `follow: false` shows:

<!-- test: file notes.txt
buy milk
-->

```ant
try File.symlink("notes.txt", "current")
defer { File.remove("current") }
try File.symlink("gone.txt", "broken")
defer { File.remove("broken") }

println(File.exists("current"))
println(File.exists("broken"))
println(File.stat("broken", follow: false).isOk())
```

```text
true
false
true
```

## Applications

Use the first configuration file that is present, or the defaults when there is none:

<!-- test: file app.toml
port = 8080
-->

```ant
let candidates = ["app.local.toml", "app.toml", "/etc/app/app.toml"]
let config = candidates.find({ File.exists(x) })
println("config: ${config ?? "built-in defaults"}")
```

```text
config: app.toml
```

Tell an existing report apart from other failures after a write with `replace: false`. The `Err`
does not say why the write failed, so check `File.exists` after it:

<!-- test: file report.txt
yesterday
-->

```ant
func save(path: string, text: string): Result[void, Error] {
  let written = File.write(path, text, replace: false)
  if written.isOk() || !File.exists(path) { return written }
  Result.fail("${path} is already there; remove it to write a new one")
}

println(save("report.txt", "today"))
```

```text
Err{ error: Error{ message: "report.txt is already there; remove it to write a new one" } }
```

## Properties and relations

A file exists after a successful write, and no longer after a successful remove:

```ant
try File.write("draft.txt", "first line")
assert(File.exists("draft.txt"))
try File.remove("draft.txt")
assert(!File.exists("draft.txt"))
```

## Possible issues

Checking and then writing is two steps, and another program can create the file between them. Let
`replace: false` decide in one step instead:

```ant
let output = "out.txt"
// Racy: the file can appear after the check.
if !File.exists(output) { try File.write(output, "first") }
// One step: Err when anything is at the path.
println(File.write(output, "second", replace: false))
```

```text
Err{ error: Error{ message: "open out.txt: file exists" } }
```

A dangling link is not there for `File.exists`, but it still occupies the name, so a write with
`replace: false` refuses it:

```ant
try File.symlink("gone.txt", "broken")
defer { File.remove("broken") }

println(File.exists("broken"))
println(File.write("broken", "text", replace: false))
```

```text
false
Err{ error: Error{ message: "open broken: file exists" } }
```

`true` does not mean the path is a file. Reading a directory is an `Err`:

<!-- test: file logs/today.log
started
-->

```ant
if File.exists("logs") {
  println(File.read("logs"))
}
```

```text
Err{ error: Error{ message: "read logs: is a directory" } }
```

## See also

[File.stat](stat.md) · [File.write](write.md) · [File.symlink](symlink.md) ·
[File.realPath](realPath.md) · [File.remove](remove.md) · [File.mkdir](mkdir.md)
