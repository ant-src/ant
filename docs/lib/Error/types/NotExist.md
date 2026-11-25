<!-- status: draft -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Error.NotExist

```ant signature
type Error.NotExist = {}
```

Error.NotExist is the kind of an error for a path that nothing is at.

<!-- end generated -->

## Details

- The library sets it when nothing is at a path a call names: a read, a stat, an open or a remove of
  a missing path, a write into a directory that is missing, a `Cmd` whose `dir` or program path is
  missing.
- A path that leads through a file, such as `notes.txt/today`, is not `Error.NotExist`: its `kind`
  is `null`.
- A `Cmd` that names a program without a slash, which no directory of `PATH` holds, has no kind. Ask
  [Os.which](../../Os/which.md) before the call.
- `Error.NotExist` is an empty record, so the value is `Error.NotExist{}`.

## Basic examples

Use defaults when a file is missing:

```ant
let saved = File.read("theme.txt")
let theme = if saved is Err && saved.error.kind is Error.NotExist { "light" } else { try saved }
println(theme)
```

```text
light
```

The printed error carries the kind:

```ant
println(File.read("theme.txt"))
```

```text
Err{ error: Error{ message: "open theme.txt: no such file or directory", kind: Error.NotExist{} } }
```

## Scope

Calls that name a missing path give it, whatever they were going to do there:

```ant
println(File.stat("draft.md"))
println(File.remove("draft.md"))
println(File.write("drafts/today.md", "x"))
println(File.rename("draft.md", "final.md"))
```

```text
Err{ error: Error{ message: "stat draft.md: no such file or directory", kind: Error.NotExist{} } }
Err{ error: Error{ message: "remove draft.md: no such file or directory", kind: Error.NotExist{} } }
Err{ error: Error{ message: "open drafts/today.md: no such file or directory", kind: Error.NotExist{} } }
Err{ error: Error{ message: "rename draft.md final.md: no such file or directory", kind: Error.NotExist{} } }
```

A command whose working directory is missing gives it too:

```ant
println(Cmd.new("ls", [], dir: "nowhere").run())
```

```text
Err{ error: Error{ message: "chdir nowhere: no such file or directory", kind: Error.NotExist{} } }
```

The kind survives the context a caller adds:

```ant
let r = File.read("app.json") | Result.wrapErr("load config")
if r is Err {
  println(r.error.message)
  println(r.error.kind is Error.NotExist)
}
```

```text
load config: open app.json: no such file or directory
true
```

## Applications

Load a setting with a default, and fail on a file that is there and holds something else:

```ant
func loadPort(path: string): Result[int, Error] {
  let text = File.read(path)
  if text is Err && text.error.kind is Error.NotExist {
    return Result.ok(8080)
  }
  let port = (try text).trim().toInt() ?? return Result.fail("${path}: not a port")
  Result.ok(port)
}

println(loadPort("port.txt"))
try File.write("port.txt", "80x")
println(loadPort("port.txt"))
```

```text
Ok{ value: 8080 }
Err{ error: Error{ message: "port.txt: not a port", kind: null } }
```

Count the lines of the logs that are there, and skip the ones a rotation already removed:

```ant
func lineCount(path: string): Result[int, Error] {
  let text = File.read(path)
  if text is Err && text.error.kind is Error.NotExist {
    return Result.ok(0)
  }
  Result.ok((try text).lines().length())
}

try File.write("app.log", "started\nstopped\n")
println(try lineCount("app.log"))
println(try lineCount("app.log.1"))
```

```text
2
0
```

## Properties and relations

[File.removeAll](../../File/removeAll.md) and [File.exists](../../File/exists.md) take a missing
path as an answer, not a failure, so they give no `Error.NotExist`:

```ant
assert(File.removeAll("build").isOk())
assert(!File.exists("build"))
```

## Possible issues

A path that leads through a file is a different failure, with no kind:

```ant
try File.write("notes.txt", "x")
println(File.read("notes.txt/today"))
```

```text
Err{ error: Error{ message: "open notes.txt/today: not a directory", kind: null } }
```

`File.exists` is `false` for a missing path and also for one it cannot reach. When the difference
matters, test the kind of [File.stat](../../File/stat.md)'s `Err`:

```ant
let info = File.stat("backups/monday.tar")
if info is Err && !(info.error.kind is Error.NotExist) {
  println("cannot look: ${info.error.message}")
} else {
  println("there or missing")
}
```

```text
there or missing
```

## See also

[Error.Kind](Kind.md) · [Error.Exist](Exist.md) · [Error.Permission](Permission.md) ·
[Error](../index.md) · [File.read](../../File/read.md) · [File.stat](../../File/stat.md) ·
[File.exists](../../File/exists.md) · [Os.which](../../Os/which.md)
