<!-- status: reviewed -->

# Files and paths

Two namespaces share the work. [Path](../Path/index.md) builds and takes apart paths as text and
never reads the disk, so most of its functions return plain values. `Path.abs`, `Path.rel`, and
`Path.match` return a `Result`: a working directory that cannot be read, a target with no relative
path from the base, or a malformed pattern makes them fail. [File](../File/index.md) reads, writes,
and changes what is on disk, so every call that can fail returns a `Result`. You build a path with
`Path`, then hand it to `File`:

```ant
let dir = "notes"
try File.mkdir(dir)
let path = Path.join([dir, "todo.txt"])
try File.write(path, "buy milk\n")
println(path)
print(try File.read(path))
println(File.read(Path.join([dir, "done.txt"])))
```

```text
notes/todo.txt
buy milk
Err{ error: Error{ message: "open notes/done.txt: no such file or directory", kind: Error.NotExist{} } }
```

A read of a missing file is an `Err`, and its message names the path. Only `File.mkdir` makes
missing directories, with all their parents. `File.write` and the other calls that create a file are
`Err` when the directory that would hold it is missing, so the example makes `notes` first.

`File.write`, `File.writeBytes`, `File.create`, and `File.copy` replace what is at the path. Pass
`replace: false` to create the file only when nothing is there; otherwise the call is `Err` and
changes nothing:

```ant
try File.write("out.txt", "first")
println(File.write("out.txt", "second", replace: false))
println(try File.read("out.txt"))
```

```text
Err{ error: Error{ message: "open out.txt: file exists", kind: Error.Exist{} } }
first
```

A handle is open until it is closed. Defer the close right after opening, and close again at the end
of a write: a close can fail, and a deferred close cannot report it.

```ant
let rows = ["north,12", "south,7"]
let out = (try File.create("report.txt")).buffered()
defer { out.close() }
for row in rows {
  try out.writeLine(row)
}
try out.close()
print(try File.read("report.txt"))
```

```text
north,12
south,7
```

A close hands the data to the OS but does not wait for the disk. When a file must survive a crash or
a power loss, call `try out.sync()` before the final close, as when a temporary file is renamed into
place: see [Replace a file without a partial write](../../howto/files/replace-safely.md).

A `Path` function works on the text alone, so the path does not have to exist:

```ant
Path.join(["logs", "..", "app.log"]) // "app.log"
Path.base("reports/2026/march.csv")  // "march.csv"
Path.ext("reports/2026/march.csv")   // ".csv"
Path.isAbs("reports/2026/march.csv") // false
```

[File.walk](../File/walk.md) and [File.glob](../File/glob.md) give paths relative to the directory
they search, so `Path.join` puts the directory back in front before another `File` call:

```ant
try File.mkdir("site/posts")
try File.write("site/index.md", "# Home\n")
try File.write("site/posts/first.md", "# First post\n")
let dir = "site"
for path in try File.glob("**/*.md", dir: dir) {
  let info = try File.stat(Path.join([dir, path]))
  println("${path}: ${info.size} bytes")
}
```

```text
index.md: 7 bytes
posts/first.md: 13 bytes
```

A path is used as written: `~` is not the home directory, and a relative path is taken against the
working directory. Build a path under the home directory with `Path.join([try Os.home(), ...])`.

A mode is a number written in octal. On Unix, [File.tempFile](../File/tempFile.md) makes a file only
you can read, with mode `0o600`:

```ant
let tmp = try File.tempFile("token-*", dir: ".")
defer { File.remove(tmp) }
let mode = (try File.stat(tmp)).mode
println(String.fromInt(mode, base: 8))
println(mode == 0o600)
```

```text
600
true
```

## Functions by job

### Whole files

- [File.read](../File/read.md), [File.readBytes](../File/readBytes.md): the contents of a file, as
  text or as bytes
- [File.write](../File/write.md), [File.writeBytes](../File/writeBytes.md): create a file, or
  replace what it held

### Building and taking paths apart

- [Path.join](../Path/join.md): join parts with the separator, cleaned
- [Path.clean](../Path/clean.md): the shortest equivalent path
- [Path.dir](../Path/dir.md): all but the last element
- [Path.base](../Path/base.md): the last element
- [Path.ext](../Path/ext.md): the extension of the last element

### Absolute and relative paths

- [Path.isAbs](../Path/isAbs.md): whether a path is absolute
- [Path.abs](../Path/abs.md): a path made absolute against the working directory
- [Path.rel](../Path/rel.md): a target as a path from a base
- [Os.cwd](../Os/cwd.md), [Os.home](../Os/home.md): the working directory and the home directory
- [Os.cacheDir](../Os/cacheDir.md), [Os.configDir](../Os/configDir.md): where the user's programs
  keep what can be made again, and their configuration

### Finding files

- [File.exists](../File/exists.md): whether anything is at a path
- [File.list](../File/list.md): the names in one directory
- [File.readDir](../File/readDir.md): the entries of one directory, each a
  [File.Entry](../File/types/Entry.md) with its name and whether it is a directory or a symbolic
  link
- [File.walk](../File/walk.md): every file below a directory
- [File.glob](../File/glob.md): the paths below a directory that match a pattern
- [Path.match](../Path/match.md): whether a path matches a pattern, without reading the disk

### Opening and closing a handle

- [File.Handle](../File/types/Handle.md): an open file, a standard stream, a connection, or a buffer
  over one
- [File.open](../File/open.md): open a file for reading
- [File.create](../File/create.md): create or truncate a file for writing
- [File.append](../File/append.md): open a file for writing at its end
- [File.close](../File/close.md): close a handle, writing out a buffer first
- [Os.stdin](../Os/stdin.md), [Os.stdout](../Os/stdout.md), [Os.stderr](../Os/stderr.md): the
  standard streams as handles
- [File.isTerminal](../File/isTerminal.md): whether a handle is a terminal

### Reading and writing a handle

- [File.readLine](../File/readLine.md): the next line, or `null` at the end
- [File.readAll](../File/readAll.md): the rest of the input as text
- [File.readChunk](../File/readChunk.md): what is there, up to a number of bytes
- [File.readExactly](../File/readExactly.md): exactly a number of bytes
- [File.writeLine](../File/writeLine.md), [File.writeText](../File/writeText.md): text, with or
  without a newline
- [File.writeChunk](../File/writeChunk.md): bytes as they are
- [File.pipe](../File/pipe.md): copy the rest of one handle to another
- [File.buffered](../File/buffered.md): a handle that holds its writes until full, flush, or close
- [File.flush](../File/flush.md): hand what a buffered handle holds to the OS
- [File.sync](../File/sync.md): flush, then have the OS write the file to storage, so it survives a
  crash

### Metadata and modes

- [File.stat](../File/stat.md): the [File.Info](../File/types/Info.md) of a path: name, size, kind,
  modified time, mode, and owner ids
- [File.chmod](../File/chmod.md): set the mode
- [File.chown](../File/chown.md): set the user and group that own a file, by id
- [File.touch](../File/touch.md): set the modified time, creating an empty file when missing

### Creating, moving, copying, and removing

- [File.mkdir](../File/mkdir.md): make a directory with its parents
- [File.rename](../File/rename.md): move a file or directory
- [File.copy](../File/copy.md): copy one file's contents and mode
- [File.copyAll](../File/copyAll.md): copy a file or a whole tree
- [File.remove](../File/remove.md): remove a file or an empty directory
- [File.removeAll](../File/removeAll.md): remove a whole tree

### Links

- [File.symlink](../File/symlink.md): make a symbolic link
- [File.readLink](../File/readLink.md): the target of a link, as stored
- [File.realPath](../File/realPath.md): resolve every link along a path

### Temporary files and locks

- [File.tempFile](../File/tempFile.md), [File.tempDir](../File/tempDir.md): a new file or directory
  under a free name, only for you
- [File.lock](../File/lock.md): open a file holding its lock, waiting for it
- [File.tryLock](../File/tryLock.md): the same without waiting, or `null` when another handle holds
  it

## See also

[File](../File/index.md) · [Path](../Path/index.md) · [Os](../Os/index.md) ·
[Read and write a whole file](../../howto/files/read-write.md) ·
[Read a large file line by line](../../howto/files/lines.md) ·
[Walk a directory and match globs](../../howto/files/walk-glob.md) ·
[Replace a file without a partial write](../../howto/files/replace-safely.md) ·
[Use a temporary file or directory](../../howto/files/temp.md) ·
[Copy, move, and link files](../../howto/files/copy-link.md) ·
[Read and change modes, owners, and times](../../howto/files/metadata.md) ·
[Lock a file between processes](../../howto/files/lock.md)
