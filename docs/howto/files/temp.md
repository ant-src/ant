<!-- status: reviewed -->

# Use a temporary file or directory

Give a step of your program a place of its own for files that last only one run, and remove it when
the step is done.

This script compares two texts with `diff`, which reads files, not standard input. It writes both
texts into a temporary directory and runs `diff` there:

```ant
func diffText(before: string, after: string): Result[string, Error] {
  let dir = try File.tempDir("diff-*")
  defer { File.removeAll(dir) }
  try File.write(Path.join([dir, "before.txt"]), before)
  try File.write(Path.join([dir, "after.txt"]), after)
  let out = try Cmd.capture(Cmd.new("diff", ["before.txt", "after.txt"], dir: dir))
  if out.status > 1 {
    return Result.fail("diff: ${out.stderr.trim()}")
  }
  Result.ok(out.stdout)
}

print(try diffText("apples\npears\nplums\n", "apples\nquinces\nplums\n"))
```

```text
2c2
< pears
---
> quinces
```

1. Make the directory with [File.tempDir](../../lib/File/tempDir.md). It creates a new, empty
   directory in the host's temporary directory, under a name nothing else has, that only you can
   open, and returns its path. The random string replaces the last `*` of the pattern, or goes at
   the end when the pattern has no `*`.
2. Remove it with [File.removeAll](../../lib/File/removeAll.md) in a `defer` right after you make
   it. Nothing else removes it, at exit or later. The deferred block runs on every path out of the
   function, so the directory is removed even when a `try` returns early.
3. Build the paths inside it with [Path.join](../../lib/Path/join.md), and write the files with
   [File.write](../../lib/File/write.md).
4. Run the command in the directory with `dir:` on [Cmd.new](../../lib/Cmd/new.md), so its arguments
   and its output name the files without the random part. [Cmd.capture](../../lib/Cmd/capture.md)
   returns `Ok` for every status, since `diff` exits with 1 when the texts differ.

The host's temporary directory is `$TMPDIR`, or `/tmp` when that is unset, on Unix, and `%TEMP%` on
Windows.

## Make one temporary file

When one file is enough, call [File.tempFile](../../lib/File/tempFile.md) and remove the file with
[File.remove](../../lib/File/remove.md). It returns a path, not a handle, so write to it as to any
other path. Put the `*` before the extension, so the name keeps it for a program that looks at it:

```ant
let path = try File.tempFile("names-*.txt")
defer { File.remove(path) }
try File.write(path, "Cy\nAnn\nBob\n")
println(Path.ext(path))
println(try Cmd.read(Cmd.new("sort", [path])))
```

```text
.txt
Ann
Bob
Cy
```

On Unix the new file has mode `0o600` and the new directory `0o700`, so other users cannot read what
you write there. [File.stat](../../lib/File/stat.md) shows it:

```ant
let path = try File.tempFile("token-*")
defer { File.remove(path) }
let info = try File.stat(path)
println(String.fromInt(info.mode, base: 8))
```

```text
600
```

On Windows the default temporary directory is under the user's profile, so what you make there is
yours alone.

## Choose the directory

Give `dir:` to create the file or directory somewhere other than the host's temporary directory. The
directory must exist. The pattern names only the file, so a path separator in it is an `Err`:

```ant
try File.mkdir("work")
let log = try File.tempFile("job-*.log", dir: "work")
defer { File.remove(log) }
println(Path.dir(log))
println(File.tempFile("work/job-*.log"))
```

```text
work
Err{ error: Error{ message: "createtemp work/job-*.log: pattern contains path separator" } }
```

A temporary file in the destination's own directory is how you replace a file so that no reader sees
half of it: write the temporary file, then rename it over the destination.
[Replace a file without a partial write](replace-safely.md) shows how.

## Check that a directory can be written

The mode of a directory does not tell whether you can create files in it. Try it instead: create a
temporary file there and remove it at once:

```ant
func canWrite(dir: string): bool {
  let probe = File.tempFile(".probe-*", dir: dir)
  if probe is Err {
    return false
  }
  _ = File.remove(probe.value)
  true
}

println(canWrite("."))
println(canWrite("missing"))
```

```text
true
false
```

## Remove it before you exit

[Os.exit](../../lib/Os/exit.md) ends the program at once and runs no deferred block. Remove the
directory yourself before you call it:

<!-- test: exit 3 -->

```ant
let work = try File.tempDir("convert-*")
defer { File.removeAll(work) }
let input = Path.join([work, "input.txt"])
try File.write(input, "")
if (try File.read(input)).isEmpty() {
  eprintln("nothing to convert")
  _ = File.removeAll(work)
  Os.exit(3)
}
```

Returning from the script, with `return` or with a `try` that fails, runs the deferred block, so
only `Os.exit` needs this.

## See also

[File.tempDir](../../lib/File/tempDir.md) · [File.tempFile](../../lib/File/tempFile.md) ·
[File.removeAll](../../lib/File/removeAll.md) · [File.remove](../../lib/File/remove.md) ·
[File.write](../../lib/File/write.md) · [File.read](../../lib/File/read.md) ·
[File.stat](../../lib/File/stat.md) · [File.mkdir](../../lib/File/mkdir.md) ·
[Path.join](../../lib/Path/join.md) · [Path.ext](../../lib/Path/ext.md) ·
[Path.dir](../../lib/Path/dir.md) · [Cmd.new](../../lib/Cmd/new.md) ·
[Cmd.capture](../../lib/Cmd/capture.md) · [Cmd.read](../../lib/Cmd/read.md) ·
[Os.exit](../../lib/Os/exit.md) · [Replace a file without a partial write](replace-safely.md)
