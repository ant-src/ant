<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.readLink

```ant signature
func File.readLink(path: string): Result[string, Error]
```

File.readLink returns the target of the symbolic link at path, as it is stored.

<!-- end generated -->

## Details

- The target comes back as text, exactly as [File.symlink](symlink.md) stored it: a relative target
  stays relative to the link's directory, and nothing is cleaned or resolved.
- `File.readLink` reads one link. When the target is itself a link, it is not followed; use
  [File.realPath](realPath.md) for the file at the end of the chain.
- A dangling link still has a target, so `File.readLink` is `Ok` for it.
- `File.readLink` is `Err` when nothing is at `path`, or when what is there is not a symbolic link.

## Basic examples

Read the target of a link:

```ant
try File.write("notes.txt", "buy milk")
try File.symlink("notes.txt", "current")
println(File.readLink("current"))
```

```text
Ok{ value: "notes.txt" }
```

A relative target is stored relative to the link's directory, and comes back that way:

```ant
try File.mkdir("site")
try File.write("site/index.html", "<h1>Home</h1>")
try File.mkdir("public")
try File.symlink("../site/index.html", "public/home.html")
println(File.readLink("public/home.html"))
println(File.read("public/home.html"))
```

```text
Ok{ value: "../site/index.html" }
Ok{ value: "<h1>Home</h1>" }
```

## Scope

An absolute target comes back absolute:

```ant
try File.symlink("/etc/hosts", "hosts")
println(File.readLink("hosts"))
```

```text
Ok{ value: "/etc/hosts" }
```

The text is not cleaned: a trailing slash, `.` and `..` stay as they were written. Unicode names
come back unchanged:

```ant
try File.symlink("./drafts//old/../new/", "latest")
try File.symlink("café ☕.txt", "menu")
println(File.readLink("latest"))
println(File.readLink("menu"))
```

```text
Ok{ value: "./drafts//old/../new/" }
Ok{ value: "café ☕.txt" }
```

A link whose target does not exist still has a target:

```ant
try File.symlink("gone.txt", "broken")
println(File.readLink("broken"))
println(File.exists("broken"))
```

```text
Ok{ value: "gone.txt" }
false
```

A link to a link gives the next link, not the file at the end. [File.realPath](realPath.md) follows
the whole chain:

```ant
try File.write("v2.txt", "second")
try File.symlink("v2.txt", "stable")
try File.symlink("stable", "default")
println(File.readLink("default"))
println(File.realPath("default").map({ Path.base(x) }))
```

```text
Ok{ value: "stable" }
Ok{ value: "v2.txt" }
```

A regular file, a directory, and a missing path are each an `Err`:

```ant
try File.write("plain.txt", "text")
try File.mkdir("folder")
println(File.readLink("plain.txt"))
println(File.readLink("folder"))
println(File.readLink("nowhere"))
```

```text
Err{ error: Error{ message: "readlink plain.txt: invalid argument" } }
Err{ error: Error{ message: "readlink folder: invalid argument" } }
Err{ error: Error{ message: "readlink nowhere: no such file or directory" } }
```

## Applications

Report which release a `live` link points at, as deploy tools do:

```ant
try File.mkdir("releases")
for name in ["v1.4.0", "v1.5.0"] {
  try File.mkdir("releases/${name}")
}
try File.symlink("releases/v1.5.0", "live")

let target = try File.readLink("live")
println("live release: ${Path.base(target)}")
```

```text
live release: v1.5.0
```

Find the dangling links in a directory: [File.stat](stat.md) with `follow: false` picks the links,
and `File.exists` follows each one:

```ant
try File.mkdir("bin")
try File.write("bin/tool-1.0", "")
try File.symlink("tool-1.0", "bin/tool")
try File.symlink("helper-0.9", "bin/helper")

for name in try File.list("bin") {
  let path = Path.join(["bin", name])
  let info = try File.stat(path, follow: false)
  if !info.isSymlink || File.exists(path) { continue }
  let target = try File.readLink(path)
  println("${path} -> ${target} (missing)")
}
```

```text
bin/helper -> helper-0.9 (missing)
```

Turn a relative target into a path from the working directory by joining it to the link's directory:

```ant
func linkTarget(link: string): Result[string, Error] {
  let target = try File.readLink(link)
  Result.ok(if Path.isAbs(target) { target } else { Path.join([Path.dir(link), target]) })
}

try File.mkdir("docs")
try File.write("docs/guide.md", "# Guide")
try File.mkdir("docs/old")
try File.symlink("../guide.md", "docs/old/guide.md")
println(linkTarget("docs/old/guide.md"))
```

```text
Ok{ value: "docs/guide.md" }
```

Say which link failed with [Result.wrapErr](../Result/wrapErr.md):

```ant
let r = File.readLink("config/active") | Result.wrapErr("find the active config")
println(r)
```

```text
Err{ error: Error{ message: "find the active config: readlink config/active: no such file or directory" } }
```

## Properties and relations

`File.readLink` gives back the target [File.symlink](symlink.md) was given, byte for byte:

```ant
let pairs = [
  ["a.txt", "l1"],
  ["../up/b.txt", "l2"],
  ["/abs/c", "l3"],
  ["dir/", "l4"],
  ["./x//y", "l5"],
  ["ünï cødé", "l6"],
]
for [target, link] in pairs {
  try File.symlink(target, link)
  assert((try File.readLink(link)) == target)
}
```

## Possible issues

A relative target is relative to the link's directory, not to the working directory. Reading the
target as a path from the working directory misses the file:

```ant
try File.mkdir("conf")
try File.write("conf/app.toml", "port = 8080")
try File.mkdir("conf/active")
try File.symlink("../app.toml", "conf/active/app.toml")

let target = try File.readLink("conf/active/app.toml")
println(File.read(target))
println(File.read("conf/active/app.toml"))
```

```text
Err{ error: Error{ message: "open ../app.toml: no such file or directory" } }
Ok{ value: "port = 8080" }
```

To read what a link leads to, read the link itself: [File.read](read.md) follows it.

## See also

[File.symlink](symlink.md) · [File.realPath](realPath.md) · [File.stat](stat.md) ·
[File.exists](exists.md) · [File.remove](remove.md) · [Path.join](../Path/join.md) ·
[Path.dir](../Path/dir.md)
