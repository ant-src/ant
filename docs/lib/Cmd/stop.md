<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.stop

```ant signature
func Cmd.stop(p: Cmd.Process, grace: Time.Duration = Time.duration(seconds: 5)): Result[int, Error]
```

Cmd.stop ends p: SIGTERM, then SIGKILL when it is still running after grace, or SIGKILL alone when
grace is zero or negative, and returns its status as [`Cmd.wait`](wait.md) does. It panics when
grace has years, months or weeks.

<!-- end generated -->

## Details

- `Cmd.stop` signals the whole process group that [`Cmd.start`](start.md) made, so every stage of a
  pipeline gets the signal.
- It waits for the command to exit and returns its status. A command ended by a signal has
  `128 + n`: `143` for SIGTERM, `137` for SIGKILL. That status is `Ok`, not `Err`.
- On a command that has exited already it sends nothing and returns its status, as
  [`Cmd.wait`](wait.md) does. Calling it twice is safe.
- A command still running when the script ends is left running. Stop it in a deferred block.
- On Windows a command that `Cmd.stop` ends has status `1`.

## Basic examples

`Cmd.stop(p)` sends SIGTERM and returns the status. `sleep` dies of it, `128 + 15`:

```ant
let p = try Cmd.start(Cmd.new("sleep", ["30"]))
println(try Cmd.stop(p))
```

```text
143
```

On a command that has exited, `Cmd.stop` sends nothing and returns the status it exited with:

```ant
let p = try Cmd.start(Cmd.new("sh", ["-c", "exit 3"]))
println(try p.wait())
println(try Cmd.stop(p))
```

```text
3
3
```

## Scope

Call it with a dot on the process:

```ant
let p = try Cmd.start(Cmd.new("sleep", ["30"]))
println(try p.stop())
```

```text
143
```

A pipeline's stages share one process group, so one call stops them all. The status is that of the
rightmost stage that failed:

```ant
let p = try Cmd.start(Cmd.pipe([Cmd.new("sleep", ["30"]), Cmd.new("cat")]))
println(try p.stop())
```

```text
143
```

A command that handles SIGTERM can finish its work and exit with its own status. This shell script
traps the signal, prints a line, and exits with `0`. The program reads `ready` first, so the trap is
in place before the signal arrives:

```ant
let out = try Os.openPipe()
defer { out.read.close() }
let script = "trap 'echo saving; exit 0' TERM; echo ready; while true; do sleep 1; done"
let server = try Cmd.start(Cmd.new("sh", ["-c", script]), stdout: out.write)
try out.write.close()
println(try out.read.readLine())
println(try server.stop())
println(try out.read.readLine())
```

```text
ready
0
saving
```

## Named arguments

`grace:` is how long `Cmd.stop` waits after SIGTERM before it sends SIGKILL. This script ignores
SIGTERM, so it runs until the grace ends and SIGKILL ends it, `128 + 9`:

```ant
let out = try Os.openPipe()
defer { out.read.close() }
let script = "trap '' TERM; echo ready; while true; do sleep 1; done"
let p = try Cmd.start(Cmd.new("sh", ["-c", script]), stdout: out.write)
try out.write.close()
println(try out.read.readLine())
println(try p.stop(grace: Time.duration(milliseconds: 200)))
```

```text
ready
137
```

A `grace` of zero, or a negative one, sends SIGKILL alone, with no SIGTERM before it:

```ant
let p = try Cmd.start(Cmd.new("sleep", ["30"]))
println(try p.stop(grace: Time.duration(seconds: 0)))
```

```text
137
```

## Applications

Stop a helper process when the script ends, however it ends. A deferred `stop` runs on `return`, on
`Err`, and on a panic:

```ant
let server = try Cmd.start(Cmd.new("sleep", ["30"]))
defer { server.stop(grace: Time.duration(seconds: 1)) }
println("server ${server.pid() > 0}")
```

```text
server true
```

Run a command with a time limit. A context on [`Cmd.wait`](wait.md) ends only the wait, and the
command goes on running, so stop it once the wait fails:

```ant
func runFor(c: Cmd, limit: Time.Duration): Result[int, Error] {
  let p = try Cmd.start(c)
  let ctx = Context.new(timeout: limit)
  defer { ctx.cancel() }
  let status = p.wait(ctx: ctx)
  if status.isOk() { return status }
  println("timed out: ${c.string()}")
  p.stop(grace: Time.duration(seconds: 1))
}

println(try runFor(Cmd.new("sh", ["-c", "exit 4"]), Time.duration(seconds: 5)))
println(try runFor(Cmd.new("sleep", ["30"]), Time.duration(milliseconds: 100)))
```

```text
4
timed out: sleep 30
143
```

## Properties and relations

`Cmd.stop` returns what [`Cmd.wait`](wait.md) returns afterward, and a second `stop` sends nothing
and returns the same status:

```ant
let p = try Cmd.start(Cmd.new("sleep", ["30"]))
let status = try p.stop()
assert(try p.wait() == status)
assert(try p.stop() == status)
println(status)
```

```text
143
```

## Possible issues

A command that SIGTERM ends has a nonzero status. `Cmd.stop` returns it as `Ok`, so check the status
when it matters:

```ant
let p = try Cmd.start(Cmd.new("sleep", ["30"]))
let status = try p.stop()
println(if status == 0 { "clean exit" } else { "status ${status}" })
```

```text
status 143
```

A `grace` with years, months or weeks has no fixed length, so `Cmd.stop` panics:

```ant
let p = try Cmd.start(Cmd.new("sleep", ["1"]))
let status = p.stop(grace: Time.duration(weeks: 1))
// panic: Cmd.stop needs a duration without years, months or weeks, which have no fixed length
println(status)
```

## See also

[Cmd.kill](kill.md), [Cmd.wait](wait.md), [Cmd.start](start.md), [Cmd.pid](pid.md),
[Cmd.Process](types/Process.md), [Context.new](../Context/new.md),
[Time.duration](../Time/duration.md)
