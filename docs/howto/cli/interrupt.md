<!-- status: reviewed -->

# Handle Ctrl-C

By default the first Ctrl-C ends an Ant program at once. To finish the current piece of work and
save what you have instead, ask for the interrupt and watch it.

<!-- test: before
go {
  Time.sleep(100ms)
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
let stop = Os.interrupt()
var batch = 0
while !stop.isDone() {
  batch += 1
  println("processed batch ${batch}")
  Time.sleep(10s, ctx: stop)
}
println("interrupted, saving progress at batch ${batch}")
```

```text
processed batch 1
interrupted, saving progress at batch 1
```

1. Call [Os.interrupt](../../lib/Os/interrupt.md) before the work starts. From then on the first
   Ctrl-C, SIGINT, or SIGTERM no longer ends the run: it makes the returned context done.
2. Check [Context.isDone](../../lib/Context/isDone.md) between pieces of work, so a piece already
   started runs to its end.
3. Pass the context as `ctx:` to every wait, such as [Time.sleep](../../lib/Time/sleep.md), so the
   wait ends at Ctrl-C instead of running its full length.
4. After the loop, write what the program has, then let the script end.

A second Ctrl-C ends the run even while the program saves, so a stuck cleanup can still be stopped.

## Exit with the interrupt status

A shell sees status 130 when Ctrl-C ends a program. To give the same status after cleaning up, exit
with [Os.exit](../../lib/Os/exit.md). Unlike a run the interrupt ends, this exit does not stop a
shell loop that runs the program:

<!-- test: exit 130 -->
<!-- test: before
go {
  Time.sleep(100ms)
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
let stop = Os.interrupt()
stop.wait()
println("cleaned up")
Os.exit(130)
```

```text
cleaned up
```

[Context.wait](../../lib/Context/wait.md) blocks until the interrupt comes. Os.exit runs no deferred
block, so finish the cleanup before you call it.

## Stop a worker that waits on a channel

To stop a worker that waits on a channel, wait in a `select` for either the next job or
[Context.wait](../../lib/Context/wait.md) on the interrupt. A
[Channel.receive](../../lib/Channel/receive.md) with `ctx:` gives up too, but it returns `null` for
a close and for the interrupt alike, and the `select` tells them apart:

<!-- test: before
go {
  Time.sleep(100ms)
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
let stop = Os.interrupt()
let jobs: Channel[string] = Channel.new()
go {
  jobs.send("a.txt")
  jobs.send("b.txt")
}
while true {
  select {
    let job = Channel.receive(jobs) { println("processing ${job ?? "nothing"}") }
    Context.wait(stop) { break }
  }
}
println("worker stopped")
```

```text
processing a.txt
processing b.txt
worker stopped
```

## Stop a server

Open a listener with [Net.listen](../../lib/Net/listen.md) and pass the context to
[Http.serve](../../lib/Http/serve.md), and print [Net.address](../../lib/Net/address.md) to see the
port. Http.serve returns `Ok` once the context is done and the requests in flight have finished,
waiting at most five seconds before it closes their connections:

<!-- test: before
go {
  Time.sleep(100ms)
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
func hello(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: "hello" }
}

let l = try Net.listen("127.0.0.1:0")
println("serving on ${l.address()}")
try Http.serve(l, hello, ctx: Os.interrupt())
println("server stopped")
```

A server that waits for `Os.interrupt()` does not call `Cmd.run`: while `Cmd.run` waits, Ctrl-C goes
to the child, and the context is done only if the child exits because of it.
