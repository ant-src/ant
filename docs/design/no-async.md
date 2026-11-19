<!-- status: reviewed -->

# Why goroutines and not async

Ant has no `async` functions, no `await`, and no promises. Any function can wait, and any function
can start a goroutine. A call that waits, such as a file read, a `Time.sleep`, or a channel receive,
blocks only the goroutine that makes it, and the other goroutines keep running. This page explains
why Ant chose goroutines over `async`, and what the choice costs.

## Async colors the call chain

In a language with `async`, a function that waits for I/O is marked `async` and returns a promise.
Its caller must `await` the promise, which makes the caller `async` too, and so on up to the top of
the program. The mark spreads through every function between the wait and `main`. Functions come in
two colors, and a function of one color cannot call one of the other the way it calls its own.

The split reaches past your own code. A callback passed to `map` or `sort` must be of the color the
library expects, so `map` with an `async` callback returns an array of promises, not of values. A
library ships each operation twice, one blocking and one `async`, or picks a color for all of its
users. Adding one wait deep inside a function changes the signature of every caller above it.

Ant has one kind of function. Waiting is not part of a function's type, so it does not spread.
Goroutines are cheap, so a goroutine that blocks costs little, and the program needs no second kind
of function to avoid blocking.

## A blocking call is an ordinary call

<!-- test: file a.txt
one two three
-->
<!-- test: file b.txt
four five
-->

```ant
func countWords(path: string): Result[int, Error] {
  let text = try File.read(path)
  Result.ok(text.fields().length())
}

let counts = ["a.txt", "b.txt"].map(countWords)
println(counts)
```

```text
[Ok{ value: 3 }, Ok{ value: 2 }]
```

`countWords` waits for the disk, and nothing in its signature says so. `Array.map` takes it as it
takes any function, and the result is an array of the values `countWords` returns. The same holds
for every callback in the language: a sort key, a filter, a sequence, a test, or a deferred block
may wait, and the code around it does not change.

The standard library has one `File.read`, not a blocking one and an `async` one. A function you
write is usable from any caller the moment it compiles, whether it waits or not.

## The caller chooses concurrency

Since a function does not decide whether it runs concurrently, its caller does. The same function
runs one call after another, or many at once:

```ant
func price(item: string, ctx: Context | null = null): int {
  Time.sleep(10ms, ctx: ctx) // stands in for a network call
  item.length() * 100
}

let items = ["tea", "coffee", "milk"]
println(items.map({ price(x) }))
println(Parallel.map(items, price, limit: 3))
```

```text
[300, 600, 400]
[300, 600, 400]
```

`items.map` calls `price` three times in a row. `Parallel.map` calls it on three goroutines at once,
waits for all of them, and returns the results in input order. Neither call changes `price`. The
decision to run work concurrently sits where the work is started, which is where the reader looks
for it.

`Parallel` is the counterpart of `Promise.all` and its relatives, with two differences. Each call
returns only when every goroutine it started has ended, so no work outlives the call by accident.
And a panic in one of them is raised again in the caller, so a bug in concurrent code stops the
caller as a bug in an ordinary call does. `Parallel.race` takes the first value to return and
`Parallel.any` the first success, and a `select` waits for the first of several channels, contexts,
or timers.

## Cancellation is a value you pass

A promise has no built-in way to stop the work behind it, so `async` code passes a separate signal
when it needs one. Ant makes that signal the one mechanism: a `Context`. Blocking calls such as
`Channel.receive`, `Time.sleep`, and `Http.get` take one as their last parameter, `ctx:`, and give
up when it is done:

```ant
func waitForReply(replies: Channel[string], ctx: Context): string {
  replies.receive(ctx: ctx) ?? "no reply: ${ctx.reason()}"
}

let replies: Channel[string] = Channel.new()
println(waitForReply(replies, Context.new(20ms)))
```

```text
no reply: timeout
```

A deadline, a cancel call, and the end of the run all reach a wait the same way. `Parallel.race`
uses it to stop the losers once it has a winner:

```ant
func fetchQuote(source: string, delay: Time.Duration, ctx: Context): string {
  Time.sleep(delay, ctx: ctx)
  "quote from ${source}"
}

let quote = Parallel.race([
  { fetchQuote("primary", 200ms, x) },
  { fetchQuote("mirror", 10ms, x) },
])
println(quote ?? "cancelled")
```

```text
quote from mirror
```

Each callback gets the race's own context as `x`. When the mirror answers first, the race cancels
that context, the primary's sleep returns early, and the call returns without waiting the full 200
milliseconds.

## No handle from go

`go` starts a goroutine and returns nothing: no handle to join, no future to read a result from. A
handle would bring promises back under another name, with the same question of who waits for it and
what happens to its failure when nobody does.

Ant answers those questions in two places instead. Work that the caller waits for goes through
`Parallel`, which joins, collects results, and raises panics in one call. Work with a lifetime of
its own, such as a producer feeding a channel, a background loop, or a goroutine per connection,
goes through `go`, and reports back through a channel, a `WaitGroup`, or shared memory.

```ant
let done: Channel[string] = Channel.new()
go {
  Time.sleep(10ms) // stands in for slow work
  done.send("report sent")
}
println(done.receive() ?? "closed")
```

```text
report sent
```

The goroutine sends its result on a channel, and the script waits on the receive.

## Goroutines are preemptive

A goroutine that never blocks cannot keep the others from running. A long computation on one
goroutine does not freeze a server's other requests, as it would on an event loop, where every task
waits until the running one reaches its next `await`.

The same property removes a guarantee that `async` code relies on. On a single-threaded event loop,
the code between two `await`s runs without interruption, so `count += 1` is safe there. In Ant
another goroutine may run between any two steps. This counter can lose updates, so the number it
prints can change from run to run:

<!-- test: check-only the count depends on goroutine timing -->

```ant
var count = 0
Parallel.each(Array.range(0, 1000), func(_: int, _: Context) {
  count += 1
}, limit: 8)
println(count)
```

Two goroutines that write the same variable with nothing ordering them are a data race, which is a
bug. The interpreter stays memory safe, but the result is not certain. Shared state needs a `Mutex`,
an `Atomic`, or a channel:

```ant
let count = Atomic.new(0)
Parallel.each(Array.range(0, 1000), func(_: int, _: Context) {
  count.add(1)
}, limit: 8)
println(count.load())
```

```text
1000
```

## What it costs

- **Shared memory needs synchronization.** Goroutines share memory and run preemptively, so any
  state two of them touch needs a lock, an `Atomic`, or a channel. `async` code on one thread gets
  that for free between its `await`s.
- **Waiting is invisible in a signature.** An `async` mark tells a reader that a function waits. In
  Ant only the body says so, and a `ctx` parameter is the closest hint.
- **Contexts are passed by hand.** A wait stops early only when its context reaches it. A function
  that sleeps or receives without passing `ctx` on holds its caller until the wait ends on its own.
- **`go` has no result.** The program exits when the script ends, whether or not goroutines are
  still running, and a panic in a `go` goroutine is reported on standard error rather than returned.
  Code that needs a goroutine's work waits on a `WaitGroup` or a channel, or uses `Parallel`.
- **Nothing detects a deadlock.** Goroutines that all wait on each other wait until the run ends.
  Bound a wait that may never end with a context.

## JavaScript habits are errors

`async` and `await` are reserved words. Neither can name a variable, function, type, or parameter,
so code written from habit stops at `ant check` with an error instead of meaning something else:

```ant
async func load(path: string): Result[string, Error] { // error: 'async' is a reserved word
  File.read(path)
}
```

Write the function without the mark. To run it alongside other work, call it from `Parallel` or a
`go` block. The same holds at the call site:

```ant
func load(path: string): Result[string, Error] {
  await File.read(path) // error: 'await' is a reserved word
}
```

Drop the `await`, since the call already waits.

## See also

The [concurrency guide](../guide/concurrency.md) shows how goroutines, channels, contexts, and
`Parallel` fit together, and the tour pages [Goroutines](../tour/17-goroutines.md) and
[Channels and select](../tour/18-channels.md) introduce them. Reference:
[`go`](../language/keywords/go.md), [`select`](../language/keywords/select.md),
[`Parallel`](../lib/Parallel/index.md), [`Context`](../lib/Context/index.md),
[`Channel`](../lib/Channel/index.md), [`Atomic`](../lib/Atomic/index.md),
[`Mutex`](../lib/Mutex/index.md). The design note [Why no exceptions](errors-are-values.md) covers a
related decision, and [Coming from JavaScript](../coming-from/javascript.md) lists the differences
side by side.
