<!-- status: reviewed -->

# Stop or time out a process

A command runs until it exits. To give it a time limit, pass a context with a timeout. This script
gives a command 200 milliseconds:

<!-- test: exit 1 -->

```ant
let check = Cmd.new("sleep", ["5"])
let status = try (Cmd.run(check, ctx: Context.new(200ms)) | Result.wrapErr(check.string()))
println("exit status ${status}")
```

```sh
$ ant run check.ant; echo $?
sleep 5: timeout
1
```

1. Build the command with [Cmd.new](../../lib/Cmd/new.md).
2. Create the context with [Context.new](../../lib/Context/new.md) right before the call. Its time
   limit starts when you create it.
3. Pass the context as `ctx:` to [Cmd.run](../../lib/Cmd/run.md), [Cmd.read](../../lib/Cmd/read.md),
   or [Cmd.capture](../../lib/Cmd/capture.md). When the time runs out, the call kills the command
   and returns `Err` with the message `timeout`.
4. Name the command in the error with [Result.wrapErr](../../lib/Result/wrapErr.md) and
   [Cmd.string](../../lib/Cmd/string.md), and parenthesize the pipe under `try`.

A command that exits in time gives its status, as if there were no context. The timeout kills the
command itself, not the processes it started, so the children of a shell script run on. To stop
those too, start the command in the background, as the next sections do.

## Stop a background process

A command started with [Cmd.start](../../lib/Cmd/start.md) runs on after the script ends. Stop it in
a deferred block right after you start it. [Cmd.pid](../../lib/Cmd/pid.md) gives its process id:

```ant
let server = try Cmd.start(Cmd.new("sleep", ["30"]))
defer { server.stop(grace: 1s) }
println("server running: ${server.pid() > 0}")
```

```text
server running: true
```

[Cmd.stop](../../lib/Cmd/stop.md) sends SIGTERM, sends SIGKILL if the command still runs after
`grace:`, and returns its status. It signals every process in the command's process group, so a
pipeline's stages and the children of a shell script get the SIGTERM too.

## Give a background process a time limit

Wait with a timeout, and stop the command when the wait fails. A context on
[Cmd.wait](../../lib/Cmd/wait.md) ends only the wait, so the command still gets its SIGTERM and its
grace. Test the wait's result with [Result.isOk](../../lib/Result/isOk.md):

```ant
let job = try Cmd.start(Cmd.new("sleep", ["30"]))
defer { job.stop(grace: 1s) }
let status = job.wait(ctx: Context.new(200ms))
if status.isOk() {
  println("finished with ${try status}")
  return
}
println("still running after 200ms")
println("stopped with ${try job.stop(grace: 1s)}")
```

```text
still running after 200ms
stopped with 143
```

Where the host has signals, a command that signal `n` ended has the status `128 + n`: `143` for
SIGTERM, `137` for SIGKILL. On Windows it has status 1.

To kill the command outright when the time runs out, pass the context to `Cmd.start` instead. Every
process in its group is killed, and every wait returns `Err`:

```ant
let script = "sleep 30 & sleep 30; wait"
let job = try Cmd.start(Cmd.new("sh", ["-c", script]), ctx: Context.new(200ms))
println(job.wait())
```

```text
Err{ error: Error{ message: "timeout" } }
```

## Stop a background process on Ctrl-C

The terminal's Ctrl-C does not reach a command that `Cmd.start` started, since it runs in a process
group of its own. Wait for the command with the context from
[Os.interrupt](../../lib/Os/interrupt.md), and stop it when the interrupt comes:

<!-- test: before
go {
  Time.sleep(100ms)
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
let stop = Os.interrupt()
let worker = try Cmd.start(Cmd.new("sleep", ["30"]))
defer { worker.stop(grace: 2s) }
let status = worker.wait(ctx: stop)
if status.isOk() {
  println("worker exited with ${try status}")
  return
}
println("interrupted, stopping the worker")
println("worker stopped with ${try worker.stop(grace: 2s)}")
```

```text
interrupted, stopping the worker
worker stopped with 143
```

A command that `Cmd.run` runs shares the terminal, so Ctrl-C reaches it directly, and the command
decides what to do with it. See [Handle Ctrl-C](../cli/interrupt.md).

## Stop on a time limit or Ctrl-C

Give the timeout context the interrupt as its `parent:`. The wait then ends on the timeout or on
Ctrl-C, whichever comes first, and [Context.reason](../../lib/Context/reason.md) tells which one it
was:

```ant
let ctx = Context.new(200ms, parent: Os.interrupt())
let worker = try Cmd.start(Cmd.new("sleep", ["30"]))
defer { worker.stop(grace: 2s) }
let status = worker.wait(ctx: ctx)
if status.isOk() {
  println("worker exited with ${try status}")
  return
}
println("wait ended: ${Context.reason(ctx) ?? "?"}")
println("worker stopped with ${try worker.stop(grace: 2s)}")
```

```text
wait ended: timeout
worker stopped with 143
```

## Send another signal

[Cmd.kill](../../lib/Cmd/kill.md) sends one signal by name and does not wait. Wait for the command
afterward to get its status:

```ant
let job = try Cmd.start(Cmd.new("sleep", ["30"]))
try job.kill("SIGINT")
println(try job.wait())
```

```text
130
```

It takes `SIGTERM`, `SIGKILL`, `SIGINT`, `SIGHUP`, `SIGQUIT`, `SIGUSR1`, and `SIGUSR2`. Any other
name returns `Err`.

## See also

[Cmd.run](../../lib/Cmd/run.md) · [Cmd.start](../../lib/Cmd/start.md) ·
[Cmd.wait](../../lib/Cmd/wait.md) · [Cmd.stop](../../lib/Cmd/stop.md) ·
[Cmd.kill](../../lib/Cmd/kill.md) · [Cmd.pid](../../lib/Cmd/pid.md) ·
[Context.new](../../lib/Context/new.md) · [Context.reason](../../lib/Context/reason.md) ·
[Os.interrupt](../../lib/Os/interrupt.md) · [Result.isOk](../../lib/Result/isOk.md) ·
[Result.wrapErr](../../lib/Result/wrapErr.md) · [Run a command and capture its output](run.md) ·
[Handle Ctrl-C](../cli/interrupt.md)
