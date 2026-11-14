<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# WaitGroup.new

```ant signature
func WaitGroup.new(): WaitGroup
```

WaitGroup.new returns a new wait group with its counter at zero.

<!-- end generated -->

## Details

- Each call makes a distinct wait group, compared by identity. Passing it to a function, storing it
  in a record field, or capturing it in a closure shares that one wait group.
- The counter starts at zero, so `wait` on a new wait group returns at once.

## Basic examples

Make a wait group:

```ant
WaitGroup.new() // WaitGroup
```

Start three goroutines and wait for all of them. Each one writes its own slot, and `defer` marks it
done even if it panics:

```ant
let wg = WaitGroup.new()
let squares = Array.filled(3, 0)
for i in Array.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    squares[i] = i * i
  }
}
wg.wait()
println(squares)
```

```text
[0, 1, 4]
```

## Scope

A wait group is a value like any other: keep it in a `let`, an array, or a map. It renders as its
type's name:

```ant
let wg = WaitGroup.new()
[wg, WaitGroup.new()]  // [WaitGroup, WaitGroup]
Map{ "downloads": wg } // Map{ "downloads": WaitGroup }
```

A record field can default to `WaitGroup.new()`. The default runs afresh for each record, so each
record gets its own wait group:

```ant
type Batch = { pending: WaitGroup = WaitGroup.new(), name: string }
let a = Batch{ name: "a" }
let b = Batch{ name: "b" }
a.pending == b.pending // false
```

Add the whole count at once when it is known before the loop:

```ant
let jobs = ["parse", "check", "emit"]
let wg = WaitGroup.new()
let finished = Atomic.new(0)
wg.add(jobs.length())
for job in jobs {
  go {
    defer { wg.done() }
    finished.add(1)
  }
}
wg.wait()
println(finished.load())
```

```text
3
```

## Applications

Run a fixed pool of workers over a channel of jobs, and close the results only after every worker
has finished:

```ant
let jobs: Channel[int] = Channel.new(capacity: 10)
let results: Channel[int] = Channel.new(capacity: 10)
let workers = WaitGroup.new()
for _ in Array.range(0, 3) {
  workers.add(1)
  go {
    defer { workers.done() }
    for n in jobs.values() {
      results.send(n * 10)
    }
  }
}
for n in Array.range(1, 6) {
  jobs.send(n)
}
jobs.close()
workers.wait()
results.close()
let total = results.values() | Iter.toArray() | Array.sum()
println(total)
```

```text
150
```

Bound a wait with a context, so that a goroutine that never finishes does not block forever. After
the wait, `ctx.isDone()` is false only when the counter reached zero before the context ended:

```ant
let wg = WaitGroup.new()
wg.add(1)
go {
  Time.sleep(1h)
  wg.done()
}
let ctx = Context.new(timeout: 50ms)
wg.wait(ctx)
println(if ctx.isDone() { "timed out" } else { "finished" })
```

```text
timed out
```

## Properties and relations

Two calls make two wait groups. A wait group equals only itself:

```ant
let a = WaitGroup.new()
let b = WaitGroup.new()
assert(a != b)
assert(a == a)
```

A new wait group's counter is zero, so `wait` returns at once. `done` takes one from the counter, as
`add(-1)` does here: adding and taking away the same amount brings the counter back to zero, where
`done` panics:

```ant
let wg = WaitGroup.new()
wg.wait()
wg.add(2)
wg.done()
wg.add(-1)
wg.wait()
assert(Panic.catch({ wg.done() }) is Err)
```

## Possible issues

The counter of a new wait group is zero, so `done` panics:

```ant
WaitGroup.new().done() // panic: WaitGroup.done: the counter is already zero
```

Call `add` before the `go` that starts the work, not inside it. With `wg.add(1)` as the first line
inside `go { ... }`, `wait` can find the counter still at zero and return before the goroutine has
run, so the work may not be done when the program reads its result. Add first, then start the
goroutine:

```ant
let wg = WaitGroup.new()
let ran = Atomic.new(false)
wg.add(1)
go {
  defer { wg.done() }
  ran.store(true)
}
wg.wait()
println(ran.load())
```

```text
true
```

A goroutine that panics before its `done` never decrements the counter, and `wait` blocks for good.
Put `done` in a `defer` at the top of the goroutine, so it runs on a panic too.

## See also

[WaitGroup.add](add.md) · [WaitGroup.done](done.md) · [WaitGroup.wait](wait.md) ·
[WaitGroup](index.md) · [Parallel.each](../Parallel/each.md) · [Mutex.new](../Mutex/new.md) ·
[Once.new](../Once/new.md) · [Channel.new](../Channel/new.md)
