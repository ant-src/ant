<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.configDir

```ant signature
func Os.configDir(): Result[string, Error]
```

Os.configDir returns the directory where the user's programs keep their configuration:
$XDG\_CONFIG\_HOME or \~/.config on Unix, \~/Library/Application Support on macOS, %AppData% on
Windows. A program keeps its files under a directory of its own name there.

<!-- end generated -->

## Details

- It is `Err` when the variable it rests on is unset, such as `$HOME`, or, on Unix other than macOS,
  when `$XDG_CONFIG_HOME` holds a relative path.
- Make your program's directory with [File.mkdir](../File/mkdir.md) before you write to it. It makes
  any missing parents and is `Ok` when the directory already exists, so calling it on every run is
  safe.

## Basic examples

Find the user's configuration directory. On macOS it is `/Users/<name>/Library/Application Support`,
on Linux usually `/home/<name>/.config`:

<!-- test: check-only the directory depends on the host -->

```ant
let dir = try Os.configDir()
println(dir)
```

Without `try`, `println(Os.configDir())` shows the `Result`; on macOS it prints
`Ok{ value: "/Users/ann/Library/Application Support" }`.

Unwrap it with `try`. The path is absolute:

```ant
let dir = try Os.configDir()
println(Path.isAbs(dir))
```

```text
true
```

Build the path of a settings file under a directory of the program's own name:

<!-- test: check-only the directory depends on the host -->

```ant
let path = Path.join([try Os.configDir(), "greeter", "settings.txt"])
println(path)
```

## Applications

Read a setting, and fall back to a default when the user has not saved one. Add context to the
error, so a missing `$HOME` reads as a problem with the settings:

<!-- test: check-only the directory and its contents depend on the host -->

```ant
func settingsPath(): Result[string, Error] {
  let root = try (Os.configDir() | Result.wrapErr("find the settings"))
  Result.ok(Path.join([root, "greeter", "name.txt"]))
}

func loadName(): Result[string, Error] {
  let path = try settingsPath()
  if !File.exists(path) { return Result.ok("world") }
  let text = try File.read(path)
  Result.ok(text.trim())
}

println("Hello, ${try loadName()}")
```

Save a setting, making the program's directory first:

<!-- test: check-only it writes to the user's configuration directory -->

```ant
func saveName(name: string): Result[void, Error] {
  let dir = Path.join([try Os.configDir(), "greeter"])
  try File.mkdir(dir)
  File.write(Path.join([dir, "name.txt"]), name)
}

try saveName("Ann")
```

## Properties and relations

The directory follows from [Os.home](home.md) and the environment, as the platform names it:

```ant
let dir = try Os.configDir()
let home = try Os.home()
let expected = if Os.platform() == "darwin" {
  Path.join([home, "Library", "Application Support"])
} else if Os.platform() == "windows" {
  Os.env("AppData") ?? ""
} else {
  Os.env("XDG_CONFIG_HOME") ?? Path.join([home, ".config"])
}
assert(dir == expected)
```

## Possible issues

`Os.configDir` does not make a directory for your program. Writing under one that is not there yet
is `Err`; call [File.mkdir](../File/mkdir.md) first, as `saveName` does above:

<!-- test: check-only the directory depends on the host -->

```ant
let path = Path.join([try Os.configDir(), "greeter-new", "name.txt"])
println(File.write(path, "Ann"))
```

A program started with a bare environment, as some services and containers are, may have no `$HOME`,
and then `Os.configDir` is `Err`. A `try` at the top of a script ends it. To go on with default
settings, check the result instead:

```ant
let root = Os.configDir()
if root is Err {
  eprintln("greeter: using default settings: ${root.error.message}")
}
```

## See also

[Os.cacheDir](cacheDir.md) · [Os.home](home.md) · [Os.env](env.md) · [File.mkdir](../File/mkdir.md)
· [Path.join](../Path/join.md)
