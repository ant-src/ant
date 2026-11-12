<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.tempDir

```ant signature
func File.tempDir(pattern: string, dir: string | null = null): Result[string, Error]
```

File.tempDir creates a new empty directory, mode 0o700, in dir or the host's temporary directory,
and returns its path. Its name is pattern with the last \* replaced by a random string, or with the
string appended when there is no \*. Nothing removes it for the program.

<!-- end generated -->

## Details

- The mode is exactly `0o700`; the umask does not apply, unlike [File.mkdir](mkdir.md).
- Without `dir`, it is made in the host's temporary directory: `$TMPDIR`, or `/tmp` when that is
  unset, on Unix, and `%TEMP%` on Windows.
- Only the last `*` of `pattern` is replaced, so `"cache-*.d"` keeps its suffix. A `pattern` with a
  path separator is `Err`; put the parent in `dir`.
- Nothing removes the directory, at exit or later. Remove it with `defer { File.removeAll(dir) }`
  right after making it, and before calling [Os.exit](../Os/exit.md), which runs no deferred block.

## Basic examples

Make a scratch directory, remove it when the script ends, and work inside it:

```ant
let dir = try File.tempDir("build-*")
defer { File.removeAll(dir) }
try File.write(Path.join([dir, "out.txt"]), "done")
println(try File.list(dir))
```

```text
["out.txt"]
```

The name is the pattern with a random string in place of the `*`:

```ant
let dir = try File.tempDir("build-*")
defer { File.removeAll(dir) }
let name = Path.base(dir)
println(name.startsWith("build-"))
println(name.length() > "build-".length())
```

```text
true
true
```

## Scope

A `*` in the middle keeps what follows it, such as an extension:

```ant
let dir = try File.tempDir("snapshot-*.d")
defer { File.removeAll(dir) }
println(Path.base(dir).endsWith(".d"))
```

```text
true
```

Without a `*`, the random string goes at the end. Only the last `*` is replaced, so the others stay
in the name:

```ant
let plain = try File.tempDir("cache")
defer { File.removeAll(plain) }
let stars = try File.tempDir("a*b*")
defer { File.removeAll(stars) }
println(Path.base(plain).startsWith("cache"))
println(Path.base(stars).startsWith("a*b"))
```

```text
true
true
```

An empty pattern gives a name that is the random string alone:

```ant
let dir = try File.tempDir("")
defer { File.removeAll(dir) }
println(Path.base(dir).isEmpty())
```

```text
false
```

## Named arguments

### dir

`dir:` makes the directory there instead of in the host's temporary directory:

```ant
try File.mkdir("work")
let job = try File.tempDir("job-*", dir: "work")
println(Path.dir(job))
let entries = try File.list("work")
println(entries.length())
```

```text
work
1
```

The directory must already exist; `File.tempDir` does not create it:

```ant
println(File.tempDir("job-*", dir: "missing"))
try File.mkdir("missing")
println(File.tempDir("job-*", dir: "missing").isOk())
```

```text
Err{ error: Error{ message: "stat missing: no such file or directory" } }
true
```

## Applications

Run work in a directory of its own, removed however the work ends. The helper hands the directory to
a function and returns its result:

```ant
func withScratch[T](f: func(string): Result[T, Error]): Result[T, Error] {
  let dir = try File.tempDir("scratch-*")
  defer { File.removeAll(dir) }
  f(dir)
}

func countWords(dir: string): Result[int, Error] {
  let path = Path.join([dir, "words.txt"])
  try File.write(path, "one two three")
  let text = try File.read(path)
  Ok{ value: text.split(" ").length() }
}

println(try withScratch(countWords))
```

```text
3
```

Build a whole directory out of sight, then move it into place with [File.rename](rename.md). Making
it next to the target, with `dir: "."`, keeps the move inside one directory:

```ant
let tmp = try File.tempDir(".site-*", dir: ".")
try File.write(Path.join([tmp, "index.html"]), "<h1>Hi</h1>")
try File.write(Path.join([tmp, "about.html"]), "<h1>About</h1>")
try File.rename(tmp, "site")
println(try File.list("site"))
```

```text
["about.html", "index.html"]
```

## Properties and relations

Each call makes a different directory, new, empty, and private to its owner:

```ant
let a = try File.tempDir("t-*")
defer { File.removeAll(a) }
let b = try File.tempDir("t-*")
defer { File.removeAll(b) }
assert(a != b)
let info = try File.stat(a)
assert(info.isDir)
assert(info.mode == 0o700)
let entries = try File.list(a)
assert(entries.isEmpty())
```

## Possible issues

A path separator in the pattern is `Err`. Name the parent with `dir:` instead:

```ant
try File.mkdir("out")
println(File.tempDir("out/run-*"))
println(File.tempDir("run-*", dir: "out").isOk())
```

```text
Err{ error: Error{ message: "mkdirtemp out/run-*: pattern contains path separator" } }
true
```

[Os.exit](../Os/exit.md) runs no deferred block, so the directory stays behind. Remove it before
exiting:

<!-- test: exit 1 -->

```ant
let dir = try File.tempDir("job-*")
defer { File.removeAll(dir) }
if !File.exists(Path.join([dir, "result.txt"])) {
  File.removeAll(dir)
  Os.exit(1)
}
```

In a test, use [Test.tempDir](../Test/tempDir.md): it makes the directory and removes it when the
test ends, with no `defer` to forget.

## See also

[File.tempFile](tempFile.md) · [File.removeAll](removeAll.md) · [File.mkdir](mkdir.md) ·
[File.rename](rename.md) · [Test.tempDir](../Test/tempDir.md) · [Path.join](../Path/join.md) ·
[Os.exit](../Os/exit.md)
