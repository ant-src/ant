<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.interrupt

```ant signature
func Os.interrupt(): Context
```

Os.interrupt returns a context that is done, with the reason "interrupt", when the process is
interrupted by Ctrl-C, SIGINT, or SIGTERM. It is the same context at every call. Once it has been
called, the first interrupt no longer ends the run; a second does, as does the first after the
program cancels the context itself.

<!-- end generated -->

## Details

- The program has to watch the context and finish on its own. If it cancels the context itself, the
  reason becomes `"cancelled"` and the next interrupt ends the run.
- When an interrupt ends the run, `ant` prints nothing and ends by that signal, so a shell sees
  status 130 for Ctrl-C and SIGINT and 143 for SIGTERM.
- While a goroutine waits in [Cmd.run](../Cmd/run.md), `ant` holds back SIGINT: the child, or a
  pipeline's last stage, gets Ctrl-C too. The context is done only when that child exits because of
  the signal, so quitting `less` with Ctrl-C does not end the script. SIGTERM is not held back.
  Holding back is process-wide, so a server that waits for `Os.interrupt()` does not call `Cmd.run`.

## Basic examples

Wait for Ctrl-C, then finish the work and exit:

<!-- test: before
go {
  Time.sleep(Time.duration(milliseconds: 100))
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
println("running, press Ctrl-C to stop")
Os.interrupt().wait()
println("stopping")
```

```text
running, press Ctrl-C to stop
stopping
```

Before any interrupt the context is not done and has no reason. After one, its reason is
`"interrupt"`:

<!-- test: before
go {
  Time.sleep(Time.duration(milliseconds: 100))
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
let stop = Os.interrupt()
println(stop.isDone())
println(stop.reason())
stop.wait()
println(stop.reason())
```

```text
false
null
interrupt
```

## Scope

Every call returns the same context, so a library and the script that uses it watch one interrupt:

```ant
let stop = Os.interrupt()
println(Os.interrupt() == stop)
```

```text
true
```

SIGTERM, which `kill` and service managers send, makes the context done as Ctrl-C does. Here the
program sends it to itself:

```ant
let stop = Os.interrupt()
try Cmd.read(Cmd.new("kill", ["-TERM", "${Os.pid()}"]))
stop.wait()
println(stop.reason())
```

```text
interrupt
```

## Applications

Run a loop until Ctrl-C. Passing the context to [Time.sleep](../Time/sleep.md) ends the pause at
once instead of after the full ten seconds:

<!-- test: before
go {
  Time.sleep(Time.duration(milliseconds: 100))
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
let stop = Os.interrupt()
var rounds = 0
while !stop.isDone() {
  rounds += 1
  Time.sleep(Time.duration(seconds: 10), ctx: stop)
}
println("stopped after ${rounds} round")
```

```text
stopped after 1 round
```

Serve HTTP until Ctrl-C. [Http.serve](../Http/serve.md) returns `Ok` once the context is done and
the requests in flight have finished:

<!-- test: before
go {
  Time.sleep(Time.duration(milliseconds: 100))
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
let l = try Net.listen("127.0.0.1:0")
let served = Http.serve(l, { Http.Reply{ status: 200, headers: Map{}, body: "hi" } }, ctx: Os.interrupt())
println(served)
```

```text
Ok{ value: void }
```

Give the shutdown a limit of its own. A child of the interrupt context is done at Ctrl-C, with the
reason `"interrupt"`, or when its timeout passes, with `"timeout"`:

```ant
let job = Context.new(timeout: Time.duration(milliseconds: 50), parent: Os.interrupt())
job.wait()
println(job.reason())
```

```text
timeout
```

Drain a queue of jobs, and stop taking new ones at Ctrl-C. [Channel.values](../Channel/values.md)
ends when the context is done:

<!-- test: before
go {
  Time.sleep(Time.duration(milliseconds: 100))
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
let stop = Os.interrupt()
let jobs: Channel[string] = Channel.new(capacity: 2)
jobs.send("resize")
jobs.send("upload")
for job in jobs.values(ctx: stop) {
  println("done ${job}")
}
println(stop.reason())
```

```text
done resize
done upload
interrupt
```

## Properties and relations

The context is the same at every call, and an interrupt gives it the reason `"interrupt"`:

```ant
let stop = Os.interrupt()
assert(stop == Os.interrupt())
assert(stop.reason() == null)
try Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"]))
stop.wait()
assert(stop.reason() == "interrupt")
assert(Os.interrupt().isDone())
```

## Possible issues

A program that never calls `Os.interrupt` ends at the first Ctrl-C. This one prints `start` and
exits with status 130, and nothing after the interrupt runs:

<!-- test: check-only the run ends by SIGINT, which the docs test cannot expect as status 130 -->

```ant
println("start")
try Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"]))
Time.sleep(Time.duration(seconds: 1))
println("not reached")
```

A program that calls `Os.interrupt` and never watches the context ignores the first Ctrl-C. Only the
second one ends it:

```ant
Os.interrupt()
try Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"]))
Time.sleep(Time.duration(milliseconds: 100))
println("still running")
```

```text
still running
```

The second interrupt ends the run even while the program is shutting down. This one prints
`cleaning up` and exits with status 130:

<!-- test: check-only the run ends by SIGINT, which the docs test cannot expect as status 130 -->

```ant
let stop = Os.interrupt()
try Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"]))
stop.wait()
println("cleaning up")
try Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"]))
Time.sleep(Time.duration(seconds: 1))
println("not reached")
```

Cancelling the context yourself gives it the reason `"cancelled"`, and the next interrupt ends the
run at once. This one prints `cancelled` and exits with status 130:

<!-- test: check-only the run ends by SIGINT, which the docs test cannot expect as status 130 -->

```ant
let stop = Os.interrupt()
stop.cancel()
println(stop.reason())
try Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"]))
Time.sleep(Time.duration(seconds: 1))
println("not reached")
```

While `Cmd.run` waits, Ctrl-C goes to the child, and the context becomes done only if the child
exits because of it. A pager quit with Ctrl-C leaves the context running, so this prints `false`.
The holding back is process-wide, so a server that waits for `Os.interrupt()` does not call
`Cmd.run`:

<!-- test: check-only it needs a terminal and a press of Ctrl-C inside the pager -->

```ant
let stop = Os.interrupt()
try Cmd.run(Cmd.new("less", ["notes.txt"]))
println(stop.isDone())
```

## See also

[Context.wait](../Context/wait.md) · [Context.reason](../Context/reason.md) ·
[Context.new](../Context/new.md) · [Os.exit](exit.md) · [Cmd.run](../Cmd/run.md) ·
[Http.serve](../Http/serve.md)
