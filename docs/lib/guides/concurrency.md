<!-- status: reviewed -->

# Concurrency

Ant runs concurrent work on goroutines. A [`go`](../../language/keywords/go.md) block starts one and
returns at once, with no handle to join and no result to read. The program ends when the script
ends, without waiting for goroutines still running, so wait for them with a `WaitGroup` or a
channel. A panic in a goroutine ends only that goroutine, and the runtime reports it on standard
error. The standard library gives you four kinds of tools to coordinate them:

- [`Parallel`](../Parallel/index.md) runs work whose shape you know when it starts: every element of
  an array, every job on a channel, or a few different tasks. Each call starts its goroutines, waits
  for all of them, and raises a callback's panic in the caller.
- [`Channel`](../Channel/index.md) carries values from one goroutine to another and blocks the side
  that is ahead. A [`select`](../../language/keywords/select.md) waits on several channels, a
  context, a wait group, or a timer at once.
- [`Context`](../Context/index.md) bounds a wait with a deadline or ends it with a cancellation.
  Every blocking library call takes one as its last parameter, `ctx:`, except the locks and
  `Cond.wait`.
- The sync types guard memory that goroutines share: [`Mutex`](../Mutex/index.md),
  [`RwMutex`](../RwMutex/index.md), [`Atomic`](../Atomic/index.md), [`Once`](../Once/index.md),
  [`Cond`](../Cond/index.md), and [`WaitGroup`](../WaitGroup/index.md), which waits for goroutines
  to finish.

## Picking a tool

### Work over an array

`Parallel.map` calls a function on every element, at most `limit` at once, and returns the results
in input order. The callback gets a context to pass to its own waits. A lambda takes one parameter,
so the callback is a `func` with two. The result is `null` when the context ends before every
element starts, so give it a fallback:

```ant
let words = ["ant", "bee", "wasp"]
let lengths = Parallel.map(words, func(word: string, _: Context) { word.length() }, limit: 2) ?? []
println(lengths.sum())
```

```text
10
```

### A stream of values

A goroutine sends on a channel and closes it when it is done. A `for` loop over `Channel.values`
receives until the close:

```ant
let names: Channel[string] = Channel.new()
go {
  defer { names.close() }
  for name in ["ant", "bee", "wasp"] {
    names.send(name.toUpper())
  }
}
for name in names.values() {
  println(name)
}
```

```text
ANT
BEE
WASP
```

### A worker pool

A goroutine feeds jobs into a channel and closes it, and `Parallel.each` hands each job to one of at
most `limit` workers. `each` stops when the channel closes; it does not close the channel itself:

```ant
let jobs: Channel[int] = Channel.new()
go {
  defer { jobs.close() }
  for n in [1, 2, 3, 4] {
    jobs.send(n)
  }
}
let total = Atomic.new(0)
Parallel.each(jobs, func(n: int, _: Context) { total.add(n * n) }, limit: 2)
println(total.load())
```

```text
30
```

### A bounded wait

A context with a timeout ends the wait. A receive that gives up returns `null`, and `reason` says
why:

```ant
let replies: Channel[string] = Channel.new()
let ctx = Context.new(20ms)
println(replies.receive(ctx: ctx) ?? "no reply")
println(ctx.reason())
```

```text
no reply
timeout
```

### The first of several things

A `select` takes whichever head can proceed first. Its heads are written in namespace form:

```ant
let replies: Channel[string] = Channel.new()
go {
  Time.sleep(200ms)
  replies.send("late")
}
let reply = select {
  let r = Channel.receive(replies) { r }
  Time.sleep(20ms) { "timed out" }
}
println(reply)
```

```text
timed out
```

For the first of several tasks, use `Parallel.race`, which cancels the others once one returns. Each
task is a lambda whose `x` is the context to pass to its waits:

```ant
let winner = Parallel.race([
  {
    Time.sleep(200ms, ctx: x)
    "slow"
  },
  { "fast" },
])
println(winner ?? "cancelled")
```

```text
fast
```

`Parallel.any` works the same way for tasks that return a `Result`, and gives the first `Ok`.

### Shared data

Goroutines that write the same map take a mutex around each change. A wait group tells the main
goroutine when they have all finished, and `defer` counts a goroutine done even if it panics:

```ant
let counts: Map[string, int] = Map{}
let lock = Mutex.new()
let wg = WaitGroup.new()
for word in ["ant", "bee", "ant", "wasp", "ant"] {
  wg.add(1)
  go {
    defer { wg.done() }
    lock.withLock({ counts[word] = (counts[word] ?? 0) + 1 })
  }
}
wg.wait()
println(counts["ant"])
```

```text
3
```

A goroutine hands back a result through a container, such as an array slot, a map entry, a record
field, or an `Atomic`, not by reassigning a local of its parent.

## Functions by job

### Starting work

- [go](../../language/keywords/go.md): run a block on a new goroutine and return at once
- [Parallel.run](../Parallel/run.md): call every function at once and wait for all of them

### Working through many values

- [Parallel.map](../Parallel/map.md): apply a function to every element, at most `limit` at once,
  results in input order
- [Parallel.each](../Parallel/each.md): call a function on every element of an array, or every value
  of a channel, at most `limit` at once

### Taking the first result

- [Parallel.race](../Parallel/race.md): the first value any function returns, `Err` included
- [Parallel.any](../Parallel/any.md): the first `Ok`, or `Err` of every error when all fail
- [select](../../language/keywords/select.md): wait for the first of several channels, contexts,
  wait groups, or timers

### Passing values between goroutines

- [Channel.new](../Channel/new.md): an open channel that buffers up to `capacity` values; 0 is a
  rendezvous
- [Channel.send](../Channel/send.md): send a value, waiting until it is taken or buffered
- [Channel.receive](../Channel/receive.md): the next value, or `null` once the channel is closed and
  drained
- [Channel.values](../Channel/values.md): a sequence of every value until the channel is closed and
  drained
- [Channel.close](../Channel/close.md): close the channel; receivers drain it and then get `null`
- [Channel.isClosed](../Channel/isClosed.md): whether the channel has been closed
- [Channel.length](../Channel/length.md): the number of values buffered now
- [Time.ticker](../Time/ticker.md): a channel that receives the current instant at intervals

### Bounding and cancelling waits

- [Context.new](../Context/new.md): a child of a parent or of the run, with an optional timeout
- [Context.cancel](../Context/cancel.md): make the context done now, with the reason `"cancelled"`
- [Context.isDone](../Context/isDone.md): whether the context is done
- [Context.reason](../Context/reason.md): why it is done, or `null` while it runs
- [Context.deadline](../Context/deadline.md): the earliest deadline of the context and its parents
- [Context.wait](../Context/wait.md): wait until the context is done; also a head of `select`
- [Os.interrupt](../Os/interrupt.md): the context that Ctrl-C, SIGINT, or SIGTERM ends
- [Test.context](../Test/context.md): the context that ends with the current test
- [Time.sleep](../Time/sleep.md): block for a duration, or until the context is done

### Waiting for goroutines to finish

- [WaitGroup.new](../WaitGroup/new.md): a new group, its counter at zero
- [WaitGroup.add](../WaitGroup/add.md): add `n`, which may be negative, to the counter
- [WaitGroup.done](../WaitGroup/done.md): take one off the counter, as a goroutine finishes
- [WaitGroup.wait](../WaitGroup/wait.md): block until the counter is zero, or until `ctx` is done

### Guarding shared data

- [Mutex.new](../Mutex/new.md): a new, unlocked mutex
- [Mutex.withLock](../Mutex/withLock.md): lock, run a function, unlock even if it panics, and return
  its result
- [Mutex.lock](../Mutex/lock.md): lock, waiting while the mutex is locked
- [Mutex.unlock](../Mutex/unlock.md): unlock; panics if the mutex is not locked
- [RwMutex.new](../RwMutex/new.md): a new, unlocked reader-writer lock
- [RwMutex.withRead](../RwMutex/withRead.md): run a function under the read lock and return its
  result
- [RwMutex.withWrite](../RwMutex/withWrite.md): run a function under the write lock and return its
  result
- [RwMutex.readLock](../RwMutex/readLock.md): lock for reading, waiting while a writer holds the
  lock or waits for it
- [RwMutex.readUnlock](../RwMutex/readUnlock.md): release one read lock
- [RwMutex.writeLock](../RwMutex/writeLock.md): lock for writing, waiting while any reader or writer
  holds it
- [RwMutex.writeUnlock](../RwMutex/writeUnlock.md): release the write lock

### Changing one value without a lock

- [Atomic.new](../Atomic/new.md): a new cell holding a value
- [Atomic.load](../Atomic/load.md): the value in the cell
- [Atomic.store](../Atomic/store.md): set the value in the cell
- [Atomic.add](../Atomic/add.md): add to a number and return the new value; panics on overflow
- [Atomic.compareAndSwap](../Atomic/compareAndSwap.md): set the value only if it equals an expected
  one

### Doing work once

- [Once.new](../Once/new.md): a new once that has not run yet
- [Once.do](../Once/do.md): run a function on the first call only; later calls wait for it and do
  nothing
- [Once.value](../Once/value.md): a function that runs the loader on its first call and caches its
  value

### Waiting for a condition

- [Cond.new](../Cond/new.md): a new cond over a mutex
- [Cond.wait](../Cond/wait.md): release the mutex, wait for a wake-up, and lock the mutex again
- [Cond.signal](../Cond/signal.md): wake one waiting goroutine, if any
- [Cond.broadcast](../Cond/broadcast.md): wake every waiting goroutine

### Containing panics

- [Panic.catch](../Panic/catch.md): call a function and return its panic as an `Err`, so a worker
  goes on to its next job
- [Panic.raise](../Panic/raise.md): panic again with a caught panic's message and trace

## See also

- [Concurrency](../../guide/concurrency.md): how goroutines, channels, contexts, and the sync types
  fit together
- How-to guides: [process a list in parallel](../../howto/concurrency/parallel-map.md),
  [run a worker pool](../../howto/concurrency/worker-pool.md),
  [give up after a timeout](../../howto/concurrency/timeout.md),
  [take the first result](../../howto/concurrency/first-result.md),
  [cancel a group of goroutines](../../howto/concurrency/cancel.md),
  [protect shared state](../../howto/concurrency/shared-state.md),
  [run something periodically](../../howto/concurrency/ticker.md)
- [Time](../Time/index.md): durations, sleeps, and tickers
