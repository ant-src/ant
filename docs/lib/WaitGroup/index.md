<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# WaitGroup

```ant signature
type WaitGroup
```

WaitGroup is a counter that lets a goroutine wait until a group of others has finished. WaitGroups
compare by identity.

<!-- end generated -->

A wait group counts work that is still running. You add one to its counter for each goroutine you
start, each goroutine calls `done` when it finishes, and `wait` blocks until the counter is back at
zero. A `go` block hands back no result and cannot be joined, so a wait group is how a goroutine
learns that others have finished. The `WaitGroup` namespace holds the functions that make a group,
change its counter, and wait on it; `WaitGroup` is also the type, for an annotation, a parameter, or
a record field.

## Details

- `wait` returns at once when the counter is zero. Call `add` before the `go` that starts the work.
- `done` panics when the counter is already zero, and `add` panics when the counter would go
  negative.
- Everything a goroutine writes before its `done` is visible after the `wait` it unblocks returns.
  Hand results back through an array slot, a map entry, a record field, or an `Atomic`.
- A panic ends its goroutine but not the program. Write `defer { wg.done() }` first in the
  goroutine, so that `done` runs on every path, a panic included.
- Ending a `wait` through its context does not stop the goroutines; check `Context.isDone` to tell a
  timeout from a finished group.

## Basic examples

`WaitGroup.new` makes a group. `WaitGroup.add` counts each goroutine before it starts, each one
calls `WaitGroup.done` as it ends, and `WaitGroup.wait` returns once all have finished. Each
goroutine writes its own slot, so the results are complete after the wait:

```ant
let words = ["ant", "bee", "wasp"]
let lengths = Array.filled(words.length(), 0)
let wg = WaitGroup.new()
for entry in words.entries() {
  WaitGroup.add(wg, 1)
  go {
    defer { WaitGroup.done(wg) }
    lengths[entry.index] = entry.value.length()
  }
}
WaitGroup.wait(wg)
println(lengths)
```

```text
[3, 3, 4]
```

A context bounds the wait. Here the worker takes a second, the wait gives up after 20 milliseconds,
and `Context.isDone` tells that apart from a finished group:

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

A wait group renders as its type's name:

```ant
WaitGroup.new() // WaitGroup
```

A wait group is a reference value. A function that takes one as a parameter changes the caller's
group, so a helper can count the goroutines it starts:

```ant
func startSquare(wg: WaitGroup, out: Array[int], i: int) {
  wg.add(1)
  go {
    defer { wg.done() }
    out[i] = i * i
  }
}

let squares = Array.filled(4, 0)
let wg = WaitGroup.new()
for i in Array.range(0, 4) {
  startSquare(wg, squares, i)
}
wg.wait()
println(squares)
```

```text
[0, 1, 4, 9]
```

A group in a record field counts the work of that record, here the uploads a client has in flight:

```ant
type Uploader = { sent: Atomic[int] = Atomic.new(0), inFlight: WaitGroup = WaitGroup.new() }

func Uploader.upload(u: Uploader) {
  u.inFlight.add(1)
  go {
    defer { u.inFlight.done() }
    Time.sleep(5ms)
    u.sent.add(1)
  }
}

let uploader = Uploader{}
for _ in Array.range(0, 3) {
  uploader.upload()
}
uploader.inFlight.wait()
println("${uploader.sent.load()} uploads sent")
```

```text
3 uploads sent
```

When the number of goroutines is known up front, one `add` can count them all before the loop:

```ant
let jobs = ["build", "test", "lint"]
let letters = Atomic.new(0)
let wg = WaitGroup.new()
wg.add(jobs.length())
for job in jobs {
  go {
    defer { wg.done() }
    letters.add(job.length())
  }
}
wg.wait()
println("${jobs.length()} jobs, ${letters.load()} letters")
```

```text
3 jobs, 13 letters
```

In a `select`, `WaitGroup.wait(wg)` is a head that proceeds once the counter is zero. Beside a
`Context.wait` head, it waits for the group or for the deadline, whichever comes first:

```ant
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  Time.sleep(10ms)
}
let ctx = Context.new(timeout: 1s)
select {
  WaitGroup.wait(wg) { println("finished") }
  Context.wait(ctx) { println("deadline passed") }
}
```

```text
finished
```

## Functions by job

### Making a wait group

- [WaitGroup.new](new.md): a new group

### Counting work

- [WaitGroup.add](add.md): add `n`, which may be negative, to the counter
- [WaitGroup.done](done.md): take one off the counter, as a goroutine finishes

### Waiting

- [WaitGroup.wait](wait.md): block until the counter is zero, or until `ctx` is done

## Applications

### Close a results channel when the workers finish

Workers send results on one channel. A separate goroutine waits for the workers and then closes the
channel, so the `for` loop over its values ends after the last result:

```ant
let jobs = [3, 1, 4, 1, 5]
let results: Channel[int] = Channel.new()
let wg = WaitGroup.new()
for job in jobs {
  wg.add(1)
  go {
    defer { wg.done() }
    results.send(job * 10)
  }
}
go {
  wg.wait()
  results.close()
}
var total = 0
for r in results.values() {
  total += r
}
println(total)
```

```text
140
```

### Collect into a map under a mutex

Goroutines that write one shared map take a mutex around each write. The wait group says when the
map is complete:

```ant
let words = ["ant", "bee", "ant", "wasp", "bee", "ant"]
let mu = Mutex.new()
let counts: Map[string, int] = Map{}
let wg = WaitGroup.new()
for word in words {
  wg.add(1)
  go {
    defer { wg.done() }
    mu.withLock({ counts[word] = (counts.get(word) ?? 0) + 1 })
  }
}
wg.wait()
println(counts.get("ant"))
```

```text
3
```

## Properties and relations

A wait group equals itself and its copies, and no other group:

```ant
let wg = WaitGroup.new()
let alias = wg
assert(wg == alias)
assert(wg != WaitGroup.new())
```

A goroutine per element writing its own slot, joined by a wait group, gives the same array as
`Parallel.map` with a limit that lets every element run at once:

```ant
let xs = [1, 2, 3, 4]
let byHand = Array.filled(xs.length(), 0)
let wg = WaitGroup.new()
for entry in xs.entries() {
  wg.add(1)
  go {
    defer { wg.done() }
    byHand[entry.index] = entry.value * entry.value
  }
}
wg.wait()
assert(byHand == Parallel.map(xs, func(x: int, work: Context): int { x * x }, limit: 4))
```

## Possible issues

### One done too many

`done` on a zero counter panics, as does an `add` that takes the counter below zero:

```ant
let wg = WaitGroup.new()
wg.add(1)
wg.done()
wg.done() // panic: WaitGroup.done: WaitGroup.done when the counter is already zero
```

### Adding inside the goroutine

An `add` inside the goroutine comes too late: `wait` can run first, see a zero counter, and return
while the work is still going. This program may print `0`:

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

### A panic skips a done that is not deferred

A goroutine that panics before a plain `done` never calls it, so the counter stays above zero and a
`wait` without a context blocks until the run is cancelled. Here a context bounds the wait to show
it; `defer` the `done` instead:

```ant
let wg = WaitGroup.new()
wg.add(1)
go {
  panic("worker failed")
  wg.done()
}
let ctx = Context.new(timeout: 50ms)
wg.wait(ctx: ctx)
println(ctx.reason())
```

```text
timeout
```

### Reassigning a local of the parent

A local assigned inside a `go` block does not narrow after the wait, even though the wait group
makes the read safe. `ant check` rejects the use:

```ant
let wg = WaitGroup.new()
var best: int | null = null
wg.add(1)
go {
  defer { wg.done() }
  best = 42
}
wg.wait()
println(best + 1) // error: + needs both operands of one type, got int | null and int
```

Write into a container instead, and read the value into a fresh local after the wait:

```ant
let wg = WaitGroup.new()
let best: Atomic[int | null] = Atomic.new(null)
wg.add(1)
go {
  defer { wg.done() }
  best.store(42)
}
wg.wait()
let got = best.load()
if got != null { println(got + 1) }
```

```text
43
```

## See also

- [Parallel](../Parallel/index.md): run a known set of tasks and wait for all of them
- [Mutex](../Mutex/index.md): guard data the goroutines share
- [Atomic](../Atomic/index.md): one value read and written atomically
- [Channel](../Channel/index.md): hand values between goroutines
- [Context](../Context/index.md): bound a wait with a deadline
- [Once](../Once/index.md): run work exactly once
- [go](../../language/keywords/go.md) · [defer](../../language/keywords/defer.md) ·
  [select](../../language/keywords/select.md)
- [Concurrency](../../guide/concurrency.md): how goroutines, channels, and sync objects fit together
