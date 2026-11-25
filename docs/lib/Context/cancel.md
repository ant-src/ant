<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Context.cancel

```ant signature
func Context.cancel(ctx: Context): void
```

Context.cancel makes ctx done now, with the reason "cancelled". It does nothing if ctx is already
done.

<!-- end generated -->

## Details

- Every wait given `ctx`, or a context that descends from it, gives up: a `Result` call returns
  `Err` with the message `"cancelled"`, a value-or-null call returns `null`, `Channel.send` returns
  `false`, and a `void` call returns early.
- The children of `ctx` become done with it, with the same reason. Its parent is not affected.
- A done context stays done. Cancelling it again, or after its timeout, changes nothing, not even
  its reason.
- Cancellation is cooperative: it ends only the waits given the context or a descendant of it, and
  code that checks `Context.isDone`.
- Everything a goroutine did before `Context.cancel` is visible to a goroutine whose wait the cancel
  ends.

## Basic examples

Cancel a context, and it is done with the reason `"cancelled"`:

```ant
let ctx = Context.new()
Context.isDone(ctx) // false
Context.cancel(ctx)
Context.isDone(ctx) // true
Context.reason(ctx) // "cancelled"
```

Cancel a context to end a wait in another goroutine. The receive gives up and returns `null`:

```ant
let jobs: Channel[string] = Channel.new()
let ctx = Context.new()
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  let job = jobs.receive(ctx: ctx)
  println("received ${job}")
}
Context.cancel(ctx)
wg.wait()
```

```text
received null
```

## Scope

Cancelling a child leaves its parent running. Cancelling the parent makes every child done, the ones
made after it included:

```ant
let parent = Context.new()
let child = Context.new(parent: parent)
child.cancel()
println("parent ${parent.isDone()}, child ${child.isDone()}")
parent.cancel()
let late = Context.new(parent: parent)
println("late ${late.isDone()}, ${late.reason()}")
```

```text
parent false, child true
late true, cancelled
```

A context that is done already keeps its reason. Here the timeout comes first:

```ant
let ctx = Context.new(timeout: 10ms)
ctx.wait()
ctx.cancel()
println(ctx.reason())
```

```text
timeout
```

One cancel wakes every goroutine that waits on the context, so a context works as a stop signal for
any number of them:

```ant
let stop = Context.new()
let wg = WaitGroup.new()
let woken = Atomic.new(0)
for _ in Array.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    stop.wait()
    woken.add(1)
  }
}
stop.cancel()
wg.wait()
println("${woken.load()} goroutines woken")
```

```text
3 goroutines woken
```

In a `select`, a `Context.wait` clause can proceed once the context is cancelled:

```ant
let stop = Context.new()
let jobs: Channel[string] = Channel.new()
let finished: Channel[string] = Channel.new(1)
go {
  select {
    let job = Channel.receive(jobs) { finished.send("job ${job}") }
    Context.wait(stop) { finished.send("stopped") }
  }
}
stop.cancel()
println(finished.receive())
```

```text
stopped
```

A call that returns a `Result` gives up with an `Err` whose message is the reason:

```ant
let listener = Result.unwrap(Net.listen("127.0.0.1:0"))
defer { listener.close() }
let ctx = Context.new()
go {
  Time.sleep(10ms)
  ctx.cancel()
}
println(Net.accept(listener, ctx: ctx))
```

```text
Err{ error: Error{ message: "cancelled", kind: null } }
```

## Applications

### Stop a producer when you have enough

A function that starts a goroutine owns its context. `defer` right after creating it cancels the
context on every exit, so the producer stops when the function returns:

```ant
func generate(out: Channel[int], ctx: Context) {
  var n = 0
  while !ctx.isDone() {
    n += 1
    out.send(n * n, ctx: ctx)
  }
}

func firstSquares(count: int): Array[int] {
  let ctx = Context.new()
  defer { ctx.cancel() }
  let squares: Channel[int] = Channel.new()
  go { generate(squares, ctx) }
  let result: Array[int] = []
  for _ in Array.range(0, count) {
    result.push(squares.receive() ?? 0)
  }
  result
}

println(firstSquares(4))
```

```text
[1, 4, 9, 16]
```

### Fail fast in Parallel work

`Parallel` never cancels the context it is given, so a callback that wants the whole call to stop
cancels a context the program owns. Here the bad element stops the admission of the rest, and
`Parallel.map` returns `null`:

```ant
func double(n: int, all: Context): int {
  if n < 0 { all.cancel() }
  n * 2
}

let all = Context.new()
let doubled = Parallel.map(
  [1, 2, -3, 4, 5],
  func(n: int, work: Context) { double(n, all) },
  limit: 1,
  ctx: all,
)
println(doubled)
println(all.reason())
```

```text
null
cancelled
```

### Signal that a result is ready

A cancel orders what came before it, so a goroutine can store a value in a shared array, cancel a
context, and the waiter reads it with no lock:

```ant
let messages: Array[string] = []
let ready = Context.new()
go {
  messages.push("hello")
  ready.cancel()
}
ready.wait()
println(messages[0])
```

```text
hello
```

## Properties and relations

A cancel makes the context and its children done, leaves the parent running, and a second cancel
changes nothing:

```ant
let parent = Context.new()
let ctx = Context.new(parent: parent)
let child = Context.new(parent: ctx)
ctx.cancel()
assert(ctx.isDone() && ctx.reason() == "cancelled")
assert(child.isDone() && child.reason() == "cancelled")
assert(!parent.isDone() && parent.reason() == null)
ctx.cancel()
assert(ctx.reason() == "cancelled")
```

## Possible issues

### A cancelled context cannot be reused

A done context stays done, and a channel send or receive given one gives up at once, even when the
channel is ready. Make a new context for each attempt:

```ant
let ctx = Context.new()
ctx.cancel()
let replies: Channel[string] = Channel.new(1)
replies.send("pong")
println(replies.receive(ctx: ctx))
println(replies.receive(ctx: Context.new()))
```

```text
null
pong
```

### A void wait returns as if it finished

`Time.sleep` and the other `void` waits return early on a cancel, and the code after them cannot
tell from the return. Ask `Context.isDone`:

```ant
let ctx = Context.new()
go { ctx.cancel() }
Time.sleep(1m, ctx: ctx)
if ctx.isDone() {
  println("woken early: ${ctx.reason()}")
}
```

```text
woken early: cancelled
```

### An optional context cannot be cancelled

A `ctx: Context | null` parameter may be `null`, so `Context.cancel` rejects it. A function that
needs to cancel its work makes a child of its caller's context, which `Context.new` accepts as
`null`, and cancels that:

```ant
func stopAll(ctx: Context | null = null) {
  Context.cancel(ctx) // error: cannot use Context | null where Context is expected
}
```

```ant
func withWork(ctx: Context | null = null): string {
  let work = Context.new(parent: ctx)
  defer { work.cancel() }
  "done"
}

println(withWork())
```

```text
done
```

## See also

[Context.new](new.md) · [Context.isDone](isDone.md) · [Context.reason](reason.md) ·
[Context.wait](wait.md) · [Channel.receive](../Channel/receive.md) ·
[Parallel.map](../Parallel/map.md) · [select](../../language/keywords/select.md) ·
[Concurrency guide](../../guide/concurrency.md)
