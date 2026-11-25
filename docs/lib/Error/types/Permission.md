<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Error.Permission

```ant signature
type Error.Permission = {}
```

Error.Permission is the kind of an error for an operation the OS denied.

<!-- end generated -->

## Details

- The library sets it when the operating system denies an operation: a read of a file you may not
  read, a write where you may not write, a [File.chown](../../File/chown.md) to another user.
- A path in a directory you may not search is `Error.Permission`, not [Error.NotExist](NotExist.md),
  although [File.exists](../../File/exists.md) is `false` for both.
- Run as root, most of these calls succeed, so the examples here show an ordinary user.
- `Error.Permission` is an empty record, so the value is `Error.Permission{}`.

## Basic examples

Say what to do about a denied operation, and report any other failure as it is:

<!-- test: check-only the result depends on the user -->

```ant
try File.write("server.key", "")
let given = File.chown("server.key", 0, 0)
if given is Err && given.error.kind is Error.Permission {
  eprintln("run this script as root")
  Os.exit(1)
}
try given
```

Run by an ordinary user, the script writes `run this script as root` and exits with status 1.

The printed error carries the kind:

<!-- test: check-only the result depends on the user -->

```ant
try File.write("server.key", "")
println(File.chown("server.key", 0, 0))
```

It prints
`Err{ error: Error{ message: "chown server.key: operation not permitted", kind: Error.Permission{} } }`.

## Scope

A file whose mode lets nobody read it:

<!-- test: check-only the result depends on the user -->

```ant
try File.write("secret.txt", "x")
try File.chmod("secret.txt", 0)
println(File.read("secret.txt"))
```

It prints
`Err{ error: Error{ message: "open secret.txt: permission denied", kind: Error.Permission{} } }`.

The kind survives the context a caller adds, and an error you build can carry it:

```ant
let denied = Error{ message: "open /etc/shadow: permission denied", kind: Error.Permission{} }
Error.wrap(denied, "read users").kind is Error.Permission // true
Error.wrap(denied, "read users").message                  // "read users: open /etc/shadow: permission denied"
```

## Applications

Read the files you may, and name the ones you may not:

<!-- test: check-only the result depends on the user -->

```ant
try File.write("public.txt", "hello")
try File.write("private.txt", "secret")
try File.chmod("private.txt", 0)

for path in ["public.txt", "private.txt"] {
  let text = File.read(path)
  if text is Ok {
    println("${path}: ${text.value}")
  } else if text.error.kind is Error.Permission {
    println("${path}: skipped, not readable")
  } else {
    println("${path}: ${text.error.message}")
  }
}
```

It prints `public.txt: hello` and `private.txt: skipped, not readable`.

## Possible issues

[File.exists](../../File/exists.md) is `false` for a path it may not reach, as for a missing one.
The kind of [File.stat](../../File/stat.md)'s `Err` tells them apart:

<!-- test: check-only the result depends on the user -->

```ant
try File.mkdir("vault")
try File.write("vault/key.pem", "x")
try File.chmod("vault", 0)

println(File.exists("vault/key.pem"))
let info = File.stat("vault/key.pem")
if info is Err {
  println(info.error.kind)
}
```

It prints `false` and then `Error.Permission{}`.

Do not test the user instead of the kind: modes and owners decide what the OS allows, not the user
alone. Try the operation and test the kind of its `Err`.

## See also

[Error.Kind](Kind.md) · [Error.NotExist](NotExist.md) · [Error.Exist](Exist.md) ·
[Error](../index.md) · [File.chown](../../File/chown.md) · [File.chmod](../../File/chmod.md) ·
[Os.uid](../../Os/uid.md) · [File.exists](../../File/exists.md)
