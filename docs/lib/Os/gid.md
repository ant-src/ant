<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.gid

```ant signature
func Os.gid(): int | null
```

Os.gid returns the effective group id, or null on Windows and WebAssembly, which have none.

<!-- end generated -->

## Details

- It is one id; the user's other groups are not part of it.

## Basic examples

Print the id of the group the program runs as:

<!-- test: check-only the id differs between users -->

```ant
println(Os.gid())
```

Print both ids of the process:

<!-- test: check-only the ids differ between users -->

```ant
println("uid ${Os.uid()}, gid ${Os.gid()}")
```

## Scope

### Handling `null`

`??` gives a value to use on Windows:

<!-- test: check-only the id differs between users -->

```ant
let gid = Os.gid() ?? -1
println("gid ${gid}")
```

A guard narrows the result to `int` for the rest of the block:

<!-- test: check-only the id differs between users -->

```ant
let gid = Os.gid()
if gid == null {
  println("no group ids on this platform")
  return
}
println("gid ${gid}, next ${gid + 1}")
```

### As a function value

```ant
let current: func(): int | null = Os.gid
assert(current() == Os.gid())
```

## Applications

### Check a file's group

[File.stat](../File/stat.md) reports the group that owns a file. On Windows both the file's group
and `Os.gid()` are null and so compare equal, so the guard comes first:

<!-- test: check-only the answer depends on the user and the file -->

```ant
let gid = Os.gid()
if gid == null {
  println("no group ids on this platform")
  return
}
let info = try File.stat("/etc/hosts")
println(if info.gid == gid { "in my group" } else { "group ${info.gid}" })
```

### Give a file to the program's group

[File.chown](../File/chown.md) takes `-1` to keep the owner and changes only the group. The file's
owner may give it to the program's own group:

```ant
let gid = Os.gid()
if gid == null {
  return
}
let dir = try File.tempDir("gid")
defer { File.removeAll(dir) }
let path = Path.join([dir, "shared.txt"])
try File.write(path, "")
try File.chown(path, -1, gid)
let info = try File.stat(path)
assert(info.gid == gid)
```

## Properties and relations

On Unix the id is what `id -g` prints for the same process:

```ant
let printed = try Cmd.read(Cmd.new("id", ["-g"]))
assert(printed.trim() == "${Os.gid()}")
```

## Possible issues

### The result may be `null`

The result is `int | null`, so it does not go where an `int` is expected:

```ant
let next: int = Os.gid() + 1 // error: + needs both operands of one type, got int | null and int
```

### Other groups do not count

A file whose group differs from `Os.gid()` may still belong to one of the user's other groups, so
`info.gid != Os.gid()` does not mean the program lacks the group's access. Try the operation and
handle its `Err` instead of predicting it.

## See also

[Os.uid](uid.md) · [File.stat](../File/stat.md) · [File.chown](../File/chown.md) · [Os](index.md)
