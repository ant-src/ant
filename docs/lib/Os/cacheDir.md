<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.cacheDir

```ant signature
func Os.cacheDir(): Result[string, Error]
```

Os.cacheDir returns the directory where the user's programs keep what can be made again:
$XDG\_CACHE\_HOME, or $HOME/.cache when that is unset or empty, on Unix, $HOME/Library/Caches on
macOS, %LocalAppData% on Windows. The directory need not exist.

<!-- end generated -->

## Details

- The result is `Err` when the variable it rests on is unset, such as `$HOME` on Unix and macOS, or
  when `$XDG_CACHE_HOME` holds a relative path.
- The directory is a root shared by every program. Keep your files under a directory of your
  program's name, made with [File.mkdir](../File/mkdir.md), which also creates any parents it lacks.
- Use it for what can be made again, such as downloads and computed results. Settings belong in
  [Os.configDir](configDir.md).

## Basic examples

Get the user's cache directory:

<!-- test: check-only the cache directory depends on the host -->

```ant
println(Os.cacheDir())
```

On macOS this prints `Ok{ value: "/Users/ann/Library/Caches" }`, and on Linux, with
`$XDG_CACHE_HOME` unset, `Ok{ value: "/home/ann/.cache" }`.

Unwrap it with `try`, and name the directory of your program under it with
[Path.join](../Path/join.md):

<!-- test: check-only the cache directory depends on the host -->

```ant
let dir = Path.join([try Os.cacheDir(), "wordcount"])
println(dir)
```

## Applications

Keep a computed result between runs. The first call counts the words and saves the count; the next
call reads it back:

<!-- test: check-only writes to the user's cache directory -->

```ant
func cacheFile(name: string): Result[string, Error] {
  let dir = Path.join([try Os.cacheDir(), "wordcount"])
  try File.mkdir(dir)
  Result.ok(Path.join([dir, name]))
}

func countWords(path: string): Result[int, Error] {
  let cached = try cacheFile("${Path.base(path)}.count")
  if File.exists(cached) {
    let saved = (try File.read(cached)).toInt()
    if saved != null { return Result.ok(saved) }
  }
  let words = (try File.read(path)).fields().length()
  try File.write(cached, "${words}")
  Result.ok(words)
}

try File.write("notes.txt", "buy milk and bread")
println(try countWords("notes.txt"))
println(try countWords("notes.txt"))
```

Let one run of a script at a time do its work, with a lock file at a fixed path. The cache directory
is the place for it, not the shared temporary directory, where someone else can take the name first:

<!-- test: check-only writes to the user's cache directory -->

```ant
let lock = try File.tryLock(Path.join([try Os.cacheDir(), "backup.lock"]))
if lock == null {
  eprintln("backup: already running")
  Os.exit(1)
}
defer { File.close(lock) }
println("backing up")
```

Clear a program's cache. Everything in it can be made again, so removing it is safe:

<!-- test: check-only removes a directory in the user's cache directory -->

```ant
let dir = Path.join([try Os.cacheDir(), "wordcount"])
try File.removeAll(dir)
println("cleared ${dir}")
```

## Properties and relations

The cache directory follows the platform's rule: `~/Library/Caches` on macOS, `%LocalAppData%` on
Windows, and `$XDG_CACHE_HOME` or `~/.cache` elsewhere:

```ant
let dir = try Os.cacheDir()
let home = try Os.home()
let expected = if Os.platform() == "darwin" {
  Path.join([home, "Library", "Caches"])
} else if Os.platform() == "windows" {
  Os.env("LocalAppData")
} else {
  Os.env("XDG_CACHE_HOME") ?? Path.join([home, ".cache"])
}
assert(dir == expected)
```

## Possible issues

A program started with a bare environment, as some services and containers are, may have no `$HOME`,
and then `Os.cacheDir` is `Err`. A `try` at the top of a script ends it. When the cache only saves
time, check the result and go on without it:

```ant
let root = Os.cacheDir()
let cache = if root is Ok {
  Path.join([root.value, "wordcount"])
} else {
  eprintln("wordcount: running without a cache: ${root.error.message}")
  null
}
println(cache ?? "no cache")
```

## See also

[Os.configDir](configDir.md) · [Os.home](home.md) · [Os.env](env.md) ·
[File.mkdir](../File/mkdir.md) · [File.tryLock](../File/tryLock.md) ·
[File.tempDir](../File/tempDir.md) · [Path.join](../Path/join.md)
