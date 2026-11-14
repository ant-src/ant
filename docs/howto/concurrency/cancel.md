<!-- status: reviewed -->

# Cancel a group of goroutines

This script starts three workers, lets them run, then stops them all with one context and waits
until each has finished:

```ant
let names = ["indexer", "uploader", "cleaner"]
let reports = Array.filled(names.length(), "")
let ctx = Context.new()
let wg = WaitGroup.new()
for entry in names.entries() {
  wg.add(1)
  go {
    defer { wg.done() }
    while !ctx.isDone() {
      // one step of work
      Time.sleep(10ms, ctx: ctx)
    }
    reports[entry.index] = "${entry.value} stopped: ${ctx.reason() ?? "running"}"
  }
}
Time.sleep(50ms)
ctx.cancel()
wg.wait()
println(reports | String.join("\n"))
```

```text
indexer stopped: cancelled
uploader stopped: cancelled
cleaner stopped: cancelled
```

1. Create one context for the group with [Context.new](../../lib/Context/new.md) and a counter with
   [WaitGroup.new](../../lib/WaitGroup/new.md).
2. Call [WaitGroup.add](../../lib/WaitGroup/add.md) before each `go`, not inside the goroutine, so
   the wait cannot run before the counter goes up.
3. Make `defer { wg.done() }` the first line of the goroutine.
   [WaitGroup.done](../../lib/WaitGroup/done.md) then runs even when the goroutine panics, and the
   wait does not hang.
4. Check [Context.isDone](../../lib/Context/isDone.md) between steps, and pass the context as `ctx:`
   to every wait, such as [Time.sleep](../../lib/Time/sleep.md), so a wait in progress ends at the
   cancel. A goroutine that ignores the context runs until the script ends.
5. Write each goroutine's result into its own slot of an array made with
   [Array.filled](../../lib/Array/filled.md), indexed through
   [Array.entries](../../lib/Array/entries.md). A local the goroutine assigns never narrows, so a
   nullable result written there cannot be narrowed after the wait.
   [String.join](../../lib/String/join.md) then prints the slots in order.
6. Call [Context.cancel](../../lib/Context/cancel.md), then
   [WaitGroup.wait](../../lib/WaitGroup/wait.md). [Context.reason](../../lib/Context/reason.md)
   tells each goroutine why it stopped.

To cancel part of the group alone, give that part a child context, `Context.new(parent: ctx)`.
Cancelling `ctx` still stops the child too.

## Let Parallel wait for the group

[Parallel.run](../../lib/Parallel/run.md) starts every function at once and returns when all have
ended, so it needs no `WaitGroup`. Pass the group's context as `ctx:`, and pass each function's own
context, `x`, to its waits:

```ant
func poll(name: string, ctx: Context): string {
  while !ctx.isDone() {
    Time.sleep(10ms, ctx: ctx)
  }
  "${name} stopped: ${ctx.reason() ?? "running"}"
}

let stop = Context.new()
let reports = ["", ""]
go {
  Time.sleep(50ms)
  stop.cancel()
}
Parallel.run([
  { reports[0] = poll("inbox", x) },
  { reports[1] = poll("outbox", x) },
], ctx: stop)
println(reports | String.join("\n"))
```

```text
inbox stopped: cancelled
outbox stopped: cancelled
```

The functions share the call's own context, a child of `stop`, so cancelling `stop` ends them all.
Parallel.run never cancels `stop` itself.

## Stop the group at the first error

Parallel.each and Parallel.run do not look inside the `Result` a callback returns;
[Parallel.any](../../lib/Parallel/any.md) stops at the first `Ok` (see
[Take the first result](first-result.md)). To stop the rest when one fails, cancel a context you
own, keep the first error in an atomic made with [Atomic.new](../../lib/Atomic/new.md), and record
it with [Atomic.compareAndSwap](../../lib/Atomic/compareAndSwap.md):

```ant
func download(name: string, ctx: Context): Result[string, Error] {
  if name == "b.zip" {
    return Result.fail("${name}: connection refused")
  }
  Time.sleep(10s, ctx: ctx)
  if ctx.isDone() {
    return Result.fail("${name}: ${ctx.reason() ?? "running"}")
  }
  Result.ok(name)
}

let names = ["a.zip", "b.zip", "c.zip"]
let failed = Context.new()
let firstErr: Atomic[Error | null] = Atomic.new(null)
Parallel.each(names, func(name: string, work: Context) {
  let r = download(name, work)
  if r is Err {
    firstErr.compareAndSwap(null, r.error)
    failed.cancel()
  }
}, limit: names.length(), ctx: failed)
let err = firstErr.load()
println(if err == null { "all downloaded" } else { "download failed: ${err.message}" })
```

```text
download failed: b.zip: connection refused
```

1. Run the work with [Parallel.each](../../lib/Parallel/each.md) under a context made for this
   group.
2. On the first `Err`, store it, then call [Context.cancel](../../lib/Context/cancel.md). The other
   downloads wake from their waits and fail with `cancelled`, but `compareAndSwap` keeps only the
   first error. An element not yet started is never started.
3. After the call returns, read the error with [Atomic.load](../../lib/Atomic/load.md).

## Stop the group when one panics

A panic in a goroutine started with `go` ends that goroutine alone; the others go on. A panic in a
Parallel callback cancels the other callbacks, waits for them, and panics again in the caller, where
[Panic.catch](../../lib/Panic/catch.md) can stop it:

```ant
func poll(name: string, ctx: Context): string {
  while !ctx.isDone() {
    Time.sleep(10ms, ctx: ctx)
  }
  "${name} stopped: ${ctx.reason() ?? "running"}"
}

let reports = [""]
let caught = Panic.catch({
  Parallel.run([
    { reports[0] = poll("inbox", x) },
    {
      Time.sleep(20ms, ctx: x)
      panic("outbox: bad message")
    },
  ])
})
if caught is Err {
  println("group panicked: ${caught.error.message}")
}
println(reports[0])
```

```text
group panicked: outbox: bad message
inbox stopped: cancelled
```

## See also

[Context.new](../../lib/Context/new.md) · [Context.cancel](../../lib/Context/cancel.md) ·
[Context.isDone](../../lib/Context/isDone.md) · [Context.reason](../../lib/Context/reason.md) ·
[WaitGroup.add](../../lib/WaitGroup/add.md) · [WaitGroup.done](../../lib/WaitGroup/done.md) ·
[WaitGroup.wait](../../lib/WaitGroup/wait.md) · [Parallel.run](../../lib/Parallel/run.md) ·
[Parallel.each](../../lib/Parallel/each.md) · [Panic.catch](../../lib/Panic/catch.md) ·
[Atomic.compareAndSwap](../../lib/Atomic/compareAndSwap.md) · [Give up after a timeout](timeout.md)
· [Take the first result](first-result.md) · [Run a worker pool](worker-pool.md) ·
[Concurrency](../../guide/concurrency.md)
