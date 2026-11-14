<!-- status: reviewed -->

# `go`

`go { ... }` runs a block on a new goroutine and returns at once.

## Details

- `go` has type `void`. There is no handle to join or to read a result from: a goroutine hands back
  its result through a channel or a shared container, and signals completion through a `WaitGroup`.
- The program exits when the script body ends, whatever other goroutines are still doing.
- A panic ends only the goroutine it occurs in, after that goroutine's deferred blocks run. The
  message and trace go to standard error, and the rest of the program continues.
- A `go` block is a boundary: `return` in it ends the goroutine and takes no value, `try` directly
  in it is a compile error, and `break` and `continue` cannot reach a loop outside it.
- Everything the starting goroutine did before the `go` is visible to the new goroutine.

## Basic examples

Start a goroutine and wait for it with a `WaitGroup`:

```ant
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  println("hello from a goroutine")
}
wg.wait()
println("done")
```

```text
hello from a goroutine
done
```

Send a goroutine's result back on a channel:

```ant
let results: Channel[int] = Channel.new()
go { results.send(6 * 7) }
println(results.receive())
```

```text
42
```

## Scope

A `go` block captures the variables around it. Each loop iteration binds a fresh loop variable, so
every goroutine sees its own `i`. Each goroutine writes its own array slot:

```ant
let squares = Array.filled(4, 0)
let wg = WaitGroup.new()
for i in [0, 1, 2, 3] {
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
[0, 1, 4, 9]
```

A `go` block can call any function. Here the goroutine produces values that the script body reads:

```ant
func produce(out: Channel[int]) {
  defer { out.close() }
  for n in [1, 2, 3] {
    out.send(n)
  }
}

let numbers: Channel[int] = Channel.new()
go { produce(numbers) }
for n in numbers.values() {
  println(n * 10)
}
```

```text
10
20
30
```

A function can start a goroutine and return before it ends. `return` inside a `go` block ends the
goroutine, not the function that started it:

```ant
func notify(user: string, out: Channel[string]) {
  go {
    if user.isEmpty() {
      out.send("skipped")
      return
    }
    out.send("sent to ${user}")
  }
}

let out: Channel[string] = Channel.new()
notify("ann", out)
println(out.receive())
notify("", out)
println(out.receive())
```

```text
sent to ann
skipped
```

A panic ends its goroutine and nothing else. Its deferred blocks still run, so `defer { wg.done() }`
lets the wait return. The panic is reported on standard error, and the program goes on:

```ant
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  panic("worker failed")
}
wg.wait()
println("still running")
```

```text
still running
```

## Applications

A worker reads jobs from a channel and calls each through `Panic.catch`, so a bug in one job does
not end the worker:

```ant
func handle(job: string): int {
  100 / job.length()
}

let jobs: Channel[string] = Channel.new()
let report: Channel[string] = Channel.new()
go {
  defer { report.close() }
  for job in jobs.values() {
    let r = Panic.catch({ handle(job) })
    let line = if r is Ok { "${job}: ${r.value}" } else { "${job}: ${r.error.message}" }
    report.send(line)
  }
}
go {
  defer { jobs.close() }
  for job in ["abcd", "", "ab"] {
    jobs.send(job)
  }
}
for line in report.values() {
  println(line)
}
```

```text
abcd: 25
: division by zero
ab: 50
```

Give up on a slow goroutine with a context deadline. The receive returns `null` when the context is
done:

```ant
let reply: Channel[string] = Channel.new()
go {
  Time.sleep(5s)
  reply.send("late answer")
}
let ctx = Context.new(50ms)
println(reply.receive(ctx) ?? "timed out")
```

```text
timed out
```

Stop a goroutine with a context. The goroutine checks `isDone` between steps and passes the context
into its sleep, so `cancel` ends the wait at once. The caller cancels and then waits for the
goroutine to finish:

```ant
let ctx = Context.new()
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  while !ctx.isDone() {
    Time.sleep(1s, ctx)
  }
  println("worker stopped: ${ctx.reason() ?? "running"}")
}
Time.sleep(20ms)
ctx.cancel()
wg.wait()
println("done")
```

```text
worker stopped: cancelled
done
```

Count across goroutines with an `Atomic`:

```ant
let hits = Atomic.new(0)
let wg = WaitGroup.new()
for _ in Array.range(0, 100) {
  wg.add(1)
  go {
    defer { wg.done() }
    hits.add(1)
  }
}
wg.wait()
println(hits.load())
```

```text
100
```

## Properties and relations

A goroutine per element, each writing its own slot, joined with a `WaitGroup`, computes what
`Parallel.map` computes. `Parallel.map` also bounds how many run at once:

```ant
func squareAll(xs: Array[int]): Array[int] {
  let out = Array.filled(xs.length(), 0)
  let wg = WaitGroup.new()
  for e in xs.entries() {
    wg.add(1)
    go {
      defer { wg.done() }
      out[e.index] = e.value * e.value
    }
  }
  wg.wait()
  out
}

let xs = [3, 1, 4, 1, 5]
assert(squareAll(xs) == Parallel.map(xs, func(n: int, work: Context): int { n * n }, limit: 2))
```

A write the starting goroutine makes before the `go` is visible inside the goroutine:

```ant
let settings = Map{ mode: "fast" }
settings["mode"] = "safe"
let done: Channel[bool] = Channel.new()
go {
  assert(settings["mode"] == "safe")
  done.send(true)
}
done.receive()
```

## Possible issues

The program exits when the script body ends. A goroutine that has not finished by then never does:

```ant
go {
  Time.sleep(1s)
  println("never printed")
}
println("script ends")
```

```text
script ends
```

`go` gives `void`. The value of its block is dropped:

```ant
let v = go { 42 }
println(v)
```

```text
void
```

A local assigned inside a `go` block never narrows outside it, since the goroutine may assign it at
any moment. Send the value on a channel or write it into a container instead:

```ant
var hit: string | null = null
go { hit = "found" }
if hit != null { println(hit.length()) }
// error: string | null has no field or function length on every member: narrow the value first
```

A `go` block has no result, so `try` cannot return from it:

```ant
func load(): Result[int, Error] {
  Result.ok(1)
}

func start(): Result[int, Error] {
  go {
    let n = try load() // error: try needs an enclosing function that returns a Result
    println(n)
  }
  Result.ok(0)
}
```

Handle the `Result` inside the block, or send it to the code that waits.

`return` in a `go` block takes no value:

```ant
go { return 1 } // error: 'return' takes no value in a go block
```

`break` and `continue` cannot leave a loop outside the `go` block:

```ant
for _ in [1, 2] {
  go {
    break // error: 'break' outside a loop
  }
}
```

A goroutine that panics before it signals completion leaves a waiter blocked for good. Put
`defer { wg.done() }` first in the block, as the examples above do, so it runs on every exit.

Two goroutines that access the same variable, field, or element without synchronization, at least
one of them writing, are a data race. A data race is a program error: a read sees some stored value
but not a particular one, and the program may end with exit status 1. Guard shared state with an
`Atomic`, a `Mutex`, or a channel, as in the counter above.

A goroutine that ignores its context runs until the program ends. Check `isDone` between steps and
pass the context into waits, as the stopping example above does.

## See also

[defer](defer.md) · [select](select.md) · [return](return.md) ·
[Channel](../../lib/Channel/index.md) · [WaitGroup](../../lib/WaitGroup/index.md) ·
[Atomic](../../lib/Atomic/index.md) · [Mutex](../../lib/Mutex/index.md) ·
[Context](../../lib/Context/index.md) · [Parallel](../../lib/Parallel/index.md) ·
[Panic.catch](../../lib/Panic/catch.md) · [Concurrency](../../guide/concurrency.md)
