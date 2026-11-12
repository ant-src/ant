<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.home

```ant signature
func Os.home(): Result[string, Error]
```

Os.home returns the user's home directory: $HOME on Unix, %USERPROFILE% on Windows.

<!-- end generated -->

## Details

- `Os.home` is `Err` when `$HOME` (`%USERPROFILE%` on Windows) is unset or empty.
- A path is used as written: `~` and `$HOME` in a path are not expanded, so a path under the home
  directory is `Path.join([try Os.home(), ...])`.
- The home directory is shared by every program. Keep configuration under
  [Os.configDir](configDir.md) and files that can be made again under [Os.cacheDir](cacheDir.md), in
  a directory named after your program.

## Basic examples

Print the home directory:

<!-- test: check-only the home directory differs between machines -->

```ant
let home = try Os.home()
println(home)
```

The home directory is an absolute path:

```ant
let home = try Os.home()
println(Path.isAbs(home))
```

```text
true
```

## Applications

Find a file that a tool keeps in the home directory by convention, such as the SSH client
configuration. [Result.wrapErr](../Result/wrapErr.md) says which step failed when `$HOME` is unset:

```ant
func sshConfigPath(): Result[string, Error] {
  let home = try (Os.home() | Result.wrapErr("locate the SSH configuration"))
  Result.ok(Path.join([home, ".ssh", "config"]))
}

let home = try Os.home()
let path = try sshConfigPath()
println(try Path.rel(home, path))
```

```text
.ssh/config
```

Expand a leading `~` in a path the user typed, as a shell does:

```ant
func expandHome(path: string): Result[string, Error] {
  if path != "~" && !path.startsWith("~/") { return Result.ok(path) }
  let home = try Os.home()
  Result.ok(Path.join([home, path.trimPrefix("~")]))
}

let home = try Os.home()
let config = try expandHome("~/.ssh/config")
println(try Path.rel(home, config))
println(try expandHome("/etc/hosts"))
println(try expandHome("~") == home)
```

```text
.ssh/config
/etc/hosts
true
```

Shorten a path under the home directory to `~` for display:

```ant
func tildify(path: string, home: string): string {
  if path == home { return "~" }
  if path.startsWith("${home}/") { return "~${path.trimPrefix(home)}" }
  path
}

let home = try Os.home()
println(tildify(Path.join([home, "src", "app"]), home))
println(tildify(home, home))
println(tildify("/etc/hosts", home))
```

```text
~/src/app
~
/etc/hosts
```

## Properties and relations

On Unix, `Os.home` is the value of `$HOME`:

```ant
if Os.platform() != "windows" {
  let home = try Os.home()
  assert(Os.env("HOME") == home)
}
```

## Possible issues

`~` is shell syntax. `File.read("~/notes.txt")` looks for a directory named `~` in the working
directory. Expand it with `Os.home` first, as `expandHome` does under Applications:

```ant
println(File.read("~/notes.txt"))
```

```text
Err{ error: Error{ message: "open ~/notes.txt: no such file or directory" } }
```

A program started without `$HOME` gets `Err` from `Os.home`. Here a child Ant program runs with an
`env` that holds only `PATH`, which [Cmd.new](../Cmd/new.md) gives it in place of the parent's:

```ant
let dir = try File.tempDir("home-*")
defer { File.removeAll(dir) }
let script = Path.join([dir, "main.ant"])
try File.write(script, "println(Os.home())")
let child = Cmd.new("ant", ["run", script], env: Map{ "PATH": Os.env("PATH") ?? "" })
print(try Cmd.read(child))
```

```text
Err{ error: Error{ message: "$HOME is not defined" } }
```

Where the program can go on without it, fall back to a directory of its own choosing with
[Result.unwrapOr](../Result/unwrapOr.md):

```ant
let home = Os.home().unwrapOr(try Os.cwd())
println(Path.isAbs(home))
```

```text
true
```

## See also

[Os.configDir](configDir.md) · [Os.cacheDir](cacheDir.md) · [Os.cwd](cwd.md) · [Os.env](env.md) ·
[Path.join](../Path/join.md) · [Path.rel](../Path/rel.md) · [File.mkdir](../File/mkdir.md)
