<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Channel.isClosed

```ant signature
func Channel.isClosed[T](c: Channel[T]): bool
```

Channel.isClosed reports whether c has been closed.

<!-- end generated -->

## Details

- A closed channel can still hold buffered values: `Channel.isClosed` is true from the moment
  `Channel.close` returns, before the values are received. Receive until `null` to drain it.
- The answer is the state at the moment of the call. Another goroutine can close the channel right
  after it returns false.
- A `Time.ticker` channel is never closed, so `Channel.isClosed` is always false for it.

## Basic examples

A new channel is open, and closing it makes `Channel.isClosed` true:

```ant
let c: Channel[int] = Channel.new()
println(Channel.isClosed(c))
Channel.close(c)
println(Channel.isClosed(c))
```

```text
false
true
```

A producer closes the channel when it is done. The producer blocks on its first send until a value
is received, so the channel is still open on the first check, and closed once every value is in:

```ant
func produce(out: Channel[int]) {
  defer { out.close() }
  for n in [1, 2, 3] {
    out.send(n)
  }
}

let numbers: Channel[int] = Channel.new()
go { produce(numbers) }
println(numbers.isClosed())
let total = numbers.values().toArray().sum()
println("${total} ${numbers.isClosed()}")
```

```text
false
6 true
```

## Scope

A buffered channel is closed while it still holds values. `Channel.isClosed` and `Channel.length`
answer different questions: whether the channel has been closed, and how many values are waiting:

```ant
let c: Channel[string] = Channel.new(2)
c.send("a")
c.send("b")
c.close()
println("${c.isClosed()} ${c.length()}")
println(c.receive())
println(c.receive())
println(c.receive())
```

```text
true 2
a
b
null
```

A ticker's channel is never closed, not even after its context is done:

```ant
let ctx = Context.new()
let ticks = Time.ticker(10ms, ctx: ctx)
ctx.cancel()
println(ticks.isClosed())
```

```text
false
```

## Applications

### Tell a close from a timeout

`Channel.receive` gives `null` both when the channel is closed and drained and when its context ends
first. Ask the context which one happened: a closed channel can still hold values when the context
ends, so `Channel.isClosed` alone does not tell the two apart. The last call gets a fresh context,
because the first one is done for good:

```ant
func next(jobs: Channel[string], ctx: Context): string {
  let job = jobs.receive(ctx: ctx)
  if job != null { return "job ${job}" }
  if ctx.isDone() { "gave up: ${ctx.reason()}" } else { "no more jobs" }
}

let jobs: Channel[string] = Channel.new(1)
jobs.send("build")
let ctx = Context.new(timeout: 20ms)
println(next(jobs, ctx))
println(next(jobs, ctx))
jobs.close()
println(next(jobs, Context.new()))
```

```text
job build
gave up: timeout
no more jobs
```

### Close a channel once from several goroutines

Closing a channel signals every receiver at once, but a second close panics. Check and close under
one lock, so that no other goroutine closes the channel between the two calls:

```ant
type Shutdown = { done: Channel[bool], lock: Mutex }

func Shutdown.new(): Shutdown {
  Shutdown{ done: Channel.new(), lock: Mutex.new() }
}

func Shutdown.trigger(s: Shutdown) {
  s.lock.withLock({
    if !s.done.isClosed() { s.done.close() }
  })
}

let shutdown = Shutdown.new()
let wg = WaitGroup.new()
for _ in [1, 2, 3] {
  wg.add(1)
  go {
    defer { wg.done() }
    shutdown.trigger()
  }
}
wg.wait()
println(shutdown.done.receive())
```

```text
null
```

## Properties and relations

A channel stays closed while it is drained, and a receive on a closed, drained channel gives `null`:

```ant
let c: Channel[int] = Channel.new(3)
for n in [1, 2, 3] {
  c.send(n)
}
c.close()
assert(c.isClosed())
assert(c.receive() == 1)
assert(c.values().toArray() == [2, 3])
assert(c.receive() == null)
assert(c.isClosed())
```

## Possible issues

### Closed is not empty

A loop that runs while the channel is open misses the values buffered before the close. Receive
until `null`, or use `Channel.values`, to get them:

```ant
let c: Channel[int] = Channel.new(3)
for n in [1, 2, 3] {
  c.send(n)
}
c.close()

let got: Array[int] = []
while !c.isClosed() {
  got.push(c.receive() ?? 0)
}
println(got)
println(c.values().toArray())
```

```text
[]
[1, 2, 3]
```

### Closing twice panics

`Channel.close` on a closed channel panics. Here the second call of `finish` closes the channel
again:

```ant
func finish(results: Channel[int]) {
  results.close() // panic: Channel.close: the channel is closed already
}

let results: Channel[int] = Channel.new()
finish(results)
finish(results)
```

A send on a closed channel panics too, even when the buffer has room:

```ant
let c: Channel[int] = Channel.new(1)
c.close()
let sent = c.send(1) // panic: Channel.send: the channel is closed
```

Checking `Channel.isClosed` first avoids the panic only when no other goroutine can close the
channel between the check and the close; otherwise hold a lock around both, as in
[Close a channel once from several goroutines](#close-a-channel-once-from-several-goroutines).

## See also

[Channel.close](close.md), [Channel.receive](receive.md), [Channel.values](values.md),
[Channel.length](length.md), [Channel.send](send.md), [Context.isDone](../Context/isDone.md),
[Time.ticker](../Time/ticker.md), [Channel](index.md)
