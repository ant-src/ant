<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# WaitGroup.add

```ant signature
func WaitGroup.add(wg: WaitGroup, n: int): void
```

WaitGroup.add adds n, which may be negative, to the counter of wg. It panics if the counter would go
negative or past \[Math.maxInt\].

<!-- end generated -->

## Details

- Call `add` before the `go` that starts the work, not inside it. A `wait` that runs while the
  counter is zero returns at once.
- `add(wg, -1)` takes one off the counter, as `done` does. A goroutine signals that it finished with
  `done`.
- A negative `add` that brings the counter to zero releases every `WaitGroup.wait` on `wg`, and what
  the goroutine wrote before it is visible once `wait` returns, as with `done`.

## Basic examples

Add one to the counter for each goroutine before starting it. Each goroutine calls `done` when it
finishes, and `wait` returns when the counter is back at zero:

```ant
let wg = WaitGroup.new()
let lengths = Array.filled(3, 0)
for entry in ["ant", "bee", "wasp"].entries() {
  WaitGroup.add(wg, 1)
  go {
    defer { wg.done() }
    lengths[entry.index] = entry.value.length()
  }
}
wg.wait()
println(lengths)
```

```text
[3, 3, 4]
```

Add the whole batch at once when its size is known up front:

```ant
let words = ["red", "green", "blue"]
let wg = WaitGroup.new()
let upper = Array.filled(words.length(), "")
WaitGroup.add(wg, words.length())
for entry in words.entries() {
  go {
    defer { wg.done() }
    upper[entry.index] = entry.value.toUpper()
  }
}
wg.wait()
println(upper)
```

```text
["RED", "GREEN", "BLUE"]
```

## Scope

`n` may be zero, which changes nothing, or negative, which takes work off the counter:

```ant
let wg = WaitGroup.new()
wg.add(0)
wg.add(3)
wg.add(-3)
wg.add(-1) // panic: WaitGroup.add: the counter would go negative
```

A wait group in a record field counts the work of that record. Here a server counts the requests it
is handling, so `close` can wait for them:

```ant
type Server = { name: string, active: WaitGroup = WaitGroup.new(), served: Atomic[int] = Atomic.new(0) }

func Server.handle(s: Server, request: string) {
  s.active.add(1)
  go {
    defer { s.active.done() }
    Time.sleep(5ms)
    s.served.add(1)
  }
}

func Server.close(s: Server) {
  s.active.wait()
  println("${s.name} closed after ${s.served.load()} requests")
}

let server = Server{ name: "api" }
for request in ["/", "/users", "/health"] {
  server.handle(request)
}
server.close()
```

```text
api closed after 3 requests
```

## Applications

Reserve the whole batch, then give back the items that need no goroutine with a negative `add`:

```ant
let inputs = ["12", "", "7", "", "30"]
let wg = WaitGroup.new()
let total = Atomic.new(0)
wg.add(inputs.length())
for input in inputs {
  if input.isEmpty() {
    wg.add(-1)
    continue
  }
  go {
    defer { wg.done() }
    total.add(input.toInt() ?? 0)
  }
}
wg.wait()
println(total.load())
```

```text
49
```

A goroutine may start more work and add it to the same group, as long as it adds before its own
`done`. The counter then never reaches zero early:

```ant
let wg = WaitGroup.new()
let visited = Atomic.new(0)

func visit(depth: int): void {
  visited.add(1)
  if depth == 0 { return }
  for _ in Array.range(0, 2) {
    wg.add(1)
    go {
      defer { wg.done() }
      visit(depth - 1)
    }
  }
}

wg.add(1)
go {
  defer { wg.done() }
  visit(3)
}
wg.wait()
println(visited.load())
```

```text
15
```

## Properties and relations

`add(wg, -1)` and `done` both take one off the counter, so after both a `wait` returns at once:

```ant
let wg = WaitGroup.new()
wg.add(2)
wg.done()
wg.add(-1)
wg.wait()
println("counter is back at zero")
```

```text
counter is back at zero
```

## Possible issues

An `add` inside the goroutine comes too late. Here the goroutine has not run its `add` yet, so the
counter is zero and `wait` returns before the work is done:

```ant
let wg = WaitGroup.new()
let start: Channel[bool] = Channel.new()
let finished = Atomic.new(false)
go {
  start.receive()
  wg.add(1)
  defer { wg.done() }
  finished.store(true)
}
wg.wait()
println("finished: ${finished.load()}")
start.send(true)
```

```text
finished: false
```

An `add` with no matching `done` keeps the counter above zero, and a `wait` without a context never
returns. A context bounds the wait, and `Context.isDone` reports whether it gave up:

```ant
let wg = WaitGroup.new()
wg.add(2)
go { wg.done() }
let ctx = Context.new(timeout: 50ms)
wg.wait(ctx: ctx)
println(if ctx.isDone() { "gave up: one done is missing" } else { "all done" })
```

```text
gave up: one done is missing
```

## See also

[WaitGroup.done](done.md) · [WaitGroup.wait](wait.md) · [WaitGroup.new](new.md) ·
[WaitGroup](index.md) · [Parallel.each](../Parallel/each.md) · [Atomic](../Atomic/index.md) ·
[defer](../../language/keywords/defer.md) · [Concurrency guide](../../guide/concurrency.md)
