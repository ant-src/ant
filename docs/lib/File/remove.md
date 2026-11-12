<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.remove

```ant signature
func File.remove(path: string): Result[void, Error]
```

File.remove removes the file or empty directory at path. It is Err on a directory that has entries;
[`File.removeAll`](removeAll.md) removes a whole tree.

<!-- end generated -->

## Details

- Nothing at `path` is `Err`, unlike [File.removeAll](removeAll.md), which is `Ok` then.
- At a symbolic link it removes the link, never what the link points at.

## Basic examples

Remove a file:

<!-- test: file draft.txt
first line
-->

```ant
println(File.remove("draft.txt"))
println(File.exists("draft.txt"))
```

```text
Ok{ value: void }
false
```

Remove a directory once it is empty:

```ant
try File.mkdir("empty")
println(File.remove("empty"))
```

```text
Ok{ value: void }
```

## Scope

A directory with entries is `Err`, and so is a path where nothing is:

<!-- test: file build/app.js
console.log("hi")
-->

```ant
println(File.remove("build"))
println(File.remove("missing.txt"))
```

```text
Err{ error: Error{ message: "remove build: directory not empty" } }
Err{ error: Error{ message: "remove missing.txt: no such file or directory" } }
```

A symbolic link is removed itself, and its target stays. A dangling link is removed the same way:

<!-- test: file notes.txt
buy milk
-->

```ant
try File.symlink("notes.txt", "current")
try File.symlink("gone.txt", "broken")
println(File.remove("current"))
println(File.remove("broken"))
println(File.exists("notes.txt"))
```

```text
Ok{ value: void }
Ok{ value: void }
true
```

## Applications

Remove a temporary file however the function returns. After the rename the deferred remove is `Err`,
and a deferred block's value is discarded:

```ant
func save(path: string, text: string): Result[void, Error] {
  let tmp = try File.tempFile(".save-*", dir: Path.dir(path))
  defer { File.remove(tmp) }
  try File.write(tmp, text)
  File.rename(tmp, path)
}

println(save("out.txt", "hi"))
println(try File.glob(".save-*", hidden: true))
```

```text
Ok{ value: void }
[]
```

Delete the log files of a directory and keep everything else:

<!-- test: file logs/app.log
started
-->
<!-- test: file logs/db.log
connected
-->
<!-- test: file logs/README
rotated daily
-->

```ant
for path in try File.glob("*.log", dir: "logs") {
  try File.remove(Path.join(["logs", path]))
}
println(try File.list("logs"))
```

```text
["README"]
```

Remove a file that may already be gone, and fail only when it is still there afterwards:

<!-- test: file cache.json
{}
-->

```ant
func discard(path: string): Result[void, Error] {
  let removed = File.remove(path)
  if removed.isOk() || !File.exists(path) { return Result.ok(void) }
  removed
}

println(discard("cache.json"))
println(discard("cache.json"))
```

```text
Ok{ value: void }
Ok{ value: void }
```

## Properties and relations

After an `Ok`, nothing is at the path, and a second call is `Err`:

```ant
try File.write("draft.txt", "first line")
assert(File.remove("draft.txt").isOk())
assert(!File.exists("draft.txt"))
assert(!File.remove("draft.txt").isOk())
```

On a file or an empty directory `File.remove` and [File.removeAll](removeAll.md) do the same; on a
missing path only `File.removeAll` is `Ok`:

```ant
try File.mkdir("a")
try File.mkdir("b")
assert(File.remove("a").isOk() == File.removeAll("b").isOk())
assert(!File.remove("a").isOk())
assert(File.removeAll("b").isOk())
```

## Possible issues

`File.remove` does not empty a directory first. Use [File.removeAll](removeAll.md) for a tree:

<!-- test: file out/report.txt
done
-->

```ant
println(File.remove("out"))
println(File.removeAll("out"))
```

```text
Err{ error: Error{ message: "remove out: directory not empty" } }
Ok{ value: void }
```

A call on its own line drops the `Result`, so a failed removal goes unnoticed. Use `try` to stop at
the error; at the top of a script it prints the message and exits with status 1:

<!-- test: exit 1 -->

```ant
File.remove("report.txt")
println("removed, or so it seems")
try File.remove("report.txt")
println("not reached")
```

```text
removed, or so it seems
```

## See also

[File.removeAll](removeAll.md) · [File.exists](exists.md) · [File.rename](rename.md) ·
[File.tempFile](tempFile.md) · [File.tempDir](tempDir.md) · [File.mkdir](mkdir.md) ·
[File.symlink](symlink.md) · [File.glob](glob.md)
