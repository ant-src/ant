<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.symlink

```ant signature
func File.symlink(target: string, link: string): Result[void, Error]
```

File.symlink makes a symbolic link at link whose target is target, in the order of ln -s. A relative
target is relative to the link's directory. It is Err when anything is at link.

<!-- end generated -->

## Details

- `target` is stored as text and resolved only when the link is used, so nothing needs to exist at
  it and a link may dangle from the start.
- `File.symlink` never replaces anything at `link`: a file, a directory, or another link, even a
  dangling one, makes it `Err`.
- On Windows it makes a directory link when `target` is an existing directory and a file link
  otherwise, and it is `Err` without Developer Mode or administrator rights.

## Basic examples

Make a link to a file, then read the file through it:

```ant
try File.write("notes.txt", "buy milk")
try File.symlink("notes.txt", "current")
println(File.read("current"))
```

```text
Ok{ value: "buy milk" }
```

A link in another directory names its target relative to that directory:

```ant
try File.mkdir("site")
try File.write("site/index.html", "<h1>Home</h1>")
try File.mkdir("public")
try File.symlink("../site/index.html", "public/home.html")
println(File.read("public/home.html"))
```

```text
Ok{ value: "<h1>Home</h1>" }
```

## Scope

An absolute target works from anywhere:

```ant
let dir = try File.realPath(".")
try File.write("config.toml", "port = 8080")
try File.symlink(Path.join([dir, "config.toml"]), "app.toml")
println(Path.isAbs(try File.readLink("app.toml")))
println(File.read("app.toml"))
```

```text
true
Ok{ value: "port = 8080" }
```

A link to a directory reaches the files inside it:

```ant
try File.mkdir("assets")
try File.write("assets/logo.svg", "<svg/>")
try File.symlink("assets", "static")
println(File.read("static/logo.svg"))
```

```text
Ok{ value: "<svg/>" }
```

The target need not exist. The link dangles until something appears there:

```ant
try File.symlink("later.txt", "soon")
println(File.exists("soon"))
try File.write("later.txt", "here now")
println(File.read("soon"))
```

```text
false
Ok{ value: "here now" }
```

A link may point at another link. Reading follows the whole chain:

```ant
try File.write("v2.txt", "second")
try File.symlink("v2.txt", "stable")
try File.symlink("stable", "default")
println(File.readLink("default"))
println(File.read("default"))
```

```text
Ok{ value: "stable" }
Ok{ value: "second" }
```

## Applications

Switch a `live` link to a new release without a moment when it is missing: make the new link under a
temporary name, then [File.rename](rename.md) it over the old one. `File.rename` acts on the link,
not on its target:

```ant
try File.mkdir("releases")
for name in ["v1.4.0", "v1.5.0"] {
  try File.mkdir("releases/${name}")
  try File.write("releases/${name}/VERSION", name)
}
try File.symlink("releases/v1.4.0", "live")

try File.symlink("releases/v1.5.0", "live.next")
try File.rename("live.next", "live")
println(File.read("live/VERSION"))
```

```text
Ok{ value: "v1.5.0" }
```

Link every script of a directory into `bin`, as an installer does:

```ant
try File.mkdir("tools")
for name in ["fmt.sh", "lint.sh"] {
  try File.write("tools/${name}", "echo ${name}\n")
}
try File.mkdir("bin")

for name in try File.list("tools") {
  let command = name.trimSuffix(Path.ext(name))
  try File.symlink("../tools/${name}", "bin/${command}")
  println("bin/${command} -> ${try File.readLink("bin/${command}")}")
}
```

```text
bin/fmt -> ../tools/fmt.sh
bin/lint -> ../tools/lint.sh
```

Say which link failed with [Result.wrapErr](../Result/wrapErr.md):

```ant
let r = File.symlink("app.log", "logs/latest") | Result.wrapErr("link the latest log")
println(r)
```

```text
Err{ error: Error{ message: "link the latest log: symlink app.log logs/latest: no such file or directory" } }
```

## Properties and relations

[File.readLink](readLink.md) gives back the target as written, and [File.stat](stat.md) with
`follow: false` sees a link:

```ant
for target in ["a.txt", "../up/b.txt", "/abs/c", "dir/", "ünï cødé"] {
  let link = "link-${target.length()}"
  try File.symlink(target, link)
  assert((try File.readLink(link)) == target)
  assert((try File.stat(link, follow: false)).isSymlink)
}
```

[File.remove](remove.md) removes the link and leaves its target alone:

```ant
try File.write("keep.txt", "safe")
try File.symlink("keep.txt", "alias")
try File.remove("alias")
assert(!File.exists("alias"))
assert(File.exists("keep.txt"))
```

## Possible issues

The target comes first, as in `ln -s`. Swapped arguments try to make the link where the file already
is:

```ant
try File.write("report.pdf", "%PDF")
println(File.symlink("latest.pdf", "report.pdf"))
```

```text
Err{ error: Error{ message: "symlink latest.pdf report.pdf: file exists" } }
```

A relative target is read from the link's directory, not from the working directory. A target
written as a path from the working directory leaves the link dangling:

```ant
try File.mkdir("data")
try File.write("data/users.csv", "id,name")
try File.mkdir("out")
try File.symlink("data/users.csv", "out/users.csv")
println(File.read("out/users.csv"))
```

```text
Err{ error: Error{ message: "open out/users.csv: no such file or directory" } }
```

`File.symlink` does not replace a link. Remove the old one first, or use the rename above when the
link must never be missing:

```ant
try File.symlink("v1", "previous")
println(File.symlink("v2", "previous"))
try File.remove("previous")
try File.symlink("v2", "previous")
println(File.readLink("previous"))
```

```text
Err{ error: Error{ message: "symlink v2 previous: file exists" } }
Ok{ value: "v2" }
```

## See also

[File.readLink](readLink.md) · [File.realPath](realPath.md) · [File.stat](stat.md) ·
[File.rename](rename.md) · [File.remove](remove.md) · [File.exists](exists.md) ·
[Path.join](../Path/join.md)
