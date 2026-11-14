<!-- status: reviewed -->

# Exit with a status

A script exits with status `0` when it ends, and with `1` when `try` returns an `Err` from it or the
script panics. Call [`Os.exit`](../../lib/Os/exit.md) for any other status.

1. Pick a status for each outcome: `0` for success, and other numbers from 1 to 125 for the rest.
   List them in the script's doc comment. `ant` itself uses `1` for a panic and `2` for a program it
   cannot load, so pick `3` and up for outcomes a caller must tell apart from those.
2. Check the arguments first. On a usage error, write a usage line to standard error with
   [`eprintln`](../../lib/prelude/eprintln.md) and call `Os.exit(2)`.
3. Use [`try`](../../language/keywords/try.md) for expected failures, such as a file that cannot be
   read. The script writes the error to standard error and exits with `1`.
4. Call `Os.exit` with the status of any other outcome. Let the script end for `0`.

This script lists the TODO lines of the files given to it:

<!-- test: file notes.txt
buy milk
TODO: call Ann
-->
<!-- test: file clean.txt
all done
-->
<!-- test: cmd ant run todo.ant notes.txt clean.txt -->
<!-- test: exit 3 -->

```ant [todo.ant]
//! Lists the TODO lines of the files given as arguments.
//!
//! Exit status: 0 when no file has one, 1 when a file cannot be read, 2 on a usage error, and 3 when
//! some file has a TODO line.

if Os.args().isEmpty() {
  eprintln("usage: ${Path.base(Os.program())} <file>...")
  Os.exit(2)
}

var found = false
for path in Os.args() {
  let text = try File.read(path)
  for line in text.lines().filter({ x.contains("TODO") }) {
    println("${path}: ${line}")
    found = true
  }
}
if found {
  Os.exit(3)
}
```

```text
notes.txt: TODO: call Ann
```

Read the status in the shell with `$?`:

```sh
$ ant run todo.ant notes.txt clean.txt; echo $?
notes.txt: TODO: call Ann
3
$ ant run todo.ant clean.txt; echo $?
0
$ ant run todo.ant; echo $?
usage: todo.ant <file>...
2
$ ant run todo.ant missing.txt; echo $?
open missing.txt: no such file or directory
1
```

## Clean up before you exit

`Os.exit` ends the process at once, and no deferred block runs. Remove what a
[`defer`](../../language/keywords/defer.md) would remove before the call. This version keeps the
TODO lines of each file in a temporary directory:

<!-- test: file notes.txt
buy milk
TODO: call Ann
-->
<!-- test: args notes.txt -->
<!-- test: exit 3 -->

```ant
if Os.args().isEmpty() {
  eprintln("usage: todo <file>...")
  Os.exit(2)
}
let dir = try File.tempDir("todo-*")
defer { File.removeAll(dir) }
var found = false
for path in Os.args() {
  let text = try File.read(path)
  let todos = text.lines().filter({ x.contains("TODO") })
  if todos.isEmpty() {
    continue
  }
  try File.write(Path.join([dir, Path.base(path)]), todos | String.join("\n"))
  found = true
}
if found {
  File.removeAll(dir)
  Os.exit(3)
}
```

To fail with status `1` and run the deferred blocks, return an `Err` from the script with
[`Result.fail`](../../lib/Result/fail.md) instead:

<!-- test: exit 1 -->

```ant
defer { println("cleanup") }
if Os.args().isEmpty() {
  return Result.fail("no files given")
}
```

```text
cleanup
```

## See also

[Os.exit](../../lib/Os/exit.md) · [Os.args](../../lib/Os/args.md) ·
[Os.program](../../lib/Os/program.md) · [Path.base](../../lib/Path/base.md) ·
[File.read](../../lib/File/read.md) · [String.lines](../../lib/String/lines.md) ·
[Array.filter](../../lib/Array/filter.md) · [String.contains](../../lib/String/contains.md) ·
[File.tempDir](../../lib/File/tempDir.md) · [File.removeAll](../../lib/File/removeAll.md) ·
[File.write](../../lib/File/write.md) · [Path.join](../../lib/Path/join.md) ·
[String.join](../../lib/String/join.md) · [Array.isEmpty](../../lib/Array/isEmpty.md) ·
[println](../../lib/prelude/println.md) · [eprintln](../../lib/prelude/eprintln.md) ·
[Read command-line arguments](args.md)
