<!-- status: reviewed -->

# Read environment variables and standard directories

Configure a script from its environment, and keep its files where the user's other programs keep
theirs.

## Read a variable with a default

Call [Os.env](../../lib/Os/env.md). It returns `null` when the variable is not set, so give the
default with `??`:

<!-- test: env EDITOR=nano -->
<!-- test: env PAGER -->

```ant
let editor = Os.env("EDITOR") ?? "vi"
let pager = Os.env("PAGER") ?? "less"
println("${editor} ${pager}")
```

```text
nano less
```

A variable set to the empty string is `""`, not `null`. To treat a flag as on whenever it is set,
compare with `null`:

<!-- test: env VERBOSE= -->

```ant
let verbose = Os.env("VERBOSE") != null
println(verbose)
```

```text
true
```

## Parse a number

Convert the text with [String.toInt](../../lib/String/toInt.md), which returns `null` for text that
is not a number. Chain it with `?.` so that a missing and a malformed variable both fall back to the
default:

<!-- test: env WORKERS=8 -->
<!-- test: env RETRIES=many -->

```ant
let workers = Os.env("WORKERS")?.toInt() ?? 4
let retries = Os.env("RETRIES")?.toInt() ?? 3
println("${workers} workers, ${retries} retries")
```

```text
8 workers, 3 retries
```

`toInt` takes no surrounding spaces, so `" 8"` falls back to the default too; call
[String.trim](../../lib/String/trim.md) before `.toInt()` to accept them:

<!-- test: env WORKERS= 8 -->

```ant
let workers = Os.env("WORKERS")?.trim().toInt() ?? 4
println(workers)
```

```text
8
```

## Require a variable

Stop the script with an error when a variable it cannot do without is missing. `?? return` returns
the `Err` that [Result.fail](../../lib/Result/fail.md) makes from the script, which prints
`API_TOKEN is not set` to standard error and exits with status 1.
[String.length](../../lib/String/length.md) counts what it got:

<!-- test: exit 1 -->
<!-- test: env API_TOKEN -->

```ant
let token = Os.env("API_TOKEN") ?? return Result.fail("API_TOKEN is not set")
println("token has ${token.length()} characters")
```

## List variables

Call [Os.environ](../../lib/Os/environ.md) for a map of every variable. Its order is the host's, so
take the names with [Map.keys](../../lib/Map/keys.md), keep the ones you want with
[Array.filter](../../lib/Array/filter.md) and [String.startsWith](../../lib/String/startsWith.md),
and sort them with [Array.sort](../../lib/Array/sort.md):

<!-- test: env APP_PORT=8080 -->
<!-- test: env APP_HOST=localhost -->

```ant
let names = Os.environ()
  | Map.keys()
  | Array.filter({ x.startsWith("APP_") })
  | Array.sort()
println(names)
```

```text
["APP_HOST", "APP_PORT"]
```

The map is a copy: changing it changes neither the environment nor what `Os.env` returns.

## Find the standard directories

[Os.home](../../lib/Os/home.md) returns the user's home directory,
[Os.configDir](../../lib/Os/configDir.md) the directory for configuration, and
[Os.cacheDir](../../lib/Os/cacheDir.md) the directory for files the program can make again. Each
returns a `Result`, which is `Err` when the variable it rests on, such as `HOME`, is unset or empty,
or when an XDG variable holds a relative path.

Build a path under the home directory with [Path.join](../../lib/Path/join.md);
[Path.base](../../lib/Path/base.md) gives its last element:

<!-- test: env HOME=home -->

```ant
let rc = Path.join([try Os.home(), ".greetrc"])
println(Path.base(rc))
```

```text
.greetrc
```

The directories are shared by every program, and may not exist yet, so make one of your own under
them with [File.mkdir](../../lib/File/mkdir.md), which creates the parents too:

<!-- test: env HOME=home -->
<!-- test: env XDG_CONFIG_HOME= -->
<!-- test: env XDG_CACHE_HOME= -->

```ant
let config = try Os.configDir()
let cache = try Os.cacheDir()
let settings = Path.join([config, "greet"])
let downloads = Path.join([cache, "greet"])
try File.mkdir(settings)
try File.mkdir(downloads)
println(File.exists(settings) && File.exists(downloads))
```

```text
true
```

The directory names differ by platform: the configuration directory is `$XDG_CONFIG_HOME` or
`~/.config` on Unix, `~/Library/Application Support` on macOS, and `%AppData%` on Windows; the cache
directory is `$XDG_CACHE_HOME` or `~/.cache`, `~/Library/Caches`, and `%LocalAppData%`.

## Use a temporary directory

For files that last only one run, call [File.tempDir](../../lib/File/tempDir.md). It creates a new
directory in the host's temporary directory, under a name no other program has. Nothing removes it,
so remove it with [File.removeAll](../../lib/File/removeAll.md) in a `defer` right after you make
it. `Path.base` gives the name it picked:

```ant
let scratch = try File.tempDir("greet-*")
defer { File.removeAll(scratch) }
try File.write(Path.join([scratch, "draft.txt"]), "hello")
println(Path.base(scratch).startsWith("greet-"))
```

```text
true
```

Keep a file with a fixed name in your configuration or cache directory, not in the temporary one,
where another user can take the name first. [Use a temporary file or directory](../files/temp.md)
shows more.

## The whole program

This script greets a name saved in its configuration directory, which it creates with a default on
the first run. `GREET_NAME` overrides the saved name, and `GREET_TIMES` says how many times to
greet:

<!-- test: env HOME=home -->
<!-- test: env XDG_CONFIG_HOME= -->
<!-- test: env GREET_TIMES=2 -->

```ant
let config = try Os.configDir()
let dir = Path.join([config, "greet"])
try File.mkdir(dir)
let path = Path.join([dir, "name.txt"])
if !File.exists(path) {
  try File.write(path, "world\n")
}
let saved = try File.read(path)
let name = Os.env("GREET_NAME") ?? saved.trim()
let times = Os.env("GREET_TIMES")?.toInt() ?? 1
for _ in Iter.range(0, times) {
  println("Hello, ${name}")
}
```

```text
Hello, world
Hello, world
```

Besides the functions above, it uses [File.exists](../../lib/File/exists.md),
[File.write](../../lib/File/write.md), [File.read](../../lib/File/read.md),
[String.trim](../../lib/String/trim.md), and [Iter.range](../../lib/Iter/range.md).
