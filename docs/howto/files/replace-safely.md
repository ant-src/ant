<!-- status: reviewed -->

# Replace a file without a partial write

Write the new contents to a temporary file beside the old one, then rename it over the old one. On
Unix, a program that reads the file at the same time sees the old contents or the new ones, never
half of each, and a write that fails, or a crash, leaves the old file as it was.

```ant
func replaceFile(path: string, text: string): Result[void, Error] {
  let tmp = try File.tempFile(".${Path.base(path)}-*", dir: Path.dir(path))
  defer { File.remove(tmp) }
  let out = try File.create(tmp)
  defer { out.close() }
  try out.writeText(text)
  try out.sync()
  try out.close()
  File.rename(tmp, path)
}

try File.mkdir("config")
try File.write("config/settings.txt", "theme=light\n")
try replaceFile("config/settings.txt", "theme=dark\n")
print(try File.read("config/settings.txt"))
println(try File.list("config"))
```

```text
theme=dark
["settings.txt"]
```

1. Create the temporary file with [File.tempFile](../../lib/File/tempFile.md) in the directory of
   the target, which [Path.dir](../../lib/Path/dir.md) gives. A rename across file systems is an
   `Err`, so do not use the host's temporary directory. Start the pattern with a `.` and the name
   from [Path.base](../../lib/Path/base.md), so a glob such as `*` does not match it.
2. Remove the temporary file in a [defer](../../language/keywords/defer.md) right after creating it,
   so an `Err` from a later step does not leave it behind. After the rename,
   [File.remove](../../lib/File/remove.md) finds nothing there and its `Err` is discarded.
3. Open the temporary file with [File.create](../../lib/File/create.md), close it in a `defer`, and
   write the contents with [File.writeText](../../lib/File/writeText.md).
4. Call [File.sync](../../lib/File/sync.md) before the rename. Closing a file hands its data to the
   OS but does not wait for the disk, so after a crash or power loss the new name could point at an
   empty or partial file. `sync` has the OS write it to storage, so it survives a crash. Then close
   it and check the result.
5. Rename the temporary file over the target with [File.rename](../../lib/File/rename.md). On Unix
   the rename replaces the target in one step. On other systems it is not promised to be atomic.

When the path may be a symbolic link, `File.rename` replaces the link itself. To replace the file it
points at, resolve the path first with [File.realPath](../../lib/File/realPath.md) and use the
result for both `Path.dir` and the rename.

[File.list](../../lib/File/list.md) shows that no temporary file is left in the directory.

## Keep the mode of the old file

`File.tempFile` creates the file with mode `0o600`, and the rename puts that file in place, so on
Unix the new file is readable only by you. That suits a file that holds a secret, such as a token.
To keep the mode of the file you replace, read it with [File.stat](../../lib/File/stat.md) and set
it on the temporary file with [File.chmod](../../lib/File/chmod.md) before the rename. Skip it with
[File.exists](../../lib/File/exists.md) when there is no old file yet:

```ant
func replaceFile(path: string, text: string): Result[void, Error] {
  let tmp = try File.tempFile(".${Path.base(path)}-*", dir: Path.dir(path))
  defer { File.remove(tmp) }
  if File.exists(path) {
    let info = try File.stat(path)
    try File.chmod(tmp, info.mode)
  }
  let out = try File.create(tmp)
  defer { out.close() }
  try out.writeText(text)
  try out.sync()
  try out.close()
  File.rename(tmp, path)
}

try File.write("deploy.sh", "echo old\n")
try File.chmod("deploy.sh", 0o755)
try replaceFile("deploy.sh", "echo new\n")
let info = try File.stat("deploy.sh")
println(String.fromInt(info.mode, base: 8))
```

```text
755
```

The new file also belongs to the user who runs the program. Run as root,
[File.chown](../../lib/File/chown.md) gives it the old file's owner from the same `info`:
`try File.chown(tmp, info.uid ?? -1, info.gid ?? -1)` next to the `chmod`.

## Write the new file a line at a time

To write output in many small pieces, open the temporary file with
[File.create](../../lib/File/create.md) and write through
[File.buffered](../../lib/File/buffered.md). Sync and then close the handle with
[File.close](../../lib/File/close.md), and check the result before the rename: a failed sync or
close means the file may be incomplete, and the old file must stay. On a buffered handle `sync`
writes out the buffer first. The deferred close covers the paths that leave early and does nothing
after the explicit one.

```ant
func writeLines(path: string, lines: Array[string]): Result[void, Error] {
  let tmp = try File.tempFile(".${Path.base(path)}-*", dir: Path.dir(path))
  defer { File.remove(tmp) }
  let out = File.buffered(try File.create(tmp))
  defer { out.close() }
  for line in lines {
    try out.writeLine(line)
  }
  try out.sync()
  try out.close()
  File.rename(tmp, path)
}

try writeLines("squares.txt", ["1", "4", "9"])
print(try File.read("squares.txt"))
```

```text
1
4
9
```

[File.writeLine](../../lib/File/writeLine.md) writes each line to the buffer.

## Never replace an existing file

To create a file only when nothing is at the path, pass `replace: false` to `File.write`. The check
and the creation happen in one step, so another program cannot slip a file in between. When
something is there, `File.write` returns an `Err` and changes nothing:

```ant
try File.write("report.txt", "first\n", replace: false)
let saved = File.write("report.txt", "second\n", replace: false)
if saved is Err {
  println("kept: ${saved.error.message}")
}
print(try File.read("report.txt"))
```

```text
kept: open report.txt: file exists
first
```

`File.writeBytes`, `File.create` and `File.copy` take `replace: false` too.

## See also

[File.tempFile](../../lib/File/tempFile.md) · [File.rename](../../lib/File/rename.md) ·
[File.sync](../../lib/File/sync.md) · [File.writeText](../../lib/File/writeText.md) ·
[File.write](../../lib/File/write.md) · [File.remove](../../lib/File/remove.md) ·
[File.stat](../../lib/File/stat.md) · [File.chmod](../../lib/File/chmod.md) ·
[File.chown](../../lib/File/chown.md) · [File.exists](../../lib/File/exists.md) ·
[File.create](../../lib/File/create.md) · [File.buffered](../../lib/File/buffered.md) ·
[File.close](../../lib/File/close.md) · [File.writeLine](../../lib/File/writeLine.md) ·
[File.read](../../lib/File/read.md) · [File.list](../../lib/File/list.md) ·
[File.mkdir](../../lib/File/mkdir.md) · [File.realPath](../../lib/File/realPath.md) ·
[Path.dir](../../lib/Path/dir.md) · [Path.base](../../lib/Path/base.md) ·
[String.fromInt](../../lib/String/fromInt.md) · [Use a temporary file or directory](temp.md) ·
[Read and write a whole file](read-write.md)
