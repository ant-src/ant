<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Parallel

Parallel runs work on goroutines and waits for all of it. Each call makes a scope context, a child
of ctx, and passes it to every callback; the call cancels it on every exit, success included, and
never cancels ctx. A call neither returns nor panics until every callback it started has ended,
deferred blocks included, and every action of a callback happens before the call returns, so a value
a callback wrote can be read after the call with no lock. If a panic leaves a callback, the call
cancels the scope, waits for the others, and raises that panic in the caller with its original
message and trace; a later panic in the same call is reported and dropped. A go inside a callback is
not owned by the call.

<!-- end generated -->

`Parallel` runs callbacks on goroutines and waits for every one of them before it returns. Each call
owns its goroutines: none outlives the call, a panic in one reaches the caller, and every callback
gets a context that ends when the work should stop. `map` works through an array and `each` through
an array or a channel, at most `limit` at once; `run` does a fixed list of jobs together, and `race`
and `any` take the first result and cancel the rest.

## Details

- Every callback receives the scope context: the second parameter of a `map` or `each` callback, and
  `x` in a `run`, `race`, or `any` lambda. Pass it to the waits inside the callback, so the callback
  stops when `ctx` ends, when another callback panics, or when `race` or `any` has a winner.
- Cancellation is cooperative: a callback that ignores its context holds the call until it ends.
- After the call, read what the callbacks wrote with no lock. Callbacks that write the same variable
  as each other still need a [Mutex](../Mutex/index.md) or an [Atomic](../Atomic/index.md).
- `limit` counts per call, so nested `map` or `each` calls multiply it.

## Basic examples

`Parallel.map` applies a function to every element, at most `limit` at once, and keeps the input
order:

```ant
let squares = Parallel.map([1, 2, 3, 4], func(n: int, _: Context) { n * n }, limit: 2)
println(squares)
```

```text
[1, 4, 9, 16]
```

`Parallel.run` does a few different jobs at once. Each one hands back its result by writing a field:

```ant
type Profile = { name: string | null, orders: Array[string] }

func loadName(id: int, ctx: Context): string {
  Time.sleep(20ms, ctx: ctx)
  "user-${id}"
}

func loadOrders(id: int, ctx: Context): Array[string] {
  Time.sleep(10ms, ctx: ctx)
  ["order-1", "order-2"]
}

let profile = Profile{ name: null, orders: [] }
Parallel.run([
  { profile.name = loadName(7, x) },
  { profile.orders = loadOrders(7, x) },
])
println(profile)
```

```text
Profile{ name: "user-7", orders: ["order-1", "order-2"] }
```

`Parallel.any` asks several sources and takes the first `Ok`; a failure only drops out:

```ant
func fetch(server: string, latency: Time.Duration, ctx: Context): Result[string, Error] {
  Time.sleep(latency, ctx: ctx)
  if server == "primary" { return Result.fail("${server}: unavailable") }
  Result.ok("reply from ${server}")
}

let reply = Parallel.any([
  { fetch("primary", 10ms, x) },
  { fetch("mirror", 30ms, x) },
])
println(reply)
```

```text
Ok{ value: "reply from mirror" }
```

## Scope

`Parallel.each` takes an array or a channel. Over a channel it runs a callback for every value until
the channel closes:

```ant
let jobs: Channel[string] = Channel.new(capacity: 4)
go {
  defer { jobs.close() }
  for name in ["a.txt", "b.txt", "c.txt"] {
    jobs.send(name)
  }
}

let done = Atomic.new(0)
Parallel.each(jobs, func(name: string, _: Context) { done.add(1) }, limit: 2)
println("${done.load()} jobs done")
```

```text
3 jobs done
```

The scope context is a child of `ctx`. The call cancels it when it returns, even on success, and
leaves `ctx` alone:

```ant
type Seen = { scope: Context | null }

let outer = Context.new()
let seen = Seen{ scope: null }
Parallel.run([{ seen.scope = x }], ctx: outer)
println(seen.scope?.reason())
println(outer.reason())
```

```text
cancelled
null
```

A panic in one callback cancels the scope, so its siblings stop. The call waits for them, then
panics in the caller with the original message:

```ant
let caught = Panic.catch({
  Parallel.run([
    { panic("disk full") },
    {
      x.wait()
      println("sibling stopped: ${x.reason() ?? "running"}")
    },
  ])
})
if caught is Err {
  println(caught.error.message)
}
```

```text
sibling stopped: cancelled
disk full
```

## Functions by job

### Work through many values

- [Parallel.map](map.md): apply a function to every element, at most `limit` at once, results in
  input order
- [Parallel.each](each.md): call a function on every element of an array, or every value of a
  channel, at most `limit` at once

### Do a fixed list of jobs

- [Parallel.run](run.md): call every function at once and wait for all of them

### Take the first result

- [Parallel.race](race.md): the first value any function returns, `Err` included
- [Parallel.any](any.md): the first `Ok`, or `Err` of every error when all fail

## Applications

### Stop at the first error

In `map` and `each`, an `Err` from a callback stops nothing by itself. To fail fast, own the context
and cancel it. With `limit: 1` the elements after the failure are never admitted, so the result is
`null`:

```ant
func validate(n: int): Result[int, Error] {
  if n < 0 { return Result.fail("negative: ${n}") }
  Result.ok(n * 10)
}

let stop = Context.new()
let results = Parallel.map([1, -2, 3, 4], func(n: int, _: Context): Result[int, Error] {
  let checked = validate(n)
  if checked is Err { stop.cancel() }
  checked
}, limit: 1, ctx: stop)
println(results ?? "stopped at the first error")
```

```text
stopped at the first error
```

### A deadline for a whole step

A deadline on `ctx` bounds every callback. `race` returns `null` when the deadline passes before any
callback returns:

```ant
let deadline = Context.new(20ms)
let answer = Parallel.race([
  {
    Time.sleep(1h, ctx: x)
    "computed"
  },
], ctx: deadline)
println(answer ?? "timed out")
```

```text
timed out
```

## Properties and relations

Every write of a callback is visible after the call. Callbacks that write different slots need no
lock:

```ant
let words = ["ant", "bee", "wasp"]
let lengths = Array.filled(words.length(), 0)
Parallel.each(Array.range(0, words.length()), func(i: int, _: Context) {
  lengths[i] = words[i].length()
}, limit: 3)
assert(lengths == words.map({ x.length() }))
```

For callbacks that ignore their context, `Parallel.map` gives what `Array.map` gives, whatever the
limit:

```ant
let xs = [5, 3, 8]
let doubled = Parallel.map(xs, func(n: int, _: Context) { n * 2 }, limit: 2)
assert(doubled == xs.map({ x * 2 }))
```

## Possible issues

### A callback that ignores its context

A callback that waits without its context holds the call. Here the deadline passes during the first
element, yet `map` returns only after that callback wakes:

```ant
let deadline = Context.new(30ms)
let done = Parallel.map([1, 2], func(n: int, _: Context): int {
  Time.sleep(100ms)
  println("woke ${n}")
  n
}, limit: 1, ctx: deadline)
println(done)
```

```text
woke 1
null
```

### A deadline after the last element started

`map` returns `null` only when some element never started. When every element has started, the call
returns every result, even if `ctx` ended meanwhile. The callback decides what an interrupted wait
gives:

```ant
let deadline = Context.new(30ms)
let sizes = Parallel.map(["small", "huge"], func(name: string, work: Context): int {
  Time.sleep(if name == "huge" { 1h } else { 1ms }, ctx: work)
  name.length()
}, limit: 2, ctx: deadline)
println(sizes)
```

```text
[5, 4]
```

Check `work.isDone()`, or return a `Result`, when a callback cut short must say so.

### Waiting on the caller's context

A callback that waits on the caller's `ctx` instead of its scope context does not see a sibling's
panic or a `race` or `any` winner. Here the slow callback sleeps on `outer`, so `race` has its
winner at once but returns only when `outer` times out:

```ant
let outer = Context.new(30ms)
let winner = Parallel.race([
  { "fast" },
  {
    Time.sleep(1h, ctx: outer)
    "slow"
  },
], ctx: outer)
println(winner)
println(outer.reason() ?? "outer still running")
```

```text
fast
timeout
```

Passing the scope context `x` stops the loser as soon as the winner returns, and leaves `outer`
running:

```ant
let outer = Context.new(30ms)
let winner = Parallel.race([
  { "fast" },
  {
    Time.sleep(1h, ctx: x)
    "slow"
  },
], ctx: outer)
println(winner)
println(outer.reason() ?? "outer still running")
```

```text
fast
outer still running
```

### A race whose value can be null

`race` and `any` return `null` when the scope ends before a winner, so when a callback's value can
be `null` too, the result cannot tell the two apart. Here the cache answers first with `null`, and
`race` returns `null` as it would on cancellation:

```ant
func lookup(source: string, latency: Time.Duration, ctx: Context): string | null {
  Time.sleep(latency, ctx: ctx)
  if source == "cache" { null } else { "value from ${source}" }
}

let found = Parallel.race([
  { lookup("cache", 1ms, x) },
  { lookup("database", 30ms, x) },
])
println(found ?? "missing or cancelled")
```

```text
missing or cancelled
```

Wrap the value in a record, so a `null` result means only cancellation:

```ant
type Found = { value: string | null }

func lookup(source: string, latency: Time.Duration, ctx: Context): string | null {
  Time.sleep(latency, ctx: ctx)
  if source == "cache" { null } else { "value from ${source}" }
}

let found = Parallel.race([
  { Found{ value: lookup("cache", 1ms, x) } },
  { Found{ value: lookup("database", 30ms, x) } },
])
println(found ?? "cancelled")
```

```text
Found{ value: null }
```

### A producer blocked after `each` ends

`each` neither closes nor drains its channel, and a producer blocked in `send` is not woken when
`each` returns. Give the producer the same context, so its `send` gives up when the work stops:

```ant
let deadline = Context.new(30ms)
let jobs: Channel[string] = Channel.new()
let producer = WaitGroup.new()
producer.add(1)
go {
  defer { producer.done() }
  defer { jobs.close() }
  for name in ["a.txt", "b.txt", "c.txt"] {
    if !jobs.send(name, ctx: deadline) {
      println("producer stopped at ${name}")
      break
    }
  }
}

Parallel.each(jobs, func(name: string, work: Context) { Time.sleep(1h, ctx: work) }, limit: 1, ctx: deadline)
producer.wait()
```

```text
producer stopped at b.txt
```

Without `ctx: deadline` on the `send`, the producer stays blocked on `b.txt` after `each` returns.

### A `go` inside a callback

The call waits only for its own callbacks. A goroutine a callback starts with `go` may still run
after the call returns:

```ant
let finished = Atomic.new(false)
Parallel.run([
  {
    go {
      Time.sleep(50ms)
      finished.store(true)
    }
  },
])
println(finished.load())
```

```text
false
```

### A lambda for `map` or `each`

Their callbacks take the element and the context, so a lambda does not fit:

```ant
let doubled: Array[int] | null = Parallel.map([1, 2], { x * 2 }, limit: 2)
// error: a lambda has at most one parameter, x; func(int, Context): int expects 2: write func with named parameters
```

### A limit below 1, an empty race

`map` and `each` panic on a `limit` below 1, and `race` panics on an empty list, since it has no
value to return:

```ant
Parallel.each([1], func(n: int, _: Context) { n }, limit: 0) // panic: Parallel.each needs a limit of at least 1, got 0
```

```ant
let empty: Array[func(Context): int] = []
let winner = Parallel.race(empty) // panic: Parallel.race needs at least one function
```

## See also

- [Context](../Context/index.md): deadlines and cancellation for the scope
- [WaitGroup](../WaitGroup/index.md): wait for goroutines started with `go`
- [Channel](../Channel/index.md): feed `Parallel.each` and hand values between goroutines
- [Atomic](../Atomic/index.md) and [Mutex](../Mutex/index.md): share state between callbacks
- [Panic.catch](../Panic/catch.md): stop a panic a callback raised
- [Concurrency](../../guide/concurrency.md): how goroutines, channels, and sync objects fit together
