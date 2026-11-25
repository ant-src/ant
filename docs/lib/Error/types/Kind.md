<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Error.Kind

```ant signature
type Error.Kind = Error.NotExist | Error.Exist | Error.Permission
```

Error.Kind is what failed, for the failures of the OS that Go tells apart on every platform. Test it
with is, on the result of the call it is about: a kind survives [`Error.wrap`](../wrap.md).

<!-- end generated -->

## Details

- `kind` is the second field of an [Error](../index.md). The library sets it where the failure of
  the operating system is one Go tells apart, and where a function reports one of those conditions
  itself, as [File.copyAll](../../File/copyAll.md) onto something it refuses. Most other errors have
  `kind: null`.
- The three members are records with no fields, so a value is written with braces:
  `Error.NotExist{}`.
- `r.error.kind` does not narrow. To cover every kind in one chain, bind the kind to a `let` first
  and add an arm for `null`.

## Basic examples

Read a file that may be missing, and still fail on anything else:

```ant
func readOr(path: string, fallback: string): Result[string, Error] {
  let text = File.read(path)
  if text is Err && text.error.kind is Error.NotExist {
    return Result.ok(fallback)
  }
  text
}

println(readOr("settings.json", "{}"))
try File.mkdir("cache")
println(readOr("cache", "{}"))
```

```text
Ok{ value: "{}" }
Err{ error: Error{ message: "read cache: is a directory", kind: null } }
```

A printed error shows its kind, and `null` when it has none:

```ant
println(File.read("notes.txt"))
println(Json.parse("{"))
```

```text
Err{ error: Error{ message: "open notes.txt: no such file or directory", kind: Error.NotExist{} } }
Err{ error: Error{ message: "unexpected end of JSON input", kind: null } }
```

## Scope

`Error.NotExist` and `Error.Exist`, from calls that give them. `Error.Exist` is also the kind of a
directory that is not empty:

```ant
try File.write("report.txt", "monday")
try File.mkdir("logs")
try File.write("logs/app.log", "started")

println(File.stat("missing.txt"))
println(File.write("report.txt", "tuesday", replace: false))
println(File.remove("logs"))
```

```text
Err{ error: Error{ message: "stat missing.txt: no such file or directory", kind: Error.NotExist{} } }
Err{ error: Error{ message: "open report.txt: file exists", kind: Error.Exist{} } }
Err{ error: Error{ message: "remove logs: directory not empty", kind: Error.Exist{} } }
```

Bind the kind to a local, and a chain with an arm for each member and one for `null` covers every
case:

```ant
func describe(e: Error): string {
  let kind = e.kind
  if kind is Error.NotExist {
    "nothing is there"
  } else if kind is Error.Exist {
    "something is there already"
  } else if kind is Error.Permission {
    "the OS denied it"
  } else if kind is null {
    e.message
  }
}

describe(Error{ message: "open a.txt: not found", kind: Error.NotExist{} }) // "nothing is there"
describe(Error{ message: "disk full" })                                     // "disk full"
```

Set a kind on an `Error` you build, such as a stand-in for the file system in a test. The kind then
means what the library's means:

```ant
func fakeRead(path: string): Result[string, Error] {
  Result.err(Error{ message: "open ${path}: no such file", kind: Error.NotExist{} })
}

let r = fakeRead("a.txt")
r is Err && r.error.kind is Error.NotExist // true
```

[Error.wrap](../wrap.md) and [Result.wrapErr](../../Result/wrapErr.md) keep the kind.
[Result.fail](../../Result/fail.md) builds an error with none:

```ant
let e = Error{ message: "open app.json: no such file or directory", kind: Error.NotExist{} }
Error.wrap(e, "load config").kind // Error.NotExist{}
let failed: Result[int, Error] = Result.fail("no input")
failed // Err{ error: Error{ message: "no input", kind: null } }
```

## Applications

Create a file only when nothing is there, and tell "someone was first" from a real failure:

```ant
func claim(path: string, owner: string): Result[bool, Error] {
  let written = File.write(path, owner, replace: false)
  if written is Err && written.error.kind is Error.Exist {
    return Result.ok(false)
  }
  try written
  Result.ok(true)
}

println(claim("job.lock", "worker 1"))
println(claim("job.lock", "worker 2"))
println(claim("missing/job.lock", "worker 3"))
```

```text
Ok{ value: true }
Ok{ value: false }
Err{ error: Error{ message: "open missing/job.lock: no such file or directory", kind: Error.NotExist{} } }
```

Remove a file that may already be gone:

```ant
func forget(path: string): Result[void, Error] {
  let removed = File.remove(path)
  if removed is Err && removed.error.kind is Error.NotExist {
    return Result.ok(void)
  }
  removed
}

try File.write("session.tmp", "x")
println(forget("session.tmp"))
println(forget("session.tmp"))
```

```text
Ok{ value: void }
Ok{ value: void }
```

## Properties and relations

A kind is part of the error, so `==` compares it:

```ant
let plain = Error{ message: "open a.txt: not found" }
let missing = Error{ message: "open a.txt: not found", kind: Error.NotExist{} }
assert(plain != missing)
assert(missing.kind == Error.NotExist{})
assert(plain.kind == null)
```

`Result.mapErr` returns the error its function builds, so the kind stays only when the function
copies it:

```ant
let e = Error{ message: "open a.txt: not found", kind: Error.NotExist{} }
let r: Result[int, Error] = Result.err(e)
let plain = r | Result.mapErr({ Error{ message: x.message } })
plain // Err{ error: Error{ message: "open a.txt: not found", kind: null } }
```

## Possible issues

A kind travels with a wrapped error, so far from the call it no longer says which path it is about.
Here the site is there and its theme is missing:

```ant
func loadSite(dir: string): Result[string, Error] {
  let config = try (File.read("${dir}/site.json") | Result.wrapErr("load config"))
  let theme = try (File.read("${dir}/theme.css") | Result.wrapErr("load theme"))
  Result.ok(config + theme)
}

try File.mkdir("blog")
try File.write("blog/site.json", "{}")
let site = loadSite("blog")
if site is Err && site.error.kind is Error.NotExist {
  println("no site in blog: ${site.error.message}")
}
```

```text
no site in blog: load theme: open blog/theme.css: no such file or directory
```

Test the kind inside `loadSite`, on the result of the `File.read` it is about.

A test of `e.kind` does not narrow it, so the same chain on the field is not complete, even with an
arm for every member. Bind the kind to a `let` first, as in Scope:

```ant
func describe(e: Error): string {
  if e.kind is Error.NotExist {
    "nothing is there"
  } else if e.kind is Error.Exist {
    "something is there already"
  } else if e.kind is Error.Permission {
    "the OS denied it"
  } else if e.kind is null {
    // error: an if without else has type void when the condition is false, which cannot be used where string is expected
    e.message
  }
}
```

Most errors have no kind: a parse error, a refused connection, a timeout, a path that leads through
a file. Do not read `null` as one kind of failure:

```ant
try File.write("todo.txt", "x")
println(File.read("todo.txt/today"))
```

```text
Err{ error: Error{ message: "open todo.txt/today: not a directory", kind: null } }
```

`Json.stringify` writes a record without its name, so every kind becomes `{}` in JSON. Write
`String.from(e.kind)` to keep it in a log:

```ant
let e = Error{ message: "open a.txt: not found", kind: Error.NotExist{} }
Json.stringify(e)   // "{\"message\":\"open a.txt: not found\",\"kind\":{}}"
String.from(e.kind) // "Error.NotExist{}"
```

A program cannot add kinds. An error of your own that code branches on is a type of your own:

```ant
let e = Error{ message: "user 7 not found", kind: "not-found" }
// error: cannot use string where Error.Kind | null is expected
```

A kind is a value, so it takes `{}`:

```ant
let e = Error{ message: "open a.txt: not found", kind: Error.NotExist }
// error: Error.NotExist is a type, not a value; a type appears only as a literal head, after is, or before a dot
```

## See also

[Error.NotExist](NotExist.md) · [Error.Exist](Exist.md) · [Error.Permission](Permission.md) ·
[Error](../index.md) · [Error.wrap](../wrap.md) · [Result.wrapErr](../../Result/wrapErr.md) ·
[File.exists](../../File/exists.md) · [Errors](../../guides/errors.md)
