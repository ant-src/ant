<!-- status: reviewed -->

# Concurrency

Ant runs concurrent work on goroutines. Any function can start one, and any function can block
inside one: there are no `async` functions, no `await`, and no promises. A blocking call such as a
file read or a `Time.sleep` blocks only the goroutine that makes it, and the others keep running.

The rest is a small set of parts that each do one job:

- `Parallel` runs a batch of work, waits for all of it, and hands you the results.
- `go` starts one goroutine and returns at once, with no handle and no result.
- A `WaitGroup` waits for goroutines that `go` started.
- `Mutex`, `RwMutex`, `Atomic`, `Once`, and `Cond` guard memory that goroutines share.
- A `Channel` carries values from one goroutine to another.
- A `Context` puts a deadline on a wait, or cancels it.
- A `select` waits for the first of several channels, contexts, groups, or timers.

This chapter shows how they fit together and which one to pick.

## Start with Parallel

When you know the shape of the work when it starts, such as every element of an array or a few
different calls, use `Parallel`. Each of its calls starts the goroutines, waits for all of them, and
returns only when every one has ended:

```ant
func wordCount(text: string, _: Context): int { text.split(" ").length() }

let pages = ["one two", "three four five", "six"]
let counts = Parallel.map(pages, wordCount, limit: 2) ?? []
println(counts)
```

```text
[2, 3, 1]
```

`Parallel.map` calls `wordCount` on every page, at most `limit: 2` at once, and returns the results
in input order, whichever finished first. `limit` is required, so the choice of how many run at once
stays visible: pass `pages.length()` for one goroutine per element. The result is `null` only when
the run is cancelled before every element starts, which `?? []` covers.

The callback takes a second parameter, a context. Pass it to every wait inside the callback, so the
call can stop the callback early:

```ant
func slowSquare(n: int, work: Context): int {
  Time.sleep(Time.duration(milliseconds: 10 * n), ctx: work)
  n * n
}

println(Parallel.map([3, 1, 2], slowSquare, limit: 3))
```

```text
[9, 1, 4]
```

A function that takes a caller's context passes it to the `Parallel` call as `ctx:`, so the call
stops admitting work when the caller's context is done. The call's own scope context, the one its
callbacks get, is a child of `ctx` and is cancelled on every exit; `ctx` itself is never cancelled:

```ant
func slowSquare(n: int, work: Context): int {
  Time.sleep(Time.duration(milliseconds: 10 * n), ctx: work)
  n * n
}

func squares(xs: Array[int], ctx: Context): Array[int] | null {
  Parallel.map(xs, slowSquare, limit: 4, ctx: ctx)
}

let request = Context.new(1s)
println(squares([1, 2, 3], request))
```

```text
[1, 4, 9]
```

The other `Parallel` functions cover the other shapes:

| You have                                  | Use                                 |
| ----------------------------------------- | ----------------------------------- |
| An array, and want a result per element   | `Parallel.map(xs, f, limit: n)`     |
| An array or a channel of jobs, no results | `Parallel.each(xs, f, limit: n)`    |
| A few different calls                     | `Parallel.run([{ ... }, { ... }])`  |
| Several ways to get one value, first wins | `Parallel.race([{ ... }, { ... }])` |
| Several sources, the first success wins   | `Parallel.any([{ ... }, { ... }])`  |

The callbacks of `run`, `race`, and `any` are lambdas whose `x` is the context to pass to their
waits. `race` and `any` cancel that context once they have a winner, so the slower callbacks stop:

```ant
let answer = Parallel.race([
  {
    Time.sleep(1s, ctx: x)
    "from the slow mirror"
  },
  { "from the cache" },
])
println(answer ?? "cancelled")
```

```text
from the cache
```

`Parallel.run` hands back nothing, so each callback writes its own field of a record, and you read
the fields after the call with no lock:

```ant
type Profile = { name: string, posts: int }

func loadName(_: Context): string { "Ann" }
func countPosts(_: Context): int { 8 }

let profile = Profile{ name: "", posts: 0 }
Parallel.run([
  { profile.name = loadName(x) },
  { profile.posts = countPosts(x) },
])
println(profile)
```

```text
Profile{ name: "Ann", posts: 8 }
```

Every action of a callback happens before the `Parallel` call returns, which is what makes that read
safe. Two callbacks that write the same field still need a lock.

## Goroutines with go

`go` runs a block on a new goroutine and returns at once. It gives you no handle to join and no
value to read, and the program exits when the script ends, whether or not other goroutines are still
running. So a goroutine that the script needs to finish is counted on a `WaitGroup`:

```ant
func report(name: string): void { println("report for ${name}") }

let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  report("ann")
}
wg.wait()
println("all reports sent")
```

```text
report for ann
all reports sent
```

Call `add` before the `go`, so `wait` cannot find the count at zero before the goroutine is counted,
and put `done` in a `defer` right after the goroutine starts, so it runs even when the goroutine
panics.

Use `go` for work that has its own lifetime, such as a producer feeding a channel, a background
loop, or a goroutine per connection. For a batch of work that you wait for, `Parallel` does the
counting for you.

### Handing back a result

A goroutine cannot hand a result back by assigning a local of its parent. The type checker does not
narrow a local that a `go` block assigns, even after a `wait`, because it cannot see the wait:

```ant
let wg = WaitGroup.new()
var answer: string | null = null
wg.add(1)
go {
  defer { wg.done() }
  answer = "42"
}
wg.wait()
if answer != null {
  println(answer.length())
  // error: string | null has no field or function length on every member: narrow the value first
}
```

Write the result into a container instead: an array slot, a map entry, a record field, or an
`Atomic`. After the wait, read it into a fresh local, which narrows as usual:

```ant
let words = ["ant", "bee", "wasp"]
let lengths: Array[int | null] = Array.filled(words.length(), null)
let wg = WaitGroup.new()
for entry in words.entries() {
  wg.add(1)
  go {
    defer { wg.done() }
    lengths[entry.index] = entry.value.length()
  }
}
wg.wait()
let last = lengths[2]
if last != null {
  println("wasp has ${last} letters")
}
```

```text
wasp has 4 letters
```

This is `Parallel.map` written by hand, and it shows why `Parallel.map` is the better choice: the
hand-written version starts every goroutine at once, and a goroutine that panics leaves its slot
`null`, where `Parallel.map` honors `limit` and raises the panic in the caller.

## Shared memory

Goroutines share memory. Two goroutines that touch the same variable, field, element, map, or set,
at least one of them writing, with nothing ordering the two, are a data race. A data race is a bug.
The interpreter never crashes because of one, and a racy read of a variable sees some value that was
stored there, but which one is not certain: a counter that two goroutines increment without a lock
can end up short.

A local `let` never races: it is written once, before any goroutine started after it can read it.
Locals, parameters, and loop variables that a `go` block reads are safe for the same reason. What a
`let` holds is another matter, since an array or record can still change.

### Mutex

A `Mutex` lets one goroutine at a time into a piece of code. `withLock` takes the lock, runs the
function, and releases the lock even when the function panics:

```ant
type Account = { balance: int, history: Array[int] }

let account = Account{ balance: 0, history: [] }
let lock = Mutex.new()
Parallel.each([50, -20, 30], func(amount: int, _: Context) {
  lock.withLock({
    account.balance += amount
    account.history.push(amount)
  })
}, limit: 3)
println(account.balance)
println(account.history.length())
```

```text
60
3
```

The balance and the history change together, so one lock covers both. Prefer `withLock` to a `lock`
and `unlock` pair: it cannot forget the unlock. A mutex is not reentrant: a goroutine that locks a
mutex it already holds waits forever, so a function called under the lock must not take it again.

`RwMutex` lets many readers in at once, or one writer. Use it when reads far outnumber writes;
otherwise a `Mutex` is simpler.

### Atomic

An `Atomic` holds one value that goroutines load, store, add to, or compare and swap without a lock:

```ant
let hits = Atomic.new(0)
Parallel.each(Iter.range(0, 100).toArray(), func(_: int, _: Context) {
  hits.add(1)
}, limit: 8)
println(hits.load())
```

```text
100
```

An `Atomic` guards one value. When two values must change together, as the balance and the history
above, use a mutex.

### Once

`Once.value` turns a loader into a function that runs it on the first call and returns the cached
value after that, however many goroutines call it at once:

```ant
type Config = { name: string }

func loadConfig(): Config {
  println("loading")
  Config{ name: "ant" }
}

let config = Once.value(loadConfig)
Parallel.each([1, 2, 3], func(_: int, _: Context) { config() }, limit: 3)
println(config().name)
```

```text
loading
ant
```

`Cond` is for a goroutine that waits until some shared state changes. Prefer a channel or a
`WaitGroup` when one of them can express the wait; it says the same thing more directly.

### What orders goroutines

A read in one goroutine is certain to see a write in another only when something orders the two.
These do:

- `go` comes before everything the new goroutine does.
- An unlock comes before the next lock of the same mutex.
- A `WaitGroup.done` comes before the `wait` it unblocks returns.
- A channel send comes before the receive that takes the value.
- A channel close comes before the receive that returns `null` because of it.
- `Context.cancel`, or a passed deadline, comes before the return of every wait the context ends.
- Every action of a `Parallel` callback comes before the call returns.
- An `Atomic` store comes before the load that sees it, and the first run of a `Once` before every
  call returns.

A goroutine ending orders nothing. To use what a goroutine wrote, wait for it on a `WaitGroup`, a
channel, or a `Parallel` call.

## Channels

A channel carries values from one goroutine to another, and blocks whichever side is ahead. Use a
channel when goroutines pass work or results along, and shared memory with a lock when they update
the same state.

```ant
let lines: Channel[string] = Channel.new(capacity: 2)
go {
  defer { lines.close() }
  for n in Iter.range(1, 4) {
    lines.send("line ${n}")
  }
}
for line in lines.values() {
  println(line)
}
```

```text
line 1
line 2
line 3
```

- `Channel.new` cannot tell what the channel carries, so annotate it: `Channel[string]`.
- A channel with capacity 0, the default, hands each value over directly: `send` returns when a
  receiver has taken it. A channel with capacity `n` holds up to `n` values before a send waits.
- The sender closes the channel when it has nothing more to send. `values` ends once the channel is
  closed and drained, which ends the `for` loop. Leave closing to the sender: a send on a closed
  channel panics, even one already waiting, so a receiver that closes the channel can crash a sender
  that is still sending.
- A receive on a closed, drained channel returns `null` at once. A send on a closed channel panics.

A channel of jobs pairs with `Parallel.each`, which runs a worker on each value as it arrives, at
most `limit` at once, and returns when the channel closes:

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

If the call ends early, on a panic or cancellation, a producer blocked in `send` stays blocked: give
it a context, `jobs.send(n, ctx: ctx)`, and cancel that context when the consumer is done.

## Contexts

A context bounds a wait. It becomes done once, for one of three reasons: you cancel it, its timeout
passes, or its parent is done. Every blocking library call takes one as its last parameter, `ctx:`,
and gives up when the context is done. Locks, `Cond.wait`, `Once`, and `Context.wait` itself take
none.

What giving up looks like follows the call's return type. A call that returns a `Result` returns an
`Err` whose message is the context's reason. A call that returns a value or `null` returns `null`. A
call that returns `void`, such as `Time.sleep`, returns early, so ask `isDone` when you need to
know:

```ant
let replies: Channel[string] = Channel.new()
let ctx = Context.new(20ms)
let reply = replies.receive(ctx: ctx)
println(reply ?? "no reply")
println(ctx.reason())

let nap = Context.new(10ms)
Time.sleep(1s, ctx: nap)
println(if nap.isDone() { "woke early: ${nap.reason()}" } else { "slept" })
```

```text
no reply
timeout
woke early: timeout
```

Contexts form a tree. `Context.new(parent: ctx)` makes a child that ends when its parent does, so a
function can put a tighter deadline on one step without losing the caller's cancellation. Every
context descends from the run's, so when the run ends, every wait ends:

```ant
func step(name: string, ctx: Context): string {
  let local = Context.new(50ms, parent: ctx)
  Time.sleep(1s, ctx: local)
  "${name}: ${local.reason()}"
}

let job = Context.new()
println(step("first", job))
job.cancel()
println(step("second", job))
```

```text
first: timeout
second: cancelled
```

A goroutine that should stop when asked passes its context into every wait and checks `isDone`
between steps. Cancellation is cooperative: a goroutine that ignores its context runs until it ends
on its own.

## select

A `select` waits for the first of several things and runs the block of the one that happens. Each
head is a blocking call written in namespace form: `Channel.receive(c)`, `Channel.send(c, v)`,
`Context.wait(ctx)`, `WaitGroup.wait(wg)`, or `Time.sleep(d)`. The call is not made until the
`select` picks its clause, and `let name =` before a head binds what it returns:

```ant
let replies: Channel[string] = Channel.new()
go {
  Time.sleep(10ms)
  replies.send("pong")
}
let reply = select {
  let r = Channel.receive(replies) { r ?? "closed" }
  Time.sleep(1s) { "timed out" }
}
println(reply)
```

```text
pong
```

A `select` is an expression, so its value here is the value of the block that ran. With an `else`
clause, written last, a `select` never waits: `else` runs when no head can proceed at once.

A goroutine that handles several sources runs a `select` in a loop. A `break` or `return` in a
clause acts on the loop or function around the `select`:

```ant
let events: Channel[string] = Channel.new()
let ctx = Context.new(1s)
go {
  defer { events.close() }
  for event in ["start", "tick", "stop"] {
    events.send(event)
  }
}
var seen = 0
while true {
  select {
    let event = Channel.receive(events) {
      if event == null { break }
      seen += 1
    }
    Context.wait(ctx) { break }
  }
}
println("seen ${seen} events")
```

```text
seen 3 events
```

The `if event == null { break }` matters. A closed channel's receive is always ready, so a loop that
keeps receiving from one spins on `null` forever. When several heads can proceed at once, `select`
picks one at random, so do not rely on the order of the clauses.

`Time.ticker(d, ctx: ctx)` is a channel that receives the time every `d` until `ctx` is done, for a
`select` that does something on a schedule beside its other work. Without a context it ticks until
the run ends, and it is never closed.

## Panics in goroutines

A panic in a `go` goroutine ends that goroutine only. Its deferred blocks run, the runtime reports
the message and trace on standard error, and the rest of the program goes on:

```ant
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  let xs = [1, 2, 3]
  println(xs[5]) // this goroutine panics: index 5 out of range for an array of length 3
}
wg.wait()
println("the program goes on")
```

```text
the program goes on
```

The `defer` is what lets `wait` return: without it, the count never reaches zero and the main
goroutine waits forever. A goroutine that works through many jobs calls each one through
`Panic.catch`, so a bug in one job ends that job and the goroutine goes on to the next:

```ant
func process(job: int): int {
  if job == 2 {
    panic("job ${job} is broken")
  }
  job * 10
}

let jobs: Channel[int] = Channel.new()
go {
  defer { jobs.close() }
  for n in [1, 2, 3] {
    jobs.send(n)
  }
}
for job in jobs.values() {
  let result = Panic.catch({ process(job) })
  if result is Ok {
    println("job ${job}: ${result.value}")
  } else {
    println("job ${job} failed: ${result.error.message}")
  }
}
```

```text
job 1: 10
job 2 failed: job 2 is broken
job 3: 30
```

`Parallel` treats a panic differently. A panic in a callback cancels the call's context, the call
waits for the other callbacks, and then raises the panic in the code that made the call, with its
original message and trace. So a panic in a `Parallel` callback behaves as a panic in an ordinary
function call:

```ant
func half(n: int, _: Context): int {
  if n % 2 != 0 {
    panic("${n} is odd")
  }
  n / 2
}

let result = Panic.catch({ Parallel.map([2, 3, 4], half, limit: 3) })
println(if result is Err { result.error.message } else { "no panic" })
```

```text
3 is odd
```

## Interrupts

When you press Ctrl-C, or the process gets SIGINT or SIGTERM, `ant` ends the run: every goroutine
stops at its next call or loop iteration, blocking calls return with the reason `cancelled`, and the
process exits. A program that needs to finish cleanly, such as a server that drains its connections
or a batch job that saves its progress, asks for the interrupt instead by calling `Os.interrupt()`.
The first interrupt then makes that context done with the reason `interrupt`, and the program goes
on:

```ant
func save(job: int): void { println("saved job ${job}") }

let stop = Os.interrupt()
var finished = 0
for job in Iter.range(1, 4) {
  if stop.isDone() { break }
  save(job)
  finished += 1
}
println("finished ${finished} jobs, interrupted: ${stop.isDone()}")
```

```text
saved job 1
saved job 2
saved job 3
finished 3 jobs, interrupted: false
```

Pass `stop` to the waits that should end on Ctrl-C, or make it the parent of the contexts you
create. `Context.wait(Os.interrupt())` is how a main goroutine waits while a server runs. A second
interrupt ends the run, so a program that hangs while shutting down can still be stopped.

## Pitfalls

Coming from other languages, these are the surprises:

- **The program does not wait.** It exits when the script ends, with goroutines still running. Wait
  on a `WaitGroup` or a channel, or use `Parallel`.
- **There is no `await` and no join.** A `go` block has no result. Use `Parallel.map`, or write into
  a container and wait.
- **A goroutine's panic does not stop the program.** It ends the goroutine and appears on standard
  error. Put `done` in a `defer`, and catch per job in long-running workers.
- **Nothing detects a deadlock.** Goroutines that all wait on each other wait until the run is
  interrupted. Bound waits with a context where an answer may never come.
- **Locks are not reentrant.** Locking a mutex you hold waits forever.
- **A send on a closed channel panics**, and a receive on one returns `null` at once. Let the sender
  close it, and leave a `select` loop on `null`.
- **A context only stops code that passes it on.** A callback that sleeps without its context holds
  its `Parallel` call until the sleep ends.

## See also

- Tour: [Goroutines](../tour/17-goroutines.md) and [Channels and select](../tour/18-channels.md)
- Reference: [`go`](../language/keywords/go.md), [`select`](../language/keywords/select.md),
  [`defer`](../language/keywords/defer.md)
- Library: [Concurrency](../lib/guides/concurrency.md), [Parallel](../lib/Parallel/index.md),
  [Channel](../lib/Channel/index.md), [Context](../lib/Context/index.md),
  [WaitGroup](../lib/WaitGroup/index.md), [Mutex](../lib/Mutex/index.md),
  [Atomic](../lib/Atomic/index.md), [Once](../lib/Once/index.md),
  [Os.interrupt](../lib/Os/interrupt.md), [Panic.catch](../lib/Panic/catch.md)
- How-to: [process a list in parallel](../howto/concurrency/parallel-map.md),
  [run a worker pool](../howto/concurrency/worker-pool.md),
  [give up after a timeout](../howto/concurrency/timeout.md),
  [cancel a group of goroutines](../howto/concurrency/cancel.md),
  [protect shared state](../howto/concurrency/shared-state.md)
- [Errors](errors.md): `Result`, `panic`, and `Panic.catch`
