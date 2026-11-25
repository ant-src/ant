<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.chown

```ant signature
func File.chown(path: string, uid: int, gid: int, follow: bool = true): Result[void, Error]
```

File.chown sets the user and the group that own the file at path to the ids uid and gid; -1 keeps
one as it is. It follows a symbolic link, unless follow is false, when it changes the link itself.
Only root may give a file away. It is Err on Windows.

<!-- end generated -->

## Details

- `uid` and `gid` are numeric ids, as `ls -n` shows them. `File.chown` does not look up names.
- The file's owner may change its group to a group it belongs to.

## Basic examples

Give a file to the program's own group, keeping its user:

```ant
try File.write("notes.txt", "draft")
let gid = Os.gid() ?? -1
try File.chown("notes.txt", -1, gid)
let info = try File.stat("notes.txt")
println(info.gid == Os.gid())
```

```text
true
```

As root, hand a file to the user and the group with id 1000:

<!-- test: check-only only root may give a file away -->

```ant
try File.write("authorized_keys", "")
try File.chown("authorized_keys", 1000, 1000)
```

## Scope

A directory changes owner as a file does:

```ant
try File.mkdir("site")
try File.chown("site", Os.uid() ?? -1, Os.gid() ?? -1)
println((try File.stat("site")).uid == Os.uid())
```

```text
true
```

`-1` for both ids changes nothing, and is still `Err` when nothing is at the path:

```ant
try File.write("same.txt", "")
println(File.chown("same.txt", -1, -1))
println(File.chown("absent.txt", -1, -1))
```

```text
Ok{ value: void }
Err{ error: Error{ message: "chown absent.txt: no such file or directory", kind: Error.NotExist{} } }
```

## Named arguments

`follow:` decides what a symbolic link stands for. By default `File.chown` changes the file the link
points to, so a dangling link is an `Err`. With `follow: false` it changes the link itself, which
exists:

```ant
try File.symlink("gone.txt", "current")
println(File.chown("current", -1, -1))
println(File.chown("current", -1, -1, follow: false))
```

```text
Err{ error: Error{ message: "chown current: no such file or directory", kind: Error.NotExist{} } }
Ok{ value: void }
```

## Applications

Look up a user by name with `id`, then give a file to that user and the user's group:

```ant
func lookUp(flag: string, user: string): Result[int, Error] {
  let out = try Cmd.read(Cmd.new("id", [flag, user]))
  let id = out.toInt()
  if id == null { return Result.fail("id ${flag} ${user}: ${out} is not an id") }
  Result.ok(id)
}

func chownTo(path: string, user: string): Result[void, Error] {
  let uid = try lookUp("-u", user)
  let gid = try lookUp("-g", user)
  File.chown(path, uid, gid)
}

println(try lookUp("-u", "root"))
let me = try Cmd.read(Cmd.new("id", ["-un"]))
try File.write("app.log", "")
try chownTo("app.log", me)
println((try File.stat("app.log")).uid == Os.uid())
```

```text
0
true
```

As root, give a copy made with [File.copy](copy.md) the owner of the original; `File.copy` copies
the contents and the mode, not the owner:

```ant
func copyOwner(from: string, to: string): Result[void, Error] {
  let info = try File.stat(from)
  File.chown(to, info.uid ?? -1, info.gid ?? -1)
}

try File.write("app.conf", "port = 80\n")
try File.copy("app.conf", "app.conf.bak")
try copyOwner("app.conf", "app.conf.bak")
let a = try File.stat("app.conf")
let b = try File.stat("app.conf.bak")
println(a.uid == b.uid && a.gid == b.gid)
```

```text
true
```

Say which step failed with [Result.wrapErr](../Result/wrapErr.md):

```ant
let r = File.chown("deploy.log", -1, -1) | Result.wrapErr("hand the log to deploy")
println(r)
```

```text
Err{ error: Error{ message: "hand the log to deploy: chown deploy.log: no such file or directory", kind: Error.NotExist{} } }
```

## Properties and relations

`-1` keeps both ids, and [File.stat](stat.md) reads back the ids `File.chown` set:

```ant
try File.write("f.txt", "")
let before = try File.stat("f.txt")
try File.chown("f.txt", -1, -1)
let after = try File.stat("f.txt")
assert(after.uid == before.uid && after.gid == before.gid)

try File.chown("f.txt", Os.uid() ?? -1, Os.gid() ?? -1)
let mine = try File.stat("f.txt")
assert(mine.uid == Os.uid() && mine.gid == Os.gid())
```

## Possible issues

Only root may give a file to another user. Anyone else gets an `Err`, such as
`chown key.pem: operation not permitted`, and the file keeps its owner. Check [Os.uid](../Os/uid.md)
first to fail with a clearer message:

<!-- test: check-only only root may give a file away -->

```ant
if Os.uid() != 0 {
  eprintln("run this script as root")
  Os.exit(1)
}

try File.write("key.pem", "")
try File.chown("key.pem", 0, 0)
```

The ids are numbers. A user name is a compile error; look the id up as in Applications:

```ant
try File.write("x", "")
try File.chown("x", "deploy", -1) // error: cannot use string where int is expected
```

The `uid` and `gid` of a [File.Info](types/Info.md) are `int | null`, so they do not go straight
into `File.chown`. Write `info.uid ?? -1` to keep the id where there is none:

```ant
let info = try File.stat("x")
try File.chown("y", info.uid, -1) // error: cannot use int | null where int is expected
```

## See also

[File.stat](stat.md) · [File.Info](types/Info.md) · [File.chmod](chmod.md) · [File.copy](copy.md) ·
[Os.uid](../Os/uid.md) · [Os.gid](../Os/gid.md) · [Cmd.read](../Cmd/read.md)
