<!-- status: reviewed -->

# Processes and the OS

Two namespaces cover a program's place in the operating system. [Os](../Os/index.md) is the
program's own process: its arguments, environment, standard streams, exit status, and interrupts.
[Cmd](../Cmd/index.md) runs other programs. A `Cmd` is a record that says what to run, and the
function you call on it says how: in the foreground, for its output, or in the background.

<!-- test: args Ann -->

```ant
let who = Os.args().get(0) ?? "world"
println(try Cmd.new("echo", ["hello", who]).read())
```

```text
hello Ann
```

No shell runs a command, so its arguments arrive as they are. Join commands with
[Cmd.pipe](../Cmd/pipe.md) instead of `|`, and every function that runs a `Cmd` runs the pipeline
too:

```ant
let fixes = Cmd.pipe([Cmd.new("grep", ["fix"]), Cmd.new("wc", ["-l"])])
println(fixes.string())
let count = try fixes.read(input: "fix a\nadd b\nfix c\n")
println(count.trim())
```

```text
grep fix | wc -l
2
```

## Choosing how to run a command

| You want                                                | Use                                                                                                 |
| ------------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| to run a program the user watches or types into         | [Cmd.run](../Cmd/run.md)                                                                            |
| one value a program prints                              | [Cmd.read](../Cmd/read.md), or [Cmd.readBytes](../Cmd/readBytes.md)                                 |
| the status and both outputs, without failing on nonzero | [Cmd.capture](../Cmd/capture.md), or [Cmd.captureBytes](../Cmd/captureBytes.md)                     |
| the output as it comes, or a program that keeps running | [Cmd.start](../Cmd/start.md) with [Os.openPipe](../Os/openPipe.md), then [Cmd.wait](../Cmd/wait.md) |

`Cmd.run` gives the command the program's own streams and returns its status, so a script can pass a
failure on with [Os.exit](../Os/exit.md). A nonzero status is `Ok` here, not `Err`:

<!-- test: exit 3 -->

```ant
let step = Cmd.new("sh", ["-c", "echo testing; exit 3"])
eprintln("+ ${step.string()}")
let status = try step.run()
if status != 0 {
  Os.exit(status)
}
```

```text
testing
```

`Cmd.read` turns a nonzero status into an `Err` that carries the command's standard error, and
`Cmd.capture` keeps the status and both outputs:

```ant
println(Cmd.new("sh", ["-c", "echo oops >&2; exit 2"]).read())
let out = try Cmd.new("sh", ["-c", "echo data; echo oops >&2; exit 2"]).capture()
println("${out.status}: ${out.stdout.trim()} / ${out.stderr.trim()}")
```

```text
Err{ error: Error{ message: "sh: exit status 2: oops" } }
2: data / oops
```

`env` is the child's whole environment, so start from `Os.environ()` to add one variable:

```ant
let env = Os.environ()
env["GREETING"] = "hi"
println(try Cmd.new("sh", ["-c", "echo $GREETING"], env: env).read())
```

```text
hi
```

`Cmd.start` returns as soon as the command runs. Give it the writing end of a pipe to read the
output line by line, and close your copy of the writing end, so reads reach end of file once the
command exits:

```ant
let out = try Os.openPipe()
defer { out.read.close() }
let job = try Cmd.start(Cmd.new("sh", ["-c", "echo one; echo two"]), stdout: out.write)
try out.write.close()
while true {
  let line = try out.read.readLine() ?? break
  println("got ${line}")
}
println("status ${try job.wait()}")
```

```text
got one
got two
status 0
```

A command started in the background keeps running when the script ends. Stop it in a deferred block.
Here the script runs a helper until Ctrl-C, which [Os.interrupt](../Os/interrupt.md) turns into a
done context instead of the end of the run:

<!-- test: before
go {
  Time.sleep(100ms)
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
let stop = Os.interrupt()
let helper = try Cmd.start(Cmd.new("sleep", ["10"]))
defer { helper.stop() }
stop.wait()
println("shutting down after ${stop.reason()}")
```

```text
shutting down after interrupt
```

## Functions by job

### Arguments and identity

- [Os.args](../Os/args.md): the arguments after the program path
- [Os.program](../Os/program.md): the program path, as the command gave it
- [Os.pid](../Os/pid.md): this process's id
- [Os.uid](../Os/uid.md): the effective user id, or `null` on Windows

### Environment and host

- [Os.env](../Os/env.md): one variable, or `null` when it is not set
- [Os.environ](../Os/environ.md): every variable, as a new map
- [Os.platform](../Os/platform.md), [Os.arch](../Os/arch.md): the operating system and processor, by
  Go's names
- [Os.hostname](../Os/hostname.md): the host's name
- [Os.cwd](../Os/cwd.md): the working directory
- [Os.home](../Os/home.md): the user's home directory
- [Os.configDir](../Os/configDir.md), [Os.cacheDir](../Os/cacheDir.md): where programs keep
  configuration and what can be made again

### Streams and pipes

- [Os.stdin](../Os/stdin.md), [Os.stdout](../Os/stdout.md), [Os.stderr](../Os/stderr.md): the
  standard streams, as handles
- [Os.openPipe](../Os/openPipe.md): two connected handles, as an [Os.Pipe](../Os/types/Pipe.md)

### Ending the process

- [Os.exit](../Os/exit.md): end the process now with a status; no deferred block runs
- [Os.interrupt](../Os/interrupt.md): a context done at Ctrl-C, SIGINT, or SIGTERM

### Building commands

- [Cmd.new](../Cmd/new.md): a command from a program, its arguments, a directory, and an environment
- [Cmd.pipe](../Cmd/pipe.md): commands joined as a shell's `|` joins them, as a
  [Cmd.Pipeline](../Cmd/types/Pipeline.md)
- [Cmd.string](../Cmd/string.md): a command as a shell-quoted line for a log
- [Os.which](../Os/which.md): the path of the program a `Cmd` would start, or `null`

### Running to the end

- [Cmd.run](../Cmd/run.md): run in the foreground with the program's streams, and return the status
- [Cmd.read](../Cmd/read.md): the standard output as text without one final newline, `Err` on a
  nonzero status
- [Cmd.readBytes](../Cmd/readBytes.md): the standard output as bytes, exactly as written
- [Cmd.capture](../Cmd/capture.md): the status and both outputs as text, in a
  [Cmd.Output](../Cmd/types/Output.md)
- [Cmd.captureBytes](../Cmd/captureBytes.md): the same, with standard output as bytes, in a
  [Cmd.OutputBytes](../Cmd/types/OutputBytes.md)

### Running in the background

- [Cmd.start](../Cmd/start.md): start a command with the handles given, and return a
  [Cmd.Process](../Cmd/types/Process.md)
- [Cmd.wait](../Cmd/wait.md): the status once the command has exited
- [Cmd.kill](../Cmd/kill.md): send a signal by name
- [Cmd.stop](../Cmd/stop.md): SIGTERM, then SIGKILL after a grace period
- [Cmd.pid](../Cmd/pid.md): the process id of the last stage

## See also

[Os](../Os/index.md) · [Cmd](../Cmd/index.md) · [Files and paths](files.md) ·
[Concurrency](concurrency.md) · [File.Handle](../File/types/Handle.md) ·
[Run a command and capture its output](../../howto/processes/run.md) ·
[Stream a command's output](../../howto/processes/stream.md) ·
[Handle Ctrl-C](../../howto/cli/interrupt.md) ·
[Build a pipeline](../../howto/processes/pipeline.md) ·
[Stop or time out a process](../../howto/processes/stop.md) ·
[Read command-line arguments](../../howto/cli/args.md) ·
[Read environment variables and standard directories](../../howto/cli/env.md) ·
[Read standard input line by line](../../howto/cli/stdin.md) ·
[Exit with a status](../../howto/cli/exit-status.md)
