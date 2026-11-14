<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Context.isDone

```ant signature
func Context.isDone(ctx: Context): bool
```

Context.isDone reports whether ctx is done.

<!-- end generated -->

## Details

- A context becomes done once: when `Context.cancel` cancels it, when its deadline passes, or when
  its parent is done. After that, `Context.isDone` stays true for good.
- `Context.isDone` does not wait. It answers for the moment of the call; to block until the context
  is done, call `Context.wait`.
- A wait that takes a context and returns `void`, such as `Time.sleep` or `WaitGroup.wait`, returns
  early when the context ends without saying so. Ask `Context.isDone` afterwards to tell the two
  apart.

## Basic examples

A new context is not done; `Context.cancel` makes it done:

```ant
let ctx = Context.new()
Context.isDone(ctx) // false
Context.cancel(ctx)
Context.isDone(ctx) // true
```

A context with a timeout becomes done when its deadline passes:

```ant
let ctx = Context.new(20ms)
println(Context.isDone(ctx))
Time.sleep(50ms)
println(Context.isDone(ctx))
```

```text
false
true
```

## Scope

A deadline far in the future leaves the context running, and so does a context with no timeout at
all:

```ant
Context.new(1h).isDone() // false
Context.new().isDone()   // false
```

A child is done as soon as its parent is, with the parent's reason. A child made from a parent that
is done already starts out done:

```ant
let parent = Context.new()
let child = Context.new(parent: parent)
parent.cancel()
child.isDone()                           // true
child.reason()                           // "cancelled"
Context.new(1h, parent: parent).isDone() // true
```

Cancelling a child leaves its parent running:

```ant
let parent = Context.new()
let child = Context.new(parent: parent)
child.cancel()
child.isDone()  // true
parent.isDone() // false
```

## Applications

### Stop a worker between steps

A goroutine that should stop when asked checks `Context.isDone` before each step, and passes the
context to its waits so that a sleep in progress ends too:

```ant
let ctx = Context.new()
let stopped: Channel[string] = Channel.new()
go {
  while !ctx.isDone() {
    Time.sleep(5ms, ctx: ctx)
  }
  stopped.send("worker stopped: ${ctx.reason()}")
}
Time.sleep(20ms)
ctx.cancel()
println(stopped.receive())
```

```text
worker stopped: cancelled
```

### Tell a timeout from the end of a channel

`Channel.receive` returns `null` both when the channel is closed and drained and when its context
ends first. `Context.isDone` tells which one happened:

```ant
func nextJob(jobs: Channel[string], ctx: Context): string {
  let job = jobs.receive(ctx: ctx)
  if job != null { return "job ${job}" }
  if ctx.isDone() { "timed out" } else { "no more jobs" }
}

let jobs: Channel[string] = Channel.new(1)
println(nextJob(jobs, Context.new(20ms)))
jobs.close()
println(nextJob(jobs, Context.new(20ms)))
```

```text
timed out
no more jobs
```

### Retry until a deadline

`Time.sleep` returns early when its context ends, and returns nothing either way. Check
`Context.isDone` after the pause to give up instead of trying again:

```ant
func connect(attempts: Atomic[int]): Result[string, Error] {
  if attempts.add(1) < 3 { return Result.fail("connection refused") }
  Result.ok("connected")
}

func retry(attempts: Atomic[int], ctx: Context): Result[string, Error] {
  while true {
    let r = connect(attempts)
    if r is Ok { return r }
    Time.sleep(10ms, ctx: ctx)
    if ctx.isDone() { return Result.fail("gave up: ${ctx.reason()}") }
  }
}

println(retry(Atomic.new(0), Context.new(1s)))
println(retry(Atomic.new(-100), Context.new(25ms)))
```

```text
Ok{ value: "connected" }
Err{ error: Error{ message: "gave up: timeout" } }
```

### Stop a losing Parallel.race callback

`Parallel.race` cancels its scope context `x` once there is a winner. Cancellation is cooperative,
so a compute-bound callback checks `x.isDone()` between steps and stops when it lost. A callback
that never checks it would hold the call until it ends:

```ant
var slowSteps = 0
let winner = Parallel.race([
  {
    var steps = 0
    while !x.isDone() {
      steps += 1
      if steps == 1000 { return "fast" }
    }
    "fast stopped"
  },
  {
    while !x.isDone() {
      slowSteps += 1
      if slowSteps == 100000000 { return "slow" }
    }
    "slow stopped"
  },
])
println(winner)
println(slowSteps < 100000000)
```

```text
fast
true
```

## Properties and relations

A context is done exactly when it has a reason:

```ant
let ctx = Context.new()
assert(ctx.isDone() == (ctx.reason() != null))
ctx.cancel()
assert(ctx.isDone() == (ctx.reason() != null))
```

Once done, a context stays done; a second `Context.cancel` changes nothing:

```ant
let ctx = Context.new()
ctx.cancel()
ctx.cancel()
assert(ctx.isDone())
assert(ctx.reason() == "cancelled")
```

After `Context.wait` returns, the context is done:

```ant
let ctx = Context.new(10ms)
ctx.wait()
assert(ctx.isDone())
```

## Possible issues

### Polling instead of waiting

`Context.isDone` returns at once, so a loop that only polls it keeps the goroutine busy until the
context ends. To wait for the context, call `Context.wait`, or pass the context to the wait in the
loop, as in [Stop a worker between steps](#stop-a-worker-between-steps). This loop spins the whole
50 milliseconds:

```ant
let ctx = Context.new(50ms)
var polls = 0
while !ctx.isDone() {
  polls += 1
}
println(polls > 1)
```

```text
true
```

### A done context cannot be reused

A context never becomes not done again. A loop that shares one context across rounds gives up on
every round after the first timeout; make a fresh context for each round instead:

```ant
let jobs: Channel[int] = Channel.new(4)
let shared = Context.new(20ms)
for _ in [1, 2] {
  let got = jobs.receive(ctx: shared)
  println("shared: ${got}, done: ${shared.isDone()}")
}
jobs.send(7)
println("shared again: ${jobs.receive(ctx: shared)}")
println("fresh: ${jobs.receive(ctx: Context.new(20ms))}")
```

```text
shared: null, done: true
shared: null, done: true
shared again: null
fresh: 7
```

## See also

[Context.new](new.md) · [Context.cancel](cancel.md) · [Context.reason](reason.md) ·
[Context.wait](wait.md) · [Context.deadline](deadline.md) · [Context](index.md) ·
[Channel.receive](../Channel/receive.md) · [WaitGroup.wait](../WaitGroup/wait.md) ·
[Time.sleep](../Time/sleep.md) · [Concurrency guide](../../guide/concurrency.md)
