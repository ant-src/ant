<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.run

```ant signature
func Cmd.run(c: Cmd | Cmd.Pipeline, ctx: Context | null = null): Result[int, Error]
```

Cmd.run runs c in the foreground, with the program's own standard input, output and error, so a
pager, an editor or ssh can use the terminal, and returns its status, nonzero included. While it
runs the ant command holds back Ctrl-C, which every stage gets and decides on, and passes it on when
the last stage dies of it. It is Err when a stage could not be started, and when ctx is done, which
kills every stage.

<!-- end generated -->

## Details

- A nonzero status is `Ok`, not `Err`: `Cmd.run` fails only when a stage cannot be started or `ctx`
  ends. Check the status yourself.
- The command writes straight to the program's standard output and error, and reads its standard
  input. Nothing is captured; use [Cmd.read](read.md) or [Cmd.capture](capture.md) for the output.
- There is no shell: `args` arrive as they are, with no globbing, variables, or pipes.
- A pipeline's status is that of the rightmost stage that failed, or 0. Only its first stage reads
  standard input.
- Where the host has signals, a child that a signal ended has status `128 + n`, so SIGKILL is `137`
  and Ctrl-C is `130`.

## Basic examples

`Cmd.run(c)` runs a command, lets its output through, and returns its status:

```ant
let status = try Cmd.run(Cmd.new("echo", ["hello"]))
println("status ${status}")
```

```text
hello
status 0
```

A command that fails still gives `Ok` of its status:

```ant
let status = try Cmd.run(Cmd.new("sh", ["-c", "exit 3"]))
println("status ${status}")
```

```text
status 3
```

## Scope

A pipeline from [Cmd.pipe](pipe.md) runs its stages at once, and the last one writes to standard
output:

```ant
let unique = Cmd.pipe([Cmd.new("printf", ["b\\na\\nb\\n"]), Cmd.new("sort", ["-u"])])
let status = try unique.run()
println("status ${status}")
```

```text
a
b
status 0
```

The command reads the program's own standard input:

<!-- test: stdin
pear
apple
-->

```ant
try Cmd.new("sort").run()
```

```text
apple
pear
```

`dir` runs the command in another directory:

<!-- test: file notes/todo.txt
buy milk
-->

```ant
try Cmd.new("ls", dir: "notes").run()
```

```text
todo.txt
```

`env` is the whole environment of the command. To add one variable, start from
[Os.environ](../Os/environ.md):

```ant
let env = Os.environ()
env["GREETING"] = "hello"
try Cmd.new("sh", ["-c", "echo $GREETING"], env: env).run()
```

```text
hello
```

A program that cannot be started is `Err`:

```ant
let result = Cmd.new("no-such-program").run()
println(result.isOk())
```

```text
false
```

## Named arguments

### ctx

When `ctx` ends while the command runs, `Cmd.run` kills every stage and returns `Err` with the
context's reason:

```ant
let ctx = Context.new(timeout: 100ms)
let result = Cmd.new("sleep", ["10"]).run(ctx: ctx)
println(result)
```

```text
Err{ error: Error{ message: "timeout" } }
```

A context that is already done starts nothing:

```ant
let ctx = Context.new()
ctx.cancel()
println(Cmd.new("echo", ["never printed"]).run(ctx: ctx))
```

```text
Err{ error: Error{ message: "cancelled" } }
```

## Applications

Run build steps in order, log each, and stop at the first that fails:

```ant
let steps = [
  Cmd.new("echo", ["compiling"]),
  Cmd.new("sh", ["-c", "echo testing; exit 1"]),
  Cmd.new("echo", ["deploying"]),
]
for step in steps {
  println("+ ${step.string()}")
  let status = try step.run()
  if status != 0 {
    println("failed with status ${status}")
    break
  }
}
```

```text
+ echo compiling
compiling
+ sh -c 'echo testing; exit 1'
testing
failed with status 1
```

Exit with a failed command's status, and go on when it succeeds:

<!-- test: exit 3 -->

```ant
let status = try Cmd.new("sh", ["-c", "echo working; exit 3"]).run()
if status != 0 { Os.exit(status) }
println("done")
```

```text
working
```

Turn a nonzero status into an error, so `try` stops the script:

<!-- test: exit 1 -->

```ant
func check(c: Cmd): Result[void, Error] {
  let status = try c.run()
  if status != 0 {
    return Result.fail("${c.string()}: exit status ${status}")
  }
  Result.ok(void)
}

try check(Cmd.new("true"))
try check(Cmd.new("false"))
println("not reached")
```

## Properties and relations

`Cmd.run` returns the status that [Cmd.capture](capture.md) reports, without capturing the output:

```ant
let c = Cmd.new("sh", ["-c", "exit 4"])
let output = try c.capture()
let status = try c.run()
assert(status == output.status)
```

A pipeline's status is that of its rightmost failed stage, so a later stage that succeeds does not
hide a failure:

```ant
func exitWith(n: int): Cmd {
  Cmd.new("sh", ["-c", "exit ${n}"])
}

let status = try Cmd.pipe([exitWith(2), exitWith(3), Cmd.new("true")]).run()
assert(status == 3)
```

A stage before the last that SIGPIPE ends does not count:

```ant
let firstLine = Cmd.pipe([Cmd.new("yes"), Cmd.new("head", ["-n", "1"])])
let status = try firstLine.run()
assert(status == 0)
```

```text
y
```

Where the host has signals, a child that a signal ends has status `128 + n`:

```ant
let status = try Cmd.new("sh", ["-c", "kill -9 $$"]).run()
assert(status == 128 + 9)
```

## Possible issues

`try` does not stop on a nonzero status, since that is `Ok`. Compare the status:

```ant
let status = try Cmd.new("false").run()
if status != 0 {
  println("failed with status ${status}")
}
```

```text
failed with status 1
```

No shell reads the arguments, so variables and globs stay as written:

```ant
try Cmd.new("echo", ["$HOME", "*.txt"]).run()
```

```text
$HOME *.txt
```

For the same reason a whole command line in `program` names a program that does not exist. Split it
into `program` and `args`:

```ant
let result = Cmd.new("echo hello").run()
println(result.isOk())
```

```text
false
```

An `env` that is not built from `Os.environ()` leaves the command only those variables:

```ant
try Cmd.new("env", env: Map{ "GREETING": "hi" }).run()
```

```text
GREETING=hi
```

The output is not returned. To use what a command prints, call [Cmd.read](read.md):

```ant
let text = try Cmd.read(Cmd.new("echo", ["hello"]))
println(text.toUpper())
```

```text
HELLO
```

What the program has already read ahead from standard input stays the program's, so a command run
after the program read a line does not see the rest. To hand the rest on, read it and pass it as
`input` to [Cmd.read](read.md) or [Cmd.capture](capture.md):

<!-- test: stdin
name
pear
apple
-->

```ant
let header = try File.readLine(Os.stdin())
println("header: ${header}")
try Cmd.new("cat").run()
println("cat saw nothing")
```

```text
header: name
cat saw nothing
```

While `Cmd.run` waits, the program holds back Ctrl-C for the whole process, so a server that waits
for [Os.interrupt](../Os/interrupt.md) runs its commands with [Cmd.start](start.md) instead.

A pipeline of no commands panics:

```ant
let empty = Cmd.Pipeline{ cmds: [] }
let result = empty.run() // panic: Cmd.run needs a pipeline of at least one command
```

## See also

[Cmd.new](new.md), [Cmd.pipe](pipe.md), [Cmd.read](read.md), [Cmd.capture](capture.md),
[Cmd.start](start.md), [Cmd.string](string.md), [Os.exit](../Os/exit.md),
[Os.environ](../Os/environ.md)
