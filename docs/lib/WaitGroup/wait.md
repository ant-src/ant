<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# WaitGroup.wait

```ant signature
func WaitGroup.wait(wg: WaitGroup, ctx: Context | null = null): void
```

WaitGroup.wait blocks until the counter of wg is zero, or until ctx is done. When ctx ends the wait
it returns early; [`Context.isDone`](../Context/isDone.md) tells the two apart.

<!-- end generated -->

## Details

- Every write a goroutine makes before its `WaitGroup.done` is visible after the `wait` that it
  unblocks returns.
- Any number of goroutines can wait on the same group; each `wait` returns once the counter is zero.
- A wait that `ctx` ends does not stop the group's goroutines; they keep running.
- `Context.isDone` is `true` once the context is done, which can also happen after the counter
  reaches zero. `true` means the context ended, not that work is still running. Read the results the
  goroutines wrote when you need to know what finished.
- `WaitGroup.wait(wg)` is a head of `select`, where it takes no context.

## Basic examples

Start three goroutines and wait for all of them. Each one writes its own slot, so the results are
ready to read after the wait:

```ant
let words = ["ant", "bee", "cat"]
let lengths = Array.filled(words.length(), 0)
let wg = WaitGroup.new()
for entry in words.entries() {
  wg.add(1)
  go {
    defer { wg.done() }
    lengths[entry.index] = entry.value.length()
  }
}
WaitGroup.wait(wg)
println(lengths)
```

```text
[3, 3, 3]
```

Bound the wait with a context. The worker takes a second, the context ends after 20 milliseconds,
and `Context.isDone` shows that the wait gave up:

```ant
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  Time.sleep(1s)
}
let ctx = Context.new(timeout: 20ms)
WaitGroup.wait(wg, ctx: ctx)
println(if ctx.isDone() { "gave up: ${ctx.reason()}" } else { "all done" })
```

```text
gave up: timeout
```

## Scope

A group whose counter is zero has nothing to wait for, so `wait` returns at once:

```ant
let wg = WaitGroup.new()
wg.wait()
println("nothing to wait for")
```

```text
nothing to wait for
```

Several goroutines can wait on one group. Here the group works as a starting gate: three runners
wait on it, and one `done` releases all of them:

```ant
let gate = WaitGroup.new()
gate.add(1)
let started = Atomic.new(0)
let runners = WaitGroup.new()
for _ in Array.range(0, 3) {
  runners.add(1)
  go {
    defer { runners.done() }
    gate.wait()
    started.add(1)
  }
}
println("started before the gate opens: ${started.load()}")
gate.done()
runners.wait()
println("started after: ${started.load()}")
```

```text
started before the gate opens: 0
started after: 3
```

A group in a record field waits for the record's own goroutines:

```ant
type Batch = { results: Array[int], pending: WaitGroup = WaitGroup.new() }

func Batch.start(b: Batch, n: int) {
  let slot = b.results.length()
  b.results.push(0)
  b.pending.add(1)
  go {
    defer { b.pending.done() }
    b.results[slot] = n * n
  }
}

let batch = Batch{ results: [] }
for n in [2, 3, 4] {
  batch.start(n)
}
batch.pending.wait()
println(batch.results)
```

```text
[4, 9, 16]
```

In a `select`, `WaitGroup.wait(wg)` is a head that can proceed when the counter is zero. Beside a
`Time.sleep` head it waits for the group or for the time, whichever comes first:

```ant
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  Time.sleep(1s)
}
select {
  WaitGroup.wait(wg) { println("finished") }
  Time.sleep(20ms) { println("still running after 20ms") }
}
```

```text
still running after 20ms
```

## Named arguments

### ctx

`ctx:` gives up the wait when the context is done: by its timeout, or by a `Context.cancel`
anywhere. A context that is cancelled already ends the wait at once:

```ant
let wg = WaitGroup.new()
wg.add(1)
let ctx = Context.new()
ctx.cancel()
wg.wait(ctx: ctx)
println(ctx.reason())
```

```text
cancelled
```

A context with a deadline that the group meets does not change the result. `Context.isDone` is
`false` after the wait, since the counter reached zero first:

```ant
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  Time.sleep(10ms)
}
let ctx = Context.new(timeout: 5s)
wg.wait(ctx: ctx)
println(ctx.isDone())
```

```text
false
```

## Applications

Shut down workers with a grace period. The workers loop until their context is cancelled; after
cancelling it, the main goroutine waits for them, but no longer than the grace period:

```ant
let work = Context.new()
let stopped = Atomic.new(0)
let wg = WaitGroup.new()
for _ in Array.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    while !work.isDone() {
      Time.sleep(5ms, ctx: work)
    }
    stopped.add(1)
  }
}
Time.sleep(20ms)
work.cancel()
let grace = Context.new(timeout: 1s)
wg.wait(ctx: grace)
println(if grace.isDone() { "some workers did not stop" } else { "${stopped.load()} workers stopped" })
```

```text
3 workers stopped
```

Fetch several pages at once within one deadline. Each fetch gives up when the shared context ends,
so a plain `wait` returns by the deadline, and every goroutine has written its slot by then; `null`
marks a page that did not finish:

```ant
func fetch(page: int, ctx: Context): string | null {
  let delay = if page == 3 { 1s } else { 10ms }
  Time.sleep(delay, ctx: ctx)
  if ctx.isDone() { return null }
  "page ${page}"
}

let pages = [1, 2, 3]
let bodies: Array[string | null] = Array.filled(pages.length(), null)
let ctx = Context.new(timeout: 100ms)
let wg = WaitGroup.new()
for entry in pages.entries() {
  wg.add(1)
  go {
    defer { wg.done() }
    bodies[entry.index] = fetch(entry.value, ctx)
  }
}
wg.wait()
println(bodies)
```

```text
["page 1", "page 2", null]
```

## Properties and relations

Everything the goroutines wrote before `done` is visible once `wait` returns, so the sum of the
slots is complete:

```ant
let n = 100
let squares = Array.filled(n, 0)
let wg = WaitGroup.new()
for i in Array.range(0, n) {
  wg.add(1)
  go {
    defer { wg.done() }
    squares[i] = i * i
  }
}
wg.wait()
assert(squares.sum() == 328350)
assert(squares[n - 1] == (n - 1) * (n - 1))
```

A goroutine per task joined by a `WaitGroup` does what `Parallel.run` does for tasks that cannot
panic: both return once every task has finished:

```ant
let a = Array.filled(2, "")
let wg = WaitGroup.new()
wg.add(2)
go {
  defer { wg.done() }
  a[0] = "left"
}
go {
  defer { wg.done() }
  a[1] = "right"
}
wg.wait()

let b = Array.filled(2, "")
Parallel.run([
  func(work: Context) { b[0] = "left" },
  func(work: Context) { b[1] = "right" },
])
assert(a == b)
```

## Possible issues

Call `WaitGroup.add` before `go`, not inside the goroutine. Otherwise `wait` can run before the
goroutine has added anything, see a zero counter, and return while the work is still going. This
program may print `0`:

<!-- test: check-only whether the goroutine adds before the wait varies between runs -->

```ant
let wg = WaitGroup.new()
let total = Atomic.new(0)
go {
  wg.add(1) // too late: the wait may have returned already
  defer { wg.done() }
  total.add(1)
}
wg.wait()
println(total.load())
```

Every path of a goroutine must call `WaitGroup.done`. Here the worker returns early for an empty
name and skips its `done`, so the wait would never return; the context makes it give up instead.
Write `defer { wg.done() }` as the goroutine's first line so that every `return` calls it:

```ant
func greet(name: string, wg: WaitGroup) {
  if name.isEmpty() { return }
  println("hello, ${name}")
  wg.done()
}

let wg = WaitGroup.new()
wg.add(2)
go { greet("Ann", wg) }
go { greet("", wg) }
let ctx = Context.new(timeout: 50ms)
wg.wait(ctx: ctx)
println(ctx.reason())
```

```text
hello, Ann
timeout
```

A panic ends its goroutine but not the program. Without a deferred `done`, the counter never reaches
zero and `wait` blocks for good. With one, the deferred block runs as the panic leaves the
goroutine, and `wait` returns:

```ant
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  panic("worker failed")
}
wg.wait()
println("the wait returned")
```

```text
the wait returned
```

A wait that the context ends does not stop the goroutines; they go on running. Pass the same context
to their own waits so that they stop too:

```ant
let ctx = Context.new(timeout: 20ms)
let finished = Atomic.new(0)
let wg = WaitGroup.new()
for _ in Array.range(0, 2) {
  wg.add(1)
  go {
    defer { wg.done() }
    Time.sleep(1s, ctx: ctx)
    if !ctx.isDone() { finished.add(1) }
  }
}
wg.wait(ctx: ctx)
println("finished: ${finished.load()}")
```

```text
finished: 0
```

## See also

[WaitGroup.add](add.md) · [WaitGroup.done](done.md) · [WaitGroup.new](new.md) ·
[WaitGroup](index.md) · [Context.isDone](../Context/isDone.md) · [Context.new](../Context/new.md) ·
[Parallel.run](../Parallel/run.md) · [Parallel.map](../Parallel/map.md) ·
[select](../../language/keywords/select.md) · [go](../../language/keywords/go.md) ·
[Concurrency guide](../../guide/concurrency.md)
