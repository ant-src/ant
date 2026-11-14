<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.kill

```ant signature
func Cmd.kill(p: Cmd.Process, signal: string = "SIGTERM"): Result[void, Error]
```

Cmd.kill sends the signal named, "SIGTERM", "SIGKILL", "SIGINT", "SIGHUP", "SIGQUIT", "SIGUSR1" or
"SIGUSR2", to p's process group, and does not wait. It is Ok when p has exited already, and Err for
a name the host does not have.

<!-- end generated -->

## Details

- The signal goes to the command's process group: every stage of a pipeline, and the processes they
  started, get it.
- `Cmd.kill` returns once the signal is sent. [Cmd.wait](wait.md) gives the status. A command ended
  by signal `n` has the status `128 + n`, as a shell reports it: `143` for SIGTERM, `137` for
  SIGKILL.
- The name is the full name in capitals, such as `"SIGTERM"`. Any other name is `Err`.
- On Windows the signal reaches each stage alone, and a command it ends has the status `1`.

## Basic examples

Send SIGTERM, the default, and wait for the command to end:

```ant
let p = try Cmd.start(Cmd.new("sleep", ["10"]))
try Cmd.kill(p)
println(try Cmd.wait(p))
```

```text
143
```

Send SIGKILL instead, by name:

```ant
let p = try Cmd.start(Cmd.new("sleep", ["10"]))
try Cmd.kill(p, signal: "SIGKILL")
println(try Cmd.wait(p))
```

```text
137
```

## Scope

A command that has exited already is `Ok`, and nothing is sent:

```ant
let p = try Cmd.start(Cmd.new("true"))
try Cmd.wait(p)
println(Cmd.kill(p))
```

```text
Ok{ value: void }
```

A pipeline gets the signal in every stage, so one call ends all of them:

```ant
let p = try Cmd.start(Cmd.pipe([Cmd.new("sleep", ["10"]), Cmd.new("cat")]))
try Cmd.kill(p)
println(try Cmd.wait(p))
```

```text
143
```

A program that handles the signal ends its own way. This shell script saves its state on SIGTERM and
exits with `0`. It prints `ready` once its handler is set, so the signal does not arrive before it:

```ant
let out = try Os.openPipe()
defer { out.read.close() }
let script = "trap 'echo saving state; exit 0' TERM; echo ready; while :; do sleep 0.1; done"
let worker = try Cmd.start(Cmd.new("sh", ["-c", script]), stdout: out.write)
try out.write.close()
println(try out.read.readLine())
try Cmd.kill(worker)
println(try out.read.readLine())
println(try Cmd.wait(worker))
```

```text
ready
saving state
0
```

## Named arguments

`signal` names the signal. Each one ends `sleep`, which handles none of them, with its own status:

```ant
for name in ["SIGTERM", "SIGKILL", "SIGINT", "SIGHUP", "SIGQUIT"] {
  let p = try Cmd.start(Cmd.new("sleep", ["10"]))
  try Cmd.kill(p, signal: name)
  println("${name}: ${try Cmd.wait(p)}")
}
```

```text
SIGTERM: 143
SIGKILL: 137
SIGINT: 130
SIGHUP: 129
SIGQUIT: 131
```

## Applications

Ask a server to reload with SIGUSR1, which leaves it running:

```ant
let out = try Os.openPipe()
defer { out.read.close() }
let script = "trap 'echo reloading' USR1; trap 'exit 0' TERM; echo ready; while :; do sleep 0.1; done"
let server = try Cmd.start(Cmd.new("sh", ["-c", script]), stdout: out.write)
try out.write.close()
defer { Cmd.stop(server) }
println(try out.read.readLine())
try Cmd.kill(server, signal: "SIGUSR1")
println(try out.read.readLine())
try Cmd.kill(server, signal: "SIGUSR1")
println(try out.read.readLine())
```

```text
ready
reloading
reloading
```

Give a command a deadline, then interrupt it as Ctrl-C would:

```ant
let job = try Cmd.start(Cmd.new("sleep", ["10"]))
let ctx = Context.new(timeout: 100ms)
if Cmd.wait(job, ctx: ctx) is Err {
  println("no answer in time, interrupting")
  try Cmd.kill(job, signal: "SIGINT")
}
println(try Cmd.wait(job))
```

```text
no answer in time, interrupting
130
```

## Properties and relations

The status is `128` plus the signal's number, and a second `Cmd.kill` after the exit is `Ok`:

```ant
let p = try Cmd.start(Cmd.new("sleep", ["10"]))
try Cmd.kill(p)
assert(try Cmd.wait(p) == 128 + 15)
assert(Cmd.kill(p) is Ok)
assert(Cmd.kill(p, signal: "SIGKILL") is Ok)
assert(try Cmd.wait(p) == 143)
```

## Possible issues

A short or lowercase name is `Err`, and nothing is sent:

```ant
let p = try Cmd.start(Cmd.new("sleep", ["10"]))
defer { Cmd.stop(p, grace: 0s) }
println(Cmd.kill(p, signal: "TERM"))
println(Cmd.kill(p, signal: "sigterm"))
```

```text
Err{ error: Error{ message: "unknown signal TERM" } }
Err{ error: Error{ message: "unknown signal sigterm" } }
```

Without `try`, the `Err` from a misspelled name is dropped and the command keeps running. Write
`try Cmd.kill(...)`.

`Cmd.kill` does not wait, and a program may handle SIGTERM and keep running. This script ignores it,
so a wait with a deadline runs out, and only SIGKILL ends it. [Cmd.stop](stop.md) does both steps:

```ant
let out = try Os.openPipe()
defer { out.read.close() }
let script = "trap '' TERM; echo ready; while :; do sleep 0.1; done"
let stubborn = try Cmd.start(Cmd.new("sh", ["-c", script]), stdout: out.write)
try out.write.close()
println(try out.read.readLine())
try Cmd.kill(stubborn)
let ctx = Context.new(timeout: 300ms)
println(Cmd.wait(stubborn, ctx: ctx))
try Cmd.kill(stubborn, signal: "SIGKILL")
println(try Cmd.wait(stubborn))
```

```text
ready
Err{ error: Error{ message: "timeout" } }
137
```

## See also

[Cmd.stop](stop.md) · [Cmd.wait](wait.md) · [Cmd.start](start.md) · [Cmd.pid](pid.md) ·
[Cmd.Process](types/Process.md) · [Context.new](../Context/new.md)
