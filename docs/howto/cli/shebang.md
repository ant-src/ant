<!-- status: reviewed -->

# Run a script as a command

Give a script a `#!` line and the execute bit, and it runs as `./lines.ant` instead of
`ant run lines.ant`.

## The program

`lines.ant` counts the lines of each file named on its command line:

<!-- test: file notes.txt
milk
eggs
bread
-->

<!-- test: cmd ant lines.ant notes.txt -->

```ant [lines.ant]
#!/usr/bin/env ant
//! Counts the lines of each file given as an argument.

let paths = Os.args()
if paths.isEmpty() {
  eprintln("usage: ${Path.base(Os.program())} <file>...")
  Os.exit(2)
}
for path in paths {
  let text = try File.read(path)
  println("${text.lines().length()} ${path}")
}
```

```text
3 notes.txt
```

## Steps

1. Put `#!/usr/bin/env ant` on the first line of the script, starting at its first byte. Module doc
   lines (`//!`) go after it.
2. Make the file executable:

   ```sh
   chmod +x lines.ant
   ```

3. Run it by its path. The arguments after the path reach [`Os.args()`](../../lib/Os/args.md), as
   they do after `ant run lines.ant`:

   ```sh
   ./lines.ant notes.txt
   ```

4. To run it from any directory, copy it into a directory on your `PATH`, such as `~/bin`. It may
   keep its `.ant` extension or drop it: `ant` runs any existing file, so a copy named `lines` runs
   as `lines notes.txt`.

The `#!` line makes the system run `ant ./lines.ant notes.txt`, which is
`ant run ./lines.ant notes.txt`. So `ant` must be on your `PATH` too.

[`Os.program()`](../../lib/Os/program.md) is the script path as it was run, such as `./lines.ant`
when the `#!` line runs it from the working directory. [`Path.base`](../../lib/Path/base.md) keeps
its last element, so the usage line reads `lines.ant` however the script is started.

## Keep the `#!` line first

`ant` skips a `#!` line only when it is the first line of the file and starts at its first byte.
Anywhere else, after a comment, a blank line, or a space, it is a syntax error:

```ant
//! Counts the lines of each file given as an argument.
#!/usr/bin/env ant
// error: unexpected character '#'
```

## See also

- [ant run and the REPL](../../tools/run.md): running a file directly and the exit statuses.
- [Read command-line arguments](args.md)
- [Exit with a status](exit-status.md)
- [`Os.exit`](../../lib/Os/exit.md), [`File.read`](../../lib/File/read.md),
  [`String.lines`](../../lib/String/lines.md), [`Array.isEmpty`](../../lib/Array/isEmpty.md),
  [`Array.length`](../../lib/Array/length.md), [`eprintln`](../../lib/prelude/eprintln.md),
  [`println`](../../lib/prelude/println.md), [`Path.base`](../../lib/Path/base.md)
