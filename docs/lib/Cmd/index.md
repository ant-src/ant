<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd

```ant signature
type Cmd = { program: string, args: Array[string] = [], dir: string | null = null, env: Map[string, string] | null = null }
```

Cmd is a program to run, as a value: a Cmd says what to run, and a function of Cmd says how. Nothing
runs it through a shell. program is looked up in PATH when it has no path separator, and is taken
relative to dir otherwise; the child runs in dir and with env as its whole environment when they are
given. A child ended by signal n has the status 128 + n, as a shell reports it.

<!-- end generated -->

A `Cmd` is a record that names a program, its arguments, and optionally the directory and
environment it runs in. Building one runs nothing. The functions of the `Cmd` namespace run it: in
the foreground, for its output, for its status and both outputs, or in the background. A command is
built once and then run, logged, or passed to a helper like any other record, and
[Cmd.pipe](pipe.md) joins commands into a pipeline that every one of those functions runs too.

## Details

- No shell runs the command: `args` arrive as they are, with no globbing, variables, quoting, or
  `|`. Join commands with [Cmd.pipe](pipe.md) instead.
- A `program` without a path separator is looked up in the program's own `PATH`, never in the `PATH`
  of `env`. [Os.which](../Os/which.md) shows what that lookup finds.
- `env` is the child's whole environment. To add one variable, start from
  [Os.environ](../Os/environ.md).
- A nonzero status is `Ok` for [Cmd.run](run.md), [Cmd.capture](capture.md), and
  [Cmd.wait](wait.md), and `Err` for [Cmd.read](read.md) and [Cmd.readBytes](readBytes.md). Every
  function that runs a command is `Err` when the command cannot be started. When its `ctx` ends
  first, the command is killed and the call is `Err` with the context's reason; for
  [Cmd.start](start.md), every `Cmd.wait` is `Err` instead.
- A command started with [Cmd.start](start.md) and still running when the script ends is left
  running. Stop it in a deferred block.

## Basic examples

`Cmd.new` builds a command, and `Cmd.read` runs it and returns what it prints:

```ant
let greeting = try Cmd.read(Cmd.new("echo", ["hello"]))
println(greeting.toUpper())
```

```text
HELLO
```

A `Cmd` is a record, and `Cmd.new` fills in the same defaults as a literal:

```ant
Cmd.new("git", ["status"])            // Cmd{ program: "git", args: ["status"], dir: null, env: null }
Cmd.new("ls") == Cmd{ program: "ls" } // true
```

`Cmd.run` lets the command use the program's own streams and returns its status:

```ant
let status = try Cmd.run(Cmd.new("sh", ["-c", "echo working; exit 3"]))
println("status ${status}")
```

```text
working
status 3
```

## Scope

A `Cmd` changes like any record before it runs. [Cmd.string](string.md) writes it as a shell would
need it quoted, for a log:

```ant
let verbose = true
let build = Cmd.new("go", ["build"])
if verbose {
  build.args.push("-v")
}
build.args.push("./...")
println("running ${build.string()}")
```

```text
running go build -v ./...
```

A [Cmd.Pipeline](types/Pipeline.md) runs wherever a `Cmd` runs:

```ant
let paths = Cmd.pipe([
  Cmd.new("grep", ["^GET"]),
  Cmd.new("cut", ["-d", " ", "-f", "2"]),
  Cmd.new("sort", ["-u"]),
])
println(try paths.read(input: "GET /\nPOST /login\nGET /about\nGET /\n"))
```

```text
/
/about
```

[Cmd.capture](capture.md) gives the status and both outputs, and a nonzero status is still `Ok`:

```ant
let out = try Cmd.new("sh", ["-c", "echo data; echo oops >&2; exit 2"]).capture()
println(out.status)
println(out.stdout.trim())
println(out.stderr.trim())
```

```text
2
data
oops
```

[Cmd.start](start.md) runs a command in the background and returns a
[Cmd.Process](types/Process.md). [Cmd.wait](wait.md) gives its status, and [Cmd.stop](stop.md) ends
it. `Cmd.stop` gives `143`, the status of a child that SIGTERM ended:

```ant
let quick = try Cmd.start(Cmd.new("echo", ["in the background"]))
println(try quick.wait())
let slow = try Cmd.start(Cmd.new("sleep", ["10"]))
println(try slow.stop())
```

```text
in the background
0
143
```

## Named arguments

`dir` runs the command in another directory:

<!-- test: file notes/todo.txt
buy milk
-->

```ant
println(try Cmd.new("ls", dir: "notes").read())
```

```text
todo.txt
```

`env` replaces the environment, so add a variable to a copy of [Os.environ](../Os/environ.md):

```ant
let env = Os.environ()
env["GREETING"] = "hi"
println(try Cmd.new("sh", ["-c", "echo $GREETING"], env: env).read())
```

```text
hi
```

`input` is what the command reads on standard input:

```ant
println(try Cmd.new("sort").read(input: "pear\napple\n"))
```

```text
apple
pear
```

Every function that runs a command takes a `ctx`. When it ends before the command exits, the command
is killed and the call is `Err` with the context's reason:

```ant
let ctx = Context.new(timeout: 100ms)
println(Cmd.new("sleep", ["10"]).run(ctx: ctx))
```

```text
Err{ error: Error{ message: "timeout" } }
```

## Functions by job

### Building

- [Cmd.new](new.md): a command from a program, its arguments, a directory, and an environment
- [Cmd.pipe](pipe.md): commands joined as a shell's `|` joins them
- [Cmd.string](string.md): a command as a shell-quoted line for a log

### Running to the end

- [Cmd.run](run.md): run in the foreground with the program's streams, and return the status
- [Cmd.read](read.md): the standard output as text, `Err` on a nonzero status
- [Cmd.readBytes](readBytes.md): the standard output as bytes, exactly as written
- [Cmd.capture](capture.md): the status and both outputs, `Ok` on a nonzero status

### Running in the background

- [Cmd.start](start.md): start a command with the handles given, and return at once
- [Cmd.wait](wait.md): the status once the command has exited
- [Cmd.kill](kill.md): send a signal by name
- [Cmd.stop](stop.md): SIGTERM, then SIGKILL after a grace period
- [Cmd.pid](pid.md): the process id of the last stage

### Types

- [Cmd.Pipeline](types/Pipeline.md): commands run at once, each one's output the next one's input
- [Cmd.Output](types/Output.md): what `Cmd.capture` collects
- [Cmd.Process](types/Process.md): a command started by `Cmd.start`

## Applications

Read a value from a command, and fall back when the command fails:

```ant
func branch(dir: string): string {
  let head = Cmd.new("git", ["rev-parse", "--abbrev-ref", "HEAD"], dir: dir).read()
  head.unwrapOr("unknown")
}

println(branch("/"))
```

```text
unknown
```

Check that the tools a script needs are installed before it runs any of them:

```ant
let missing = ["sh", "no-such-tool-ant"].filter({ Os.which(x) == null })
if !missing.isEmpty() {
  println("missing: ${missing | String.join(", ")}")
}
```

```text
missing: no-such-tool-ant
```

Start a helper in the background, stop it when the script ends, and read its output through a pipe:

```ant
let out = try Os.openPipe()
defer { out.read.close() }
let ticker = try Cmd.start(Cmd.new("sh", ["-c", "echo ready; sleep 10"]), stdout: out.write)
defer { ticker.stop() }
try out.write.close()
println(try out.read.readLine())
```

```text
ready
```

Run steps in order, log each, and stop at the first that fails:

<!-- test: exit 1 -->

```ant
let steps = [Cmd.new("echo", ["compiling"]), Cmd.new("false"), Cmd.new("echo", ["deploying"])]
for step in steps {
  println("+ ${step.string()}")
  let status = try step.run()
  if status != 0 {
    eprintln("failed with status ${status}")
    Os.exit(status)
  }
}
```

```text
+ echo compiling
compiling
+ false
```

## Properties and relations

`Cmd.read` is the standard output that `Cmd.capture` collects, without one final newline:

```ant
let c = Cmd.new("printf", ["one\ntwo\n"])
let out = try c.capture()
assert(try c.read() == out.stdout.trimSuffix("\n"))
```

`Cmd.string` leaves out `env` and `dir`, so a secret in the environment stays out of the log:

```ant
let deploy = Cmd.new("deploy", ["--prod"], dir: "/srv", env: Map{ "TOKEN": "secret" })
assert(deploy.string() == Cmd.new("deploy", ["--prod"]).string())
```

## Possible issues

`Cmd` is a type. Build a command with `Cmd.new` or a literal:

```ant
let ls = Cmd("ls") // error: Cmd is a type, not a function
```

A whole command line in `program` names a program that does not exist. Split it into `program` and
`args`:

```ant
println(Cmd.new("echo hello").run().isOk())
println(try Cmd.new("echo", ["hello"]).read())
```

```text
false
hello
```

`try` does not stop on a nonzero status from `Cmd.run`, since that is `Ok`. `Cmd.read` makes it an
`Err`, with the program's standard error in the message:

```ant
println(Cmd.new("false").run())
println(Cmd.new("sh", ["-c", "echo bad >&2; exit 2"]).read())
```

```text
Ok{ value: 1 }
Err{ error: Error{ message: "sh: exit status 2: bad" } }
```

`Cmd.read` removes one final newline and no other white space:

```ant
println(Cmd.new("printf", ["a\n\n"]).read())
```

```text
Ok{ value: "a\n" }
```

## See also

[Cmd.new](new.md) · [Cmd.pipe](pipe.md) · [Cmd.run](run.md) · [Cmd.read](read.md) ·
[Cmd.capture](capture.md) · [Cmd.start](start.md) · [Os](../Os/index.md) ·
[Os.which](../Os/which.md) · [Os.environ](../Os/environ.md) · [Os.openPipe](../Os/openPipe.md) ·
[Context](../Context/index.md)
