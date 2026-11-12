<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.wait

```ant signature
func Cmd.wait(p: Cmd.Process, ctx: Context | null = null): Result[int, Error]
```

Cmd.wait waits until every stage of p has exited and returns its status. Every call gets the same
outcome. It is Err when ctx is done first, and p goes on; and when the context given to
[`Cmd.start`](start.md) killed p, with that context's reason.

<!-- end generated -->

## Details

- A nonzero status is `Ok`, as is the status of a command that [Cmd.kill](kill.md) or
  [Cmd.stop](stop.md) ended, so `try` does not reject a failed command.
- A wait that its `ctx` ends leaves the command running. Wait again, or end it with
  [Cmd.stop](stop.md).
- A pipeline's status is that of the rightmost stage that failed, or 0; a stage before the last that
  SIGPIPE ended does not count, so `yes | head -n 1` is 0.
- Where the host has signals, a child ended by signal `n` has the status `128 + n`, as a shell
  reports it: SIGTERM is 143. On Windows a child that [Cmd.kill](kill.md) or [Cmd.stop](stop.md)
  ended has status 1.
- A child writing into a pipe nobody reads blocks once the pipe is full, and so does the wait. Read
  the pipe before the wait, or on another goroutine.

## Basic examples

Start a command, then wait for it to exit and get its status:

```ant
let child = try Cmd.new("echo", ["hello"]).start()
println(Cmd.wait(child))
```

```text
hello
Ok{ value: 0 }
```

A nonzero status is `Ok` too, so `try` passes it through:

```ant
let child = try Cmd.new("sh", ["-c", "exit 3"]).start()
let status = try Cmd.wait(child)
println("exited with ${status}")
```

```text
exited with 3
```

## Scope

### A pipeline

The wait ends when every stage has exited. The status is that of the rightmost stage that failed:

```ant
let pipeline = try Cmd.pipe([Cmd.new("sh", ["-c", "exit 2"]), Cmd.new("cat")]).start()
println(try pipeline.wait())
```

```text
2
```

### A command ended by a signal

Where the host has signals, a command that [Cmd.kill](kill.md) ends has the status `128 + n`, here
128 + 15 for SIGTERM:

```ant
let sleeper = try Cmd.new("sleep", ["30"]).start()
try sleeper.kill()
println(try sleeper.wait())
```

```text
143
```

### From several goroutines

Every wait on the same process gets the same status, whichever goroutine waits:

```ant
let child = try Cmd.new("sh", ["-c", "sleep 0.2; exit 4"]).start()
let seen = Parallel.map([1, 2, 3], func(_: int, _: Context): Result[int, Error] {
  child.wait()
}, limit: 3)
println(seen)
```

```text
[Ok{ value: 4 }, Ok{ value: 4 }, Ok{ value: 4 }]
```

## Named arguments

### ctx

A context bounds the wait. When it is done first the wait is `Err` with its reason, and the command
goes on, so a later wait still gets its status:

```ant
let child = try Cmd.new("sleep", ["1"]).start()
let ctx = Context.new(timeout: Time.duration(milliseconds: 100))
println(child.wait(ctx: ctx))
println(child.wait())
```

```text
Err{ error: Error{ message: "timeout" } }
Ok{ value: 0 }
```

A context given to [Cmd.start](start.md) bounds the command's life instead. When it ends, the
command is killed and every wait is `Err` with that context's reason:

```ant
let ctx = Context.new()
let child = try Cmd.new("sleep", ["30"]).start(ctx: ctx)
ctx.cancel()
println(child.wait())
println(child.wait())
```

```text
Err{ error: Error{ message: "cancelled" } }
Err{ error: Error{ message: "cancelled" } }
```

## Applications

### Run commands at once

Start every command before waiting for any, so they run side by side, then collect their statuses in
order:

```ant
type Job = { cmd: Cmd, process: Cmd.Process }

let cmds = [
  Cmd.new("sh", ["-c", "sleep 0.2; exit 0"]),
  Cmd.new("sh", ["-c", "sleep 0.1; exit 1"]),
  Cmd.new("true"),
]
let jobs: Array[Job] = []
for cmd in cmds {
  jobs.push(Job{ cmd, process: try cmd.start() })
}
for job in jobs {
  println("${job.cmd.string()}: ${try job.process.wait()}")
}
```

```text
sh -c 'sleep 0.2; exit 0': 0
sh -c 'sleep 0.1; exit 1': 1
true: 0
```

### Wait with a limit, then stop

Give a command a while to finish, and stop it when it takes longer:

```ant
func waitOrStop(p: Cmd.Process, limit: Time.Duration): Result[int, Error] {
  let waited = p.wait(ctx: Context.new(timeout: limit))
  if waited is Ok { return waited }
  p.stop(grace: Time.duration(seconds: 1))
}

let quick = try Cmd.new("true").start()
println(try waitOrStop(quick, Time.duration(seconds: 5)))
let stuck = try Cmd.new("sleep", ["30"]).start()
println(try waitOrStop(stuck, Time.duration(milliseconds: 200)))
```

```text
0
143
```

### Read the output, then wait

Read a pipe to its end first; the end comes when the command has closed its output, and the wait
then gives its status:

```ant
let out = try Os.openPipe()
defer { out.read.close() }
let child = try Cmd.new("sh", ["-c", "echo one; echo two; exit 5"]).start(stdout: out.write)
try out.write.close()
let text = try out.read.readAll()
println(text.lines())
println(try child.wait())
```

```text
["one", "two"]
5
```

## Properties and relations

Waiting twice gives the same outcome:

```ant
let child = try Cmd.new("sh", ["-c", "exit 7"]).start()
assert(child.wait() == child.wait())
```

[Cmd.stop](stop.md) returns the status as `Cmd.wait` does, so a wait after a stop agrees with it:

```ant
let child = try Cmd.new("sleep", ["30"]).start()
let stopped = try child.stop()
assert(try child.wait() == stopped)
```

## Possible issues

### A pipe nobody reads

A child that fills a pipe blocks until someone reads it, so a wait before the read never ends. Here
a bounded wait shows it, and reading the pipe lets the child finish:

```ant
let out = try Os.openPipe()
defer { out.read.close() }
let child = try Cmd.new("seq", ["100000"]).start(stdout: out.write)
try out.write.close()
println(child.wait(ctx: Context.new(timeout: Time.duration(milliseconds: 300))))
let text = try out.read.readAll()
println(text.lines().length())
println(try child.wait())
```

```text
Err{ error: Error{ message: "timeout" } }
100000
0
```

### `try` does not reject a failed command

The wait is `Ok` of a nonzero status, so check the status where failure matters:

```ant
func build(): Result[void, Error] {
  let child = try Cmd.new("sh", ["-c", "exit 2"]).start()
  let status = try child.wait()
  if status != 0 {
    return Result.fail("build: exit status ${status}")
  }
  Result.ok(void)
}

println(build())
```

```text
Err{ error: Error{ message: "build: exit status 2" } }
```

### A Result is not a process

`Cmd.wait` takes the process unwrapped from the `Result` of [Cmd.start](start.md), and its own
result is a `Result` too:

```ant
let started = Cmd.new("true").start()
println(started.wait()) // error: Result[Cmd.Process, Error] has no field or function wait
let child = try Cmd.new("false").start()
let status: int = child.wait() // error: cannot use Result[int, Error] where int is expected
println(status)
```

## See also

[Cmd.start](start.md) · [Cmd.stop](stop.md) · [Cmd.kill](kill.md) · [Cmd.pid](pid.md) ·
[Cmd.run](run.md) · [Cmd.capture](capture.md) · [Context.new](../Context/new.md) · [Cmd](index.md)
