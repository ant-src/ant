<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Parallel.each

```ant signature
func Parallel.each[T, R](xs: Array[T] | Channel[T], f: func(T, Context): R, limit: int, ctx: Context | null = null): void
```

Parallel.each calls f on every element of xs, or on every value received from xs until it is closed,
with the scope context, running at most limit callbacks at once, and discards f's value. Over a
channel it receives only when a slot is free and under the scope context, so a value it receives
always runs; it neither closes nor drains the channel, and a null value is passed to f like any
other. It reports nothing: a done ctx after the call does not say which elements ran. It panics if
limit \< 1, even when xs is empty.

<!-- end generated -->

## Details

- `Parallel.each` returns only after every callback it started has ended, deferred blocks included.
  A field or slot a callback wrote can be read after the call with no lock.
- Callbacks run on goroutines of their own, in an unspecified order. Callbacks that touch the same
  variable, field, or element need synchronization, such as a [Mutex](../Mutex/index.md).
- Elements are admitted in index order, or channel values in receive order. With `limit: 1`, each
  callback ends before the next one starts.
- Over a channel, the call ends when the channel is closed and empty, or when `ctx` is done; values
  still in the channel stay there.
- A `go` inside a callback starts a goroutine the call does not wait for, which may outlive it; use
  a nested `Parallel` call instead.
- A panic in a callback cancels the scope context, waits for the other callbacks, and is raised in
  the caller with its original message.

## Basic examples

Call a function on every element, one at a time:

```ant
Parallel.each(["one", "two", "three"], func(word: string, work: Context) { println(word) }, limit: 1)
```

```text
one
two
three
```

Each callback fills in its own record, so the results can be read after the call with no lock:

```ant
type Job = { name: string, size: int | null }

let jobs = [Job{ name: "ant", size: null }, Job{ name: "beetle", size: null }]
Parallel.each(jobs, func(job: Job, work: Context) { job.size = job.name.length() }, limit: 2)
println(jobs.map({ x.size }))
```

```text
[3, 6]
```

Take work from a channel until the producer closes it:

```ant
let words: Channel[string] = Channel.new()
go {
  defer { words.close() }
  for w in ["red", "green", "blue"] {
    words.send(w)
  }
}
Parallel.each(words, func(w: string, work: Context) { println(w.toUpper()) }, limit: 1)
```

```text
RED
GREEN
BLUE
```

## Scope

An empty array returns at once, without calling `f`:

```ant
Parallel.each([], func(n: int, work: Context) { println(n) }, limit: 4)
println("done")
```

```text
done
```

A closed channel still hands over the values it holds:

```ant
let queue: Channel[int] = Channel.new(capacity: 3)
for n in [10, 20, 30] {
  queue.send(n)
}
queue.close()
Parallel.each(queue, func(n: int, work: Context) { println(n / 10) }, limit: 1)
```

```text
1
2
3
```

A `null` sent on a `Channel[T | null]` is a value like any other, not the end of the channel:

```ant
let readings: Channel[int | null] = Channel.new(capacity: 3)
readings.send(20)
readings.send(null)
readings.send(22)
readings.close()
Parallel.each(readings, func(r: int | null, work: Context) { println(r ?? "missing") }, limit: 1)
```

```text
20
missing
22
```

`f` may be a named function, and its value is discarded. Here each check returns a `Result` that
nothing reads; the check records its outcome in the record instead:

```ant
type Site = { host: string, up: bool | null }

func probe(site: Site, work: Context): Result[void, Error] {
  let up = !site.host.startsWith("old.")
  site.up = up
  if !up { return Result.fail("${site.host} is down") }
  Result.ok(void)
}

let sites = [Site{ host: "api.example", up: null }, Site{ host: "old.example", up: null }]
Parallel.each(sites, probe, limit: 2)
for s in sites {
  println("${s.host}: ${s.up}")
}
```

```text
api.example: true
old.example: false
```

The input comes first, so `Parallel.each` ends a pipe:

```ant
Iter.range(1, 4)
  | Iter.map({ x * x })
  | Iter.toArray()
  | Parallel.each(func(n: int, work: Context) { println(n) }, limit: 1)
```

```text
1
4
9
```

## Named arguments

### limit

`limit:` caps how many callbacks run at once. Twenty jobs with `limit: 4` never run more than four
at a time:

```ant
let m = Mutex.new()
var running = 0
var peak = 0
let jobs = Iter.range(0, 20) | Iter.toArray()
Parallel.each(jobs, func(n: int, work: Context) {
  m.withLock({
    running += 1
    peak = Math.max(peak, running)
  })
  Time.sleep(5ms)
  m.withLock({ running -= 1 })
}, limit: 4)
println(peak <= 4)
```

```text
true
```

To start a goroutine for every element, pass the length, which keeps the choice visible:

```ant
let urls = ["a.example", "b.example", "c.example"]
let seen: Channel[string] = Channel.new(capacity: urls.length())
Parallel.each(urls, func(url: string, work: Context) { seen.send(url) }, limit: urls.length())
println(seen.length())
```

```text
3
```

For CPU-bound work, pass [`Os.cpus()`](../Os/cpus.md), the CPUs the program may use at once; work
that waits on the network or disk takes a limit chosen for the other side. Each callback writes its
own slot:

```ant
let sums = Array.filled(3, 0)
Parallel.each([0, 1, 2], func(i: int, work: Context) {
  sums[i] = Array.range(1, (i + 1) * 100 + 1).sum() ?? 0
}, limit: Os.cpus())
println(sums)
```

```text
[5050, 20100, 45150]
```

### ctx

The scope context is a child of `ctx:`. A `ctx` that is done already admits nothing:

```ant
let ctx = Context.new()
ctx.cancel()
var ran = 0
Parallel.each([1, 2, 3], func(n: int, work: Context) { ran += 1 }, limit: 3, ctx: ctx)
println(ran)
```

```text
0
```

`ctx:` also bounds the wait for a channel that is never closed. Here nobody sends or closes, and the
timeout ends the call:

```ant
let lines: Channel[string] = Channel.new()
let ctx = Context.new(30ms)
Parallel.each(lines, func(s: string, work: Context) { println(s) }, limit: 2, ctx: ctx)
println(ctx.reason())
```

```text
timeout
```

## Applications

### A worker pool

A producer feeds a channel, and `Parallel.each` runs three workers over it. The workers add to a
shared map, so they take a lock:

```ant
type Order = { item: string, qty: int }

let orders: Channel[Order] = Channel.new()
go {
  defer { orders.close() }
  for order in [
    Order{ item: "tea", qty: 2 },
    Order{ item: "milk", qty: 1 },
    Order{ item: "tea", qty: 3 },
    Order{ item: "bread", qty: 1 },
  ] {
    orders.send(order)
  }
}

let m = Mutex.new()
let totals: Map[string, int] = Map{}
Parallel.each(orders, func(order: Order, work: Context) {
  m.withLock({ totals[order.item] = (totals[order.item] ?? 0) + order.qty })
}, limit: 3)
for item in totals.keys() | Array.sort() {
  println("${item}: ${totals[item]}")
}
```

```text
bread: 1
milk: 1
tea: 5
```

### Stop at the first failure

`Parallel.each` ignores the value `f` returns, so an `Err` stops nothing by itself. To stop at the
first failure, cancel a context you own and pass it as `ctx:`. With `limit: 1`, the steps after the
failing one are never admitted:

```ant
type Step = { name: string, ok: bool }

let steps = [
  Step{ name: "fetch", ok: true },
  Step{ name: "build", ok: false },
  Step{ name: "deploy", ok: true },
]
let stop = Context.new()
Parallel.each(steps, func(step: Step, work: Context) {
  println("run ${step.name}")
  if !step.ok { stop.cancel() }
}, limit: 1, ctx: stop)
println(stop.reason())
```

```text
run fetch
run build
cancelled
```

### A panic stops the siblings

A panic in one callback cancels the scope context. A sibling that waits on its `work` context wakes
up, and the call raises the panic in the caller:

```ant
let r = Panic.catch({
  Parallel.each([1, 2], func(n: int, work: Context) {
    if n == 2 { panic("job 2 failed") }
    Time.sleep(1h, ctx: work)
  }, limit: 2)
})
if r is Err {
  println(r.error.message)
}
```

```text
job 2 failed
```

## Properties and relations

With `limit: 1`, `Parallel.each` calls `f` in input order, as a `for` loop does:

```ant
let seen: Array[int] = []
Parallel.each([3, 1, 2], func(n: int, work: Context) { seen.push(n) }, limit: 1)
assert(seen == [3, 1, 2])
```

Every callback has ended when the call returns, so no `WaitGroup` is needed to read what they wrote:

```ant
type Cell = { n: int, square: int }

let cells = [1, 2, 3, 4].map({ Cell{ n: x, square: 0 } })
Parallel.each(cells, func(c: Cell, work: Context) { c.square = c.n * c.n }, limit: 4)
assert(cells.map({ x.square }) == [1, 4, 9, 16])
```

When the results matter, [Parallel.map](map.md) returns them in input order:

```ant
let squares = Parallel.map([1, 2, 3], func(n: int, work: Context) { n * n }, limit: 3)
assert(squares == [1, 4, 9])
```

## Possible issues

### limit is required

`limit:` has no default:

```ant
Parallel.each([1, 2], func(n: int, work: Context) { println(n) })
// error: Parallel.each needs an argument for limit
```

### A limit below 1 panics, even for empty input

```ant
let r = Panic.catch({ Parallel.each([], func(n: int, work: Context) {}, limit: 0) })
if r is Err {
  println(r.error.message)
}
```

```text
Parallel.each needs a limit of at least 1, got 0
```

### Wait on the scope context, not the caller's

A callback that waits on the caller's context, or on none, does not see a sibling's panic. In the
panic example above, a callback that slept with `Time.sleep(1h)`, without `ctx: work`, would hold
the call for the whole hour. Pass `work`, the context the callback receives.

### Nested calls multiply the limit

The limit is per call. Two groups with `limit: 3`, each running its items with `limit: 3`, can run
up to six callbacks at once:

```ant
let m = Mutex.new()
var running = 0
var peak = 0
let groups = [[1, 2, 3], [4, 5, 6]]
Parallel.each(groups, func(group: Array[int], outer: Context) {
  Parallel.each(group, func(n: int, work: Context) {
    m.withLock({
      running += 1
      peak = Math.max(peak, running)
    })
    Time.sleep(5ms)
    m.withLock({ running -= 1 })
  }, limit: 3, ctx: outer)
}, limit: 3)
println(peak <= 6)
```

```text
true
```

### Unsynchronized writes to one variable

Callbacks run at the same time, so `total += n` from several of them races. Take a lock, or give
each element a slot of its own:

```ant
let m = Mutex.new()
var total = 0
Parallel.each([1, 2, 3, 4], func(n: int, work: Context) { m.withLock({ total += n }) }, limit: 4)
println(total)
```

```text
10
```

### A producer blocked in send stays blocked

When `ctx` ends the call, `Parallel.each` stops receiving. A producer waiting in `Channel.send` is
not woken by that; give it the same context:

```ant
let jobs: Channel[int] = Channel.new()
let ctx = Context.new(30ms)
let producer: Channel[bool] = Channel.new(capacity: 1)
go {
  var n = 0
  while jobs.send(n, ctx: ctx) {
    n += 1
  }
  producer.send(true)
}
Parallel.each(jobs, func(n: int, work: Context) { Time.sleep(10ms, ctx: work) }, limit: 1, ctx: ctx)
println(producer.receive() ?? false)
```

```text
true
```

## See also

[Parallel.map](map.md), [Parallel.run](run.md), [Parallel.race](race.md), [Parallel.any](any.md),
[Channel.values](../Channel/values.md), [Context.new](../Context/new.md),
[Mutex.withLock](../Mutex/withLock.md), [WaitGroup](../WaitGroup/index.md),
[Panic.catch](../Panic/catch.md), [Os.cpus](../Os/cpus.md)
