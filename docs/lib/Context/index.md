<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Context

```ant signature
type Context
```

Context bounds a wait with a deadline or ends it with a cancellation. A context becomes done once:
when it is cancelled, when its deadline passes, or when its parent is done. Every blocking library
call takes one as its last parameter. Contexts compare by identity.

<!-- end generated -->

A context tells blocking calls when to stop waiting. You make one with `Context.new`, optionally
with a timeout and a parent, pass it to the calls that wait, and cancel it when the work is no
longer wanted. The `Context` namespace holds the functions that make a context, cancel it, and ask
whether and why it is done; `Context` is also the type, for an annotation, a parameter, or a record
field.

## Details

- A context is done once and stays done. Its reason is `"cancelled"`, `"timeout"`, `"interrupt"`,
  or, for a child, the reason of the parent that ended it.
- Every context is a child of a parent or of the run. When a parent is done, its children are done
  too; cancelling a child leaves its parent running.
- A blocking call that takes `ctx:` gives up when the context is done. A call that returns a
  `Result` gives `Err` with the reason as its message, a call that returns a value or `null` gives
  `null`, and a `void` call returns early, so ask [`Context.isDone`](isDone.md) when you need to
  know whether the wait finished. Without a context, the call waits until it can finish or the run
  ends.
- A channel send or receive given a context that is done already gives up at once, even when it
  could proceed.
- Locks and conditions, [`Mutex.lock`](../Mutex/lock.md) and [`Cond.wait`](../Cond/wait.md) among
  them, take no context.
- Everything a goroutine writes before `Context.cancel` is visible after a wait that the cancel
  ends.

## Basic examples

`Context.new` makes a running context. `Context.cancel` makes it done, and `Context.reason` says
why:

```ant
let ctx = Context.new()
Context.isDone(ctx) // false
Context.cancel(ctx)
Context.isDone(ctx) // true
Context.reason(ctx) // "cancelled"
```

A context with a timeout bounds a wait. The receive gives up with `null` when no value comes in
time:

```ant
let replies: Channel[string] = Channel.new()
let ctx = Context.new(50ms)
println(replies.receive(ctx: ctx) ?? "no reply: ${Context.reason(ctx)}")
```

```text
no reply: timeout
```

## Scope

How a call gives up follows its return type. A `Result` call fails with the reason, a
value-or-`null` call returns `null`, and a `void` call returns early:

```ant
let ctx = Context.new()
ctx.cancel()
let jobs: Channel[string] = Channel.new()
println(Cmd.read(Cmd.new("sleep", ["5"]), ctx: ctx))
println(jobs.receive(ctx: ctx))
Time.sleep(1h, ctx: ctx)
println("slept: ${ctx.isDone()}")
```

```text
Err{ error: Error{ message: "cancelled" } }
null
slept: true
```

A child ends with its parent and takes the parent's reason. Its deadline is the earlier of its own
and its parent's:

```ant
let request = Context.new(1s)
let step = Context.new(1h, parent: request)
println(step.deadline() == request.deadline())
request.cancel()
println(step.reason())
```

```text
true
cancelled
```

[`Os.interrupt`](../Os/interrupt.md) returns the context that Ctrl-C ends, with the reason
`interrupt`:

<!-- test: before
go {
  Time.sleep(100ms)
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
let stop = Os.interrupt()
stop.wait()
println(stop.reason())
```

```text
interrupt
```

Contexts compare by identity, and print as their type's name:

```ant
let a = Context.new()
let alias = a
println(a == Context.new())
println(a == alias)
println(a)
```

```text
false
true
Context
```

## Functions by job

### Making a context

- [Context.new](new.md): a child of a parent or of the run, with an optional timeout
- [Os.interrupt](../Os/interrupt.md): the context that Ctrl-C, SIGINT, or SIGTERM ends
- [Test.context](../Test/context.md): the context that ends with the current test

### Ending a context

- [Context.cancel](cancel.md): make the context done now, with the reason `"cancelled"`

### Asking about a context

- [Context.isDone](isDone.md): whether the context is done
- [Context.reason](reason.md): why it is done, or `null` while it runs
- [Context.deadline](deadline.md): the earliest deadline of the context and its parents

### Waiting

- [Context.wait](wait.md): wait until the context is done; also a head of `select`

## Applications

### One budget for several steps

Pass one context to every step, and each step waits only as long as the budget has left. A `void`
call such as `Time.sleep` returns early, so the loop checks the context between steps:

```ant
let ctx = Context.new(100ms)
for step in ["fetch", "parse", "store"] {
  Time.sleep(40ms, ctx: ctx)
  if ctx.isDone() {
    println("${step}: ${ctx.reason()}")
    break
  }
  println("${step}: done")
}
```

```text
fetch: done
parse: done
store: timeout
```

### Write a function that waits

A function that blocks takes `ctx: Context | null = null` as its last parameter and passes it on to
the calls it waits in. Here the caller's context ends the wait, and the reason becomes the error:

```ant
func nextJob(jobs: Channel[string], ctx: Context | null = null): Result[string, Error] {
  let job = jobs.receive(ctx: ctx) ?? return Result.fail(ctx?.reason() ?? "no more jobs")
  Result.ok(job)
}

let jobs: Channel[string] = Channel.new(1)
jobs.send("resize")
println(nextJob(jobs, ctx: Context.new(50ms)))
println(nextJob(jobs, ctx: Context.new(50ms)))
jobs.close()
println(nextJob(jobs))
```

```text
Ok{ value: "resize" }
Err{ error: Error{ message: "timeout" } }
Err{ error: Error{ message: "no more jobs" } }
```

### Stop the rest when one finishes

[`Parallel.race`](../Parallel/race.md) runs callbacks at once, takes the first value returned, and
cancels its scope context, so the others stop waiting:

```ant
func search(name: string, delay: Time.Duration, ctx: Context): string | null {
  Time.sleep(delay, ctx: ctx)
  if ctx.isDone() { null } else { name }
}

let winner = Parallel.race([{ search("fast", 10ms, x) }, { search("slow", 1h, x) }])
println(winner)
```

```text
fast
```

### Wait for a reply or the deadline

A `select` waits on a channel and a context at once. A `select` head takes no `ctx:`; the
`Context.wait` clause takes its place:

```ant
let replies: Channel[string] = Channel.new()
let ctx = Context.new(50ms)
select {
  let reply = Channel.receive(replies) { println(reply ?? "closed") }
  Context.wait(ctx) { println("gave up: ${ctx.reason()}") }
}
```

```text
gave up: timeout
```

### Finish the work at Ctrl-C

After `Os.interrupt`, the first Ctrl-C ends the waits on its context instead of the program, so the
program can report what it has done before it exits:

<!-- test: before
go {
  Time.sleep(100ms)
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
let stop = Os.interrupt()
var processed = 0
while !stop.isDone() {
  Time.sleep(30ms, ctx: stop)
  if !stop.isDone() { processed += 1 }
}
println("stopped by ${stop.reason()}")
println("work done before the stop: ${processed > 0}")
```

```text
stopped by interrupt
work done before the stop: true
```

## Properties and relations

A context is done exactly when it has a reason:

```ant
let ctx = Context.new()
assert(ctx.isDone() == (ctx.reason() != null))
ctx.cancel()
assert(ctx.isDone() == (ctx.reason() != null))
```

A context is done once. Cancelling it after its timeout keeps the reason `timeout`:

```ant
let ctx = Context.new(10ms)
ctx.wait()
ctx.cancel()
assert(ctx.reason() == "timeout")
```

A child of a done parent is done at once with the parent's reason:

```ant
let parent = Context.new()
parent.cancel()
let child = Context.new(parent: parent)
assert(child.reason() == parent.reason())
```

## Possible issues

### A void call does not say that it gave up

`Time.sleep` and `WaitGroup.wait` return early when the context ends, and the code after them runs
as if the wait had finished. Ask `isDone` before you act on the wait:

```ant
let ctx = Context.new(20ms)
let started = Time.monotonic()
Time.sleep(1h, ctx: ctx)
println("slept an hour: ${!ctx.isDone()}")
println("short: ${Time.monotonic() - started < 1000}")
```

```text
slept an hour: false
short: true
```

### A done context stops a call that could proceed

A send on a buffered channel with room still fails when its context is done already, and buffers
nothing:

```ant
let ctx = Context.new()
ctx.cancel()
let jobs: Channel[string] = Channel.new(4)
println(jobs.send("resize", ctx: ctx))
println(jobs.length())
```

```text
false
0
```

### A reason names no call

A failed call's message is only the reason, so it does not say which call gave up. Add the operation
with `Result.wrapErr`:

```ant
let ctx = Context.new()
ctx.cancel()
println(Cmd.read(Cmd.new("sleep", ["5"]), ctx: ctx).wrapErr("run sleep"))
```

```text
Err{ error: Error{ message: "run sleep: cancelled" } }
```

### Locks take no context

`Mutex.lock` waits until the mutex is free and cannot give up:

```ant
let m = Mutex.new()
let ctx = Context.new(1s)
m.lock(ctx: ctx) // error: Mutex.lock has no parameter ctx
```

Keep the lock short, or hand work over through a channel, whose waits take a context.

## See also

- [Os.interrupt](../Os/interrupt.md): the context that Ctrl-C ends
- [Channel](../Channel/index.md): send and receive with a context
- [Parallel](../Parallel/index.md): run callbacks on goroutines that share a scope context
- [WaitGroup](../WaitGroup/index.md): wait for goroutines, with a context
- [Time.sleep](../Time/sleep.md): sleep until a duration passes or a context ends
- [select](../../language/keywords/select.md): wait for a context and channels at once
- [Concurrency](../../guide/concurrency.md): how goroutines, channels, and contexts fit together
