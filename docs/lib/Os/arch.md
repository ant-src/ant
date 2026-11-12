<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.arch

```ant signature
func Os.arch(): string
```

Os.arch returns the host's processor architecture by Go's name for it, GOARCH: "arm64", "amd64", and
so on.

<!-- end generated -->

## Details

- The name is Go's, in lower case: `"amd64"` for 64-bit x86 and `"arm64"` for 64-bit ARM.

## Basic examples

Print the architecture of the host. On an Apple silicon Mac it prints `arm64`:

<!-- test: check-only the output depends on the host -->

```ant
println(Os.arch())
```

Compare it with a name to choose a value by architecture:

<!-- test: check-only the output depends on the host -->

```ant
let libDir = if Os.arch() == "arm64" { "lib/arm64" } else { "lib/x64" }
println(libDir)
```

## Scope

Join it with [Os.platform](platform.md) to name the whole target, such as `darwin-arm64` or
`linux-amd64`:

<!-- test: check-only the output depends on the host -->

```ant
let target = "${Os.platform()}-${Os.arch()}"
println(target)
```

## Applications

Name the release archive to download for this host:

<!-- test: check-only the output depends on the host -->

```ant
let version = "1.4.0"
let asset = "tool-${version}-${Os.platform()}-${Os.arch()}.tar.gz"
println(asset)
```

Translate the name to the one Linux's `uname -m` prints, keeping any other name as it is:

<!-- test: check-only the output depends on the host -->

```ant
let arch = Os.arch()
let unameNames = Map{ "amd64": "x86_64", "arm64": "aarch64", "386": "i386" }
println(unameNames[arch] ?? arch)
```

Stop early on an architecture the script does not support:

<!-- test: check-only the outcome depends on the host -->

```ant
let arch = Os.arch()
let supported = ["amd64", "arm64"]
if !supported.contains(arch) {
  eprintln("unsupported architecture: ${arch}")
  Os.exit(1)
}
println("installing for ${arch}")
```

## Properties and relations

The name is lower case:

```ant
let arch = Os.arch()
assert(arch == arch.toLower())
```

## Possible issues

`Os.arch` uses Go's names, so `"x86_64"` and `"aarch64"` never match: compare with `"amd64"` and
`"arm64"`.

```ant
assert(Os.arch() != "x86_64")
assert(Os.arch() != "aarch64")
```

## See also

[Os.platform](platform.md) · [Os.which](which.md) · [Os](index.md)
