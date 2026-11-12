<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Channel.send

```ant signature
func Channel.send[T](c: Channel[T], v: T, ctx: Context | null = null): bool
```

Channel.send sends v on c, blocking until a receiver takes it or the buffer has room. It returns
true once v is taken or buffered, and false if ctx is done first; a ctx done already buffers
nothing. It panics if c is closed, even when ctx is done, or if c closes while the send waits.

<!-- end generated -->

## Details

- On a channel of capacity 0, `send` returns once a receiver has taken the value. On a buffered
  channel it returns as soon as the value is buffered, and blocks only while the buffer is full.
- It returns `false` only when `ctx` is done before the value is taken or buffered; the value is
  then not sent.
- Sending on a closed channel panics, even with a `ctx` that is done, and so does sending on a
  [Time.ticker](../Time/ticker.md) channel. A send waiting when the channel closes panics too.
- A `ctx` that is done already gives up at once, so `send` buffers nothing and returns `false`, even
  when the buffer has room.

## Basic examples

Hand a value from one goroutine to another:

```ant
let c: Channel[string] = Channel.new()
go { Channel.send(c, "hello") }
println(c.receive())
```

```text
hello
```

On a buffered channel, a send returns at once while there is room:

```ant
let c: Channel[int] = Channel.new(2)
println(Channel.send(c, 1))
println(Channel.send(c, 2))
println(c.length())
```

```text
true
true
2
```

## Scope

Any value can be sent, `null` and records included:

```ant
type Job = { id: int, name: string }
let jobs: Channel[Job | null] = Channel.new(2)
jobs.send(Job{ id: 1, name: "build" })
jobs.send(null)
println(jobs.receive())
println(jobs.receive())
```

```text
Job{ id: 1, name: "build" }
null
```

A producer sends every value and closes the channel when it is done, so the receiver's loop ends:

```ant
let words: Channel[string] = Channel.new()
go {
  defer { words.close() }
  for w in ["ant", "bee", "cat"] {
    words.send(w)
  }
}
for w in words.values() {
  println(w.toUpper())
}
```

```text
ANT
BEE
CAT
```

As a `select` head, a send is written in namespace form and performed only when its clause is
chosen. With `else`, the `select` tries the send without waiting:

```ant
let c: Channel[int] = Channel.new(1)
c.send(1)
let sent = select {
  let ok = Channel.send(c, 2) { ok }
  else { false }
}
println(sent)
```

```text
false
```

## Named arguments

### ctx

With `ctx:`, the send gives up when the context is done and returns `false`. Here the buffer is full
and no one receives:

```ant
let c: Channel[int] = Channel.new(1)
c.send(1)
let ctx = Context.new(timeout: Time.duration(milliseconds: 20))
println(c.send(2, ctx: ctx))
println(ctx.reason())
```

```text
false
timeout
```

A producer loops while its sends succeed, and stops when the consumer cancels the context:

```ant
let numbers: Channel[int] = Channel.new()
let ctx = Context.new()
let stopped = WaitGroup.new()
stopped.add(1)
go {
  defer { stopped.done() }
  var n = 1
  while numbers.send(n, ctx: ctx) {
    n *= 2
  }
  println("producer stopped: ${ctx.reason()}")
}
for _ in Iter.range(0, 4) {
  println(numbers.receive())
}
ctx.cancel()
stopped.wait()
```

```text
1
2
4
8
producer stopped: cancelled
```

## Applications

A mailbox: several producers send different event records into one channel, and one receiver handles
them in arrival order. The context stops the producer still waiting once the receiver is done:

```ant
type Progress = { percent: int }
type Done = { files: int }
type Timeout = {}

let events: Channel[Progress | Done | Timeout] = Channel.new()
let ctx = Context.new()
defer { ctx.cancel() }
go {
  for p in [25, 50, 100] {
    events.send(Progress{ percent: p }, ctx: ctx)
  }
  events.send(Done{ files: 12 }, ctx: ctx)
}
go {
  Time.sleep(Time.duration(seconds: 5), ctx: ctx)
  events.send(Timeout{}, ctx: ctx)
}
while true {
  let event = events.receive()
  if event is Progress {
    println("${event.percent}%")
  } else if event is Done {
    println("done: ${event.files} files")
    break
  } else {
    println("timed out")
    break
  }
}
```

```text
25%
50%
100%
done: 12 files
```

A pipeline stage receives from one channel, transforms each value, and sends it on. Each stage
closes its output when its input ends:

```ant
let words: Channel[string] = Channel.new()
let lengths: Channel[int] = Channel.new()
go {
  defer { words.close() }
  for w in ["ant", "beetle", "cicada"] {
    words.send(w)
  }
}
go {
  defer { lengths.close() }
  for w in words.values() {
    lengths.send(w.length())
  }
}
for n in lengths.values() {
  println(n)
}
```

```text
3
6
6
```

Offer a value, but give up when no one takes it in time, with a [Context.wait](../Context/wait.md)
clause beside the send:

```ant
let out: Channel[string] = Channel.new()
let ctx = Context.new(timeout: Time.duration(milliseconds: 20))
let outcome = select {
  Channel.send(out, "report") { "sent" }
  Context.wait(ctx) { "nobody listened" }
}
println(outcome)
```

```text
nobody listened
```

## Properties and relations

Values come out of a channel in the order they were sent:

```ant
let c: Channel[int] = Channel.new(3)
for n in [1, 2, 3] {
  assert(c.send(n))
}
assert(c.length() == 3)
assert(c.receive() == 1)
assert(c.receive() == 2)
assert(c.receive() == 3)
```

On an unbuffered channel, `send` returns only after a receiver has taken the value:

```ant
let c: Channel[int] = Channel.new()
go {
  Time.sleep(Time.duration(milliseconds: 30))
  c.receive()
}
let start = Time.monotonic()
c.send(1)
assert(Time.monotonic() - start >= 30)
```

## Possible issues

Sending on a closed channel panics. Close a channel from the sending side, after the last send:

```ant
let c: Channel[int] = Channel.new(1)
c.close()
let sent = c.send(1) // panic: Channel.send: the channel is closed
```

A send clause on a closed channel is always ready, so a `select` chooses it over `else` and panics:

```ant
let c: Channel[int] = Channel.new()
c.close()
let sent = select { // panic: select: a send clause was chosen on a closed channel
  let ok = Channel.send(c, 1) { ok }
  else { false }
}
```

A ticker's channel only receives:

```ant
let ticks = Time.ticker(Time.duration(seconds: 1))
let sent = ticks.send(Time.now()) // panic: Channel.send: a ticker's channel only receives
```

A send on an unbuffered channel with no receiver waits until the run is cancelled. In a script with
one goroutine, `Channel.send(c, 1)` never returns on its own. Give the send a context, or start the
receiver first:

```ant
let c: Channel[int] = Channel.new()
let ctx = Context.new(timeout: Time.duration(milliseconds: 20))
println(c.send(1, ctx: ctx))
```

```text
false
```

A producer blocked in `send` is not woken when a [Parallel.each](../Parallel/each.md) that reads its
channel ends. Here `each` stops after the third job, and a plain `jobs.send(n)` would then wait
until the run is cancelled. Give the producer a context, so its send returns `false`:

```ant
let jobs: Channel[int] = Channel.new()
let ctx = Context.new()
let producer = WaitGroup.new()
producer.add(1)
go {
  defer { producer.done() }
  var n = 1
  while jobs.send(n, ctx: ctx) {
    n += 1
  }
  println("producer stopped: ${ctx.reason()}")
}
Parallel.each(jobs, func(n: int, work: Context) {
  println("job ${n}")
  if n == 3 { ctx.cancel() }
}, limit: 1, ctx: ctx)
producer.wait()
```

```text
job 1
job 2
job 3
producer stopped: cancelled
```

A send head in a `select` takes no context; the context is a clause of its own:

```ant
let c: Channel[int] = Channel.new()
let ctx = Context.new()
let r = select {
  Channel.send(c, 1, ctx) { "sent" } // error: a select clause's head takes no context: the context is a clause of its own, Context.wait(ctx) { ... }
  Context.wait(ctx) { "stopped" }
}
```

## See also

[Channel.receive](receive.md) · [Channel.new](new.md) · [Channel.close](close.md) ·
[Channel.values](values.md) · [Channel.length](length.md) · [Context.new](../Context/new.md) ·
[Context.wait](../Context/wait.md) · [Parallel.each](../Parallel/each.md)
