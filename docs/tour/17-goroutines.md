<!-- status: reviewed -->

# Goroutines

`go` runs a block on a new goroutine, next to the code that started it. `Parallel` runs a batch of
work on goroutines and waits for all of it.

```ant
let wg = WaitGroup.new()
let lock = Mutex.new()
var total = 0
for n in [1, 2, 3, 4] {
  wg.add(1)
  go {
    defer { wg.done() }
    let square = n * n
    lock.withLock({ total += square })
  }
}
wg.wait()
println("total: ${total}")

func slowSquare(n: int, work: Context): int {
  Time.sleep(Time.duration(milliseconds: 10 * (5 - n)), work)
  n * n
}

let squares = Parallel.map([1, 2, 3, 4], slowSquare, limit: 2) ?? []
println(squares)

type Summary = { words: int, longest: string }
let words = "go runs a block on a new goroutine".split(" ")
let summary = Summary{ words: 0, longest: "" }
Parallel.run([
  { summary.words = words.length() },
  { summary.longest = words.maxBy({ x.length() }) ?? "" },
])
println(summary)
```

```text
total: 30
[1, 4, 9, 16]
Summary{ words: 8, longest: "goroutine" }
```

Each `go { ... }` starts a goroutine and returns at once, so the loop starts four of them without
waiting. A goroutine hands back no value and no handle. Two things make up for it here: the
`WaitGroup` counts the goroutines still running, and the `Mutex` lets only one of them at a time add
to `total`.

`wg.add(1)` counts a goroutine before it starts, and `wg.done()` counts it off when it ends. Putting
`done` in a `defer` counts it off even if the goroutine panics. A panic in a `go` goroutine ends
only that goroutine. The runtime reports it on standard error and the rest of the program keeps
running, which is why `done` belongs in a `defer`. `wg.wait()` blocks until the count is zero, so
`total` is read only after every goroutine has added its square. Without the lock, two goroutines
could update `total` at the same time. That is a data race, and the total it leaves is not certain.

`Parallel.map` does the same job in one call. It runs `slowSquare` on every element, at most
`limit: 2` at a time, waits for all of them, and returns the results in input order. The later
elements sleep less, so `2` finishes before `1`, and the results still line up with the input. Each
callback gets a context, `work`, that it passes to its waits, so the call can stop it early.
`Parallel.map` returns `null` only when the run is cancelled before it starts every element, which
`?? []` turns into an empty array.

`Parallel.run` runs a list of different tasks at once. Each one writes its own field of `summary`,
and after the call you read the fields with no lock, because `Parallel.run` returns only once every
task has ended. If a task panics, the panic comes out of the `Parallel` call, in the code that made
it.

The program exits when the script ends, whether or not other goroutines are still running. Try
deleting `wg.wait()`: nothing then makes the goroutines finish before the read, so reading `total`
is a data race and the printed number is not certain. It is often 0, because the script reaches
`println` before the goroutines run.

The [concurrency chapter](../guide/concurrency.md) explains how goroutines, locks, and `Parallel`
fit together. The reference has [`go`](../language/keywords/go.md),
[WaitGroup](../lib/WaitGroup/index.md), [Mutex.withLock](../lib/Mutex/withLock.md),
[Parallel.map](../lib/Parallel/map.md), and [Parallel.run](../lib/Parallel/run.md).

Next: [channels](18-channels.md).
