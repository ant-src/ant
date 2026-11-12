<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.platform

```ant signature
func Os.platform(): string
```

Os.platform returns the host's operating system by Go's name for it, GOOS: "darwin", "linux",
"windows", and so on.

<!-- end generated -->

## Details

- The name is Go's, always lower case: `"darwin"` for macOS, never `"macos"` or `"Darwin"` as
  `uname` prints it.

## Basic examples

Print the operating system of the host. On a Mac it prints `darwin`:

<!-- test: check-only the output depends on the host -->

```ant
println(Os.platform())
```

Compare it with a name to choose a value by platform:

<!-- test: check-only the output depends on the host -->

```ant
let opener = if Os.platform() == "darwin" { "open" } else { "xdg-open" }
println(opener)
```

## Scope

Look the name up in a map to show it to a person, keeping any other name as it is:

<!-- test: check-only the output depends on the host -->

```ant
let names = Map{ "darwin": "macOS", "linux": "Linux", "windows": "Windows" }
let os = Os.platform()
println(names[os] ?? os)
```

Join it with [Os.arch](arch.md) to name the whole target, such as `darwin-arm64` or `linux-amd64`:

<!-- test: check-only the output depends on the host -->

```ant
println("${Os.platform()}-${Os.arch()}")
```

Pass it as a function value:

```ant
let current: func(): string = Os.platform
assert(current() == Os.platform())
```

## Applications

Name the executable a build produces, which carries `.exe` on Windows:

<!-- test: check-only the output depends on the host -->

```ant
let exe = if Os.platform() == "windows" { "tool.exe" } else { "tool" }
println(exe)
```

Pick the command that opens a file in its default application, with an `if` expression that chooses
among the platforms:

<!-- test: check-only the output depends on the host -->

```ant
let os = Os.platform()
let opener = if os == "darwin" { "open" } else if os == "windows" { "explorer" } else { "xdg-open" }
println(Cmd.string(Cmd.new(opener, ["report.html"])))
```

Split `$PATH` by the separator the platform uses, `;` on Windows and `:` elsewhere:

<!-- test: check-only the output depends on the host -->

```ant
let sep = if Os.platform() == "windows" { ";" } else { ":" }
let dirs = (Os.env("PATH") ?? "").split(sep)
println("${dirs.length()} directories on the PATH")
```

## Properties and relations

The name is never empty and is lower case:

```ant
let os = Os.platform()
assert(!os.isEmpty())
assert(os == os.toLower())
```

[Os.uid](uid.md) is `null` exactly on Windows:

```ant
assert((Os.uid() == null) == (Os.platform() == "windows"))
```

## Possible issues

A comparison with the name a person would write, or the one `uname` prints, is never true:

```ant
assert(Os.platform() != "macos")
assert(Os.platform() != "Darwin")
```

## See also

[Os.arch](arch.md) · [Os.uid](uid.md) · [Os.env](env.md) · [Cmd.new](../Cmd/new.md) · [Os](index.md)
