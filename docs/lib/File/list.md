<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.list

```ant signature
func File.list(dir: string): Result[Array[string], Error]
```

File.list returns the names of the entries directly in dir, in lexical order, the order in which
strings compare, so "10.txt" comes before "9.txt". It follows a link at dir. Unlike
[`File.walk`](walk.md) it does not descend, and it returns names rather than paths.

<!-- end generated -->

## Details

- The names are bare: to read or stat an entry, join it to `dir` with
  [`Path.join`](../Path/join.md).
- Subdirectories, hidden names that begin with a dot, and symbolic links are listed like files.
- An empty directory gives `Ok` with an empty array. A missing `dir`, or a `dir` that is a file, is
  `Err`.

## Basic examples

List the entries of a directory:

<!-- test: file notes/todo.txt
buy milk
-->
<!-- test: file notes/ideas.md
# Ideas
-->
<!-- test: file notes/archive/2025.md
# 2025
-->

```ant
println(File.list("notes"))
```

```text
Ok{ value: ["archive", "ideas.md", "todo.txt"] }
```

Unwrap the names with `try` and loop over them. In a script, `try` ends the script on an `Err`:

<!-- test: file notes/todo.txt
buy milk
-->

```ant
for name in try File.list("notes") {
  println(name)
}
```

```text
archive
ideas.md
todo.txt
```

## Scope

An empty directory gives an empty array:

```ant
try File.mkdir("empty")
println(File.list("empty"))
```

```text
Ok{ value: [] }
```

Names that begin with a dot are listed too:

<!-- test: file config/.env
PORT=8080
-->
<!-- test: file config/app.json
{}
-->

```ant
println(File.list("config"))
```

```text
Ok{ value: [".env", "app.json"] }
```

The order compares names as strings do, code point by code point: uppercase letters before
lowercase, `10` before `9`, and `Ä` after every ASCII letter:

<!-- test: file mixed/apple
-->
<!-- test: file mixed/Zeta
-->
<!-- test: file mixed/Äpfel
-->
<!-- test: file mixed/9.txt
-->
<!-- test: file mixed/10.txt
-->

```ant
println(File.list("mixed"))
```

```text
Ok{ value: ["10.txt", "9.txt", "Zeta", "apple", "Äpfel"] }
```

A symbolic link is an entry of its own, listed by its name:

<!-- test: file links/target.txt
hello
-->

```ant
try File.symlink("target.txt", "links/latest")
println(File.list("links"))
```

```text
Ok{ value: ["latest", "target.txt"] }
```

## Applications

Keep the names with one extension:

<!-- test: file posts/hello.md
# Hello
-->
<!-- test: file posts/draft.txt
draft
-->
<!-- test: file posts/world.md
# World
-->

```ant
func markdownFiles(dir: string): Result[Array[string], Error] {
  let names = try File.list(dir)
  Result.ok(names.filter({ x.endsWith(".md") }))
}

println(markdownFiles("posts"))
```

```text
Ok{ value: ["hello.md", "world.md"] }
```

Report the size of each file in a directory, skipping subdirectories:

<!-- test: file notes/todo.txt
buy milk
-->
<!-- test: file notes/ideas.md
# Ideas
-->
<!-- test: file notes/archive/2025.md
# 2025
-->

```ant
func fileSizes(dir: string): Result[Map[string, int], Error] {
  let sizes: Map[string, int] = Map{}
  for name in try File.list(dir) {
    let info = try File.stat(Path.join([dir, name]))
    if !info.isDir {
      sizes[name] = info.size
    }
  }
  Result.ok(sizes)
}

println(fileSizes("notes"))
```

```text
Ok{ value: Map{ "ideas.md": 8, "todo.txt": 9 } }
```

Refuse to work in a directory that is not empty:

<!-- test: file build/old.js
-->

```ant
func checkEmpty(dir: string): Result[void, Error] {
  let names = try File.list(dir)
  if !names.isEmpty() {
    return Result.fail("${dir} is not empty: ${names | String.join(", ")}")
  }
  Result.ok(void)
}

println(checkEmpty("build"))
```

```text
Err{ error: Error{ message: "build is not empty: old.js" } }
```

## Properties and relations

The names come sorted, so sorting them again changes nothing:

<!-- test: file mixed/apple
-->
<!-- test: file mixed/Zeta
-->

```ant
let names = try File.list("mixed")
assert(names == names.sort())
```

`File.list` gives what [`File.glob`](glob.md) gives for `*` with hidden names included:

<!-- test: file config/.env
PORT=8080
-->
<!-- test: file config/app.json
{}
-->

```ant
assert(File.list("config") == File.glob("*", dir: "config", hidden: true))
```

In a directory that holds only files, `File.list` and [`File.walk`](walk.md) agree. Once there is a
subdirectory, `File.list` names it and `File.walk` lists the files inside it instead:

<!-- test: file notes/todo.txt
buy milk
-->
<!-- test: file notes/ideas.md
# Ideas
-->
<!-- test: file notes/archive/2025.md
# 2025
-->

```ant
println(File.list("notes"))
println(File.walk("notes"))
```

```text
Ok{ value: ["archive", "ideas.md", "todo.txt"] }
Ok{ value: ["archive/2025.md", "ideas.md", "todo.txt"] }
```

## Possible issues

The names are relative to `dir`, not to the working directory. Passing one straight to `File.read`
looks for it in the wrong place:

<!-- test: file inbox/todo.txt
buy milk
-->

```ant
let names = try File.list("inbox")
println(File.read(names[0]))
println(File.read(Path.join(["inbox", names[0]])))
```

```text
Err{ error: Error{ message: "open todo.txt: no such file or directory" } }
Ok{ value: "buy milk\n" }
```

A directory that does not exist, and a file in place of a directory, are errors:

<!-- test: file notes/todo.txt
buy milk
-->

```ant
println(File.list("missing"))
println(File.list("notes/todo.txt"))
```

```text
Err{ error: Error{ message: "open missing: no such file or directory" } }
Err{ error: Error{ message: "open notes/todo.txt: not a directory" } }
```

## See also

[File.walk](walk.md), [File.glob](glob.md), [File.stat](stat.md), [File.exists](exists.md),
[File.mkdir](mkdir.md), [Path.join](../Path/join.md)
