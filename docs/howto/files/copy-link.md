<!-- status: reviewed -->

# Copy, move, and link files

Copy a file or a whole directory, move it to a new name, and point a symbolic link at it.

## Copy a file

Call [File.copy](../../lib/File/copy.md) with the path of the copy. To copy into a directory, name
the file there with [Path.join](../../lib/Path/join.md):

```ant
try File.write("app.conf", "port=8080\n")
try File.chmod("app.conf", 0o600)
try File.mkdir("backup")
let copy = Path.join(["backup", "app.conf"])
try File.copy("app.conf", copy)
print(try File.read(copy))
println(String.fromInt((try File.stat(copy)).mode, base: 8))
```

```text
port=8080
600
```

1. Make the directory that holds the copy with [File.mkdir](../../lib/File/mkdir.md). `copy` creates
   no directories, and is `Err` when the one that would hold the copy is missing.
2. Copy the file. `copy` replaces a file already at the path, and is `Err` when a directory is
   there. When the path is a symbolic link, `copy` writes through it and replaces the target's
   contents. To put a file where the link is, remove the link with
   [File.remove](../../lib/File/remove.md) first.
3. The copy has the contents and the mode of the source, as [File.stat](../../lib/File/stat.md)
   shows; [String.fromInt](../../lib/String/fromInt.md) prints the mode in octal. The times are not
   copied, so the copy is new to a build tool. Nor is the owner: `copy` gives the copy the source's
   contents and mode only. Run as root, [File.chown](../../lib/File/chown.md) gives it the source's
   owner from the `uid` and `gid` of the source's `File.stat`:
   `let info = try File.stat("app.conf")` and then
   `try File.chown(copy, info.uid ?? -1, info.gid ?? -1)`.

A reader can see the copy half-written. To replace a file that others read, see
[Replace a file without a partial write](replace-safely.md).

## Keep a file that is already there

Pass `replace: false` to copy only when nothing is at the path. The `Err` does not say why the copy
failed, so check [File.exists](../../lib/File/exists.md) to tell an existing file from other
failures:

```ant
try File.write("report.txt", "monday\n")
try File.write("draft.txt", "tuesday\n")
let copied = File.copy("draft.txt", "report.txt", replace: false)
if !copied.isOk() && File.exists("report.txt") {
  println("report.txt is already there; keeping it")
}
print(try File.read("report.txt"))
```

```text
report.txt is already there; keeping it
monday
```

[Result.isOk](../../lib/Result/isOk.md) tests the outcome. Anything at the path counts, a directory
and a symbolic link included.

## Copy a directory

Call [File.copyAll](../../lib/File/copyAll.md) to copy a directory with everything in it. The
destination must not exist yet, or be an empty directory:

```ant
try File.mkdir("site/css")
try File.write("site/index.html", "<h1>Home</h1>\n")
try File.write("site/css/main.css", "h1 { color: navy }\n")
try File.copyAll("site", "site-backup")
println(try File.walk("site-backup"))
println(File.copyAll("site", "site-backup"))
```

```text
["css/main.css", "index.html"]
Err{ error: Error{ message: "copy site to site-backup: site-backup exists" } }
```

1. Copy the tree. Every directory and file keeps its mode, and a symbolic link inside the tree is
   copied as a link with the same target.
2. List the copy with [File.walk](../../lib/File/walk.md) to see what arrived.
3. Copy again onto the same destination, and `copyAll` is `Err` before it copies anything, so two
   trees are never merged. Remove the old copy first with
   [File.removeAll](../../lib/File/removeAll.md) to start over. `copyAll` is also `Err` when the
   destination is inside the source, such as `File.copyAll("site", "site/backup")`, since the copy
   would contain itself. Put the copy beside the tree.

## Move or rename

Call [File.rename](../../lib/File/rename.md) to give a file or a directory a new path. It replaces a
file at the new path:

```ant
try File.write("draft.md", "# Plan\n")
try File.write("plan.md", "old plan\n")
try File.rename("draft.md", "plan.md")
println(File.exists("draft.md"))
print(try File.read("plan.md"))
```

```text
false
# Plan
```

`rename` has no "into" form: it is `Err` when a directory is at the new path, so name the file
inside the directory with `Path.join`:

```ant
try File.mkdir("archive")
try File.write("2025.log", "closed\n")
println(File.rename("2025.log", "archive").isOk())
try File.rename("2025.log", Path.join(["archive", "2025.log"]))
println(try File.list("archive"))
```

```text
false
["2025.log"]
```

[File.list](../../lib/File/list.md) shows the directory's names. `rename` is also `Err` across file
systems. To move a file there, copy it and then remove the original with
[File.remove](../../lib/File/remove.md):

```ant
func move(from: string, to: string): Result[void, Error] {
  if File.rename(from, to).isOk() {
    return Result.ok(void)
  }
  try File.copy(from, to)
  File.remove(from)
}

try File.write("upload.tmp", "data\n")
try move("upload.tmp", "upload.dat")
print(try File.read("upload.dat"))
```

```text
data
```

`isOk` tells the fast path from the fallback, and [Result.ok](../../lib/Result/ok.md) returns from
it.

`move` handles files: `File.copy` copies one regular file. To move a directory to another file
system, copy it with `File.copyAll` and remove the original with `File.removeAll`.

## Link to a file

Call [File.symlink](../../lib/File/symlink.md) with the target first and the link second, in the
order of `ln -s`. A relative target is relative to the link's directory, not to the working
directory:

```ant
try File.mkdir("releases/v1")
try File.write("releases/v1/app.txt", "version 1\n")
try File.mkdir("bin")
try File.symlink("../releases/v1/app.txt", "bin/app")
print(try File.read("bin/app"))
println(try File.readLink("bin/app"))
println((try File.realPath("bin/app")) == (try File.realPath("releases/v1/app.txt")))
```

```text
version 1
../releases/v1/app.txt
true
```

1. Make the link. Nothing needs to exist at the target yet.
2. Read through the link: [File.read](../../lib/File/read.md) and most other functions follow it.
3. Read the target as stored with [File.readLink](../../lib/File/readLink.md).
4. Resolve every link on the way with [File.realPath](../../lib/File/realPath.md), which returns an
   absolute path. It is `Err` when the link leads nowhere.

## Point a link somewhere else

`symlink` is `Err` when anything is at the link's path. To switch a link to a new target in one
step, make the new link under a temporary name and rename it over the old one, since `rename` acts
on the link and not on its target:

```ant
try File.mkdir("releases/v1")
try File.mkdir("releases/v2")
try File.write("releases/v1/version.txt", "1\n")
try File.write("releases/v2/version.txt", "2\n")
try File.symlink("releases/v1", "current")
println(File.symlink("releases/v2", "current").isOk())

try File.symlink("releases/v2", "current.new")
try File.rename("current.new", "current")
print(try File.read("current/version.txt"))
println(try File.readLink("current"))
```

```text
false
2
releases/v2
```

On Unix the rename is atomic, so a reader that opens `current/version.txt` during the switch sees
the old release or the new one, never no link at all. To change the link in two steps instead,
remove it with `File.remove`, which removes the link and leaves its target, and then call `symlink`
again.

## See also

[File.copy](../../lib/File/copy.md) · [File.copyAll](../../lib/File/copyAll.md) ·
[File.rename](../../lib/File/rename.md) · [File.symlink](../../lib/File/symlink.md) ·
[File.readLink](../../lib/File/readLink.md) · [File.realPath](../../lib/File/realPath.md) ·
[File.remove](../../lib/File/remove.md) · [File.removeAll](../../lib/File/removeAll.md) ·
[File.stat](../../lib/File/stat.md) · [File.chown](../../lib/File/chown.md) ·
[File.mkdir](../../lib/File/mkdir.md) · [Path.join](../../lib/Path/join.md) ·
[Replace a file without a partial write](replace-safely.md)
