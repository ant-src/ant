<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.read

```ant signature
func Cmd.read(c: Cmd | Cmd.Pipeline, input: string | Bytes = "", ctx: Context | null = null): Result[string, Error]
```

Cmd.read runs c with input as its standard input, waits for it to exit, and returns its standard
output as text without one final newline. A nonzero status is Err with the message "program: exit
status n", followed by ": " and standard error when it wrote any; program is that of the stage that
set the status, with "(stage i)" after it when another stage runs the same program. Err also means a
stage could not be started or ctx was done, which kills every stage.

<!-- end generated -->

## Details

- `Cmd.read` removes one final `\n` or `\r\n` from the output, as a shell's `$(...)` does, and keeps
  all other white space. [`Cmd.readBytes`](readBytes.md) returns the output exactly.
- Standard error is collected, not shown. It appears only in the message of an `Err`, with its
  trailing white space removed.
- Any nonzero status is `Err`. For a program whose nonzero status is an answer, such as `grep` with
  no match, use [`Cmd.capture`](capture.md).
- No shell runs the command, so `$HOME`, `*.txt`, and `|` in `args` reach the program as they are.
  Join commands with [`Cmd.pipe`](pipe.md).
- A string `input` is given as its UTF-8, and `Bytes` as they are. The default `""` gives the
  command an empty standard input, never the program's own.

## Basic examples

Run a program and get what it prints:

```ant
let greeting = try Cmd.read(Cmd.new("echo", ["hello", "world"]))
println(greeting)
```

```text
hello world
```

Give the command standard input:

```ant
let loud = try Cmd.read(Cmd.new("tr", ["a-z", "A-Z"]), "quiet please")
println(loud)
```

```text
QUIET PLEASE
```

A nonzero status is `Err`, with what the command wrote to standard error:

```ant
println(Cmd.read(Cmd.new("sh", ["-c", "echo 'disk full' >&2; exit 3"])))
```

```text
Err{ error: Error{ message: "sh: exit status 3: disk full" } }
```

## Scope

Only one final line break goes. Other line breaks, and spaces, stay:

```ant
println(Cmd.read(Cmd.new("printf", ["a\nb\n\n"])))
println(Cmd.read(Cmd.new("printf", ["  42  \r\n"])))
```

```text
Ok{ value: "a\nb\n" }
Ok{ value: "  42  " }
```

A command that prints nothing gives `""`:

```ant
println(Cmd.read(Cmd.new("true")))
```

```text
Ok{ value: "" }
```

Input of several lines reaches the command as it is:

```ant
let out = try Cmd.read(Cmd.new("sort"), "pear\napple\nfig\n")
println(out.lines())
```

```text
["apple", "fig", "pear"]
```

`Bytes` as input reach the command as they are:

```ant
let input = Bytes.fromArray([0x41, 0x6e, 0x74, 0x0a])
println(try Cmd.read(Cmd.new("cat"), input))
```

```text
Ant
```

A command that writes to standard error and exits with status 0 is `Ok`. Its standard error is
dropped:

```ant
println(Cmd.read(Cmd.new("sh", ["-c", "echo result; echo 'warning: slow' >&2"])))
```

```text
Ok{ value: "result" }
```

The `dir` and `env` of the `Cmd` set where the command runs and what environment it sees:

```ant
println(try Cmd.read(Cmd.new("pwd", dir: "/")))
let env = Map{ NAME: "Ant" }
println(try Cmd.read(Cmd.new("sh", ["-c", `echo "hi $NAME"`], env: env)))
```

```text
/
hi Ant
```

A pipeline reads `input` in its first stage and returns the output of its last:

```ant
let words = "b\na\nb\nc\na\n"
let unique = try Cmd.read(Cmd.pipe([Cmd.new("sort"), Cmd.new("uniq")]), words)
println(unique.lines())
```

```text
["a", "b", "c"]
```

The status of a pipeline is that of its rightmost stage that failed. When another stage runs the
same program, the message names the stage, counting from 1:

```ant
let pipeline = Cmd.pipe([
  Cmd.new("printf", ["ant\nbee\n"]),
  Cmd.new("grep", ["a"]),
  Cmd.new("grep", ["z"]),
])
println(Cmd.read(pipeline))
```

```text
Err{ error: Error{ message: "grep (stage 3): exit status 1" } }
```

A stage that ends because a later stage stopped reading does not fail the pipeline:

```ant
println(Cmd.read(Cmd.pipe([Cmd.new("yes"), Cmd.new("head", ["-n", "2"])])))
```

```text
Ok{ value: "y\ny" }
```

A program that cannot be started is `Err` too:

```ant
println(Cmd.read(Cmd.new("no-such-program")))
println(Cmd.read(Cmd.new("ls", dir: "/no/such/dir")))
```

```text
Err{ error: Error{ message: "exec: \"no-such-program\": executable file not found in $PATH" } }
Err{ error: Error{ message: "chdir /no/such/dir: no such file or directory" } }
```

## Named arguments

### input

`input:` names the standard input, which reads well next to the other arguments:

```ant
let first = try Cmd.read(Cmd.new("head", ["-n", "1"]), input: "one\ntwo\nthree\n")
println(first)
```

```text
one
```

### ctx

When `ctx` is done while the command runs, every stage is killed and the result is `Err` with the
context's reason:

```ant
let ctx = Context.new(50ms)
println(Cmd.read(Cmd.new("sleep", ["10"]), ctx: ctx))
```

```text
Err{ error: Error{ message: "timeout" } }
```

A context that is already done starts nothing:

```ant
let ctx = Context.new()
ctx.cancel()
println(Cmd.read(Cmd.new("echo", ["never"]), ctx: ctx))
```

```text
Err{ error: Error{ message: "cancelled" } }
```

## Applications

Ask a program for an answer only it has. `git` gives the id it would store a blob under:

```ant
let id = try Cmd.read(Cmd.new("git", ["hash-object", "--stdin"]), "hello\n")
println(id)
```

```text
ce013625030ba8dba906f756967f9e9ca394464a
```

Wrap a command in a function that names the task in its error, so the caller sees what failed:

```ant
func lineCount(paths: Array[string]): Result[int, Error] {
  let count = Cmd.pipe([Cmd.new("cat", paths), Cmd.new("wc", ["-l"])])
  let out = try (Cmd.read(count) | Result.wrapErr("count lines"))
  Result.ok(out.trim().toInt() ?? 0)
}

try File.write("a.txt", "one\ntwo\n")
try File.write("b.txt", "three\n")
println(lineCount(["a.txt", "b.txt"]))
println(lineCount(["a.txt", "missing.txt"]))
```

```text
Ok{ value: 3 }
Err{ error: Error{ message: "count lines: cat: exit status 1: cat: missing.txt: No such file or directory" } }
```

Run a short shell script when you need the shell's features, passing values as arguments, not by
building the script text:

```ant
let script = `for f in "$@"; do echo "[$f]"; done`
let files = ["notes.txt", "my file.txt"]
let out = try Cmd.read(Cmd.new("sh", ["-c", script, "sh", ...files]))
println(out)
```

```text
[notes.txt]
[my file.txt]
```

## Properties and relations

`Cmd.read` gives the `stdout` of [`Cmd.capture`](capture.md) without its final line break:

```ant
let c = Cmd.new("printf", ["one\ntwo\n"])
let out = try Cmd.capture(c)
assert(try Cmd.read(c) == out.stdout.trimSuffix("\n"))
```

[`Cmd.readBytes`](readBytes.md) keeps the line break that `Cmd.read` removes:

```ant
let c = Cmd.new("echo", ["hi"])
let bytes = try Cmd.readBytes(c)
assert(String.fromBytes(bytes) == "${try Cmd.read(c)}\n")
```

## Possible issues

No shell runs the command. A whole command line as `program` names a program that does not exist,
and `$HOME` or `*` in `args` is not expanded:

```ant
println(Cmd.read(Cmd.new("echo hello")))
println(try Cmd.read(Cmd.new("echo", ["$HOME", "*.txt"])))
```

```text
Err{ error: Error{ message: "exec: \"echo hello\": executable file not found in $PATH" } }
$HOME *.txt
```

A program that reports an answer with its status fails `Cmd.read`. `grep` exits with 1 when nothing
matches; [`Cmd.capture`](capture.md) gives the status as a value:

```ant
let c = Cmd.new("grep", ["panic"])
println(Cmd.read(c, "all good\n"))
let out = try Cmd.capture(c, "all good\n")
println(out.status)
```

```text
Err{ error: Error{ message: "grep: exit status 1" } }
1
```

`env` is the child's whole environment, not variables added to the program's own. With only `NAME`,
the child has no `HOME`, and no `PATH` either. Start from `Os.environ()`, a new map at each call,
and set the variable on it:

<!-- test: env HOME=/home/ant -->

```ant
let script = `echo "HOME=$HOME"`
println(try Cmd.read(Cmd.new("sh", ["-c", script], env: Map{ NAME: "Ant" })))
let env = Os.environ()
env["NAME"] = "Ant"
println(try Cmd.read(Cmd.new("sh", ["-c", script], env: env)))
```

```text
HOME=
HOME=/home/ant
```

Padding stays in the output, so trim it before parsing a number:

```ant
let out = try Cmd.read(Cmd.new("printf", ["%8d\n", "42"]))
println(out.toInt())
println(out.trim().toInt())
```

```text
null
42
```

## See also

[Cmd.readBytes](readBytes.md) · [Cmd.capture](capture.md) · [Cmd.run](run.md) ·
[Cmd.start](start.md) · [Cmd.new](new.md) · [Cmd.pipe](pipe.md) · [Os.which](../Os/which.md)
