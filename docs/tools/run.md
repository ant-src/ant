<!-- status: reviewed -->

# ant run and the REPL

`ant run` runs a program. A bare `ant` opens the REPL, where you type Ant and see each value.

```sh
ant run main.ant [args...]
ant main.ant [args...]
ant
```

## Details

- The file you give `ant run` is the program's script: a module with a statement at its top level.
  `ant run` loads it and every module it imports, initializes each library once, after its own
  imports, and runs the script last.
- `ant run` does not type-check. It parses every module first, so a syntax error anywhere stops the
  program before any of it runs. Annotations do nothing at run time, so a type mistake shows up only
  as a panic where a value is misused, if it shows up at all. Run [`ant check`](check.md) in your
  editor and in CI.
- The arguments after the file name reach the script through `Os.args()`. The file name is not one
  of them.

## Basic examples

A script is a file of statements. `ant run` runs them from top to bottom:

```ant
let langs = ["Ant", "Go", "Zig"]
for lang in langs {
  println("Hello, ${lang}")
}
```

```text
Hello, Ant
Hello, Go
Hello, Zig
```

The words after the file name are the script's arguments, and `Os.args()` returns them. This block
runs as `ant run main.ant Ann Bob`:

<!-- test: args Ann Bob -->

```ant
for name in Os.args() {
  println("Hello, ${name}")
}
```

```text
Hello, Ann
Hello, Bob
```

## Running a file directly

`ant greet.ant Ann` is the same as `ant run greet.ant Ann`. With a `#!/usr/bin/env ant` line first
and the execute bit set, the file runs on its own, as `./greet.ant Ann`:

<!-- test: cmd ant greet.ant Ann -->

```ant [greet.ant]
#!/usr/bin/env ant
for name in Os.args() {
  println("Hello, ${name}")
}
```

```text
Hello, Ann
```

```sh
chmod +x greet.ant
./greet.ant Ann
```

The `#!` line must start at the first byte of the file, before any blank line, space, or comment.
Module doc comments (`//!`) go after it. `ant` must be on your `PATH`.

To run the script from any directory, put it in a directory on your `PATH`, such as `~/bin`. It may
keep its `.ant` extension or drop it: `ant` runs any existing file.

## Exit status

| Status       | When                                                                                                       |
| ------------ | ---------------------------------------------------------------------------------------------------------- |
| `0`          | The script ends with its last line, or returns anything but an `Err`.                                      |
| `1`          | The script returns an `Err`, or the main goroutine panics.                                                 |
| `2`          | The program cannot be loaded: a syntax error, a missing module, an import cycle, or an import of a script. |
| `141`        | A write to standard output or standard error after its reader has gone, as in `ant run gen.ant \| head`.   |
| `130`, `143` | Ctrl-C or SIGTERM ends a program that has not called `Os.interrupt()`.                                     |
| any          | `Os.exit(n)` ends the process with `n`.                                                                    |

### Failing with an error

`return` and `try` work at the top level of a script as they do in a function. A script that returns
an `Err` writes its error to standard error and exits with `1`. This block exits with `1` and writes
`usage: main.ant <name>...` to standard error:

<!-- test: exit 1 -->

```ant
let names = Os.args()
if names.isEmpty() {
  return Result.fail("usage: main.ant <name>...")
}
for name in names {
  println("Hello, ${name}")
}
```

`try` is the short way to end the script with an error that a call returns. Deferred blocks run
before the script exits:

<!-- test: exit 1 -->

```ant
defer { println("done") }
let text = try File.read("missing.txt")
println(text)
```

```text
done
```

Standard error then holds the error alone, `open missing.txt: no such file or directory`, with no
stack trace, since `try` reports an expected failure. `Result.unwrap` in its place panics, which
reports a bug: the message and a stack trace.

### Panics

A panic in the main goroutine ends the program with `1`. A panic in any other goroutine ends only
that goroutine: its message and trace go to standard error, and the program goes on. `ant run`
writes `panic: `, the message, and a stack trace to standard error:

```ant
let xs = [1, 2, 3]
println(xs[0])
let last = xs[3] // panic: index 3 out of range for an array of length 3
println(last)
```

```text
1
```

### Os.exit

`Os.exit` ends the process at once with the status it is given. It skips deferred blocks:

<!-- test: exit 3 -->

```ant
defer { println("never printed") }
println("stopping")
Os.exit(3)
```

```text
stopping
```

## Running without checking

`ant run` skips the type checker, so it runs a program that `ant check` rejects until an operation
gets a value it cannot use:

```ant [total.ant]
println("start")
let n: int = "five" // error: cannot use string where int is expected
println(n + 1)
```

```sh
ant run total.ant
```

```text
start
panic: + needs two numbers of one kind or two strings, got string and int
  at module total.ant (total.ant:3:11)
```

The annotation `int` does nothing at run time: `n` holds a string, and the panic comes from `+` two
lines later. Run `ant check total.ant` first and the error points at the line that is wrong.

## Libraries and scripts

A module with no top-level statement is a library. Give one to `ant run` and it is initialized, and
the program exits with `0`. A script cannot be imported, so `ant run` refuses a program that imports
one, before anything runs, with status `2`.

`ant run` refuses a test file, a file whose name ends in `_test.ant`; run it with
[`ant test`](test.md). It refuses a declaration file too, and a program that reads a name which only
a declaration file beside it declares: such a program runs in the host that provides the name
([Modules](../guide/modules.md#declaration-files)).

## The REPL

`ant` with no arguments starts the REPL. Each input runs as its own small script, and what it
declares stays for the inputs after it. An input that ends in a value prints the value; `_` reads
the last value printed:

```text
> let prices = [3, 5, 8]
> prices.sum()
16
> _ * 2
32
```

`try` works in an input as it does in a script. An `Err` is written to standard error, and the REPL
goes on:

```text
> let config = try File.read("missing.txt")
open missing.txt: no such file or directory
> 1 + 1
2
```

A `let` holds for the whole session: a later input cannot assign it or declare the name again. A
`var` can be declared again.

A declaration or a `void` result prints nothing. Other results print as `println` writes them: a
string has no surrounding quotes, while a string inside an array keeps them. The prompt is `> ` for
a new input and `. ` for a line that continues an unfinished input. Prompts also appear when you
pipe input into `ant`.

### Commands and keys

Type `:help` to list the commands and editing keys:

| Command                | What it does                                                                                           |
| ---------------------- | ------------------------------------------------------------------------------------------------------ |
| `:type expr`           | Print the type of an expression.                                                                       |
| `:doc name`            | Look up standard-library documentation as `ant doc` does; for a name not found it suggests near names. |
| `:load file.ant`       | Evaluate a file in the current session.                                                                |
| `:reset`               | Start a new session, forgetting its declarations.                                                      |
| `:quit`, `:q`, `:exit` | End the session with status `0`.                                                                       |

In a terminal, Tab completes names and Up and Down browse the history, kept in `~/.ant_history`.
Ctrl-D on an empty line ends the session with status `0`, as does the end of piped input. Ctrl-C
drops the current input; press it again on an empty line to exit with status `130`.

## See also

- [ant check](check.md)
- [ant doc](doc.md)
- [ant test](test.md)
- [Runtime panics](../language/panics.md)
- [Os.args](../lib/Os/args.md)
- [Os.exit](../lib/Os/exit.md)
- [Os.interrupt](../lib/Os/interrupt.md)
