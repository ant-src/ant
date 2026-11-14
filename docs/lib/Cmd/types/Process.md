<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.Process

```ant signature
type Cmd.Process
```

Cmd.Process is an opaque command started by [`Cmd.start`](../start.md), waited for with
[`Cmd.wait`](../wait.md).

<!-- end generated -->

## Details

- A process comes only from [Cmd.start](../start.md). It has no literal. A started pipeline is one
  process: its stages share a process group of their own.
- Four functions take it: [Cmd.wait](../wait.md), [Cmd.kill](../kill.md), [Cmd.stop](../stop.md),
  and [Cmd.pid](../pid.md).
- `Cmd.wait` may be called any number of times, from any goroutine, and each call gets the same
  outcome. `Cmd.kill` and `Cmd.stop` send nothing to a command that has exited.
- A command still running when the script ends is left running. Stop it in a deferred block.
- Two processes are equal only when they are the same process. `String.from` renders every process
  as `Cmd.Process`.

## Basic examples

Start a command in the background with [Cmd.start](../start.md), and wait for its status with
[Cmd.wait](../wait.md):

```ant
let process: Cmd.Process = try Cmd.start(Cmd.new("sh", ["-c", "sleep 0.1; echo working; exit 3"]))
println("started")
println(Cmd.wait(process))
```

```text
started
working
Ok{ value: 3 }
```

Stop a command you started when you are done with it. The deferred block stops it on every path out
of the script:

```ant
let server = try Cmd.new("sleep", ["30"]).start()
defer { server.stop() }
println("serving")
```

```text
serving
```

## Scope

### Several waits

Every wait gets the same status, also from goroutines that wait at once:

```ant
let job = try Cmd.new("sh", ["-c", "sleep 0.1; exit 4"]).start()
let statuses: Channel[int] = Channel.new(2)
for _ in [1, 2] {
  go { statuses.send(job.wait().unwrap()) }
}
println(statuses.receive())
println(statuses.receive())
println(job.wait())
```

```text
4
4
Ok{ value: 4 }
```

### The process id

[Cmd.pid](../pid.md) gives the id of a started command, to log or report it. For a started pipeline
it is the last stage's:

```ant
let server = try Cmd.new("sleep", ["30"]).start()
defer { server.stop() }
println(server.pid() > 0)
```

```text
true
```

### A pipeline

A started pipeline is one process. A wait gives the pipeline's status, and a kill reaches every
stage:

```ant
let pipeline = try Cmd.pipe([Cmd.new("sleep", ["30"]), Cmd.new("cat")]).start()
try pipeline.kill()
println(pipeline.wait())
```

```text
Ok{ value: 143 }
```

### Output as it comes

Give the process the write end of a pipe, and read lines while it runs. Close the program's copy of
the write end, so the read ends when the command exits:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
let counter = try Cmd.new("sh", ["-c", "for i in 1 2 3; do echo $i; sleep 0.05; done"])
  .start(stdout: pipe.write)
try pipe.write.close()
while true {
  let line = try pipe.read.readLine()
  if line == null { break }
  println("got ${line}")
}
println(counter.wait())
```

```text
got 1
got 2
got 3
Ok{ value: 0 }
```

### A process or null

A process may be absent, as `Cmd.Process | null`. `is` cannot test a process, so narrow by testing
`null`:

```ant
func stopIfStarted(p: Cmd.Process | null): string {
  if p == null { return "nothing to stop" }
  "stopped with ${p.stop().unwrap()}"
}

println(stopIfStarted(null))
println(stopIfStarted(try Cmd.new("sleep", ["30"]).start()))
```

```text
nothing to stop
stopped with 143
```

### A context

A context given to `Cmd.start` bounds the command's life. When it ends, the command is killed and
every wait is `Err`:

```ant
let ctx = Context.new(timeout: 100ms)
let slow = try Cmd.new("sleep", ["30"]).start(ctx: ctx)
println(slow.wait())
```

```text
Err{ error: Error{ message: "timeout" } }
```

A context given to `Cmd.wait` bounds only that wait. The command goes on running:

```ant
let slow = try Cmd.new("sleep", ["30"]).start()
println(slow.wait(ctx: Context.new(timeout: 100ms)))
println(slow.stop(grace: 0s))
```

```text
Err{ error: Error{ message: "timeout" } }
Ok{ value: 137 }
```

## Applications

### Run commands side by side

Start every command first, then wait for each. The commands run at once, so the whole takes as long
as the slowest:

```ant
let scripts = ["sleep 0.2; exit 0", "sleep 0.2; exit 1", "sleep 0.2; exit 0"]
let started = Time.monotonic()
let jobs: Array[Cmd.Process] = []
for script in scripts {
  jobs.push(try Cmd.new("sh", ["-c", script]).start())
}
let statuses: Array[int] = []
for job in jobs {
  statuses.push(try job.wait())
}
println(statuses)
println(Time.monotonic() - started < 500)
```

```text
[0, 1, 0]
true
```

### Give a command a deadline

Wait for a while, and stop the command if it has not exited by then:

```ant
func finish(p: Cmd.Process, limit: Time.Duration): Result[int, Error] {
  let waited = p.wait(ctx: Context.new(timeout: limit))
  if waited is Ok { return waited }
  p.stop(grace: 1s)
}

let quick = try Cmd.new("sh", ["-c", "exit 2"]).start()
let stuck = try Cmd.new("sleep", ["30"]).start()
println(finish(quick, 200ms))
println(finish(stuck, 200ms))
```

```text
Ok{ value: 2 }
Ok{ value: 143 }
```

## Properties and relations

A process is a reference: a copy is the same process, and another start makes another one:

```ant
let a = try Cmd.new("sleep", ["30"]).start()
defer { a.stop() }
let b = try Cmd.new("sleep", ["30"]).start()
defer { b.stop() }
let same = a
assert(same == a)
assert(a != b)
assert(String.from(a) == "Cmd.Process")
```

Once a command has exited, a wait, a stop, and a kill change nothing, and the status stays the same:

```ant
let p = try Cmd.new("sh", ["-c", "exit 5"]).start()
let status = try p.wait()
assert(try p.stop() == status)
assert(p.kill() is Ok)
assert(try p.wait() == status)
```

## Possible issues

### A command that cannot start gives no process

`Cmd.start` returns a `Result`. A program that is not found is `Err`, and there is no process to
wait for:

```ant
let started = Cmd.new("no-such-program").start()
println(started is Err)
```

```text
true
```

### A `Result` is not a process

Unwrap the result of `Cmd.start` before calling a function of `Cmd.Process`:

```ant
let started = Cmd.new("sleep", ["5"]).start()
println(started.wait()) // error: Result[Cmd.Process, Error] has no field or function wait
```

### No literal, no `is`

`Cmd.Process` is not a record, so it has no literal, and `is` cannot test it:

```ant
let p = Cmd.Process{} // error: Cmd.Process is not a record type, so it has no literal
```

```ant
func describe(x: Cmd.Process | string): string {
  if x is Cmd.Process { "process" } else { x }
  // error: is cannot test Cmd.Process; narrow by testing the other members
}
```

### A wait before reading the pipe can block forever

A child that writes into a pipe nobody reads blocks once the pipe is full, and a `Cmd.wait` before
the pipe is read then never returns. Read a child's pipe while it waits, before the wait or on
another goroutine, as in [Output as it comes](#output-as-it-comes).

### No JSON form

A process has no JSON form, so `Json.stringify` panics on one:

```ant
let p = try Cmd.new("true").start()
println(p.wait())
let json = Json.stringify(p) // panic: Json.stringify: Cmd.Process has no JSON form
```

## See also

[Cmd.start](../start.md) · [Cmd.wait](../wait.md) · [Cmd.kill](../kill.md) · [Cmd.stop](../stop.md)
· [Cmd.pid](../pid.md) · [Os.openPipe](../../Os/openPipe.md) · [Cmd.Pipeline](Pipeline.md) ·
[Cmd](../index.md)
