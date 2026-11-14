<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Context.new

```ant signature
func Context.new(timeout: Time.Duration | null = null, parent: Context | null = null): Context
```

Context.new returns a new context that is a child of parent, or of the run when parent is null. With
a timeout it is done once the timeout passes; a child of a done parent is done at once with the
parent's reason. It panics if timeout has years, months, or weeks, or is not positive.

<!-- end generated -->

## Details

- A context without a timeout is done only when a call of [Context.cancel](cancel.md) cancels it,
  when its parent is done, or when the run is cancelled.
- A timeout of days counts each day as 24 hours.
- A context is done once and stays done. A timeout that has passed ends every later wait on that
  context at once.
- Cancelling a child leaves its parent running.

## Basic examples

A new context is running, with no reason yet:

```ant
let ctx = Context.new()
ctx.isDone() // false
ctx.reason() // null
```

A context with a timeout becomes done when the timeout passes, with the reason `timeout`:

```ant
let ctx = Context.new(50ms)
ctx.wait()
println(ctx.reason())
```

```text
timeout
```

Pass the context to a blocking call to bound its wait:

```ant
let replies: Channel[string] = Channel.new()
let ctx = Context.new(50ms)
println(replies.receive(ctx: ctx) ?? "no reply")
```

```text
no reply
```

## Scope

Any duration of hours and smaller units works as a timeout, and so do days:

```ant
Context.new(1s500ms).isDone()                // false
Context.new(Time.duration(days: 1)).isDone() // false
```

A context with a timeout has a deadline, the instant the timeout passes. Without a timeout, it has a
deadline only when a parent has one:

```ant
println(Context.new(1h).deadline() != null)
println(Context.new().deadline())
```

```text
true
null
```

Each call makes a new context. Contexts compare by identity:

```ant
let ctx = Context.new()
ctx == ctx                     // true
Context.new() == Context.new() // false
```

## Named arguments

### timeout

`timeout:` is the first parameter, so the name is optional. With the name, the call reads like an
[Http.get](../Http/get.md) call that takes one:

```ant
let ctx = Context.new(timeout: 20ms)
ctx.wait()
println(ctx.reason())
```

```text
timeout
```

### parent

With `parent:`, the new context is done when its parent is done, with the parent's reason:

```ant
let parent = Context.new()
let child = Context.new(parent: parent)
parent.cancel()
println(child.reason())
```

```text
cancelled
```

A child of a parent with a timeout ends when the parent times out:

```ant
let parent = Context.new(20ms)
let child = Context.new(parent: parent)
child.wait()
println(child.reason())
```

```text
timeout
```

A child of a done parent is done at once, even with a timeout of its own:

```ant
let parent = Context.new()
parent.cancel()
let child = Context.new(1h, parent: parent)
child.isDone() // true
child.reason() // "cancelled"
```

Combine `timeout:` and `parent:` to give one step a shorter budget inside a larger one:

```ant
let request = Context.new(5s)
let lookup = Context.new(100ms, parent: request)
lookup.wait()
println("lookup: ${lookup.reason()}")
println("request done: ${request.isDone()}")
```

```text
lookup: timeout
request done: false
```

## Applications

### Stop a group of goroutines

Give each goroutine a child of one context, and cancel that context to stop them all:

```ant
let run = Context.new()
let results: Channel[string] = Channel.new()
for name in ["fetch", "index"] {
  let work = Context.new(parent: run)
  go {
    work.wait()
    results.send("${name}: ${work.reason()}")
  }
}
Time.sleep(20ms)
run.cancel()
let lines = [results.receive() ?? "", results.receive() ?? ""]
for line in lines.sort() {
  println(line)
}
```

```text
fetch: cancelled
index: cancelled
```

### Retry with a timeout per attempt

Each attempt gets its own short timeout, and all of them stay under an overall budget:

```ant
func attempt(n: int, ctx: Context): Result[string, Error] {
  let replies: Channel[string] = Channel.new(1)
  if n == 3 { replies.send("pong") }
  let reply = replies.receive(ctx: ctx) ?? return Result.fail("attempt ${n}: ${ctx.reason()}")
  Result.ok(reply)
}

let overall = Context.new(1s)
for n in Iter.range(1, 5) {
  let result = attempt(n, Context.new(20ms, parent: overall))
  println(result)
  if result.isOk() { break }
}
```

```text
Err{ error: Error{ message: "attempt 1: timeout" } }
Err{ error: Error{ message: "attempt 2: timeout" } }
Ok{ value: "pong" }
```

### Stop at Ctrl-C or after a budget

A child of [Os.interrupt](../Os/interrupt.md) ends on the first interrupt or when its own timeout
passes, whichever comes first. Its reason tells which:

```ant
let ctx = Context.new(50ms, parent: Os.interrupt())
while !ctx.isDone() {
  // one step of work
  Time.sleep(10ms, ctx: ctx)
}
println("stopped: ${ctx.reason()}")
```

```text
stopped: timeout
```

## Properties and relations

A child's deadline is the earlier of its own and its parent's:

```ant
let outer = Context.new(100ms)
let inner = Context.new(1h, parent: outer)
assert(inner.deadline() == outer.deadline())
```

Cancelling a child leaves its parent running:

```ant
let parent = Context.new()
let child = Context.new(parent: parent)
child.cancel()
assert(child.isDone())
assert(!parent.isDone())
```

## Possible issues

### The parent is not the first argument

The first parameter is the timeout, so a parent passed by position is a compile error:

```ant
let parent = Context.new()
let child = Context.new(parent) // error: cannot use Context where Time.Duration | null is expected
```

Name it: `Context.new(parent: parent)`.

### A timeout that is not positive panics

A zero or negative timeout panics, rather than making a context that is done at once:

```ant
Context.new(0s) // panic: Context.new: the timeout is zero or negative
```

To make a context that is done now, cancel it:

```ant
let ctx = Context.new()
ctx.cancel()
ctx.isDone() // true
```

### Years, months and weeks panic

They have no fixed length, so the type checker accepts them but the call panics:

```ant
Context.new(Time.duration(weeks: 1))
// panic: Context.new: the timeout has years, months or weeks, which have no fixed length
```

Use days or hours, which `Context.new` takes as fixed spans: `Time.duration(days: 7)`.

### A shared timeout runs out for every later wait

The timeout belongs to the context, not to each wait. Once it passes, every wait on the context
gives up at once:

```ant
let ctx = Context.new(30ms)
let replies: Channel[string] = Channel.new()
for n in Iter.range(1, 3) {
  let started = Time.monotonic()
  replies.receive(ctx: ctx)
  let waited = Time.monotonic() - started >= 25
  println("wait ${n}: waited ${waited}")
}
```

```text
wait 1: waited true
wait 2: waited false
```

Make a new context for each wait that needs a budget of its own, as in the retry example above.

## See also

[Context.cancel](cancel.md) · [Context.wait](wait.md) · [Context.isDone](isDone.md) ·
[Context.reason](reason.md) · [Context.deadline](deadline.md) · [Os.interrupt](../Os/interrupt.md) ·
[Time.sleep](../Time/sleep.md) · [Parallel.map](../Parallel/map.md)
