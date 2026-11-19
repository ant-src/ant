<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Context.wait

```ant signature
func Context.wait(ctx: Context): void
```

Context.wait blocks until ctx is done.

<!-- end generated -->

## Details

- `wait` returns at once when `ctx` is done already.
- A context that nothing cancels and that has no deadline is done only when its parent is, so a wait
  on it can last until the run ends.
- `wait` takes no context of its own. To bound it, wait on a child with a timeout instead.
- Every write a goroutine makes before `Context.cancel` is visible after the `wait` that the cancel
  ends returns.
- `Context.wait(ctx)` is a head of `select`, which can proceed once the context is done.

## Basic examples

Wait for a context's timeout to pass:

```ant
let ctx = Context.new(timeout: 20ms)
Context.wait(ctx)
println(ctx.reason())
```

```text
timeout
```

Wait until another goroutine cancels the context:

```ant
let ctx = Context.new()
go {
  Time.sleep(20ms)
  ctx.cancel()
}
Context.wait(ctx)
println(ctx.reason())
```

```text
cancelled
```

## Scope

A context that is done already has nothing to wait for, so `wait` returns at once:

```ant
let ctx = Context.new()
ctx.cancel()
ctx.wait()
println("returned at once")
```

```text
returned at once
```

A child is done when its parent is, with the parent's reason, so a wait on the child ends when the
parent is cancelled:

```ant
let parent = Context.new()
let child = Context.new(timeout: 1h, parent: parent)
go {
  Time.sleep(20ms)
  parent.cancel()
}
child.wait()
println(child.reason())
```

```text
cancelled
```

Any number of goroutines can wait on one context, and one cancel releases all of them. Here the
context works as a starting gate:

```ant
let gate = Context.new()
let started = Atomic.new(0)
let runners = WaitGroup.new()
for _ in Array.range(0, 3) {
  runners.add(1)
  go {
    defer { runners.done() }
    gate.wait()
    started.add(1)
  }
}
println("started before the gate opens: ${started.load()}")
gate.cancel()
runners.wait()
println("started after: ${started.load()}")
```

```text
started before the gate opens: 0
started after: 3
```

The context of [Os.interrupt](../Os/interrupt.md) is done at Ctrl-C, so a wait on it lasts until the
user stops the program:

<!-- test: before
go {
  Time.sleep(100ms)
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
let stop = Os.interrupt()
println("press Ctrl-C to stop")
stop.wait()
println("stopped: ${stop.reason()}")
```

```text
press Ctrl-C to stop
stopped: interrupt
```

In a `select`, `Context.wait(ctx)` is a head that can proceed once the context is done. Beside a
receive it ends the wait for a value that does not come in time:

```ant
let replies: Channel[string] = Channel.new()
let ctx = Context.new(timeout: 20ms)
let reply = select {
  let r = Channel.receive(replies) { r ?? "closed" }
  Context.wait(ctx) { "no reply: ${ctx.reason()}" }
}
println(reply)
```

```text
no reply: timeout
```

## Applications

### Serve until Ctrl-C

Run a server on its own goroutine and keep the main goroutine waiting for the interrupt. The server
sees the same context, and the main goroutine waits for it to finish the requests in flight:

<!-- test: before
go {
  Time.sleep(100ms)
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
func hello(req: Http.Request): Http.Reply {
  Http.Reply{ status: 200, headers: Map{}, body: "hi" }
}

let stop = Os.interrupt()
let listener = try Net.listen("127.0.0.1:0")
let served = WaitGroup.new()
served.add(1)
go {
  defer { served.done() }
  _ = Http.serve(listener, hello, ctx: stop)
}
println("serving")
stop.wait()
println("shutting down")
served.wait()
```

```text
serving
shutting down
```

### Stop a worker loop

A worker that waits on several things writes a `select` in a loop, with a `Context.wait` clause that
returns when the work is cancelled:

```ant
func work(jobs: Channel[string], ctx: Context) {
  while true {
    select {
      Context.wait(ctx) {
        println("stopped: ${ctx.reason()}")
        return
      }
      let job = Channel.receive(jobs) {
        if job == null { return }
        println("done ${job}")
      }
    }
  }
}

let jobs: Channel[string] = Channel.new()
let ctx = Context.new()
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  work(jobs, ctx)
}
for job in ["resize", "upload"] {
  jobs.send(job)
}
ctx.cancel()
wg.wait()
```

```text
done resize
done upload
stopped: cancelled
```

### Wait with a limit

`wait` takes no context, so bound it with a child that has a timeout. The child is done when the
parent is or when the time passes, and the parent's `isDone` tells which:

```ant
func waitAtMost(ctx: Context, limit: Time.Duration): bool {
  Context.new(timeout: limit, parent: ctx).wait()
  ctx.isDone()
}

let job = Context.new()
println(waitAtMost(job, 20ms))
job.cancel()
println(waitAtMost(job, 1s))
```

```text
false
true
```

## Properties and relations

Once `wait` returns, the context is done and has a reason:

```ant
let ctx = Context.new(timeout: 10ms)
ctx.wait()
assert(ctx.isDone())
assert(ctx.reason() == "timeout")
```

What a goroutine writes before it cancels a context is visible after the wait returns, so a cancel
can hand over a result:

```ant
let result = ["pending"]
let ready = Context.new()
go {
  result[0] = "computed"
  ready.cancel()
}
ready.wait()
assert(result[0] == "computed")
```

`Context.wait(ctx)` waits as a `select` with that one head does:

```ant
let a = Context.new(timeout: 10ms)
a.wait()
let b = Context.new(timeout: 10ms)
select {
  Context.wait(b) {}
}
assert(a.reason() == b.reason())
```

## Possible issues

### A context nothing ends

A context without a timeout is done only when a call cancels it or its parent is done. If nothing
does, a wait on it lasts until the run ends. Cancelling a child does not end the parent either. Here
a bounded wait shows that the parent is still running:

```ant
let parent = Context.new()
let child = Context.new(parent: parent)
child.cancel()
Context.new(timeout: 20ms, parent: parent).wait()
println(parent.isDone())
```

```text
false
```

### No context argument

The one parameter of `wait` is `ctx`, the context it waits for, so `ctx:` names it a second time. To
bound the wait, wait on a child with a timeout, as in Wait with a limit:

```ant
let job = Context.new()
let limit = Context.new(timeout: 1s)
Context.wait(job, ctx: limit) // error: the parameter ctx is given twice
```

### Heads in a select

A `select` head is written in the namespace form. `ctx.wait()` is the same call elsewhere, but in a
`select` it is a syntax error:

```ant
let ctx = Context.new(timeout: 10ms)
select {
  ctx.wait() { println("done") }
  // error: a select clause's head must be a call of Channel.receive, Channel.send, Context.wait, WaitGroup.wait or Time.sleep
}
```

A `Context.wait` head binds nothing, since `wait` returns `void`:

```ant
let ctx = Context.new(timeout: 10ms)
select {
  let r = Context.wait(ctx) { println("done") } // error: Context.wait returns void, so there is nothing to bind to r
}
```

## See also

[Context.new](new.md) · [Context.cancel](cancel.md) · [Context.isDone](isDone.md) ·
[Context.reason](reason.md) · [Context](index.md) · [Os.interrupt](../Os/interrupt.md) ·
[WaitGroup.wait](../WaitGroup/wait.md) · [Time.sleep](../Time/sleep.md) ·
[select](../../language/keywords/select.md) · [Concurrency guide](../../guide/concurrency.md)
