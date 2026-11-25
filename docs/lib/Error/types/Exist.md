<!-- status: draft -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Error.Exist

```ant signature
type Error.Exist = {}
```

Error.Exist is the kind of an error for a path that something is at already, as a write with
replace: false finds, and for a directory that is not empty.

<!-- end generated -->

## Details

- The library sets it when something is at a path where a call needs nothing: a write, a create or a
  copy with `replace: false`, a [File.symlink](../../File/symlink.md) or a
  [File.copyAll](../../File/copyAll.md) onto something.
- It is also the kind of a directory that is not empty, so [File.remove](../../File/remove.md) of a
  directory with entries gives it.
- Anything at the path counts: a file, a directory, a link.
- `Error.Exist` is an empty record, so the value is `Error.Exist{}`.

## Basic examples

Write a file only when nothing is there, and say so when something is:

```ant
try File.write("report.txt", "monday")
let saved = File.write("report.txt", "tuesday", replace: false)
if saved is Err && saved.error.kind is Error.Exist {
  println("report.txt is already there")
}
println(try File.read("report.txt"))
```

```text
report.txt is already there
monday
```

The printed error carries the kind:

```ant
try File.write("report.txt", "monday")
println(File.write("report.txt", "tuesday", replace: false))
```

```text
Err{ error: Error{ message: "open report.txt: file exists", kind: Error.Exist{} } }
```

## Scope

Every call that refuses what is there gives it:

```ant
try File.write("page.html", "<p>")
try File.mkdir("site")
try File.write("site/index.html", "<p>")

println(File.create("page.html", replace: false) is Err)
println(File.copy("site/index.html", "page.html", replace: false))
println(File.symlink("site", "page.html"))
println(File.copyAll("site", "page.html"))
```

```text
true
Err{ error: Error{ message: "open page.html: file exists", kind: Error.Exist{} } }
Err{ error: Error{ message: "symlink site page.html: file exists", kind: Error.Exist{} } }
Err{ error: Error{ message: "copy site to page.html: page.html exists", kind: Error.Exist{} } }
```

Removing a directory that has entries gives it too:

```ant
try File.mkdir("logs")
try File.write("logs/app.log", "started")
println(File.remove("logs"))
```

```text
Err{ error: Error{ message: "remove logs: directory not empty", kind: Error.Exist{} } }
```

## Applications

Save under the first free name. Only `Error.Exist` moves on to the next name; any other failure ends
the search:

```ant
func saveNew(stem: string, text: string): Result[string, Error] {
  for n in Iter.range(0, 100) {
    let path = if n == 0 { "${stem}.txt" } else { "${stem}-${n}.txt" }
    let saved = File.write(path, text, replace: false)
    if saved is Ok {
      return Result.ok(path)
    }
    if !(saved.error.kind is Error.Exist) {
      return Result.err(saved.error)
    }
  }
  Result.fail("${stem}: no free name")
}

println(try saveNew("summary", "first"))
println(try saveNew("summary", "second"))
println(saveNew("missing/summary", "third"))
```

```text
summary.txt
summary-1.txt
Err{ error: Error{ message: "open missing/summary.txt: no such file or directory", kind: Error.NotExist{} } }
```

Take a lock file, and tell a job that already runs from a failure to take the lock:

```ant
func start(job: string): Result[string, Error] {
  let taken = File.write("${job}.lock", "running", replace: false)
  if taken is Err && taken.error.kind is Error.Exist {
    return Result.ok("${job} already runs")
  }
  try taken
  Result.ok("${job} started")
}

println(try start("backup"))
println(try start("backup"))
```

```text
backup started
backup already runs
```

## Properties and relations

[File.mkdir](../../File/mkdir.md) takes a directory that is already there as done, so it gives no
`Error.Exist`:

```ant
try File.mkdir("out")
assert(File.mkdir("out").isOk())
```

## Possible issues

The name says more than the kind does: for [File.remove](../../File/remove.md) it means the
directory is not empty. Read the message, or use [File.removeAll](../../File/removeAll.md):

```ant
try File.mkdir("cache")
try File.write("cache/a.bin", "x")
let removed = File.remove("cache")
if removed is Err && removed.error.kind is Error.Exist {
  println(removed.error.message)
}
println(File.removeAll("cache"))
```

```text
remove cache: directory not empty
Ok{ value: void }
```

Without `replace: false` a write replaces what is there, and nothing is an error:

```ant
try File.write("report.txt", "monday")
println(File.write("report.txt", "tuesday"))
```

```text
Ok{ value: void }
```

## See also

[Error.Kind](Kind.md) · [Error.NotExist](NotExist.md) · [Error.Permission](Permission.md) ·
[Error](../index.md) · [File.write](../../File/write.md) · [File.create](../../File/create.md) ·
[File.copy](../../File/copy.md) · [File.remove](../../File/remove.md)
