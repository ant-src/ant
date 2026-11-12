<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.chmod

```ant signature
func File.chmod(path: string, mode: int): Result[void, Error]
```

File.chmod sets the mode of the file at path to mode exactly, following a symbolic link. It is Err
for a mode outside 0 ..= 0o7777. On Windows only 0o200 counts: without it the file becomes
read-only.

<!-- end generated -->

## Details

- The mode replaces the old one, bits that are clear included. To add or remove a few bits, read the
  current mode with [File.stat](stat.md) and change it with [Bits.or](../Bits/or.md) or
  [Bits.and](../Bits/and.md).

## Basic examples

Make a script executable by its owner and readable by everyone:

```ant
try File.write("deploy.sh", "echo deploying\n")
try File.chmod("deploy.sh", 0o755)
let info = try File.stat("deploy.sh")
println(String.fromInt(info.mode, base: 8))
```

```text
755
```

Make a file readable and writable by its owner only:

```ant
try File.write("notes.txt", "draft")
try File.chmod("notes.txt", 0o600)
let info = try File.stat("notes.txt")
println(String.fromInt(info.mode, base: 8))
```

```text
600
```

## Scope

A mode may include the setuid, setgid and sticky bits, `0o4000`, `0o2000` and `0o1000`. A shared
directory such as `/tmp` is `0o1777`:

```ant
try File.mkdir("shared")
try File.chmod("shared", 0o1777)
let info = try File.stat("shared")
println(String.fromInt(info.mode, base: 8))
```

```text
1777
```

The mode is set exactly, without the umask that applies when a file is created:

```ant
try File.write("open.txt", "")
try File.chmod("open.txt", 0o777)
let info = try File.stat("open.txt")
println(String.fromInt(info.mode, base: 8))
```

```text
777
```

Through a symbolic link, `File.chmod` changes the file the link points to:

```ant
try File.write("run.sh", "make\n")
try File.chmod("run.sh", 0o644)
try File.symlink("run.sh", "latest")
try File.chmod("latest", 0o700)
let info = try File.stat("run.sh")
println(String.fromInt(info.mode, base: 8))
```

```text
700
```

## Applications

Add the execute bits, as `chmod +x` does, keeping the rest of the mode:

```ant
func makeExecutable(path: string): Result[void, Error] {
  let info = try File.stat(path)
  File.chmod(path, Bits.or(info.mode, 0o111))
}

try File.write("build.sh", "make\n")
try File.chmod("build.sh", 0o640)
try makeExecutable("build.sh")
let info = try File.stat("build.sh")
println(String.fromInt(info.mode, base: 8))
```

```text
751
```

Take every permission away from the group and others on a set of keys, as `chmod go-rwx` does:

```ant
try File.mkdir("keys")
for name in ["a.pem", "b.pem"] {
  try File.write("keys/${name}", "secret")
}

for path in try File.glob("keys/*.pem") {
  let info = try File.stat(path)
  try File.chmod(path, Bits.and(info.mode, 0o700))
  let after = try File.stat(path)
  println("${path} ${String.fromInt(after.mode, base: 8)}")
}
```

```text
keys/a.pem 600
keys/b.pem 600
```

Say which step failed with [Result.wrapErr](../Result/wrapErr.md):

```ant
let r = File.chmod("server.key", 0o600) | Result.wrapErr("protect the key")
println(r)
```

```text
Err{ error: Error{ message: "protect the key: chmod server.key: no such file or directory" } }
```

## Properties and relations

[File.stat](stat.md) reads back the mode `File.chmod` set:

```ant
try File.write("f.txt", "")
for mode in [0o644, 0o600, 0o755, 0o400, 0o4755, 0] {
  try File.chmod("f.txt", mode)
  let info = try File.stat("f.txt")
  assert(info.mode == mode)
}
```

## Possible issues

A mode written in decimal is a different mode: `755` is `0o1363`. Write modes as octal literals:

```ant
try File.write("deploy.sh", "")
try File.chmod("deploy.sh", 755)
let info = try File.stat("deploy.sh")
println(String.fromInt(info.mode, base: 8))
```

```text
1363
```

A mode outside `0 ..= 0o7777` is an `Err`, and the file keeps its mode:

```ant
try File.write("deploy.sh", "")
println(File.chmod("deploy.sh", 0o10000))
println(File.chmod("deploy.sh", -1))
```

```text
Err{ error: Error{ message: "chmod deploy.sh: mode 0o10000 is outside 0 ..= 0o7777" } }
Err{ error: Error{ message: "chmod deploy.sh: mode -1 is outside 0 ..= 0o7777" } }
```

`File.chmod` does not create the file. A missing file is an `Err`, which `try` turns into the end of
a script:

<!-- test: exit 1 -->

```ant
try File.chmod("missing.sh", 0o755)
println("not reached")
```

Writing a file and then calling `File.chmod` leaves it readable by others until `File.chmod` runs,
since `File.write` creates it with the umask's mode, usually `0o644`. To write a secret, write a
temporary file beside the destination with [File.tempFile](tempFile.md), which is `0o600` from the
start, and move it into place with [File.rename](rename.md):

```ant
func saveToken(path: string, token: string): Result[void, Error] {
  let tmp = try File.tempFile(".token-*", dir: Path.dir(path))
  defer { File.remove(tmp) }
  try File.write(tmp, token)
  File.rename(tmp, path)
}

try saveToken("token.txt", "s3cr3t")
let info = try File.stat("token.txt")
println(String.fromInt(info.mode, base: 8))
```

```text
600
```

## See also

[File.stat](stat.md) · [File.Info](types/Info.md) · [File.copy](copy.md) ·
[File.tempFile](tempFile.md) · [File.rename](rename.md) · [File.touch](touch.md) ·
[Bits.or](../Bits/or.md) · [Bits.and](../Bits/and.md) · [String.fromInt](../String/fromInt.md)
