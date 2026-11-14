<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Channel.close

```ant signature
func Channel.close[T](c: Channel[T]): void
```

Channel.close closes c: receivers drain what is buffered and then get null. It panics if c is
already closed.

<!-- end generated -->

## Details

- Receivers take the buffered values in order. Once the buffer is empty, every receive returns
  `null` at once.
- A close wakes every receiver waiting on `c`, so one close reaches all of them.
- A send on a closed channel panics, so close a channel only once nothing will send on it.
- It also panics on the channel of a `Time.ticker`.

## Basic examples

Close a buffered channel, then receive. The buffered values come first, then `null`:

```ant
let c: Channel[int] = Channel.new(3)
c.send(1)
c.send(2)
Channel.close(c)
println(c.receive())
println(c.receive())
println(c.receive())
```

```text
1
2
null
```

A producer closes its channel when it has sent everything, and `defer` makes sure it does. The `for`
loop over `Channel.values` ends at the close:

```ant
func produce(out: Channel[string]) {
  defer { out.close() }
  for word in ["ant", "bee", "cat"] {
    out.send(word)
  }
}

let words: Channel[string] = Channel.new()
go { produce(words) }
for word in words.values() {
  println(word)
}
println("done")
```

```text
ant
bee
cat
done
```

## Scope

An empty channel, buffered or not, can be closed. A receive on it returns `null` without waiting,
and `Channel.isClosed` reports the close:

```ant
let c: Channel[string] = Channel.new()
c.close()
println(c.receive())
println(c.isClosed())
```

```text
null
true
```

A close wakes every goroutine waiting to receive. Each gets `null`, so a channel that never carries
a value works as a stop signal for any number of workers:

```ant
type Stop = {}

let stop: Channel[Stop] = Channel.new()
let stopped = Atomic.new(0)
let wg = WaitGroup.new()
for _ in Array.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    if stop.receive() == null { stopped.add(1) }
  }
}
stop.close()
wg.wait()
println("${stopped.load()} workers stopped")
```

```text
3 workers stopped
```

In a `select`, a receive on a closed channel is always ready and binds `null`, so it is chosen
before `else`:

```ant
let lines: Channel[string] = Channel.new()
lines.close()
let got = select {
  var line = Channel.receive(lines) { line ?? "closed" }
  else { "nothing ready" }
}
println(got)
```

```text
closed
```

## Applications

Fill a job queue, close it, and let `Parallel.each` work through it. The call returns once the
channel is closed and drained:

```ant
let jobs: Channel[int] = Channel.new(10)
for n in Array.range(1, 6) {
  jobs.send(n)
}
jobs.close()
let total = Atomic.new(0)
Parallel.each(jobs, func(n: int, work: Context) { total.add(n * n) }, limit: 3)
println(total.load())
```

```text
55
```

When several goroutines send on one channel, none of them may close it alone, since the others would
then send on a closed channel. Count them with a `WaitGroup` and close once the count is zero:

```ant
let results: Channel[int] = Channel.new()
let wg = WaitGroup.new()
for n in [10, 20, 30] {
  wg.add(1)
  go {
    defer { wg.done() }
    results.send(n * 2)
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
120
```

## Properties and relations

A close leaves the buffer as it was. Draining it gives the values in the order they were sent, and
every receive after that is `null`:

```ant
let c: Channel[string] = Channel.new(4)
c.send("a")
c.send("b")
c.close()
assert(c.isClosed())
assert(c.length() == 2)
assert((c.values() | Iter.toArray()) == ["a", "b"])
assert(c.length() == 0)
assert(c.receive() == null)
assert(c.receive() == null)
```

## Possible issues

A second close panics. Close a channel in one place, the place that knows the sending is over:

```ant
func finish(c: Channel[int]) {
  c.close() // panic: Channel.close: the channel is closed already
}

let results: Channel[int] = Channel.new()
finish(results)
finish(results)
```

A send after the close panics too. To stop a sender from the receiving side, do not close its
channel: cancel a `Context` that the sender passes to `Channel.send`, which then returns `false`:

```ant
let c: Channel[int] = Channel.new(1)
c.close()
let sent = c.send(1) // panic: Channel.send: the channel is closed
```

A `select` with a send clause on a closed channel panics as well, even when it has an `else`:

```ant
let c: Channel[int] = Channel.new(1)
c.close()
select { // panic: select: a send clause was chosen on a closed channel
  Channel.send(c, 1) { println("sent") }
  else { println("full") }
}
```

A ticker's channel only receives, so closing it panics. Stop a ticker by cancelling the context it
was given instead:

```ant
func stopTicks(ticks: Channel[Time.Instant]) {
  ticks.close() // panic: Channel.close: a ticker's channel only receives
}

stopTicks(Time.ticker(1s))
```

When `T` includes `null`, a `null` value looks like the close. Here the first receive takes the
value `null` and the second sees the close, yet both print the same:

```ant
let c: Channel[int | null] = Channel.new(2)
c.send(null)
c.close()
println(c.receive())
println(c.receive())
```

```text
null
null
```

Wrap such values in a record, so that only the close gives `null`.

A channel is not a sequence, so `for` over it is a compile error. Loop over `Channel.values`:

```ant
let c: Channel[int] = Channel.new()
c.close()
for x in c { println(x) }
// error: for needs an array, a map, a set, a string, Bytes, or a sequence, got Channel[int]
```

## See also

[Channel.new](new.md) · [Channel.send](send.md) · [Channel.receive](receive.md) ·
[Channel.values](values.md) · [Channel.isClosed](isClosed.md) · [Parallel.each](../Parallel/each.md)
· [WaitGroup](../WaitGroup/index.md) · [select](../../language/keywords/select.md) ·
[Concurrency guide](../../guide/concurrency.md)
