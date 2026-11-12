<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.copyAll

```ant signature
func File.copyAll(from: string, to: string): Result[void, Error]
```

File.copyAll copies the file or the whole tree at from to to, keeping modes and copying the links
inside the tree as links. to must not exist or, for a tree, be an empty directory, which keeps its
own mode; anything else is Err before anything is copied, as is a to that lies inside from. An error
partway through leaves what was copied.

<!-- end generated -->

## Details

- `to` must not exist, or, when `from` is a directory, must be an empty directory, which becomes the
  root of the copy and keeps its own mode. Anything else at `to`, including an empty directory when
  `from` is a file, is an `Err` before anything is copied, so two trees are never merged.
- Links inside the tree are copied as links with the same target, not followed. A link at `from`
  itself is followed.
- It is an `Err` when `to`, with its links resolved, lies inside `from`, such as
  `File.copyAll("site", "site/backup")`, which would copy its own copy.
- An error partway through leaves what was copied.

## Basic examples

Copy a directory and everything below it:

<!-- test: file site/index.html
<h1>Hello</h1>
-->
<!-- test: file site/css/main.css
h1 { color: red }
-->

```ant
try File.copyAll("site", "backup")
println(try File.walk("backup"))
println(try File.read("backup/css/main.css"))
```

```text
["css/main.css", "index.html"]
h1 { color: red }

```

Copy one file, as [File.copy](copy.md) does:

```ant
try File.copyAll("site/index.html", "home.html")
println(try File.read("home.html"))
```

```text
<h1>Hello</h1>

```

## Scope

### An empty directory at `to`

An empty directory at `to` becomes the root of the copy and keeps its own mode:

```ant
try File.mkdir("public")
try File.chmod("public", 0o700)
try File.copyAll("site", "public")
println(try File.list("public"))
println((try File.stat("public")).mode == 0o700)
```

```text
["css", "index.html"]
true
```

### Modes

Every file and directory keeps its mode:

```ant
try File.mkdir("tools")
try File.write("tools/deploy.sh", "echo deploy")
try File.chmod("tools/deploy.sh", 0o755)
try File.copyAll("tools", "tools-copy")
println((try File.stat("tools-copy/deploy.sh")).mode == 0o755)
```

```text
true
```

### Links inside the tree

A link inside the tree comes out as a link with the same target:

```ant
try File.mkdir("docs")
try File.write("docs/guide.md", "# Guide")
try File.symlink("guide.md", "docs/README.md")
try File.copyAll("docs", "docs-copy")
println((try File.stat("docs-copy/README.md", follow: false)).isSymlink)
println(try File.readLink("docs-copy/README.md"))
println(try File.read("docs-copy/README.md"))
```

```text
true
guide.md
# Guide
```

### A link at `from`

A link at `from` is followed, so the copy is a real directory:

```ant
try File.symlink("site", "current")
try File.copyAll("current", "release")
println((try File.stat("release", follow: false)).isSymlink)
println(try File.walk("release"))
```

```text
false
["css/main.css", "index.html"]
```

### An empty directory at `from`

```ant
try File.mkdir("cache")
try File.copyAll("cache", "cache-copy")
println(try File.list("cache-copy"))
```

```text
[]
```

## Named arguments

Name the arguments to make the direction plain at the call:

```ant
try File.copyAll(from: "site", to: "mirror")
println(File.exists("mirror/index.html"))
```

```text
true
```

## Applications

### Start a project from a template

Copy a template tree, then fill in what differs:

<!-- test: file template/README.md
# Project
-->
<!-- test: file template/src/main.ant
println("hello")
-->

```ant
func scaffold(name: string): Result[void, Error] {
  try (File.copyAll("template", name) | Result.wrapErr("scaffold ${name}"))
  File.write(Path.join([name, "README.md"]), "# ${name}\n")
}

println(scaffold("rocket"))
println(try File.walk("rocket"))
println(scaffold("rocket"))
```

```text
Ok{ value: void }
["README.md", "src/main.ant"]
Err{ error: Error{ message: "scaffold rocket: copy template to rocket: rocket exists" } }
```

### Copy into a directory

`File.copyAll` has no "into" form. Join the destination directory and the base name of `from`:

```ant
try File.mkdir("archive")
let from = "site"
try File.copyAll(from, Path.join(["archive", Path.base(from)]))
println(try File.walk("archive"))
```

```text
["site/css/main.css", "site/index.html"]
```

### Work on a scratch copy

Copy a tree into a new [temporary directory](tempDir.md), work there, and remove it when done. The
directory is empty, so the copy goes into it, and it keeps its mode `0o700`:

```ant
let scratch = try File.tempDir("site-*")
defer { File.removeAll(scratch) }
try File.copyAll("site", scratch)
try File.write(Path.join([scratch, "index.html"]), "<h1>Draft</h1>")
println((try File.stat(scratch)).mode == 0o700)
println(try File.read("site/index.html"))
```

```text
true
<h1>Hello</h1>

```

## Properties and relations

The copy has the same files with the same contents and modes:

```ant
try File.copyAll("site", "twin")
let files = try File.walk("site")
assert(files == try File.walk("twin"))
for file in files {
  let a = Path.join(["site", file])
  let b = Path.join(["twin", file])
  assert((try File.read(a)) == (try File.read(b)))
  assert((try File.stat(a)).mode == (try File.stat(b)).mode)
}
println("same")
```

```text
same
```

On a single file, `File.copyAll` copies like `File.copy`:

```ant
try File.copy("site/index.html", "one.html")
try File.copyAll("site/index.html", "two.html")
assert((try File.read("one.html")) == (try File.read("two.html")))
println("same")
```

```text
same
```

## Possible issues

### Something at `to`

A directory with entries at `to` is an `Err`, and so is a file. Nothing is merged or replaced:

```ant
try File.mkdir("full")
try File.write("full/old.txt", "old")
println(File.copyAll("site", "full"))
println(File.copyAll("site/index.html", "home-page.html"))
println(File.copyAll("site/index.html", "home-page.html"))
try File.mkdir("pages")
println(File.copyAll("site/index.html", "pages"))
```

```text
Err{ error: Error{ message: "copy site to full: full exists" } }
Ok{ value: void }
Err{ error: Error{ message: "copy site/index.html to home-page.html: home-page.html exists" } }
Err{ error: Error{ message: "copy site/index.html to pages: pages exists" } }
```

A file copied onto an empty directory is an `Err` too, not a copy into it. Join the directory and
the base name instead, as in [Copy into a directory](#copy-into-a-directory). To replace a tree,
remove it with [File.removeAll](removeAll.md) first.

### `to` inside `from`

Copying a tree into a folder inside itself is an `Err`:

```ant
println(File.copyAll("site", "site/backup"))
```

```text
Err{ error: Error{ message: "copy site to site/backup: site/backup lies inside site" } }
```

Copy to a sibling path such as `backup` instead.

### Relative links that leave the tree

A link keeps its target text, so a relative link that points out of the tree resolves from the
copy's place, where that target may not exist:

```ant
try File.mkdir("app/config")
try File.write("shared.conf", "port=8080")
try File.symlink("../../shared.conf", "app/config/app.conf")
try File.mkdir("deploy")
try File.copyAll("app", "deploy/app")
println(File.exists("app/config/app.conf"))
println(File.exists("deploy/app/config/app.conf"))
```

```text
true
false
```

### `File.copy` on a directory

[File.copy](copy.md) copies one file and is an `Err` on a directory:

```ant
println(File.copy("site", "site-copy"))
```

```text
Err{ error: Error{ message: "copy site: is a directory; File.copyAll copies a tree" } }
```

## See also

[File.copy](copy.md), [File.removeAll](removeAll.md), [File.rename](rename.md),
[File.walk](walk.md), [File.symlink](symlink.md), [File.tempDir](tempDir.md),
[Path.join](../Path/join.md)
