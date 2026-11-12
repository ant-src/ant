<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.tempFile

```ant signature
func File.tempFile(pattern: string, dir: string | null = null): Result[string, Error]
```

File.tempFile creates a new empty file, mode 0o600, in dir or the host's temporary directory, and
returns its path. Its name is pattern with the last \* replaced by a random string, or with the
string appended when there is no \*. Nothing removes it for the program.

<!-- end generated -->

## Details

- `File.tempFile` creates the file only when its name is free, so it never replaces anything, a
  planted symbolic link included.
- `File.tempFile` returns a path, not a handle. Write to it with [`File.write`](write.md) or open it
  with [`File.create`](create.md).
- A `pattern` that holds a path separator is `Err`. Give the directory with `dir:` instead.
- `dir` is the host's temporary directory when it is `null`: `$TMPDIR`, or `/tmp` when that is
  unset, on Unix, and `%TEMP%` on Windows. The directory must exist.
- Nothing removes the file, at exit or later. Remove it yourself, usually with
  `defer { File.remove(path) }` right after the call. [`Os.exit`](../Os/exit.md) runs no deferred
  block.

## Basic examples

Create a temporary file, use it, and remove it when the program is done:

```ant
let path = try File.tempFile("notes-*.txt")
defer { File.remove(path) }
try File.write(path, "draft")
println(try File.read(path))
```

```text
draft
```

The random string replaces the `*`, so the name keeps the prefix and the extension of `pattern`:

```ant
let path = try File.tempFile("report-*.csv")
defer { File.remove(path) }
let name = Path.base(path)
println(name.startsWith("report-"))
println(name.endsWith(".csv"))
```

```text
true
true
```

## Scope

Without a `*`, the random string goes at the end of `pattern`:

```ant
let path = try File.tempFile("job-", dir: ".")
defer { File.remove(path) }
let suffix = Path.base(path).trimPrefix("job-")
println(suffix.isEmpty())
```

```text
false
```

Only the last `*` is replaced; the others stay in the name:

```ant
let path = try File.tempFile("a*b*.txt", dir: ".")
defer { File.remove(path) }
println(Path.base(path).startsWith("a*b"))
println(Path.base(path).contains("*b*"))
```

```text
true
false
```

The new file is empty, and only its owner can read and write it:

```ant
let path = try File.tempFile("secret-*")
defer { File.remove(path) }
println((try File.read(path)).isEmpty())
let info = try File.stat(path)
println(String.fromInt(info.mode, base: 8))
```

```text
true
600
```

A `pattern` with a path separator is `Err`:

```ant
println(File.tempFile("cache/entry-*"))
```

```text
Err{ error: Error{ message: "createtemp cache/entry-*: pattern contains path separator" } }
```

## Named arguments

### dir

With `dir:`, `File.tempFile` creates the file in that directory instead of the host's temporary
directory:

```ant
try File.mkdir("work")
let path = try File.tempFile("job-*.log", dir: "work")
defer { File.remove(path) }
println(Path.dir(path))
println((try File.glob("job-*.log", dir: "work")).length())
```

```text
work
1
```

The directory must exist:

```ant
println(File.tempFile("job-*", dir: "missing") is Err)
```

```text
true
```

## Applications

Save a file so that a reader sees the old contents or the new ones, never half a write. Write a
temporary file next to the target, in the same directory, then rename it over the target. The
deferred `File.remove` cleans up when a step fails; after the rename nothing is left for it to
remove:

```ant
func save(path: string, text: string): Result[void, Error] {
  let tmp = try File.tempFile(".${Path.base(path)}-*", dir: Path.dir(path))
  defer { File.remove(tmp) }
  try File.write(tmp, text)
  File.rename(tmp, path)
}

try save("config.json", "{\"debug\": false}")
try save("config.json", "{\"debug\": true}")
println(try File.read("config.json"))
println(try File.glob(".config.json-*", hidden: true))
```

```text
{"debug": true}
[]
```

Hand a function that takes a path some input of your own, without touching the working directory, as
a test does. Add context to a failure with [`Result.wrapErr`](../Result/wrapErr.md):

```ant
func countLines(path: string): Result[int, Error] {
  let text = try File.read(path)
  Result.ok(text.split("\n").filter({ !x.isEmpty() }).length())
}

func countText(text: string): Result[int, Error] {
  let path = try (File.tempFile("input-*.txt") | Result.wrapErr("count lines"))
  defer { File.remove(path) }
  try File.write(path, text)
  countLines(path)
}

println(countText("one\ntwo\nthree\n"))
```

```text
Ok{ value: 3 }
```

## Properties and relations

Every call picks a new name, so two calls with the same `pattern` never return the same path:

```ant
let a = try File.tempFile("same-*")
defer { File.remove(a) }
let b = try File.tempFile("same-*")
defer { File.remove(b) }
assert(a != b)
```

The path lies in `dir`:

```ant
try File.mkdir("out")
let path = try File.tempFile("part-*", dir: "out")
defer { File.remove(path) }
assert(Path.dir(path) == "out")
assert(File.exists(path))
```

## Possible issues

The file already exists when `File.tempFile` returns, so a write with `replace: false` fails. Write
with the default `replace: true`:

```ant
let path = try File.tempFile("out-*", dir: ".")
defer { File.remove(path) }
let refused = File.write(path, "data", replace: false)
println(refused is Err)
println(File.write(path, "data"))
```

```text
true
Ok{ value: void }
```

[`Os.exit`](../Os/exit.md) runs no deferred block, so a file meant to be removed by `defer` stays.
Remove it before the exit:

<!-- test: exit 2 -->

```ant
let path = try File.tempFile("run-*")
defer { File.remove(path) }
try File.remove(path)
Os.exit(2)
```

A file made by `File.tempFile` keeps mode `0o600` when it is renamed into place, so only its owner
can read the result:

```ant
let target = "shared.txt"
let tmp = try File.tempFile(".shared-*", dir: ".")
defer { File.remove(tmp) }
try File.write(tmp, "hello")
try File.rename(tmp, target)
println(String.fromInt((try File.stat(target)).mode, base: 8))
```

```text
600
```

When others must read the file, set its mode with [`File.chmod`](chmod.md) before the rename:

```ant
let target = "public.txt"
let tmp = try File.tempFile(".public-*", dir: ".")
defer { File.remove(tmp) }
try File.write(tmp, "hello")
try File.chmod(tmp, 0o644)
try File.rename(tmp, target)
println(String.fromInt((try File.stat(target)).mode, base: 8))
```

```text
644
```

## See also

[File.tempDir](tempDir.md) · [File.remove](remove.md) · [File.removeAll](removeAll.md) ·
[File.rename](rename.md) · [File.write](write.md) · [File.create](create.md)
