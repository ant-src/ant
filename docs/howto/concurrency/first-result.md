<!-- status: reviewed -->

# Take the first result of several

Ask several sources the same question, use the first answer, and stop the others.

1. Make one channel with [Channel.new](../../lib/Channel/new.md) for the answers.
2. Make a context with [Context.new](../../lib/Context/new.md). Give it a timeout to bound the whole
   wait, and [`defer`](../../language/keywords/defer.md) its
   [Context.cancel](../../lib/Context/cancel.md) right after.
3. Start one goroutine per source with [`go`](../../language/keywords/go.md). Pass the context to
   every wait inside it, and to its [Channel.send](../../lib/Channel/send.md).
4. Receive once with [Channel.receive](../../lib/Channel/receive.md), under the same context. It
   returns `null` when the context ends first.
5. The deferred cancel stops the sources still working when the function or script ends. When more
   work follows in the same body, also call `ctx.cancel()` right after the receive; keep the
   deferred cancel for early exits.

## Ask several mirrors

This script asks three mirrors for the same file and prints the first answer:

```ant
type Mirror = { name: string, delay: Time.Duration }

func fetch(m: Mirror, ctx: Context): string | null {
  Time.sleep(m.delay, ctx: ctx)
  if ctx.isDone() { return null }
  "release notes from ${m.name}"
}

let mirrors = [
  Mirror{ name: "eu", delay: 300ms },
  Mirror{ name: "us", delay: 20ms },
  Mirror{ name: "asia", delay: 1s },
]
let replies: Channel[string] = Channel.new()
let ctx = Context.new(2s)
defer { ctx.cancel() }
for m in mirrors {
  go {
    let reply = fetch(m, ctx)
    if reply != null { replies.send(reply, ctx: ctx) }
  }
}
let first = replies.receive(ctx: ctx) ?? "no mirror answered"
println(first)
```

```text
release notes from us
```

`fetch` stands in for a real request: [Time.sleep](../../lib/Time/sleep.md) plays the network, and
it returns early once the context is done. A source that finishes after the winner has no receiver,
so its send waits until the cancel and then returns `false`. Without `ctx:` on the send, that
goroutine waits until the program ends.

## Report when nobody answers in time

When the timeout passes before any answer, the receive returns `null`.
[Context.reason](../../lib/Context/reason.md) says why the wait ended:

```ant
func ask(name: string, delay: Time.Duration, ctx: Context): string | null {
  Time.sleep(delay, ctx: ctx)
  if ctx.isDone() { return null }
  name
}

let replies: Channel[string] = Channel.new()
let ctx = Context.new(100ms)
defer { ctx.cancel() }
for name in ["eu", "us"] {
  go {
    let reply = ask(name, 1s, ctx)
    if reply != null { replies.send(reply, ctx: ctx) }
  }
}
let first = replies.receive(ctx: ctx)
if first == null {
  println("no mirror answered: ${ctx.reason()}")
}
```

```text
no mirror answered: timeout
```

## Wait on sources of different types

When the sources give different types, give each its own channel and wait on all of them in one
[`select`](../../language/keywords/select.md). It runs the clause of the first one ready.

When the sources are your own goroutines, you can instead make one channel of a union, such as
`Channel[string | Price]`, keep the single receive, and tell the answers apart with
[`is`](../../language/operators/is.md).

With a `select`, add a [Context.wait](../../lib/Context/wait.md) clause for the timeout:

```ant
type Price = { cents: int }

let cached: Channel[string] = Channel.new()
let live: Channel[Price] = Channel.new()
let ctx = Context.new(1s)
defer { ctx.cancel() }
go {
  Time.sleep(500ms, ctx: ctx)
  cached.send("cached: 1990 cents", ctx: ctx)
}
go {
  Time.sleep(10ms, ctx: ctx)
  live.send(Price{ cents: 2015 }, ctx: ctx)
}
select {
  let text = Channel.receive(cached) { println(text ?? "no cache") }
  let price = Channel.receive(live) {
    if price != null { println("live: ${price.cents} cents") }
  }
  Context.wait(ctx) { println("no price: ${ctx.reason()}") }
}
```

```text
live: 2015 cents
```

Write each head in namespace form, `Channel.receive(live)`: `live.receive()` is not a `select` head.
A head binds `null` when its channel is closed, so check the bound value before you use it.

## Let Parallel start the goroutines

When each source is a function, [Parallel.race](../../lib/Parallel/race.md) starts one goroutine per
function, returns the first value, and cancels the rest. The first value wins even when it is an
`Err`. To wait for the first `Ok` instead, use [Parallel.any](../../lib/Parallel/any.md):

```ant
func lookup(server: string, delay: Time.Duration, ctx: Context): Result[string, Error] {
  Time.sleep(delay, ctx: ctx)
  if ctx.isDone() { return Result.fail("${server}: ${ctx.reason()}") }
  if server == "a" { return Result.fail("a: refused") }
  Result.ok("${server}: 93.184.216.34")
}

let first = Parallel.race([
  { lookup("a", 10ms, x) },
  { lookup("b", 50ms, x) },
])
println(first)
let firstOk = Parallel.any([
  { lookup("a", 10ms, x) },
  { lookup("b", 50ms, x) },
])
println(firstOk)
```

```text
Err{ error: Error{ message: "a: refused" } }
Ok{ value: "b: 93.184.216.34" }
```

Each lambda gets the scope context as `x`; pass it to every wait. Both calls return only after every
function has ended, so a function that ignores `x` holds the call until it finishes.

Pass `ctx:` to bound the whole call. Both functions return `null` when it ends before a winner, and
when every callback fails, `any` returns `Err` of every error, in input order:

```ant
func lookup(server: string, delay: Time.Duration, ctx: Context): Result[string, Error] {
  Time.sleep(delay, ctx: ctx)
  if ctx.isDone() { return Result.fail("${server}: ${ctx.reason()}") }
  if server != "b" { return Result.fail("${server}: refused") }
  Result.ok("${server}: 93.184.216.34")
}

let ctx = Context.new(100ms)
defer { ctx.cancel() }
let reply = Parallel.any([
  { lookup("a", 10ms, x) },
  { lookup("b", 1s, x) },
], ctx: ctx)
println(reply ?? "no answer in time")
let refused = Parallel.any([
  { lookup("c", 20ms, x) },
  { lookup("a", 10ms, x) },
])
println(refused)
```

```text
no answer in time
Err{ error: [Error{ message: "c: refused" }, Error{ message: "a: refused" }] }
```

The errors keep the input order: `c` is first although it failed last.

## See also

[Channel.new](../../lib/Channel/new.md) · [Channel.send](../../lib/Channel/send.md) ·
[Channel.receive](../../lib/Channel/receive.md) · [Context.new](../../lib/Context/new.md) ·
[Context.cancel](../../lib/Context/cancel.md) · [Context.isDone](../../lib/Context/isDone.md) ·
[Context.reason](../../lib/Context/reason.md) · [Context.wait](../../lib/Context/wait.md) ·
[Time.sleep](../../lib/Time/sleep.md) · [Parallel.race](../../lib/Parallel/race.md) ·
[Parallel.any](../../lib/Parallel/any.md) · [`select`](../../language/keywords/select.md) ·
[`go`](../../language/keywords/go.md) · [`is`](../../language/operators/is.md) ·
[Give up after a timeout](timeout.md) · [Cancel a group of goroutines](cancel.md) ·
[Concurrency](../../guide/concurrency.md)
