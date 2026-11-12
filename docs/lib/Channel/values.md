<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Channel.values

```ant signature
func Channel.values[T](c: Channel[T], ctx: Context | null = null): Seq[T]
```

Channel.values returns a sequence of every value received from c until it is closed and drained, or
until ctx is done.

<!-- end generated -->

## Details

- The loop receives each value, so a value it yields is gone from the channel.
- The sequence ends only when the channel is closed and drained, or when `ctx` is done. A loop over
  a channel nobody closes, without a context, waits until the run is cancelled.
- Unlike `Channel.receive`, the sequence tells a `null` value from a close, so a `Channel[T | null]`
  yields its `null` values.
- When `ctx` ends the loop, nothing tells it apart from a close. Ask `Context.isDone(ctx)`
  afterward.

## Basic examples

Print every value a goroutine sends, until it closes the channel:

```ant
let numbers: Channel[int] = Channel.new()
go {
  defer { numbers.close() }
  for n in [1, 2, 3] {
    numbers.send(n)
  }
}
for n in Channel.values(numbers) {
  println(n)
}
```

```text
1
2
3
```

A closed buffered channel still yields what it holds:

```ant
let words: Channel[string] = Channel.new(capacity: 2)
words.send("ant")
words.send("bee")
words.close()
for w in Channel.values(words) {
  println(w)
}
```

```text
ant
bee
```

## Scope

A closed, empty channel yields nothing, so the loop body never runs:

```ant
let empty: Channel[int] = Channel.new()
empty.close()
for n in empty.values() {
  println(n)
}
println("done")
```

```text
done
```

A channel of a union yields each value with its own case. Here `null` is an ordinary value, not the
end of the channel:

```ant
let readings: Channel[int | null] = Channel.new(capacity: 3)
readings.send(20)
readings.send(null)
readings.send(22)
readings.close()
for r in readings.values() {
  println(r ?? "missing")
}
```

```text
20
missing
22
```

Records and tagged unions travel as they are:

```ant
type Job = { id: int, name: string }
type Stop = { reason: string }

let events: Channel[Job | Stop] = Channel.new(capacity: 3)
events.send(Job{ id: 1, name: "build" })
events.send(Job{ id: 2, name: "test" })
events.send(Stop{ reason: "shutdown" })
events.close()
for e in events.values() {
  if e is Job {
    println("job ${e.id}: ${e.name}")
  } else {
    println("stop: ${e.reason}")
  }
}
```

```text
job 1: build
job 2: test
stop: shutdown
```

The result is a `Seq[T]`, so the `Iter` functions chain on it in a pipe:

```ant
let c: Channel[int] = Channel.new()
go {
  defer { c.close() }
  for n in Iter.range(1, 10) {
    c.send(n)
  }
}
let evenSquares = c.values()
  | Iter.filter({ x % 2 == 0 })
  | Iter.map({ x * x })
  | Iter.toArray()
println(evenSquares)
```

```text
[4, 16, 36, 64]
```

## Named arguments

### ctx

With `ctx:`, the sequence also ends when the context is done. Here the channel holds two values and
is never closed, so the timeout ends the loop, and `Context.reason` tells why:

```ant
let c: Channel[string] = Channel.new(capacity: 2)
c.send("first")
c.send("second")
let ctx = Context.new(Time.duration(milliseconds: 50))
for s in c.values(ctx: ctx) {
  println(s)
}
println(ctx.reason())
```

```text
first
second
timeout
```

## Applications

Collect the results of several workers. Each worker sends on one channel, and a goroutine closes it
when the `WaitGroup` reaches zero, which ends the loop:

```ant
let results: Channel[string] = Channel.new()
let wg = WaitGroup.new()
for name in ["ann", "bob", "cid"] {
  wg.add(1)
  go {
    defer { wg.done() }
    results.send(name.toUpper())
  }
}
go {
  wg.wait()
  results.close()
}
let names = results.values() | Iter.toArray() | Array.sort()
println(names)
```

```text
["ANN", "BOB", "CID"]
```

Stop at the first value that matches:

```ant
let queue: Channel[int] = Channel.new(capacity: 5)
for n in [3, 8, 12, 5, 20] {
  queue.send(n)
}
queue.close()
var found: int | null = null
for n in queue.values() {
  if n > 10 {
    found = n
    break
  }
}
println(found)
```

```text
12
```

Sum a stream of numbers from a producer that reads them from text:

```ant
func produce(text: string, out: Channel[int]) {
  defer { out.close() }
  for line in text.split("\n") {
    let n = line.trim().toInt()
    if n == null { continue }
    out.send(n)
  }
}

let numbers: Channel[int] = Channel.new()
go { produce("4\n x\n10\n6", numbers) }
var total = 0
for n in numbers.values() {
  total += n
}
println(total)
```

```text
20
```

## Properties and relations

For a channel whose element type excludes `null`, the sequence yields the same values as calling
`Channel.receive` until it returns `null`:

```ant
func fill(): Channel[int] {
  let c: Channel[int] = Channel.new(capacity: 3)
  for n in [1, 2, 3] {
    c.send(n)
  }
  c.close()
  c
}

let viaValues = fill().values() | Iter.toArray()

let c = fill()
let viaReceive: Array[int] = []
while true {
  let n = c.receive() ?? break
  viaReceive.push(n)
}
assert(viaValues == viaReceive)
```

The loop takes the values out, so a second run over the same closed channel yields nothing:

```ant
let c: Channel[int] = Channel.new(capacity: 2)
c.send(1)
c.send(2)
c.close()
let vs = c.values()
assert((vs | Iter.toArray()) == [1, 2])
assert((vs | Iter.toArray()).isEmpty())
assert(c.isClosed())
```

## Possible issues

The producer must close the channel. Without `close` and without a context, the loop waits for a
value that never comes, until the run is cancelled. Close with `defer` as the producer's first line,
so a `return` or a panic in the producer still ends the loop:

```ant
let lines: Channel[string] = Channel.new()
go {
  defer { lines.close() }
  lines.send("only line")
}
for line in lines.values() {
  println(line)
}
```

```text
only line
```

A ticker is never closed, so a loop over its ticks ends only by `break`, a context, or `Iter.take`:

```ant
let ticks = Time.ticker(Time.duration(milliseconds: 10))
let first = ticks.values() | Iter.take(3) | Iter.toArray()
println(first.length())
```

```text
3
```

## See also

[Channel.receive](receive.md), [Channel.close](close.md), [Channel.new](new.md),
[Iter.toArray](../Iter/toArray.md), [Iter.take](../Iter/take.md),
[Parallel.each](../Parallel/each.md), [Time.ticker](../Time/ticker.md),
[for](../../language/keywords/for.md), [go](../../language/keywords/go.md)
